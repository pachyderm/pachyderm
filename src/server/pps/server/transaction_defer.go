package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
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

// Run creates any jobs for the modified Commitsets
func (t *Propagater) Run() error {
	if t.notified {
		return t.a.propagateJobs(t.txnCtx)
	}
	return nil
}
