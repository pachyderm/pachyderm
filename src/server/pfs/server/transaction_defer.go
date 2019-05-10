package server

import (
	"fmt"

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
}

func (a *apiServer) NewTransactionDefer(stm col.STM) txnenv.PfsTransactionDefer {
	return &TransactionDefer{
		d:   a.driver,
		stm: stm,
	}
}

// Run performs any final tasks and cleanup tasks in the STM, such as
// propagating branches
func (t *TransactionDefer) Run() error {
	return t.d.propagateCommits(t.stm, t.propagateBranches)
}

// PropagateCommit marks a branch as needing propagation once the transaction
// successfully ends.  This will be performed by the Run function.
func (t *TransactionDefer) PropagateCommit(branch *pfs.Branch) error {
	fmt.Printf("PropagateCommit: %s@%s\n", branch.Repo.Name, branch.Name)
	if branch == nil {
		return fmt.Errorf("cannot propagate nil branch")
	}
	t.propagateBranches = append(t.propagateBranches, branch)
	return nil
}
