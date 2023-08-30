package pfsdb_test

import (
	"context"
	"testing"

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

func TestGetBranch(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	options := dockertestenv.NewTestDBOptions(t)
	dsn := dbutil.GetDSN(options...)
	listener := collection.NewPostgresListener(dsn)
	db, err := dbutil.NewDB(options...)
	require.NoError(t, err)
	commitsCol := pfsdb.Commits(db, listener)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	repoInfo := testRepo(testRepoName, testRepoType)
	commitInfo := &pfs.CommitInfo{Commit: &pfs.Commit{Repo: repoInfo.Repo, Id: random.String(32)}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}}
	branchInfo := &pfs.BranchInfo{Head: commitInfo.Commit, Branch: &pfs.Branch{Repo: repoInfo.Repo, Name: "master"}}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		// create repo
		if err := pfsdb.CreateRepo(cbCtx, tx, repoInfo); err != nil {
			return err
		}
		// create commit
		if err := commitsCol.ReadWrite(tx).Put(commitInfo.Commit, commitInfo); err != nil {
			return err
		}
		if err := pfsdb.AddCommit(tx, commitInfo.Commit); err != nil {
			return err
		}

		// create branch
		id, err := pfsdb.CreateBranch(cbCtx, tx, branchInfo)
		if err != nil {
			return err
		}

		// get goBranch
		gotBranch, err := pfsdb.GetBranch(cbCtx, tx, id)
		if err != nil {
			return err
		}
		if gotBranch.Branch.Repo.Project.Name != branchInfo.Branch.Repo.Project.Name ||
			gotBranch.Branch.Repo.Name != branchInfo.Branch.Repo.Name ||
			gotBranch.Branch.Name != branchInfo.Branch.Name {
			return errors.Errorf("expected branch name %+v, got %+v", branchInfo.Branch, gotBranch.Branch)
		}
		return nil
	}))
}
