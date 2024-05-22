// Package pachctl contains utilities for implementing pachctl commands.
package pachctl

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	ci "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging/client"
)

type Config struct {
	Verbose bool
	Timeout time.Duration
}

func (cfg *Config) NewOnUserMachine(ctx context.Context, enterprise bool, opts ...client.Option) (*client.APIClient, error) {
	if cfg.Verbose {
		opts = append(opts, client.WithAdditionalStreamClientInterceptors(ci.LogStream), client.WithAdditionalUnaryClientInterceptors(ci.LogUnary))
	}

	cancel := func() {}
	if cfg.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
	}

	var c *client.APIClient
	var err error
	if enterprise {
		c, err = client.NewEnterpriseClientOnUserMachineContext(ctx, "user", opts...)
	} else {
		c, err = client.NewOnUserMachine(ctx, "user", opts...)
	}
	if err != nil {
		cancel()
		return nil, err
	}
	if enterprise {
		fmt.Fprintf(os.Stderr, "Using enterprise context: %v\n", c.ClientContextName())
	}
	return c.WithCtxCancel(ctx, cancel), nil
}

func (cfg *Config) NewInWorker(ctx context.Context, opts ...client.Option) (*client.APIClient, error) {
	if cfg.Verbose {
		opts = append(opts, client.WithAdditionalStreamClientInterceptors(ci.LogStream), client.WithAdditionalUnaryClientInterceptors(ci.LogUnary))
	}
	cancel := func() {}
	if cfg.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
	}
	c, err := client.NewInWorker(ctx, opts...)
	if err != nil {
		cancel()
		return nil, err
	}
	return c.WithCtxCancel(ctx, cancel), nil
}
