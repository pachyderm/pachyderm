package log

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level is the level at which to generate Span logs.
type Level int

const (
	DebugLevel Level = 1
	InfoLevel  Level = 2
	ErrorLevel Level = 3
)

func (l Level) log(z *zap.Logger, msg string, fields ...Field) {
	switch l { //exhaustive:enforce
	case DebugLevel:
		z.Debug(msg, fields...)
		return
	case InfoLevel:
		z.Info(msg, fields...)
		return
	case ErrorLevel:
		z.Error(msg, fields...)
		return
	}
	z.DPanic("log: internal error: unknown level in Level.log call", zap.Int("level", int(l)), zap.Stack("stack"))
	z.Debug(msg, fields...)
}

func (l Level) coreLevel() zapcore.Level {
	switch l { //exhaustive:enforce
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	}
	zap.L().DPanic("unknown log level in (log.Level).coreLevel", zap.Int("level", int(l)), zap.Stack("stack"))
	return zapcore.DebugLevel
}

// EndSpanFunc is a function that ends a span.
type EndSpanFunc = func(fields ...Field)

// ErrorL is a Field that marks a span as failed, and logs the end of the span at the provided level.
func ErrorL(err error, level Level) Field {
	if err == nil {
		return zap.Skip()
	}
	f := zap.Error(err)
	f.Integer = int64(level)
	return f
}

// ErrorpL is a Field that marks a span as failed if *err is a non-nil error at the time when the
// span ends.  See ErrorL.
func ErrorpL(err *error, level Level) Field {
	// If you're wondering why this ends in "pL" with that capitalization; zap uses the
	// convention 'lowercase p = pointer', and we use the convention 'uppercase L = level'.
	f := Errorp(err)
	f.Integer = int64(level)
	return f
}

const errorpType = zapcore.InlineMarshalerType + 100

// Errorp is a Field that marks a span as failed.  See ErrorpL.
func Errorp(err *error) Field {
	return zapcore.Field{
		Key:       "error",
		Type:      errorpType,
		Interface: err,
	}
}

type spanStatus string

const (
	spanStarting spanStatus = "span start"
	spanOK       spanStatus = "span finished ok"
	spanFailed   spanStatus = "span failed"
)

func makeSpanEndFunc(l *zap.Logger, event string, level Level, start time.Time) EndSpanFunc {
	return func(rawFields ...Field) {
		fields := []zap.Field{zap.Duration("spanDuration", time.Since(start))}
		msg := spanOK
		for _, f := range rawFields {
			if i := f.Interface; i != nil {
				// Handle ordinary zap.Error, zap.NamedError.  Extract Level if it's
				// there.
				if _, ok := i.(error); ok {
					msg = spanFailed
					if f.Type == zapcore.ErrorType && f.Integer > 0 {
						level = Level(f.Integer)
					}
					fields = append(fields, f)
					continue
				}
				// Handle ErrorpL/Errorp.
				if f.Type == errorpType {
					if errp, ok := i.(*error); ok {
						if *errp != nil {
							msg = spanFailed
							if f.Integer > 0 {
								level = Level(f.Integer)
							}
							fields = append(fields, zap.Error(*errp))
						}
					}
					continue // No errorpType fields should end up in fields.
				}
				// Note: zap.Errors wraps itself in a type that we can only
				// introspect with reflection.  It's not worth the compute cost to
				// look for; use multierr in zap.Error or ErrorL/ErrorP.
			}
			fields = append(fields, f)
		}
		level.log(l, event+": "+string(msg), fields...)
	}
}

// SpanContextL starts a new span, returning a context with a logger scoped to that span and a
// function to end the span.  A span is a simple way to mark the start and end of an operation.  To
// end a span in failure, pass a log.ErrorL(err, level), log.Error(err), log.Errors([]error),
// etc. to the end function.  If error is nil, the span is considered successful.
//
// The returned EndSpanFunc must be called from defer(), due to how Go stacks work.
//
// It is safe to use the returned context to log from multiple goroutines, and safe to log after the
// EndSpanFunc has been called.  (But might be confusing!)
func SpanContextL(rctx context.Context, event string, level Level, fields ...Field) (context.Context, EndSpanFunc) {
	return spanContextL(rctx, event, level, 1, 0, fields...)
}

func spanContextL(rctx context.Context, event string, level Level, startSkip, endSkip int, fields ...Field) (context.Context, EndSpanFunc) {
	deadlineField := zap.Skip()
	if dl, ok := rctx.Deadline(); ok {
		deadlineField = zap.Duration("timeout", time.Until(dl))
	}
	l := extractLogger(rctx).Named(event).With(fields...)
	level.log(l.WithOptions(zap.AddCallerSkip(1+startSkip)).With(deadlineField), event+": "+string(spanStarting))
	ctx := withLogger(rctx, l)
	return ctx, makeSpanEndFunc(l.WithOptions(zap.AddCallerSkip(1+endSkip)), event, level, time.Now())
}

// SpanContext starts a new span at level debug. See SpanContextL for details.
func SpanContext(rctx context.Context, event string, fields ...Field) (context.Context, EndSpanFunc) {
	return spanContextL(rctx, event, DebugLevel, 1, 0, fields...)
}

// SpanL starts a new span, returning a function that marks the end of the span. See SpanContextL
// for details.
func SpanL(ctx context.Context, event string, level Level, fields ...Field) EndSpanFunc {
	_, end := spanContextL(ctx, event, level, 1, 0, fields...)
	return end
}

// Span starts a new span at level debug.  See SpanContextL for details.
func Span(ctx context.Context, event string, fields ...Field) EndSpanFunc {
	_, end := spanContextL(ctx, event, DebugLevel, 1, 0, fields...)
	return end
}
