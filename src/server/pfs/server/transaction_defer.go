package server

import (
	"github.com/pachyderm/pachyderm/src/client"
	col "github.com/pachyderm/pachyderm/src/internal/collection"
	"github.com/pachyderm/pachyderm/src/internal/errors"
	txnenv "github.com/pachyderm/pachyderm/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/src/pfs"
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
		return errors.Errorf("cannot propagate nil branch")
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

// PipelineFinisher closes any open commits on a pipeline output branch,
// preventing any in-progress jobs from pipelines about to be replaced.
// The Run method is called at the end of any transaction
type PipelineFinisher struct {
	d      *driver
	txnCtx *txnenv.TransactionContext

	// pipeline output branches to finish commits on
	branches []*pfs.Branch
}

func (a *apiServer) NewPipelineFinisher(txnCtx *txnenv.TransactionContext) txnenv.PipelineCommitFinisher {
	return &PipelineFinisher{
		d:      a.driver,
		txnCtx: txnCtx,
	}
}

// FinishPipelineCommits marks a branch as belonging to a pipeline which is being modified and thus
// needs old open commits finished
func (f *PipelineFinisher) FinishPipelineCommits(branch *pfs.Branch) error {
	if branch == nil || branch.Repo == nil {
		return errors.Errorf("cannot finish commits on nil branch")
	}
	f.branches = append(f.branches, branch)
	return nil
}

// Run finishes any open commits on output branches of pipelines modified in the preceeding transaction
func (f *PipelineFinisher) Run() error {
	for _, branch := range f.branches {
		if err := f.d.listCommit(
			f.txnCtx.Client,
			branch.Repo,
			client.NewCommit(branch.Repo.Name, branch.Name), // to
			nil,   // from
			0,     // number
			false, // reverse
			func(commitInfo *pfs.CommitInfo) error {
				// finish all open commits on the branch
				if commitInfo.Finished != nil {
					return nil
				}
				return f.d.finishCommit(
					f.txnCtx,
					client.NewCommit(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID),
					"",
				)
			}); err != nil && !isNotFoundErr(err) {
			return err
		}
	}
	return nil
}
