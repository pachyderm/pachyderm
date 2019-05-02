package transactionenv

import (
	"context"

	etcd "github.com/coreos/etcd/clientv3"

	"github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/transaction"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

// PfsWrapper is an interface providing a wrapper for each operation that
// may be appended to a transaction through PFS.  Each call may either
// directly run the request through PFS or append it to the active transaction,
// depending on if there is an active transaction in the client context.
type PfsWrapper interface {
	CreateRepo(*pfs.CreateRepoRequest) error
	DeleteRepo(*pfs.DeleteRepoRequest) error

	StartCommit(*pfs.StartCommitRequest, *pfs.Commit) (*pfs.Commit, error)
	FinishCommit(*pfs.FinishCommitRequest) error
	DeleteCommit(*pfs.DeleteCommitRequest) error

	CreateBranch(*pfs.CreateBranchRequest) error
	DeleteBranch(*pfs.DeleteBranchRequest) error

	CopyFile(*pfs.CopyFileRequest) error
	DeleteFile(*pfs.DeleteFileRequest) error
}

type PfsTransactionDefer interface {
	PropagateBranch(branch *pfs.Branch)
	DeleteScratch(commit *pfs.Commit)
	Run() error
}

type TransactionContext struct {
	stm      col.STM
	ctx      context.Context
	txnEnv   *TransactionEnv
	pfsDefer PfsTransactionDefer
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
	AuthorizeInTransaction(*TransactionContext, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error)

	GetScopeInTransaction(*TransactionContext, *auth.GetScopeRequest) (*auth.GetScopeResponse, error)
	SetScopeInTransaction(*TransactionContext, *auth.SetScopeRequest) (*auth.SetScopeResponse, error)

	GetACLInTransaction(*TransactionContext, *auth.GetACLRequest) (*auth.GetACLResponse, error)
	SetACLInTransaction(*TransactionContext, *auth.SetACLRequest) (*auth.SetACLResponse, error)
}

// PfsTransactionServer is an interface for the transactionally-supported
// methods that can be called through the PFS server.
type PfsTransactionServer interface {
	NewTransactionDefer(col.STM) PfsTransactionDefer

	CreateRepoInTransaction(*TransactionContext, *pfs.CreateRepoRequest) error
	InspectRepoInTransaction(*TransactionContext, *pfs.InspectRepoRequest) (*pfs.RepoInfo, error)
	DeleteRepoInTransaction(*TransactionContext, *pfs.DeleteRepoRequest) error

	StartCommitInTransaction(*TransactionContext, *pfs.StartCommitRequest, *pfs.Commit) (*pfs.Commit, error)
	FinishCommitInTransaction(*TransactionContext, *pfs.FinishCommitRequest) error
	DeleteCommitInTransaction(*TransactionContext, *pfs.DeleteCommitRequest) error

	CreateBranchInTransaction(*TransactionContext, *pfs.CreateBranchRequest) error
	DeleteBranchInTransaction(*TransactionContext, *pfs.DeleteBranchRequest) error

	CopyFileInTransaction(*TransactionContext, *pfs.CopyFileRequest) error
	DeleteFileInTransaction(*TransactionContext, *pfs.DeleteFileRequest) error
}

// TransactionEnv contains the APIServer instances for each subsystem that may
// be involved in running transactions so that they can make calls to each other
// without leaving the context of a transaction.  This is a separate object
// because there are cyclic dependencies between APIServer instances.
type TransactionEnv struct {
	etcdClient *etcd.Client
	txnServer  TransactionServer
	authServer AuthTransactionServer
	pfsServer  PfsTransactionServer
}

// Initialize stores the references to APIServer instances in the TransactionEnv
func (env *TransactionEnv) Initialize(
	etcdClient *etcd.Client,
	txnServer TransactionServer,
	authServer AuthTransactionServer,
	pfsServer PfsTransactionServer,
) {
	env.etcdClient = etcdClient
	env.txnServer = txnServer
	env.authServer = authServer
	env.pfsServer = pfsServer
}

// transactionServer returns a reference to the interface for modifying
// transactions from other API servers.
func (env *TransactionEnv) transactionServer() TransactionServer {
	return env.txnServer
}

// PfsServer returns a reference to the interface for making transactional
// calls through the PFS subsystem.
func (env *TransactionEnv) PfsServer() PfsTransactionServer {
	return env.pfsServer
}

type Transaction interface {
	PfsWrapper
}

type DirectTransaction struct {
	txnCtx *TransactionContext
}

func NewDirectTransaction(ctx context.Context, stm col.STM, txnEnv *TransactionEnv) *DirectTransaction {
	return &DirectTransaction{
		txnCtx: &TransactionContext{
			ctx:      ctx,
			stm:      stm,
			txnEnv:   txnEnv,
			pfsDefer: txnEnv.PfsServer().NewTransactionDefer(stm),
		},
	}
}

func (t *DirectTransaction) Finish() error {
	return t.txnCtx.pfsDefer.Run()
}

func (t *DirectTransaction) CreateRepo(original *pfs.CreateRepoRequest) error {
	req := proto.Clone(original).(*pfs.CreateRepoRequest)
	return t.txnCtx.txnEnv.PfsServer().CreateRepoInTransaction(t.txnCtx, req)
}

func (t *DirectTransaction) DeleteRepo(original *pfs.DeleteRepoRequest) error {
	req := proto.Clone(original).(*pfs.DeleteRepoRequest)
	return t.txnCtx.txnEnv.PfsServer().DeleteRepoInTransaction(t.txnCtx, req)
}

func (t *DirectTransaction) StartCommit(original *pfs.StartCommitRequest, commit *pfs.Commit) (*pfs.Commit, error) {
	req := proto.Clone(original).(*pfs.StartCommitRequest)
	return t.txnCtx.txnEnv.PfsServer().StartCommitInTransaction(t.txnCtx, req, commit)
}

func (t *DirectTransaction) FinishCommit(original *pfs.FinishCommitRequest) error {
	req := proto.Clone(original).(*pfs.FinishCommitRequest)
	return t.txnCtx.txnEnv.PfsServer().FinishCommitInTransaction(t.txnCtx, req)
}

func (t *DirectTransaction) DeleteCommit(original *pfs.DeleteCommitRequest) error {
	req := proto.Clone(original).(*pfs.DeleteCommitRequest)
	return t.txnCtx.txnEnv.PfsServer().DeleteCommitInTransaction(t.txnCtx, req)
}

func (t *DirectTransaction) CreateBranch(original *pfs.CreateBranchRequest) error {
	req := proto.Clone(original).(*pfs.CreateBranchRequest)
	return t.txnCtx.txnEnv.PfsServer().CreateBranchInTransaction(t.txnCtx, req)
}

func (t *DirectTransaction) DeleteBranch(original *pfs.DeleteBranchRequest) error {
	req := proto.Clone(original).(*pfs.DeleteBranchRequest)
	return t.txnCtx.txnEnv.PfsServer().DeleteBranchInTransaction(t.txnCtx, req)
}

func (t *DirectTransaction) CopyFile(original *pfs.CopyFileRequest) error {
	req := proto.Clone(original).(*pfs.CopyFileRequest)
	return t.txnCtx.txnEnv.PfsServer().CopyFileInTransaction(t.txnCtx, req)
}

func (t *DirectTransaction) DeleteFile(original *pfs.DeleteFileRequest) error {
	req := proto.Clone(original).(*pfs.DeleteFileRequest)
	return t.txnCtx.txnEnv.PfsServer().DeleteFileInTransaction(t.txnCtx, req)
}

type AppendTransaction struct {
	ctx       context.Context
	activeTxn *transaction.Transaction
	txnEnv    *TransactionEnv
}

func newAppendTransaction(ctx context.Context, activeTxn *transaction.Transaction, txnEnv *TransactionEnv) *AppendTransaction {
	return &AppendTransaction{
		ctx:       ctx,
		activeTxn: activeTxn,
		txnEnv:    txnEnv,
	}
}

func (t *AppendTransaction) CreateRepo(req *pfs.CreateRepoRequest) error {
	_, err := t.txnEnv.transactionServer().AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{CreateRepo: req})
	return err
}

func (t *AppendTransaction) DeleteRepo(req *pfs.DeleteRepoRequest) error {
	_, err := t.txnEnv.transactionServer().AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{DeleteRepo: req})
	return err
}

func (t *AppendTransaction) StartCommit(req *pfs.StartCommitRequest, _ *pfs.Commit) (*pfs.Commit, error) {
	res, err := t.txnEnv.transactionServer().AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{StartCommit: req})
	if err != nil {
		return nil, err
	}
	return res.Commit, nil
}

func (t *AppendTransaction) FinishCommit(req *pfs.FinishCommitRequest) error {
	_, err := t.txnEnv.transactionServer().AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{FinishCommit: req})
	return err
}

func (t *AppendTransaction) DeleteCommit(req *pfs.DeleteCommitRequest) error {
	_, err := t.txnEnv.transactionServer().AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{DeleteCommit: req})
	return err
}

func (t *AppendTransaction) CreateBranch(req *pfs.CreateBranchRequest) error {
	_, err := t.txnEnv.transactionServer().AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{CreateBranch: req})
	return err
}

func (t *AppendTransaction) DeleteBranch(req *pfs.DeleteBranchRequest) error {
	_, err := t.txnEnv.transactionServer().AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{DeleteBranch: req})
	return err
}

func (t *AppendTransaction) CopyFile(req *pfs.CopyFileRequest) error {
	_, err := t.txnEnv.transactionServer().AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{CopyFile: req})
	return err
}

func (t *AppendTransaction) DeleteFile(req *pfs.DeleteFileRequest) error {
	_, err := t.txnEnv.transactionServer().AppendRequest(t.ctx, t.activeTxn, &transaction.TransactionRequest{DeleteFile: req})
	return err
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

	_, err = col.NewSTM(ctx, env.etcdClient, func(stm col.STM) error {
		directTxn := NewDirectTransaction(ctx, stm, env)
		err = cb(directTxn)
		if err != nil {
			return err
		}
		return directTxn.Finish()
	})
	return err
}

// ReadTransaction will call the given callback with a col.STM object which
// can be used to perform reads of the current cluster state.  If the client
// context passed in has an active transaction, it will be run in dryrun mode
// first so that reads may be performed on the latest state of the transaction.
//
// If the col.STM is used to perform any writes, they will be silently
// discarded.
/*
func (env *TransactionEnv) ReadTransaction(ctx context.Context, cb func(col.STM) error) error {
	activeTxn, err := client.GetTransaction(ctx)
	if err != nil {
		return err
	}

	_, err = col.NewDryrunSTM(ctx, env.etcdClient, func(stm col.STM) error {
		if activeTxn != nil {
			_, err := env.TransactionServer().DryrunTransaction(ctx, stm, activeTxn)
			if err != nil {
				return nil
			}
		}
		return cb(stm)
	})
	return err
}*/
