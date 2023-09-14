package pfsdb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil/random"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func compareHead(expected, got *pfs.Commit) bool {
	return expected.Id == got.Id &&
		expected.Repo.Name == got.Repo.Name &&
		expected.Repo.Type == got.Repo.Type &&
		expected.Repo.Project.Name == got.Repo.Project.Name
}

func compareBranch(expected, got *pfs.Branch) bool {
	return expected.Name == got.Name &&
		expected.Repo.Name == got.Repo.Name &&
		expected.Repo.Type == got.Repo.Type &&
		expected.Repo.Project.Name == got.Repo.Project.Name
}

func compareBranchOpts() []cmp.Option {
	return []cmp.Option{
		cmpopts.IgnoreUnexported(pfs.BranchInfo{}),
		cmpopts.SortSlices(func(a, b *pfs.Branch) bool { return a.Key() < b.Key() }), // Note that this is before compareBranch because we need to sort first.
		cmpopts.EquateEmpty(),
		cmp.Comparer(compareBranch),
		cmp.Comparer(compareHead),
	}
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
			// Note that commitsCol implicity creates an entry in pfs.commits table via a PutHook
			// TODO replace this with pfsdb.CreateCommit
			require.NoError(t, commitsCol.ReadWrite(tx).Put(commitInfo.Commit, commitInfo))
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
		gotBranch, err := pfsdb.GetBranchInfo(ctx, tx, id)
		require.NoError(t, err)
		require.True(t, cmp.Equal(branchInfo, gotBranch, compareBranchOpts()...))
		gotBranchByName, err := pfsdb.GetBranchInfoByName(ctx, tx, branchInfo.Branch.Repo.Project.Name, branchInfo.Branch.Repo.Name, branchInfo.Branch.Repo.Type, branchInfo.Branch.Name)
		require.NoError(t, err)
		require.True(t, cmp.Equal(branchInfo, gotBranchByName, compareBranchOpts()...))

		// Update branch to point to second commit
		branchInfo.Head = commit2Info.Commit
		id2, err := pfsdb.UpsertBranch(ctx, tx, branchInfo)
		require.NoError(t, err)
		require.Equal(t, id, id2, "UpsertBranch should keep id stable")
		gotBranch2, err := pfsdb.GetBranchInfo(ctx, tx, id2)
		require.NoError(t, err)
		require.True(t, cmp.Equal(branchInfo, gotBranch2, compareBranchOpts()...))
	})
}

func TestCreateAndGetBranchProvenance(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	commitsCol := pfsdb.Commits(db, nil)
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		// Create 3 repos, A, B, C
		repoAInfo := &pfs.RepoInfo{Repo: &pfs.Repo{Name: "A", Type: pfs.UserRepoType, Project: &pfs.Project{Name: pfs.DefaultProjectName}}}
		repoBInfo := &pfs.RepoInfo{Repo: &pfs.Repo{Name: "B", Type: pfs.UserRepoType, Project: &pfs.Project{Name: pfs.DefaultProjectName}}}
		repoCInfo := &pfs.RepoInfo{Repo: &pfs.Repo{Name: "C", Type: pfs.UserRepoType, Project: &pfs.Project{Name: pfs.DefaultProjectName}}}
		for _, repoInfo := range []*pfs.RepoInfo{repoAInfo, repoBInfo, repoCInfo} {
			_, err := pfsdb.UpsertRepo(ctx, tx, repoInfo)
			require.NoError(t, err)
		}
		// Create 3 commits, one in each repo
		commitSetID := random.String(32)
		commitAInfo := &pfs.CommitInfo{Commit: &pfs.Commit{Repo: repoAInfo.Repo, Id: commitSetID}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}}
		commitBInfo := &pfs.CommitInfo{Commit: &pfs.Commit{Repo: repoBInfo.Repo, Id: commitSetID}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}}
		commitCInfo := &pfs.CommitInfo{Commit: &pfs.Commit{Repo: repoCInfo.Repo, Id: commitSetID}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}}
		for _, commitInfo := range []*pfs.CommitInfo{commitAInfo, commitBInfo, commitCInfo} {
			// Note that commitsCol implicity creates an entry in pfs.commits table via a PutHook
			// TODO replace this with pfsdb.CreateCommit
			require.NoError(t, commitsCol.ReadWrite(tx).Put(commitInfo.Commit, commitInfo))
		}
		// Create 3 branches, one for each repo, pointing to the corresponding commit
		branchAInfo := &pfs.BranchInfo{Branch: &pfs.Branch{Name: "master", Repo: repoAInfo.Repo}, Head: commitAInfo.Commit}
		branchBInfo := &pfs.BranchInfo{Branch: &pfs.Branch{Name: "master", Repo: repoBInfo.Repo}, Head: commitBInfo.Commit}
		branchCInfo := &pfs.BranchInfo{Branch: &pfs.Branch{Name: "master", Repo: repoCInfo.Repo}, Head: commitCInfo.Commit}
		// Provenance info: A <- B <- C, and A <- C
		branchAInfo.Subvenance = []*pfs.Branch{branchBInfo.Branch, branchCInfo.Branch}
		branchBInfo.DirectProvenance = []*pfs.Branch{branchAInfo.Branch}
		branchBInfo.Provenance = []*pfs.Branch{branchAInfo.Branch}
		branchBInfo.Subvenance = []*pfs.Branch{branchCInfo.Branch}
		branchCInfo.DirectProvenance = []*pfs.Branch{branchAInfo.Branch, branchBInfo.Branch}
		branchCInfo.Provenance = []*pfs.Branch{branchBInfo.Branch, branchAInfo.Branch}
		// Create all branches, and provenance relationships
		allBranches := make(map[pfsdb.BranchID]*pfs.BranchInfo)
		for _, branchInfo := range []*pfs.BranchInfo{branchAInfo, branchBInfo, branchCInfo} {
			id, err := pfsdb.UpsertBranch(ctx, tx, branchInfo) // implicitly creates prov relationships
			require.NoError(t, err)
			allBranches[id] = branchInfo
		}
		// Verify direct provenance, full provenance, and full subvenance relationships
		for id, branchInfo := range allBranches {
			gotDirectProv, err := pfsdb.GetDirectBranchProvenance(ctx, tx, id)
			require.NoError(t, err)
			require.True(t, cmp.Equal(branchInfo.DirectProvenance, gotDirectProv, compareBranchOpts()...))
			gotProv, err := pfsdb.GetBranchProvenance(ctx, tx, id)
			require.NoError(t, err)
			require.True(t, cmp.Equal(branchInfo.Provenance, gotProv, compareBranchOpts()...))
			gotSubv, err := pfsdb.GetBranchSubvenance(ctx, tx, id)
			require.NoError(t, err)
			require.True(t, cmp.Equal(branchInfo.Subvenance, gotSubv, compareBranchOpts()...))
		}
		// Update provenance DAG to A <- B -> C, to test adding and deleting prov relationships
		branchAInfo.DirectProvenance = nil
		branchAInfo.Provenance = nil
		branchAInfo.Subvenance = []*pfs.Branch{branchBInfo.Branch}
		branchBInfo.DirectProvenance = []*pfs.Branch{branchAInfo.Branch, branchCInfo.Branch}
		branchBInfo.Provenance = []*pfs.Branch{branchAInfo.Branch, branchCInfo.Branch}
		branchBInfo.Subvenance = nil
		branchCInfo.DirectProvenance = nil
		branchCInfo.Provenance = nil
		branchCInfo.Subvenance = []*pfs.Branch{branchBInfo.Branch}
		for id, branchInfo := range allBranches {
			gotID, err := pfsdb.UpsertBranch(ctx, tx, branchInfo)
			require.NoError(t, err)
			require.Equal(t, id, gotID, "UpsertBranch should keep id stable")
		}
		for id, branchInfo := range allBranches {
			gotDirectProv, err := pfsdb.GetDirectBranchProvenance(ctx, tx, id)
			require.NoError(t, err)
			require.True(t, cmp.Equal(branchInfo.DirectProvenance, gotDirectProv, compareBranchOpts()...))
			gotProv, err := pfsdb.GetBranchProvenance(ctx, tx, id)
			require.NoError(t, err)
			require.True(t, cmp.Equal(branchInfo.Provenance, gotProv, compareBranchOpts()...))
			gotSubv, err := pfsdb.GetBranchSubvenance(ctx, tx, id)
			require.NoError(t, err)
			require.True(t, cmp.Equal(branchInfo.Subvenance, gotSubv, compareBranchOpts()...))
		}
	})

	// Test listing branches
	// TODO move this to a separate test
	qb := pfsdb.QueryBuilder[pfsdb.BranchField]{
		AndFilters: []pfsdb.Filter[pfsdb.BranchField]{
			{Field: pfsdb.BranchFieldRepoName, Op: pfsdb.ValueIn, Values: []any{"A", "C"}},
			{Field: pfsdb.BranchFieldName, Op: pfsdb.Equal, Value: "master"},
		},
	}
	branchIterator, err := pfsdb.NewBranchIterator(ctx, db, qb, 2)
	require.NoError(t, err)
	require.NoError(t, stream.ForEach[pfsdb.BranchPair](ctx, branchIterator, func(branchPair pfsdb.BranchPair) error {
		fmt.Println("qqq", branchPair)
		return nil
	}))
}
