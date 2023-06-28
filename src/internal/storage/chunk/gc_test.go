package chunk

import (
	"context"
	"io"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

func TestGC(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tracker := track.NewTestTracker(t, db)
	oc, s := NewTestStorage(t, db, tracker)

	writeRandom(t, s)
	count, err := countObjects(ctx, oc)
	require.NoError(t, err)
	require.True(t, count > 0)

	// set everything to expire
	_, err = db.ExecContext(ctx, `UPDATE storage.tracker_objects SET expires_at = CURRENT_TIMESTAMP - interval '1 hour'`)
	require.NoError(t, err)
	// run the tracker GC
	deleter := track.DeleterMux(func(tid string) track.Deleter {
		switch {
		case strings.HasPrefix(tid, TrackerPrefix):
			return s.NewDeleter()
		case strings.HasPrefix(tid, renew.TmpTrackerPrefix):
			return renew.NewTmpDeleter()
		default:
			return nil
		}
	})
	tgc := track.NewGarbageCollector(tracker, time.Minute, deleter)
	require.NoError(t, tgc.RunUntilEmpty(ctx))

	// run the chunk GC
	gc := NewGC(s, time.Minute)
	require.NoError(t, gc.RunOnce(ctx))

	// make sure there are no objects
	count, err = countObjects(ctx, oc)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func countObjects(ctx context.Context, store kv.Store) (int, error) {
	it := store.NewKeyIterator(kv.Span{})
	var count int
	if err := stream.ForEach(ctx, it, func([]byte) error {
		count++
		return nil
	}); err != nil {
		return -1, errors.EnsureStack(err)
	}
	return count, nil
}

func writeRandom(t testing.TB, s *Storage) {
	ctx := context.Background()
	const seed = 10
	const size = 1e8
	rng := rand.New(rand.NewSource(seed))

	cb := func(_ interface{}, _ []*DataRef) error { return nil }
	u := s.NewUploader(ctx, "test-writer", false, cb)
	require.NoError(t, u.Upload(nil, io.LimitReader(rng, size)))
	require.NoError(t, u.Close())
}
