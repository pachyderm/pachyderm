package auth

import (
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
)

// an authHandler can optionally return a key-value that will be cached in the request's context
type authHandler func(*client.APIClient, string) (*keyValue, error)

type ctxKey string

type keyValue struct {
	k ctxKey
	v interface{}
}

const whoAmIKey ctxKey = "ctxKey_WhoAmI"

// authDisabledOr wraps an authHandler and permits the RPC if authHandler succeeds or
// if auth is disabled on the cluster
func authDisabledOr(h authHandler) authHandler {
	return func(pachClient *client.APIClient, fullMethod string) (*keyValue, error) {
		memoized, err := h(pachClient, fullMethod)

		// TODO: check this
		if auth.IsErrNotActivated(err) {
			return nil, nil
		}
		return memoized, err
	}
}

// unauthenticated permits any RPC even if the user has no authentication token
func unauthenticated(pachClient *client.APIClient, fullMethod string) (*keyValue, error) {
	return nil, nil
}

// authenticated permits an RPC if auth is fully enabled and the user is authenticated
func authenticated(pachClient *client.APIClient, fullMethod string) (*keyValue, error) {
	// consider the request authenticated if WhoAmI has a value
	if v := pachClient.Ctx().Value(whoAmIKey); v != nil {
		return nil, nil
	}
	r, err := pachClient.WhoAmI(pachClient.Ctx(), &auth.WhoAmIRequest{})
	return &keyValue{whoAmIKey, r}, err
}

// clusterPermissions permits an RPC if the user is authorized with the given permissions on the cluster
func clusterPermissions(permissions ...auth.Permission) authHandler {
	return func(pachClient *client.APIClient, fullMethod string) (*keyValue, error) {
		resp, err := pachClient.Authorize(pachClient.Ctx(), &auth.AuthorizeRequest{
			Resource:    &auth.Resource{Type: auth.ResourceType_CLUSTER},
			Permissions: permissions,
		})
		if err != nil {
			return nil, err
		}

		if resp.Authorized {
			return nil, nil
		}

		return nil, &auth.ErrNotAuthorized{
			Subject:  resp.Principal,
			Resource: auth.Resource{Type: auth.ResourceType_CLUSTER},
			Required: permissions,
		}
	}
}
