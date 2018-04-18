package pachyderm

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	pclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
)

func (b *backend) revokePath() *framework.Path {
	return &framework.Path{
		Pattern: "revoke",
		Fields: map[string]*framework.FieldSchema{
			"user_token": &framework.FieldSchema{
				Type: framework.TypeString,
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathRevoke,
		},
	}
}

func (b *backend) pathRevoke(ctx context.Context, req *logical.Request, data *framework.FieldData) (resp *logical.Response, retErr error) {
	b.Logger().Debug(fmt.Sprintf("(%s) %s received at %s", req.ID, req.Operation, req.Path))
	defer func() {
		b.Logger().Debug(fmt.Sprintf("(%s) %s finished at %s with result (success=%t)", req.ID, req.Operation, req.Path, retErr == nil && !resp.IsError()))
	}()

	userToken, errResp := getStringField(data, "user_token")
	if errResp != nil {
		return errResp, nil
	}

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
	client = client.WithCtx(ctx)
	client.SetAuthToken(adminToken)
	_, err = client.AuthAPIClient.RevokeAuthToken(client.Ctx(), &auth.RevokeAuthTokenRequest{
		Token: userToken,
	})
	return err
}
