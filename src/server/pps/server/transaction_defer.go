package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/proto"
)

// Propagater is an object that is used to create jobs in response to a new
// commitset in PFS.  The transactionenv package provides the interface for this
// and will call the Run function at the end of a transaction.
type Propagater struct {
	a        *apiServer
	txnCtx   *txncontext.TransactionContext
	notified bool
}

func (a *apiServer) NewPropagater(txnCtx *txncontext.TransactionContext) txncontext.PpsPropagater {
	return &Propagater{
		a:      a,
		txnCtx: txnCtx,
	}
}

// PropagateJobs notifies PPS that a commitset has been modified in the
// transaction, and any jobs will be created for it at the end of the
// transaction.  This will be performed by the Run function.
func (t *Propagater) PropagateJobs() {
	t.notified = true
}

// Run creates any jobs for the modified CommitSets
func (t *Propagater) Run(ctx context.Context) error {
	if t.notified {
		return t.a.propagateJobs(ctx, t.txnCtx)
	}
	return nil
}

// JobStopper is an object that is used to stop jobs in response to a commitset
// being removed by PFS.  The transactionenv package provides the interface for
// this and will call the Run function at the end of a transaction.
type JobStopper struct {
	a          *apiServer
	txnCtx     *txncontext.TransactionContext
	commitsets []*pfs.CommitSet
	commits    []*pfs.Commit
}

func (a *apiServer) NewJobStopper(txnCtx *txncontext.TransactionContext) txncontext.PpsJobStopper {
	return &JobStopper{
		a:      a,
		txnCtx: txnCtx,
	}
}

// StopJobSet notifies PPS that a commitset has been deleted in the
// transaction, and any jobs will be stopped for it at the end of the
// transaction. This will be performed by the Run function.
func (t *JobStopper) StopJobSet(commitset *pfs.CommitSet) {
	t.commitsets = append(t.commitsets, commitset)
}

// StopJob notifies PPS that a commit has been deleted in the
// transaction, and any jobs will be stopped for it at the end of the
// transaction. This will be performed by the Run function.
func (t *JobStopper) StopJob(commit *pfs.Commit) {
	t.commits = append(t.commits, commit)
}

// Run stops any jobs for the removed commitsets or commits.
func (t *JobStopper) Run(ctx context.Context) error {
	for _, commitset := range t.commitsets {
		jobInfo := &pps.JobInfo{}
		if err := t.a.jobs.ReadWrite(t.txnCtx.SqlTx).GetByIndex(ppsdb.JobsJobSetIndex, commitset.Id, jobInfo, col.DefaultOptions(), func(string) error {
			return t.a.stopJob(ctx, t.txnCtx, jobInfo.Job, "output commit removed")
		}); err != nil {
			return errors.EnsureStack(err)
		}
	}
	// Cache entries are preset for filtering out unecessary jobs when performing commitset lookups.
	cache := make(map[string]*pps.Job)
	for _, commit := range t.commits {
		cache[pfsdb.CommitKey(commit)] = nil
	}
	for _, commit := range t.commits {
		job := cache[pfsdb.CommitKey(commit)]
		if job == nil {
			jobInfo := &pps.JobInfo{}
			if err := t.a.jobs.ReadWrite(t.txnCtx.SqlTx).GetByIndex(ppsdb.JobsJobSetIndex, commit.Id, jobInfo, col.DefaultOptions(), func(string) error {
				if _, ok := cache[pfsdb.CommitKey(jobInfo.OutputCommit)]; ok {
					cache[pfsdb.CommitKey(jobInfo.OutputCommit)] = proto.Clone(jobInfo.Job).(*pps.Job)
				}
				return nil
			}); err != nil {
				return errors.EnsureStack(err)
			}
		}
		job = cache[pfsdb.CommitKey(commit)]
		if job == nil {
			continue
		}
		if err := t.a.stopJob(ctx, t.txnCtx, job, "output commit removed"); err != nil {
			return err
		}
	}
	return nil
}

type JobFinisher struct {
	a           *apiServer
	txnCtx      *txncontext.TransactionContext
	commitInfos []*pfs.CommitInfo
}

func (a *apiServer) NewJobFinisher(txnCtx *txncontext.TransactionContext) txncontext.PpsJobFinisher {
	return &JobFinisher{
		a:      a,
		txnCtx: txnCtx,
	}
}

func (jf *JobFinisher) FinishJob(commitInfo *pfs.CommitInfo) {
	jf.commitInfos = append(jf.commitInfos, commitInfo)
}

func (jf *JobFinisher) Run(ctx context.Context) error {
	if len(jf.commitInfos) > 0 {
		pipelines := jf.a.pipelines.ReadWrite(jf.txnCtx.SqlTx)
		jobs := jf.a.jobs.ReadWrite(jf.txnCtx.SqlTx)
		for _, commitInfo := range jf.commitInfos {
			if commitInfo.Commit.Repo.Type != pfs.UserRepoType {
				continue
			}
			jobKey := ppsdb.JobKey(client.NewJob(commitInfo.Commit.Repo.Project.GetName(), commitInfo.Commit.Repo.Name, commitInfo.Commit.Id))
			jobInfo := &pps.JobInfo{}
			if err := jobs.Get(ctx, jobKey, jobInfo); err != nil {
				// Commits in source repos will not have a job associated with them.
				if col.IsErrNotFound(err) {
					continue
				}
				return errors.EnsureStack(err)
			}
			if jobInfo.State != pps.JobState_JOB_FINISHING {
				return nil
			}
			state := pps.JobState_JOB_SUCCESS
			var reason string
			if commitInfo.Error != "" {
				state = pps.JobState_JOB_FAILURE
				reason = commitInfo.Error
			}
			if err := ppsutil.UpdateJobState(pipelines, jobs, jobInfo, state, reason); err != nil {
				return err
			}
		}
	}
	return nil
}
