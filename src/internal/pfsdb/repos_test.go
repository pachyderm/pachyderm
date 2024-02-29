package pfsdb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/go-cmp/cmp"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
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

func TestGetRepoByNameMissingProject(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	repo := testRepo(testRepoName, testRepoType)
	repo.Repo.Project.Name = "doesNotExist"
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		_, err := pfsdb.GetRepoByName(ctx, tx, repo.Repo.Project.Name, repo.Repo.Name, repo.Repo.Type)
		require.True(t, errors.As(err, &pfsdb.ProjectNotFoundError{}))
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
		require.True(t, errors.As(err, &pfsdb.RepoNotFoundError{}))
		require.True(t,
			errors.As(pfsdb.DeleteRepo(ctx, tx, expectedInfo.Repo.Project.Name, expectedInfo.Repo.Name, expectedInfo.Repo.Type),
				&pfsdb.RepoNotFoundError{Project: expectedInfo.Repo.Project.Name, Name: testRepoName, Type: expectedInfo.Repo.Type}),
		)
	})
}

func TestDeleteRepoMissingProject(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	expectedInfo := testRepo(testRepoName, testRepoType)
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		err := pfsdb.DeleteRepo(ctx, tx, "doesNotExist", expectedInfo.Repo.Name, expectedInfo.Repo.Type)
		require.True(t, errors.As(err, &pfsdb.ProjectNotFoundError{}))
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
		// validate GetRepo.
		getInfo, err := pfsdb.GetRepo(ctx, tx, repoID)
		require.NoError(t, err, "should be able to get a repo")
		require.True(t, cmp.Equal(createInfo, getInfo, cmp.Comparer(compareRepos)))
		// validate GetRepoInfoWithID.
		getInfoWithID, err := pfsdb.GetRepoInfoWithID(ctx, tx, pfs.DefaultProjectName, testRepoName, testRepoType)
		require.NoError(t, err, "should be able to get a repo")
		require.True(t, cmp.Equal(createInfo, getInfoWithID.RepoInfo, cmp.Comparer(compareRepos)))
		// validate error for attempting to get non-existent repo.
		_, err = pfsdb.GetRepo(ctx, tx, 3)
		require.True(t, errors.As(err, &pfsdb.RepoNotFoundError{}))
	})
}

func TestForEachRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		size := 210
		expectedInfos := make([]*pfs.RepoInfo, size)
		for i := 0; i < size; i++ {
			createInfo := testRepo(fmt.Sprintf("%s%03d", testRepoName, i), "unknown")
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
		i := 0
		require.NoError(t, pfsdb.ForEachRepo(ctx, tx, nil, nil, func(repoWithID pfsdb.RepoInfoWithID) error {
			require.True(t, cmp.Equal(expectedInfos[i], repoWithID.RepoInfo, cmp.Comparer(compareRepos)))
			i++
			return nil
		}))
		require.Equal(t, size, i)
	})
}

func TestForEachRepoFilter(t *testing.T) {
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
		filter := &pfs.Repo{Name: "repoA", Type: "meta", Project: &pfs.Project{Name: "default"}}
		require.NoError(t, pfsdb.ForEachRepo(ctx, tx, filter, nil, func(repoWithID pfsdb.RepoInfoWithID) error {
			require.Equal(t, "repoA", repoWithID.RepoInfo.Repo.Name)
			require.Equal(t, "meta", repoWithID.RepoInfo.Repo.Type)
			return nil
		}), "should be able to call for each repo")
		filter = &pfs.Repo{Name: "repoB", Type: "user", Project: &pfs.Project{Name: "default"}}
		require.NoError(t, pfsdb.ForEachRepo(ctx, tx, filter, nil, func(repoWithID pfsdb.RepoInfoWithID) error {
			require.Equal(t, "repoB", repoWithID.RepoInfo.Repo.Name)
			require.Equal(t, "user", repoWithID.RepoInfo.Repo.Type)
			return nil
		}), "should be able to call for each repo")
	})
}

func TestForEachRepoPage(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	collectRepoPage := func(tx *pachsql.Tx, page *pfs.RepoPage) []pfsdb.RepoInfoWithID {
		repos := make([]pfsdb.RepoInfoWithID, 0)
		require.NoError(t, pfsdb.ForEachRepo(ctx, tx, nil, page, func(repoWithID pfsdb.RepoInfoWithID) error {
			repos = append(repos, repoWithID)
			return nil
		}), "should be able to call for each repo")
		return repos
	}
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		for _, repoName := range []string{"A", "B", "C", "D", "E"} {
			createInfo := testRepo(repoName, "user")
			_, err := pfsdb.UpsertRepo(ctx, tx, createInfo)
			require.NoError(t, err, "should be able to create repo")
		}
		// page larger than number of repos
		rs := collectRepoPage(tx, &pfs.RepoPage{
			PageSize:  20,
			PageIndex: 0,
		})
		require.Equal(t, 5, len(rs))
		rs = collectRepoPage(tx, &pfs.RepoPage{
			PageSize:  3,
			PageIndex: 0,
		})
		assertRepoSequence(t, []string{"A", "B", "C"}, rs)
		rs = collectRepoPage(tx, &pfs.RepoPage{
			PageSize:  3,
			PageIndex: 1,
		})
		assertRepoSequence(t, []string{"D", "E"}, rs)
		// page overbounds
		rs = collectRepoPage(tx, &pfs.RepoPage{
			PageSize:  3,
			PageIndex: 2,
		})
		assertRepoSequence(t, []string{}, rs)
	})
}

// the default order for ForEachRepo is dictated by the page size and is (project.name, repo.name)
func TestForEachRepoDefaultOrder(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		project := &pfs.Project{Name: "alpha"}
		require.NoError(t,
			pfsdb.CreateProject(ctx, tx, &pfs.ProjectInfo{Project: project}),
			"should be able to create repo",
		)
		for i, repoName := range []string{"A", "B", "C"} {
			createInfo := testRepo(repoName, "user")
			if i == 2 {
				createInfo.Repo.Project.Name = "alpha"
			}
			_, err := pfsdb.UpsertRepo(ctx, tx, createInfo)
			require.NoError(t, err, "should be able to create repo")
		}
		repos := make([]pfsdb.RepoInfoWithID, 0)
		require.NoError(t, pfsdb.ForEachRepo(ctx, tx, nil, nil, func(repoWithID pfsdb.RepoInfoWithID) error {
			repos = append(repos, repoWithID)
			return nil
		}), "should be able to call for each repo")
		assertRepoSequence(t, []string{"C", "A", "B"}, repos)
	})
}

func assertRepoSequence(t *testing.T, names []string, repos []pfsdb.RepoInfoWithID) {
	require.Equal(t, len(names), len(repos))
	for i, n := range names {
		require.Equal(t, n, repos[i].RepoInfo.Repo.Name)
	}
}
