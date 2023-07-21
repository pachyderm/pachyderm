package transactionenv

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
)

// PfsWrites is an interface providing a wrapper for each operation that
// may be appended to a transaction through PFS.  Each call may either
// directly run the request through PFS or append it to the active transaction,
// depending on if there is an active transaction in the client context.
type PfsWrites interface {
	CreateRepo(*pfs.CreateRepoRequest) error
	DeleteRepo(*pfs.DeleteRepoRequest) error

	StartCommit(*pfs.StartCommitRequest) (*pfs.Commit, error)
	FinishCommit(*pfs.FinishCommitRequest) error
	SquashCommitSet(*pfs.SquashCommitSetRequest) error

	CreateBranch(*pfs.CreateBranchRequest) error
	DeleteBranch(*pfs.DeleteBranchRequest) error
}

// PpsWrites is an interface providing a wrapper for each operation that
// may be appended to a transaction through PPS.  Each call may either
// directly run the request through PPS or append it to the active transaction,
// depending on if there is an active transaction in the client context.
type PpsWrites interface {
	StopJob(*pps.StopJobRequest) error
	UpdateJobState(*pps.UpdateJobStateRequest) error
	CreatePipeline(*pps.CreatePipelineRequest) error
}

// AuthWrites is an interface providing a wrapper for each operation that
// may be appended to a transaction through the Auth server.  Each call may
// either directly run the request through Auth or append it to the active
// transaction, depending on if there is an active transaction in the client
// context.
type AuthWrites interface {
	ModifyRoleBinding(*auth.ModifyRoleBindingRequest) (*auth.ModifyRoleBindingResponse, error)
	DeleteRoleBinding(*auth.Resource) error
}

// TransactionServer is an interface used by other servers to append a request
// to an existing transaction.
type TransactionServer interface {
	AppendRequest(
		context.Context,
		*transaction.Transaction,
		*transaction.TransactionRequest,
	) (*transaction.TransactionResponse, error)
}

// TransactionEnv contains the APIServer instances for each subsystem that may
// be involved in running transactions so that they can make calls to each other
// without leaving the context of a transaction.  This is a separate object
// because there are cyclic dependencies between APIServer instances.
type TransactionEnv struct {
	serviceEnv serviceenv.ServiceEnv
	txnServer  TransactionServer
	initDone   chan struct{}
}

func New() *TransactionEnv {
	return &TransactionEnv{
		initDone: make(chan struct{}),
	}
}

// Initialize stores the references to APIServer instances in the TransactionEnv
func (tnxEnv *TransactionEnv) Initialize(
	serviceEnv serviceenv.ServiceEnv,
	txnServer TransactionServer,
) {
	tnxEnv.serviceEnv = serviceEnv
	tnxEnv.txnServer = txnServer
	close(tnxEnv.initDone)
}

// Transaction is an interface to unify the code that may either perform an
// action directly or append an action to an existing transaction (depending on
// if there is an active transaction in the client context metadata).  There
// are two implementations of this interface:
//
//	directTransaction: all operations will be run directly through the relevant
//	  server, all inside the same STM.
//	appendTransaction: all operations will be appended to the active transaction
//	  which will then be dryrun so that the response for the operation can be
//	  returned.  Each operation that is appended will do a new dryrun, so this
//	  isn't as efficient as it could be.
type Transaction interface {
	PfsWrites
	PpsWrites
	AuthWrites
}

type directTransaction struct {
	txnEnv *TransactionEnv
	txnCtx *txncontext.TransactionContext
	ctx    context.Context
}

// NewDirectTransaction is a helper function to instantiate a directTransaction
// object.  It is exposed so that the transaction API server can run a direct
// transaction even though there is an active transaction in the context (which
// is why it cannot use `WithTransaction`).
func NewDirectTransaction(ctx context.Context, txnEnv *TransactionEnv, txnCtx *txncontext.TransactionContext) Transaction {
	return &directTransaction{
		ctx:    ctx,
		txnEnv: txnEnv,
		txnCtx: txnCtx,
	}
}

func (t *directTransaction) CreateRepo(original *pfs.CreateRepoRequest) error {
	req := proto.Clone(original).(*pfs.CreateRepoRequest)
	return errors.EnsureStack(t.txnEnv.serviceEnv.PfsServer().CreateRepoInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) DeleteRepo(original *pfs.DeleteRepoRequest) error {
	req := proto.Clone(original).(*pfs.DeleteRepoRequest)
	return errors.EnsureStack(t.txnEnv.serviceEnv.PfsServer().DeleteRepoInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) StartCommit(original *pfs.StartCommitRequest) (*pfs.Commit, error) {
	req := proto.Clone(original).(*pfs.StartCommitRequest)
	res, err := t.txnEnv.serviceEnv.PfsServer().StartCommitInTransaction(t.ctx, t.txnCtx, req)
	return res, errors.EnsureStack(err)
}

func (t *directTransaction) FinishCommit(original *pfs.FinishCommitRequest) error {
	req := proto.Clone(original).(*pfs.FinishCommitRequest)
	return errors.EnsureStack(t.txnEnv.serviceEnv.PfsServer().FinishCommitInTransaction(t.txnCtx, req))
}

func (t *directTransaction) SquashCommitSet(original *pfs.SquashCommitSetRequest) error {
	req := proto.Clone(original).(*pfs.SquashCommitSetRequest)
	return errors.EnsureStack(t.txnEnv.serviceEnv.PfsServer().SquashCommitSetInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) CreateBranch(original *pfs.CreateBranchRequest) error {
	req := proto.Clone(original).(*pfs.CreateBranchRequest)
	return errors.EnsureStack(t.txnEnv.serviceEnv.PfsServer().CreateBranchInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) DeleteBranch(original *pfs.DeleteBranchRequest) error {
	req := proto.Clone(original).(*pfs.DeleteBranchRequest)
	return errors.EnsureStack(t.txnEnv.serviceEnv.PfsServer().DeleteBranchInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) StopJob(original *pps.StopJobRequest) error {
	req := proto.Clone(original).(*pps.StopJobRequest)
	return errors.EnsureStack(t.txnEnv.serviceEnv.PpsServer().StopJobInTransaction(t.txnCtx, req))
}

func (t *directTransaction) UpdateJobState(original *pps.UpdateJobStateRequest) error {
	req := proto.Clone(original).(*pps.UpdateJobStateRequest)
	return errors.EnsureStack(t.txnEnv.serviceEnv.PpsServer().UpdateJobStateInTransaction(t.txnCtx, req))
}

func (t *directTransaction) ModifyRoleBinding(original *auth.ModifyRoleBindingRequest) (*auth.ModifyRoleBindingResponse, error) {
	req := proto.Clone(original).(*auth.ModifyRoleBindingRequest)
	res, err := t.txnEnv.serviceEnv.AuthServer().ModifyRoleBindingInTransaction(t.txnCtx, req)
	return res, errors.EnsureStack(err)
}

func (t *directTransaction) CreatePipeline(original *pps.CreatePipelineRequest) error {
	req := proto.Clone(original).(*pps.CreatePipelineRequest)
	return errors.EnsureStack(t.txnEnv.serviceEnv.PpsServer().CreatePipelineInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) DeleteRoleBinding(original *auth.Resource) error {
	req := proto.Clone(original).(*auth.Resource)
	return errors.EnsureStack(t.txnEnv.serviceEnv.AuthServer().DeleteRoleBindingInTransaction(t.txnCtx, req))
}

type appendTransaction struct {
	ctx       context.Context
	activeTxn *transaction.Transaction
	txnEnv    *TransactionEnv
}

func newAppendTransaction(ctx context.Context, activeTxn *transaction.Transaction, txnEnv *TransactionEnv) Transaction {
	return &appendTransaction{
		ctx:       ctx,
		activeTxn: activeTxn,
		txnEnv:    txnEnv,
	}
}

func (t *appendTransaction) CreateRepo(req *pfs.CreateRepoRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{CreateRepo: req})
	return errors.EnsureStack(err)
}

func (t *appendTransaction) DeleteRepo(req *pfs.DeleteRepoRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{DeleteRepo: req})
	return errors.EnsureStack(err)
}

func (t *appendTransaction) StartCommit(req *pfs.StartCommitRequest) (*pfs.Commit, error) {
	res, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{StartCommit: req})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return res.Commit, nil
}

func (t *appendTransaction) FinishCommit(req *pfs.FinishCommitRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{FinishCommit: req})
	return errors.EnsureStack(err)
}

func (t *appendTransaction) SquashCommitSet(req *pfs.SquashCommitSetRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{SquashCommitSet: req})
	return errors.EnsureStack(err)
}

func (t *appendTransaction) CreateBranch(req *pfs.CreateBranchRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{CreateBranch: req})
	return errors.EnsureStack(err)
}

func (t *appendTransaction) DeleteBranch(req *pfs.DeleteBranchRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{DeleteBranch: req})
	return errors.EnsureStack(err)
}

func (t *appendTransaction) StopJob(req *pps.StopJobRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{StopJob: req})
	return errors.EnsureStack(err)
}

func (t *appendTransaction) UpdateJobState(req *pps.UpdateJobStateRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{UpdateJobState: req})
	return errors.EnsureStack(err)
}

func (t *appendTransaction) CreatePipeline(req *pps.CreatePipelineRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{CreatePipeline: req})
	return errors.EnsureStack(err)
}

func (t *appendTransaction) ModifyRoleBinding(original *auth.ModifyRoleBindingRequest) (*auth.ModifyRoleBindingResponse, error) {
	panic("ModifyRoleBinding not yet implemented in transactions")
}

func (t *appendTransaction) DeleteRoleBinding(original *auth.Resource) error {
	panic("DeleteRoleBinding not yet implemented in transactions")
}

// WithTransaction will call the given callback with a txnenv.Transaction
// object, which is instantiated differently based on if an active
// transaction is present in the RPC context.  If an active transaction is
// present, any calls into the Transaction are first dry-run then appended
// to the transaction.  If there is no active transaction, the request will be
// run directly through the selected server.
func (env *TransactionEnv) WithTransaction(ctx context.Context, cb func(Transaction) error) error {
	activeTxn, err := client.GetTransaction(ctx)
	if err != nil {
		return err
	}
	if activeTxn != nil {
		appendTxn := newAppendTransaction(ctx, activeTxn, env)
		return cb(appendTxn)
	}
	return env.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		directTxn := NewDirectTransaction(ctx, env, txnCtx)
		return cb(directTxn)
	})

}

func (env *TransactionEnv) attemptTx(ctx context.Context, sqlTx *pachsql.Tx, cb func(*txncontext.TransactionContext) error) error {
	txnCtx, err := txncontext.New(ctx, sqlTx, env.serviceEnv.AuthServer())
	if err != nil {
		return err
	}
	if env.serviceEnv.PfsServer() != nil {
		txnCtx.PfsPropagater = env.serviceEnv.PfsServer().NewPropagater(txnCtx)
	}
	if env.serviceEnv.PpsServer() != nil {
		txnCtx.PpsPropagater = env.serviceEnv.PpsServer().NewPropagater(txnCtx)
		txnCtx.PpsJobStopper = env.serviceEnv.PpsServer().NewJobStopper(txnCtx)
		txnCtx.PpsJobFinisher = env.serviceEnv.PpsServer().NewJobFinisher(txnCtx)
	}

	err = cb(txnCtx)
	if err != nil {
		return err
	}
	return txnCtx.Finish()
}

func (env *TransactionEnv) waitReady(ctx context.Context) error {
	select {
	case <-env.initDone:
		return nil
	case <-ctx.Done():
		return errors.EnsureStack(context.Cause(ctx))
	}
}

// WithWriteContext will call the given callback with a txncontext.TransactionContext
// which can be used to perform reads and writes on the current cluster state.
func (env *TransactionEnv) WithWriteContext(ctx context.Context, cb func(*txncontext.TransactionContext) error) error {
	if err := env.waitReady(ctx); err != nil {
		return err
	}
	return dbutil.WithTx(ctx, env.serviceEnv.GetDBClient(), func(ctx context.Context, sqlTx *pachsql.Tx) error {
		return env.attemptTx(ctx, sqlTx, cb)
	})
}

// WithReadContext will call the given callback with a txncontext.TransactionContext
// which can be used to perform reads of the current cluster state. If the
// transaction is used to perform any writes, they will be silently discarded.
func (env *TransactionEnv) WithReadContext(ctx context.Context, cb func(*txncontext.TransactionContext) error) error {
	if err := env.waitReady(ctx); err != nil {
		return err
	}
	return col.NewDryrunSQLTx(ctx, env.serviceEnv.GetDBClient(), func(sqlTx *pachsql.Tx) error {
		return env.attemptTx(ctx, sqlTx, cb)
	})
}

// PreTxOps defines what operations to run related to the transaction, but before the physical database
// transaction is opened. If doing any I/O as a part of a transaction is necessary, this is the place for it.
//
// NOTES:
// - PreTxOps may be called multiple times for a Pachyderm Transaction and should therefore be idempotent
// - in most cases some background job will also be necessary to cleanup resources created here
func (env *TransactionEnv) PreTxOps(ctx context.Context, reqs []*transaction.TransactionRequest) error {
	for _, r := range reqs {
		if r.CreatePipeline != nil {
			if r.CreatePipeline.Determined != nil {
				if err := env.serviceEnv.PpsServer().CreateDetPipelineSideEffects(ctx, r.CreatePipeline.Pipeline, r.CreatePipeline.Determined.Workspaces); err != nil {
					return errors.Wrap(err, "apply determined pipeline side effects")
				}
			}
		}
	}
	return nil
}
