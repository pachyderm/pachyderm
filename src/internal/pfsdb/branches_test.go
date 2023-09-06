package pfsdb_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

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
)

func compareBranches(expected, got *pfs.BranchInfo) bool {
	return expected.Branch.Name == got.Branch.Name &&
		expected.Branch.Repo.Name == got.Branch.Repo.Name &&
		expected.Branch.Repo.Type == got.Branch.Repo.Type &&
		expected.Branch.Repo.Project.Name == got.Branch.Repo.Project.Name &&
		expected.Head.Id == got.Head.Id
}

func TestCreateAndGetBranch(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	options := dockertestenv.NewTestDBOptions(t)
	dsn := dbutil.GetDSN(options...)
	listener := collection.NewPostgresListener(dsn)
	db, err := dbutil.NewDB(options...)
	require.NoError(t, err)
	commitsCol := pfsdb.Commits(db, listener)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")

	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		// Create repo, commit, and branch, then try to fetch that branch
		repoInfo := testRepo(testRepoName, testRepoType)
		if err := pfsdb.CreateRepo(cbCtx, tx, repoInfo); err != nil {
			return err
		}

		commit1Info := &pfs.CommitInfo{Commit: &pfs.Commit{Repo: repoInfo.Repo, Id: random.String(32)}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}}
		commit2Info := &pfs.CommitInfo{Commit: &pfs.Commit{Repo: repoInfo.Repo, Id: random.String(32)}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}, ParentCommit: commit1Info.Commit}
		for _, commitInfo := range []*pfs.CommitInfo{commit1Info, commit2Info} {
			if err := commitsCol.ReadWrite(tx).Put(commitInfo.Commit, commitInfo); err != nil {
				return err
			}
			if err := pfsdb.AddCommit(tx, commitInfo.Commit); err != nil {
				return err
			}
		}

		branchInfo := &pfs.BranchInfo{
			Branch: &pfs.Branch{
				Name: "master",
				Repo: &pfs.Repo{
					Name:    repoInfo.Repo.Name,
					Type:    repoInfo.Repo.Type,
					Project: &pfs.Project{Name: repoInfo.Repo.Project.Name},
				}},
			Head: commit1Info.Commit,
		}
		id, err := pfsdb.UpsertBranch(cbCtx, tx, branchInfo)
		if err != nil {
			return err
		}
		gotBranch, err := pfsdb.GetBranch(cbCtx, tx, id)
		if err != nil {
			return err
		}
		if !cmp.Equal(branchInfo, gotBranch, cmp.Comparer(compareBranches)) {
			return errors.Errorf("expected branch %+v, got %+v", branchInfo, gotBranch)
		}

		// Update branch to point to second commit
		branchInfo.Head = commit2Info.Commit
		id2, err := pfsdb.UpsertBranch(cbCtx, tx, branchInfo)
		if err != nil {
			return err
		}
		if id != id2 {
			return errors.Errorf("expected branch id to be stable: %d != %d", id, id2)
		}
		gotBranch2, err := pfsdb.GetBranch(cbCtx, tx, id2)
		if err != nil {
			return err
		}
		if !cmp.Equal(branchInfo, gotBranch2, cmp.Comparer(compareBranches)) {
			return errors.Errorf("expected branch %+v, got %+v", branchInfo, gotBranch)
		}
		return nil
	}))
}
