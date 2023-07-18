package pfsdb_test

import (
	"context"
	"fmt"
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
	createInfo := testRepo(testRepoName, "user")
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, createInfo), "should be able to create repo")
		getInfo, err := pfsdb.GetRepoByName(cbCtx, tx, "default", testRepoName, "user")
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
		require.True(t, pfsdb.ErrRepoAlreadyExists{Project: "default", Name: testRepoName}.Is(err))
		fmt.Println("hello")
		return nil
	}), "double create should fail and result in rollback")
}

func TestDeleteRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	createInfo := testRepo(testRepoName, "")
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, createInfo), "should be able to create repo")
		require.NoError(t, pfsdb.DeleteRepo(cbCtx, tx, createInfo.Repo.Project.Name, createInfo.Repo.Name, createInfo.Repo.Type), "should be able to delete repo")
		_, err := pfsdb.GetRepoByName(cbCtx, tx, "default", testRepoName, "unknown")
		require.YesError(t, err, "get repo should not find row")
		require.True(t, pfsdb.ErrRepoNotFound{Project: "default", Name: testRepoName, Type: "unknown"}.Is(err))
		err = pfsdb.DeleteRepo(cbCtx, tx, createInfo.Repo.Project.Name, createInfo.Repo.Name, createInfo.Repo.Type)
		require.YesError(t, err, "double delete should be an error")
		require.True(t, pfsdb.ErrRepoNotFound{Project: createInfo.Repo.Project.Name, Name: testRepoName, Type: createInfo.Repo.Type}.Is(err))
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
	createInfo := testRepo(testRepoName, "user")
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
		require.True(t, pfsdb.ErrRepoNotFound{ID: 3}.Is(err))
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
		iter, err := pfsdb.ListRepo(cbCtx, tx)
		require.NoError(t, err, "should be able to list repos")
		i := 0
		require.NoError(t, stream.ForEach[*pfs.RepoInfo](cbCtx, iter, func(repo *pfs.RepoInfo) error {
			if err != nil {
				require.NoError(t, err, "should be able to iterate over repos")
			}
			require.Equal(t, expectedInfos[i].Repo.Name, repo.Repo.Name)
			require.Equal(t, expectedInfos[i].Repo.Type, repo.Repo.Type)
			require.Equal(t, expectedInfos[i].Repo.Project.Name, repo.Repo.Project.Name)
			require.Equal(t, expectedInfos[i].Description, repo.Description)
			require.Equal(t, len(expectedInfos[i].Branches), len(repo.Branches))
			i++
			return nil
		}))
		require.Equal(t, size, i)
		return nil
	}))
}

func TestListReposByIdxType(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		size := 210
		expectedInfos := make([]*pfs.RepoInfo, 0)
		for i := 0; i < size; i++ {
			createInfo := testRepo(fmt.Sprintf("%s%d", testRepoName, i), "unknown")
			if i%2 == 0 {
				createInfo.Repo.Type = "user"
				expectedInfos = append(expectedInfos, createInfo)
			}
			require.NoError(t, pfsdb.CreateRepo(ctx, tx, createInfo), "should be able to create repo")
		}
		iter, err := pfsdb.ListRepoByIdxType(cbCtx, tx, "user")
		require.NoError(t, err, "should be able to list repos")
		i := 0
		require.NoError(t, stream.ForEach[*pfs.RepoInfo](cbCtx, iter, func(repo *pfs.RepoInfo) error {
			if err != nil {
				require.NoError(t, err, "should be able to iterate over repos")
			}
			require.Equal(t, expectedInfos[i].Repo.Name, repo.Repo.Name)
			require.Equal(t, expectedInfos[i].Repo.Type, repo.Repo.Type)
			require.Equal(t, expectedInfos[i].Repo.Project.Name, repo.Repo.Project.Name)
			require.Equal(t, expectedInfos[i].Description, repo.Description)
			i++
			return nil
		}))
		require.Equal(t, size/2, i)
		return nil
	}))
}

func TestListReposByIdxName(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		expectedInfos := make([]*pfs.RepoInfo, 0)
		for _, repoName := range []string{"repoA", "repoB", "repoC"} {
			createInfo := testRepo(repoName, "unknown")
			require.NoError(t, pfsdb.CreateRepo(ctx, tx, createInfo), "should be able to create repo")
		}
		for _, repoType := range []string{"user", "unknown", "meta"} {
			createInfo := testRepo(testRepoName, repoType)
			require.NoError(t, pfsdb.CreateRepo(ctx, tx, createInfo), "should be able to create repo")
			expectedInfos = append(expectedInfos, createInfo)
		}
		iter, err := pfsdb.ListRepoByIdxName(cbCtx, tx, testRepoName)
		require.NoError(t, err, "should be able to list repos")
		i := 0
		require.NoError(t, stream.ForEach[*pfs.RepoInfo](cbCtx, iter, func(repo *pfs.RepoInfo) error {
			if err != nil {
				require.NoError(t, err, "should be able to iterate over repos")
			}
			require.Equal(t, expectedInfos[i].Repo.Name, repo.Repo.Name)
			require.Equal(t, expectedInfos[i].Repo.Type, repo.Repo.Type)
			require.Equal(t, expectedInfos[i].Repo.Project.Name, repo.Repo.Project.Name)
			require.Equal(t, expectedInfos[i].Description, repo.Description)
			i++
			return nil
		}))
		require.Equal(t, 3, i)
		return nil
	}))
}

func TestUpdateRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	repoInfo := testRepo(testRepoName, "")
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
	repoInfo := testRepo(testRepoName, "")
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, repoInfo), "should be able to create repo")
		require.NoError(t, pfsdb.UpdateRepo(cbCtx, tx, 1, repoInfo), "should be able to update repo")
		return nil
	}))
}
