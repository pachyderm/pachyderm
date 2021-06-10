package txncontext

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// TransactionContext is a helper type to encapsulate the state for a given
// set of operations being performed in the Pachyderm API.  When a new
// transaction is started, a context will be created for it containing these
// objects, which will be threaded through to every API call:
type TransactionContext struct {
	// ClientContext is the incoming context.Context for the request.
	ClientContext context.Context
	// SqlTx is the ongoing database transaction.
	SqlTx *sqlx.Tx
	// CommitsetID is the ID of the Commitset corresponding to PFS changes in this transaction.
	CommitsetID string
	// Timestamp is the canonical timestamp to be used for writes in this transaction.
	Timestamp *types.Timestamp
	// PfsPropagater applies commits at the end of the transaction.
	PfsPropagater PfsPropagater
	// PpsPropagater starts Jobs in any pipelines that have new output commits at the end of the transaction.
	PpsPropagater PpsPropagater
	// PpsJobStopper stops Jobs in any pipelines that are associated with a removed commitset
	PpsJobStopper PpsJobStopper
}

// PropagateJobs notifies PPS that there are new commits in the transaction's
// commitset that need jobs to be created at the end of the transaction
// transaction (if all operations complete successfully).
func (t *TransactionContext) PropagateJobs() {
	t.PpsPropagater.PropagateJobs()
}

// StopJobs notifies PPS that some commits have been removed and the jobs
// associated with them should be stopped.
func (t *TransactionContext) StopJobs(commitset *pfs.Commitset) {
	t.PpsJobStopper.StopJobs(commitset)
}

// PropagateBranch saves a branch to be propagated at the end of the transaction
// (if all operations complete successfully).  This is used to batch together
// propagations and dedupe downstream commits in PFS.
func (t *TransactionContext) PropagateBranch(branch *pfs.Branch) error {
	return t.PfsPropagater.PropagateBranch(branch)
}

// Finish applies the deferred logic in the pfsPropagator and ppsPropagator to
// the transaction
func (t *TransactionContext) Finish() error {
	if t.PfsPropagater != nil {
		if err := t.PfsPropagater.Run(); err != nil {
			return err
		}
	}
	if t.PpsPropagater != nil {
		if err := t.PpsPropagater.Run(); err != nil {
			return err
		}
	}
	return nil
}

// PfsPropagater is the interface that PFS implements to propagate commits at
// the end of a transaction.  It is defined here to avoid a circular dependency.
type PfsPropagater interface {
	PropagateBranch(branch *pfs.Branch) error
	Run() error
}

// PpsPropagater is the interface that PPS implements to start jobs at the end
// of a transaction.  It is defined here to avoid a circular dependency.
type PpsPropagater interface {
	PropagateJobs()
	Run() error
}

// PpsJobStopper is the interface that PPS implements to stop jobs of deleted
// commitsets at the end of a transaction.  It is defined here to avoid a
// circular dependency.
type PpsJobStopper interface {
	StopJobs(commitset *pfs.Commitset)
	Run() error
}
