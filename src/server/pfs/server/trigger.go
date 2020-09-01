package server

import (
	"time"

	"github.com/docker/go-units"
	"github.com/gogo/protobuf/types"
	"github.com/robfig/cron"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
)

// triggerCommit is called when a commit is finished, it updates branches in
// the repo if they trigger on the change and returns all branches which were
// moved by this call.
func (d *driver) triggerCommit(
	txnCtx *txnenv.TransactionContext,
	commit *pfs.Commit,
) ([]*pfs.Branch, error) {
	repos := d.repos.ReadWrite(txnCtx.Stm)
	branches := d.branches(commit.Repo.Name).ReadWrite(txnCtx.Stm)
	commits := d.commits(commit.Repo.Name).ReadWrite(txnCtx.Stm)
	repoInfo := &pfs.RepoInfo{}
	if err := repos.Get(commit.Repo.Name, repoInfo); err != nil {
		return nil, err
	}
	newHead := &pfs.CommitInfo{}
	if err := commits.Get(commit.ID, newHead); err != nil {
		return nil, err
	}
	// find which branches this commit is the head of
	headBranches := make(map[string]bool)
	bi := &pfs.BranchInfo{}
	for _, b := range repoInfo.Branches {
		if err := branches.Get(b.Name, bi); err != nil {
			return nil, err
		}
		if bi.Head != nil && bi.Head.ID == commit.ID {
			headBranches[b.Name] = true
		}
	}
	var result []*pfs.Branch
	for _, b := range repoInfo.Branches {
		if err := branches.Get(b.Name, bi); err != nil {
			return nil, err
		}
		if bi.Trigger != nil && headBranches[bi.Trigger.Branch] {
			var oldHead *pfs.CommitInfo
			if bi.Head != nil {
				oldHead = &pfs.CommitInfo{}
				if err := commits.Get(bi.Head.ID, oldHead); err != nil {
					return nil, err
				}
			}
			triggered, err := d.isTriggered(txnCtx, bi.Trigger, oldHead, newHead)
			if err != nil {
				return nil, err
			}
			if triggered {
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
func (d *driver) isTriggered(txnCtx *txnenv.TransactionContext, t *pfs.Trigger, oldHead, newHead *pfs.CommitInfo) (bool, error) {
	result := t.All
	merge := func(cond bool) {
		if t.All {
			result = result && cond
		} else {
			result = result || cond
		}
	}
	if t.Size_ != "" {
		size, err := units.FromHumanSize(t.Size_)
		if err != nil {
			// Shouldn't be possible to error here since we validate on ingress
			return false, err
		}
		var oldSize uint64
		if oldHead != nil {
			oldSize = oldHead.SizeBytes
		}
		merge(int64(newHead.SizeBytes-oldSize) >= size)
	}
	if t.CronSpec != "" {
		// Shouldn't be possible to error here since we validate on ingress
		schedule, err := cron.ParseStandard(t.CronSpec)
		if err != nil {
			// Shouldn't be possible to error here since we validate on ingress
			return false, err
		}
		var oldTime, newTime time.Time
		if oldHead != nil && oldHead.Finished != nil {
			oldTime, err = types.TimestampFromProto(oldHead.Finished)
			if err != nil {
				return false, err
			}
		}
		if newHead.Finished != nil {
			newTime, err = types.TimestampFromProto(newHead.Finished)
			if err != nil {
				return false, err
			}
		}
		merge(schedule.Next(oldTime).Before(newTime))
	}
	if t.Commits != 0 {
		ci := newHead
		var commits int64
		for commits < t.Commits {
			commits++
			if ci.ParentCommit != nil && (oldHead == nil || oldHead.Commit.ID != ci.ParentCommit.ID) {
				var err error
				ci, err = d.inspectCommit(txnCtx.Client, ci.ParentCommit, pfs.CommitState_STARTED)
				if err != nil {
					return false, err
				}
			} else {
				break
			}
		}
		merge(commits == t.Commits)
	}
	return result, nil
}
