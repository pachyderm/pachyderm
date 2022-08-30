/*
Package txncontext manages the life cycle of transactions.
It provides all the resources transactional operations generally require, and triggers deferred transaction steps.

The transaction defer model is used to maintain various pachyderm invariants that have performance implications,
unlike branch provenance lists which are updated after every individual operation.
Rather than e.g. create new jobs after every individual state change, these operations happen once at the end.
In most cases, callers must mark specific objects (branches or jobs) for processing during [TransactionContext.Finish].
*/
package txncontext

import (
	"context"
	"time"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// TransactionContext is a helper type to encapsulate the state for a given
// set of operations being performed in the Pachyderm API.  When a new
// transaction is started, a context will be created for it containing these
// objects, which will be threaded through to every API call.
// Note: TransactionContext deliberately does not expose a context object.
// This is to discourage interacting with external services or invoking API endpoints, during a transaction,
// as these operations can lead to impossible-to-satisfy requests or accidental repetition of non-idempotent operations.
type TransactionContext struct {
	// username is the user running the transaction.
	// It is computed once initially, and used to respond to subsequent WhoAmI requests.
	username string
	// SqlTx is the ongoing database transaction.
	SqlTx *pachsql.Tx
	// CommitSetID is the ID of the CommitSet corresponding to PFS changes in this transaction.
	// It is initialized to a random ID, unless overwritten (which should be done with great care).
	CommitSetID string
	// Timestamp is the canonical timestamp to be used for writes in this transaction.
	Timestamp *types.Timestamp
	// PfsPropagater applies commits at the end of the transaction.
	PfsPropagater PfsPropagater
	// PpsPropagater starts Jobs in any pipelines that have new output commits at the end of the transaction.
	PpsPropagater PpsPropagater
	// PpsJobStopper stops Jobs in any pipelines that are associated with a removed commit set
	PpsJobStopper PpsJobStopper
	// PpsJobFinisher changes Job states to SUCCESS or FAILURE based on the results of commit finishing
	PpsJobFinisher PpsJobFinisher
}

type identifier interface {
	WhoAmI(context.Context, *auth.WhoAmIRequest) (*auth.WhoAmIResponse, error)
}

func New(ctx context.Context, sqlTx *pachsql.Tx, authServer identifier) (*TransactionContext, error) {
	var username string
	// check auth once now so that we can refer to it later
	if authServer != nil {
		if me, err := authServer.WhoAmI(ctx, &auth.WhoAmIRequest{}); err != nil && !auth.IsErrNotActivated(err) {
			return nil, errors.EnsureStack(err)
		} else if err == nil {
			username = me.Username
		}
	}
	var currTime time.Time
	if err := sqlTx.GetContext(ctx, &currTime, "select CURRENT_TIMESTAMP as Timestamp"); err != nil {
		return nil, errors.EnsureStack(err)
	}

	ts, err := types.TimestampProto(currTime)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting transaction timestamp")
	}
	return &TransactionContext{
		SqlTx:       sqlTx,
		CommitSetID: uuid.NewWithoutDashes(),
		Timestamp:   ts,
		username:    username,
	}, nil
}

func (t *TransactionContext) WhoAmI() (*auth.WhoAmIResponse, error) {
	if t.username == "" {
		return nil, auth.ErrNotActivated
	}
	return &auth.WhoAmIResponse{Username: t.username}, nil
}

// PropagateJobs notifies PPS that there are new commits in the transaction's
// commitset that need jobs to be created at the end of the transaction
// transaction (if all operations complete successfully).
func (t *TransactionContext) PropagateJobs() {
	t.PpsPropagater.PropagateJobs()
}

// StopJobs notifies PPS that some commits have been removed and the jobs
// associated with them should be stopped.
func (t *TransactionContext) StopJobs(commitset *pfs.CommitSet) {
	t.PpsJobStopper.StopJobs(commitset)
}

func (t *TransactionContext) FinishJob(commitInfo *pfs.CommitInfo) {
	t.PpsJobFinisher.FinishJob(commitInfo)
}

// PropagateBranch saves a branch to be propagated at the end of the transaction
// (if all operations complete successfully).  This is used to batch together
// propagations and dedupe downstream commits in PFS.
func (t *TransactionContext) PropagateBranch(branch *pfs.Branch) error {
	return errors.EnsureStack(t.PfsPropagater.PropagateBranch(branch))
}

// DeleteBranch removes a branch from the list of branches to propagate, if
// it is present.
func (t *TransactionContext) DeleteBranch(branch *pfs.Branch) {
	t.PfsPropagater.DeleteBranch(branch)
}

// Finish applies the deferred logic in the pfsPropagator and ppsPropagator to
// the transaction
func (t *TransactionContext) Finish() error {
	if t.PfsPropagater != nil {
		if err := t.PfsPropagater.Run(); err != nil {
			return errors.EnsureStack(err)
		}
	}
	if t.PpsPropagater != nil {
		if err := t.PpsPropagater.Run(); err != nil {
			return errors.EnsureStack(err)
		}
	}
	if t.PpsJobStopper != nil {
		if err := t.PpsJobStopper.Run(); err != nil {
			return errors.EnsureStack(err)
		}
	}
	if t.PpsJobFinisher != nil {
		if err := t.PpsJobFinisher.Run(); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

// PfsPropagater is the interface that PFS implements to propagate commits at
// the end of a transaction.  It is defined here to avoid a circular dependency.
type PfsPropagater interface {
	PropagateBranch(branch *pfs.Branch) error
	DeleteBranch(branch *pfs.Branch)
	Run() error
}

// PpsPropagater is the interface that PPS implements to start jobs at the end
// of a transaction.  It is defined here to avoid a circular dependency.
type PpsPropagater interface {
	PropagateJobs()
	Run() error
}

// PpsJobStopper is the interface that PPS implements to stop jobs of deleted
// commitsets at the end of a transaction.  It is defined here to avoid a
// circular dependency.
type PpsJobStopper interface {
	StopJobs(commitset *pfs.CommitSet)
	Run() error
}

type PpsJobFinisher interface {
	FinishJob(commitInfo *pfs.CommitInfo)
	Run() error
}
