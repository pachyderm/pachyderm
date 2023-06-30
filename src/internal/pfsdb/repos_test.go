package pfsdb_test

import (
	"context"
	"fmt"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	testRepo     = "testRepo"
	testRepoDesc = "this is a test repo"
)

func TestCreateRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	createInfo := &pfs.RepoInfo{Repo: &pfs.Repo{Name: testRepo, Type: "user", Project: &pfs.Project{Name: "default"}}, Description: testRepoDesc}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, createInfo), "should be able to create repo")
		getInfo, err := pfsdb.GetRepoByName(cbCtx, tx, testRepo)
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
		require.True(t, pfsdb.ErrRepoAlreadyExists{Project: "default", Name: testRepo}.Is(err))
		fmt.Println("hello")
		return nil
	}), "double create should fail and result in rollback")
}

func TestDeleteRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		createInfo := &pfs.RepoInfo{Repo: &pfs.Repo{Name: testRepo, Project: &pfs.Project{Name: "default"}}, Description: testRepoDesc}
		require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, createInfo), "should be able to create repo")
		require.NoError(t, pfsdb.DeleteRepo(cbCtx, tx, createInfo.Repo.Name), "should be able to delete repo")
		_, err := pfsdb.GetRepoByName(cbCtx, tx, testRepo)
		require.YesError(t, err, "get repo should not find row")
		require.True(t, pfsdb.ErrRepoNotFound{Name: testRepo}.Is(err))
		err = pfsdb.DeleteRepo(cbCtx, tx, createInfo.Repo.Name)
		require.YesError(t, err, "double delete should be an error")
		require.True(t, pfsdb.ErrRepoNotFound{Name: testRepo}.Is(err))
		return nil
	}))
}

func TestDeleteAllRepos(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, migrations.ApplyMigrations(cbCtx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
		size := 3
		for i := 0; i < size; i++ {
			createInfo := &pfs.RepoInfo{Repo: &pfs.Repo{Name: fmt.Sprintf("%s%d", testRepo, i), Project: &pfs.Project{Name: "default"}}, Description: testRepoDesc}
			require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, createInfo), "should be able to create project %d", i)
		}
		require.NoError(t, pfsdb.DeleteAllRepos(cbCtx, tx))
		_, err := pfsdb.GetRepo(cbCtx, tx, 1)
		require.YesError(t, err, "should not have any project entries")
		return nil
	}))
}

func TestGetRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		createInfo := &pfs.RepoInfo{Repo: &pfs.Repo{Name: testRepo, Project: &pfs.Project{Name: "default"}}, Description: testRepoDesc}
		require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, createInfo), "should be able to create repo")
		getInfo, err := pfsdb.GetRepo(cbCtx, tx, 1)
		require.NoError(t, err, "should be able to get a repo")
		require.Equal(t, createInfo.Repo.Name, getInfo.Repo.Name)
		require.Equal(t, createInfo.Description, getInfo.Description)
		_, err = pfsdb.GetRepo(cbCtx, tx, 3)
		require.YesError(t, err, "should not be able to get non-existent repo")
		require.True(t, pfsdb.ErrRepoNotFound{ID: 3}.Is(err))
		return nil
	}))
}

func TestListRepos(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		size := 210
		expectedInfos := make([]*pfs.RepoInfo, size)
		for i := 0; i < size; i++ {
			createInfo := &pfs.RepoInfo{
				Repo: &pfs.Repo{
					Name:    fmt.Sprintf("%s%d", testRepo, i),
					Project: &pfs.Project{Name: "default"},
				},
				Description: testRepoDesc,
			}
			expectedInfos[i] = createInfo
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
			require.Equal(t, "unknown", repo.Repo.Type)
			require.Equal(t, expectedInfos[i].Repo.Project.Name, repo.Repo.Project.Name)
			require.Equal(t, expectedInfos[i].Description, repo.Description)
			i++
			return nil
		}))
		return nil
	}))
}

func TestUpdateRepo(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		// test upsert correctness
		repoInfo := &pfs.RepoInfo{Repo: &pfs.Repo{Name: testRepo, Project: &pfs.Project{Name: "default"}}, Description: testRepoDesc}
		require.YesError(t, pfsdb.UpdateRepo(cbCtx, tx, 99, repoInfo), "should not be able to update repo with an id out of range")
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
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		repoInfo := &pfs.RepoInfo{Repo: &pfs.Repo{Name: testRepo, Project: &pfs.Project{Name: "default"}}, Description: testRepoDesc}
		require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, repoInfo), "should be able to create repo")
		require.NoError(t, pfsdb.UpdateRepo(cbCtx, tx, 1, repoInfo), "should be able to update repo")
		return nil
	}))
}
