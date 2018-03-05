package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/fatih/structs"
	"github.com/hashicorp/vault/helper/pluginutil"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/hashicorp/vault/logical/plugin"
	pclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
)

type config struct {
	// AdminToken is pachyderm admin token used to generate credentials
	AdminToken string `json:"admin_token" structs:"-"`
}

func main() {
	apiClientMeta := &pluginutil.APIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:])

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := pluginutil.VaultPluginTLSProvider(tlsConfig)

	if err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: Factory,
		TLSProviderFunc:    tlsProviderFunc,
	}); err != nil {
		log.Fatal(err)
	}
}

type factory struct {
	*logical.Factory
}

//type Factory func(context.Context, *BackendConfig) (Backend, error)
func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	b, err := Backend(c)
	if err != nil {
		return nil, err
	}
	if err := b.Setup(ctx, c); err != nil {
		return nil, err
	}
	return b, nil
}

type backend struct {
	*framework.Backend
	PachydermClient *pclient.APIClient
}

func Backend(c *logical.BackendConfig) (*backend, error) {
	var b backend
	client, err := pclient.NewFromAddress("127.0.0.1:30650")
	if err != nil {
		return nil, err
	}
	b = backend{
		Backend: &framework.Backend{
			BackendType: logical.TypeCredential,
			AuthRenew:   b.pathAuthRenew,
			PathsSpecial: &logical.Paths{
				Unauthenticated: []string{"login"},
			},
			Paths: []*framework.Path{
				&framework.Path{
					Pattern: "login",
					Fields: map[string]*framework.FieldSchema{
						"username": &framework.FieldSchema{
							Type: framework.TypeString,
						},
					},
					Callbacks: map[logical.Operation]framework.OperationFunc{
						logical.UpdateOperation: b.pathAuthLogin,
					},
				},
				// auth/pachyderm/config
				&framework.Path{
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
				},
			},
		},
		PachydermClient: client,
	}

	return &b, nil
}
func (b *backend) generateUserCredentials(ctx context.Context, username string, adminToken string) (string, error) {
	// This is where we'd make the actual pachyderm calls to create the user
	// token using the admin token. For now, we just do an action that only an
	// admin could do
	b.PachydermClient.SetAuthToken(adminToken)
	_, err := b.PachydermClient.AuthAPIClient.ModifyAdmins(ctx, &auth.ModifyAdminsRequest{
		Add: []string{username},
	})

	if err != nil {
		return "", err
	}

	return "ARealLiveTemporaryAccessToken", nil
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

func (b *backend) pathAuthRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	if req.Auth == nil {
		return nil, errors.New("request auth was nil")
	}

	// Grab the token
	adminTokenRaw, ok := req.Auth.InternalData["pachyderm_admin_token"]
	if !ok {
		return nil, errors.New("no internal token found in the store")
	}
	adminToken, ok := adminTokenRaw.(string)
	if !ok {
		return nil, errors.New("stored admin token is not a string")
	}

	// Grab the token
	usernameRaw, ok := req.Auth.InternalData["username"]
	if !ok {
		return nil, errors.New("no internal token found in the store")
	}
	username, ok := usernameRaw.(string)
	if !ok {
		return nil, errors.New("stored admin token is not a string")
	}

	// Use the admin token to perform an action
	// for testing, hardcoding username to something else so that I can validate
	// renew has an effect:
	username = "tweetybird"
	userToken, err := b.generateUserCredentials(ctx, username, adminToken)
	if err != nil {
		return nil, err
	}
	fmt.Printf("generated new user token for (%v): %v\n", username, userToken)

	ttl, maxTTL, err := b.SanitizeTTLStr("30s", "1h")
	if err != nil {
		return nil, err
	}

	return framework.LeaseExtend(ttl, maxTTL, b.System())(ctx, req, d)
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

// errMissingField returns a logical response error that prints a consistent
// error message for when a required field is missing.
func errMissingField(field string) *logical.Response {
	return logical.ErrorResponse(fmt.Sprintf("Missing required field '%s'", field))
}

// validateFields verifies that no bad arguments were given to the request.
func validateFields(req *logical.Request, data *framework.FieldData) error {
	var unknownFields []string
	for k := range req.Data {
		if _, ok := data.Schema[k]; !ok {
			unknownFields = append(unknownFields, k)
		}
	}

	if len(unknownFields) > 0 {
		return fmt.Errorf("unknown fields: %q", unknownFields)
	}

	return nil
}

// Config parses and returns the configuration data from the storage backend.
func (b *backend) Config(ctx context.Context, s logical.Storage) (*config, error) {
	entry, err := s.Get(ctx, "config")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v: failed to get config from storage", err))
	}
	if entry == nil || len(entry.Value) == 0 {
		return nil, errors.New("no configuration in storage")
	}

	var result config
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, errors.New(fmt.Sprintf("%v: failed to decode configuration", err))
	}

	return &result, nil
}
