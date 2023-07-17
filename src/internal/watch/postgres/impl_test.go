package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
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
	ctx, cancel := context.WithCancel(pctx.TestContext(t))
	defer cancel()
	dbOpts := dockertestenv.NewTestDirectDBOptions(t)
	db := testutil.OpenDB(t, dbOpts...)
	defer db.Close()

	// Apply migrations
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")

	// Create a watcher. The Watcher interfaces with the listener and already starts buffering events.
	dsn := dbutil.GetDSN(dbOpts...)
	listener := collection.NewPostgresListener(dsn)
	watcher, err := postgresWatcher.NewWatcher(ctx, db, listener, t.Name(), "pfs.repos", postgresWatcher.WithBufferSize(10))
	require.NoError(t, err)

	// Generate events by creating repos in the default project.
	var projectID uint64
	require.NoError(t, db.QueryRowxContext(ctx, `SELECT id FROM core.projects WHERE name = 'default'`).Scan(&projectID))
	for i := 0; i < 10; i++ {
		_, err := db.ExecContext(ctx, "INSERT INTO pfs.repos (name, type, project_id) VALUES ($1, $2, $3)", fmt.Sprintf("repo%d", i), "user", projectID)
		require.NoError(t, err)
	}

	// Start watching for events.
	// Note the reason we can start watching *after* we insert the rows is because the watcher is buffering events.
	var results []*postgresWatcher.Event
	events := watcher.Watch()
	for i := 0; i < 10; i++ {
		event := <-events
		require.Equal(t, postgresWatcher.EventInsert, event.EventType)
		require.Equal(t, uint64(i+1), event.Id)
		results = append(results, event)
		if i == 9 {
			cancel()
		}
	}
	require.Len(t, results, 10)
}
