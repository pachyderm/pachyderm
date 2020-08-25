package server

import (
	"github.com/docker/go-units"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
)

// triggerBranch is called when a branch is moved to point to a new commit, it
// updates other branches in the repo if they trigger on the change and returns
// all branches which were moved by this call.
func (d *driver) triggerBranch(
	txnCtx *txnenv.TransactionContext,
	branch *pfs.Branch,
) ([]*pfs.Branch, error) {
	repos := d.repos.ReadWrite(txnCtx.Stm)
	branches := d.branches(branch.Repo.Name).ReadWrite(txnCtx.Stm)
	commits := d.commits(branch.Repo.Name).ReadWrite(txnCtx.Stm)
	repoInfo := &pfs.RepoInfo{}
	if err := repos.Get(branch.Repo.Name, repoInfo); err != nil {
		return nil, err
	}
	bi := &pfs.BranchInfo{}
	if err := branches.Get(branch.Name, bi); err != nil {
		return nil, err
	}
	newHead := &pfs.CommitInfo{}
	if err := commits.Get(bi.Head.ID, newHead); err != nil {
		return nil, err
	}
	var result []*pfs.Branch
	for _, b := range repoInfo.Branches {
		if err := branches.Get(b.Name, bi); err != nil {
			return nil, err
		}
		if bi.Trigger != nil && bi.Trigger.Branch == branch.Name {
			oldHead := &pfs.CommitInfo{}
			if bi.Head != nil {
				if err := commits.Get(bi.Head.ID, oldHead); err != nil {
					return nil, err
				}
			}
			if isTriggered(bi.Trigger, oldHead, newHead) {
				if err := branches.Update(bi.Name, bi, func() error {
					bi.Head = newHead.Commit
					return nil
				}); err != nil {
					return nil, err
				}
				result = append(result, b)
			}
		}
	}
	return result, nil
}

// isTriggered checks to see if a branch should be updated from oldHead to
// newHead based on a trigger.
func isTriggered(t *pfs.Trigger, oldHead, newHead *pfs.CommitInfo) bool {
	if t.Size_ != "" {
		// Shouldn't be possible to error here since we validate on egress
		size, _ := units.FromHumanSize(t.Size_)
		if int64(newHead.SizeBytes-oldHead.SizeBytes) >= size {
			return true
		}
	}
	return false
}
