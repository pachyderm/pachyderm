package pachyderm

import (
	"context"
	"errors"
	"fmt"

	"github.com/fatih/structs"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

type config struct {
	// AdminToken is pachyderm admin token used to generate credentials
	AdminToken string `json:"admin_token" structs:"-"`
}

func (b *backend) configPath() *framework.Path {

	return &framework.Path{
		Pattern:      "config",
		HelpSynopsis: "Configure the admin token",
		HelpDescription: `

Read or writer configuration to Vault's storage backend to specify the Pachyderm admin token. For example:

    $ vault write auth/pachyderm/config \
        admin_token="xxx"

For more information and examples, please see the online documentation.

`,

		Fields: map[string]*framework.FieldSchema{
			"admin_token": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Pachyderm admin token used to generate user credentials",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathConfigWrite,
			logical.ReadOperation:   b.pathConfigRead,
		},
	}
}

// pathConfigRead corresponds to READ auth/pachyderm/config.
func (b *backend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := b.Config(ctx, req.Storage)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v: failed to get configuration from storage", err))
	}

	resp := &logical.Response{
		Data: structs.New(config).Map(),
	}
	return resp, nil
}

// pathConfigRead corresponds to POST auth/pachyderm/config.
func (b *backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	// Validate we didn't get extraneous fields
	if err := validateFields(req, data); err != nil {
		return nil, logical.CodedError(422, err.Error())
	}

	// Get the admin token
	adminToken := data.Get("admin_token").(string)
	if adminToken == "" {
		return errMissingField("admin_token"), nil
	}

	// Built the entry
	entry, err := logical.StorageEntryJSON("config", &config{
		AdminToken: adminToken,
	})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v: failed to generate storage entry", err))
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, errors.New(fmt.Sprintf("%v: failed to write configuration to storage", err))
	}
	return nil, nil
}
