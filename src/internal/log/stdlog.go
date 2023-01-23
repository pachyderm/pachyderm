package log

import (
	"context"
	"log"

	"go.uber.org/zap"
)

// NewStdLogAt returns a *log.Logger (from the standard library "log" package) that logs to the
// provided context, at the provided level.
func NewStdLogAt(ctx context.Context, lvl Level) *log.Logger {
	// NewStdLogAt cannot return an error for any of the log levels that lvl.coreLevel()
	// returns, since coreLevel will change the level to Debug if it's unknown.
	l, _ := zap.NewStdLogAt(extractLogger(ctx).WithOptions(zap.AddCallerSkip(-1)), lvl.coreLevel())
	return l
}
