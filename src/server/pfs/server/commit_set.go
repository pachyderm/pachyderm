package server

import (
	"context"
	"sort"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

// returns CommitInfos in a commit set, topologically sorted.
// A commit set will include all the commits that were created across repos for a run, along
// with all of the commits that the run's commit's rely on (present in previous commit sets).
func (d *driver) inspectCommitSetImmediateTx(txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet, filterAliases bool) ([]*pfs.CommitInfo, error) {
	cis := make([]*pfs.CommitInfo, 0)
	if !filterAliases {
		cs, err := pfsdb.CommitSetProvenance(txnCtx.SqlTx, commitset.ID)
		if err != nil {
			return nil, err
		}
		for _, c := range cs {
			ci := &pfs.CommitInfo{}
			if err := d.commits.ReadWrite(txnCtx.SqlTx).Get(c, ci); err != nil {
				return nil, err
			}
			cis = append(cis, ci)
		}
	}
	ci := &pfs.CommitInfo{}
	if err := d.commits.ReadWrite(txnCtx.SqlTx).GetByIndex(pfsdb.CommitsCommitSetIndex, commitset.ID, ci, col.DefaultOptions(), func(string) error {
		cis = append(cis, proto.Clone(ci).(*pfs.CommitInfo))
		return nil
	}); err != nil {
		return nil, err
	}
	return TopologicalSort(cis), nil
}

// TopologicalSort sorts a slice of commit infos topologically based on their provenance
func TopologicalSort(cis []*pfs.CommitInfo) []*pfs.CommitInfo {
	commits := make(map[string]*pfs.CommitInfo)
	for _, ci := range cis {
		commits[pfsdb.CommitKey(ci.Commit)] = ci
	}
	commitSubv := make(map[string][]string) // maps commit key -> []commit keys
	queue := make([]string, 0)
	seenQueue := make(map[string]struct{})
	res := make([]*pfs.CommitInfo, 0)
	// set up commitSubv
	for _, ci := range commits {
		for _, p := range ci.DirectProvenance {
			if _, ok := commits[pfsdb.CommitKey(p)]; ok {
				commitSubv[pfsdb.CommitKey(p)] = append(commitSubv[pfsdb.CommitKey(p)], pfsdb.CommitKey(ci.Commit))
			}
		}
	}
	canPop := func(k string) bool {
		for _, p := range commits[k].DirectProvenance {
			if _, ok := commitSubv[pfsdb.CommitKey(p)]; ok {
				return false
			}
		}
		return true
	}
	// load queue with commits that can be popped
	for _, ci := range commits {
		if canPop(pfsdb.CommitKey(ci.Commit)) {
			queue = append(queue, pfsdb.CommitKey(ci.Commit))
		}
	}
	for len(res) < len(commits) {
		for i, k := range queue {
			if canPop(k) {
				res = append(res, commits[k])
				// add k's commitSubv to queue
				for _, s := range commitSubv[k] {
					if _, seen := seenQueue[s]; !seen {
						queue = append(queue, s)
						seenQueue[s] = struct{}{}
					}
				}
				delete(commitSubv, k)
				queue = append(queue[:i], queue[i+1:]...) // pop from queue
				break
			}
		}
	}
	return res
}

func (d *driver) inspectCommitSetImmediate(ctx context.Context, commitset *pfs.CommitSet, cb func(*pfs.CommitInfo) error) error {
	var commitInfos []*pfs.CommitInfo
	if err := d.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		commitInfos, err = d.inspectCommitSetImmediateTx(txnCtx, commitset, false)
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

// applies a callback, cb, to all the commits in a commit set. A commit set includes
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
	unfinishedCommits := make([]*pfs.Commit, 0)
	// NOTE: before 2.5, a triggered commits would be included as part of the same commit set as the triggering commit set,
	// so we would wait for all of the current commit set to finish, then check if new previously unknown commits are added due to triggers.
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

func (d *driver) listCommitSet(ctx context.Context, project *pfs.Project, cb func(*pfs.CommitSetInfo) error) error {
	// Track the commitsets we've already processed
	seen := map[string]struct{}{}
	// Return commitsets by the newest commit in each set (which can be at a different
	// timestamp due to triggers or deferred processing)
	commitInfo := &pfs.CommitInfo{}
	err := d.commits.ReadOnly(ctx).List(commitInfo, col.DefaultOptions(), func(string) error {
		if commitInfo.GetCommit().GetBranch().GetRepo().GetProject().GetName() != project.GetName() {
			return nil
		}
		if _, ok := seen[commitInfo.Commit.ID]; ok {
			return nil
		}
		seen[commitInfo.Commit.ID] = struct{}{}
		var commitInfos []*pfs.CommitInfo
		err := d.inspectCommitSet(ctx, &pfs.CommitSet{ID: commitInfo.Commit.ID}, false, func(ci *pfs.CommitInfo) error {
			commitInfos = append(commitInfos, ci)
			return nil
		})
		if err != nil {
			return err
		}
		return cb(&pfs.CommitSetInfo{
			CommitSet: client.NewCommitSet(commitInfo.Commit.ID),
			Commits:   commitInfos,
		})
	})
	return errors.EnsureStack(err)
}

// dropCommitSet is only implemented for commits with no children, so if any
// commits in the commitSet have children the operation will fail.
func (d *driver) dropCommitSet(txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet) error {
	css, err := d.subvenantCommitSets(txnCtx, commitset)
	if err != nil {
		return err
	}
	if len(css) > 0 {
		return &pfsserver.ErrSquashWithSubvenance{CommitSet: commitset, SubvenantCommitSets: css}
	}
	cis, err := d.inspectCommitSetImmediateTx(txnCtx, commitset, true)
	if err != nil {
		return err
	}
	for _, ci := range cis {
		if ci.Commit.Branch.Repo.Type == pfs.SpecRepoType && ci.Origin.Kind == pfs.OriginKind_USER {
			return errors.Errorf("cannot squash commit %s because it updated a pipeline", ci.Commit)
		}
		// TODO(acohen4): can drop commits & squash just be modeled the same for now?
		// if len(ci.ChildCommits) > 0 {
		// 	return &pfsserver.ErrDropWithChildren{Commit: ci.Commit}
		// }
	}
	// While this is a 'drop' operation and not a 'squash', proper drop semantics
	// aren't implemented at the moment.  Squashing the head of a branch is
	// effectively a drop, though, because there is no child commit that contains
	// the data from the given commits, which is why it is an error to drop any
	// non-head commits (until generalized drop semantics are implemented).
	if err := d.deleteCommits(txnCtx, cis); err != nil {
		return err
	}
	// notify PPS that this commitset has been dropped so it can clean up any
	// jobs associated with it at the end of the transaction
	txnCtx.StopJobs(commitset)
	return nil
}

func (d *driver) squashCommitSets(txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet, force bool) error {
	css, err := d.subvenantCommitSets(txnCtx, commitset)
	if err != nil {
		return err
	}
	if len(css) > 0 && !force {
		return &pfsserver.ErrSquashWithSubvenance{CommitSet: commitset, SubvenantCommitSets: css}
	}
	commitsets := append([]*pfs.CommitSet{commitset}, css...)
	var commitInfos []*pfs.CommitInfo
	for _, commitset := range commitsets {
		var err error
		cis, err := d.inspectCommitSetImmediateTx(txnCtx, commitset, true)
		if err != nil {
			return err
		}
		commitInfos = append(commitInfos, cis...)
	}
	for _, ci := range commitInfos {
		if ci.Commit.Branch.Repo.Type == pfs.SpecRepoType && ci.Origin.Kind == pfs.OriginKind_USER {
			return errors.Errorf("cannot squash commit %s because it updated a pipeline", ci.Commit)
		}
		if len(ci.ChildCommits) == 0 {
			return &pfsserver.ErrSquashWithoutChildren{Commit: ci.Commit}
		}
	}
	if err := d.deleteCommits(txnCtx, commitInfos); err != nil {
		return err
	}
	// notify PPS that this commitset has been squashed so it can clean up any
	// jobs associated with it at the end of the transaction
	for _, commitset := range commitsets {
		txnCtx.StopJobs(commitset)
	}
	return nil
}

// deleteCommits() accepts commitInfos that may span commit sets.
//
// Since commits are only deleted as part of deleting a commit set, in most cases
// we will delete many commits at a time. The graph traversal computations can result
// in many I/Os. Therefore, this function bulkifies the graph traversals to run the
// entire operation performantly. This function takes care to load and save an object
// no more than one time.
//
// to delete a single commit
// 1. delete the commit's file set
// 2. check to whether the commit was at the head of a branch, and update the branch head if necessary
// 3. updating the ChildCommits pointers of deletedCommit.ParentCommit
// 4. updating the ParentCommit pointer of deletedCommit.ChildCommits
func (d *driver) deleteCommits(txnCtx *txncontext.TransactionContext, commitInfos []*pfs.CommitInfo) error {
	deleteCommits := make(map[string]*pfs.CommitInfo)
	repos := make(map[string]*pfs.Repo)
	for _, ci := range commitInfos {
		deleteCommits[pfsdb.CommitKey(ci.Commit)] = ci
		repos[pfsdb.RepoKey(ci.Commit.Repo)] = ci.Commit.Repo
	}
	// delete the commits and their filesets
	for _, ci := range commitInfos {
		// make sure all children are finished, so we don't lose data
		for _, child := range ci.ChildCommits {
			if _, ok := deleteCommits[pfsdb.CommitKey(child)]; ok {
				// this child is being deleted, any files from this commit will end up
				// as part of *its* children, which have already been checked
				continue
			}
			var childInfo pfs.CommitInfo
			if err := d.commits.ReadWrite(txnCtx.SqlTx).Get(child, &childInfo); err != nil {
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
		if err := d.commits.ReadWrite(txnCtx.SqlTx).Delete(ci.Commit); err != nil {
			return errors.EnsureStack(err)
		}
		// Delete the commit's filesets
		if err := d.commitStore.DropFileSetsTx(txnCtx.SqlTx, ci.Commit); err != nil {
			return errors.EnsureStack(err)
		}
	}
	// update branch heads
	headlessBranches := make([]*pfs.BranchInfo, 0)
	for _, repo := range repos {
		repoInfo := &pfs.RepoInfo{}
		if err := d.repos.ReadWrite(txnCtx.SqlTx).Get(repo, repoInfo); err != nil {
			return err
		}
		branchInfo := &pfs.BranchInfo{}
		for _, b := range repoInfo.Branches {
			if err := d.branches.ReadWrite(txnCtx.SqlTx).Update(b, branchInfo, func() error {
				if ci, ok := deleteCommits[pfsdb.CommitKey(branchInfo.Head)]; ok {
					branchInfo.Head = oldestAncestor(ci, deleteCommits)
					if branchInfo.Head == nil {
						headlessBranches = append(headlessBranches, proto.Clone(branchInfo).(*pfs.BranchInfo))
					}
				}
				return nil
			}); err != nil {
				return err
			}
		}
	}
	sort.Slice(headlessBranches, func(i, j int) bool { return len(headlessBranches[i].Provenance) < len(headlessBranches[j].Provenance) })
	newRepoCommits := make(map[string]*pfs.Commit)
	for _, bi := range headlessBranches {
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Update(bi.Branch, bi, func() error {
			// Create a new empty commit for the branch head
			var repoCommit *pfs.Commit
			var err error
			if c, ok := newRepoCommits[pfsdb.RepoKey(bi.Branch.Repo)]; !ok {
				repoCommit, err = d.makeEmptyCommit(txnCtx, bi.Branch, bi.DirectProvenance, nil)
				if err != nil {
					return err
				}
				newRepoCommits[pfsdb.RepoKey(bi.Branch.Repo)] = repoCommit
			} else {
				repoCommit = c
			}
			bi.Head = repoCommit
			return nil
		}); err != nil {
			return errors.Wrapf(err, "error updating branch %s", bi.Branch)
		}
	}
	// update parent/child relationships on commits where necessary.
	// for each deleted commit, update its parent's children and childrens' parents.
	parentsToNewChildren := make(map[*pfs.Commit]map[*pfs.Commit]struct{})
	parentsToDeleteChildren := make(map[*pfs.Commit]map[*pfs.Commit]struct{})
	childrenToNewParent := make(map[*pfs.Commit]*pfs.Commit)
	for _, ci := range commitInfos {
		parent := oldestAncestor(ci, deleteCommits)
		newDescendants := traverseToEdges(ci, deleteCommits, func(commitInfo *pfs.CommitInfo) []*pfs.Commit {
			return commitInfo.ChildCommits
		})
		for _, descendant := range newDescendants {
			childrenToNewParent[descendant] = parent
		}
		if parent != nil {
			deleteChildren, ok := parentsToDeleteChildren[parent]
			if !ok {
				deleteChildren = make(map[*pfs.Commit]struct{})
				parentsToDeleteChildren[parent] = deleteChildren
			}
			deleteChildren[ci.Commit] = struct{}{}
			newChildren, ok := parentsToNewChildren[parent]
			if !ok {
				newChildren = make(map[*pfs.Commit]struct{})
				parentsToNewChildren[parent] = newChildren
			}
			for _, child := range newDescendants {
				newChildren[child] = struct{}{}
			}
		}
	}
	for child, parent := range childrenToNewParent {
		childInfo := &pfs.CommitInfo{}
		if err := d.commits.ReadWrite(txnCtx.SqlTx).Update(child, childInfo, func() error {
			childInfo.ParentCommit = parent
			return nil
		}); err != nil {
			return errors.Wrapf(err, "error updating parent/child pointers")
		}
	}
	for parent := range parentsToNewChildren {
		parentInfo := &pfs.CommitInfo{}
		if err := d.commits.ReadWrite(txnCtx.SqlTx).Update(parent, parentInfo, func() error {
			childrenSet := make(map[string]*pfs.Commit)
			for _, fc := range parentInfo.ChildCommits {
				childrenSet[pfsdb.CommitKey(fc)] = fc
			}
			for newChild := range parentsToNewChildren[parent] {
				childrenSet[pfsdb.CommitKey(newChild)] = newChild
			}
			for deleteChild := range parentsToDeleteChildren[parent] {
				delete(childrenSet, pfsdb.CommitKey(deleteChild))
			}
			parentInfo.ChildCommits = make([]*pfs.Commit, 0)
			for _, c := range childrenSet {
				parentInfo.ChildCommits = append(parentInfo.ChildCommits, c)
			}
			return nil
		}); err != nil {
			return errors.Wrapf(err, "error updating parent/child pointers")
		}
	}
	return nil
}

// oldestAncestor returns the oldest ancestor commit of 'startCommit' leveraging the pool of known commits in the 'skipSet'
func oldestAncestor(startCommit *pfs.CommitInfo, skipSet map[string]*pfs.CommitInfo) *pfs.Commit {
	oldest := traverseToEdges(startCommit, skipSet, func(commitInfo *pfs.CommitInfo) []*pfs.Commit {
		return []*pfs.Commit{commitInfo.ParentCommit}
	})
	if len(oldest) == 0 {
		return nil
	}
	return oldest[0]
}

// traverseToEdges does a breadth first search using a traverse function.
// returns all of the commits that can not continue traversal within the skip set, hence returning the known 'leaves' of the skipSet graph
func traverseToEdges(startCommit *pfs.CommitInfo, skipSet map[string]*pfs.CommitInfo, traverse func(*pfs.CommitInfo) []*pfs.Commit) []*pfs.Commit {
	cs := []*pfs.Commit{startCommit.Commit}
	result := make([]*pfs.Commit, 0)
	var c *pfs.Commit
	for len(cs) > 0 {
		c, cs = cs[0], cs[1:]
		if c == nil {
			continue
		}
		ci, ok := skipSet[pfsdb.CommitKey(c)]
		if ok {
			cs = append(cs, traverse(ci)...)
		} else {
			result = append(result, proto.Clone(c).(*pfs.Commit))
		}
	}
	return result
}

// a commit set 'X' is subvenant to another commit set 'Y' if it contains commits
// that are subvenant to commits in 'Y'.
// commit set subvenance is transitivie. So if 'X' is subvenant to 'Y', and 'Y', is provenant to
//
// imagine the commit provenance graph where r@X & q@Y are in p@Y's provenance. q@Y is also in s@Z's provenance.
// i.e.
// r@X    q@Y
//  \     / \
//   \   /   \
//    p@Y    s@Z
// CommitSetSubvenance(X) evaluates to [p@Y]. Since we delete commit sets in batches, we would delete all of
// commit set Y. But this would inadvertently kill commit q@Y which is in s@Z's provenance. Therefore,
// we re-evaluate CommitSetSubvenance for each collected commit set until or resulting set becomes stable.
func (d *driver) subvenantCommitSets(txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet) ([]*pfs.CommitSet, error) {
	collectSubvCommitSets := func(setIDs map[string]struct{}) (map[string]struct{}, error) {
		subvCommitSets := make(map[string]struct{})
		for id := range setIDs {
			subvCommits, err := pfsdb.CommitSetSubvenance(txnCtx.SqlTx, id)
			if err != nil {
				return nil, err
			}
			for _, subvCommit := range subvCommits {
				if _, ok := setIDs[subvCommit.ID]; !ok {
					subvCommitSets[subvCommit.ID] = struct{}{}
				}
			}
		}
		return subvCommitSets, nil
	}
	subvCSs, err := collectSubvCommitSets(map[string]struct{}{
		commitset.ID: {},
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
		result = append(result, &pfs.CommitSet{ID: cs})
	}
	return result, nil
}
