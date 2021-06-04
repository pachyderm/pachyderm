package auth

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	authiface "github.com/pachyderm/pachyderm/v2/src/server/auth"
)

// an authHandler can optionally return a username string that will be cached in the request's context
type authHandler func(context.Context, authiface.APIServer, string) (string, error)

type ContextKey string

const whoAmIResultKey = ContextKey("WhoAmI")

// authDisabledOr wraps an authHandler and permits the RPC if authHandler succeeds or
// if auth is disabled on the cluster
func authDisabledOr(h authHandler) authHandler {
	return func(ctx context.Context, authApi authiface.APIServer, fullMethod string) (string, error) {
		username, err := h(ctx, authApi, fullMethod)

		if auth.IsErrNotActivated(err) {
			return "", nil
		}
		return username, err
	}
}

// unauthenticated permits any RPC even if the user has no authentication token
func unauthenticated(ctx context.Context, _ authiface.APIServer, fullMethod string) (string, error) {
	return "", nil
}

// authenticated permits an RPC if auth is fully enabled and the user is authenticated
func authenticated(ctx context.Context, authApi authiface.APIServer, fullMethod string) (string, error) {
	r, err := authApi.WhoAmI(ctx, &auth.WhoAmIRequest{})
	var username string
	if err == nil {
		username = r.Username
	}
	return username, err
}

// clusterPermissions permits an RPC if the user is authorized with the given permissions on the cluster
func clusterPermissions(permissions ...auth.Permission) authHandler {
	return func(ctx context.Context, authApi authiface.APIServer, fullMethod string) (string, error) {
		resp, err := authApi.Authorize(ctx, &auth.AuthorizeRequest{
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

func GetWhoAmI(ctx context.Context) string {
	if v := ctx.Value(whoAmIResultKey); v != nil {
		return v.(string)
	}
	return ""
}

func setWhoAmI(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, whoAmIResultKey, username)
}
