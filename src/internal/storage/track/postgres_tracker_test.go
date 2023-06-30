// TODO: remove _test suffix
// the _test suffix is necessary because
// track -> dockertestenv -> (testutil ->) serviceenv -> client -> renew -> track
package track_test

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

func TestPostgresTracker(t *testing.T) {
	t.Parallel()
	track.TestTracker(t, func(testing.TB) track.Tracker {
		db := dockertestenv.NewTestDB(t)
		ctx := context.Background()
		err := dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
			db.MustExec("CREATE SCHEMA storage")
			return track.SetupPostgresTrackerV0(cbCtx, tx)
		})
		require.NoError(t, err)
		return track.NewPostgresTracker(db)
	})
}
