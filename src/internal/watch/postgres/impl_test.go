package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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
	watcher, err := postgresWatcher.NewWatcher(db, listener, t.Name(), "pfs.repos", postgresWatcher.WithBufferSize(10))
	require.NoError(t, err)
	defer watcher.Close()

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
	}
	require.Len(t, results, 10)
	watcher.Close()

	// Error handling for when the channel is blocked.
	newWatcher, err := postgresWatcher.NewWatcher(db, listener, t.Name(), "pfs.repos", postgresWatcher.WithBufferSize(0))
	require.NoError(t, err)
	defer newWatcher.Close()

	_, err = db.ExecContext(ctx, "INSERT INTO pfs.repos (name, type, project_id) VALUES ($1, $2, $3)", fmt.Sprintf("repo%d", 10), "user", projectID)
	require.NoError(t, err)
	// Sleep to ensure channel is blocked before we start consuming events.
	time.Sleep(100 * time.Millisecond)

	events = newWatcher.Watch()
	event := <-events
	require.Equal(t, postgresWatcher.EventError, event.EventType)
	require.ErrorContains(t, event.Error, fmt.Sprintf("failed to send event, watcher %s is blocked", t.Name()))
	newWatcher.Close()
}
