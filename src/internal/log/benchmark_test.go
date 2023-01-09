package log

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type count struct {
	n int64
}

func (c *count) Write(p []byte) (int, error) {
	atomic.AddInt64(&c.n, int64(len(p)))
	return len(p), nil
}

func (c *count) Sync() error { return nil }

func newBenchLogger(sample bool) (context.Context, *count) {
	enc := zapcore.NewJSONEncoder(pachdEncoder)
	w := new(count)
	l := makeLogger(enc, zapcore.Lock(w), zapcore.DebugLevel, sample, []zap.Option{zap.AddCaller()})
	return withLogger(context.Background(), l), w
}

func BenchmarkFields(b *testing.B) {
	ctx, w := newBenchLogger(false)
	for i := 0; i < b.N; i++ {
		Debug(ctx, "debug", zap.Int("i", i))
	}
	if w.n == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkFieldsSampled(b *testing.B) {
	ctx, w := newBenchLogger(true)
	for i := 0; i < b.N; i++ {
		Debug(ctx, "debug", zap.Int("i", i))
	}
	if w.n == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkSpan(b *testing.B) {
	ctx, w := newBenchLogger(false)
	errEven := errors.New("even")
	for i := 0; i < b.N; i++ {
		func() (retErr error) {
			defer Span(ctx, "bench", zap.Int("i", i))()
			if i%2 == 0 {
				return errEven //nolint:wrapcheck
			}
			return nil
		}() //nolint:errcheck
	}
	if w.n == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkSpanWithError(b *testing.B) {
	ctx, w := newBenchLogger(false)
	errEven := errors.New("even")
	for i := 0; i < b.N; i++ {
		func() (retErr error) {
			defer Span(ctx, "bench", zap.Int("i", i))(Errorp(&retErr))
			if i%2 == 0 {
				return errEven //nolint:wrapcheck
			}
			return nil
		}() //nolint:errcheck
	}
	if w.n == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkLogrusFields(b *testing.B) {
	l := logrus.New()
	w := new(count)
	l.Formatter = &logrus.JSONFormatter{}
	l.Out = w
	l.Level = logrus.DebugLevel
	for i := 0; i < b.N; i++ {
		l.WithField("i", i).Debug("debug")
	}
	if w.n == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkLogrusSugar(b *testing.B) {
	l := logrus.New()
	w := new(count)
	l.Formatter = &logrus.JSONFormatter{}
	l.Out = w
	l.Level = logrus.DebugLevel
	for i := 0; i < b.N; i++ {
		l.Debugf("debug: %d", i)
	}
	if w.n == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkLogrusWrapper(b *testing.B) {
	ctx, w := newBenchLogger(false)
	l := NewLogrus(ctx)
	for i := 0; i < b.N; i++ {
		l.Debugf("debug: %d", i)
	}
	if w.n == 0 {
		b.Fatal("no bytes added to logger")
	}
}
