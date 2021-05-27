package txncontext

import (
	"context"

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
	// PfsPropagater applies commits at the end of the transaction.
	PfsPropagater PfsPropagater
	// CommitFinisher finishes commits for a pipeline at the end of a transaction
	CommitFinisher PipelineCommitFinisher
}

// PropagateCommit saves a branch to be propagated at the end of the transaction
// (if all operations complete successfully).  This is used to batch together
// propagations and dedupe downstream commits in PFS.
func (t *TransactionContext) PropagateBranch(branch *pfs.Branch) error {
	return t.PfsPropagater.PropagateBranch(branch)
}

// Finish applies the commitFinisher and pfsPropagator, is set
func (t *TransactionContext) Finish() error {
	if t.CommitFinisher != nil {
		if err := t.CommitFinisher.Run(); err != nil {
			return err
		}
	}
	if t.PfsPropagater != nil {
		return t.PfsPropagater.Run()
	}
	return nil
}

// FinishPipelineCommits saves a pipeline output branch to have its commits
// finished at the end of the transaction
func (t *TransactionContext) FinishPipelineCommits(branch *pfs.Branch) error {
	if t.CommitFinisher != nil {
		return t.CommitFinisher.FinishPipelineCommits(branch)
	}
	return nil
}

// PfsPropagater is the interface that PFS implements to propagate commits at
// the end of a transaction.  It is defined here to avoid a circular dependency.
type PfsPropagater interface {
	PropagateBranch(branch *pfs.Branch) error
	Run() error
}

// PipelineCommitFinisher is an interface to facilitate finishing pipeline commits
// at the end of a transaction
type PipelineCommitFinisher interface {
	FinishPipelineCommits(branch *pfs.Branch) error
	Run() error
}
