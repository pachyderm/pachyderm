package pctx

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
)

// TestContext returns a context suitable for use as the root context of tests.
func TestContext(t testing.TB) context.Context {
	return log.TestParallel(t)
}
