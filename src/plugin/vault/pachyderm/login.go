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

func (b *backend) loginPath() *framework.Path {

	return &framework.Path{
		Pattern: "login",
		Fields: map[string]*framework.FieldSchema{
			"username": &framework.FieldSchema{
				Type: framework.TypeString,
			},
			"ttl": &framework.FieldSchema{
				Type: framework.TypeString,
			},
			"max_ttl": &framework.FieldSchema{
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
	ttlString := d.Get("ttl").(string)
	if len(ttlString) == 0 {
		ttlString = "45s"
	}
	maxTtlString := d.Get("max_ttl").(string)
	if len(maxTtlString) == 0 {
		maxTtlString = "2h"
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

	ttl, _, err := b.SanitizeTTLStr(ttlString, maxTtlString)
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
				"ttl":        ttlString,
				"max_ttl":    maxTtlString,
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

func (b *backend) generateUserCredentials(ctx context.Context, pachdAddress string, adminToken string, username string, ttl time.Duration) (string, error) {
	// This is where we'd make the actual pachyderm calls to create the user
	// token using the admin token. For now, for testing purposes, we just do an action that only an
	// admin could do

	// Setup a single use client w the given admin token / address
	client, err := pclient.NewFromAddress(pachdAddress)
	if err != nil {
		return "", err
	}
	client = client.WithCtx(ctx)
	client.SetAuthToken(adminToken)

	_, err = client.AuthAPIClient.ModifyAdmins(client.Ctx(), &auth.ModifyAdminsRequest{
		Add: []string{username},
	})

	if err != nil {
		return "", err
	}

	return "ARealLiveTemporaryAccessToken", nil
}
