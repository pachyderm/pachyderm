package chunk

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

func TestReduceObjectsPerChunk(t *testing.T) {
	ctx := context.Background()
	db := dbutil.NewTestDB(t)
	tracker := track.NewPostgresTracker(db)
	NewTestStorage(t, db, tracker)
	require.NoError(t, ReduceObjectsPerChunk(ctx, db))
}
