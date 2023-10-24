package pfsdb_test

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil/random"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	testRepoName = "testRepoName"
	testRepoType = "user"
	testRepoDesc = "this is a test repo"
)

func testRepo(name, repoType string) *pfs.RepoInfo {
	repo := &pfs.Repo{Name: name, Type: repoType, Project: &pfs.Project{Name: pfs.DefaultProjectName}}
	return &pfs.RepoInfo{
		Repo:        repo,
		Description: testRepoDesc,
	}
}

func compareRepos(expected, got *pfs.RepoInfo) bool {
	return expected.Repo.Name == got.Repo.Name &&
		expected.Repo.Type == got.Repo.Type &&
		expected.Repo.Project.Name == got.Repo.Project.Name &&
		expected.Description == got.Description &&
		len(expected.Branches) == len(got.Branches)
}

func newTestDB(t testing.TB, ctx context.Context) *pachsql.DB {
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	return db
}

func TestUpsertRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	expectedInfo := testRepo(testRepoName, testRepoType)
	var repoID pfsdb.RepoID
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		var err error
		repoID, err = pfsdb.UpsertRepo(ctx, tx, expectedInfo)
		require.NoError(t, err)
		getByIDInfo, err := pfsdb.GetRepo(ctx, tx, repoID)
		require.NoError(t, err)
		require.True(t, cmp.Equal(expectedInfo, getByIDInfo, cmp.Comparer(compareRepos)))
		getByNameInfo, err := pfsdb.GetRepoByName(ctx, tx, expectedInfo.Repo.Project.Name, expectedInfo.Repo.Name, expectedInfo.Repo.Type)
		require.NoError(t, err)
		require.True(t, cmp.Equal(expectedInfo, getByNameInfo, cmp.Comparer(compareRepos)))
	})
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		expectedInfo.Description = "new desc"
		id, err := pfsdb.UpsertRepo(ctx, tx, expectedInfo)
		require.NoError(t, err)
		require.Equal(t, repoID, id, "UpsertRepo should keep id stable")
		getInfo, err := pfsdb.GetRepo(ctx, tx, id)
		require.NoError(t, err)
		require.True(t, cmp.Equal(expectedInfo, getInfo, cmp.Comparer(compareRepos)))
	})
}

func TestDeleteRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	expectedInfo := testRepo(testRepoName, testRepoType)
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		id, err := pfsdb.UpsertRepo(ctx, tx, expectedInfo)
		require.NoError(t, err)
		require.NoError(t, pfsdb.DeleteRepo(ctx, tx, expectedInfo.Repo.Project.Name, expectedInfo.Repo.Name, expectedInfo.Repo.Type), "should be able to delete repo")
		_, err = pfsdb.GetRepo(ctx, tx, id)
		require.ErrorIs(t, err, pfsdb.ErrRepoNotFound{ID: id})
		require.ErrorIs(t,
			pfsdb.DeleteRepo(ctx, tx, expectedInfo.Repo.Project.Name, expectedInfo.Repo.Name, expectedInfo.Repo.Type),
			pfsdb.ErrRepoNotFound{Project: expectedInfo.Repo.Project.Name, Name: testRepoName, Type: expectedInfo.Repo.Type},
		)
	})
}

func createCommitAndBranches(ctx context.Context, tx *pachsql.Tx, t *testing.T, repoInfo *pfs.RepoInfo) {
	for _, branch := range repoInfo.Branches {
		commit := &pfs.Commit{Repo: repoInfo.Repo, Branch: nil, Id: random.String(32)}
		commitInfo := &pfs.CommitInfo{Commit: commit,
			Origin:  &pfs.CommitOrigin{Kind: pfs.OriginKind_USER},
			Started: timestamppb.Now()}
		_, err := pfsdb.UpsertCommit(ctx, tx, commitInfo)
		require.NoError(t, err, "should be able to create commit")
		branchInfo := &pfs.BranchInfo{Branch: branch, Head: commit}
		_, err = pfsdb.UpsertBranch(ctx, tx, branchInfo)
		require.NoError(t, err, "should be able to create branch")
		commitInfo.Commit.Branch = branch
		_, err = pfsdb.UpsertCommit(ctx, tx, commitInfo)
		require.NoError(t, err, "should be able to update commit")
	}
}

func TestGetRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	createInfo := testRepo(testRepoName, testRepoType)
	createInfo.Branches = []*pfs.Branch{
		{Repo: createInfo.Repo, Name: "master"},
		{Repo: createInfo.Repo, Name: "a"},
		{Repo: createInfo.Repo, Name: "b"},
	}
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		repoID, err := pfsdb.UpsertRepo(ctx, tx, createInfo)
		require.NoError(t, err, "should be able to create repo")
		createCommitAndBranches(ctx, tx, t, createInfo)
		getInfo, err := pfsdb.GetRepo(ctx, tx, repoID)
		require.NoError(t, err, "should be able to get a repo")
		require.True(t, cmp.Equal(createInfo, getInfo, cmp.Comparer(compareRepos)))
		_, err = pfsdb.GetRepo(ctx, tx, 3)
		require.ErrorIs(t, err, pfsdb.ErrRepoNotFound{ID: 3})
	})
}

func TestListRepos(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		size := 210
		expectedInfos := make([]*pfs.RepoInfo, size)
		for i := 0; i < size; i++ {
			createInfo := testRepo(fmt.Sprintf("%s%d", testRepoName, i), "unknown")
			id, err := pfsdb.UpsertRepo(ctx, tx, createInfo)
			require.NoError(t, err, "should be able to create repo")
			require.Equal(t, pfsdb.RepoID(i+1), id, "id should be auto incremented")
			createInfo.Branches = []*pfs.Branch{
				{Repo: createInfo.Repo, Name: "master"},
				{Repo: createInfo.Repo, Name: "a"},
				{Repo: createInfo.Repo, Name: "b"},
			}
			createCommitAndBranches(ctx, tx, t, createInfo)
			expectedInfos[i] = createInfo
		}
		iter, err := pfsdb.ListRepo(ctx, tx, nil)
		require.NoError(t, err, "should be able to list repos")
		var i int
		require.NoError(t, stream.ForEach[pfsdb.RepoPair](ctx, iter, func(repoPair pfsdb.RepoPair) error {
			require.True(t, cmp.Equal(expectedInfos[i], repoPair.RepoInfo, cmp.Comparer(compareRepos)))
			i++
			return nil
		}))
		require.Equal(t, size, i)
	})
}

func TestListReposFilter(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		for _, repoName := range []string{"repoA", "repoB", "repoC"} {
			for _, repoType := range []string{"user", "unknown", "meta"} {
				createInfo := testRepo(repoName, repoType)
				_, err := pfsdb.UpsertRepo(ctx, tx, createInfo)
				require.NoError(t, err, "should be able to create repo")
			}
		}
		filter := make(pfsdb.RepoListFilter)
		filter[pfsdb.RepoNames] = []string{"repoA"}
		iter, err := pfsdb.ListRepo(ctx, tx, filter)
		i := 0
		require.NoError(t, err, "should be able to list repos")
		require.NoError(t, stream.ForEach[pfsdb.RepoPair](ctx, iter, func(repoPair pfsdb.RepoPair) error {
			if err != nil {
				require.NoError(t, err, "should be able to iterate over repos")
			}
			require.Equal(t, "repoA", repoPair.RepoInfo.Repo.Name)
			i++
			return nil
		}))
		require.Equal(t, 3, i)
		i = 0
		filter[pfsdb.RepoTypes] = []string{testRepoType, "meta"}
		filter[pfsdb.RepoProjects] = []string{pfs.DefaultProjectName, "random"}
		iter, err = pfsdb.ListRepo(ctx, tx, filter)
		seen := make(map[string]bool)
		require.NoError(t, err, "should be able to list repos")
		require.NoError(t, stream.ForEach[pfsdb.RepoPair](ctx, iter, func(repoPair pfsdb.RepoPair) error {
			if err != nil {
				require.NoError(t, err, "should be able to iterate over repos")
			}
			require.Equal(t, "repoA", repoPair.RepoInfo.Repo.Name)
			seen[repoPair.RepoInfo.Repo.Type] = true
			i++
			return nil
		}))
		require.Equal(t, 2, i)
		require.Equal(t, true, seen[testRepoType])
		require.Equal(t, true, seen["meta"])
		require.Equal(t, 2, len(seen))
	})
}
