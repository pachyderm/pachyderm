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
func (a *InactiveAPIServer) Activate(ctx context.Context, req *auth.ActivateRequest) (resp *auth.ActivateResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// Deactivate implements the Deactivate RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Deactivate(ctx context.Context, req *auth.DeactivateRequest) (resp *auth.DeactivateResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// GetAdmins implements the GetAdmins RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetAdmins(ctx context.Context, req *auth.GetAdminsRequest) (resp *auth.GetAdminsResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// ModifyAdmins implements the ModifyAdmins RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) ModifyAdmins(ctx context.Context, req *auth.ModifyAdminsRequest) (resp *auth.ModifyAdminsResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// Authenticate implements the Authenticate RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Authenticate(ctx context.Context, req *auth.AuthenticateRequest) (resp *auth.AuthenticateResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// Authorize implements the Authorize RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Authorize(ctx context.Context, req *auth.AuthorizeRequest) (resp *auth.AuthorizeResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// WhoAmI implements the WhoAmI RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) WhoAmI(ctx context.Context, req *auth.WhoAmIRequest) (resp *auth.WhoAmIResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// SetScope implements the SetScope RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) SetScope(ctx context.Context, req *auth.SetScopeRequest) (resp *auth.SetScopeResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// GetScope implements the GetScope RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetScope(ctx context.Context, req *auth.GetScopeRequest) (resp *auth.GetScopeResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// GetACL implements the GetACL RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetACL(ctx context.Context, req *auth.GetACLRequest) (resp *auth.GetACLResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// SetACL implements the SetACL RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) SetACL(ctx context.Context, req *auth.SetACLRequest) (resp *auth.SetACLResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// GetAuthToken implements the GetAuthToken RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetAuthToken(ctx context.Context, req *auth.GetAuthTokenRequest) (resp *auth.GetAuthTokenResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// GetAuthToken implements the ExtendAuthToken RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) ExtendAuthToken(ctx context.Context, req *auth.ExtendAuthTokenRequest) (resp *auth.ExtendAuthTokenResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// RevokeAuthToken implements the RevokeAuthToken RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) RevokeAuthToken(ctx context.Context, req *auth.RevokeAuthTokenRequest) (resp *auth.RevokeAuthTokenResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}
