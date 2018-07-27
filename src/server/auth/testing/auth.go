package testing

import (
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/client/auth"
)

// InactiveAPIServer (in the auth/testing package) is an implementation of the
// pachyderm auth api that returns NotActivatedError for all requests. This is
// meant to be used with local PFS and PPS servers for testing, and should
// never be used in a real Pachyderm cluster
type InactiveAPIServer struct{}

// Activate implements the Activate RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Activate(_ context.Context, _ *auth.ActivateRequest) (*auth.ActivateResponse, error) {
	return nil, auth.ErrNotActivated
}

// Deactivate implements the Deactivate RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Deactivate(_ context.Context, _ *auth.DeactivateRequest) (*auth.DeactivateResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetAdmins implements the GetAdmins RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetAdmins(_ context.Context, _ *auth.GetAdminsRequest) (*auth.GetAdminsResponse, error) {
	return nil, auth.ErrNotActivated
}

// ModifyAdmins implements the ModifyAdmins RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) ModifyAdmins(_ context.Context, _ *auth.ModifyAdminsRequest) (*auth.ModifyAdminsResponse, error) {
	return nil, auth.ErrNotActivated
}

// Authenticate implements the Authenticate RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Authenticate(_ context.Context, _ *auth.AuthenticateRequest) (*auth.AuthenticateResponse, error) {
	return nil, auth.ErrNotActivated
}

// Authorize implements the Authorize RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Authorize(_ context.Context, _ *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	return nil, auth.ErrNotActivated
}

// WhoAmI implements the WhoAmI RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) WhoAmI(_ context.Context, _ *auth.WhoAmIRequest) (*auth.WhoAmIResponse, error) {
	return nil, auth.ErrNotActivated
}

// SetScope implements the SetScope RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) SetScope(_ context.Context, _ *auth.SetScopeRequest) (*auth.SetScopeResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetScope implements the GetScope RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetScope(_ context.Context, _ *auth.GetScopeRequest) (*auth.GetScopeResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetACL implements the GetACL RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetACL(_ context.Context, _ *auth.GetACLRequest) (*auth.GetACLResponse, error) {
	return nil, auth.ErrNotActivated
}

// SetACL implements the SetACL RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) SetACL(_ context.Context, _ *auth.SetACLRequest) (*auth.SetACLResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetAuthToken implements the GetAuthToken RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetAuthToken(_ context.Context, _ *auth.GetAuthTokenRequest) (*auth.GetAuthTokenResponse, error) {
	return nil, auth.ErrNotActivated
}

// ExtendAuthToken implements the ExtendAuthToken RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) ExtendAuthToken(_ context.Context, _ *auth.ExtendAuthTokenRequest) (*auth.ExtendAuthTokenResponse, error) {
	return nil, auth.ErrNotActivated
}

// RevokeAuthToken implements the RevokeAuthToken RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) RevokeAuthToken(_ context.Context, _ *auth.RevokeAuthTokenRequest) (*auth.RevokeAuthTokenResponse, error) {
	return nil, auth.ErrNotActivated
}

// SetGroupsForUser implements the SetGroupsForUser RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) SetGroupsForUser(_ context.Context, _ *auth.SetGroupsForUserRequest) (*auth.SetGroupsForUserResponse, error) {
	return nil, auth.ErrNotActivated
}

// ModifyMembers implements the ModifyMembers RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) ModifyMembers(_ context.Context, _ *auth.ModifyMembersRequest) (*auth.ModifyMembersResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetGroups implements the GetGroups RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetGroups(_ context.Context, _ *auth.GetGroupsRequest) (*auth.GetGroupsResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetUsers implements the GetUsers RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetUsers(_ context.Context, _ *auth.GetUsersRequest) (*auth.GetUsersResponse, error) {
	return nil, auth.ErrNotActivated
}

// SetConfiguration implements the SetConfiguration RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) SetConfiguration(_ context.Context, _ *auth.SetConfigurationRequest) (*auth.SetConfigurationResponse, error) {
	return nil, auth.ErrNotActivated
}

// GetConfiguration implements the GetConfiguration RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetConfiguration(_ context.Context, _ *auth.GetConfigurationRequest) (*auth.GetConfigurationResponse, error) {
	return nil, auth.ErrNotActivated
}
