package log

import (
	"context"
	"log"

	"go.uber.org/zap"
)

// NewStdLog returns a *log.Logger (from the standard library "log" package) that logs to the provided context.
func NewStdLog(ctx context.Context) *log.Logger {
	return zap.NewStdLog(extractLogger(ctx).WithOptions(zap.AddCallerSkip(-1)))
}
