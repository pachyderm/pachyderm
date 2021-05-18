package server

import (
	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// Propagater is an object that is used to propagate PFS branches at the end of
// a transaction.  The transactionenv package provides the interface for this
// and will call the Run function at the end of a transaction.
type Propagater struct {
	d     *driver
	sqlTx *sqlx.Tx
	job   *pfs.Job

	// Branches to propagate when the transaction completes
	branches    []*pfs.Branch
	isNewCommit bool
}

func (a *apiServer) NewPropagater(sqlTx *sqlx.Tx, job *pfs.Job) txncontext.PfsPropagater {
	return &Propagater{
		d:     a.driver,
		sqlTx: sqlTx,
		job:   job,
	}
}

// PropagateCommit marks a branch as needing propagation once the transaction
// successfully ends.  This will be performed by the Run function.
func (t *Propagater) PropagateCommit(branch *pfs.Branch, isNewCommit bool) error {
	if branch == nil {
		return errors.Errorf("cannot propagate nil branch")
	}
	t.branches = append(t.branches, branch)
	t.isNewCommit = isNewCommit
	return nil
}

// Run performs any final tasks and cleanup tasks in the transaction, such as
// propagating branches
func (t *Propagater) Run() error {
	return t.d.propagateCommits(t.sqlTx, t.job, t.branches, t.isNewCommit)
}

// PipelineFinisher closes any open commits on a pipeline output branch,
// preventing any in-progress jobs from pipelines about to be replaced.
// The Run method is called at the end of any transaction
type PipelineFinisher struct {
	d      *driver
	txnCtx *txncontext.TransactionContext

	// pipeline output branches to finish commits on
	branches []*pfs.Branch
}

func (a *apiServer) NewPipelineFinisher(txnCtx *txncontext.TransactionContext) txncontext.PipelineCommitFinisher {
	return &PipelineFinisher{
		d:      a.driver,
		txnCtx: txnCtx,
	}
}

// FinishPipelineCommits marks a branch as belonging to a pipeline which is being modified and thus
// needs old open commits finished
// TODO: this should really be in a seperate transaction-defer for PPS code
func (f *PipelineFinisher) FinishPipelineCommits(branch *pfs.Branch) error {
	if branch == nil || branch.Repo == nil {
		return errors.Errorf("cannot finish commits on nil branch")
	}
	f.branches = append(f.branches, branch)
	return nil
}

// Run finishes any open commits on output branches of pipelines modified in the preceeding transaction
func (f *PipelineFinisher) Run() error {
	for _, branch := range f.branches {
		// TODO: I believe this takes linear time based on the number of commits in a pipeline
		// the listCommit is also outside of the transaction, so new commits may be
		// made that can't be found in the transaction (hopefully this causes a
		// transaction conflict, though)
		if err := f.d.listCommit(
			f.txnCtx.ClientContext,
			branch.Repo,
			branch.NewCommit(""), // to
			nil,                  // from
			0,                    // number
			false,                // reverse
			func(commitInfo *pfs.CommitInfo) error {
				return f.d.env.PpsServer().StopPipelineJobInTransaction(f.txnCtx, &pps.StopPipelineJobRequest{
					OutputCommit: commitInfo.Commit,
				})
			}); err != nil && !isNotFoundErr(err) {
			return err
		}
	}
	return nil
}
