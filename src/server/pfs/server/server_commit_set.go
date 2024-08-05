package server

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

// returns CommitInfos in a commit set, topologically sorted.
// A commit set will include all the commits that were created across repos for a run, along
// with all of the commits that the run's commit's rely on (present in previous commit sets).
func (a *apiServer) inspectCommitSetImmediateTx(ctx context.Context, txnCtx *txncontext.TransactionContext, commitSet *pfs.CommitSet, includeAliases bool) ([]*pfsdb.Commit, error) {
	var commits []*pfsdb.Commit
	if includeAliases {
		cs, err := pfsdb.CommitSetProvenance(txnCtx.SqlTx, commitSet.Id)
		if err != nil {
			return nil, err
		}
		for _, c := range cs {
			commit, err := pfsdb.GetCommitByKey(ctx, txnCtx.SqlTx, c)
			if err != nil {
				return nil, errors.Wrap(err, "inspect commit set immediate")
			}
			commits = append(commits, commit)
		}
	}
	commitsInSet, err := pfsdb.ListCommitTxByFilter(ctx, txnCtx.SqlTx, &pfs.Commit{Id: commitSet.Id})
	if err != nil {
		return nil, errors.Wrap(err, "inspect commit set immediate")
	}
	commits = append(commits, commitsInSet...)
	return TopologicalSort(commits), nil
}

func topSortHelper(ci *pfsdb.Commit, visited map[string]struct{}, commits map[string]*pfsdb.Commit) []*pfsdb.Commit {
	if _, ok := visited[pfsdb.CommitKey(ci.Commit)]; ok {
		return nil
	}
	var result []*pfsdb.Commit
	for _, p := range ci.DirectProvenance {
		provCI, commitExists := commits[pfsdb.CommitKey(p)]
		_, commitVisited := visited[pfsdb.CommitKey(p)]
		if commitExists && !commitVisited {
			result = append(result, topSortHelper(provCI, visited, commits)...)
		}
	}
	result = append(result, ci)
	visited[pfsdb.CommitKey(ci.Commit)] = struct{}{}
	return result
}

// TopologicalSort sorts a slice of commit with ids topologically based on their provenance
func TopologicalSort(cis []*pfsdb.Commit) []*pfsdb.Commit {
	commits := make(map[string]*pfsdb.Commit)
	visited := make(map[string]struct{})
	for _, ci := range cis {
		commits[pfsdb.CommitKey(ci.Commit)] = ci
	}
	var result []*pfsdb.Commit
	for _, ci := range cis {
		result = append(result, topSortHelper(ci, visited, commits)...)
	}
	return result
}

func (a *apiServer) inspectCommitSetImmediate(ctx context.Context, commitset *pfs.CommitSet, cb func(*pfs.CommitInfo) error) error {
	var commits []*pfsdb.Commit
	if err := a.txnEnv.WithReadContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		var err error
		commits, err = a.inspectCommitSetImmediateTx(ctx, txnCtx, commitset, true)
		return err
	}); err != nil {
		return err
	}
	for _, commit := range commits {
		if err := cb(commit.CommitInfo); err != nil {
			return err
		}
	}
	return nil
}

func (a *apiServer) inspectCommitSet(ctx context.Context, commitset *pfs.CommitSet, wait bool, cb func(*pfs.CommitInfo) error) error {
	if !wait {
		return a.inspectCommitSetImmediate(ctx, commitset, cb)
	}
	sent := map[string]struct{}{}
	send := func(ci *pfs.CommitInfo) error {
		if _, ok := sent[pfsdb.CommitKey(ci.Commit)]; ok {
			return nil
		}
		sent[pfsdb.CommitKey(ci.Commit)] = struct{}{}
		return cb(ci)
	}
	var unfinishedCommits []*pfs.Commit
	if err := a.inspectCommitSetImmediate(ctx, commitset, func(ci *pfs.CommitInfo) error {
		if ci.Finished != nil {
			return send(ci)
		}
		unfinishedCommits = append(unfinishedCommits, proto.Clone(ci.Commit).(*pfs.Commit))
		return nil
	}); err != nil {
		return err
	}
	for _, uc := range unfinishedCommits {
		// TODO: make a dedicated call just for the blocking part, inspectCommitInfo is a little heavyweight?
		ci, err := a.inspectCommitInfo(ctx, uc, pfs.CommitState_FINISHED)
		if err != nil {
			return err
		}
		if err := send(ci); err != nil {
			return err
		}
	}
	return nil
}

// TODO(provenance): performance concerns in inspecting each commit set
// TODO(fahad/albert): change list commit set to query the pfsdb and return a list of all unique commit sets.
func (a *apiServer) listCommitSet(ctx context.Context, project *pfs.Project, cb func(*pfs.CommitSetInfo) error) error {
	// Track the commitsets we've already processed
	seen := map[string]struct{}{}
	// Return commitsets by the newest commit in each set (which can be at a different
	// timestamp due to triggers or deferred processing)
	err := pfsdb.ForEachCommit(ctx, a.env.DB, nil, func(commit pfsdb.Commit) error {
		commitInfo := commit.CommitInfo
		if project != nil && commitInfo.Commit.AccessRepo().Project.Name != project.Name {
			return nil
		}
		if _, ok := seen[commitInfo.Commit.Id]; ok {
			return nil
		}
		seen[commitInfo.Commit.Id] = struct{}{}
		var commitInfos []*pfs.CommitInfo
		if err := a.inspectCommitSet(ctx, &pfs.CommitSet{Id: commitInfo.Commit.Id}, false, func(ci *pfs.CommitInfo) error {
			commitInfos = append(commitInfos, ci)
			return nil
		}); err != nil {
			return err
		}
		return cb(&pfs.CommitSetInfo{
			CommitSet: client.NewCommitSet(commitInfo.Commit.Id),
			Commits:   commitInfos,
		})
	}, pfsdb.OrderByCommitColumn{Column: pfsdb.CommitColumnID, Order: pfsdb.SortOrderDesc})
	return errors.Wrap(err, "list commit set")
}

// dropCommitSet is only implemented for commits with no children, so if any
// commits in the commitSet have children the operation will fail.
func (a *apiServer) dropCommitSet(ctx context.Context, txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet) error {
	css, err := a.subvenantCommitSets(txnCtx, commitset)
	if err != nil {
		return err
	}
	if len(css) > 0 {
		return &pfsserver.ErrSquashWithSubvenance{CommitSet: commitset, SubvenantCommitSets: css}
	}
	cis, err := a.inspectCommitSetImmediateTx(ctx, txnCtx, commitset, false)
	if err != nil {
		return err
	}
	for _, ci := range cis {
		if ci.Commit.AccessRepo().Type == pfs.SpecRepoType && ci.Origin.Kind == pfs.OriginKind_USER {
			return errors.Errorf("cannot squash commit %s because it updated a pipeline", ci.Commit)
		}
		if len(ci.ChildCommits) > 0 {
			return &pfsserver.ErrDropWithChildren{Commit: ci.Commit}
		}
	}
	// While this is a 'drop' operation and not a 'squash', proper drop semantics
	// aren't implemented at the moment.  Squashing the head of a branch is
	// effectively a drop, though, because there is no child commit that contains
	// the data from the given commits, which is why it is an error to drop any
	// non-head commits (until generalized drop semantics are implemented).
	for _, ci := range cis {
		if err := a.deleteCommit(ctx, txnCtx, ci); err != nil {
			return err
		}
	}
	// notify PPS that this commitset has been dropped so it can clean up any
	// jobs associated with it at the end of the transaction
	txnCtx.StopJobSet(commitset)
	return nil
}

func (a *apiServer) squashCommitSet(ctx context.Context, txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet) error {
	css, err := a.subvenantCommitSets(txnCtx, commitset)
	if err != nil {
		return err
	}
	if len(css) > 0 {
		return &pfsserver.ErrSquashWithSubvenance{CommitSet: commitset, SubvenantCommitSets: css}
	}
	commitInfos, err := a.inspectCommitSetImmediateTx(ctx, txnCtx, commitset, false)
	if err != nil {
		return err
	}
	for _, ci := range commitInfos {
		if ci.Commit.Repo.Type == pfs.SpecRepoType && ci.Origin.Kind == pfs.OriginKind_USER {
			return errors.Errorf("cannot squash commit %s because it updated a pipeline", ci.Commit)
		}
		if len(ci.ChildCommits) == 0 {
			return &pfsserver.ErrSquashWithoutChildren{Commit: ci.Commit}
		}
	}
	for _, ci := range commitInfos {
		if err := a.deleteCommit(ctx, txnCtx, ci); err != nil {
			return err
		}
	}
	// notify PPS that this commitset has been squashed so it can clean up any
	// jobs associated with it at the end of the transaction
	txnCtx.StopJobSet(commitset)
	return nil
}

// Since commits are only deleted as part of deleting a commit set, in most cases
// we will delete many commits at a time. The graph traversal computations can result
// in many I/Os. Therefore, this function bulkifies the graph traversals to run the
// entire operation performantly. This function takes care to load and save an object
// no more than one time.
//
// to delete a single commit
// 1. delete the commit and its associated file set
// 2. check whether the commit was at the head of a branch, and update the branch head if necessary
// 3. updating the ChildCommits pointers of deletedCommit.ParentCommit
// 4. updating the ParentCommit pointer of deletedCommit.ChildCommits
func (a *apiServer) deleteCommit(ctx context.Context, txnCtx *txncontext.TransactionContext, ci *pfsdb.Commit) error {
	for _, child := range ci.ChildCommits {
		childInfo, err := pfsdb.GetCommitInfoByKey(ctx, txnCtx.SqlTx, child)
		if err != nil {
			return errors.Wrapf(err, "error checking child commit state")
		}
		if childInfo.Finished == nil {
			var suffix string
			if childInfo.Finishing != nil {
				// user might already have called "finish",
				suffix = ", consider using WaitCommit"
			}
			return errors.Errorf("cannot squash until child commit %s is finished%s", child, suffix)
		}
	}

	// Delete the commit's filesets
	if err := a.commitStore.DropFileSetsTx(txnCtx.SqlTx, ci); err != nil {
		return errors.EnsureStack(err)
	}
	// update branch heads
	headlessBranches := make([]*pfs.BranchInfo, 0)
	repoInfo, err := pfsdb.GetRepoByName(ctx, txnCtx.SqlTx, ci.Commit.Repo.Project.Name, ci.Commit.Repo.Name, ci.Commit.Repo.Type)
	if err != nil {
		return err
	}
	branchInfo := &pfs.BranchInfo{}
	for _, b := range repoInfo.Branches {
		branch, err := pfsdb.GetBranch(ctx, txnCtx.SqlTx, b)
		if err != nil {
			return errors.Wrapf(err, "delete commit: getting branch %s", b)
		}
		branchInfo = branch.BranchInfo
		if pfsdb.CommitKey(branchInfo.Head) == pfsdb.CommitKey(ci.Commit) {
			if ci.ParentCommit == nil {
				headlessBranches = append(headlessBranches, proto.Clone(branchInfo).(*pfs.BranchInfo))
			} else {
				branchInfo.Head = ci.ParentCommit
			}
		}
		if _, err := pfsdb.UpsertBranch(ctx, txnCtx.SqlTx, branchInfo); err != nil {
			return errors.Wrapf(err, "delete commit: updating branch %s", branchInfo.Branch)
		}
	}
	if len(headlessBranches) > 0 {
		repoCommit, err := a.makeEmptyCommit(ctx, txnCtx, headlessBranches[0].Branch, headlessBranches[0].DirectProvenance, nil)
		if err != nil {
			return err
		}
		for _, bi := range headlessBranches {
			bi.Head = repoCommit.CommitInfo.Commit
			if _, err := pfsdb.UpsertBranch(ctx, txnCtx.SqlTx, bi); err != nil {
				return errors.Wrapf(err, "delete commit: updating branch %s", bi.Branch)
			}
		}
	}
	// delete commit.
	if err := pfsdb.DeleteCommit(ctx, txnCtx.SqlTx, ci.Commit); err != nil {
		return errors.EnsureStack(err)
	}
	return nil
}

// A commit set 'X' is subvenant to another commit set 'Y' if it contains commits
// that are subvenant to commits in 'Y'. Commit set subvenance is transitivie.
//
// The implementation repeatedly applies CommitSetSubvenance() to compute all of the Subvenant commit sets.
// To understand why, first consider the simple case with the commit provenance graph where r@X & q@Y are in p@Y's provenance.
// For this graph, CommitSetSubvenance("X") evaluates to [p@Y] which we can use to infer that commit set Y is subvenant to commit set X.
// Now consider the same graph, with the addition of a commit s@Z that has q@Y in its subvenance.
// In this case, CommitSetSubvenance(X) still evaluates to [p@Y]. But since a commit in 'Z', depends on a commit
// in 'Y', we haven't yet computed all of 'X”s subvenant commit sets. Therefore,
// we re-evaluate CommitSetSubvenance for each collected commit set until our resulting set becomes stable.
func (a *apiServer) subvenantCommitSets(txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet) ([]*pfs.CommitSet, error) {
	collectSubvCommitSets := func(setIDs map[string]struct{}) (map[string]struct{}, error) {
		subvCommitSets := make(map[string]struct{})
		for id := range setIDs {
			subvCommits, err := pfsdb.CommitSetSubvenance(txnCtx.SqlTx, id)
			if err != nil {
				return nil, err
			}
			for _, subvCommit := range subvCommits {
				if _, ok := setIDs[subvCommit.Id]; !ok {
					subvCommitSets[subvCommit.Id] = struct{}{}
				}
			}
		}
		return subvCommitSets, nil
	}
	subvCSs, err := collectSubvCommitSets(map[string]struct{}{
		commitset.Id: {},
	})
	if err != nil {
		return nil, err
	}
	completeSubvCSs := make(map[string]struct{})
	for len(subvCSs) > 0 {
		for cs := range subvCSs {
			completeSubvCSs[cs] = struct{}{}
		}
		subvCSs, err = collectSubvCommitSets(subvCSs)
		if err != nil {
			return nil, err
		}
	}
	result := make([]*pfs.CommitSet, 0)
	for cs := range completeSubvCSs {
		result = append(result, &pfs.CommitSet{Id: cs})
	}
	return result, nil
}
