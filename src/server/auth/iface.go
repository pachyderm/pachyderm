package auth

import (
	auth_client "github.com/pachyderm/pachyderm/v2/src/auth"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
)

type APIServer interface {
	auth_client.APIServer

	AuthorizeInTransaction(*txnenv.TransactionContext, *auth_client.AuthorizeRequest) (*auth_client.AuthorizeResponse, error)

	ModifyRoleBindingInTransaction(*txnenv.TransactionContext, *auth_client.ModifyRoleBindingRequest) (*auth_client.ModifyRoleBindingResponse, error)
	GetRoleBindingInTransaction(*txnenv.TransactionContext, *auth_client.GetRoleBindingRequest) (*auth_client.GetRoleBindingResponse, error)

	// Methods to add and remove pipelines from input and output repos. These do their own auth checks
	// for specific permissions required to use a repo as a pipeline input/output.
	AddPipelineReaderToRepoInTransaction(*txnenv.TransactionContext, string, string) error
	AddPipelineWriterToRepoInTransaction(*txnenv.TransactionContext, string) error
	RemovePipelineReaderFromRepoInTransaction(*txnenv.TransactionContext, string, string) error

	// Create and Delete are internal-only APIs used by other services when creating/destroying resources.
	CreateRoleBindingInTransaction(*txnenv.TransactionContext, string, []string, *auth_client.Resource) error
	DeleteRoleBindingInTransaction(*txnenv.TransactionContext, *auth_client.Resource) error

	// GetPipelineAuthTokenInTransaction is an internal API used by PPS to generate tokens for pipelines
	GetPipelineAuthTokenInTransaction(*txnenv.TransactionContext, string) (string, error)
	RevokeAuthTokenInTransaction(*txnenv.TransactionContext, *auth_client.RevokeAuthTokenRequest) (*auth_client.RevokeAuthTokenResponse, error)

	CheckClusterIsAuthorizedInTransaction(txnCtx *txnenv.TransactionContext, p ...auth_client.Permission) error
	CheckRepoIsAuthorizedInTransaction(txnCtx *txnenv.TransactionContext, r string, p ...auth_client.Permission) error
}
