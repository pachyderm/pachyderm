package clusterstate

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	v2_8_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.8.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
)

func commitsMatch(t *testing.T, a, b *pfs.CommitInfo) {
	require.Equal(t, a.Commit.Repo.Name, b.Commit.Repo.Name)
	require.Equal(t, a.Commit.Id, b.Commit.Id)
	if a.Commit.Branch != nil || b.Commit.Branch != nil {
		require.Equal(t, a.Commit.Branch.Name, b.Commit.Branch.Name)
	}
	require.Equal(t, a.Origin.Kind, b.Origin.Kind)
	require.Equal(t, a.Description, b.Description)
	require.Equal(t, a.Started.Seconds, b.Started.Seconds)
	if a.ParentCommit != nil || b.ParentCommit != nil {
		require.Equal(t, a.ParentCommit.Id, b.ParentCommit.Id)
		require.Equal(t, a.ParentCommit.Repo.Name, b.ParentCommit.Repo.Name)
	}
	require.Equal(t, len(a.ChildCommits), len(b.ChildCommits))
	if len(a.ChildCommits) != 0 || len(b.ChildCommits) != 0 {
		childMap := make(map[string]*pfs.Commit)
		for _, commit := range a.ChildCommits {
			childMap[pfsdb.CommitKey(commit)] = commit
		}
		for _, commit := range b.ChildCommits {
			require.Equal(t, commit.Id, childMap[pfsdb.CommitKey(commit)].Id)
			require.Equal(t, commit.Repo.Name, childMap[pfsdb.CommitKey(commit)].Repo.Name)
		}
	}
}

func Test_v2_8_0_ClusterState(t *testing.T) {
	ctx := pctx.TestContext(t)
	db, _ := dockertestenv.NewEphemeralPostgresDB(ctx, t)
	defer db.Close()
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}

	// Pre-migration
	// Note that we are applying 2.6 migration here because we need to create collections.repos table
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_6_0))
	setupTestData(t, ctx, db)

	// Apply migrations up to and including 2.8.0
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_8_0))
	require.NoError(t, migrations.BlockUntil(ctx, db, state_2_8_0))

	// Get all collections commits.
	commitsCol, err := v2_8_0.ListCommitsFromCollection(ctx, db)
	require.NoError(t, err, "should be able to list commits from collection")
	// Get all pfs.commits.
	iter, err := pfsdb.ListCommit(ctx, db, nil, false)
	require.NoError(t, err, "should be able to list commits from pfs.commits")
	commitCmp := make(map[string]*pfs.CommitInfo)
	require.NoError(t, stream.ForEach[pfsdb.CommitPair](ctx, iter, func(commitPair pfsdb.CommitPair) error {
		commitCmp[pfsdb.CommitKey(commitPair.CommitInfo.Commit)] = commitPair.CommitInfo
		return nil
	}))
	// compare collections commits to pfs.commits.
	require.Equal(t, len(commitsCol), len(commitCmp), "all rows should have been migrated")
	for _, info := range commitsCol {
		commitsMatch(t, info, commitCmp[pfsdb.CommitKey(info.Commit)])
	}

	// Get all existing data from collections.
	// Note that we convert the proto object to a model that conforms better to our new relational schema,
	// but the value is from the collections tables not the new relational tables.
	expectedRepos, err := v2_8_0.ListReposFromCollection(ctx, db)
	require.NoError(t, err)
	expectedBranches, expectedEdges, expectedTriggers, err := v2_8_0.ListBranchesEdgesTriggersFromCollections(ctx, db)
	require.NoError(t, err)

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

	// Verify Branches
	// Check whether all the data is migrated to pfs.branches table
	var gotBranches []*v2_8_0.Branch
	require.NoError(t, db.SelectContext(ctx, &gotBranches, `
		SELECT branch.id, branch.name, branch.head, repo.id as repo_id, branch.created_at, branch.updated_at
		FROM pfs.branches branch JOIN pfs.repos repo ON  repo.id = branch.repo_id
			JOIN core.projects project ON project.id = repo.project_id
		ORDER BY id`))
	require.Equal(t, len(expectedBranches), len(gotBranches))
	if diff := cmp.Diff(expectedBranches, gotBranches); diff != "" {
		t.Errorf("branches differ: (-want +got)\n%s", diff)
	}
	// Check whether all provenance data is migrated to pfs.branch_provenance table
	var gotEdges []*v2_8_0.Edge
	require.NoError(t, db.SelectContext(ctx, &gotEdges, `SELECT from_id, to_id FROM pfs.branch_provenance ORDER BY from_id, to_id`))
	require.Equal(t, len(expectedEdges), len(gotEdges))
	sort.Slice(expectedEdges, func(i, j int) bool {
		if expectedEdges[i].FromID == expectedEdges[j].FromID {
			return expectedEdges[i].ToID < expectedEdges[j].ToID
		}
		return expectedEdges[i].FromID < expectedEdges[j].FromID
	})
	if diff := cmp.Diff(expectedEdges, gotEdges); diff != "" {
		t.Errorf("edges differ: (-want +got)\n%s", diff)
	}

	// Verify triggers
	var gotTriggers []*v2_8_0.BranchTrigger
	require.NoError(t, db.SelectContext(ctx, &gotTriggers, `SELECT from_branch_id, to_branch_id, cron_spec, rate_limit_spec, size, num_commits, all_conditions FROM pfs.branch_triggers ORDER BY from_branch_id, to_branch_id`))
	require.Equal(t, len(expectedTriggers), len(gotTriggers))
	if diff := cmp.Diff(expectedTriggers, gotTriggers); diff != "" {
		t.Errorf("triggers differ: (-want +got)\n%s", diff)
	}
}
