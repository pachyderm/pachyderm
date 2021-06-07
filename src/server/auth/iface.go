package auth

import (
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/transactionenv/txncontext"
)

// TransactionServer is an interface for the transactionally-supported
// methods that can be called through the auth server.
type TransactionServer interface {
	AuthorizeInTransaction(*txncontext.TransactionContext, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error)

	GetScopeInTransaction(*txncontext.TransactionContext, *auth.GetScopeRequest) (*auth.GetScopeResponse, error)
	SetScopeInTransaction(*txncontext.TransactionContext, *auth.SetScopeRequest) (*auth.SetScopeResponse, error)

	GetACLInTransaction(*txncontext.TransactionContext, *auth.GetACLRequest) (*auth.GetACLResponse, error)
	SetACLInTransaction(*txncontext.TransactionContext, *auth.SetACLRequest) (*auth.SetACLResponse, error)

	GetAuthTokenInTransaction(*txncontext.TransactionContext, *auth.GetAuthTokenRequest) (*auth.GetAuthTokenResponse, error)
	RevokeAuthTokenInTransaction(*txncontext.TransactionContext, *auth.RevokeAuthTokenRequest) (*auth.RevokeAuthTokenResponse, error)

	CheckIsAuthorizedInTransaction(*txncontext.TransactionContext, *pfs.Repo, auth.Scope) error
}

// APIServer represents an auth api server
type APIServer interface {
	auth.APIServer
	TransactionServer
}
