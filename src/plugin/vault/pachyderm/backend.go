package pachyderm

import (
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	pclient "github.com/pachyderm/pachyderm/src/client"
)

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
		PachydermClient: client,
	}

	return &b, nil
}
