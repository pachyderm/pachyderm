package log

import (
	"context"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// pgxLogger is a logger for pgx to log SQL queries to.
type pgxLogger struct {
	name string
}

var _ pgx.Logger = new(pgxLogger)

// Log implements pgx.Logger.
func (pl *pgxLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]any) {
	// We do some work to avoid DPANICs about missing contexts, because a lot of our
	// SQL-handling code is missing contexts.  We also have some paranoia that ctx might be nil.
	var zl *zap.Logger
	if ctx != nil {
		if l, ok := ctx.Value(pachydermLogger{}).(*zap.Logger); ok {
			zl = l.Named(pl.name)
		}
	}
	if zl == nil {
		zl = zap.L().Named(pl.name).WithOptions(zap.AddCallerSkip(2))
	}

	// Check is a performance optimization to skip everything else if debug logs are disabled.
	// This avoids allocating the fields slice when debug logs are not going to be printed.
	//
	// Note: We always log at severity debug; pgx has the potential to cause alarm with its own
	// definition of errors, so we map everything down to Debug.
	e := zl.Check(zapcore.DebugLevel, msg)
	if e == nil {
		return
	}

	fields := []Field{zap.Stringer("pgx.severity", level)}
	for k, v := range data {
		fields = append(fields, zap.Any("pgx."+k, v))
	}
	fields = append(fields, ContextInfo(ctx))
	e.Write(fields...)
}

// NewPGX returns a new logger for pgx.
func NewPGX(name string) pgx.Logger { return &pgxLogger{name: name} }
