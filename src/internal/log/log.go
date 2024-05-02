package log

import (
	"context"
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"
)

// Field is a typed log field.  It is lazily evaluated at log marshaling time.
type Field = zap.Field

// Debug logs a message, with fields, at level DEBUG.  Level debug is appropriate for messages that
// are interesting to Pachyderm developers, but not operators or users; or logs that are generated
// by internal action at a rate of more than 1 message per second.
//
// Most tracing and background operations are DEBUG logs.
func Debug(ctx context.Context, msg string, fields ...Field) {
	l := extractLogger(ctx)
	if e := l.Check(zapcore.DebugLevel, msg); e != nil {
		fields = append(fields, ContextInfo(ctx))
		e.Write(fields...)
	}
}

// Info logs a message, with fields, at level INFO.  Level info is appropriate for messages that are
// interesting to operators of Pachyderm (like information about applying configuration changes), or
// that represent user action (like incoming RPC requests).
//
// Most warnings and user-caused errors are INFO logs.
func Info(ctx context.Context, msg string, fields ...Field) {
	fields = append(fields, ContextInfo(ctx))
	extractLogger(ctx).Info(msg, fields...)
}

// Error logs a message, with fields, at level ERROR.  Level error is appropriate for messages that
// indicate a malfunction that an operator of Pachyderm can repair.
func Error(ctx context.Context, msg string, fields ...Field) {
	fields = append(fields, ContextInfo(ctx))
	extractLogger(ctx).Error(msg, fields...)
}

// DPanic is a message that panics in development, but is only logged in production.
func DPanic(ctx context.Context, msg string, fields ...Field) {
	fields = append(fields, ContextInfo(ctx))
	extractLogger(ctx).DPanic(msg, fields...)
}

// Exit logs a message, with fields, at level FATAL and then exits with status 1.  Level fatal is
// only appropriate for use in interactive scripts.
func Exit(ctx context.Context, msg string, fields ...Field) {
	fields = append(fields, ContextInfo(ctx))
	extractLogger(ctx).Fatal(msg, fields...)
}

// WriterAt creates a new io.Writer that logs each line as a log message at the provided levels.
func WriterAt(ctx context.Context, lvl Level) io.WriteCloser {
	l := extractLogger(ctx).WithOptions(zap.WithCaller(false))
	return &zapio.Writer{
		Log:   l,
		Level: lvl.coreLevel(),
	}
}
