package pachyderm

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	pclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
)

func (b *backend) pathAuthRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	if req.Auth == nil {
		return nil, errors.New("request auth was nil")
	}

	config, err := b.Config(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if len(config.AdminToken) == 0 {
		return nil, errors.New("plugin is missing admin token")
	}
	if len(config.PachdAddress) == 0 {
		return nil, errors.New("plugin is missing pachd_address")
	}

	// Grab the user token
	userTokenRaw, ok := req.Auth.InternalData["user_token"]
	if !ok {
		return nil, errors.New("no internal user token found in the store")
	}
	userToken, ok := userTokenRaw.(string)
	if !ok {
		return nil, errors.New("stored user token is not a string")
	}

	ttl, maxTTL, err := b.SanitizeTTLStr(config.TTL, DefaultTTL)
	if err != nil {
		return nil, err
	}
	// Use the admin token to perform an action
	// for testing, hardcoding username to something else so that I can validate
	// renew has an effect:
	err = b.renewUserCredentials(ctx, config.PachdAddress, config.AdminToken, userToken, ttl)
	if err != nil {
		return nil, err
	}

	return framework.LeaseExtend(ttl, maxTTL, b.System())(ctx, req, d)
}

// renewUserCredentials extends the TTL of the Pachyderm authentication token
// 'userToken', using the vault plugin's Admin credentials. 'userToken' belongs
// to the user who is calling vault, and would like to extend their Pachyderm
// session.
func (b *backend) renewUserCredentials(ctx context.Context, pachdAddress string, adminToken string, userToken string, ttl time.Duration) error {
	// Setup a single use client w the given admin token / address
	client, err := pclient.NewFromAddress(pachdAddress)
	if err != nil {
		return err
	}
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
