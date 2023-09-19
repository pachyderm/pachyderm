package server

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"

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
func (d *driver) inspectCommitSetImmediateTx(ctx context.Context, txnCtx *txncontext.TransactionContext, commitSet *pfs.CommitSet, includeAliases bool) ([]*pfs.CommitInfo, error) {
	var cis []*pfs.CommitInfo
	if includeAliases {
		cs, err := pfsdb.CommitSetProvenance(txnCtx.SqlTx, commitSet.Id)
		if err != nil {
			return nil, err
		}
		for _, c := range cs {
			ci, err := pfsdb.GetCommitByCommitKey(ctx, txnCtx.SqlTx, c)
			if err != nil {
				return nil, errors.Wrap(err, "inspect commit set immediate")
			}
			cis = append(cis, ci)
		}
	}
	cis, err := pfsdb.ListCommitTxByFilter(ctx, txnCtx.SqlTx, pfsdb.CommitListFilter{pfsdb.CommitSetIDs: []string{commitSet.Id}}, false)
	if err != nil {
		return nil, errors.Wrap(err, "inspect commit set immediate")
	}
	return TopologicalSort(cis), nil
}

func topSortHelper(ci *pfs.CommitInfo, visited map[string]struct{}, commits map[string]*pfs.CommitInfo) []*pfs.CommitInfo {
	if _, ok := visited[pfsdb.CommitKey(ci.Commit)]; ok {
		return nil
	}
	var result []*pfs.CommitInfo
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

// TopologicalSort sorts a slice of commit infos topologically based on their provenance
func TopologicalSort(cis []*pfs.CommitInfo) []*pfs.CommitInfo {
	commits := make(map[string]*pfs.CommitInfo)
	visited := make(map[string]struct{})
	for _, ci := range cis {
		commits[pfsdb.CommitKey(ci.Commit)] = ci
	}
	var result []*pfs.CommitInfo
	for _, ci := range cis {
		result = append(result, topSortHelper(ci, visited, commits)...)
	}
	return result
}

func (d *driver) inspectCommitSetImmediate(ctx context.Context, commitset *pfs.CommitSet, cb func(*pfs.CommitInfo) error) error {
	var commitInfos []*pfs.CommitInfo
	if err := d.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		commitInfos, err = d.inspectCommitSetImmediateTx(ctx, txnCtx, commitset, true)
		return err
	}); err != nil {
		return err
	}
	for _, commitInfo := range commitInfos {
		if err := cb(commitInfo); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) inspectCommitSet(ctx context.Context, commitset *pfs.CommitSet, wait bool, cb func(*pfs.CommitInfo) error) error {
	if !wait {
		return d.inspectCommitSetImmediate(ctx, commitset, cb)
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
	if err := d.inspectCommitSetImmediate(ctx, commitset, func(ci *pfs.CommitInfo) error {
		if ci.Finished != nil {
			return send(ci)
		}
		unfinishedCommits = append(unfinishedCommits, proto.Clone(ci.Commit).(*pfs.Commit))
		return nil
	}); err != nil {
		return err
	}
	for _, uc := range unfinishedCommits {
		// TODO: make a dedicated call just for the blocking part, inspectCommit is a little heavyweight?
		ci, err := d.inspectCommit(ctx, uc, pfs.CommitState_FINISHED)
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
func (d *driver) listCommitSet(ctx context.Context, project *pfs.Project, cb func(*pfs.CommitSetInfo) error) error {
	// Track the commitsets we've already processed
	seen := map[string]struct{}{}
	// Return commitsets by the newest commit in each set (which can be at a different
	// timestamp due to triggers or deferred processing)
	iter, err := pfsdb.ListCommit(ctx, d.env.DB, nil, false)
	if err != nil {
		return errors.Wrap(err, "list commit set")
	}
	err = stream.ForEach[pfsdb.CommitWithID](ctx, iter, func(commitPair pfsdb.CommitWithID) error {
		commitInfo := commitPair.CommitInfo
		if project != nil && commitInfo.Commit.AccessRepo().Project.Name != project.Name {
			return nil
		}
		if _, ok := seen[commitInfo.Commit.Id]; ok {
			return nil
		}
		seen[commitInfo.Commit.Id] = struct{}{}
		var commitInfos []*pfs.CommitInfo
		if err := d.inspectCommitSet(ctx, &pfs.CommitSet{Id: commitInfo.Commit.Id}, false, func(ci *pfs.CommitInfo) error {
			commitInfos = append(commitInfos, ci)
			return nil
		}); err != nil {
			return err
		}
		return cb(&pfs.CommitSetInfo{
			CommitSet: client.NewCommitSet(commitInfo.Commit.Id),
			Commits:   commitInfos,
		})
	})
	return errors.Wrap(err, "list commit set")
}

// dropCommitSet is only implemented for commits with no children, so if any
// commits in the commitSet have children the operation will fail.
func (d *driver) dropCommitSet(ctx context.Context, txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet) error {
	css, err := d.subvenantCommitSets(txnCtx, commitset)
	if err != nil {
		return err
	}
	if len(css) > 0 {
		return &pfsserver.ErrSquashWithSubvenance{CommitSet: commitset, SubvenantCommitSets: css}
	}
	cis, err := d.inspectCommitSetImmediateTx(ctx, txnCtx, commitset, false)
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
		if err := d.deleteCommit(ctx, txnCtx, ci); err != nil {
			return err
		}
	}
	// notify PPS that this commitset has been dropped so it can clean up any
	// jobs associated with it at the end of the transaction
	txnCtx.StopJobs(commitset)
	return nil
}

func (d *driver) squashCommitSet(ctx context.Context, txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet) error {
	css, err := d.subvenantCommitSets(txnCtx, commitset)
	if err != nil {
		return err
	}
	if len(css) > 0 {
		return &pfsserver.ErrSquashWithSubvenance{CommitSet: commitset, SubvenantCommitSets: css}
	}
	commitInfos, err := d.inspectCommitSetImmediateTx(ctx, txnCtx, commitset, false)
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
		if err := d.deleteCommit(ctx, txnCtx, ci); err != nil {
			return err
		}
	}
	// notify PPS that this commitset has been squashed so it can clean up any
	// jobs associated with it at the end of the transaction
	txnCtx.StopJobs(commitset)
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
func (d *driver) deleteCommit(ctx context.Context, txnCtx *txncontext.TransactionContext, ci *pfs.CommitInfo) error {
	// Delete the commit's filesets
	if err := d.commitStore.DropFileSetsTx(txnCtx.SqlTx, ci.Commit); err != nil {
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
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Update(b, branchInfo, func() error {
			if pfsdb.CommitKey(branchInfo.Head) == pfsdb.CommitKey(ci.Commit) {
				if ci.ParentCommit == nil {
					headlessBranches = append(headlessBranches, proto.Clone(branchInfo).(*pfs.BranchInfo))
				} else {
					branchInfo.Head = ci.ParentCommit
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	if len(headlessBranches) > 0 {
		repoCommit, err := d.makeEmptyCommit(ctx, txnCtx, headlessBranches[0].Branch, headlessBranches[0].DirectProvenance, nil)
		if err != nil {
			return err
		}
		for _, bi := range headlessBranches {
			if err := d.branches.ReadWrite(txnCtx.SqlTx).Update(bi.Branch, bi, func() error {
				bi.Head = repoCommit
				return nil
			}); err != nil {
				return errors.Wrapf(err, "error updating branch %s", bi.Branch)
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
func (d *driver) subvenantCommitSets(txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet) ([]*pfs.CommitSet, error) {
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
