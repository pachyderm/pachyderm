package chunk

import (
	"bytes"
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/chmduquesne/rollinghash/buzhash64"
	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

func TestCheck(t *testing.T) {
	ctx := context.Background()
	store, chunks := newTestStorage(t)
	writeRandom(t, chunks)
	n, err := chunks.Check(ctx, nil, nil, false)
	require.NoError(t, err)
	require.True(t, n > 0)

	deleteOne(t, store) // Something terrible has happened.

	_, err = chunks.Check(ctx, nil, nil, false)
	require.YesError(t, err)
}

func deleteOne(t testing.TB, s kv.Store) {
	ctx := pctx.TestContext(t)
	it := s.NewKeyIterator(kv.Span{})
	key, err := stream.Next(ctx, it)
	require.NoError(t, err)
	require.NoError(t, s.Delete(ctx, key))
}

func TestChunkKey(t *testing.T) {
	hash := pachhash.Sum([]byte("foo"))
	eid := ID(hash[:])
	egen := uint64(1234)
	k := chunkKey(eid[:], egen)
	aid, agen, err := parseKey(k)
	require.Equal(t, eid, aid)
	require.Equal(t, egen, agen)
	require.NoError(t, err)
}

func TestList(t *testing.T) {
	ctx := pctx.TestContext(t)
	_, srv := newTestStorage(t)
	writeRandom(t, srv)

	var count int
	require.NoError(t, srv.ListStore(ctx, func(id ID, gen uint64) error {
		t.Log(id, gen)
		count++
		return nil
	}))
	require.True(t, count > 0)
}

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
func newTestStorage(t testing.TB) (kv.Store, *Storage) {
	db := dockertestenv.NewTestDB(t)
	tr := track.NewTestTracker(t, db)
	return NewTestStorage(t, db, tr)
}
