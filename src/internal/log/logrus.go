package log

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// dest is a logging destination.
type dest struct {
	ctx context.Context
}

var _ logrus.Hook = dest{}

// Fire implements logrus.Hook.
func (d dest) Fire(e *logrus.Entry) error {
	var fields []zap.Field
	for key, value := range e.Data {
		if key == logrus.ErrorKey {
			if err, ok := value.(error); ok {
				fields = append(fields, zap.Error(err))
			} else {
				fields = append(fields, zap.Any(key, value))
			}
			continue
		}
		fields = append(fields, zap.Any(key, value))
	}
	ctx := d.ctx
	if e.Context != nil {
		ctx = e.Context
	}
	if c := e.Caller; c != nil {
		fields = append(fields, zap.String("caller", zapcore.EntryCaller{
			PC:       c.PC,
			File:     c.File,
			Defined:  true,
			Line:     c.Line,
			Function: c.Function,
		}.TrimmedPath()))
		ctx = withLogger(ctx, extractLogger(ctx).WithOptions(zap.WithCaller(false)))
	}
	switch e.Level { //exhaustive:enforce
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
		Error(ctx, e.Message, fields...)
	case logrus.WarnLevel, logrus.InfoLevel:
		Info(ctx, e.Message, fields...)
	case logrus.DebugLevel, logrus.TraceLevel:
		Debug(ctx, e.Message, fields...)
	}

	return nil
}

// Levels implements logrus.Hook.
func (d dest) Levels() []logrus.Level { return logrus.AllLevels }

// NewLogrus returns a *logrus.Logger that logs to the provided context (or a context passed with
// each logged entry, if available).
func NewLogrus(ctx context.Context) *logrus.Logger {
	l := logrus.New()
	l.Out = io.Discard
	l.Level = logrus.TraceLevel
	l.ReportCaller = true
	l.AddHook(dest{ctx: ctx})
	return l
}
