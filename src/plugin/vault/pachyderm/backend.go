package pachyderm

import (
	"context"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

type backend struct {
	*framework.Backend
}

// Factory is the function that the Pachyderm Vault plugin exports to let Vault
// create/refresh/revoke Pachyderm tokens
func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	result := &backend{}
	result.Backend = &framework.Backend{
		BackendType: logical.TypeLogical,
		AuthRenew:   result.renew,
		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{"login"},
		},
		Paths: []*framework.Path{
			result.configPath(),
			result.loginPath(),
			result.revokePath(),
		},
	}
	if err := result.Setup(ctx, c); err != nil {
		return nil, err
	}
	return result, nil
}
