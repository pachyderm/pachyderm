// TODO: remove _test suffix
// the _test suffix is necessary because
// track -> dockertestenv -> (testutil ->) serviceenv -> client -> renew -> track
package track_test

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
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
			return track.SetupPostgresTrackerV0(cbCtx, tx)
		})
		require.NoError(t, err)
		return track.NewPostgresTracker(db)
	})
}

func TestTrackerDump(t *testing.T) {
	db := dockertestenv.NewMigratedTestDB(t, clusterstate.DesiredClusterState)
	tr := track.NewPostgresTracker(db)
	dt, ok := tr.(track.Dumper)
	if !ok {
		t.Fatalf("should be able convert %#v to a track.Dumper", tr)
	}
	ctx := pctx.TestContext(t)
	require.NoError(t, track.Create(ctx, tr, "foo/1", nil, 0), "should create foo/1")
	require.NoError(t, track.Create(ctx, tr, "foo/2", []string{"foo/1"}, 0), "should create foo/2")
	var buf bytes.Buffer
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		return dt.DumpTracker(cbCtx, tx, &buf)
	}))
	want := strings.Join([]string{
		"",
		"",
		"COPY storage.tracker_objects (int_id, str_id, created_at, expires_at) FROM stdin;",
		"1\tfoo/1\tXXX-XXX-XXX XXX:XXX:XXX.XXX\t\\N",
		"2\tfoo/2\tXXX-XXX-XXX XXX:XXX:XXX.XXX\t\\N",
		`\.`,
		"",
		"",
		"COPY storage.tracker_refs (from_id, to_id) FROM stdin;",
		"2\t1",
		`\.`,
		"",
	}, "\n")
	got := buf.String()
	got = regexp.MustCompile(`[0-9]{2,}`).ReplaceAllString(got, "XXX")
	require.NoDiff(t, want, got, nil, "should produce expected dump")
}
