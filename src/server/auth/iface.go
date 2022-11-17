package auth

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// APIServer is the internal interface for other services to call this one.
// This includes all the public RPC methods and additional internal-only methods for use within pachd.
// These methods *do not* check that a user is authorized unless otherwise noted.
type APIServer interface {
	auth.APIServer

	CheckRepoIsAuthorized(context.Context, *pfs.Repo, ...auth.Permission) error
	CheckClusterIsAuthorized(ctx context.Context, p ...auth.Permission) error
	CheckClusterIsAuthorizedInTransaction(*txncontext.TransactionContext, ...auth.Permission) error
	CheckProjectIsAuthorizedInTransaction(*txncontext.TransactionContext, *pfs.Project, ...auth.Permission) error
	CheckRepoIsAuthorizedInTransaction(*txncontext.TransactionContext, *pfs.Repo, ...auth.Permission) error

	AuthorizeInTransaction(*txncontext.TransactionContext, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error)
	ModifyRoleBindingInTransaction(*txncontext.TransactionContext, *auth.ModifyRoleBindingRequest) (*auth.ModifyRoleBindingResponse, error)
	GetRoleBindingInTransaction(*txncontext.TransactionContext, *auth.GetRoleBindingRequest) (*auth.GetRoleBindingResponse, error)

	// Methods to add and remove pipelines from input and output repos. These do their own auth checks
	// for specific permissions required to use a repo as a pipeline input/output.
	AddPipelineReaderToRepoInTransaction(*txncontext.TransactionContext, *pfs.Repo, *pps.Pipeline) error
	AddPipelineWriterToRepoInTransaction(*txncontext.TransactionContext, *pps.Pipeline) error
	AddPipelineWriterToSourceRepoInTransaction(*txncontext.TransactionContext, *pfs.Repo, *pps.Pipeline) error
	RemovePipelineReaderFromRepoInTransaction(*txncontext.TransactionContext, *pfs.Repo, *pps.Pipeline) error

	// Create and Delete are internal-only APIs used by other services when creating/destroying resources.
	CreateRoleBindingInTransaction(*txncontext.TransactionContext, string, []string, *auth.Resource) error
	DeleteRoleBindingInTransaction(*txncontext.TransactionContext, *auth.Resource) error

	// GetPipelineAuthTokenInTransaction is an internal API used by PPS to generate tokens for pipelines
	GetPipelineAuthTokenInTransaction(*txncontext.TransactionContext, *pps.Pipeline) (string, error)
	RevokeAuthTokenInTransaction(*txncontext.TransactionContext, *auth.RevokeAuthTokenRequest) (*auth.RevokeAuthTokenResponse, error)

	GetPermissionsInTransaction(*txncontext.TransactionContext, *auth.GetPermissionsRequest) (*auth.GetPermissionsResponse, error)
}
