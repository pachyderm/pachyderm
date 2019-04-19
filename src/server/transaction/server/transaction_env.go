package server

import (
	"context"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

type pfsTransactionServer interface {
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

type authTransactionServer interface {
	AuthorizeInTransaction(context.Context, col.STM, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error)

	GetScopeInTransaction(context.Context, col.STM, *auth.GetScopeRequest) (*auth.GetScopeResponse, error)
	SetScopeInTransaction(context.Context, col.STM, *auth.SetScopeRequest) (*auth.SetScopeResponse, error)

	GetACLInTransaction(context.Context, col.STM, *auth.GetACLRequest) (*auth.GetACLResponse, error)
	SetACLInTransaction(context.Context, col.STM, *auth.SetACLRequest) (*auth.SetACLResponse, error)
}

type ppsTransactionServer interface {
	// TODO: add these once PPS is supported in transactions
}

// TransactionEnv contains the APIServer instances for each subsystem that may
// be involved in running transactions so that they can make calls to each other
// without leaving the context of a transaction.  This is a separate object
// because there are cyclic dependencies between APIServer instances.
type TransactionEnv struct {
	authClient authTransactionServer
	pfsClient  pfsTransactionServer
	ppsClient  ppsTransactionServer
}

// Initialize stores the references to APIServer instances in the TransactionEnv
func (env *TransactionEnv) Initialize(
	authClient authTransactionServer,
	pfsClient pfsTransactionServer,
	ppsClient ppsTransactionServer,
) {
	env.authClient = authClient
	env.pfsClient = pfsClient
	env.ppsClient = ppsClient
}

func (env *TransactionEnv) AuthServer() authTransactionServer {
	return env.authClient
}

func (env *TransactionEnv) PfsServer() pfsTransactionServer {
	return env.pfsClient
}

func (env *TransactionEnv) PpsServer() ppsTransactionServer {
	return env.ppsClient
}
