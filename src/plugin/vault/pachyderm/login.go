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
		return nil, errors.New("plugin is missing admin_token")
	}
	if len(config.PachdAddress) == 0 {
		return nil, errors.New("plugin is missing pachd_address")
	}
	if len(config.TTL) == 0 {
		return nil, errors.New("plugin is missing ttl")
	}

	ttl, _, err := b.SanitizeTTLStr(config.TTL, DefaultTTL)
	if err != nil {
		return nil, err
	}

	userToken, err := b.generateUserCredentials(ctx, config.PachdAddress, config.AdminToken, username, ttl)
	if err != nil {
		return nil, err
	}

	// Compose the response
	return &logical.Response{
		Auth: &logical.Auth{
			InternalData: map[string]interface{}{
				"user_token": userToken,
			},
			Metadata: map[string]string{
				"user_token":    userToken,
				"pachd_address": config.PachdAddress,
			},
			LeaseOptions: logical.LeaseOptions{
				TTL:       ttl,
				Renewable: true,
			},
		},
	}, nil
}

// generateUserCredentials uses the vault plugin's Admin credentials to generate
// a new Pachyderm authentication token for 'username' (i.e. the user who is
// currently requesting a Pachyderm token from Vault).
func (b *backend) generateUserCredentials(ctx context.Context, pachdAddress string, adminToken string, username string, ttl time.Duration) (string, error) {
	// Setup a single use client w the given admin token / address
	client, err := pclient.NewFromAddress(pachdAddress)
	if err != nil {
		return "", err
	}
	client = client.WithCtx(ctx)
	client.SetAuthToken(adminToken)
	fmt.Printf("login using auth token (%v)\n", adminToken)

	resp, err := client.AuthAPIClient.GetAuthToken(client.Ctx(), &auth.GetAuthTokenRequest{
		Subject: username,
		TTL:     int64(ttl.Seconds()),
	})
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}
