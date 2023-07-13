package clusterstate

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	v2_7_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.7.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// Create sample test data "a la collections", i.e. insert data into collections.* tables
func setupTestData(t *testing.T, ctx context.Context, db *sqlx.DB) {
	t.Helper()
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()
	projectNames := []string{"testProject1", "testProject2", "testProject3", strings.Repeat("A", 51)}
	repoNames := []string{"testRepo1", "testRepo2", "testRepo3"}
	for _, name := range projectNames {
		projectInfo := pfs.ProjectInfo{Project: &pfs.Project{Name: name}, Description: "test project", CreatedAt: timestamppb.Now()}
		b, err := proto.Marshal(&projectInfo)
		require.NoError(t, err)
		_, err = tx.ExecContext(ctx, `INSERT INTO collections.projects(key, proto) VALUES($1, $2)`, projectInfo.Project.String(), b)
		require.NoError(t, err)

		// Create repos for each project.
		for _, repoName := range repoNames {
			repoInfo := pfs.RepoInfo{Repo: &pfs.Repo{Name: repoName, Type: pfs.UserRepoType, Project: projectInfo.Project}, Description: "test repo"}
			b, err = proto.Marshal(&repoInfo)
			require.NoError(t, err)
			_, err = tx.ExecContext(ctx, `INSERT INTO collections.repos(key, proto) VALUES($1, $2)`, repoInfo.Repo.String(), b)
			require.NoError(t, err)
		}
	}
	require.NoError(t, tx.Commit())
}

func Test_v2_7_0_ClusterState(t *testing.T) {
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
	// Get all existing repos in collections.repos
	expectedRepos, err := v2_7_0.ListReposFromCollection(ctx, db)
	require.NoError(t, err)

	// Apply 2.7.0 migration
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_7_0))
	require.NoError(t, migrations.BlockUntil(ctx, db, state_2_7_0))

	// Verify Projects
	// Check whether all the data is migrated to core.projects table
	var gotProjects []v2_7_0.Project
	require.NoError(t, db.SelectContext(ctx, &gotProjects, `SELECT id, name, description, created_at, updated_at FROM core.projects ORDER BY id`))
	require.Equal(t, len(expectedProjects), len(gotProjects))
	if diff := cmp.Diff(expectedProjects, gotProjects, cmp.Comparer(func(t1, t2 time.Time) bool {
		// Ignore sub-microsecond differences becaues postgres stores timestamps with microsecond precision
		return t1.Sub(t2) < time.Microsecond

	})); diff != "" {
		t.Errorf("projects differ: (-want +got)\n%s", diff)
	}
	// Test project names can only be 51 characters long
	_, err = db.ExecContext(ctx, `INSERT INTO core.projects(name, description) VALUES($1, $2)`, strings.Repeat("A", 52), "project name with 52 characters")
	require.ErrorContains(t, err, "value too long for type character varying(51)")

	// Verify Repos
	// Check whether all the data is migrated to pfs.repos table
	var gotRepos []v2_7_0.Repo
	require.NoError(t, db.SelectContext(ctx, &gotRepos, `
		SELECT repos.id, repos.name, repos.description, repos.type, projects.name as project_name, repos.created_at, repos.updated_at
		FROM pfs.repos repos JOIN core.projects projects ON repos.project_id = projects.id
		ORDER BY id`))
	require.Equal(t, len(expectedRepos), len(gotRepos))
	if diff := cmp.Diff(expectedRepos, gotRepos); diff != "" {
		t.Errorf("repos differ: (-want +got)\n%s", diff)
	}
}
