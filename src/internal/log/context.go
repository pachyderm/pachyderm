package log

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// General note: if you're adding a plain logger to a context here, you are responsible for setting
// up caller skip when you call withLogger.  A caller skip of 1 is traditional, because the public
// API is to call this package's Error/Info/Debug methods, which add 1 stack frame from the actual
// loggable event.
//
// Basically, if you extract a logger from a context, it has caller skip for use with this package.
// If you have a raw *zap.Logger, it probably doesn't.

type pachydermLogger struct{}

// CombineLogger extracts a logger from loggerCtx, and associates the logger with ctx.  It's only
// public until grpc interceptors move into this package.
func CombineLogger(ctx, loggerCtx context.Context) context.Context {
	l := extractLogger(loggerCtx)
	return context.WithValue(ctx, pachydermLogger{}, l)
}

// withLogger adds a zap logger to the context.  The embedded logger is used by this package's
// logging functions.  It would be unusual to use this outside of a new RPC protocol's handler or
// background task setup code.
func withLogger(ctx context.Context, l *zap.Logger) context.Context {
	if l == nil {
		zap.L().WithOptions(zap.AddCallerSkip(1)).DPanic("log: internal error: nil logger provided to WithLogger", zap.Stack("stack"))
		return ctx
	}
	return context.WithValue(ctx, pachydermLogger{}, l)
}

// extractLogger returns the logger in the provided context, or the global logger.  The returned
// logger can never be nil.  Most code should not use this directly; use Debug/Info/Error directly.
//
// Remember that the logger in the context has CallerSkip set to 1, so is only really useful when
// used with Debug/Info/Error methods.
func extractLogger(ctx context.Context) *zap.Logger {
	// CallerSkip is 2 because this wants to panic at the caller of the caller of extractLogger.
	// For example, you'd get here by calling Error(ctx, ...); Error is our caller (1), the
	// caller of Error (2) is where we want the caller field of this diagnostic message to be
	// from.
	const skip = 2
	if ctx == nil {
		zap.L().WithOptions(zap.AddCallerSkip(skip)).DPanic("log: internal error: nil context provided to ExtractLogger", zap.Stack("stack"))
		return zap.L().WithOptions(zap.AddCallerSkip(1))
	}
	if l := ctx.Value(pachydermLogger{}); l != nil {
		return l.(*zap.Logger) // let this panic; indicates a bug in this package
	}
	zap.L().WithOptions(zap.AddCallerSkip(skip)).DPanic("log: internal error: no logger in provided context", zap.Stack("stack"))
	return zap.L().WithOptions(zap.AddCallerSkip(1))
}

// LogOption is an option to add to a logger.
type LogOption func(l *zap.Logger) *zap.Logger

// WithServerID adds a server ID to each message logged by the logger.
func WithServerID() LogOption {
	return func(l *zap.Logger) *zap.Logger {
		return l.With(zap.Stringer("server-id", uuid(backgroundTrace)))
	}
}

// WithFields adds fields to each message logged by the logger.
func WithFields(f ...Field) LogOption {
	return func(l *zap.Logger) *zap.Logger {
		return l.With(f...)
	}
}

// WithOptions adds zap options to the logger.
func WithOptions(o ...zap.Option) LogOption {
	return func(l *zap.Logger) *zap.Logger {
		return l.WithOptions(o...)
	}
}

// WithoutRatelimit removes any rate limiting on the child logger.  Essential for things like user
// code output.
func WithoutRatelimit() LogOption {
	return func(l *zap.Logger) *zap.Logger {
		return l.WithOptions(zap.WrapCore(func(old zapcore.Core) zapcore.Core {
			if c, ok := cloneWithSampling(old, false); ok {
				return c
			}
			zap.L().DPanic("attempt to remove rate limiting from a zap core that does not have rate limiting", zap.Stack("stack"))
			return old
		}))
	}
}

// AddLogger is used by the pctx package to create empty contexts.  It should not be used outside of
// the pctx package.
func AddLogger(ctx context.Context) context.Context {
	return withLogger(ctx, zap.L().WithOptions(zap.AddCallerSkip(1)))
}

// ChildLogger is used by the pctx package to create child contexts.  It should not be used outside of the
// pctx package.
func ChildLogger(ctx context.Context, name string, opts ...LogOption) context.Context {
	l := extractLogger(ctx)
	for _, o := range opts {
		l = o(l)
	}
	if name != "" {
		l = l.Named(name)
	}
	return withLogger(ctx, l)
}
