package bloom

import (
	"crypto/rand"
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

func requirePanic(t *testing.T, cb func()) {
	defer func() {
		r := recover()
		require.NotNil(t, r)
	}()

	cb()
}

const maxFilterSize = 1048576

func TestInvalidFalsePositive(t *testing.T) {
	requirePanic(t, func() { NewFilterWithFalsePositiveRate(-0.4, 1000, maxFilterSize) })
	requirePanic(t, func() { NewFilterWithFalsePositiveRate(-0.0, 1000, maxFilterSize) })
	requirePanic(t, func() { NewFilterWithFalsePositiveRate(0.0, 1000, maxFilterSize) })
	requirePanic(t, func() { NewFilterWithFalsePositiveRate(1.0, 1000, maxFilterSize) })
}

func TestHashLength(t *testing.T) {
	// We can handle arbitrary hash lengths greater than 8 bytes
	filter := NewFilterWithFalsePositiveRate(0.01, 1000, maxFilterSize)

	hash := makeHash(t)
	for i := 0; i < 8; i += 1 {
		partial := hash[0:i]
		requirePanic(t, func() { filter.Add(partial) })
		requirePanic(t, func() { filter.Remove(partial) })
		requirePanic(t, func() { filter.IsNotPresent(partial) })
		requirePanic(t, func() { filter.UpperBoundCount(partial) })
	}

	for i := 8; i < 64; i += 1 {
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
	expectedSize := FilterSizeForFalsePositiveRate(0.001, 10000)
	require.Equal(t, maxFilterSize, expectedSize)
}

func TestFalsePositiveRate(t *testing.T) {
}

func TestZeroElementCount(t *testing.T) {
}
