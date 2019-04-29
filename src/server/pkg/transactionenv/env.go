package transactionenv

import (
	"context"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/transaction"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

// TransactionServer is an interface used by other servers to append a request
// to an existing transaction.
type TransactionServer interface {
	AppendRequest(context.Context, *transaction.TransactionRequest) (*transaction.TransactionResponse, error)
}

// AuthTransactionServer is an interface for the transactionally-supported
// methods that can be called through the auth server.
type AuthTransactionServer interface {
	AuthorizeInTransaction(context.Context, col.STM, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error)

	GetScopeInTransaction(context.Context, col.STM, *auth.GetScopeRequest) (*auth.GetScopeResponse, error)
	SetScopeInTransaction(context.Context, col.STM, *auth.SetScopeRequest) (*auth.SetScopeResponse, error)

	GetACLInTransaction(context.Context, col.STM, *auth.GetACLRequest) (*auth.GetACLResponse, error)
	SetACLInTransaction(context.Context, col.STM, *auth.SetACLRequest) (*auth.SetACLResponse, error)
}

// PfsTransactionServer is an interface for the transactionally-supported
// methods that can be called through the PFS server.
type PfsTransactionServer interface {
	CreateRepoInTransaction(*client.APIClient, col.STM, *pfs.CreateRepoRequest) error
	InspectRepoInTransaction(*client.APIClient, col.STM, *pfs.InspectRepoRequest) (*pfs.RepoInfo, error)
	DeleteRepoInTransaction(*client.APIClient, col.STM, *pfs.DeleteRepoRequest) error

	StartCommitInTransaction(*client.APIClient, col.STM, *pfs.StartCommitRequest, *pfs.Commit) (*pfs.Commit, error)
	FinishCommitInTransaction(*client.APIClient, col.STM, *pfs.FinishCommitRequest) error
	DeleteCommitInTransaction(*client.APIClient, col.STM, *pfs.DeleteCommitRequest) error

	CreateBranchInTransaction(*client.APIClient, col.STM, *pfs.CreateBranchRequest) error
	DeleteBranchInTransaction(*client.APIClient, col.STM, *pfs.DeleteBranchRequest) error

	CopyFileInTransaction(*client.APIClient, col.STM, *pfs.CopyFileRequest) error
	DeleteFileInTransaction(*client.APIClient, col.STM, *pfs.DeleteFileRequest) error

	DeleteAllInTransaction(*client.APIClient, col.STM) error
}

// PpsTransactionServer is an interface for the transactionally-supported
// methods that can be called through the PPS server.
type PpsTransactionServer interface {
	// TODO: add these once PPS is supported in transactions
}

// TransactionEnv contains the APIServer instances for each subsystem that may
// be involved in running transactions so that they can make calls to each other
// without leaving the context of a transaction.  This is a separate object
// because there are cyclic dependencies between APIServer instances.
type TransactionEnv struct {
	txnServer  TransactionServer
	authServer AuthTransactionServer
	pfsServer  PfsTransactionServer
	ppsServer  PpsTransactionServer
}

// Initialize stores the references to APIServer instances in the TransactionEnv
func (env *TransactionEnv) Initialize(
	txnServer TransactionServer,
	authServer AuthTransactionServer,
	pfsServer PfsTransactionServer,
	ppsServer PpsTransactionServer,
) {
	env.txnServer = txnServer
	env.authServer = authServer
	env.pfsServer = pfsServer
	env.ppsServer = ppsServer
}

// TransactionServer returns a reference to the interface for modifying
// transactions from other API servers.
func (env *TransactionEnv) TransactionServer() TransactionServer {
	return env.txnServer
}

// AuthServer returns a reference to the interface for making transactional
// calls through the auth subsystem.
func (env *TransactionEnv) AuthServer() AuthTransactionServer {
	return env.authServer
}

// PfsServer returns a reference to the interface for making transactional
// calls through the PFS subsystem.
func (env *TransactionEnv) PfsServer() PfsTransactionServer {
	return env.pfsServer
}

// PpsServer returns a reference to the interface for making transactional
// calls through the PPS subsystem.
func (env *TransactionEnv) PpsServer() PpsTransactionServer {
	return env.ppsServer
}
