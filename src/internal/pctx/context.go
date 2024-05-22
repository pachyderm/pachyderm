package pctx

import (
	"context"
	"os"
	"os/signal"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/meters"
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

// Interactive returns a context for use in command-line apps.
func Interactive() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(log.AddLogger(context.Background()), os.Interrupt)
}

// Option is an option for customizing a child context.
type Option struct {
	modifyContext func(context.Context) context.Context
	modifyLogger  log.LogOption
	addsFields    bool
}

// WithServerID generates a server ID and attaches it to the context.  It appears on each log
// produced by the child.
func WithServerID() Option {
	return Option{
		modifyLogger: log.WithServerID(),
		addsFields:   true,
	}
}

// WithFields returns a context that includes additional fields that appear on each log line.
func WithFields(fields ...zap.Field) Option {
	return Option{
		modifyLogger: log.WithFields(fields...),
		addsFields:   true,
	}
}

// WithOptions returns a context that modifies the logger with additional Zap options.
func WithOptions(opts ...zap.Option) Option {
	return Option{
		modifyLogger: log.WithOptions(opts...),
		addsFields:   true, // it might not, but saying it does is most safe
	}
}

// WithoutRatelimit returns a context that does not rate limit its logs.
func WithoutRatelimit() Option {
	return Option{
		modifyLogger: log.WithoutRatelimit(),
	}
}

// WithGauge adds an aggregated gauge metric to the context.
func WithGauge[T any](metric string, zero T, options ...meters.Option) Option {
	return Option{
		modifyContext: func(ctx context.Context) context.Context {
			return meters.NewAggregatedGauge(ctx, metric, zero, options...)
		},
	}
}

// WithCounter adds an aggregated counter metric to the context.
func WithCounter[T meters.Monoid](metric string, zero T, options ...meters.Option) Option {
	return Option{
		modifyContext: func(ctx context.Context) context.Context {
			return meters.NewAggregatedCounter(ctx, metric, zero, options...)
		},
	}
}

// WithDelta adds an aggregated delta metric to the context.
func WithDelta[T meters.Signed](metric string, threshold T, options ...meters.Option) Option {
	return Option{
		modifyContext: func(ctx context.Context) context.Context {
			return meters.NewAggregatedDelta(ctx, metric, threshold, options...)
		},
	}
}

// Child returns a named child context, with additional options.  The new name can be empty.
// Options are applied in an arbitrary order.
//
// Calling Child with a name adds the name to the current name, separated by a dot.  For example,
// Child(Child(Background(), "PPS"), "master") would result in logs and metrics in the PPS.master
// namespace.  It is okay to add interesting data to these names, like
// fmt.Sprintf("pipelineController(%s)", pipeline.Name).  We prefer camel case names, and () to
// quote dynamic data.
func Child(ctx context.Context, name string, opts ...Option) context.Context {
	var logOptions []log.LogOption
	var newAggregatesNeeded bool
	for _, opt := range opts {
		if opt.addsFields {
			newAggregatesNeeded = true
			break
		}
	}
	if name != "" {
		newAggregatesNeeded = true
	}
	if newAggregatesNeeded {
		ctx = meters.WithNewFields(ctx)
	}
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
