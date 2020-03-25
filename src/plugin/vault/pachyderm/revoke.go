package pachyderm

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	pclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

// Revoke revokes the caller's credentials (by sending a request to Pachyderm).
// Unlike other handlers, it doesn't get assigned to a path; instead it's
// called by the vault lease API when a token's lease expires or is revoked.
// It's set in Backend.Secrets[0].Revoke in backend.go
func (b *backend) Revoke(ctx context.Context, req *logical.Request, data *framework.FieldData) (resp *logical.Response, retErr error) {
	b.Logger().Debug(fmt.Sprintf("(%s) %s received at %s", req.ID, req.Operation, req.Path))
	defer func() {
		b.Logger().Debug(fmt.Sprintf("(%s) %s finished at %s with result (success=%t)", req.ID, req.Operation, req.Path, retErr == nil && !resp.IsError()))
	}()

	// Extract pachyderm token from vault secret
	tokenIface, ok := req.Secret.InternalData["user_token"]
	if !ok {
		return nil, errors.Errorf("secret is missing user_token")
	}
	userToken, ok := tokenIface.(string)
	if !ok {
		return nil, errors.Errorf("secret.user_token has wrong type (expected string but was %T)", tokenIface)
	}

	// Get pach address and admin token from config
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if len(config.AdminToken) == 0 {
		return nil, errors.New("plugin is missing admin token")
	}
	if len(config.PachdAddress) == 0 {
		return nil, errors.New("plugin is missing pachd address")
	}

	// Revoke creds
	err = revokeUserCredentials(ctx, config.PachdAddress, userToken, config.AdminToken)
	if err != nil {
		return nil, err
	}

	return &logical.Response{}, nil
}

// revokeUserCredentials revokes the Pachyderm authentication token 'userToken'
// using the vault plugin's Admin credentials.
func revokeUserCredentials(ctx context.Context, pachdAddress string, userToken string, adminToken string) error {
	// Setup a single use client w the given admin token / address
	client, err := pclient.NewFromAddress(pachdAddress)
	if err != nil {
		return err
	}
	defer client.Close() // avoid leaking connections

	client = client.WithCtx(ctx)
	client.SetAuthToken(adminToken)
	_, err = client.AuthAPIClient.RevokeAuthToken(client.Ctx(), &auth.RevokeAuthTokenRequest{
		Token: userToken,
	})
	return err
}
