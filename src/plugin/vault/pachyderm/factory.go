package pachyderm

import (
	"context"

	"github.com/hashicorp/vault/logical"
)

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
