package clusterstate

import (
	"context"
	"fmt"
	"strings"
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"

	v2_7_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.7.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// Create sample test data in collections.projects table
func setupTestData(t *testing.T, ctx context.Context, db *sqlx.DB) {
	t.Helper()
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()
	for i := 0; i < 12; i++ {
		projectInfo := pfs.ProjectInfo{Project: &pfs.Project{Name: fmt.Sprintf("project%d", i+1)}, Description: "test project"}
		b, err := proto.Marshal(&projectInfo)
		require.NoError(t, err)
		_, err = tx.ExecContext(ctx, `INSERT INTO collections.projects(key, proto) VALUES($1, $2)`, projectInfo.Project.String(), b)
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())
}

func Test_v2_7_0_ClusterState_Projects(t *testing.T) {
	ctx := pctx.TestContext(t)
	db, _ := dockertestenv.NewEphemeralPostgresDB(ctx, t)
	defer db.Close()
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}

	// Pre-migration
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_6_0))
	setupTestData(t, ctx, db)

	// Get all existing projects in collections.projects including the default project
	expectedProjects, err := v2_7_0.ListProjectsFromCollection(ctx, db)
	require.NoError(t, err)

	// Migrates collections.projects to core.projects
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_7_0))
	require.NoError(t, migrations.BlockUntil(ctx, db, state_2_7_0))

	// Check whether all the data is migrated to core.projects table
	var gotProjects []v2_7_0.Project
	require.NoError(t, db.SelectContext(ctx, &gotProjects, `SELECT id, name, description, created_at, updated_at FROM core.projects ORDER BY id`))
	require.Equal(t, len(expectedProjects), len(gotProjects))
	if diff := cmp.Diff(expectedProjects, gotProjects); diff != "" {
		t.Errorf("projects differ: (-want +got)\n%s", diff)
	}

	// Test project names can only be 51 characters long
	_, err = db.ExecContext(ctx, `INSERT INTO core.projects(name, description) VALUES($1, $2)`, strings.Repeat("A", 52), "this should error")
	require.ErrorContains(t, err, "value too long for type character varying(51)")
}
