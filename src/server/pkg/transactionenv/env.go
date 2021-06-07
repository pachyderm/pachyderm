package transactionenv

import (
	"context"

	"github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/transaction"
	pfs_iface "github.com/pachyderm/pachyderm/src/server/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/transactionenv/txncontext"
	pps_iface "github.com/pachyderm/pachyderm/src/server/pps"
)

// PfsWrites is an interface providing a wrapper for each operation that
// may be appended to a transaction through PFS.  Each call may either
// directly run the request through PFS or append it to the active transaction,
// depending on if there is an active transaction in the client context.
type PfsWrites interface {
	CreateRepo(*pfs.CreateRepoRequest) error
	DeleteRepo(*pfs.DeleteRepoRequest) error

	StartCommit(*pfs.StartCommitRequest, *pfs.Commit) (*pfs.Commit, error)
	FinishCommit(*pfs.FinishCommitRequest) error
	DeleteCommit(*pfs.DeleteCommitRequest) error

	CreateBranch(*pfs.CreateBranchRequest) error
	DeleteBranch(*pfs.DeleteBranchRequest) error
}

// PpsWrites is an interface providing a wrapper for each operation that
// may be appended to a transaction through PPS.  Each call may either
// directly run the request through PPS or append it to the active transaction,
// depending on if there is an active transaction in the client context.
type PpsWrites interface {
	UpdateJobState(*pps.UpdateJobStateRequest) error
	CreatePipeline(*pps.CreatePipelineRequest, **pfs.Commit) error
}

// AuthWrites is an interface providing a wrapper for each operation that
// may be appended to a transaction through the Auth server.  Each call may
// either directly run the request through Auth or append it to the active
// transaction, depending on if there is an active transaction in the client
// context.
type AuthWrites interface {
	SetScope(*auth.SetScopeRequest) (*auth.SetScopeResponse, error)
	SetACL(*auth.SetACLRequest) (*auth.SetACLResponse, error)
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

// AuthTransactionServer is an interface for the transactionally-supported
// methods that can be called through the auth server.
type AuthTransactionServer interface {
	AuthorizeInTransaction(*txncontext.TransactionContext, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error)

	GetScopeInTransaction(*txncontext.TransactionContext, *auth.GetScopeRequest) (*auth.GetScopeResponse, error)
	SetScopeInTransaction(*txncontext.TransactionContext, *auth.SetScopeRequest) (*auth.SetScopeResponse, error)

	GetACLInTransaction(*txncontext.TransactionContext, *auth.GetACLRequest) (*auth.GetACLResponse, error)
	SetACLInTransaction(*txncontext.TransactionContext, *auth.SetACLRequest) (*auth.SetACLResponse, error)

	GetAuthTokenInTransaction(*txncontext.TransactionContext, *auth.GetAuthTokenRequest) (*auth.GetAuthTokenResponse, error)
	RevokeAuthTokenInTransaction(*txncontext.TransactionContext, *auth.RevokeAuthTokenRequest) (*auth.RevokeAuthTokenResponse, error)
}

// TransactionEnv contains the APIServer instances for each subsystem that may
// be involved in running transactions so that they can make calls to each other
// without leaving the context of a transaction.  This is a separate object
// because there are cyclic dependencies between APIServer instances.
type TransactionEnv struct {
	serviceEnv *serviceenv.ServiceEnv
	txnServer  TransactionServer
	authServer AuthTransactionServer
	pfsServer  pfs_iface.TransactionServer
	ppsServer  pps_iface.TransactionServer
}

// Initialize stores the references to APIServer instances in the TransactionEnv
func (env *TransactionEnv) Initialize(
	serviceEnv *serviceenv.ServiceEnv,
	txnServer TransactionServer,
	authServer AuthTransactionServer,
	pfsServer pfs_iface.TransactionServer,
	ppsServer pps_iface.TransactionServer,
) {
	env.serviceEnv = serviceEnv
	env.txnServer = txnServer
	env.authServer = authServer
	env.pfsServer = pfsServer
	env.ppsServer = ppsServer
}

// Transaction is an interface to unify the code that may either perform an
// action directly or append an action to an existing transaction (depending on
// if there is an active transaction in the client context metadata).  There
// are two implementations of this interface:
//  directTransaction: all operations will be run directly through the relevant
//    server, all inside the same STM.
//  appendTransaction: all operations will be appended to the active transaction
//    which will then be dryrun so that the response for the operation can be
//    returned.  Each operation that is appended will do a new dryrun, so this
//    isn't as efficient as it could be.
type Transaction interface {
	PfsWrites
	PpsWrites
	AuthWrites
}

type directTransaction struct {
	txnCtx *txncontext.TransactionContext
	txnEnv *TransactionEnv
}

// NewDirectTransaction is a helper function to instantiate a directTransaction
// object.  It is exposed so that the transaction API server can run a direct
// transaction even though there is an active transaction in the context (which
// is why it cannot use `WithTransaction`).
func NewDirectTransaction(txnEnv *TransactionEnv, txnCtx *txncontext.TransactionContext) Transaction {
	return &directTransaction{
		txnCtx: txnCtx,
		txnEnv: txnEnv,
	}
}

func (t *directTransaction) CreateRepo(original *pfs.CreateRepoRequest) error {
	req := proto.Clone(original).(*pfs.CreateRepoRequest)
	return t.txnEnv.pfsServer.CreateRepoInTransaction(t.txnCtx, req)
}

func (t *directTransaction) DeleteRepo(original *pfs.DeleteRepoRequest) error {
	req := proto.Clone(original).(*pfs.DeleteRepoRequest)
	return t.txnEnv.pfsServer.DeleteRepoInTransaction(t.txnCtx, req)
}

func (t *directTransaction) StartCommit(original *pfs.StartCommitRequest, commit *pfs.Commit) (*pfs.Commit, error) {
	req := proto.Clone(original).(*pfs.StartCommitRequest)
	return t.txnEnv.pfsServer.StartCommitInTransaction(t.txnCtx, req, commit)
}

func (t *directTransaction) FinishCommit(original *pfs.FinishCommitRequest) error {
	req := proto.Clone(original).(*pfs.FinishCommitRequest)
	return t.txnEnv.pfsServer.FinishCommitInTransaction(t.txnCtx, req)
}

func (t *directTransaction) DeleteCommit(original *pfs.DeleteCommitRequest) error {
	req := proto.Clone(original).(*pfs.DeleteCommitRequest)
	return t.txnEnv.pfsServer.DeleteCommitInTransaction(t.txnCtx, req)
}

func (t *directTransaction) CreateBranch(original *pfs.CreateBranchRequest) error {
	req := proto.Clone(original).(*pfs.CreateBranchRequest)
	return t.txnEnv.pfsServer.CreateBranchInTransaction(t.txnCtx, req)
}

func (t *directTransaction) DeleteBranch(original *pfs.DeleteBranchRequest) error {
	req := proto.Clone(original).(*pfs.DeleteBranchRequest)
	return t.txnEnv.pfsServer.DeleteBranchInTransaction(t.txnCtx, req)
}

func (t *directTransaction) UpdateJobState(original *pps.UpdateJobStateRequest) error {
	req := proto.Clone(original).(*pps.UpdateJobStateRequest)
	return t.txnEnv.ppsServer.UpdateJobStateInTransaction(t.txnCtx, req)
}

func (t *directTransaction) SetScope(original *auth.SetScopeRequest) (*auth.SetScopeResponse, error) {
	req := proto.Clone(original).(*auth.SetScopeRequest)
	return t.txnEnv.authServer.SetScopeInTransaction(t.txnCtx, req)
}

func (t *directTransaction) SetACL(original *auth.SetACLRequest) (*auth.SetACLResponse, error) {
	req := proto.Clone(original).(*auth.SetACLRequest)
	return t.txnEnv.authServer.SetACLInTransaction(t.txnCtx, req)
}

func (t *directTransaction) CreatePipeline(original *pps.CreatePipelineRequest, specCommit **pfs.Commit) error {
	req := proto.Clone(original).(*pps.CreatePipelineRequest)
	return t.txnEnv.ppsServer.CreatePipelineInTransaction(t.txnCtx, req, specCommit)
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
	return err
}

func (t *appendTransaction) DeleteRepo(req *pfs.DeleteRepoRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{DeleteRepo: req})
	return err
}

func (t *appendTransaction) StartCommit(req *pfs.StartCommitRequest, _ *pfs.Commit) (*pfs.Commit, error) {
	res, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{StartCommit: req})
	if err != nil {
		return nil, err
	}
	return res.Commit, nil
}

func (t *appendTransaction) FinishCommit(req *pfs.FinishCommitRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{FinishCommit: req})
	return err
}

func (t *appendTransaction) DeleteCommit(req *pfs.DeleteCommitRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{DeleteCommit: req})
	return err
}

func (t *appendTransaction) CreateBranch(req *pfs.CreateBranchRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{CreateBranch: req})
	return err
}

func (t *appendTransaction) DeleteBranch(req *pfs.DeleteBranchRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{DeleteBranch: req})
	return err
}

func (t *appendTransaction) UpdateJobState(req *pps.UpdateJobStateRequest) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{UpdateJobState: req})
	return err
}

func (t *appendTransaction) CreatePipeline(req *pps.CreatePipelineRequest, _ **pfs.Commit) error {
	_, err := t.txnEnv.txnServer.AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{CreatePipeline: req})
	return err
}

func (t *appendTransaction) SetScope(original *auth.SetScopeRequest) (*auth.SetScopeResponse, error) {
	panic("SetScope not yet implemented in transactions")
}

func (t *appendTransaction) SetACL(original *auth.SetACLRequest) (*auth.SetACLResponse, error) {
	panic("SetACL not yet implemented in transactions")
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
		directTxn := NewDirectTransaction(env, txnCtx)
		return cb(directTxn)
	})
}

// WithWriteContext will call the given callback with a txncontext.TransactionContext
// which can be used to perform reads and writes on the current cluster state.
func (env *TransactionEnv) WithWriteContext(ctx context.Context, cb func(*txncontext.TransactionContext) error) error {
	_, err := col.NewSTM(ctx, env.serviceEnv.GetEtcdClient(), func(stm col.STM) error {
		pachClient := env.serviceEnv.GetPachClient(ctx)
		txnCtx := &txncontext.TransactionContext{
			Client:        pachClient,
			ClientContext: pachClient.Ctx(),
			Stm:           stm,
			PfsPropagater: env.pfsServer.NewPropagater(stm),
		}
		txnCtx.CommitFinisher = env.pfsServer.NewPipelineFinisher(txnCtx)

		err := cb(txnCtx)
		if err != nil {
			return err
		}
		return txnCtx.Finish()
	})
	return err
}

// WithReadContext will call the given callback with a txncontext.TransactionContext
// which can be used to perform reads of the current cluster state. If the
// transaction is used to perform any writes, they will be silently discarded.
func (env *TransactionEnv) WithReadContext(ctx context.Context, cb func(*txncontext.TransactionContext) error) error {
	return col.NewDryrunSTM(ctx, env.serviceEnv.GetEtcdClient(), func(stm col.STM) error {
		pachClient := env.serviceEnv.GetPachClient(ctx)
		txnCtx := &txncontext.TransactionContext{
			Client:         pachClient,
			ClientContext:  pachClient.Ctx(),
			Stm:            stm,
			PfsPropagater:  env.pfsServer.NewPropagater(stm),
			CommitFinisher: nil, // don't alter any pipeline commits in a read-only setting
		}

		err := cb(txnCtx)
		if err != nil {
			return err
		}
		return txnCtx.Finish()
	})
}
