package pfsdb_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
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
	db := newTestDB(t, ctx)
	commitsCol := pfsdb.Commits(db, nil)

	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		// Create repo, commit, and branch, then try to fetch that branch
		repoInfo := testRepo(testRepoName, testRepoType)
		_, err := pfsdb.UpsertRepo(ctx, tx, repoInfo)
		require.NoError(t, err)

		commit1Info := &pfs.CommitInfo{Commit: &pfs.Commit{Repo: repoInfo.Repo, Id: random.String(32)}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}}
		commit2Info := &pfs.CommitInfo{Commit: &pfs.Commit{Repo: repoInfo.Repo, Id: random.String(32)}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}, ParentCommit: commit1Info.Commit}
		for _, commitInfo := range []*pfs.CommitInfo{commit1Info, commit2Info} {
			require.NoError(t, commitsCol.ReadWrite(tx).Put(commitInfo.Commit, commitInfo))
			require.NoError(t, pfsdb.AddCommit(tx, commitInfo.Commit))
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
		id, err := pfsdb.UpsertBranch(ctx, tx, branchInfo)
		require.NoError(t, err)
		gotBranch, err := pfsdb.GetBranch(ctx, tx, id)
		require.NoError(t, err)
		require.True(t, cmp.Equal(branchInfo, gotBranch, cmp.Comparer(compareBranches)))

		// Update branch to point to second commit
		branchInfo.Head = commit2Info.Commit
		id2, err := pfsdb.UpsertBranch(ctx, tx, branchInfo)
		require.NoError(t, err)
		require.Equal(t, id, id2, "UpsertBranch should keep id stable")
		gotBranch2, err := pfsdb.GetBranch(ctx, tx, id2)
		require.NoError(t, err)
		require.True(t, cmp.Equal(branchInfo, gotBranch2, cmp.Comparer(compareBranches)))
	})
}
