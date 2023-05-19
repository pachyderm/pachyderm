// Package pachctl contains utilities for implementing pachctl commands.
package pachctl

import (
	"context"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
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
		c, err = client.NewEnterpriseClientOnUserMachineContext(ctx, "user", opts...)
	} else {
		c, err = client.NewOnUserMachine(ctx, "user", opts...)
	}
	if err != nil {
		return nil, err
	}
	if enterprise {
		fmt.Fprintf(os.Stderr, "Using enterprise context: %v\n", c.ClientContextName())
	}
	return c.WithCtx(ctx), nil
}

func (cfg *Config) NewInWorker(ctx context.Context, opts ...client.Option) (*client.APIClient, error) {
	if cfg.Verbose {
		opts = append(opts, client.WithAdditionalStreamClientInterceptors(ci.LogStream), client.WithAdditionalUnaryClientInterceptors(ci.LogUnary))
	}
	c, err := client.NewInWorker(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return c.WithCtx(ctx), nil
}
