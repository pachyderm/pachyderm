package auth

import (
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
)

type authHandler func(*client.APIClient, string) error

// authDisabledOr wraps an authHandler and permits the RPC if authHandler succeeds or
// if auth is disabled on the cluster
func authDisabledOr(h authHandler) authHandler {
	return func(pachClient *client.APIClient, fullMethod string) error {
		err := h(pachClient, fullMethod)
		if auth.IsErrNotActivated(err) {
			return nil
		}
		return err
	}
}

// unauthenticated permits any RPC even if the user has no authentication token
func unauthenticated(pachClient *client.APIClient, fullMethod string) error {
	return nil
}

// authenticated permits an RPC if auth is fully enabled and the user is authenticated
func authenticated(pachClient *client.APIClient, fullMethod string) error {
	_, err := pachClient.WhoAmI(pachClient.Ctx(), &auth.WhoAmIRequest{})
	return err
}

// admin permits an RPC if auth is fully enabled and the user is a cluster admin
func admin(pachClient *client.APIClient, fullMethod string) error {
	resp, err := pachClient.Authorize(pachClient.Ctx(), &auth.AuthorizeRequest{
		Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER},
		Permissions: []auth.Permission{
			auth.Permission_CLUSTER_ADMIN,
		},
	})
	if err != nil {
		return err
	}

	if resp.Authorized {
		return nil
	}

	return &auth.ErrNotAuthorized{
		Subject:  resp.Principal,
		Resource: auth.Resource{Type: auth.ResourceType_CLUSTER},
		Required: []auth.Permission{
			auth.Permission_CLUSTER_ADMIN,
		},
	}
}
