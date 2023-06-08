package clusterstate

import (
	"fmt"
	"testing"
	"time"

	proto "github.com/gogo/protobuf/proto"

	v2_7_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.7.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
)

func Test_v2_7_0_ClusterState_Projects(t *testing.T) {
	type project struct {
		ID          int       `db:"id"`
		Name        string    `db:"name"`
		Description string    `db:"description"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}

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
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("project%d", i)
		projectInfo := v2_7_0.ProjectInfo{Project: &v2_7_0.Project{Name: name}, Description: "test " + name}
		b, err := proto.Marshal(&projectInfo)
		require.NoError(t, err)
		_, err = tx.ExecContext(ctx, `INSERT INTO collections.projects(key, proto) VALUES($1, $2)`, name, b)
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())

	// Get all existing projects in collections.projects including the default project
	expectedProjects := make(map[string]project)
	collectionRecords := []v2_7_0.CollectionRecord{}
	require.NoError(t, db.SelectContext(ctx, &collectionRecords, `SELECT key, proto, createdat, updatedat FROM collections.projects`))
	for _, row := range collectionRecords {
		projectInfo := v2_7_0.ProjectInfo{}
		require.NoError(t, proto.Unmarshal(row.Proto, &projectInfo))
		expectedProjects[projectInfo.Project.Name] = project{Name: projectInfo.Project.Name, Description: projectInfo.Description, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
	}

	// Apply the migration of interest
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_7_0))
	require.NoError(t, migrations.BlockUntil(ctx, db, state_2_7_0))

	// Check whether all the data is migrated to core.gotProjects table
	gotProjects := []project{}
	require.NoError(t, db.SelectContext(ctx, &gotProjects, `SELECT id, name, description, created_at, updated_at FROM core.projects ORDER BY id`))
	require.Equal(t, len(expectedProjects), len(gotProjects))
	for i, p := range gotProjects {
		require.Equal(t, i+1, p.ID)
		require.Equal(t, expectedProjects[p.Name].Name, p.Name)
		require.Equal(t, expectedProjects[p.Name].Description, p.Description)
		require.Equal(t, expectedProjects[p.Name].CreatedAt, p.CreatedAt)
		require.Equal(t, expectedProjects[p.Name].UpdatedAt, p.UpdatedAt)
	}
}
