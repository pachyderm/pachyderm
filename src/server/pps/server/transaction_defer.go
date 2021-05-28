package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// Propagater is an object that is used to propagate PFS branches at the end of
// a transaction.  The transactionenv package provides the interface for this
// and will call the Run function at the end of a transaction.
type Propagater struct {
	a      *apiServer
	txnCtx *txncontext.TransactionContext

	// Commitsets that were modified in this transaction
	commitsets []*pfs.StoredCommitset
}

func (a *apiServer) NewPropagater(txnCtx *txncontext.TransactionContext) txncontext.PpsPropagater {
	return &Propagater{
		a:          a,
		txnCtx:     txnCtx,
		commitsets: []*pfs.StoredCommitset{},
	}
}

// PropagateCommitset notifies PPS that a commitset has been modified in the
// transaction, and any jobs will be created for it at the end of the
// transaction  This will be performed by the Run function.
func (t *Propagater) PropagateCommitset(commitset *pfs.StoredCommitset) error {
	if commitset == nil {
		return errors.New("cannot propagate nil commitset")
	}
	t.commitsets = append(t.commitsets, commitset)
	return nil
}

// Run creates any jobs for the modified Commitsets
func (t *Propagater) Run() error {
	for _, commitset := range t.commitsets {
		if err := t.a.propagateJobs(txnCtx, commitset); err != nil {
			return err
		}
	}
	return nil
}
