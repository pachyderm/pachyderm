package server

import (
	"context"

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

// DeleteBranch removes a branch from the list of those needing propagation
// if present.
func (t *Propagater) DeleteBranch(branch *pfs.Branch) {
	delete(t.branches, pfsdb.BranchKey(branch))
}

// Run performs any final tasks and cleanup tasks in the transaction, such as
// propagating branches
func (t *Propagater) Run(ctx context.Context) error {
	branches := make([]*pfs.Branch, 0, len(t.branches))
	for _, branch := range t.branches {
		branches = append(branches, branch)
	}
	if err := t.d.validateDAGStructure(ctx, t.txnCtx, branches); err != nil {
		return errors.Wrap(err, "validate DAG at end of transaction")
	}
	return t.d.propagateBranches(ctx, t.txnCtx, branches)
}
