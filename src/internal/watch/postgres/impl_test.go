package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	v2_7_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.7.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	postgresWatcher "github.com/pachyderm/pachyderm/v2/src/internal/watch/postgres"
)

func TestWatchRepos(t *testing.T) {
	dbOpts := dockertestenv.NewTestDirectDBOptions(t)
	db := testutil.OpenDB(t, dbOpts...)
	defer db.Close()

	// Apply migrations
	ctx := pctx.TestContext(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")

	// Create a listener.
	dsn := dbutil.GetDSN(dbOpts...)
	config, err := pgx.ParseConfig(dsn)
	require.NoError(t, err)
	delete(config.RuntimeParams, "statement_cache_mode") // pgx doesn't like this

	// Start listening.
	listenCtx, cancelListener := context.WithCancel(ctx)
	defer cancelListener()
	listener := postgresWatcher.NewListener(listenCtx, config)

	// Start multiple watchers.
	watchCtx, cancelWatchers := context.WithCancel(ctx)
	defer cancelWatchers()
	events1, errs1 := listener.Watch(watchCtx, v2_7_0.ReposPgChannel, 10)
	events2, errs2 := listener.Watch(watchCtx, v2_7_0.ReposPgChannel, 10)
	events3, errs3 := listener.Watch(watchCtx, v2_7_0.ReposPgChannel, 10)
	_, _ = listener.Watch(watchCtx, "projects", 10) // projects channel doesn't exist, but this should still work

	// Generate events by creating repos in the default project.
	var projectID uint64
	require.NoError(t, db.QueryRowxContext(ctx, `SELECT id FROM core.projects WHERE name = 'default'`).Scan(&projectID))
	for i := 0; i < 10; i++ {
		_, err := db.ExecContext(ctx, "INSERT INTO pfs.repos (name, type, project_id) VALUES ($1, $2, $3)", fmt.Sprintf("repo%d", i), "user", projectID)
		require.NoError(t, err)
	}

	// Start watching for events.
	// Note the reason we can start watching *after* we insert the rows is because the watcher is buffering events.
	allEvents := []<-chan *postgresWatcher.Event{events1, events2, events3}
	allErrs := []<-chan error{errs1, errs2, errs3}
	for i, events := range allEvents {
		errs := allErrs[i]
		var results []*postgresWatcher.Event
		for j := 0; j < 10; j++ {
			select {
			case err := <-errs:
				t.Fatal(err)
			case event := <-events:
				require.Equal(t, postgresWatcher.EventInsert, event.EventType)
				require.Equal(t, uint64(j+1), event.Id)
				results = append(results, event)
			}
		}
		require.Len(t, results, 10)
	}
	cancelListener()
	require.ErrorIs(t, context.Canceled, <-listener.Errs())

	// Error handling for when the channel is blocked.
	listenCtx, cancelListener = context.WithCancel(ctx)
	defer cancelListener()
	listener = postgresWatcher.NewListener(listenCtx, config)

	_, watcherErrs := listener.Watch(listenCtx, v2_7_0.ReposPgChannel, 0)
	_, err = db.ExecContext(ctx, "INSERT INTO pfs.repos (name, type, project_id) VALUES ($1, $2, $3)", "repo11", "user", projectID)
	require.NoError(t, err)

	require.ErrorContains(t, <-watcherErrs, "buffer full")

	cancelListener()
	require.ErrorIs(t, context.Canceled, <-listener.Errs())
	// require.ErrorIs(t, context.Canceled, <-watcherErrs)
}
