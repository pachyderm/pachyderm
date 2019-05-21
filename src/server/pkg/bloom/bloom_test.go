package bloom

import (
	"crypto/rand"
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
	require.Panic(t, func() { NewFilterWithFalsePositiveRate(-0.4, 1000, maxFilterSize) })
	require.Panic(t, func() { NewFilterWithFalsePositiveRate(-0.0, 1000, maxFilterSize) })
	require.Panic(t, func() { NewFilterWithFalsePositiveRate(0.0, 1000, maxFilterSize) })
	require.Panic(t, func() { NewFilterWithFalsePositiveRate(1.0, 1000, maxFilterSize) })
	require.Panic(t, func() { NewFilterWithFalsePositiveRate(0.1, 0, maxFilterSize) })
	require.Panic(t, func() { NewFilterWithFalsePositiveRate(0.1, -1, maxFilterSize) })
	require.Panic(t, func() { NewFilterWithFalsePositiveRate(0.1, 1000, 0) })
	require.Panic(t, func() { NewFilterWithFalsePositiveRate(0.1, 1000, -100) })
	require.Panic(t, func() { NewFilterWithSize(1000, 0) })
	require.Panic(t, func() { NewFilterWithSize(1000, -4) })
	require.Panic(t, func() { NewFilterWithSize(0, 400) })
	require.Panic(t, func() { NewFilterWithSize(-10, 400) })
}

func TestHashLength(t *testing.T) {
	// We can handle arbitrary hash lengths greater than 8 bytes
	filter := NewFilterWithFalsePositiveRate(0.01, 1000, maxFilterSize)

	hash := makeHash(t)
	for i := 0; i < 8; i++ {
		partial := hash[0:i]
		require.Panic(t, func() { filter.Add(partial) })
		require.Panic(t, func() { filter.Remove(partial) })
		require.Panic(t, func() { filter.IsNotPresent(partial) })
		require.Panic(t, func() { filter.UpperBoundCount(partial) })
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
	require.Equal(t, maxFilterSize, len(filter.Buckets))

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
