package pfs

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	pfs_client "github.com/pachyderm/pachyderm/v2/src/pfs"
)

// APIServer is the internal interface for other services to call this one.
// This includes all the public RPC methods and additional internal-only methods for use within pachd.
// These methods *do not* check that a user is authorized unless otherwise noted.
type APIServer interface {
	pfs_client.APIServer

	NewPropagater(*txncontext.TransactionContext) txncontext.PfsPropagater

	CreateRepoInTransaction(*txncontext.TransactionContext, *pfs_client.CreateRepoRequest) error
	InspectRepoInTransaction(*txncontext.TransactionContext, *pfs_client.InspectRepoRequest) (*pfs_client.RepoInfo, error)
	DeleteRepoInTransaction(*txncontext.TransactionContext, *pfs_client.DeleteRepoRequest) error

	StartCommitInTransaction(*txncontext.TransactionContext, *pfs_client.StartCommitRequest) (*pfs_client.Commit, error)
	FinishCommitInTransaction(*txncontext.TransactionContext, *pfs_client.FinishCommitRequest) error
	InspectCommitInTransaction(*txncontext.TransactionContext, *pfs_client.InspectCommitRequest) (*pfs_client.CommitInfo, error)

	InspectCommitSetInTransaction(*txncontext.TransactionContext, *pfs_client.CommitSet) ([]*pfs_client.CommitInfo, error)
	SquashCommitSetInTransaction(*txncontext.TransactionContext, *pfs_client.SquashCommitSetRequest) error

	CreateBranchInTransaction(*txncontext.TransactionContext, *pfs_client.CreateBranchRequest) error
	InspectBranchInTransaction(*txncontext.TransactionContext, *pfs_client.InspectBranchRequest) (*pfs_client.BranchInfo, error)
	DeleteBranchInTransaction(*txncontext.TransactionContext, *pfs_client.DeleteBranchRequest) error

	AddFileSetInTransaction(*txncontext.TransactionContext, *pfs_client.AddFileSetRequest) error

	ListRepoCallback(context.Context, *pfs_client.ListRepoRequest, func(*pfs_client.RepoInfo) error) error
	CreateFileSetCallback(ctx context.Context, cb func(*fileset.UnorderedWriter) error) (*fileset.ID, error)
	GetFileWithWriter(ctx context.Context, request *pfs_client.GetFileRequest, writer io.Writer) (int64, error)
}
