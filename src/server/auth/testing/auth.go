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

// Activate implements the Pachdyerm Auth Activate RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Activate(ctx context.Context, req *auth.ActivateRequest) (resp *auth.ActivateResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// Authenticate implements the Pachdyerm Auth Authenticate RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Authenticate(ctx context.Context, req *auth.AuthenticateRequest) (resp *auth.AuthenticateResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// Authorize implements the Pachdyerm Auth Authorize RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) Authorize(ctx context.Context, req *auth.AuthorizeRequest) (resp *auth.AuthorizeResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// WhoAmI implements the Pachdyerm Auth WhoAmI RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) WhoAmI(ctx context.Context, req *auth.WhoAmIRequest) (resp *auth.WhoAmIResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// SetScope implements the Pachdyerm Auth SetScope RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) SetScope(ctx context.Context, req *auth.SetScopeRequest) (resp *auth.SetScopeResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// GetScope implements the Pachdyerm Auth GetScope RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetScope(ctx context.Context, req *auth.GetScopeRequest) (resp *auth.GetScopeResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// GetACL implements the Pachdyerm Auth GetACL RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetACL(ctx context.Context, req *auth.GetACLRequest) (resp *auth.GetACLResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// GetCapability implements the Pachdyerm Auth GetCapability RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) GetCapability(ctx context.Context, req *auth.GetCapabilityRequest) (resp *auth.GetCapabilityResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}

// RevokeAuthToken implements the Pachdyerm Auth RevokeAuthToken RPC, but just returns NotActivatedError
func (a *InactiveAPIServer) RevokeAuthToken(ctx context.Context, req *auth.RevokeAuthTokenRequest) (resp *auth.RevokeAuthTokenResponse, retErr error) {
	return nil, auth.NotActivatedError{}
}
