package pachyderm

import (
	"context"
	"errors"

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

func (b *backend) pathRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	userToken := d.Get("user_token").(string)
	if len(userToken) == 0 {
		return nil, logical.ErrInvalidRequest
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

	err = b.revokeUserCredentials(ctx, config.PachdAddress, userToken, config.AdminToken)
	if err != nil {
		return nil, err
	}

	// Compose the response
	// TODO: Not sure if this is the right way to return a successful response
	return &logical.Response{
		Auth: &logical.Auth{},
	}, nil
}

func (b *backend) revokeUserCredentials(ctx context.Context, pachdAddress string, userToken string, adminToken string) error {
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
