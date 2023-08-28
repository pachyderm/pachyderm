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
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil/random"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func commitsMatch(t *testing.T, a, b *pfs.CommitInfo) {
	require.Equal(t, a.Commit.Repo.Name, b.Commit.Repo.Name)
	require.Equal(t, a.Commit.Id, b.Commit.Id)
	require.Equal(t, a.Origin.Kind, b.Origin.Kind)
	require.Equal(t, a.Description, b.Description)
	require.Equal(t, a.Started.Seconds, b.Started.Seconds)
}

func testCommit(ctx context.Context, t *testing.T, branchesCol collection.PostgresCollection, db *pachsql.DB, repoName string) *pfs.CommitInfo {
	repoInfo := testRepo(repoName, testRepoType)
	repoInfo.Branches = []*pfs.Branch{
		{Repo: repoInfo.Repo, Name: "master"},
	}
	var commitInfo *pfs.CommitInfo
	id := random.String(32)
	commit := &pfs.Commit{Repo: repoInfo.Repo, Branch: repoInfo.Branches[0], Id: id}
	branchInfo := &pfs.BranchInfo{Branch: repoInfo.Branches[0], Head: commit}
	commitInfo = &pfs.CommitInfo{
		Commit:      commit,
		Description: "fake commit",
		Origin: &pfs.CommitOrigin{
			Kind: pfs.OriginKind_USER,
		},
		Started: timestamppb.New(time.Now()),
	}
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, branchesCol.ReadWrite(tx).Put(repoInfo.Branches[0], branchInfo), "should be able to create branches")
		require.NoError(t, pfsdb.UpsertRepo(cbCtx, tx, repoInfo), "should be able to create repo")
		return nil
	}))
	return commitInfo
}

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
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	commitInfo := testCommit(ctx, t, branchesCol, db, testRepoName)
	repo := commitInfo.Commit.Repo
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, pfsdb.CreateCommit(ctx, tx, commitInfo), "should be able to create commit")
		getInfo, err := pfsdb.GetCommit(ctx, tx, 1)
		require.NoError(t, err)
		commitsMatch(t, commitInfo, getInfo)
		return nil
	}))
	require.YesError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		err := pfsdb.CreateCommit(cbCtx, tx, commitInfo)
		require.YesError(t, err, "should not be able to create commit again with same commit set ID")
		require.True(t, errors.Is(pfsdb.ErrCommitAlreadyExists{CommitID: commitInfo.Commit.Id, Repo: pfsdb.RepoKey(repo)}, err))
		return nil
	}), "double create should fail and result in rollback")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		commitInfo.Commit.Repo = nil
		err := pfsdb.CreateCommit(cbCtx, tx, commitInfo)
		require.YesError(t, err, "should not be able to create commit when repo is nil")
		require.True(t, errors.Is(pfsdb.ErrCommitMissingInfo{Field: "Repo"}, err))
		commitInfo.Commit.Repo = repo
		tmpOrigin := commitInfo.Origin
		commitInfo.Origin = nil
		err = pfsdb.CreateCommit(cbCtx, tx, commitInfo)
		require.YesError(t, err, "should not be able to create commit origin is nil")
		require.True(t, errors.Is(pfsdb.ErrCommitMissingInfo{Field: "Origin"}, err))
		commitInfo.Origin = tmpOrigin
		commitInfo.Commit = nil
		err = pfsdb.CreateCommit(cbCtx, tx, commitInfo)
		require.YesError(t, err, "should not be able to create commit origin is nil")
		require.True(t, errors.Is(pfsdb.ErrCommitMissingInfo{Field: "Commit"}, err))
		return nil
	}), "transaction should succeed because test is failing before calling db")
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
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	commitInfo := testCommit(ctx, t, branchesCol, db, testRepoName)
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, pfsdb.CreateCommit(ctx, tx, commitInfo), "should be able to create commit")
		getInfo, err := pfsdb.GetCommit(ctx, tx, 1)
		require.NoError(t, err, "should be able to get commit with id=1")
		commitsMatch(t, commitInfo, getInfo)
		getInfo, err = pfsdb.GetCommit(cbCtx, tx, 0)
		require.YesError(t, err, "should not be able to get commit with id=0")
		_, err = pfsdb.GetCommit(cbCtx, tx, 3)
		require.YesError(t, err, "should not be able to get non-existent commit")
		require.True(t, errors.Is(pfsdb.ErrCommitNotFound{ID: 3}, err))
		getInfo, err = pfsdb.GetCommitByCommitKey(cbCtx, tx, commitInfo.Commit)
		commitsMatch(t, commitInfo, getInfo)
		require.NoError(t, err, "should be able to get a commit by key")
		commitInfo.Commit = nil
		_, err = pfsdb.GetCommitByCommitKey(cbCtx, tx, commitInfo.Commit)
		require.YesError(t, err, "should not be able to get commit when commit is nil.")
		require.True(t, errors.Is(pfsdb.ErrCommitMissingInfo{Field: "Commit"}, err))
		return nil
	}))
}

func TestDeleteCommit(t *testing.T) {
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
	commitInfo := testCommit(ctx, t, branchesCol, db, testRepoName)
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, pfsdb.CreateCommit(ctx, tx, commitInfo), "should be able to create commit")
		err = pfsdb.DeleteCommit(cbCtx, tx, commitInfo.Commit)
		require.NoError(t, err, "should be able to delete commit with commit_id=commit.Id")
		_, err := pfsdb.GetCommitByCommitKey(cbCtx, tx, commitInfo.Commit)
		require.YesError(t, err, "should not be able to get a commit")
		require.True(t, errors.Is(pfsdb.ErrCommitNotFound{CommitID: pfsdb.CommitKey(commitInfo.Commit)}, err))
		err = pfsdb.DeleteCommit(cbCtx, tx, commitInfo.Commit)
		require.YesError(t, err, "should not be able to double delete commit")
		require.True(t, errors.Is(pfsdb.ErrCommitNotFound{CommitID: pfsdb.CommitKey(commitInfo.Commit)}, err))
		commitInfo.Commit = nil
		err = pfsdb.DeleteCommit(cbCtx, tx, commitInfo.Commit)
		require.YesError(t, err, "should not be able to double delete commit")
		require.True(t, errors.Is(pfsdb.ErrCommitMissingInfo{Field: "Commit"}, err))
		return nil
	}))
}

func TestListCommits(t *testing.T) {
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
		expectedInfos := make([]*pfs.CommitInfo, size)
		for i := 0; i < size; i++ {
			commitInfo := testCommit(ctx, t, branchesCol, db, testRepoName)
			expectedInfos[i] = commitInfo
			require.NoError(t, pfsdb.CreateCommit(ctx, tx, commitInfo))
		}
		iter, err := pfsdb.ListCommit(cbCtx, tx, nil)
		require.NoError(t, err, "should be able to list repos")
		i := 0
		require.NoError(t, stream.ForEach[pfsdb.CommitPair](cbCtx, iter, func(commitPair pfsdb.CommitPair) error {
			if err != nil {
				require.NoError(t, err, "should be able to iterate over commits")
			}
			commitsMatch(t, expectedInfos[i], commitPair.CommitInfo)
			i++
			return nil
		}))
		require.Equal(t, size, i)
		return nil
	}))
}

func TestListCommitsFilter(t *testing.T) {
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
	repos := []string{"a", "b", "c"}
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		size := 330
		expectedInfos := make([]*pfs.CommitInfo, 0)
		commitSetIds := make([]string, 0)
		for i := 0; i < size; i++ {
			commitInfo := testCommit(ctx, t, branchesCol, db, repos[i%len(repos)])
			if commitInfo.Commit.Repo.Name == "b" && i%10 == 0 {
				expectedInfos = append(expectedInfos, commitInfo)
				commitSetIds = append(commitSetIds, commitInfo.Commit.Id)
			}
			require.NoError(t, pfsdb.CreateCommit(ctx, tx, commitInfo))
		}
		filter := pfsdb.CommitListFilter{
			pfsdb.CommitRepos: []string{"b"},
			//pfsdb.CommitOrigins:  []string{pfs.OriginKind_ORIGIN_KIND_UNKNOWN.String(), pfs.OriginKind_USER.String()},
			//pfsdb.CommitBranches: []string{"master"},
			pfsdb.CommitSetIDs: commitSetIds,
		}
		iter, err := pfsdb.ListCommit(cbCtx, tx, filter)
		require.NoError(t, err, "should be able to list repos")
		i := 0
		require.NoError(t, stream.ForEach[pfsdb.CommitPair](cbCtx, iter, func(commitPair pfsdb.CommitPair) error {
			if err != nil {
				require.NoError(t, err, "should be able to iterate over commits")
			}
			commitsMatch(t, expectedInfos[i], commitPair.CommitInfo)
			i++
			return nil
		}))
		require.Equal(t, len(expectedInfos), i)
		return nil
	}))
}
