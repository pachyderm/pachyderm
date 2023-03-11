package m

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
)

func TestImmediate(t *testing.T) {
	ctx := log.TestParallel(t)
	Set(ctx, "gauge", 42)
	Inc(ctx, "counter", 1)
	Inc(ctx, "counter", 1)
	Inc(ctx, "counter", -1)
	Sample(ctx, "sample", "a")
	Sample(ctx, "sample", "b")
}

func TestAggregatedGauge(t *testing.T) {
	ctx, h := log.TestWithCapture(t)
	ctx, c := context.WithCancel(ctx)
	doneCh := make(chan struct{})
	ctx = NewAggregatedGauge(ctx, "test", 0, withDoneCh(doneCh)) // No log expected.
	Set(ctx, "test", 42)                                         // No log expected.
	for i := 0; i < 1000; i++ {
		Set(ctx, "test", 43) // No log expected.
	}
	Set(ctx, "test", "string") // Log expected because "string" doesn't match the gauge type.
	Inc(ctx, "test", 44)       // Log expected because this is a non-gauge operation.
	Sample(ctx, "test", 123)   // Log expected because this is a non-gauge operation.
	for i := 0; i < 1000; i++ {
		Set(ctx, "test", i) // 1 log expected after flush.
	}
	c()
	<-doneCh

	var got []any
	for _, l := range h.Logs() {
		if l.Message == "metric: test" {
			if x, ok := l.Keys["value"]; ok {
				got = append(got, x)
			}
			if x, ok := l.Keys["delta"]; ok {
				got = append(got, x)
			}
			if x, ok := l.Keys["sample"]; ok {
				got = append(got, x)
			}
		} else {
			t.Errorf("unexpected log line: %v", l.String())
		}
	}
	want := []any{"string", float64(44), float64(123), float64(999)} // float64 because of JSON -> Go conversion in the log history.
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("diff logged metrics (-got +want):\n%s", diff)
	}
}

func TestAggregatedCounter(t *testing.T) {
	ctx, h := log.TestWithCapture(t)
	ctx, c := context.WithCancel(ctx)
	doneCh := make(chan struct{})
	ctx = NewAggregatedCounter(ctx, "test", 0, withDoneCh(doneCh))
	for i := 0; i < 10000; i++ {
		Inc(ctx, "test", 1) // 1 log expected after flush.
	}
	c()
	<-doneCh

	var got []any
	for _, l := range h.Logs() {
		if l.Message == "metric: test" {
			if x, ok := l.Keys["delta"]; ok {
				got = append(got, x)
			}
		} else {
			t.Errorf("unexpected log line: %v", l.String())
		}
	}
	want := []any{float64(10000)}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("diff logged metrics (-got +want):\n%s", diff)
	}

}

func BenchmarkGauge(b *testing.B) {
	ctx, w := log.NewBenchLogger(true) // log rate limiting = true is the best case
	for i := 0; i < b.N; i++ {
		Set(ctx, "bench", i)
	}
	if w.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkAggregatedGauge(b *testing.B) {
	ctx, w := log.NewBenchLogger(false)
	ctx, c := context.WithCancel(ctx)
	doneCh := make(chan struct{})
	ctx = NewAggregatedGauge(ctx, "bench", 0, withDoneCh(doneCh), WithFlushInterval(100*time.Millisecond))
	writesDoneCh := make(chan struct{})
	// Write metrics from 10 goroutines, so the writes contend with each other.
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1+b.N/10; j++ {
				Set(ctx, "bench", j)
			}
			writesDoneCh <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-writesDoneCh
	}
	close(writesDoneCh)
	c()
	<-doneCh
	if w.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func TestWithNewFields(t *testing.T) {
	ctx, h := log.TestWithCapture(t)
	ctx, c := context.WithCancel(ctx)
	ctxA := NewAggregatedCounter(ctx, "test", 0, Deferred())
	ctxB := WithNewFields(log.ChildLogger(ctxA, "b", log.WithFields(zap.Bool("b", true))))
	for i := 0; i < 1000; i++ {
		Inc(ctxA, "test", 1)
		Inc(ctxB, "test", -1)
		// If WithNewFields doesn't work, the value of the metric after c() will be 0.  If
		// it does, there will be 1000 for A and -1000 for B.
	}
	c()
	time.Sleep(100 * time.Millisecond) // can't use doneCh because of the cloning

	var gotA, gotB int
	for i, l := range h.Logs() {
		if l.Message == "metric: test" {
			if x, ok := l.Keys["delta"]; ok {
				if _, ok := l.Keys["b"]; ok {
					gotB = int(x.(float64))
				} else {
					gotA = int(x.(float64))
				}
			}
		} else {
			t.Errorf("unexpected log line: %v", l.String())
		}
		if i > 1 {
			t.Errorf("unexpected extra line: %v", l.String())
		}
	}
	if got, want := gotA, 1000; got != want {
		t.Errorf("value of a:\n  got: %v\n want: %v", got, want)
	}
	if got, want := gotB, -1000; got != want {
		t.Errorf("value of b:\n  got: %v\n want: %v", got, want)
	}
}
