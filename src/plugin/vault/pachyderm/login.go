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

func (b *backend) loginPath() *framework.Path {
	return &framework.Path{
		// Pattern uses modified version of framework.GenericNameRegex which
		// requires a single colon
		Pattern: "login/(?P<username>\\w[\\w-]*:[\\w-]*\\w)",
		Fields: map[string]*framework.FieldSchema{
			"username": &framework.FieldSchema{
				Type: framework.TypeString,
			},
			"ttl": &framework.FieldSchema{
				Type: framework.TypeString,
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathAuthLogin,
		},
	}
}

func (b *backend) pathAuthLogin(ctx context.Context, req *logical.Request, d *framework.FieldData) (resp *logical.Response, retErr error) {
	b.Logger().Debug(fmt.Sprintf("(%s) %s received at %s", req.ID, req.Operation, req.Path))
	defer func() {
		b.Logger().Debug(fmt.Sprintf("(%s) %s finished at %s (success=%t)", req.ID, req.Operation, req.Path, retErr == nil && !resp.IsError()))
	}()

	username, errResp := getStringField(d, "username")
	if errResp != nil {
		return errResp, nil
	}
	var ttlArg string
	ttlArgIface, ok, err := d.GetOkErr("ttl")
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("%v: could not extract 'ttl' from request", err)), nil
	}
	if ok {
		ttlArg, ok = ttlArgIface.(string)
		if !ok {
			return logical.ErrorResponse(fmt.Sprintf("invalid type for param 'ttl' (expected string but got %T)", ttlArgIface)), nil
		}
	}

	config, err := getConfig(ctx, req.Storage)
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

	var ttl time.Duration
	if ttlArg != "" {
		ttl, _, err = sanitizeTTLStr(ttlArg, b.System().MaxLeaseTTL().String())
	} else {
		ttl, _, err = sanitizeTTLStr(config.TTL, b.System().MaxLeaseTTL().String())
	}
	if err != nil {
		return nil, err
	}

	userToken, err := generateUserCredentials(ctx, config.PachdAddress, config.AdminToken, username, ttl)
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Secret: &logical.Secret{
			InternalData: map[string]interface{}{
				"user_token":  userToken,
				"secret_type": "pachyderm_tokens",
			},
			LeaseOptions: logical.LeaseOptions{
				TTL:       ttl,
				Renewable: true,
			},
		},
		Data: map[string]interface{}{
			"user_token":    userToken,
			"pachd_address": config.PachdAddress,
		},
	}, nil
}

// generateUserCredentials uses the vault plugin's Admin credentials to generate
// a new Pachyderm authentication token for 'username' (i.e. the user who is
// currently requesting a Pachyderm token from Vault).
func generateUserCredentials(ctx context.Context, pachdAddress string, adminToken string, username string, ttl time.Duration) (string, error) {
	// Setup a single use client w the given admin token / address
	client, err := pclient.NewFromAddress(pachdAddress)
	if err != nil {
		return "", err
	}
	defer client.Close() // avoid leaking connections

	client = client.WithCtx(ctx)
	client.SetAuthToken(adminToken)

	resp, err := client.AuthAPIClient.GetAuthToken(client.Ctx(), &auth.GetAuthTokenRequest{
		Subject: username,
		TTL:     int64(ttl.Seconds()),
	})
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}

func sanitizeTTLStr(ttlStr, maxTTLStr string) (ttl, maxTTL time.Duration, err error) {
	if len(ttlStr) == 0 || ttlStr == "0" {
		ttl = 0
	} else {
		ttl, err = time.ParseDuration(ttlStr)
		if err != nil {
			return 0, 0, errors.Wrapf(err, "invalid ttl")
		}
	}

	if len(maxTTLStr) == 0 || maxTTLStr == "0" {
		maxTTL = 0
	} else {
		maxTTL, err = time.ParseDuration(maxTTLStr)
		if err != nil {
			return 0, 0, errors.Wrapf(err, "invalid max_ttl")
		}
	}

	return ttl, maxTTL, nil
}
