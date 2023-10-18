package log

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"golang.org/x/exp/maps"
)

// Test returns a new Context appropriate for use in tests.  Unless you are for some reason reliant
// upon the global logger, use pctx.TestContext(t).
func Test(ctx context.Context, t testing.TB, opts ...zaptest.LoggerOption) context.Context {
	l := zaptest.NewLogger(t, opts...)
	t.Cleanup(zap.ReplaceGlobals(l))
	t.Cleanup(zap.RedirectStdLog(l))
	t.Cleanup(zap.RedirectStdLog(l))
	return ctx
}

// Test returns a new Context appropriate for use in parallel tests, at the cost of not logging
// messages sent to the global logger.  This function is only public so that pctx.TestContext(t) can
// call it, and you should use that.
func TestParallel(ctx context.Context, t testing.TB, opts ...zaptest.LoggerOption) context.Context {
	lvl := zap.NewAtomicLevelAt(zapcore.DebugLevel)
	opts = append(opts,
		zaptest.WrapOptions(zap.AddCaller(), zap.AddCallerSkip(1)),
		zaptest.Level(lvl),
	)
	l := zaptest.NewLogger(t, opts...)
	t.Cleanup(func() {
		// TODO(jonathan): After the tests stop running, disable logging.  It seems like
		// etcd holds on to its client logger long after it is asked to close, and then it
		// logs errors, and Go causes tests to panic if they log after they are done
		// running.  The underlying problem is that the PPS master is slow to die; we can
		// rework tests to use context.WithCancel and cancel() instead of reducing the log
		// level here.
		lvl.SetLevel(zapcore.FatalLevel)
	})
	return withLogger(ctx, l)
}

// msg is a log message parsed from pachd-formatted JSON.
type msg struct {
	Time     time.Time
	Severity string
	Logger   string
	Caller   string
	Message  string
	Keys     map[string]any
	Orig     json.RawMessage
}
type _msg msg

func (m *msg) UnmarshalJSON(data []byte) error {
	var copy _msg
	var orig []byte
	orig = append(orig, data...)
	if err := json.Unmarshal(data, &copy.Keys); err != nil {
		return errors.Wrap(err, "unmarshal into _msg.Keys")
	}
	if err := json.Unmarshal(data, &copy); err != nil {
		return errors.Wrap(err, "unmarshal into _msg")
	}
	delete(copy.Keys, "time")
	delete(copy.Keys, "severity")
	delete(copy.Keys, "logger")
	delete(copy.Keys, "message")
	m.Time = copy.Time
	m.Severity = copy.Severity
	m.Logger = copy.Logger
	m.Caller = copy.Caller
	m.Message = copy.Message
	m.Keys = copy.Keys
	m.Orig = orig
	return nil
}

type formatter struct {
	name string
	f    func(m *msg) string
}

func buildSimpleMessage(b *strings.Builder, m *msg) {
	if m.Logger != "" {
		b.WriteString(m.Logger)
		b.WriteString(": ")
	}
	b.WriteString(m.Severity)
	b.WriteString(": ")
	b.WriteString(m.Message)
}

var (
	simple = formatter{
		name: "simple",
		f: func(m *msg) string {
			b := new(strings.Builder)
			buildSimpleMessage(b, m)
			return b.String()
		},
	}

	keys = formatter{
		name: "keys",
		f: func(m *msg) string {
			keys := maps.Keys(m.Keys)
			sort.Strings(keys)
			return strings.Join(keys, ",")
		},
	}

	simpleWithKeys = formatter{
		name: "simpleWithKeys",
		f: func(m *msg) string {
			b := new(strings.Builder)
			buildSimpleMessage(b, m)
			keys := maps.Keys(m.Keys)
			sort.Strings(keys)
			for _, k := range keys {
				b.WriteByte(' ')
				v := m.Keys[k]
				fmt.Fprintf(b, "%s=%v", k, v)
			}
			return b.String()
		},
	}
)

func (m *msg) Stringify(rule formatter) string {
	return rule.f(m)
}

func (m *msg) String() string {
	return simpleWithKeys.f(m)
}

func isSliceOf[T any](x any) bool {
	_, ok := x.([]T)
	return ok
}

// formatLogs returns a cmp.Option that transforms log entries using the provided template.
func formatLogs(f formatter) cmp.Option {
	return cmp.FilterValues(func(x, y any) bool {
		return (isSliceOf[*msg](x) && isSliceOf[string](y)) ||
			(isSliceOf[*msg](y) && isSliceOf[string](x))
	}, cmp.Transformer(f.name, func(x any) []string {
		if ms, ok := x.([]*msg); ok {
			var result []string
			for _, m := range ms {
				result = append(result, f.f(m))
			}
			return result
		}
		if ss, ok := x.([]string); ok {
			return ss
		}
		panic("filter failed")
	}))
}

// History is a sequence of messages logged during a test.
type history struct {
	sync.Mutex
	logs []*msg // You must hold the mutex to read logs.
}

// Write implements io.Writer.  Each write call must be a full JSON message to unmarshal.
func (h *history) Write(p []byte) (int, error) {
	var m msg
	if err := json.Unmarshal(p, &m); err != nil {
		return 0, errors.Wrap(err, "unmarshal log message")
	}
	h.Lock()
	h.logs = append(h.logs, &m)
	h.Unlock()
	return len(p), nil
}

// Logs returns all of the logs in the history.
func (h *history) Logs() []*msg {
	h.Lock()
	defer h.Unlock()
	return h.logs
}

// HasALog asserts that the logger logged a log.
func (h *history) HasALog(t testing.TB) {
	t.Helper()
	if got := len(h.Logs()); got < 1 {
		t.Errorf("number of logs:\n  got: %v\n want: %v", got, ">0")
	}
}

// Clear clears the logging history.
func (h *history) Clear() {
	h.Lock()
	h.logs = nil
	h.Unlock()
}

// Sync implements zapcore.WriteSyncer.
func (h *history) Sync() error { return nil }

// failer is a zapcore.SyncWriter that fails the test if logged to; used as a zap.ErrorOutput, fails
// the test on write errors.
type failer struct{ testing.TB }

func (t failer) Write(p []byte) (int, error) {
	t.Helper()
	t.Errorf("internal logging error: %s", p)
	return len(p), nil
}

func (t failer) Sync() error { return nil }

// TestWithCapture returns a new Context appropriate for testing the logging system itself; in
// addition to a context, you also get a capture object that retains all logged message for
// analysis.  The underling logger is not rate-limited as a production logger is.  Note: The logged
// messages are not printed out to the test log.
//
// Most logging functionality that needs tests should be in this package; this function is only
// public for the use of pctx.
func TestWithCapture(t testing.TB, opts ...zap.Option) (context.Context, *history) {
	l, h := newTestLogger(t, false, opts...)
	t.Cleanup(zap.ReplaceGlobals(l))
	t.Cleanup(zap.RedirectStdLog(l))
	return withLogger(context.Background(), l.WithOptions(zap.AddCallerSkip(1))), h
}

// testWithCaptureParallel is the same as TestWithCapture, but does not touch the global loggers,
// allowing tests that want to run in parallel to capture logs for their test.  (At the cost of
// missing zap.L() logs, etc.)
func testWithCaptureParallel(t testing.TB, opts ...zap.Option) (context.Context, *history) {
	l, h := newTestLogger(t, false, opts...)
	return withLogger(context.Background(), l.WithOptions(zap.AddCallerSkip(1))), h
}

// testWithCaptureParallelRateLimited is the same as testWithCaptureParallel but has a rate limit.
func testWithCaptureParallelRateLimited(t testing.TB, opts ...zap.Option) (context.Context, *history) {
	l, h := newTestLogger(t, true, opts...)
	t.Cleanup(func() { droppedLogs.Store(0) })
	return withLogger(context.Background(), l.WithOptions(zap.AddCallerSkip(1))), h
}

// newTestLogger returns the raw logger used for testWithCapture*
// Remeber that raw *zap.Loggers don't have caller skip; testWith... above add those.
func newTestLogger(t testing.TB, sample bool, opts ...zap.Option) (*zap.Logger, *history) {
	h := new(history)
	opts = append(opts,
		zap.ErrorOutput(&failer{t}),
		zap.AddCaller(),
	)
	l := makeLogger(zapcore.NewJSONEncoder(pachdEncoder), h, zapcore.DebugLevel, sample, opts)
	return l, h
}

type byteCounter struct {
	c atomic.Int64
}

func (c *byteCounter) Write(p []byte) (int, error) {
	c.c.Add(int64(len(p)))
	return len(p), nil
}

func (c *byteCounter) Sync() error { return nil }

// NewBenchLogger returns a logger suitable for benchmarks and an *atomic.Int64 containing the
// number of bytes written to the logger.
func NewBenchLogger(sample bool) (context.Context, *atomic.Int64) {
	enc := zapcore.NewJSONEncoder(pachdEncoder)
	w := new(byteCounter)
	l := makeLogger(enc, zapcore.Lock(w), zapcore.DebugLevel, sample, []zap.Option{zap.AddCaller()})
	return withLogger(context.Background(), l), &w.c
}
