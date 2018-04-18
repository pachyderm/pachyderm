package pachyderm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	pclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
)

// renew renews the caller's credentials (and extends the TTL of their Pachyderm
// token by sending a request to Pachyderm). Unlike other handlers, it doesn't
// get assigned to a path; instead it's placed in Backend.AuthRenew in
// backend.go
func (b *backend) renew(ctx context.Context, req *logical.Request, d *framework.FieldData) (resp *logical.Response, retErr error) {
	// renew seems to be handled specially, and req.ID doesn't seem to be set
	b.Logger().Debug(fmt.Sprintf("%s received at %s", req.Operation, req.Path))
	defer func() {
		b.Logger().Debug(fmt.Sprintf("%s finished at %s (success=%t)", req.Operation, req.Path, retErr == nil && !resp.IsError()))
	}()

	if req.Auth == nil {
		return nil, errors.New("request auth was nil")
	}

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

	userTokenIface, ok := req.Auth.InternalData["user_token"]
	if !ok {
		return nil, errors.New("no internal user token found in the store")
	}
	userToken, ok := userTokenIface.(string)
	if !ok {
		return nil, errors.New("stored user token is not a string")
	}

	ttl, maxTTL, err := b.SanitizeTTLStr(config.TTL, b.System().MaxLeaseTTL().String())
	if err != nil {
		return nil, fmt.Errorf("%v: could not sanitize config TTL", err)
	}
	err = renewUserCredentials(ctx, config.PachdAddress, config.AdminToken, userToken, ttl)
	if err != nil {
		return nil, err
	}

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
