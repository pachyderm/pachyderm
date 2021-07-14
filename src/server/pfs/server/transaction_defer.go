package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// Propagater is an object that is used to propagate PFS branches at the end of
// a transaction.  The transactionenv package provides the interface for this
// and will call the Run function at the end of a transaction.
type Propagater struct {
	d      *driver
	txnCtx *txncontext.TransactionContext

	// Branches that were modified (new commits or head commit was moved to an old commit)
	branches map[string]*pfs.Branch
}

func (a *apiServer) NewPropagater(txnCtx *txncontext.TransactionContext) txncontext.PfsPropagater {
	return &Propagater{
		d:        a.driver,
		txnCtx:   txnCtx,
		branches: map[string]*pfs.Branch{},
	}
}

// PropagateBranch marks a branch as needing propagation once the transaction
// successfully ends.  This will be performed by the Run function.
func (t *Propagater) PropagateBranch(branch *pfs.Branch) error {
	if branch == nil {
		return errors.New("cannot propagate nil branch")
	}
	t.branches[pfsdb.BranchKey(branch)] = branch
	return nil
}

// Run performs any final tasks and cleanup tasks in the transaction, such as
// propagating branches
func (t *Propagater) Run() error {
	branches := make([]*pfs.Branch, 0, len(t.branches))
	for _, branch := range t.branches {
		branches = append(branches, branch)
	}
	return t.d.propagateBranches(t.txnCtx, branches)
}
