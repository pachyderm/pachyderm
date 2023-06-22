package clusterstate

import (
	"fmt"
	"strings"
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/google/go-cmp/cmp"

	v2_7_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.7.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func Test_v2_7_0_ClusterState_Projects(t *testing.T) {
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
		projectInfo := pfs.ProjectInfo{Project: &pfs.Project{Name: name}, Description: "test " + name}
		b, err := proto.Marshal(&projectInfo)
		require.NoError(t, err)
		_, err = tx.ExecContext(ctx, `INSERT INTO collections.projects(key, proto) VALUES($1, $2)`, name, b)
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())

	// Get all existing projects in collections.projects including the default project
	var expectedProjects []v2_7_0.Project
	var collectionRecords []v2_7_0.CollectionRecord
	require.NoError(t, db.SelectContext(ctx, &collectionRecords, `SELECT key, proto, createdat, updatedat FROM collections.projects ORDER BY createdat`))
	for i, row := range collectionRecords {
		projectInfo := pfs.ProjectInfo{}
		require.NoError(t, proto.Unmarshal(row.Proto, &projectInfo))
		expectedProjects = append(expectedProjects, v2_7_0.Project{ID: i + 1, Name: projectInfo.Project.Name, Description: projectInfo.Description, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}

	// Migrates collections.projects to core.projects
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_7_0))
	require.NoError(t, migrations.BlockUntil(ctx, db, state_2_7_0))

	// Check whether all the data is migrated to core.projects table
	var gotProjects []v2_7_0.Project
	require.NoError(t, db.SelectContext(ctx, &gotProjects, `SELECT id, name, description, created_at, updated_at FROM core.projects ORDER BY created_at`))
	require.Equal(t, len(expectedProjects), len(gotProjects))
	if diff := cmp.Diff(expectedProjects, gotProjects); diff != "" {
		t.Errorf("projects differ: (-want +got)\n%s", diff)
	}

	// Test project names can only be 51 characters long
	_, err = db.ExecContext(ctx, `INSERT INTO core.projects(name, description) VALUES($1, $2)`, strings.Repeat("A", 52), "this should error")
	require.ErrorContains(t, err, "value too long for type character varying(51)")
}
