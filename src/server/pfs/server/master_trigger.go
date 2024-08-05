package server

import (
	"context"
	"time"

	units "github.com/docker/go-units"
	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/cronutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// triggerCommit is called when a commit is finished, it updates branches in
// the repo if they trigger on the change
func (m *Master) triggerCommit(ctx context.Context, txnCtx *txncontext.TransactionContext, commitInfo *pfs.CommitInfo) error {
	branchInfos := make(map[string]*pfs.BranchInfo)
	err := pfsdb.ForEachBranch(ctx, txnCtx.SqlTx, &pfs.Branch{Repo: commitInfo.Commit.Repo}, func(branch pfsdb.Branch) error {
		branchInfos[branch.Branch.Key()] = branch.BranchInfo
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "trigger commit")
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
		if newHead == nil || proto.Equal(bi.Head, newHead.Commit) {
			return nil, nil
		}
		// Check if the trigger should fire based on the new head commit.
		oldHead, err := pfsdb.GetCommitInfoByKey(ctx, txnCtx.SqlTx, bi.Head)
		if err != nil {
			return nil, errors.Wrap(err, "trigger commit")
		}
		triggered, err := m.isTriggered(ctx, txnCtx, bi.Trigger, oldHead, newHead)
		if err != nil {
			return nil, err
		}
		if !triggered {
			return nil, nil
		}
		// Fire the trigger.
		branchToTrigger, err := pfsdb.GetBranch(ctx, txnCtx.SqlTx, bi.Branch)
		if err != nil {
			return nil, errors.Wrap(err, "trigger commit")
		}
		trigBI := branchToTrigger.BranchInfo
		trigBI.Head = newHead.Commit
		_, err = pfsdb.UpsertBranch(ctx, txnCtx.SqlTx, trigBI)
		if err != nil {
			return nil, err
		}
		newHeads[branchKey] = newHead
		if err := txnCtx.PropagateBranch(trigBI.Branch); err != nil {
			return nil, err
		}
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
func (m *Master) isTriggered(ctx context.Context, txnCtx *txncontext.TransactionContext, t *pfs.Trigger, oldHead, newHead *pfs.CommitInfo) (bool, error) {
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
			if ci.ParentCommit == nil {
				// TODO: We need a better mechanism for identifying the empty commits we create.
				if ci.Origin.Kind != pfs.OriginKind_AUTO {
					commits++
				}
				break
			}
			if oldHead.Commit.Id == ci.Commit.Id {
				break
			}
			var err error
			ci, err = m.resolveCommitInfo(ctx, txnCtx.SqlTx, ci.ParentCommit)
			if err != nil {
				return false, err
			}
			commits++
		}
		merge(commits == t.Commits)
	}
	return result, nil
}
