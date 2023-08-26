package pfsdb_test

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil/random"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestCreateCommit(t *testing.T) {
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
	repoInfo := testRepo(testRepoName, testRepoType)
	repoInfo.Branches = []*pfs.Branch{
		{Repo: repoInfo.Repo, Name: "master"},
	}
	var commitInfo *pfs.CommitInfo
	id := random.String(32)
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		commit := &pfs.Commit{Repo: repoInfo.Repo, Branch: repoInfo.Branches[0], Id: id}
		branchInfo := &pfs.BranchInfo{Branch: repoInfo.Branches[0], Head: commit}
		require.NoError(t, branchesCol.ReadWrite(tx).Put(repoInfo.Branches[0], branchInfo), "should be able to create branches")
		require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, repoInfo), "should be able to create repo")
		commitInfo = &pfs.CommitInfo{
			Commit:      commit,
			Description: "fake commit",
			Origin: &pfs.CommitOrigin{
				Kind: pfs.OriginKind_USER,
			},
			Started: timestamppb.New(time.Now()),
		}
		require.NoError(t, pfsdb.CreateCommit(ctx, tx, commitInfo), "should be able to create commit")
		getInfo, err := pfsdb.GetCommit(ctx, tx, 1)
		require.NoError(t, err)
		require.Equal(t, pfsdb.CommitKey(commitInfo.Commit), getInfo.Commit.Id)
		require.Equal(t, commitInfo.Description, getInfo.Description)
		require.Equal(t, commitInfo.Origin.Kind, getInfo.Origin.Kind)
		require.Equal(t, commitInfo.Started.Seconds, getInfo.Started.Seconds)
		return nil
	}))
	require.YesError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		err := pfsdb.CreateCommit(cbCtx, tx, commitInfo)
		require.YesError(t, err, "should not be able to create commit again with same commit set ID")
		require.True(t, errors.Is(pfsdb.ErrCommitAlreadyExists{CommitID: id, Repo: pfsdb.RepoKey(repoInfo.Repo)}, err))
		return nil
	}), "double create should fail and result in rollback")
}

func TestGetCommit(t *testing.T) {
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
		require.NoError(t, pfsdb.CreateRepo(cbCtx, tx, createInfo), "should be able to create commit")
		getInfo, err := pfsdb.GetRepo(cbCtx, tx, 1)
		require.NoError(t, err, "should be able to get a repo")
		require.Equal(t, createInfo.Repo.Name, getInfo.Repo.Name)
		require.Equal(t, createInfo.Description, getInfo.Description)
		require.Equal(t, len(createInfo.Branches), len(getInfo.Branches))
		_, err = pfsdb.GetRepo(cbCtx, tx, 3)
		require.YesError(t, err, "should not be able to get non-existent commit")
		require.True(t, errors.Is(pfsdb.ErrCommitNotFound{ID: 3}, err))
		return nil
	}))
}
