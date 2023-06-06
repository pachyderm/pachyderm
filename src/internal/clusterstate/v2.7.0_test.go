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

func Test_v2_7_0_ClusterState(t *testing.T) {
	if DesiredClusterState.Number() > state_2_7_0.Number() {
		t.Skip("skipping test because desired state is newer than this migration")
	}
	ctx := pctx.TestContext(t)
	db, _ := dockertestenv.NewEphemeralPostgresDB(ctx, t)
	defer db.Close()
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}

	// apply the migration
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_7_0))
	require.NoError(t, migrations.BlockUntil(ctx, db, state_2_7_0))

	// make some assertions
	projectNames := []string{"project1", "project2"}
	stmt, err := db.Prepare("INSERT INTO core.projects(name) VALUES($1)")
	require.NoError(t, err)
	for _, name := range projectNames {
		_, err := stmt.ExecContext(ctx, name)
		require.NoError(t, err)
	}

	rowsInserted := 0
	rows, err := db.QueryContext(ctx, `SELECT id, name, created_at, updated_at FROM core.projects ORDER BY id`)
	require.NoError(t, err)
	defer rows.Close()
	var (
		id                     int
		name                   string
		created_at, updated_at time.Time
	)
	for rows.Next() {
		require.NoError(t, rows.Scan(&id, &name, &created_at, &updated_at))
		require.Equal(t, projectNames[id-1], name)
		rowsInserted++
		t.Log(id, name, created_at, updated_at)
	}
	require.Equal(t, 2, rowsInserted)
}
