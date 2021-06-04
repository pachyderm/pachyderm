package pfs

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	pfs_client "github.com/pachyderm/pachyderm/v2/src/pfs"
)

// APIServer is the internal interface for other services to call this one.
// This includes all the public RPC methods and additional internal-only methods for use within pachd.
// These methods *do not* check that a user is authorized unless otherwise noted.
type APIServer interface {
	pfs_client.APIServer

	NewPropagater(*txncontext.TransactionContext) txncontext.PfsPropagater
	NewPipelineFinisher(*txncontext.TransactionContext) txncontext.PipelineCommitFinisher

	CreateRepoInTransaction(*txncontext.TransactionContext, *pfs_client.CreateRepoRequest) error
	InspectRepoInTransaction(*txncontext.TransactionContext, *pfs_client.InspectRepoRequest) (*pfs_client.RepoInfo, error)
	DeleteRepoInTransaction(*txncontext.TransactionContext, *pfs_client.DeleteRepoRequest) error

	StartCommitInTransaction(*txncontext.TransactionContext, *pfs_client.StartCommitRequest) (*pfs_client.Commit, error)
	FinishCommitInTransaction(*txncontext.TransactionContext, *pfs_client.FinishCommitRequest) error
	InspectCommitInTransaction(*txncontext.TransactionContext, *pfs_client.InspectCommitRequest) (*pfs_client.CommitInfo, error)

	InspectCommitsetInTransaction(*txncontext.TransactionContext, *pfs_client.Commitset) ([]*pfs_client.CommitInfo, error)
	SquashCommitsetInTransaction(*txncontext.TransactionContext, *pfs_client.SquashCommitsetRequest) error

	CreateBranchInTransaction(*txncontext.TransactionContext, *pfs_client.CreateBranchRequest) error
	InspectBranchInTransaction(*txncontext.TransactionContext, *pfs_client.InspectBranchRequest) (*pfs_client.BranchInfo, error)
	DeleteBranchInTransaction(*txncontext.TransactionContext, *pfs_client.DeleteBranchRequest) error

	AddFilesetInTransaction(*txncontext.TransactionContext, *pfs_client.AddFilesetRequest) error
}
