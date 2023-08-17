package pfsdb_test

import (
	"context"
	"fmt"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
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
	repo := &pfs.Repo{Name: name, Type: repoType, Project: &pfs.Project{Name: "default"}}
	return &pfs.RepoInfo{
		Repo:        repo,
		Description: testRepoDesc,
	}
}

func TestCreateRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	createInfo := testRepo(testRepoName, testRepoType)
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, createInfo), "should be able to create repo")
		getInfo, err := pfsdb.GetRepoByName(cbCtx, tx, "default", testRepoName, testRepoType)
		require.NoError(t, err, "should be able to get a repo")
		require.Equal(t, createInfo.Repo.Name, getInfo.Repo.Name)
		require.Equal(t, createInfo.Repo.Type, getInfo.Repo.Type)
		require.Equal(t, createInfo.Repo.Project.Name, getInfo.Repo.Project.Name)
		require.Equal(t, createInfo.Description, getInfo.Description)
		return nil
	}))
	require.YesError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		err := pfsdb.CreateRepo(cbCtx, tx, createInfo)
		require.YesError(t, err, "should not be able to create repo again with same name")
		require.True(t, errors.Is(pfsdb.ErrRepoAlreadyExists{Project: "default", Name: testRepoName, Type: testRepoType}, err))
		fmt.Println("hello")
		return nil
	}), "double create should fail and result in rollback")
}

func TestDeleteRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	createInfo := testRepo(testRepoName, testRepoType)
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, createInfo), "should be able to create repo")
		require.NoError(t, pfsdb.DeleteRepo(cbCtx, tx, createInfo.Repo.Project.Name, createInfo.Repo.Name, createInfo.Repo.Type), "should be able to delete repo")
		_, err := pfsdb.GetRepoByName(cbCtx, tx, "default", testRepoName, "unknown")
		require.YesError(t, err, "get repo should not find row")
		require.True(t, errors.Is(pfsdb.ErrRepoNotFound{Project: "default", Name: testRepoName, Type: "unknown"}, err))
		err = pfsdb.DeleteRepo(cbCtx, tx, createInfo.Repo.Project.Name, createInfo.Repo.Name, createInfo.Repo.Type)
		require.YesError(t, err, "double delete should be an error")
		require.True(t, errors.Is(pfsdb.ErrRepoNotFound{Project: createInfo.Repo.Project.Name, Name: testRepoName, Type: createInfo.Repo.Type}, err))
		return nil
	}))
}

func TestGetRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	options := dockertestenv.NewTestDBOptions(t)
	dsn := dbutil.GetDSN(options...)
	listener := collection.NewPostgresListener(dsn)
	db, err := dbutil.NewDB(options...)
	require.NoError(t, err)
	branchesCol := pfsdb.Branches(db, listener)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	createInfo := testRepo(testRepoName, testRepoType)
	createInfo.Branches = []*pfs.Branch{
		{Repo: createInfo.Repo, Name: "master"},
		{Repo: createInfo.Repo, Name: "a"},
		{Repo: createInfo.Repo, Name: "b"},
	}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		for _, branch := range createInfo.Branches {
			commit := &pfs.Commit{Repo: createInfo.Repo, Branch: branch, Id: random.String(32)}
			branchInfo := &pfs.BranchInfo{Branch: branch, Head: commit}
			require.NoError(t, branchesCol.ReadWrite(tx).Put(branch, branchInfo), "should be able to create branches")
		}
		require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, createInfo), "should be able to create repo")
		getInfo, err := pfsdb.GetRepo(cbCtx, tx, 1)
		require.NoError(t, err, "should be able to get a repo")
		require.Equal(t, createInfo.Repo.Name, getInfo.Repo.Name)
		require.Equal(t, createInfo.Description, getInfo.Description)
		require.Equal(t, len(createInfo.Branches), len(getInfo.Branches))
		_, err = pfsdb.GetRepo(cbCtx, tx, 3)
		require.YesError(t, err, "should not be able to get non-existent repo")
		require.True(t, errors.Is(pfsdb.ErrRepoNotFound{ID: 3}, err))
		return nil
	}))
}

func TestListRepos(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	options := dockertestenv.NewTestDBOptions(t)
	dsn := dbutil.GetDSN(options...)
	listener := collection.NewPostgresListener(dsn)
	db, err := dbutil.NewDB(options...)
	require.NoError(t, err)
	branchesCol := pfsdb.Branches(db, listener)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		size := 210
		expectedInfos := make([]*pfs.RepoInfo, size)
		for i := 0; i < size; i++ {
			createInfo := testRepo(fmt.Sprintf("%s%d", testRepoName, i), "unknown")
			createInfo.Branches = []*pfs.Branch{
				{Repo: createInfo.Repo, Name: "master"},
				{Repo: createInfo.Repo, Name: "a"},
				{Repo: createInfo.Repo, Name: "b"},
			}
			expectedInfos[i] = createInfo
			for _, branch := range createInfo.Branches {
				commit := &pfs.Commit{Repo: createInfo.Repo, Branch: branch, Id: random.String(32)}
				branchInfo := &pfs.BranchInfo{Branch: branch, Head: commit}
				require.NoError(t, branchesCol.ReadWrite(tx).Put(branch, branchInfo), "should be able to create branches")
			}
			require.NoError(t, pfsdb.CreateRepo(ctx, tx, createInfo), "should be able to create repo")
		}
		iter, err := pfsdb.ListRepo(cbCtx, tx, nil)
		require.NoError(t, err, "should be able to list repos")
		i := 0
		require.NoError(t, stream.ForEach[pfsdb.RepoPair](cbCtx, iter, func(repoPair pfsdb.RepoPair) error {
			if err != nil {
				require.NoError(t, err, "should be able to iterate over repos")
			}
			require.Equal(t, expectedInfos[i].Repo.Name, repoPair.RepoInfo.Repo.Name)
			require.Equal(t, expectedInfos[i].Repo.Type, repoPair.RepoInfo.Repo.Type)
			require.Equal(t, expectedInfos[i].Repo.Project.Name, repoPair.RepoInfo.Repo.Project.Name)
			require.Equal(t, expectedInfos[i].Description, repoPair.RepoInfo.Description)
			require.Equal(t, len(expectedInfos[i].Branches), len(repoPair.RepoInfo.Branches))
			i++
			return nil
		}))
		require.Equal(t, size, i)
		return nil
	}))
}

func TestListReposFilter(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		for _, repoName := range []string{"repoA", "repoB", "repoC"} {
			for _, repoType := range []string{"user", "unknown", "meta"} {
				createInfo := testRepo(repoName, repoType)
				require.NoError(t, pfsdb.CreateRepo(ctx, tx, createInfo), "should be able to create repo")
			}
		}
		filter := make(pfsdb.RepoListFilter)
		filter[pfsdb.RepoNames] = []string{"repoA"}
		iter, err := pfsdb.ListRepo(cbCtx, tx, filter)
		i := 0
		require.NoError(t, err, "should be able to list repos")
		require.NoError(t, stream.ForEach[pfsdb.RepoPair](cbCtx, iter, func(repoPair pfsdb.RepoPair) error {
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
		filter[pfsdb.RepoProjects] = []string{"default", "random"}
		iter, err = pfsdb.ListRepo(cbCtx, tx, filter)
		seen := make(map[string]bool)
		require.NoError(t, err, "should be able to list repos")
		require.NoError(t, stream.ForEach[pfsdb.RepoPair](cbCtx, iter, func(repoPair pfsdb.RepoPair) error {
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
		return nil
	}))
}

func TestUpdateRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	repoInfo := testRepo(testRepoName, testRepoType)
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		// test upsert correctness
		require.YesError(t, pfsdb.UpdateRepo(cbCtx, tx, 99, repoInfo), "should not be able to update repo with an id out of range")
		// test upsert correctness
		require.NoError(t, pfsdb.UpsertRepo(cbCtx, tx, repoInfo), "should be able to create repo with upsert")
		repoInfo.Description = "new desc"
		require.NoError(t, pfsdb.UpsertRepo(cbCtx, tx, repoInfo), "should be able to update repo with upsert")
		return nil
	}))
}

func TestUpdateRepoByID(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	repoInfo := testRepo(testRepoName, testRepoType)
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, repoInfo), "should be able to create repo")
		require.NoError(t, pfsdb.UpdateRepo(cbCtx, tx, 1, repoInfo), "should be able to update repo")
		return nil
	}))
}
