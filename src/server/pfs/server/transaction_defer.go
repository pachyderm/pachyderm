package server

import (
	"github.com/pachyderm/pachyderm/src/client/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
)

// TransactionDefer is an object that is used to defer certain cleanup tasks
// until the end of a transaction.  The transactionenv package provides the
// interface for this and will call the Run function at the end of a
// transaction.
type TransactionDefer struct {
	d   *driver
	stm col.STM

	// Branches to propagate when the transaction completes
	propagateBranches []*pfs.Branch

	// Scratch spaces in etcd that need to be deleted
	scratchCommits []*pfs.Commit
}

func (a *apiServer) NewTransactionDefer(stm col.STM) txnenv.PfsTransactionDefer {
	return &TransactionDefer{
		d:   a.driver,
		stm: stm,
	}
}

// Run performs any final tasks and cleanup tasks in the STM, such as
// propagating branches or deleting scratch space
func (t *TransactionDefer) Run() error {
	err := t.d.propagateCommits(t.stm, t.propagateBranches)
	if err != nil {
		return err
	}

	return t.d.deleteScratches(t.stm, t.scratchCommits)
}

// PropagateBranch marks a branch as needing propagation once the transaction
// successfully ends.  This will be performed by the Run function.
func (t *TransactionDefer) PropagateBranch(branch *pfs.Branch) {
	t.propagateBranches = append(t.propagateBranches, commit)
}

// DeleteScratch marks the commit scratch space as needing deletion once the
// transaction successfully ends.  This will be performed by the Run function.
func (t *TransactionDefer) DeleteScratch(commit *pfs.Commit) {
	t.scratchCommits = append(t.scratchCommits, commit)
}
