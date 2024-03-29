package testloki

import (
	"context"
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
		MutateContext: func(ctx context.Context) context.Context {
			return pctx.Child(ctx, "", pctx.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
				return zapcore.NewTee(c, l.NewZapCore(ctx))
			})))
		},
	}
}

// NewZapCore returns a zapcore.Core that sends logs to this Loki instance.
func (l *TestLoki) NewZapCore(ctx context.Context) zapcore.Core {
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
		podTemplateHash: randutil.UniqueString("")[0:10],
	}
}

type lokiCore struct {
	ctx             context.Context
	l               *TestLoki
	enc             zapcore.Encoder
	podTemplateHash string
	fields          []zapcore.Field
}

var _ zapcore.Core = (*lokiCore)(nil)

// Enabled implements zapcore.LevelEnabler.
func (c *lokiCore) Enabled(zapcore.Level) bool {
	return true
}

// With implements zapcore.Core.
func (c *lokiCore) With(f []zapcore.Field) zapcore.Core {
	var fields []zapcore.Field
	fields = append(fields, c.fields...)
	fields = append(fields, f...)
	return &lokiCore{ctx: c.ctx, l: c.l, enc: c.enc, podTemplateHash: c.podTemplateHash, fields: fields}
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
	buf, err := c.enc.EncodeEntry(e, fields)
	if err != nil {
		return errors.Wrap(err, "encode log entry")
	}
	if err := c.l.AddLog(c.ctx, &Log{
		Time: time.Now(),
		Labels: map[string]string{
			"host":              "localhost",
			"app":               "pachd",
			"container":         "pachd",
			"node_name":         "localhost",
			"pod":               "pachd-" + c.podTemplateHash + c.podTemplateHash[0:5],
			"pod_template_hash": c.podTemplateHash,
			"stream":            "stderr",
			"suite":             "pachyderm",
		},
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
