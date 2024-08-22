package log

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

func TestLogrus(t *testing.T) {
	// The next nolint is for "h is never used" because of the Fatal call.  Since we have
	// overridden ExitFunc, we do actually use h.
	ctx, h := testWithCaptureParallel(t, zap.Development()) //nolint:SA1019
	lr := NewLogrus(ctx)
	lr.SetLevel(logrus.TraceLevel)
	lr.ExitFunc = func(code int) {}

	lr.Trace("trace")
	lr.Debug("debug")
	lr.Info("info")
	lr.Infof("infof %v %q", "foo", "bar")
	lr.Warn("warn")
	lr.Error("error")
	func() {
		defer func() {
			recover() //nolint:errcheck
		}()
		lr.Panic("panic")
	}()
	lr.Fatal("fatal")

	sctx, end := SpanContext(ctx, "myspan")
	lr.WithContext(sctx).Debug("ctx debug")
	lr.WithContext(sctx).WithError(errors.New("hi")).WithField("foo", "bar").Debug("tagged ctx debug")
	lr.WithContext(sctx).Info("ctx info")
	lr.WithContext(sctx).Error("ctx error")
	end()

	lr.WithField("foo", "bar").Debug("tagged debug")

	want := []string{
		"debug: trace", "debug: debug",
		"info: info", `info: infof foo "bar"`, "info: warn",
		"error: error", "error: panic", "error: fatal",
		"myspan: debug: myspan: span start",
		"myspan: debug: ctx debug", "myspan: debug: tagged ctx debug", "myspan: info: ctx info", "myspan: error: ctx error",
		"myspan: debug: myspan: span finished ok",
		"debug: tagged debug",
	}
	if diff := cmp.Diff(h.Logs(), want, formatLogs(simple)); diff != "" {
		t.Errorf("logs (-got +want):\n%s", diff)
	}
	var dump bool
	for i, l := range h.Logs() {
		if !strings.HasPrefix(l.Caller, "log/logrus_test.go:") {
			t.Errorf("line %d: caller:\n  got:  %v\n want: ^log/logrus_test.go:", i+1, l.Caller)
			dump = true
		}
	}
	if dump {
		for i, l := range h.Logs() {
			t.Logf("line %d: %s", i+1, l.Orig)
		}
	}
}
