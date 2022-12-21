package pctx

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
)

// TODO returns a context for use in Pachyderm, for code that will be updated to take a proper
// context in the near future.  It should not be used in new code.
func TODO() context.Context {
	return log.AddLogger(context.TODO())
}

// Background returns a context for use in long-running background processes.  If the background
// process needs to inherit something other than a clean non-cancelable context, use Child instead.
func Background(process string) context.Context {
	ctx := log.AddLogger(context.Background())
	return Child(ctx, process)
}

// Option is an option for customizing a child context.
type Option struct {
	modifyContext func(context.Context) context.Context
	modifyLogger  log.LogOption
}

// WithServerID generates a server ID and attaches it to the context.  It appears on each log
// produced by the child.
func WithServerID() Option {
	return Option{
		modifyLogger: log.WithServerID(),
	}
}

// WithFields returns a context that includes additional fields that appear on each log line.
func WithFields(fields ...zap.Field) Option {
	return Option{
		modifyLogger: log.WithFields(fields...),
	}
}

// WithOptions returns a context that modifies the logger with additional Zap options.
func WithOptions(opts ...zap.Option) Option {
	return Option{
		modifyLogger: log.WithOptions(opts...),
	}
}

// WithoutRatelimit returns a context that does not rate limit its logs.
func WithoutRatelimit() Option {
	return Option{
		modifyLogger: log.WithoutRatelimit(),
	}
}

// Child returns a named child context, with additional options.  The new name can be empty.
// Options are applied in an arbitrary order.
func Child(ctx context.Context, name string, opts ...Option) context.Context {
	var logOptions []log.LogOption
	for _, opt := range opts {
		if o := opt.modifyLogger; o != nil {
			logOptions = append(logOptions, o)
		}
		if o := opt.modifyContext; o != nil {
			ctx = o(ctx)
		}
	}
	ctx = log.ChildLogger(ctx, name, logOptions...)
	return ctx
}
