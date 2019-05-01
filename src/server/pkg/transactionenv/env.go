package transactionenv

import (
	"context"

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
	InspectRepo(*pfs.InspectRepoRequest) (*pfs.RepoInfo, error)
	DeleteRepo(*pfs.DeleteRepoRequest) error

	StartCommit(*pfs.StartCommitRequest, *pfs.Commit) (*pfs.Commit, error)
	FinishCommit(*pfs.FinishCommitRequest) error
	DeleteCommit(*pfs.DeleteCommitRequest) error

	CreateBranch(*pfs.CreateBranchRequest) error
	DeleteBranch(*pfs.DeleteBranchRequest) error

	CopyFile(*pfs.CopyFileRequest) error
	DeleteFile(*pfs.DeleteFileRequest) error

	DeleteAll() error
}

// TransactionServer is an interface used by other servers to append a request
// to an existing transaction.
type TransactionServer interface {
	AppendRequest(context.Context, *transaction.TransactionRequest) (*transaction.TransactionResponse, error)
}

// AuthTransactionServer is an interface for the transactionally-supported
// methods that can be called through the auth server.
type AuthTransactionServer interface {
	NewTransactionDefer(context.Context, col.STM) (TxnDefer, error)

	AuthorizeInTransaction(context.Context, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error)

	GetScopeInTransaction(context.Context, *auth.GetScopeRequest) (*auth.GetScopeResponse, error)
	SetScopeInTransaction(context.Context, *auth.SetScopeRequest) (*auth.SetScopeResponse, error)

	GetACLInTransaction(context.Context, *auth.GetACLRequest) (*auth.GetACLResponse, error)
	SetACLInTransaction(context.Context, *auth.SetACLRequest) (*auth.SetACLResponse, error)
}

// PfsTransactionServer is an interface for the transactionally-supported
// methods that can be called through the PFS server.
type PfsTransactionServer interface {
	NewTransaction(context.Context, col.STM) (PfsTransaction, error)

	CreateRepoInTransaction(context.Context, *pfs.CreateRepoRequest) error
	InspectRepoInTransaction(context.Context, *pfs.InspectRepoRequest) (*pfs.RepoInfo, error)
	DeleteRepoInTransaction(context.Context, *pfs.DeleteRepoRequest) error

	StartCommitInTransaction(context.Context, *pfs.StartCommitRequest, *pfs.Commit) (*pfs.Commit, error)
	FinishCommitInTransaction(context.Context, *pfs.FinishCommitRequest) error
	DeleteCommitInTransaction(context.Context, *pfs.DeleteCommitRequest) error

	CreateBranchInTransaction(context.Context, *pfs.CreateBranchRequest) error
	DeleteBranchInTransaction(context.Context, *pfs.DeleteBranchRequest) error

	CopyFileInTransaction(context.Context, *pfs.CopyFileRequest) error
	DeleteFileInTransaction(context.Context, *pfs.DeleteFileRequest) error

	DeleteAllInTransaction(context.Context) error
}

// TransactionEnv contains the APIServer instances for each subsystem that may
// be involved in running transactions so that they can make calls to each other
// without leaving the context of a transaction.  This is a separate object
// because there are cyclic dependencies between APIServer instances.
type TransactionEnv struct {
	txnServer  TransactionServer
	authServer AuthTransactionServer
	pfsServer  PfsTransactionServer
}

// Initialize stores the references to APIServer instances in the TransactionEnv
func (env *TransactionEnv) Initialize(
	txnServer TransactionServer,
	authServer AuthTransactionServer,
	pfsServer PfsTransactionServer,
) {
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

type TransactionWrapper interface {
	txnContext TransactionContext
}

type TransactionContext interface {
	ctx context.Context
	finish() error
}

type DirectTransaction struct {
	ctx    context.Context
	stm    col.STM
	txnEnv TransactionEnv
	pfsTxn PfsTransaction
}

func newDirectTransaction(ctx context.Context, stm col.STM, txnEnv *txnEnv) Transaction {
	return &DirectTransaction{ctx: ctx, stm: stm, txnEnv: txnEnv}
}

func (t *DirectTransaction) Pfs() PfsWrapper {
	if t.pfsTxn == nil {
		t.pfsTxn = t.txnEnv.PfsServer().NewTransaction()
	}
	return t
}

func (t *DirectTransaction) finish() error {
	if t.pfs != nil {
		return t.pfs.finish()
	}
}

func (t *DirectTransaction) CreateRepo(req *pfs.CreateRepoRequest) error {
	res, err := t.txnEnv.PfsServer().CreateRepoInTransaction(t.ctx, t.stm, req)
	return err
}

func (t *DirectTransaction) DeleteRepo(req *pfs.DeleteRepoRequest) error {
	res, err := t.txnEnv.PfsServer().DeleteRepoInTransaction(t.ctx, t.stm, req)
	return err
}

func (t *DirectTransaction) StartCommit(req *pfs.StartCommitRequest) (*pfs.Commit, error) {
	res, err := t.txnEnv.PfsServer().StartCommitInTransaction(t.ctx, t.stm, req)
	if err != nil {
		return nil, err
	}
	return res.Commit, nil
}

func (t *DirectTransaction) FinishCommit(req *pfs.FinishCommitRequest) error {
	res, err := t.txnEnv.PfsServer().FinishCommitInTransaction(t.ctx, t.stm, req)
	return err
}

func (t *DirectTransaction) DeleteCommit(req *pfs.DeleteCommitRequest) error {
	res, err := t.txnEnv.PfsServer().DeleteCommitInTransaction(t.ctx, t.stm, req)
	return err
}

func (t *DirectTransaction) CreateBranch(req *pfs.CreateBranchRequest) error {
	res, err := t.txnEnv.PfsServer().CreateBranchInTransaction(t.ctx, t.stm, req)
	return err
}

func (t *DirectTransaction) DeleteBranch(req *pfs.DeleteBranchRequest) error {
	res, err := t.txnEnv.PfsServer().DeleteBranchInTransaction(t.ctx, t.stm, req)
	return err
}

func (t *DirectTransaction) CopyFile(req *pfs.CopyFileRequest) error {
	res, err := t.txnEnv.PfsServer().CopyFileInTransaction(t.ctx, t.stm, req)
	return err
}

func (t *DirectTransaction) DeleteFile(req *pfs.DeleteFileRequest) error {
	res, err := t.txnEnv.PfsServer().DeleteFileInTransaction(t.ctx, t.stm, req)
	return err
}

type AppendTransaction struct {
	ctx       context.Context
	activeTxn *transaction.Transaction
	txnEnv    TransactionEnv
}

func newAppendTransaction(ctx context.Context, activeTxn *transaction.Transaction, txnEnv *txnEnv) Transaction {
	return &AppendTransaction{
		ctx:       ctx,
		activeTxn: activeTxn,
		txnEnv:    txnEnv,
		pfs:       newPfsAppendTransaction(ctx, txnEnv.transactionServer()),
	}
}

func (t *AppendTransaction) Pfs() PfsTransaction {
	return t
}

func (t *AppendTransaction) finish() error {
	return t.pfs.finish()
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
		return nil, err
	}

	run := func(txn Transaction) error {
		err = cb(txn)
		if err != nil {
			return err
		}
		return txn.finish()
	}

	if activeTxn != nil {
		appendTxn := newAppendTransaction(ctx, activeTxn, env)
		return run(appendTxn)
	}

	_, err = col.NewSTM(ctx, env.etcdClient, func(stm col.STM) error {
		directTxn := newDirectTransaction(ctx, stm, env)
		return run(directTxn)
	})
	return err
}

type pfsAppendTransaction struct {
	ctx    context.Context
	txnEnv TransactionEnv
}

func newPfsAppendTransaction(ctx context.Context, txnEnv TransactionEnv) PfsTransaction {
	return &pfsAppendTransaction{ctx: ctx, txnEnv: txnEnv}
}

func (t *pfsAppendTransaction) CreateRepo(req *pfs.CreateRepoRequest) error {
	res, err := t.txnEnv.TransactionServer().AppendRequest(t.ctx, &transaction.TransactionRequest{CreateRepo: req})
	return err
}

func (t *pfsAppendTransaction) DeleteRepo(req *pfs.DeleteRepoRequest) error {
	res, err := t.txnEnv.TransactionServer().AppendRequest(t.ctx, &transaction.TransactionRequest{DeleteRepo: req})
	return err
}

func (t *pfsAppendTransaction) StartCommit(req *pfs.StartCommitRequest) (*pfs.Commit, error) {
	res, err := t.txnEnv.TransactionServer().AppendRequest(t.ctx, &transaction.TransactionRequest{StartCommit: req})
	if err != nil {
		return nil, err
	}
	return res.Commit, nil
}

func (t *pfsAppendTransaction) FinishCommit(req *pfs.FinishCommitRequest) error {
	res, err := t.txnEnv.TransactionServer().AppendRequest(t.ctx, &transaction.TransactionRequest{FinishCommit: req})
	return err
}

func (t *pfsAppendTransaction) DeleteCommit(req *pfs.DeleteCommitRequest) error {
	res, err := t.txnEnv.TransactionServer().AppendRequest(t.ctx, &transaction.TransactionRequest{DeleteCommit: req})
	return err
}

func (t *pfsAppendTransaction) CreateBranch(req *pfs.CreateBranchRequest) error {
	res, err := t.txnEnv.TransactionServer().AppendRequest(t.ctx, &transaction.TransactionRequest{CreateBranch: req})
	return err
}

func (t *pfsAppendTransaction) DeleteBranch(req *pfs.DeleteBranchRequest) error {
	res, err := t.txnEnv.TransactionServer().AppendRequest(t.ctx, &transaction.TransactionRequest{DeleteBranch: req})
	return err
}

func (t *pfsAppendTransaction) CopyFile(req *pfs.CopyFileRequest) error {
	res, err := t.txnEnv.TransactionServer().AppendRequest(t.ctx, &transaction.TransactionRequest{CopyFile: req})
	return err
}

func (t *pfsAppendTransaction) DeleteFile(req *pfs.DeleteFileRequest) error {
	res, err := t.txnEnv.TransactionServer().AppendRequest(t.ctx, &transaction.TransactionRequest{DeleteFile: req})
	return err
}
