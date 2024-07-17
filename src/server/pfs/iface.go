package pfs

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	pfs_client "github.com/pachyderm/pachyderm/v2/src/pfs"
)

// APIServer is the internal interface for other services to call this one.
// This includes all the public RPC methods and additional internal-only methods for use within pachd.
// These methods *do not* check that a user is authorized unless otherwise noted.
type APIServer interface {
	pfs_client.APIServer

	NewPropagater(*txncontext.TransactionContext) txncontext.PfsPropagater
	NewRepoValidator(*txncontext.TransactionContext) txncontext.PfsRepoValidator

	CreateRepoInTransaction(context.Context, *txncontext.TransactionContext, *pfs_client.CreateRepoRequest) error
	InspectRepoInTransaction(context.Context, *txncontext.TransactionContext, *pfs_client.InspectRepoRequest) (*pfs_client.RepoInfo, error)
	DeleteRepoInTransaction(context.Context, *txncontext.TransactionContext, *pfs_client.DeleteRepoRequest) (bool, error)
	DeleteReposInTransaction(context.Context, *txncontext.TransactionContext, []*pfs_client.Repo, bool) error

	StartCommitInTransaction(context.Context, *txncontext.TransactionContext, *pfs_client.StartCommitRequest) (*pfs_client.Commit, error)
	FinishCommitInTransaction(context.Context, *txncontext.TransactionContext, *pfs_client.FinishCommitRequest) error
	InspectCommitInTransaction(context.Context, *txncontext.TransactionContext, *pfs_client.InspectCommitRequest) (*pfs_client.CommitInfo, error)

	InspectCommitSetInTransaction(context.Context, *txncontext.TransactionContext, *pfs_client.CommitSet, bool) ([]*pfs_client.CommitInfo, error)
	SquashCommitSetInTransaction(context.Context, *txncontext.TransactionContext, *pfs_client.SquashCommitSetRequest) error

	CreateBranchInTransaction(context.Context, *txncontext.TransactionContext, *pfs_client.CreateBranchRequest) error
	InspectBranchInTransaction(context.Context, *txncontext.TransactionContext, *pfs_client.InspectBranchRequest) (*pfs_client.BranchInfo, error)
	DeleteBranchInTransaction(context.Context, *txncontext.TransactionContext, *pfs_client.DeleteBranchRequest) error

	AddFileSetInTransaction(context.Context, *txncontext.TransactionContext, *pfs_client.AddFileSetRequest) error
	ActivateAuthInTransaction(context.Context, *txncontext.TransactionContext, *pfs_client.ActivateAuthRequest) (*pfs_client.ActivateAuthResponse, error)
}
