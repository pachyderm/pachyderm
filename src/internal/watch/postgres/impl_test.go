package postgres_test

import (
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
	"github.com/pachyderm/pachyderm/v2/src/internal/watch/postgres"
)

func TestWatcRepos(t *testing.T) {
	ctx := pctx.TestContext(t)
	dbOpts := dockertestenv.NewTestDirectDBOptions(t)
	db := testutil.OpenDB(t, dbOpts...)
	defer db.Close()

	// Apply migrations
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")

	dsn := dbutil.GetDSN(dbOpts...)
	listener := collection.NewPostgresListener(dsn)
	watcher, err := postgres.NewWatcher(db, listener, "test", "pfs.repos")
	require.NoError(t, err)
	defer watcher.Close()

	// start watching for events
	go func() {
		var results []*postgres.Event
		for event := range watcher.Watch() {
			results = append(results, event)
		}
		require.Len(t, results, 10)
	}()

	// generate events
	// get default project id
	var projectID uint64
	require.NoError(t, db.QueryRowxContext(ctx, `SELECT id FROM core.projects WHERE name = 'default'`).Scan(&projectID))
	for i := 0; i < 10; i++ {
		_, err := db.ExecContext(ctx, "INSERT INTO pfs.repos (name, type, project_id) VALUES ($1, $2, $3)", fmt.Sprintf("repo%d", i), "user", projectID)
		require.NoError(t, err)
	}
}
