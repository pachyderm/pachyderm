package log

import (
	"context"
)

// LogStep logs how long it takes to perform an operation.
//
// This used to be miscutil.LogStep.  New code prefers Span().
func LogStep(ctx context.Context, name string, cb func(ctx context.Context) error, fields ...Field) (retErr error) {
	ctx, end := spanContextL(ctx, name, DebugLevel, 1, 1, fields...)
	defer end(Errorp(&retErr))
	return cb(ctx)
}
