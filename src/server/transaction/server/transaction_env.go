package server

import (
	"context"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

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

	DeleteAllInTransaction(*client.APIClient, col.STM, *pfs.DeleteAllRequest) error
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
	authClient AuthTransactionServer
	pfsClient  PfsTransactionServer
	ppsClient  PpsTransactionServer
}

// Initialize stores the references to APIServer instances in the TransactionEnv
func (env *TransactionEnv) Initialize(
	authClient AuthTransactionServer,
	pfsClient PfsTransactionServer,
	ppsClient PpsTransactionServer,
) {
	env.authClient = authClient
	env.pfsClient = pfsClient
	env.ppsClient = ppsClient
}

// AuthServer returns a reference to the interface for making transactional
// calls through the auth subsystem.
func (env *TransactionEnv) AuthServer() AuthTransactionServer {
	return env.authClient
}

// PfsServer returns a reference to the interface for making transactional
// calls through the PFS subsystem.
func (env *TransactionEnv) PfsServer() PfsTransactionServer {
	return env.pfsClient
}

// PpsServer returns a reference to the interface for making transactional
// calls through the PPS subsystem.
func (env *TransactionEnv) PpsServer() PpsTransactionServer {
	return env.ppsClient
}
