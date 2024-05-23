package meters

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
)

func TestImmediate(t *testing.T) {
	ctx := log.TestParallel(context.Background(), t)
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
		if i%2 == 0 {
			Set(ctx, "test", 43) // No log expected.
		} else {
			SetGauge(ctx, "test", 43) // No log expected.
		} // Both operations above do the same thing on this numeric gauge.
	}
	SetGauge(ctx, "test", "string") // Log expected because "string" doesn't match the gauge type.
	Inc(ctx, "test", 44)            // Log expected because this is a non-gauge operation.
	Sample(ctx, "test", 123)        // Log expected because this is a non-gauge operation.
	for i := 0; i < 1000; i++ {
		Set(ctx, "test", i) // 1 log expected after flush.
	}
	c()
	<-doneCh

	var got []any
	for _, l := range h.Logs() {
		if l.Message == "meter: test" {
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
		t.Errorf("diff logged meters (-got +want):\n%s", diff)
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
		if l.Message == "meter: test" {
			if x, ok := l.Keys["delta"]; ok {
				got = append(got, x)
			}
		} else {
			t.Errorf("unexpected log line: %v", l.String())
		}
	}
	want := []any{float64(10000)}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("diff logged meters (-got +want):\n%s", diff)
	}
}

func TestAggregatedDelta(t *testing.T) {
	ctx, h := log.TestWithCapture(t)
	ctx, c := context.WithCancel(ctx)
	doneCh := make(chan struct{})
	ctx = NewAggregatedDelta(ctx, "test", 500, withDoneCh(doneCh))
	for i := 0; i < 1000; i++ {
		Set(ctx, "test", i)
	}
	c()
	<-doneCh
	var got []string
	for _, l := range h.Logs() {
		if l.Message == "meter: test" {
			if x, ok := l.Keys["value"]; ok {
				got = append(got, fmt.Sprintf("value: %v", int(x.(float64))))
			}
			if x, ok := l.Keys["delta"]; ok {
				got = append(got, fmt.Sprintf("delta: %v", int(x.(float64))))
			}
		} else {
			t.Errorf("unexpected log line: %v", l.String())
		}
	}
	want := []string{"value: 500", "delta: 499"} // last value is 999
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("diff logged meters (-got +want):\n%s", diff)
		t.Logf("logs: %v", h.Logs())
	}
}

func TestWithNewFields(t *testing.T) {
	ctx, h := log.TestWithCapture(t)
	ctx, c := context.WithCancel(ctx)
	ctx = log.ChildLogger(ctx, "a", log.WithFields(zap.Bool("a", true)))
	t.Logf("ctx: %s", reflect.ValueOf(ctx))
	ctxA := NewAggregatedCounter(ctx, "test", 0, Deferred())
	ctxA = NewAggregatedGauge(ctxA, "gauge", 0, Deferred())
	t.Logf("ctxA: %s", reflect.ValueOf(ctxA))
	ctxB := WithNewFields(log.ChildLogger(ctxA, "b", log.WithFields(zap.Bool("b", true))))
	t.Logf("ctxB: %s", reflect.ValueOf(ctxB))
	for i := 0; i < 1000; i++ {
		Inc(ctxA, "test", 1)
		Set(ctxA, "gauge", 42)
		Inc(ctxB, "test", -1)
		// If WithNewFields doesn't work, the value of the meter after c() will be 0.  If
		// it does, there will be 1000 for A and -1000 for B.
	}
	c()
	time.Sleep(100 * time.Millisecond) // can't use doneCh because of the cloning

	var gotLines, gotA, gotB, gotGauge int
	wantLines := 3
	for _, l := range h.Logs() {
		switch l.Message {
		case "meter: test":
			if x, ok := l.Keys["delta"]; ok {
				if _, ok := l.Keys["b"]; ok {
					gotLines++
					gotB = int(x.(float64))
				} else {
					gotLines++
					gotA = int(x.(float64))
				}
			}
		case "meter: gauge":
			if _, ok := l.Keys["b"]; !ok {
				if x, ok := l.Keys["value"]; ok {
					gotLines++
					gotGauge = int(x.(float64))
				}
			}
		}
	}
	if got, want := gotLines, wantLines; got != want {
		t.Errorf("line count:\n  got: %v\n want: %v", got, want)
		for _, l := range h.Logs() {
			t.Logf("log: %s", l)
		}
	}
	if got, want := gotA, 1000; got != want {
		t.Errorf("value of test{a}:\n  got: %v\n want: %v", got, want)
	}
	if got, want := gotB, -1000; got != want {
		t.Errorf("value of test{b}:\n  got: %v\n want: %v", got, want)
	}
	if got, want := gotGauge, 42; got != want {
		t.Errorf("value of gauge:\n  got: %v\n want: %v", got, want)
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
	// Write meters from 10 goroutines, so the writes contend with each other.
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
