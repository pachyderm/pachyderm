package pachyderm

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

type config struct {
	// AdminToken is pachyderm admin token used to generate credentials
	AdminToken string `json:"admin_token" structs:"-"`

	// PachdAddress is the hostport at which the client can reach Pachyderm
	PachdAddress string `json:"pachd_address" structs:"-"`

	// TTL defines the time to live for any tokens that will be provided
	// It is a duration string, e.g. "23s", "5m", "4h"
	TTL string `json:"ttl" structs:"-"`
}

func (b *backend) configPath() *framework.Path {
	return &framework.Path{
		Pattern:      "config",
		HelpSynopsis: "Configure the admin token",
		HelpDescription: `
Read or write configuration to Vault's storage backend to specify the Pachyderm admin token. For example:

    $ vault write auth/pachyderm/config \
        admin_token="xxx" \
        pachd_address="127.0.0.1:30650" \
        ttl="20m" # ttl is optional

For more information and examples, please see the online documentation.
`,
		Fields: map[string]*framework.FieldSchema{
			"admin_token": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Pachyderm admin token used to generate user credentials",
			},
			"pachd_address": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Pachyderm cluster address, e.g. 127.0.0.1:30650",
			},
			"ttl": &framework.FieldSchema{
				Type:        framework.TypeDurationSecond,
				Description: "Max TTL for any tokens issued",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathConfigWrite,
			logical.ReadOperation:   b.pathConfigRead,
		},
	}
}

func (b *backend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (resp *logical.Response, retErr error) {
	b.Logger().Debug(fmt.Sprintf("(%s) %s received at %s", req.ID, req.Operation, req.Path))
	defer func() {
		b.Logger().Debug(fmt.Sprintf("(%s) %s finished at %s (success=%t)", req.ID, req.Operation, req.Path, retErr == nil && !resp.IsError()))
	}()

	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get configuration from storage")
	}
	respData := make(map[string]interface{})
	respData["admin_token"] = config.AdminToken
	respData["pachd_address"] = config.PachdAddress
	respData["ttl"] = config.TTL
	return &logical.Response{
		Data: respData,
	}, nil
}

func (b *backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (resp *logical.Response, retErr error) {
	b.Logger().Debug(fmt.Sprintf("(%s) %s received at %s", req.ID, req.Operation, req.Path))
	defer func() {
		b.Logger().Debug(fmt.Sprintf("(%s) %s finished at %s (success=%t)", req.ID, req.Operation, req.Path, retErr == nil && !resp.IsError()))
	}()

	// Validate we didn't get extraneous fields
	if err := validateFields(req, data); err != nil {
		return nil, logical.CodedError(422, err.Error())
	}

	// Extract relevant fields from the request
	adminToken, errResp := getStringField(data, "admin_token")
	if errResp != nil {
		return errResp, nil
	}
	if adminToken == "" {
		return logical.ErrorResponse("invalid admin_token: empty string"), nil
	}
	pachdAddress, errResp := getStringField(data, "pachd_address")
	if errResp != nil {
		return errResp, nil
	}
	if pachdAddress == "" {
		return logical.ErrorResponse("invalid pachd_address: empty string"), nil
	}

	// Try to extract TTL from request
	ttl := b.System().DefaultLeaseTTL()
	if errResp = func() *logical.Response {
		ttlIface, ok, err := data.GetOkErr("ttl")
		if err != nil {
			return logical.ErrorResponse(fmt.Sprintf("%v: could not extract 'ttl' from request", err))
		}
		if !ok {
			return nil // ttl is unset -- no error, just use default
		}

		// try to parse TTL as an int (
		ttlSeconds, ok := ttlIface.(int)
		if !ok {
			return logical.ErrorResponse(fmt.Sprintf("invalid type for param 'ttl' (expected int but got %T)", ttlIface))
		}
		if ttlSeconds <= 0 {
			return logical.ErrorResponse("invalid TTL duration (must be > 0s)")
		}
		ttl = time.Duration(ttlSeconds) * time.Second
		// clamp TTL to vault's max lease
		if ttl > b.System().MaxLeaseTTL() {
			ttl = b.System().MaxLeaseTTL()
		}
		return nil
	}(); errResp != nil {
		return errResp, nil
	}

	// Build the entry and write it to vault storage
	if err := putConfig(ctx, req.Storage, &config{
		AdminToken:   adminToken,
		PachdAddress: pachdAddress,
		TTL:          ttl.String(),
	}); err != nil {
		return logical.ErrorResponse(fmt.Sprintf("%v: could not put config", err)), nil
	}
	return &logical.Response{}, nil
}
