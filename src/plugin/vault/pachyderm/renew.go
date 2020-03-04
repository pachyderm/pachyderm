package pachyderm

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	pclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

// Renew renews the caller's credentials (and extends the TTL of their Pachyderm
// token by sending a request to Pachyderm). Unlike other handlers, it doesn't
// get assigned to a path; instead it's called by the vault lease API when a
// token's lease is renewed. It's set in Backend.Secrets[0].Revoke in backend.go
func (b *backend) Renew(ctx context.Context, req *logical.Request, d *framework.FieldData) (resp *logical.Response, retErr error) {
	// renew seems to be handled specially, and req.ID doesn't seem to be set
	b.Logger().Debug(fmt.Sprintf("%s received at %s", req.Operation, req.Path))
	defer func() {
		b.Logger().Debug(fmt.Sprintf("%s finished at %s (success=%t)", req.Operation, req.Path, retErr == nil && !resp.IsError()))
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
		return nil, errors.New("plugin is missing pachd_address")
	}

	// Extract TTL from request first, and then config (if unset)
	ttl, maxTTL := req.Secret.LeaseOptions.Increment, b.System().MaxLeaseTTL()
	if ttl == 0 {
		ttl, maxTTL, err = sanitizeTTLStr(config.TTL, maxTTL.String())
		if err != nil {
			return nil, errors.Wrapf(err, "could not sanitize config TTL")
		}
	}

	// Renew creds in Pachyderm
	err = renewUserCredentials(ctx, config.PachdAddress, config.AdminToken, userToken, ttl)
	if err != nil {
		return nil, err
	}

	// Renew vault lease
	return framework.LeaseExtend(ttl, maxTTL, b.System())(ctx, req, d)
}

// renewUserCredentials extends the TTL of the Pachyderm authentication token
// 'userToken', using the vault plugin's Admin credentials. 'userToken' belongs
// to the user who is calling vault, and would like to extend their Pachyderm
// session.
func renewUserCredentials(ctx context.Context, pachdAddress string, adminToken string, userToken string, ttl time.Duration) error {
	// Setup a single use client w the given admin token / address
	client, err := pclient.NewFromAddress(pachdAddress)
	if err != nil {
		return err
	}
	defer client.Close() // avoid leaking connections

	client = client.WithCtx(ctx)
	client.SetAuthToken(adminToken)

	_, err = client.AuthAPIClient.ExtendAuthToken(client.Ctx(), &auth.ExtendAuthTokenRequest{
		Token: userToken,
		TTL:   int64(ttl.Seconds()),
	})

	if err != nil {
		return err
	}

	return nil
}
