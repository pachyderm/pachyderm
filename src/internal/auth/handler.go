package auth

import (
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
)

// an authHandler can optionally return a username string that will be cached in the request's context
type authHandler func(*client.APIClient, string) (string, error)

type ContextKey string

const WhoAmIResultKey = ContextKey("WhoAmI")

// authDisabledOr wraps an authHandler and permits the RPC if authHandler succeeds or
// if auth is disabled on the cluster
func authDisabledOr(h authHandler) authHandler {
	return func(pachClient *client.APIClient, fullMethod string) (string, error) {
		username, err := h(pachClient, fullMethod)

		// TODO: check this
		if auth.IsErrNotActivated(err) {
			return "", nil
		}
		return username, err
	}
}

// unauthenticated permits any RPC even if the user has no authentication token
func unauthenticated(pachClient *client.APIClient, fullMethod string) (string, error) {
	return "", nil
}

// authenticated permits an RPC if auth is fully enabled and the user is authenticated
func authenticated(pachClient *client.APIClient, fullMethod string) (string, error) {
	// consider the request authenticated if WhoAmI has a value
	if v := pachClient.Ctx().Value(WhoAmIResultKey); v != nil {
		return "", nil
	}
	r, err := pachClient.WhoAmI(pachClient.Ctx(), &auth.WhoAmIRequest{})
	return r.Username, err
}

// clusterPermissions permits an RPC if the user is authorized with the given permissions on the cluster
func clusterPermissions(permissions ...auth.Permission) authHandler {
	return func(pachClient *client.APIClient, fullMethod string) (string, error) {
		resp, err := pachClient.Authorize(pachClient.Ctx(), &auth.AuthorizeRequest{
			Resource:    &auth.Resource{Type: auth.ResourceType_CLUSTER},
			Permissions: permissions,
		})
		if err != nil {
			return "", err
		}

		if resp.Authorized {
			return "", nil
		}

		return "", &auth.ErrNotAuthorized{
			Subject:  resp.Principal,
			Resource: auth.Resource{Type: auth.ResourceType_CLUSTER},
			Required: permissions,
		}
	}
}
