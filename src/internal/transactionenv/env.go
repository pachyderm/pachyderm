package transactionenv

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
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
	DeleteRepo(*pfs.DeleteRepoRequest) (bool, error)

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
	CreatePipeline(*pps.CreatePipelineTransaction) error
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

type PFSBackend interface {
	NewPropagater(*txncontext.TransactionContext) txncontext.PfsPropagater
	CreateRepoInTransaction(context.Context, *txncontext.TransactionContext, *pfs.CreateRepoRequest) error
	DeleteRepoInTransaction(context.Context, *txncontext.TransactionContext, *pfs.DeleteRepoRequest) (bool, error)

	StartCommitInTransaction(context.Context, *txncontext.TransactionContext, *pfs.StartCommitRequest) (*pfs.Commit, error)
	FinishCommitInTransaction(context.Context, *txncontext.TransactionContext, *pfs.FinishCommitRequest) error
	SquashCommitSetInTransaction(context.Context, *txncontext.TransactionContext, *pfs.SquashCommitSetRequest) error

	CreateBranchInTransaction(context.Context, *txncontext.TransactionContext, *pfs.CreateBranchRequest) error
	DeleteBranchInTransaction(context.Context, *txncontext.TransactionContext, *pfs.DeleteBranchRequest) error
}

type PPSBackend interface {
	NewPropagater(*txncontext.TransactionContext) txncontext.PpsPropagater
	NewJobStopper(*txncontext.TransactionContext) txncontext.PpsJobStopper
	NewJobFinisher(*txncontext.TransactionContext) txncontext.PpsJobFinisher

	StopJobInTransaction(context.Context, *txncontext.TransactionContext, *pps.StopJobRequest) error
	UpdateJobStateInTransaction(context.Context, *txncontext.TransactionContext, *pps.UpdateJobStateRequest) error
	CreatePipelineInTransaction(context.Context, *txncontext.TransactionContext, *pps.CreatePipelineTransaction) error
	CreateDetPipelineSideEffects(context.Context, *pps.Pipeline, []string) error
}

type AuthBackend interface {
	ModifyRoleBindingInTransaction(*txncontext.TransactionContext, *auth.ModifyRoleBindingRequest) (*auth.ModifyRoleBindingResponse, error)
	DeleteRoleBindingInTransaction(*txncontext.TransactionContext, *auth.Resource) error
	WhoAmI(context.Context, *auth.WhoAmIRequest) (*auth.WhoAmIResponse, error)
}

// TransactionEnv contains the APIServer instances for each subsystem that may
// be involved in running transactions so that they can make calls to each other
// without leaving the context of a transaction.  This is a separate object
// because there are cyclic dependencies between APIServer instances.
type TransactionEnv struct {
	txnServer TransactionServer
	db        *sqlx.DB
	getPFS    func() PFSBackend
	getPPS    func() PPSBackend
	getAuth   func() AuthBackend

	initDone chan struct{}
}

func New() *TransactionEnv {
	return &TransactionEnv{
		initDone: make(chan struct{}),
	}
}

// Initialize stores the references to APIServer instances in the TransactionEnv
func (tnxEnv *TransactionEnv) Initialize(
	db *sqlx.DB,
	getAuth func() AuthBackend,
	getPFS func() PFSBackend,
	getPPS func() PPSBackend,
	txnServer TransactionServer,
) {
	tnxEnv.db = db
	tnxEnv.txnServer = txnServer
	tnxEnv.getPFS = getPFS
	tnxEnv.getPPS = getPPS
	tnxEnv.getAuth = getAuth
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
	return errors.EnsureStack(t.txnEnv.getPFS().CreateRepoInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) DeleteRepo(original *pfs.DeleteRepoRequest) (bool, error) {
	req := proto.Clone(original).(*pfs.DeleteRepoRequest)
	isRepoDeleted, err := t.txnEnv.getPFS().DeleteRepoInTransaction(t.ctx, t.txnCtx, req)
	if err != nil {
		return false, errors.EnsureStack(err)
	}
	return isRepoDeleted, nil
}

func (t *directTransaction) StartCommit(original *pfs.StartCommitRequest) (*pfs.Commit, error) {
	req := proto.Clone(original).(*pfs.StartCommitRequest)
	res, err := t.txnEnv.getPFS().StartCommitInTransaction(t.ctx, t.txnCtx, req)
	return res, errors.EnsureStack(err)
}

func (t *directTransaction) FinishCommit(original *pfs.FinishCommitRequest) error {
	req := proto.Clone(original).(*pfs.FinishCommitRequest)
	return errors.EnsureStack(t.txnEnv.getPFS().FinishCommitInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) SquashCommitSet(original *pfs.SquashCommitSetRequest) error {
	req := proto.Clone(original).(*pfs.SquashCommitSetRequest)
	return errors.EnsureStack(t.txnEnv.getPFS().SquashCommitSetInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) CreateBranch(original *pfs.CreateBranchRequest) error {
	req := proto.Clone(original).(*pfs.CreateBranchRequest)
	return errors.EnsureStack(t.txnEnv.getPFS().CreateBranchInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) DeleteBranch(original *pfs.DeleteBranchRequest) error {
	req := proto.Clone(original).(*pfs.DeleteBranchRequest)
	return errors.EnsureStack(t.txnEnv.getPFS().DeleteBranchInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) StopJob(original *pps.StopJobRequest) error {
	req := proto.Clone(original).(*pps.StopJobRequest)
	return errors.EnsureStack(t.txnEnv.getPPS().StopJobInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) UpdateJobState(original *pps.UpdateJobStateRequest) error {
	req := proto.Clone(original).(*pps.UpdateJobStateRequest)
	return errors.EnsureStack(t.txnEnv.getPPS().UpdateJobStateInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) ModifyRoleBinding(original *auth.ModifyRoleBindingRequest) (*auth.ModifyRoleBindingResponse, error) {
	req := proto.Clone(original).(*auth.ModifyRoleBindingRequest)
	res, err := t.txnEnv.getAuth().ModifyRoleBindingInTransaction(t.txnCtx, req)
	return res, errors.EnsureStack(err)
}

func (t *directTransaction) CreatePipeline(original *pps.CreatePipelineTransaction) error {
	req := proto.Clone(original).(*pps.CreatePipelineTransaction)
	return errors.EnsureStack(t.txnEnv.getPPS().CreatePipelineInTransaction(t.ctx, t.txnCtx, req))
}

func (t *directTransaction) DeleteRoleBinding(original *auth.Resource) error {
	req := proto.Clone(original).(*auth.Resource)
	return errors.EnsureStack(t.txnEnv.getAuth().DeleteRoleBindingInTransaction(t.txnCtx, req))
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

func (t *appendTransaction) DeleteRepo(req *pfs.DeleteRepoRequest) (bool, error) {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{DeleteRepo: req})
	return true, errors.EnsureStack(err)
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

func (t *appendTransaction) CreatePipeline(req *pps.CreatePipelineTransaction) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{CreatePipelineV2: req})
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
	return env.WithWriteContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		directTxn := NewDirectTransaction(ctx, env, txnCtx)
		return cb(directTxn)
	})
}

func (env *TransactionEnv) attemptTx(ctx context.Context, sqlTx *pachsql.Tx, cb func(ctx context.Context, txnCtx *txncontext.TransactionContext) error) error {
	txnCtx, err := txncontext.New(ctx, sqlTx, env.getAuth())
	if err != nil {
		return err
	}
	if env.getPFS() != nil {
		txnCtx.PfsPropagater = env.getPFS().NewPropagater(txnCtx)
	}
	if env.getPPS() != nil {
		txnCtx.PpsPropagater = env.getPPS().NewPropagater(txnCtx)
		txnCtx.PpsJobStopper = env.getPPS().NewJobStopper(txnCtx)
		txnCtx.PpsJobFinisher = env.getPPS().NewJobFinisher(txnCtx)
	}

	err = cb(ctx, txnCtx)
	if err != nil {
		return err
	}
	return txnCtx.Finish(ctx)
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
func (env *TransactionEnv) WithWriteContext(ctx context.Context, cb func(ctx context.Context, txnCtx *txncontext.TransactionContext) error) error {
	ctx = pctx.Child(ctx, "WithWriteContext")
	if err := env.waitReady(ctx); err != nil {
		return err
	}
	return dbutil.WithTx(ctx, env.db, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		return env.attemptTx(ctx, sqlTx, cb)
	})
}

// WithReadContext will call the given callback with a txncontext.TransactionContext
// which can be used to perform reads of the current cluster state. If the
// transaction is used to perform any writes, they will be silently discarded.
func (env *TransactionEnv) WithReadContext(ctx context.Context, cb func(ctx context.Context, txnCtx *txncontext.TransactionContext) error) error {
	ctx = pctx.Child(ctx, "WithReadContext")
	if err := env.waitReady(ctx); err != nil {
		return err
	}
	return col.NewDryrunSQLTx(ctx, env.db, func(ctx context.Context, sqlTx *pachsql.Tx) error {
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
		if r.CreatePipelineV2 != nil {
			if r.CreatePipelineV2.CreatePipelineRequest == nil {
				return errors.New("nil CreatePipelineRequest in CreatePipelineTransaction")
			}
			if r.CreatePipelineV2.CreatePipelineRequest.Determined != nil {
				if err := env.getPPS().CreateDetPipelineSideEffects(ctx, r.CreatePipelineV2.CreatePipelineRequest.Pipeline, r.CreatePipelineV2.CreatePipelineRequest.Determined.Workspaces); err != nil {
					return errors.Wrap(err, "apply determined pipeline side effects")
				}
			}
		}
	}
	return nil
}
