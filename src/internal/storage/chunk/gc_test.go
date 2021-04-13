package chunk

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestReduceObjectsPerChunk(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)
	tracker := track.NewPostgresTracker(db)
	NewTestStorage(t, db, tracker)
	require.NoError(t, ReduceObjectsPerChunk(ctx, db))
}
