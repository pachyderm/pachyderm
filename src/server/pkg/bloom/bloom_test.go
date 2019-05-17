package bloom

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/hash"
)

// Uses the storage layer hash
func makeHash(t *testing.T) []byte {
	data := make([]byte, 128)
	n, err := rand.Read(data)
	require.NoError(t, err)
	require.Equal(t, 128, n)
	return hash.Sum(data)
}

func TestAddRemove(t *testing.T) {
	filter, err := NewBloomFilter(64, 256)
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

	fmt.Printf("%d/1024 false positives\n", falsePositives)
}
