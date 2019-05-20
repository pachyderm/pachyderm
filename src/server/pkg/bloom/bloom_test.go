package bloom

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/hash"
)

func maxValue(f *BloomFilter) uint32 {
	max := uint32(0)
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
	filter, err := NewFilter(FilterConstraints{ElementCount: 1024, FalsePositiveRate: 0.1}, 64)
	require.NoError(t, err)

	hashes := make([][]byte, 1024)
	for i := range hashes {
		hashes[i] = makeHash(t)
		filter.Add(hashes[i])
	}

	for _, h := range hashes {
		require.False(t, filter.IsNotPresent(h))
	}

	falsePositives := 0
	for i := 0; i < 1024; i += 1 {
		if !filter.IsNotPresent(makeHash(t)) {
			falsePositives += 1
		}
	}

	fmt.Printf("%d/1024 false positives: %f\n", falsePositives, float64(falsePositives)/1024)
	fmt.Printf("expected false positive rate: %f\n", filter.ExpectedFalsePositiveRate(1024))
	fmt.Printf("max value: %d\n", maxValue(filter))
}
