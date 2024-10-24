package restart

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/admindb"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"go.uber.org/atomic"
)

func TestLoop(t *testing.T) {
	var restart atomic.Bool
	var reason, gotReason atomic.String
	r := &Restarter{
		ch: make(listenerChan),
		check: func(_ context.Context) (bool, string, error) {
			return restart.Load(), reason.Load(), nil
		},
		restart: func(_ context.Context, reason string) {
			gotReason.Store(reason)
		},
	}
	ctx := pctx.TestContext(t)
	doneCh := make(chan error)
	go func() {
		doneCh <- r.RestartWhenRequired(ctx)
		close(doneCh)
	}()

	// Send event that doesn't yield a restart.
	r.ch <- struct{}{}
	select {
	case err := <-doneCh:
		t.Fatalf("finished too early: %v", err)
	default:
	}

	// Send event that should cause a restart.
	want := "requested"
	reason.Store(want)
	restart.Store(true)
	r.ch <- struct{}{}
	<-doneCh

	// Ensure that the restart function was called with the reason.
	if got := gotReason.Load(); got != want {
		t.Errorf("reason:\n  got: %v\n want: %v", got, want)
	}
}

func TestRestart(t *testing.T) {
	ctx := pctx.TestContext(t)
	cfg := dockertestenv.NewTestDBConfig(t)
	db := testutil.OpenDB(t, cfg.PGBouncer.DBOptions()...)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	dsn := dbutil.GetDSN(ctx, cfg.Direct.DBOptions()...)
	listener := collection.NewPostgresListener(dsn)

	doneCh := make(chan string)
	now := time.Now().Truncate(time.Microsecond)
	r, err := New(ctx, db, listener)
	if err != nil {
		t.Fatalf("setup restarter: %v", err)
	}
	if math.Abs(now.Sub(r.startupTime).Seconds()) > 30 {
		t.Errorf("too much clock skew between host and postgres:\n      now: %v\n postgres: %v", now.Format(time.RFC3339Nano), r.startupTime.Format(time.RFC3339Nano))
	}
	r.restart = func(_ context.Context, reason string) {
		doneCh <- reason
	}
	go r.RestartWhenRequired(ctx) //nolint:errcheck

	// Schedule a restart that doesn't apply.
	if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		return errors.Wrap(admindb.ScheduleRestart(ctx, tx, time.Now().Add(-time.Hour), "too old", ""), "ScheduleRestart")
	}); err != nil {
		t.Fatalf("WithTx: %v", err)
	}
	select {
	case msg := <-doneCh:
		t.Fatalf("restarted too early: %v", msg)
	case <-time.After(200 * time.Millisecond):
	}

	// Schedule a restart that does apply.
	if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		return errors.Wrap(admindb.ScheduleRestart(ctx, tx, time.Now().Add(time.Hour), "very new", ""), "ScheduleRestart")
	}); err != nil {
		t.Fatalf("WithTx: %v", err)
	}
	var got string
	select {
	case <-time.After(200 * time.Millisecond):
		t.Errorf("didn't get notification after 200ms")
	case got = <-doneCh:
	}
	if want := "very new"; got != want {
		t.Errorf("reason:\n  got: %v\n want: %v", got, want)
	}
}
