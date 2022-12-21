package log

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestCombineLogger(t *testing.T) {
	lctx, h := testWithCaptureParallel(t)
	cctx, c := context.WithCancel(context.Background())
	c()
	ctx := CombineLogger(cctx, lctx)
	select {
	case <-ctx.Done():
	default:
		t.Fatal("combined context is not done")
	}
	Info(ctx, "hi")
	h.HasALog(t)
}

func TestWithoutRatelimit(t *testing.T) {
	ctx, h := testWithCaptureParallelRateLimited(t, zap.Development())

	// Sanity check that the logger works and does rate limit.
	ctx = ChildLogger(ctx, "withFields", WithFields(zap.Bool("user", true)))
	for i := 0; i < 10000; i++ {
		Info(ctx, "with a field")
	}
	logs := h.Logs()
	if got := len(logs); got < 5 {
		t.Errorf("too few logs, even with rate limiting (got %v)", got)
	}
	log := logs[0]
	if got, want := log.String(), "withFields: info: with a field caller=log/context_test.go:32 user=true"; got != want {
		t.Errorf("first rate-limited:\n  got: %v\n want: %v", got, want)
	}
	if got := len(logs); got >= 9999 {
		t.Errorf("too many logs (got %v)", got)
	}

	// Ensure that the non-rate-limited logger has the fields/name from the parent.
	h.Clear()
	ctx = ChildLogger(ctx, "noLimit", WithoutRatelimit())
	for i := 0; i < 10000; i++ {
		Info(ctx, "not-ratelimited and with a field")
	}
	logs = h.Logs()
	if got, want := len(logs), 10000; got != want {
		t.Errorf("non-rate-limited log count:\n  got: %v\n want: %v", got, want)
	}
	log = logs[0]
	if got, want := log.String(), "withFields.noLimit: info: not-ratelimited and with a field caller=log/context_test.go:50 user=true"; got != want {
		t.Errorf("first rate-limited:\n  got: %v\n want: %v", got, want)
	}
}

func TestCloneOfWeirdCore(t *testing.T) {
	ctx, h := TestWithCapture(t)
	Info(ctx, "normal core")
	ctx = ChildLogger(ctx, "", WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewCore(zapcore.NewJSONEncoder(pachdEncoder), h, zapcore.DebugLevel)
	})))
	Info(ctx, "weird core")
	ctx = ChildLogger(ctx, "", WithoutRatelimit())
	Info(ctx, "weird core without rate limit")

	want := []string{
		"info: normal core",
		"info: weird core",
		"dpanic: attempt to remove rate limiting from a zap core that does not have rate limiting",
		"info: weird core without rate limit",
	}
	if diff := cmp.Diff(h.Logs(), want, formatLogs(simple)); diff != "" {
		t.Errorf("logs (-got +want):\n%s", diff)
	}
}
