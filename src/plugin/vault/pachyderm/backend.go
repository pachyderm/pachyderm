package pachyderm

import (
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

type backend struct {
	*framework.Backend
}

func Backend(c *logical.BackendConfig) (*backend, error) {
	var b backend
	b = backend{
		Backend: &framework.Backend{
			BackendType: logical.TypeLogical,
			AuthRenew:   b.pathAuthRenew,
			PathsSpecial: &logical.Paths{
				Unauthenticated: []string{"login"},
			},
			Paths: []*framework.Path{
				b.configPath(),
				b.loginPath(),
				b.revokePath(),
			},
		},
	}

	return &b, nil
}
