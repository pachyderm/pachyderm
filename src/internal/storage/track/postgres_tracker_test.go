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
			tx.MustExec("CREATE SCHEMA storage")
			tx.MustExec(`
			CREATE TABLE IF NOT EXISTS storage.tracker_objects (
				int_id BIGSERIAL PRIMARY KEY,
				str_id VARCHAR(4096) UNIQUE,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				expires_at TIMESTAMP
			);

			CREATE TABLE IF NOT EXISTS storage.tracker_refs (
				from_id INT8 NOT NULL,
				to_id INT8 NOT NULL,
				PRIMARY KEY (from_id, to_id)
			);

			CREATE INDEX ON storage.tracker_refs (
				to_id,
				from_id
			);`)
			return nil
		})
		require.NoError(t, err)
		return track.NewPostgresTracker(db)
	})
}
