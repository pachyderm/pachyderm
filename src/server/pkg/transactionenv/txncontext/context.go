package txncontext

import (
	"context"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

// PfsPropagater is the interface that PFS implements to propagate commits at
// the end of a transaction.  It is defined here to avoid a circular dependency.
type PfsPropagater interface {
	PropagateCommit(branch *pfs.Branch, isNewCommit bool) error
	Run() error
}

// PipelineCommitFinisher is an interface to facilitate finishing pipeline commits
// at the end of a transaction
type PipelineCommitFinisher interface {
	FinishPipelineCommits(branch *pfs.Branch) error
	Run() error
}

// TransactionContext is a helper type to encapsulate the state for a given
// set of operations being performed in the Pachyderm API.  When a new
// transaction is started, a context will be created for it containing these
// objects, which will be threaded through to every API call:
//   ctx: the client context which initiated the operations being performed
//   pachClient: the APIClient associated with the client context ctx
//   stm: the object that controls transactionality with etcd.  This is to ensure
//     that all reads and writes are consistent until changes are committed.
//   txnEnv: a struct containing references to each API server, it can be used
//     to make calls to other API servers (e.g. checking auth permissions)
//   pfsDefer: an interface for ensuring certain PFS cleanup tasks are performed
//     properly (and deduped) at the end of the transaction.
type TransactionContext struct {
	ClientContext  context.Context
	Client         *client.APIClient
	Stm            col.STM
	PfsPropagater  PfsPropagater
	CommitFinisher PipelineCommitFinisher
}

// PropagateCommit saves a branch to be propagated at the end of the transaction
// (if all operations complete successfully).  This is used to batch together
// propagations and dedupe downstream commits in PFS.
func (t *TransactionContext) PropagateCommit(branch *pfs.Branch, isNewCommit bool) error {
	return t.PfsPropagater.PropagateCommit(branch, isNewCommit)
}

func (t *TransactionContext) Finish() error {
	if t.CommitFinisher != nil {
		if err := t.CommitFinisher.Run(); err != nil {
			return err
		}
	}
	return t.PfsPropagater.Run()
}

// FinishPipelineCommits saves a pipeline output branch to have its commits
// finished at the end of the transaction
func (t *TransactionContext) FinishPipelineCommits(branch *pfs.Branch) error {
	if t.CommitFinisher != nil {
		return t.CommitFinisher.FinishPipelineCommits(branch)
	}
	return nil
}
