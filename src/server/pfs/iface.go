package pfs

import (
	"context"

	pfs_client "github.com/pachyderm/pachyderm/src/client/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/transactionenv/txncontext"
)

// TransactionServer is an interface for the transactionally-supported
// methods that can be called through the PFS server.
type TransactionServer interface {
	NewPropagater(col.STM) txncontext.PfsPropagater
	NewPipelineFinisher(*txncontext.TransactionContext) txncontext.PipelineCommitFinisher

	CreateRepoInTransaction(*txncontext.TransactionContext, *pfs_client.CreateRepoRequest) error
	InspectRepoInTransaction(*txncontext.TransactionContext, *pfs_client.InspectRepoRequest) (*pfs_client.RepoInfo, error)
	DeleteRepoInTransaction(*txncontext.TransactionContext, *pfs_client.DeleteRepoRequest) error

	StartCommitInTransaction(*txncontext.TransactionContext, *pfs_client.StartCommitRequest, *pfs_client.Commit) (*pfs_client.Commit, error)
	FinishCommitInTransaction(*txncontext.TransactionContext, *pfs_client.FinishCommitRequest) error
	DeleteCommitInTransaction(*txncontext.TransactionContext, *pfs_client.DeleteCommitRequest) error

	CreateBranchInTransaction(*txncontext.TransactionContext, *pfs_client.CreateBranchRequest) error
	InspectBranchInTransaction(*txncontext.TransactionContext, *pfs_client.InspectBranchRequest) (*pfs_client.BranchInfo, error)
	DeleteBranchInTransaction(*txncontext.TransactionContext, *pfs_client.DeleteBranchRequest) error
}

// APIServer is the combination of the public RPC interface, the internal transaction interface and
// other internal methods exposed by the PFS service
type APIServer interface {
	pfs_client.APIServer
	TransactionServer

	ListRepoNoAuth(context.Context, *pfs_client.ListRepoRequest) (*pfs_client.ListRepoResponse, error)
}
