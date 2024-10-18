package cmds

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/admindb"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

var startup = time.Now()

func assertPendingRestart(ctx context.Context, t *testing.T, db *pachsql.DB, want bool) {
	t.Helper()
	var got bool
	if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		got, _, err = admindb.ShouldRestart(ctx, tx, startup, time.Now())
		if err != nil {
			return errors.Wrap(err, "ShouldRestart")
		}
		return nil
	}); err != nil {
		t.Fatalf("assertPendingRestart.WithTx: %v", err)
	}
	if got != want {
		t.Errorf("pending restart?\n  got: %v\n want: %v", got, want)
	}
}

func TestRestartPachyderm(t *testing.T) {
	var db *pachsql.DB
	c := pachd.NewTestPachd(t, pachd.TestPachdOption{
		MutateEnv: func(env *pachd.Env) {
			db = env.DB
		},
	})
	ctx := pctx.TestContext(t)
	assertPendingRestart(ctx, t, db, false)
	require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, `pachctl restart cluster`).Run(), "cmd should run")
	assertPendingRestart(ctx, t, db, true)
}
