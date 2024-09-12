package log

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestSkip(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "m",
		CallerKey:      "c",
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	})
	l := zap.New(zapcore.NewCore(enc, zapcore.AddSync(buf), zap.DebugLevel), zap.AddCaller())
	ctx := withLogger(context.Background(), l.WithOptions(zap.AddCallerSkip(1)))

	// It is important that the next line is line 26.
	if err := LogStep(ctx, "log a step", func(context.Context) error { return errors.New("oh no") }); err == nil {
		t.Errorf("expected error")
	}

	if got, want := strings.Count(buf.String(), `"c":"log/step_test.go:26"`), 2; got != want {
		t.Errorf("caller occurence count:\n  got: %v\n want: %v", got, want)
		t.Logf("log buffer:\n%s", buf.String())
	}
}
