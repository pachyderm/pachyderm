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
	listener := postgresWatcher.NewListener()
	ctx, cancelListener := context.WithCancel(ctx)
	defer cancelListener()
	go func() {
		require.ErrorIs(t, context.Canceled, listener.Start(ctx, config))
	}()

	// Start multiple watchers.
	events1, errs1 := listener.Watch(ctx, v2_7_0.ReposPgChannel)
	events2, errs2 := listener.Watch(ctx, v2_7_0.ReposPgChannel)
	events3, errs3 := listener.Watch(ctx, v2_7_0.ReposPgChannel)

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
	for _, errs := range allErrs {
		require.ErrorIs(t, context.Canceled, <-errs)
	}

	// Error handling for when the channel is blocked.
	// watchCtx, cancel := context.WithCancel(ctx)
	// defer cancel()
	// repoEvents, err := listener.Watch(watchCtx, v2_7_0.ReposPgChannel)
	// require.NoError(t, err)

	// _, err = db.ExecContext(ctx, "INSERT INTO pfs.repos (name, type, project_id) VALUES ($1, $2, $3)", fmt.Sprintf("repo%d", 10), "user", projectID)
	// require.NoError(t, err)
	// // Sleep to ensure channel is blocked before we start consuming events.
	// time.Sleep(100 * time.Millisecond)

	// events = newWatcher.Watch()
	// event := <-events
	// require.Equal(t, postgresWatcher.EventError, event.EventType)
	// require.ErrorContains(t, event.Error, fmt.Sprintf("failed to send event, watcher %s is blocked", t.Name()))
	// newWatcher.Close()
}
