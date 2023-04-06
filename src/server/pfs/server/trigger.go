package server

import (
	"time"

	units "github.com/docker/go-units"
	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/cronutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// triggerCommit is called when a commit is finished, it updates branches in
// the repo if they trigger on the change
func (d *driver) triggerCommit(
	txnCtx *txncontext.TransactionContext,
	commit *pfs.Commit,
) error {
	repoInfo := &pfs.RepoInfo{}
	if err := d.repos.ReadWrite(txnCtx.SqlTx).Get(commit.Repo, repoInfo); err != nil {
		return errors.EnsureStack(err)
	}
	newHead := &pfs.CommitInfo{}
	if err := d.commits.ReadWrite(txnCtx.SqlTx).Get(commit, newHead); err != nil {
		return errors.EnsureStack(err)
	}
	for _, b := range repoInfo.Branches {
		bi := &pfs.BranchInfo{}
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(b, bi); err != nil {
			return errors.EnsureStack(err)
		}
		oldHead := &pfs.CommitInfo{}
		if err := d.commits.ReadWrite(txnCtx.SqlTx).Get(bi.Head, oldHead); err != nil {
			return errors.EnsureStack(err)
		}
		triggered, err := d.isTriggered(txnCtx, bi.Trigger, oldHead, newHead)
		if err != nil {
			return err
		}
		if triggered {
			var trigBi pfs.BranchInfo
			if err := d.branches.ReadWrite(txnCtx.SqlTx).Update(bi.Branch, &trigBi, func() error {
				trigBi.Head = newHead.Commit
				return nil
			}); err != nil {
				return errors.EnsureStack(err)
			}
			if err := txnCtx.PropagateBranch(bi.Branch); err != nil {
				return err
			}
		}
	}
	return nil
}

// isTriggered checks to see if a branch should be updated from oldHead to
// newHead based on a trigger.
func (d *driver) isTriggered(txnCtx *txncontext.TransactionContext, t *pfs.Trigger, oldHead, newHead *pfs.CommitInfo) (bool, error) {
	if t == nil {
		return false, nil
	}
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
			return false, errors.EnsureStack(err)
		}
		var oldSize int64
		if oldHead != nil {
			oldSize = oldHead.Details.SizeBytes
		}
		merge(newHead.Details.SizeBytes-oldSize >= size)
	}
	if t.CronSpec != "" {
		// Shouldn't be possible to error here since we validate on ingress
		schedule, err := cronutil.ParseCronExpression(t.CronSpec)
		if err != nil {
			// Shouldn't be possible to error here since we validate on ingress
			return false, errors.EnsureStack(err)
		}
		var oldTime, newTime time.Time
		if oldHead != nil && oldHead.Finishing != nil {
			oldTime, err = types.TimestampFromProto(oldHead.Finishing)
			if err != nil {
				return false, errors.EnsureStack(err)
			}
		}
		if newHead.Finishing != nil {
			newTime, err = types.TimestampFromProto(newHead.Finishing)
			if err != nil {
				return false, errors.EnsureStack(err)
			}
		}
		merge(schedule.Next(oldTime).Before(newTime))
	}
	if t.Commits != 0 {
		ci := newHead
		var commits int64
		for commits < t.Commits {
			commits++
			if ci.ParentCommit != nil && oldHead.Commit.ID != ci.ParentCommit.ID {
				var err error
				ci, err = d.resolveCommit(txnCtx.SqlTx, ci.ParentCommit)
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

// validateTrigger returns an error if a trigger is invalid
func (d *driver) validateTrigger(txnCtx *txncontext.TransactionContext, branch *pfs.Branch, trigger *pfs.Trigger) error {
	if trigger == nil {
		return nil
	}
	if trigger.Branch == "" {
		return errors.Errorf("triggers must specify a branch to trigger on")
	}
	if err := ancestry.ValidateName(trigger.Branch); err != nil {
		return err
	}
	if _, err := cronutil.ParseCronExpression(trigger.CronSpec); trigger.CronSpec != "" && err != nil {
		return errors.Wrapf(err, "invalid trigger cron spec")
	}
	if _, err := units.FromHumanSize(trigger.Size_); trigger.Size_ != "" && err != nil {
		return errors.Wrapf(err, "invalid trigger size")
	}
	if trigger.Commits < 0 {
		return errors.Errorf("can't trigger on a negative number of commits")
	}

	biMaps := make(map[string]*pfs.BranchInfo)
	if err := d.listBranchInTransaction(txnCtx, branch.Repo, false, func(bi *pfs.BranchInfo) error {
		biMaps[bi.Branch.Name] = bi
		return nil
	}); err != nil {
		return err
	}
	b := trigger.Branch
	for {
		if b == branch.Name {
			return errors.Errorf("triggers cannot form a loop")
		}
		bi, ok := biMaps[b]
		if ok && bi.Trigger != nil {
			b = bi.Trigger.Branch
		} else {
			break
		}
	}
	return nil
}
