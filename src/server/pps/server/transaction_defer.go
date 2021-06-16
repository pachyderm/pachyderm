package server

import (
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
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
			if err := t.a.jobs.ReadWrite(t.txnCtx.SqlTx).GetByIndex(ppsdb.JobsJobsetIndex, commitset.ID, jobInfo, col.DefaultOptions(), func(string) error {
				return t.a.stopJob(t.txnCtx, jobInfo.Job, "output commit removed")
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
