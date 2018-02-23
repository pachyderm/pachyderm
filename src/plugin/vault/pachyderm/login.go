package pachyderm

import (
	"context"
	"errors"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/pachyderm/pachyderm/src/client/auth"
)

func (b *backend) loginPath() *framework.Path {

	return &framework.Path{
		Pattern: "login",
		Fields: map[string]*framework.FieldSchema{
			"username": &framework.FieldSchema{
				Type: framework.TypeString,
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathAuthLogin,
		},
	}
}

func (b *backend) pathAuthLogin(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	username := d.Get("username").(string)
	if len(username) == 0 {
		return nil, logical.ErrInvalidRequest
	}

	config, err := b.Config(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if len(config.AdminToken) == 0 {
		return nil, errors.New("plugin is missing admin token")
	}

	// Use the admin token to perform an action
	userToken, err := b.generateUserCredentials(ctx, username, config.AdminToken)
	if err != nil {
		return nil, err
	}

	ttl, _, err := b.SanitizeTTLStr("30s", "1h")
	if err != nil {
		return nil, err
	}

	// Compose the response
	return &logical.Response{
		Auth: &logical.Auth{
			InternalData: map[string]interface{}{
				"pachyderm_admin_token": config.AdminToken,
				"username":              username,
			},
			Metadata: map[string]string{
				"user_token": userToken,
			},
			LeaseOptions: logical.LeaseOptions{
				TTL:       ttl,
				Renewable: true,
			},
		},
	}, nil
}

func (b *backend) generateUserCredentials(ctx context.Context, username string, adminToken string) (string, error) {
	// This is where we'd make the actual pachyderm calls to create the user
	// token using the admin token. For now, we just do an action that only an
	// admin could do

	// Setup a single use client w the given auth token
	pachClient := b.PachydermClient.WithCtx(ctx)
	pachClient.SetAuthToken(adminToken)

	_, err := b.PachydermClient.AuthAPIClient.ModifyAdmins(pachClient.Ctx(), &auth.ModifyAdminsRequest{
		Add: []string{username},
	})

	if err != nil {
		return "", err
	}

	return "ARealLiveTemporaryAccessToken", nil
}
