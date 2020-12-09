package testutil

import (
	"context"
	"testing"
)

// NewTestContext returns a context scoped to the current test
// it will have a deadline if ok in: `deadline, ok := t.Deadline()`
func NewTestContext(t *testing.T) context.Context {
	ctx := context.Background()
	var cf context.CancelFunc
	if deadline, ok := t.Deadline(); ok {
		ctx, cf = context.WithDeadline(ctx, deadline)
	}
	t.Cleanup(func() {
		cf()
	})
	return ctx
}
