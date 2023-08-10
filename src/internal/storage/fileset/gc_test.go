package fileset

import (
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

func TestGC(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tr := track.NewTestTracker(t, db)
	s := NewTestStorage(ctx, t, db, tr)
	gc := s.NewGC(time.Minute)
	w := s.NewWriter(ctx, WithTTL(time.Hour))
	require.NoError(t, w.Add("a.txt", "datum1", strings.NewReader("test data")))
	h, err := w.Close()
	require.NoError(t, err)
	// check that it's there
	exists, err := s.exists(ctx, h.id)
	require.NoError(t, err)
	require.True(t, exists)
	// run the gc
	_, err = gc.RunOnce(ctx)
	require.NoError(t, err)
	// check it's still there
	exists, err = s.exists(ctx, h.id)
	require.NoError(t, err)
	require.True(t, exists)
	// expire it
	require.NoError(t, s.Drop(ctx, *h))
	// run the gc
	countDeleted, err := gc.RunOnce(ctx)
	require.NoError(t, err)
	t.Log(countDeleted)
	require.True(t, countDeleted > 0)
}
