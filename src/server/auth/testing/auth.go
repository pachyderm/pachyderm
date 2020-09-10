package testing

import (
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/client/auth"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
)

// InactiveAPIServer (in the auth/testing package) is an implementation of the
// pachyderm auth api that returns NotActivatedError for all requests. This is
// meant to be used with local PFS and PPS servers for testing, and should
// never be used in a real Pachyderm cluster
type InactiveAPIServer struct{}

// Activate implements the Activate RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Activate(context.Context, *auth.ActivateRequest) (*auth.ActivateResponse, error) {
	return nil, auth.ErrNotActivated
}

// Deactivate implements the Deactivate RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Deactivate(context.Context, *auth.DeactivateRequest) (*auth.DeactivateResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetAdmins implements the GetAdmins RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetAdmins(context.Context, *auth.GetAdminsRequest) (*auth.GetAdminsResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetClusterRoleBindings implements the GetClusterRoleBindings RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetClusterRoleBindings(context.Context, *auth.GetClusterRoleBindingsRequest) (*auth.GetClusterRoleBindingsResponse, error) {
	return nil, auth.ErrNotActivated
}

// ModifyClusterRoleBinding implements the ModifyClusterRoleBinding RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) ModifyClusterRoleBinding(context.Context, *auth.ModifyClusterRoleBindingRequest) (*auth.ModifyClusterRoleBindingResponse, error) {
	return nil, auth.ErrNotActivated
}

// ModifyAdmins implements the ModifyAdmins RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) ModifyAdmins(context.Context, *auth.ModifyAdminsRequest) (*auth.ModifyAdminsResponse, error) {
	return nil, auth.ErrNotActivated
}

// Authenticate implements the Authenticate RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Authenticate(context.Context, *auth.AuthenticateRequest) (*auth.AuthenticateResponse, error) {
	return nil, auth.ErrNotActivated
}

// Authorize implements the Authorize RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Authorize(context.Context, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	return nil, auth.ErrNotActivated
}

// AuthorizeInTransaction is the same as the Authorize RPC but for use inside a
// running transaction.  It also returns a NotActivatedError.
func (a *InactiveAPIServer) AuthorizeInTransaction(*txnenv.TransactionContext, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	return nil, auth.ErrNotActivated
}

// WhoAmI implements the WhoAmI RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) WhoAmI(context.Context, *auth.WhoAmIRequest) (*auth.WhoAmIResponse, error) {
	return nil, auth.ErrNotActivated
}

// SetScope implements the SetScope RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) SetScope(context.Context, *auth.SetScopeRequest) (*auth.SetScopeResponse, error) {
	return nil, auth.ErrNotActivated
}

// SetScopeInTransaction is the same as the SetScope RPC but for use inside a
// running transaction.  It also returns a NotActivatedError.
func (a *InactiveAPIServer) SetScopeInTransaction(*txnenv.TransactionContext, *auth.SetScopeRequest) (*auth.SetScopeResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetScope implements the GetScope RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetScope(context.Context, *auth.GetScopeRequest) (*auth.GetScopeResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetScopeInTransaction is the same as the GetScope RPC but for use inside a
// running transaction.  It also returns a NotActivatedError.
func (a *InactiveAPIServer) GetScopeInTransaction(*txnenv.TransactionContext, *auth.GetScopeRequest) (*auth.GetScopeResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetACL implements the GetACL RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetACL(context.Context, *auth.GetACLRequest) (*auth.GetACLResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetACLInTransaction is the same as the GetACL RPC but for use inside a
// running transaction.  It also returns a NotActivatedError.
func (a *InactiveAPIServer) GetACLInTransaction(*txnenv.TransactionContext, *auth.GetACLRequest) (*auth.GetACLResponse, error) {
	return nil, auth.ErrNotActivated
}

// SetACL implements the SetACL RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) SetACL(context.Context, *auth.SetACLRequest) (*auth.SetACLResponse, error) {
	return nil, auth.ErrNotActivated
}

// SetACLInTransaction is the same as the SetACL RPC but for use inside a
// running transaction.  It also returns a NotActivatedError.
func (a *InactiveAPIServer) SetACLInTransaction(*txnenv.TransactionContext, *auth.SetACLRequest) (*auth.SetACLResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetDefaultACL implements the GetDefaultACL RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetDefaultACL(context.Context, *auth.GetDefaultACLRequest) (*auth.GetDefaultACLResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetDefaultACLInTransaction implements the GetDefaultACL RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetDefaultACLInTransaction(*txnenv.TransactionContext, *auth.GetDefaultACLRequest) (*auth.GetDefaultACLResponse, error) {
	return nil, auth.ErrNotActivated
}

// SetDefaultACL implements the SetDefaultACL RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) SetDefaultACL(context.Context, *auth.SetDefaultACLRequest) (*auth.SetDefaultACLResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetAuthToken implements the GetAuthToken RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetAuthToken(context.Context, *auth.GetAuthTokenRequest) (*auth.GetAuthTokenResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetOIDCLogin implements the GetOIDCLogin RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetOIDCLogin(context.Context, *auth.GetOIDCLoginRequest) (*auth.GetOIDCLoginResponse, error) {
	return nil, auth.ErrNotActivated
}

// ExtendAuthToken implements the ExtendAuthToken RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) ExtendAuthToken(context.Context, *auth.ExtendAuthTokenRequest) (*auth.ExtendAuthTokenResponse, error) {
	return nil, auth.ErrNotActivated
}

// RevokeAuthToken implements the RevokeAuthToken RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) RevokeAuthToken(context.Context, *auth.RevokeAuthTokenRequest) (*auth.RevokeAuthTokenResponse, error) {
	return nil, auth.ErrNotActivated
}

// SetGroupsForUser implements the SetGroupsForUser RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) SetGroupsForUser(context.Context, *auth.SetGroupsForUserRequest) (*auth.SetGroupsForUserResponse, error) {
	return nil, auth.ErrNotActivated
}

// ModifyMembers implements the ModifyMembers RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) ModifyMembers(context.Context, *auth.ModifyMembersRequest) (*auth.ModifyMembersResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetGroups implements the GetGroups RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetGroups(context.Context, *auth.GetGroupsRequest) (*auth.GetGroupsResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetUsers implements the GetUsers RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetUsers(context.Context, *auth.GetUsersRequest) (*auth.GetUsersResponse, error) {
	return nil, auth.ErrNotActivated
}

// SetConfiguration implements the SetConfiguration RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) SetConfiguration(context.Context, *auth.SetConfigurationRequest) (*auth.SetConfigurationResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetConfiguration implements the GetConfiguration RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetConfiguration(context.Context, *auth.GetConfigurationRequest) (*auth.GetConfigurationResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetOneTimePassword implements the GetOneTimePassword RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetOneTimePassword(context.Context, *auth.GetOneTimePasswordRequest) (*auth.GetOneTimePasswordResponse, error) {
	return nil, auth.ErrNotActivated
}
