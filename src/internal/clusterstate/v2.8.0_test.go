package clusterstate

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	v2_8_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.8.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
)

func Test_v2_8_0_ClusterState(t *testing.T) {
	ctx := pctx.TestContext(t)
	db, _ := dockertestenv.NewEphemeralPostgresDB(ctx, t)
	defer db.Close()
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}

	// Pre-migration
	// Note that we are applying 2.6 migration here because we need to create collections.repos table
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_6_0))
	setupTestData(t, ctx, db)

	// Get all existing repos in collections.repos
	expectedRepos, err := v2_8_0.ListReposFromCollection(ctx, db)
	require.NoError(t, err)

	// Apply migration of interest, which will apply all migrations > 2.6.0
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, State_2_8_0))
	require.NoError(t, migrations.BlockUntil(ctx, db, State_2_8_0))

	// Verify Repos
	// Check whether all the data is migrated to pfs.repos table
	var gotRepos []v2_8_0.Repo
	require.NoError(t, db.SelectContext(ctx, &gotRepos, `
		SELECT repos.id, repos.name, repos.description, repos.type, projects.name as project_name, repos.created_at, repos.updated_at
		FROM pfs.repos repos JOIN core.projects projects ON repos.project_id = projects.id
		ORDER BY id`))
	require.Equal(t, len(expectedRepos), len(gotRepos))
	if diff := cmp.Diff(expectedRepos, gotRepos); diff != "" {
		t.Errorf("repos differ: (-want +got)\n%s", diff)
	}
}
