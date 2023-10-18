package pctx

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap/zaptest"
)

type testOptions struct {
	loggingOptions []zaptest.LoggerOption
}

// TestOption allows you to supply configuration tweaks to TestContext.
type TestOption func(o *testOptions)

// WithZapTestOptions supplies additional zaptest.LoggerOptions to the logger returned by
// TestContext.  Raw zap options can be supplied by wrapping them with zaptest.WrapOptions, but note
// that behavior can be weird.  Most tests will not need any logging options.
func WithZapTestOptions(opts ...zaptest.LoggerOption) TestOption {
	return func(o *testOptions) {
		o.loggingOptions = append(o.loggingOptions, opts...)
	}
}

// TestContext returns a context suitable for use as the root context of tests.
func TestContext(t testing.TB, options ...TestOption) context.Context {
	ctx, cf := context.WithCancel(context.Background())
	t.Cleanup(cf)
	var opts testOptions
	for _, o := range options {
		o(&opts)
	}
	return log.TestParallel(ctx, t, opts.loggingOptions...)
}
