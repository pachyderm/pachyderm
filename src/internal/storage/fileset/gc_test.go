package fileset

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestGC(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)
	tr := track.NewTestTracker(t, db)
	s := testutil.NewFilesetStorage(t, db, tr)
	gc := s.newGC()
	w := s.NewWriter(ctx, WithTTL(time.Hour))
	err := w.Add("a.txt", func(fw *FileWriter) error {
		fw.Add("tag1")
		_, err := fw.Write([]byte("test data"))
		return err
	})
	require.NoError(t, err)
	id, err := w.Close()
	require.NoError(t, err)
	// check that it's there
	exists, err := s.exists(ctx, *id)
	require.NoError(t, err)
	require.True(t, exists)
	// run the gc
	_, err = gc.RunOnce(ctx)
	require.NoError(t, err)
	// check it's still there
	exists, err = s.exists(ctx, *id)
	require.NoError(t, err)
	require.True(t, exists)
	// expire it
	require.NoError(t, s.Drop(ctx, *id))
	// run the gc
	countDeleted, err := gc.RunOnce(ctx)
	require.NoError(t, err)
	t.Log(countDeleted)
	require.True(t, countDeleted > 0)

	// check that it's not there
	exists, err = s.exists(ctx, *id)
	require.NoError(t, err)
	require.False(t, exists)
}
