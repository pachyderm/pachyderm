// Package pachctl contains utilities for implementing pachctl commands.
package pachctl

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/client"
	ci "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging/client"
)

type Config struct {
	Verbose bool
}

func (cfg *Config) NewOnUserMachine(ctx context.Context, enterprise bool, opts ...client.Option) (*client.APIClient, error) {
	if cfg.Verbose {
		opts = append(opts, client.WithAdditionalStreamClientInterceptors(ci.LogStream), client.WithAdditionalUnaryClientInterceptors(ci.LogUnary))
	}

	var c *client.APIClient
	var err error
	if enterprise {
		c, err = client.NewEnterpriseClientOnUserMachine("user", opts...)
	} else {
		c, err = client.NewOnUserMachine("user", opts...)
	}
	if err != nil {
		return nil, err
	}
	return c.WithCtx(ctx), nil
}
