package server

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
)

// Propagater is an object that is used to propagate PFS branches at the end of
// a transaction.  The transactionenv package provides the interface for this
// and will call the Run function at the end of a transaction.
type Propagater struct {
	d   *driver
	stm col.STM

	// Branches to propagate when the transaction completes
	branches    []*pfs.Branch
	isNewCommit bool
}

func (a *apiServer) NewPropagater(stm col.STM) txnenv.PfsPropagater {
	return &Propagater{
		d:   a.driver,
		stm: stm,
	}
}

// PropagateCommit marks a branch as needing propagation once the transaction
// successfully ends.  This will be performed by the Run function.
func (t *Propagater) PropagateCommit(branch *pfs.Branch, isNewCommit bool) error {
	if branch == nil {
		return fmt.Errorf("cannot propagate nil branch")
	}
	t.branches = append(t.branches, branch)
	t.isNewCommit = isNewCommit
	return nil
}

// Run performs any final tasks and cleanup tasks in the STM, such as
// propagating branches
func (t *Propagater) Run() error {
	return t.d.propagateCommits(t.stm, t.branches, t.isNewCommit)
}
