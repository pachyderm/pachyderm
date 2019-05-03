package server

import (
	"github.com/pachyderm/pachyderm/src/client/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
)

type TransactionDefer struct {
	pfsServer *apiServer
	stm       col.STM

	// Branches to propagate when the transaction completes
	propagateBranches []*pfs.Branch

	// Scratch spaces in etcd that need to be deleted
	scratches []*pfs.Commit
}

func (a *apiServer) NewTransactionDefer(stm col.STM) txnenv.PfsTransactionDefer {
	return &TransactionDefer{
		pfsServer: a,
		stm:       stm,
	}
}

// Performs any final tasks and cleanup tasks in the STM, such as propagating
// branches or deleting scratch space
func (t *TransactionDefer) Run() error {
	return nil
}

func (t *TransactionDefer) PropagateBranch(branch *pfs.Branch) {
}

func (t *TransactionDefer) DeleteScratch(commit *pfs.Commit) {
}
