package pachyderm

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/helper/parseutil"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

// DefaultTTL defines the default if none is specified
const DefaultTTL = "5m"

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

Read or writer configuration to Vault's storage backend to specify the Pachyderm admin token. For example:

    $ vault write auth/pachyderm/config \
        admin_token="xxx" \
		pachd_address="127.0.0.1:30650" \
		ttl="20m"

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

func (b *backend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := b.Config(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("%v: failed to get configuration from storage", err)
	}
	respData := make(map[string]interface{})
	respData["admin_token"] = config.AdminToken
	respData["pachd_address"] = config.PachdAddress
	respData["ttl"] = config.TTL
	resp := &logical.Response{
		Data: respData,
	}
	return resp, nil
}

func (b *backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	// Validate we didn't get extraneous fields
	if err := validateFields(req, data); err != nil {
		return nil, logical.CodedError(422, err.Error())
	}

	adminToken := data.Get("admin_token").(string)
	if adminToken == "" {
		return errMissingField("admin_token"), nil
	}
	pachdAddress := data.Get("pachd_address").(string)
	if pachdAddress == "" {
		return errMissingField("pachd_address"), nil
	}
	ttlDuration, err := parseutil.ParseDurationSecond(data.Get("ttl"))
	if err != nil {
		return nil, fmt.Errorf("error parsing duration (%v): %v", data.Get("ttl"), err)
	}
	ttl := DefaultTTL
	// An empty input param results in a "0s" duration
	if ttlDuration.Seconds() > 0 {
		ttl = ttlDuration.String()
	}
	// Built the entry
	entry, err := logical.StorageEntryJSON("config", &config{
		AdminToken:   adminToken,
		PachdAddress: pachdAddress,
		TTL:          ttl,
	})
	if err != nil {
		return nil, fmt.Errorf("%v: failed to generate storage entry", err)
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, fmt.Errorf("%v: failed to write configuration to storage", err)
	}
	return nil, nil
}
