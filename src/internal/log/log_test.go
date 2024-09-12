package log

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
)

func TestBasics(t *testing.T) {
	ctx, h := TestWithCapture(t)
	var want []string
	Debug(ctx, "hello")
	want = append(want, "debug: hello")
	Info(ctx, "hello")
	want = append(want, "info: hello")
	Error(ctx, "hello")
	want = append(want, "error: hello")
	log.Println("this is the builtin go logging")
	want = append(want, "info: this is the builtin go logging")

	if diff := cmp.Diff(h.Logs(), want, formatLogs(simple)); diff != "" {
		t.Errorf("logs (-got +want):\n%s", diff)
	}
}

func TestPanics(t *testing.T) {
	testData := []struct {
		name string
		f    func(l *zap.Logger) // A scenario that should panic in development, but warn in production.
	}{
		{
			name: "nil context",
			f:    func(_ *zap.Logger) { Debug(nil, "this should panic") }, //nolint:SA1012 // intentionally testing nil context
		},
		{
			name: "empty context",
			f: func(_ *zap.Logger) {
				Debug(context.Background(), "this should also panic")
			},
		},
		{
			name: "nil logger",
			f:    func(_ *zap.Logger) { withLogger(context.Background(), nil) },
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			var msg string
			func() {
				l, _ := newTestLogger(t, false, zap.Development())
				defer zap.ReplaceGlobals(l)()
				defer func() {
					err := recover()
					if e, ok := err.(string); ok {
						msg = e
					}
				}()
				test.f(l)
			}()
			if msg == "" {
				t.Errorf("should have panicked in development")
			}
			func() {
				l, h := newTestLogger(t, false)
				defer zap.ReplaceGlobals(l)()
				test.f(l)
				if testing.Verbose() {
					for _, msg := range h.Logs() {
						t.Logf("captured log <%s: %s>", msg.Severity, msg.Message)
					}
				}
			}()
		})
	}
}

func TestNoPanicsInProduction(t *testing.T) {
	_, h := TestWithCapture(t)
	Debug(nil, "this should use the global logger") //nolint:SA1012 // Intentional nil to test error handling.
	want := []string{
		"dpanic: log: internal error: nil context provided to ExtractLogger",
		"debug: this should use the global logger",
	}
	if diff := cmp.Diff(h.Logs(), want, formatLogs(simple)); diff != "" {
		t.Errorf("logs (-got +want):\n%s", diff)
	}
}

func TestWrappedContext(t *testing.T) {
	rootCtx, h := testWithCaptureParallel(t, zap.Development())
	timeCtx, tc := context.WithTimeout(rootCtx, time.Minute)
	t.Cleanup(tc)
	ctx, cc := context.WithCancel(timeCtx)
	t.Cleanup(cc)
	Debug(ctx, "this should not panic")

	want := []string{"debug: this should not panic"}
	if diff := cmp.Diff(h.Logs(), want, formatLogs(simple)); diff != "" {
		t.Errorf("logs (-got +want):\n%s", diff)
	}
}

func TestEmptyContext(t *testing.T) {
	_, h := TestWithCapture(t)
	ctx := context.Background() // intentionally the wrong context
	var want []string
	Debug(ctx, "this is a debug log")
	want = append(want,
		"dpanic: log: internal error: no logger in provided context",
		"debug: this is a debug log")
	if diff := cmp.Diff(h.Logs(), want, formatLogs(simple)); diff != "" {
		t.Errorf("logs (-got +want):\n%s", diff)
	}
}
