package pachyderm

import (
	"context"

	"github.com/hashicorp/vault/logical"
)

// Factory is the function that the Pachyderm Vault plugin exports to let Vault
// create/refresh/revoke Pachyderm tokens
func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	b, err := newBackend(c)
	if err != nil {
		return nil, err
	}
	if err := b.Setup(ctx, c); err != nil {
		return nil, err
	}
	return b, nil
}
