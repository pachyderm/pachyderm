package clusterstate

import (
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
)

func TestCurrentClusterState(t *testing.T) {
	ctx := pctx.TestContext(t)
	db, _ := dockertestenv.NewEphemeralPostgresDB(ctx, t)
	defer db.Close()
	etcd := testetcd.NewEnv(ctx, t).EtcdClient
	migrationEnv := migrations.Env{EtcdClient: etcd}

	// apply the migration
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, DesiredClusterState))
	require.NoError(t, migrations.BlockUntil(ctx, db, DesiredClusterState))

	// make some assertions
	projectNames := []string{"project1", "project2"}
	stmt, err := db.Prepare("INSERT INTO core.projects(name) VALUES($1)")
	require.NoError(t, err)
	for _, name := range projectNames {
		_, err := stmt.ExecContext(ctx, name)
		require.NoError(t, err)
	}
	var (
		id                     int
		name                   string
		created_at, updated_at time.Time
	)
	rowsInserted := 0
	rows, err := db.QueryContext(ctx, `SELECT id, name, created_at, updated_at FROM core.projects ORDER BY id`)
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		require.NoError(t, rows.Scan(&id, &name, &created_at, &updated_at))
		require.Equal(t, projectNames[id-1], name)
		rowsInserted++
	}
	require.Equal(t, 2, rowsInserted)
}
