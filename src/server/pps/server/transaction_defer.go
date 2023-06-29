package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
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
func (t *Propagater) Run() error {
	if t.notified {
		return t.a.propagateJobs(t.txnCtx)
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
}

func (a *apiServer) NewJobStopper(txnCtx *txncontext.TransactionContext) txncontext.PpsJobStopper {
	return &JobStopper{
		a:          a,
		txnCtx:     txnCtx,
		commitsets: []*pfs.CommitSet{},
	}
}

// PropagateJobs notifies PPS that a commitset has been modified in the
// transaction, and any jobs will be created for it at the end of the
// transaction.  This will be performed by the Run function.
func (t *JobStopper) StopJobs(commitset *pfs.CommitSet) {
	t.commitsets = append(t.commitsets, commitset)
}

// Run stops any jobs for the removed CommitSets
func (t *JobStopper) Run() error {
	if len(t.commitsets) > 0 {
		for _, commitset := range t.commitsets {
			jobInfo := &pps.JobInfo{}
			if err := t.a.jobs.ReadWrite(t.txnCtx.SqlTx).GetByIndex(ppsdb.JobsJobSetIndex, commitset.Id, jobInfo, col.DefaultOptions(), func(string) error {
				return t.a.stopJob(t.txnCtx, jobInfo.Job, "output commit removed")
			}); err != nil {
				return errors.EnsureStack(err)
			}
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

func (jf *JobFinisher) Run() error {
	if len(jf.commitInfos) > 0 {
		pipelines := jf.a.pipelines.ReadWrite(jf.txnCtx.SqlTx)
		jobs := jf.a.jobs.ReadWrite(jf.txnCtx.SqlTx)
		for _, commitInfo := range jf.commitInfos {
			if commitInfo.Commit.Repo.Type != pfs.UserRepoType {
				continue
			}
			jobKey := ppsdb.JobKey(client.NewJob(commitInfo.Commit.Repo.Project.GetName(), commitInfo.Commit.Repo.Name, commitInfo.Commit.Id))
			jobInfo := &pps.JobInfo{}
			if err := jobs.Get(jobKey, jobInfo); err != nil {
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
