package log

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
)

// These two tests test our public API, but those log to t.Log, not to somewhere we can check.  So
// you have to run these in verbose mode and see if the logs make sense.  Kind of useless, but if
// you think Test is broken, at least you can manually debug it.
func TestTest(t *testing.T) {
	ctx := Test(context.Background(), t)
	Debug(ctx, "hello")
}

func TestTestParallel(t *testing.T) {
	c := make(chan struct{})

	t.Run("a", func(t *testing.T) {
		t.Parallel()
		ctx := TestParallel(context.Background(), t)
		Debug(ctx, "a")
		Debug(nil, "not seen (a)") //nolint:SA1012 // intentional nil context
		close(c)
	})
	t.Run("b", func(t *testing.T) {
		t.Parallel()
		ctx := TestParallel(context.Background(), t)
		<-c
		Debug(ctx, "b")
		Debug(nil, "not seen (b)") //nolint:SA1012 // intentional nil context
	})
}

func TestParallelTestLogging(t *testing.T) {
	var ha, hb *history
	_, hg := TestWithCapture(t, zap.Development())
	ch := make(chan struct{}) // channel to ensure tests actually run in parallel
	doneCh := make(chan struct{})

	t.Run("a", func(t *testing.T) {
		t.Parallel()
		var ctx context.Context
		ctx, ha = testWithCaptureParallel(t, zap.Development())
		Debug(ctx, "hello from a")
		<-ch
		Debug(ctx, "hello number 2 from a")
		<-ch
		zap.L().Info("global log from a")
		<-ch
		doneCh <- struct{}{}
	})
	t.Run("b", func(t *testing.T) {
		t.Parallel()
		var ctx context.Context
		ctx, hb = testWithCaptureParallel(t, zap.Development())
		ch <- struct{}{} // hello from a
		Debug(ctx, "hello from b")
		ch <- struct{}{} // hello number 2 from a
		Debug(ctx, "hello number 2 from b")
		ch <- struct{}{} // "global log from a"
		zap.L().Info("global log from b")
		doneCh <- struct{}{}
	})
	t.Run("finalize", func(t *testing.T) {
		t.Parallel()
		<-doneCh
		<-doneCh
		close(ch)
		close(doneCh)

		want := []string{
			"debug: hello from a",
			"debug: hello number 2 from a",
		}
		if diff := cmp.Diff(ha.Logs(), want, formatLogs(simple)); diff != "" {
			t.Errorf("test a's logs (-got +want):\n%s", diff)
		}

		want = []string{
			"debug: hello from b",
			"debug: hello number 2 from b",
		}
		if diff := cmp.Diff(hb.Logs(), want, formatLogs(simple)); diff != "" {
			t.Errorf("test b's logs (-got +want):\n%s", diff)
		}

		want = []string{
			"info: global log from a",
			"info: global log from b",
		}
		if diff := cmp.Diff(hg.Logs(), want, formatLogs(simple)); diff != "" {
			t.Errorf("global logs (-got +want):\n%s", diff)
		}
	})
}
