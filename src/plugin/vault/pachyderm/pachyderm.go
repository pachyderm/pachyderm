package pachyderm

import (
	"context"
	"time"

	/*
			"github.com/hashicorp/vault/helper/pluginutil"
			"github.com/hashicorp/vault/logical"
			"github.com/hashicorp/vault/logical/framework"
		"github.com/hashicorp/vault/helper/strutil"
		"github.com/hashicorp/vault/plugins/helper/database/connutil"
		"github.com/hashicorp/vault/plugins/helper/database/credsutil"
		"github.com/hashicorp/vault/plugins/helper/database/dbutil"
	*/
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/builtin/logical/database/dbplugin"
	"github.com/hashicorp/vault/plugins"
)

const (
	pachydermTypeName string = "pachyderm"
)

// New implements builtinplugins.BuiltinFactory
func New() (interface{}, error) {

	// TODO: URL should be provided by config
	client, err := client.NewFromAddress("127.0.0.1:30650")
	if err != nil {
		return nil, err
	}
	// TODO: provide adminToken via config
	client.SetAuthToken("adminToken")
	p := &Pachyderm{
		client: client,
	}

	return p, nil
}

// Run instantiates a Pachyderm object, and runs the RPC server for the plugin
func Run(apiTLSConfig *api.TLSConfig) error {
	dbType, err := New()
	if err != nil {
		return err
	}

	plugins.Serve(dbType.(*Pachyderm), apiTLSConfig)

	return nil
}

type Pachyderm struct {
	client *client.APIClient
}

func (p *Pachyderm) Type() (string, error) {
	return pachydermTypeName, nil
}

func (p *Pachyderm) CreateUser(ctx context.Context, statements dbplugin.Statements, usernameConfig dbplugin.UsernameConfig, expiration time.Time) (username string, password string, err error) {

	username = "username"

	_, err = p.client.AuthAPIClient.ModifyAdmins(p.client.Ctx(), &auth.ModifyAdminsRequest{
		Add: []string{username},
	})

	if err != nil {
		return "", "", err
	}

	return "username", "password", nil
}

func (p *Pachyderm) RenewUser(ctx context.Context, statements dbplugin.Statements, username string, expiration time.Time) error {
	return nil
}

func (p *Pachyderm) RevokeUser(ctx context.Context, statements dbplugin.Statements, username string) error {
	return nil
}
