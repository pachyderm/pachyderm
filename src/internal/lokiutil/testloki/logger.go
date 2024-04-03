package testloki

import (
	"context"
	"fmt"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func WithTestLoki(l *TestLoki) pachd.TestPachdOption {
	return pachd.TestPachdOption{
		MutateEnv: func(env *pachd.Env) {
			env.GetLokiClient = func() (*client.Client, error) {
				return l.Client, nil
			}
		},
		MutateConfig: func(config *pachconfig.PachdFullConfiguration) {
			config.LokiLogging = true
		},
	}
}

// newZapCore returns a zapcore.Core that sends logs to this Loki instance.
func (l *TestLoki) newZapCore(ctx context.Context) *lokiCore {
	return &lokiCore{
		ctx: ctx,
		l:   l,
		enc: zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "time",
			EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
			LevelKey:       "severity",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			MessageKey:     "message",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}),
		labels: map[string]string{"suite": "pachyderm"},
	}
}

// WithLoki returns a pctx context option that will send logs to this Loki instance.
func (l *TestLoki) WithLoki(sendCtx context.Context, lokiLabels map[string]string) pctx.Option {
	lc := l.newZapCore(sendCtx)
	lc.labels = lokiLabels
	return pctx.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, lc)
	}))
}

func WithLoki(ctx context.Context, l *TestLoki) context.Context {
	lc := l.newZapCore(ctx)
	templateHash := randutil.UniqueString("")[0:10]
	procHash := randutil.UniqueString("")[0:5]
	lc.labels = map[string]string{
		"host":              "localhost",
		"app":               "pachd",
		"container":         "pachd",
		"node_name":         "localhost",
		"pod":               fmt.Sprintf("pachd-%v-%v", templateHash, procHash),
		"pod_template_hash": templateHash,
		"stream":            "stderr",
		"suite":             "pachyderm",
	}
	return pctx.Child(ctx, "", pctx.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, lc)
	})))
}

// lokiCore sends JSON log messages to the provided TestLoki instance.  If you were going to use
// this code in production, you would want to skip the JSON serialization step, buffer lines to send
// them in batches, and read the time from the Entry instead of using time.Now().  (We use
// time.Now() here because it's more like what promtail does; reads the line and adds its own
// timestamp.)
type lokiCore struct {
	ctx    context.Context
	l      *TestLoki
	enc    zapcore.Encoder
	labels map[string]string
}

var _ zapcore.Core = (*lokiCore)(nil)

// Enabled implements zapcore.LevelEnabler.
func (c *lokiCore) Enabled(zapcore.Level) bool {
	return true
}

// With implements zapcore.Core.
func (c *lokiCore) With(fields []zapcore.Field) zapcore.Core {
	enc := c.enc.Clone()
	for _, f := range fields {
		f.AddTo(enc)
	}
	return &lokiCore{ctx: c.ctx, l: c.l, enc: enc, labels: c.labels}
}

// Check implements zapcore.Core.
func (c *lokiCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(e.Level) {
		return ce.AddCore(e, c)
	}
	return ce
}

// Write implements zapcore.Core.
func (c *lokiCore) Write(e zapcore.Entry, fields []zapcore.Field) error {
	select {
	case <-c.ctx.Done():
		// For tests, we do not care about writes failing beacuse the root context is done.
		return nil
	default:
	}
	buf, err := c.enc.EncodeEntry(e, fields)
	if err != nil {
		return errors.Wrap(err, "encode log entry")
	}
	if err := c.l.AddLog(c.ctx, &Log{
		Time:    time.Now(),
		Labels:  c.labels,
		Message: string(buf.Bytes()),
	}); err != nil {
		return errors.Wrap(err, "send log to loki")
	}
	return nil
}

// Sync implements zapcore.Core.
func (c *lokiCore) Sync() error {
	return nil
}
