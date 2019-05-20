package bloom

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/hash"
)

func maxValue(f *BloomFilter) uint8 {
	max := uint8(0)
	for _, value := range f.Buckets {
		if value > max {
			max = value
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

func TestAddRemove(t *testing.T) {
	filter := NewFilterWithFalsePositiveRate(0.4, 1000000)

	hashes := make([][]byte, 1000000)
	for i := range hashes {
		hashes[i] = makeHash(t)
		filter.Add(hashes[i])
	}

	for _, h := range hashes {
		require.False(t, filter.IsNotPresent(h))
	}

	falsePositives := 0
	for i := 0; i < 100000; i += 1 {
		if !filter.IsNotPresent(makeHash(t)) {
			falsePositives += 1
		}
	}

	fmt.Printf("False positive rate: %f\n", float64(falsePositives)/float64(100000))
	fmt.Printf("Filter size: %d\n", len(filter.Buckets)*4)
	fmt.Printf("Max value: %d\n", maxValue(filter))
}
