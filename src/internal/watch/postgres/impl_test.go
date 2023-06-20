package postgres_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch/postgres"
)

func TestWatcRepos(t *testing.T) {
	// ctx := pctx.TestContext(t)
	dbOpts := dockertestenv.NewTestDBOptions(t)
	db := testutil.OpenDB(t, dbOpts...)
	dsn := dbutil.GetDSN(dbOpts...)
	listener := collection.NewPostgresListener(dsn)
	watcher, err := postgres.NewWatcher(db, listener, "test", "pfs.repos")
	require.NoError(t, err)
	defer watcher.Close()

	// start watching for events
	go func() {
		for event := range watcher.Watch() {
			t.Log(event)
		}
	}()

	// generate events
}
