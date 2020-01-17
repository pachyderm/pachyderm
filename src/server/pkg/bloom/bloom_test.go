package bloom

import (
	"crypto/rand"
	"math"
	mathrand "math/rand"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/hash"
)

func maxValue(f *BloomFilter) int {
	max := 0
	for _, value := range f.Buckets {
		if int(value) > max {
			max = int(value)
		}
	}
	return max
}

// Uses the storage layer hash
func makeHash(t *testing.T) []byte {
	data := make([]byte, 128)
	n, err := rand.Read(data)
	require.NoError(t, err)
	require.Equal(t, 128, n)
	return hash.Sum(data)
}

const maxFilterSize = 1048576

func TestInvalidConstraints(t *testing.T) {
	require.YesPanic(t, func() { NewFilterWithFalsePositiveRate(-0.4, 1000, maxFilterSize) })
	require.YesPanic(t, func() { NewFilterWithFalsePositiveRate(-0.0, 1000, maxFilterSize) })
	require.YesPanic(t, func() { NewFilterWithFalsePositiveRate(0.0, 1000, maxFilterSize) })
	require.YesPanic(t, func() { NewFilterWithFalsePositiveRate(1.0, 1000, maxFilterSize) })
	require.YesPanic(t, func() { NewFilterWithFalsePositiveRate(0.1, 0, maxFilterSize) })
	require.YesPanic(t, func() { NewFilterWithFalsePositiveRate(0.1, -1, maxFilterSize) })
	require.YesPanic(t, func() { NewFilterWithFalsePositiveRate(0.1, 1000, 0) })
	require.YesPanic(t, func() { NewFilterWithFalsePositiveRate(0.1, 1000, -100) })
	require.YesPanic(t, func() { NewFilterWithSize(1000, 0) })
	require.YesPanic(t, func() { NewFilterWithSize(1000, -4) })
	require.YesPanic(t, func() { NewFilterWithSize(0, 400) })
	require.YesPanic(t, func() { NewFilterWithSize(-10, 400) })
}

func TestHashLength(t *testing.T) {
	// We can handle arbitrary hash lengths greater than 8 bytes
	filter := NewFilterWithFalsePositiveRate(0.01, 1000, maxFilterSize)

	hash := makeHash(t)
	for i := 0; i < 8; i++ {
		partial := hash[0:i]
		require.YesPanic(t, func() { filter.Add(partial) })
		require.YesPanic(t, func() { filter.Remove(partial) })
		require.YesPanic(t, func() { filter.IsNotPresent(partial) })
		require.YesPanic(t, func() { filter.UpperBoundCount(partial) })
	}

	for i := 8; i < 64; i++ {
		partial := hash[0:i]
		filter.Add(partial)
		require.False(t, filter.IsNotPresent(partial))
		require.True(t, filter.UpperBoundCount(partial) > 0)
		filter.Remove(partial)
	}
}

func TestAddRemove(t *testing.T) {
	hashA := makeHash(t)
	hashB := makeHash(t)
	filter := NewFilterWithFalsePositiveRate(0.01, 1000, maxFilterSize)

	require.True(t, filter.IsNotPresent(hashA))
	require.True(t, filter.IsNotPresent(hashB))
	require.Equal(t, 0, maxValue(filter))

	filter.Add(hashA)

	require.False(t, filter.IsNotPresent(hashA))
	require.Equal(t, 1, maxValue(filter))

	filter.Add(hashB)

	require.False(t, filter.IsNotPresent(hashA))
	require.False(t, filter.IsNotPresent(hashB))
	require.True(t, maxValue(filter) > 0)

	filter.Remove(hashA)

	require.False(t, filter.IsNotPresent(hashB))
	require.Equal(t, 1, maxValue(filter))

	filter.Remove(hashB)

	require.Equal(t, 0, maxValue(filter))
	require.True(t, filter.IsNotPresent(hashA))
	require.True(t, filter.IsNotPresent(hashB))
}

func TestLargeAllocation(t *testing.T) {
	desiredFalsePositiveRate := 0.001
	expectedSize := FilterSizeForFalsePositiveRate(desiredFalsePositiveRate, 1000000)
	require.True(t, maxFilterSize < expectedSize)

	// Allocate a filter that is larger than our max size
	filter := NewFilterWithFalsePositiveRate(desiredFalsePositiveRate, 1000000, maxFilterSize)

	// The filter should be truncated to the max size
	require.Equal(t, maxFilterSize, len(filter.Buckets)*bytesPerBucket)

	// The false positive rate should be higher than we wanted
	require.True(t, filter.FalsePositiveRate(1000000) > desiredFalsePositiveRate)
}

func TestFalsePositiveRate(t *testing.T) {
	elementCount := 1024

	// This is a probabilistic test, so seed the RNG to avoid flakiness
	mathrand.Seed(0x23505ea7bb812c38)

	mathHash := func() []byte {
		data := make([]byte, 128)
		n, err := mathrand.Read(data)
		require.NoError(t, err)
		require.Equal(t, 128, n)
		return hash.Sum(data)
	}

	filter := NewFilterWithFalsePositiveRate(0.1, elementCount, maxFilterSize)

	hashes := make([][]byte, elementCount)
	for i := range hashes {
		hashes[i] = mathHash()
		filter.Add(hashes[i])
	}

	for _, h := range hashes {
		require.False(t, filter.IsNotPresent(h))
	}

	falsePositives := 0
	falsePositiveChecks := 10000
	for i := 0; i < falsePositiveChecks; i++ {
		if !filter.IsNotPresent(mathHash()) {
			falsePositives++
		}
	}

	expectedFalsePositiveRate := filter.FalsePositiveRate(elementCount)
	require.True(t, expectedFalsePositiveRate > 0.09)
	require.True(t, expectedFalsePositiveRate < 0.11)

	actualFalsePositiveRate := float64(falsePositives) / float64(falsePositiveChecks)
	require.True(t, actualFalsePositiveRate > 0.09)
	require.True(t, actualFalsePositiveRate < 0.11)
}

func TestOverflow(t *testing.T) {
	filter := NewFilterWithSize(1024, 100)

	// Insert one element
	hash := makeHash(t)
	filter.Add(hash)

	// All buckets should be 0 or 1 - the number of buckets with '1' should be
	// equal to filter.NumSubhashes.
	indexes := map[int]bool{}
	for i, value := range filter.Buckets {
		if value == 1 {
			indexes[i] = true
		} else {
			require.True(t, value == 0)
		}
	}
	require.Equal(t, int(filter.NumSubhashes), len(indexes))

	// Directly modify the buckets to make them just shy of overflow
	for index := range indexes {
		filter.Buckets[index] = math.MaxUint32 - 1
	}

	// This should put all the buckets into overflow
	filter.Add(hash)

	// Helper function to make sure the values in the buckets are correct
	checkOverflow := func() {
		for i, value := range filter.Buckets {
			if _, found := indexes[i]; found {
				require.Equal(t, uint32(math.MaxUint32), value)
			} else {
				require.Equal(t, uint32(0), value)
			}
		}
	}

	checkOverflow()

	// Now that the buckets are in overflow, these are effectively no-ops
	filter.Add(hash)
	checkOverflow()
	filter.Remove(hash)
	checkOverflow()
}
