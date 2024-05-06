package log

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestKind(t *testing.T) {
	lvl := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	ctx := withLogger(context.Background(), zaptest.NewLogger(t, zaptest.WrapOptions(zap.AddCaller(), zap.AddCallerSkip(1)), zaptest.Level(lvl)))
	k := NewKindLogger(ctx)

	k.Warn("warn")
	k.Warnf("%s%s", "warn", "f")
	k.Error("error")
	k.Errorf("%s%s", "error", "f")

	if got, want := true, k.V(0).Enabled(); got != want {
		t.Errorf("V(0).Enabled:\n  got: %v\n want: %v", got, want)
	}
	if got, want := false, k.V(9).Enabled(); got != want {
		t.Errorf("V(9).Enabled:\n  got: %v\n want: %v", got, want)
	}

	kv := k.V(0)
	kv.Info("info")
	kv.Infof("%s%s", "info", "f")

	kv = k.V(9)
	kv.Info("not seen")
	kv.Infof("%s %s", "not", "seen")

	kv = k.V(1)
	lvl.SetLevel(zapcore.DebugLevel)
	kv.Info("debug")
	kv.Infof("%s%s", "debug", "f")
	if got, want := true, kv.Enabled(); got != want {
		t.Errorf("V(1).Enabled:\n  got: %v\n want: %v", got, want)
	}
}
