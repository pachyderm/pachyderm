package server

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/stream"

	units "github.com/docker/go-units"
	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/cronutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// triggerCommit is called when a commit is finished, it updates branches in
// the repo if they trigger on the change
func (d *driver) triggerCommit(ctx context.Context, txnCtx *txncontext.TransactionContext, commitInfo *pfs.CommitInfo) error {
	branchInfos := make(map[string]*pfs.BranchInfo)
	iter, err := pfsdb.NewBranchIterator(ctx, txnCtx.SqlTx, 0, 100, &pfs.Branch{
		Repo: commitInfo.Commit.Repo})
	if err != nil {
		return errors.Wrap(err, "trigger commit")
	}
	if err := stream.ForEach[pfsdb.BranchInfoWithID](ctx, iter, func(branchInfoWithID pfsdb.BranchInfoWithID) error {
		branchInfos[pfsdb.BranchKey(branchInfoWithID.Branch)] = branchInfoWithID.BranchInfo
		return nil
	}); err != nil {
		return err
	}
	// Recursively check / fire trigger chains.
	newHeads := make(map[string]*pfs.CommitInfo)
	// the branch can be nil in the case where the commit's branch was deleted before the commit was finished in the backend.
	if commitInfo.Commit.Branch != nil {
		newHeads[pfsdb.BranchKey(commitInfo.Commit.Branch)] = commitInfo
	}
	var triggerBranch func(*pfs.BranchInfo) (*pfs.CommitInfo, error)
	triggerBranch = func(bi *pfs.BranchInfo) (*pfs.CommitInfo, error) {
		branchKey := pfsdb.BranchKey(bi.Branch)
		head, ok := newHeads[branchKey]
		if ok {
			return head, nil
		}
		newHeads[branchKey] = nil
		if bi.Trigger == nil || bi.Trigger.CronSpec != "" {
			return nil, nil
		}
		// Recurse through the trigger chain, checking / firing earlier triggers first.
		triggerBranchKey := pfsdb.BranchKey(bi.Branch.Repo.NewBranch(bi.Trigger.Branch))
		triggerBranchInfo, ok := branchInfos[triggerBranchKey]
		if !ok {
			return nil, nil
		}
		newHead, err := triggerBranch(triggerBranchInfo)
		if err != nil {
			return nil, err
		}
		if newHead == nil {
			return nil, nil
		}
		// Check if the trigger should fire based on the new head commit.
		oldHead, err := pfsdb.GetCommitByCommitKey(ctx, txnCtx.SqlTx, bi.Head)
		if err != nil {
			return nil, errors.Wrap(err, "trigger commit")
		}
		triggered, err := d.isTriggered(ctx, txnCtx, bi.Trigger, oldHead, newHead)
		if err != nil {
			return nil, err
		}
		if !triggered {
			return nil, nil
		}
		// Fire the trigger.
		trigBI, err := pfsdb.GetBranchInfoByName(ctx, txnCtx.SqlTx, bi.Branch.Repo.Project.Name, bi.Branch.Repo.Name,
			bi.Branch.Repo.Type, bi.Branch.Name)
		if err != nil {
			return nil, errors.Wrap(err, "trigger commit")
		}
		trigBI.Head = newHead.Commit
		_, err = pfsdb.UpsertBranch(ctx, txnCtx.SqlTx, trigBI)
		if err != nil {
			return nil, err
		}
		newHeads[branchKey] = newHead
		return newHead, nil
	}
	for _, bi := range branchInfos {
		if _, err := triggerBranch(bi); err != nil {
			return err
		}
	}
	return nil
}

// isTriggered checks to see if a branch should be updated from oldHead to
// newHead based on a trigger.
func (d *driver) isTriggered(ctx context.Context, txnCtx *txncontext.TransactionContext, t *pfs.Trigger, oldHead, newHead *pfs.CommitInfo) (bool, error) {
	result := t.All
	merge := func(cond bool) {
		if t.All {
			result = result && cond
		} else {
			result = result || cond
		}
	}
	if t.Size != "" {
		size, err := units.FromHumanSize(t.Size)
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
	if t.RateLimitSpec != "" {
		// Shouldn't be possible to error here since we validate on ingress
		schedule, err := cronutil.ParseCronExpression(t.RateLimitSpec)
		if err != nil {
			// Shouldn't be possible to error here since we validate on ingress
			return false, errors.EnsureStack(err)
		}
		var oldTime, newTime time.Time
		if oldHead != nil && oldHead.Finishing != nil {
			oldTime = oldHead.Finishing.AsTime()
		}
		if newHead.Finishing != nil {
			newTime = newHead.Finishing.AsTime()
		}
		merge(schedule.Next(oldTime).Before(newTime))
	}
	if t.Commits != 0 {
		ci := newHead
		var commits int64
		for commits < t.Commits {
			commits++
			if ci.ParentCommit != nil && oldHead.Commit.Id != ci.ParentCommit.Id {
				var err error
				ci, err = d.resolveCommit(ctx, txnCtx.SqlTx, ci.ParentCommit)
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
func (d *driver) validateTrigger(ctx context.Context, txnCtx *txncontext.TransactionContext, branch *pfs.Branch, trigger *pfs.Trigger) error {
	if trigger == nil {
		return nil
	}
	if trigger.Branch == "" {
		return errors.Errorf("triggers must specify a branch to trigger on")
	}
	if err := ancestry.ValidateName(trigger.Branch); err != nil {
		return err
	}
	if _, err := cronutil.ParseCronExpression(trigger.RateLimitSpec); trigger.RateLimitSpec != "" && err != nil {
		return errors.Wrapf(err, "invalid trigger rate limit spec")
	}
	if _, err := units.FromHumanSize(trigger.Size); trigger.Size != "" && err != nil {
		return errors.Wrapf(err, "invalid trigger size")
	}
	if trigger.Commits < 0 {
		return errors.Errorf("can't trigger on a negative number of commits")
	}
	if _, err := cronutil.ParseCronExpression(trigger.CronSpec); trigger.CronSpec != "" && err != nil {
		return errors.Wrapf(err, "invalid trigger cron spec")
	}

	biMaps := make(map[string]*pfs.BranchInfo)
	if err := d.listBranchInTransaction(ctx, txnCtx, branch.Repo, false, func(bi *pfs.BranchInfo) error {
		biMaps[bi.Branch.Name] = proto.Clone(bi).(*pfs.BranchInfo)
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
