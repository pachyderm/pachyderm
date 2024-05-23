package log

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	kindlog "sigs.k8s.io/kind/pkg/log"
)

type kindLogger struct {
	v   kindlog.Level
	ctx context.Context
}

var _ kindlog.Logger = (*kindLogger)(nil)
var _ kindlog.InfoLogger = (*kindLogger)(nil)

func chomp(x string) string {
	return strings.TrimRight(x, "\n")
}

func (l *kindLogger) Warn(message string) {
	Error(l.ctx, chomp(message))
}

func (l *kindLogger) Warnf(format string, args ...any) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *kindLogger) Error(message string) {
	Error(l.ctx, chomp(message))
}

func (l *kindLogger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *kindLogger) V(v kindlog.Level) kindlog.InfoLogger {
	return &kindLogger{ctx: l.ctx, v: v}
}

func (l *kindLogger) Info(message string) {
	if l.v > 0 {
		Debug(l.ctx, chomp(message))
	} else {
		Info(l.ctx, chomp(message))
	}
}

func (l *kindLogger) Infof(format string, args ...any) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *kindLogger) Enabled() bool {
	if l.v > 0 {
		if logger := l.ctx.Value(pachydermLogger{}); logger != nil {
			return logger.(*zap.Logger).Core().Enabled(zapcore.DebugLevel)
		}
		return false
	}
	return true
}

func NewKindLogger(ctx context.Context) kindlog.Logger {
	return &kindLogger{ctx: withLogger(ctx, extractLogger(ctx).WithOptions(zap.WithCaller(false)))}
}
