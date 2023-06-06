package clusterstate

import (
	"fmt"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

type project struct {
	ID          int       `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func Test_v2_7_0_ClusterState_Projects(t *testing.T) {
	if DesiredClusterState.Number() > state_2_7_0.Number() {
		t.Skip("skipping test because desired state is newer than this migration")
	}
	ctx := pctx.TestContext(t)
	db, _ := dockertestenv.NewEphemeralPostgresDB(ctx, t)
	defer db.Close()
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}

	// Pre-migration
	// Ceate sample test data in collections.projects table
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_6_0))
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()
	projectsCollection := pfsdb.Projects(db, nil).ReadWrite(tx)
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("project%d", i)
		require.NoError(t, projectsCollection.Create(name, &pfs.ProjectInfo{Project: &pfs.Project{Name: name}, Description: "test" + name}))
	}
	// Get all existing projects in collections.projects including the default project
	expectedProjects := make(map[string]project)
	projectInfo := &pfs.ProjectInfo{}
	require.NoError(t, projectsCollection.List(projectInfo, collection.DefaultOptions(), func(string) error {
		p := project{Name: projectInfo.Project.Name, Description: projectInfo.Description}
		require.NoError(t, tx.QueryRowContext(ctx, `SELECT createdat, updatedat FROM collections.projects WHERE key = $1`, projectInfo.Project.Name).Scan(&p.CreatedAt, &p.UpdatedAt))
		expectedProjects[projectInfo.Project.Name] = p
		return nil
	}))
	require.NoError(t, tx.Commit())

	// Apply the migration of interest
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_7_0))
	require.NoError(t, migrations.BlockUntil(ctx, db, state_2_7_0))

	// Check whether all the data is migrated to core.projects table
	rows, err := db.QueryxContext(ctx, `SELECT id, name, description, created_at, updated_at FROM core.projects ORDER BY id`)
	require.NoError(t, err)
	defer rows.Close()
	count := 0
	for rows.Next() {
		var p project
		require.NoError(t, rows.StructScan(&p))
		count++
		require.Equal(t, count, p.ID)
		require.Equal(t, expectedProjects[p.Name].Name, p.Name)
		require.Equal(t, expectedProjects[p.Name].Description, p.Description)
		require.Equal(t, expectedProjects[p.Name].CreatedAt, p.CreatedAt)
		require.Equal(t, expectedProjects[p.Name].UpdatedAt, p.UpdatedAt)
	}
	require.Equal(t, len(expectedProjects), count)
}
