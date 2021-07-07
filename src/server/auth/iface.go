package auth

import (
	"context"

	auth_client "github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	pfs_client "github.com/pachyderm/pachyderm/v2/src/pfs"
)

// APIServer is the internal interface for other services to call this one.
// This includes all the public RPC methods and additional internal-only methods for use within pachd.
// These methods *do not* check that a user is authorized unless otherwise noted.
type APIServer interface {
	auth_client.APIServer

	CheckRepoIsAuthorized(context.Context, *pfs_client.Repo, ...auth_client.Permission) error
	CheckClusterIsAuthorized(ctx context.Context, p ...auth_client.Permission) error
	CheckClusterIsAuthorizedInTransaction(*txncontext.TransactionContext, ...auth_client.Permission) error
	CheckRepoIsAuthorizedInTransaction(*txncontext.TransactionContext, *pfs_client.Repo, ...auth_client.Permission) error

	AuthorizeInTransaction(*txncontext.TransactionContext, *auth_client.AuthorizeRequest) (*auth_client.AuthorizeResponse, error)
	ModifyRoleBindingInTransaction(*txncontext.TransactionContext, *auth_client.ModifyRoleBindingRequest) (*auth_client.ModifyRoleBindingResponse, error)
	GetRoleBindingInTransaction(*txncontext.TransactionContext, *auth_client.GetRoleBindingRequest) (*auth_client.GetRoleBindingResponse, error)

	// Methods to add and remove pipelines from input and output repos. These do their own auth checks
	// for specific permissions required to use a repo as a pipeline input/output.
	AddPipelineReaderToRepoInTransaction(*txncontext.TransactionContext, string, string) error
	AddPipelineWriterToRepoInTransaction(*txncontext.TransactionContext, string) error
	RemovePipelineReaderFromRepoInTransaction(*txncontext.TransactionContext, string, string) error

	// Create and Delete are internal-only APIs used by other services when creating/destroying resources.
	CreateRoleBindingInTransaction(*txncontext.TransactionContext, string, []string, *auth_client.Resource) error
	DeleteRoleBindingInTransaction(*txncontext.TransactionContext, *auth_client.Resource) error

	// GetPipelineAuthTokenInTransaction is an internal API used by PPS to generate tokens for pipelines
	GetPipelineAuthTokenInTransaction(*txncontext.TransactionContext, string) (string, error)
	RevokeAuthTokenInTransaction(*txncontext.TransactionContext, *auth_client.RevokeAuthTokenRequest) (*auth_client.RevokeAuthTokenResponse, error)
}
