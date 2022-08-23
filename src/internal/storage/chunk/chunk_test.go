package chunk

import (
	"bytes"
	"math/rand"
	"testing"
	"time"

	"github.com/chmduquesne/rollinghash/buzhash64"
	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// TODO: Write new tests.

// TODO: Reenable.
//func TestCheck(t *testing.T) {
//	ctx := context.Background()
//	objC, chunks := newTestStorage(t)
//	writeRandom(t, chunks)
//	n, err := chunks.Check(ctx, nil, nil, false)
//	require.NoError(t, err)
//	require.True(t, n > 0)
//	deleteOne(t, objC)
//	_, err = chunks.Check(ctx, nil, nil, false)
//	require.YesError(t, err)
//}
//
//func deleteOne(t testing.TB, objC obj.Client) {
//	ctx := context.Background()
//	done := false
//	err := objC.Walk(ctx, "", func(p string) error {
//		if !done {
//			require.NoError(t, objC.Delete(ctx, p))
//			done = true
//		}
//		return nil
//	})
//	require.NoError(t, err)
//}

func BenchmarkRollingHash(b *testing.B) {
	seed := time.Now().UTC().UnixNano()
	random := rand.New(rand.NewSource(seed))
	data := randutil.Bytes(random, 100*units.MB)
	b.SetBytes(int64(len(data)))
	hash := buzhash64.New()
	splitMask := uint64((1 << uint64(23)) - 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.Reset()
		_, err := hash.Write(initialWindow)
		require.NoError(b, err)
		for _, bt := range data {
			hash.Roll(bt)
			//nolint:staticcheck // benchmark is simulating exact usecase
			if hash.Sum64()&splitMask == 0 {
			}
		}
	}
}

func BenchmarkComputeChunks(b *testing.B) {
	seed := time.Now().UTC().UnixNano()
	random := rand.New(rand.NewSource(seed))
	data := randutil.Bytes(random, 100*units.MB)
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		require.NoError(b, ComputeChunks(bytes.NewReader(data), func(_ []byte) error { return nil }))
	}
}

// newTestStorage is like NewTestStorage except it doesn't need an external tracker
// it is for testing this package, not for reuse.
//func newTestStorage(t testing.TB) (obj.Client, *Storage) {
//	db := dockertestenv.NewTestDB(t)
//	tr := track.NewTestTracker(t, db)
//	return NewTestStorage(t, db, tr)
//}
