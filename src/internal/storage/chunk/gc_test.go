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
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/sirupsen/logrus"
)

func TestGC(t *testing.T) {
	ctx := context.Background()
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
	gc := NewGC(s, time.Minute, logrus.StandardLogger())
	require.NoError(t, gc.RunOnce(ctx))

	// make sure there are no objects
	count, err = countObjects(ctx, oc)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func countObjects(ctx context.Context, client obj.Client) (int, error) {
	var count int
	if err := client.Walk(ctx, "", func(string) error {
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

	cb := func(_ []*Annotation) error { return nil }
	w := s.NewWriter(ctx, "test-writer", cb)
	require.NoError(t, w.Annotate(&Annotation{}))
	_, err := io.Copy(w, io.LimitReader(rng, size))
	require.NoError(t, err)
	require.NoError(t, w.Close())
}
