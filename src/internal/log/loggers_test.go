package log

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestInit(t *testing.T) {
	buf := new(bytes.Buffer)

	// Create a development logger that logs to buf.
	developmentLogger = true
	initOnce = sync.Once{}
	warnings = nil
	warningsLogged.Store(false)

	makeLoggerOnce(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{MessageKey: "m"}),
		zapcore.AddSync(buf),
		true,
		nil,
	)
	t.Cleanup(func() {
		zap.ReplaceGlobals(zap.NewNop())
	})

	if got, want := buf.String(), ""; got != want {
		t.Errorf("unexpected log messages: %s", got)
	}
	buf.Reset()

	// Ensure we panic when creating a duplicate global logger.
	var panic string
	func() {
		defer func() {
			err := recover()
			if err != nil {
				panic = err.(string)
			}
		}()
		makeLoggerOnce(
			zapcore.NewJSONEncoder(zapcore.EncoderConfig{MessageKey: "m"}),
			zapcore.AddSync(io.Discard),
			false,
			nil,
		)
	}()
	if got, want := panic, "logger already initialized"; got != want {
		t.Errorf("init twice: expected panic:\n  got: %v\n want: %v", got, want)
	}

	// Ensure the first logger is still in use.
	buf.Reset()
	zap.L().Info("this is a test")
	if got := buf.String(); !strings.Contains(got, `"this is a test"`) {
		t.Errorf("log should contain `\"test\"`: %s", got)
	}

	// Check that rate limiting is doing something and can be turned off.
	testData := []struct {
		name        string
		makeLogger  func() *zap.Logger
		wantDropped bool
	}{
		{
			name:        "global",
			makeLogger:  zap.L,
			wantDropped: true,
		},
		{
			name:        "non-rate limited child",
			makeLogger:  func() *zap.Logger { return WithoutRatelimit()(zap.L()) },
			wantDropped: false,
		},
		{
			name:        "global after child",
			makeLogger:  zap.L,
			wantDropped: true,
		},
	}
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			l := test.makeLogger()
			for i := 0; i < 10000; i++ {
				l.Info("rate limit")
			}
			if err := zap.L().Sync(); err != nil {
				t.Fatalf("sync: %v", err)
			}

			d := droppedLogs.Swap(0)
			if (d == 0) == test.wantDropped {
				t.Errorf("rate limiting appears to not be working:\n wantDropped: %v\n     dropped: %v", test.wantDropped, d)
			}
			if got, want := strings.Count(buf.String(), "\n"), 10000-int(d); got != want {
				t.Errorf("log lines - dropped logs:\n  got: %v\n want: %v", got, want)
			}
		})
	}

	// Check that the health check logger also rate limits.
	hl := healthCheckLogger
	SetGRPCLogLevel(zapcore.InfoLevel)
	buf.Reset()
	for i := 0; i < 10000; i++ {
		hl.Info("such health, wow!")
	}
	if err := hl.Sync(); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Count(buf.String(), "\n"), 1; got != want {
		t.Errorf("health check lines:\n  got: %v\n want: %v", got, want)
	}
	if got, want := int(droppedHealthLogs.Swap(0)), 9999; got != want {
		t.Errorf("health check dropped logs:\n  got: %v\n want: %v", got, want)
	}
}

func TestWarnings(t *testing.T) {
	buf := new(bytes.Buffer)
	developmentLogger = false // don't panic on the warning
	initOnce = sync.Once{}
	warnings = nil
	warningsLogged.Store(false)
	addInitWarningf("this is an init warning")
	makeLoggerOnce(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{MessageKey: "m"}),
		zapcore.AddSync(buf),
		false,
		nil,
	)
	t.Cleanup(func() {
		zap.ReplaceGlobals(zap.NewNop())
	})
	if got, want := buf.String(), `"this is an init warning"`; !strings.Contains(got, want) {
		t.Errorf("init:\n  got: %v\n want: /%v/", got, want)
	}
}

func TestReportDroppedLogs(t *testing.T) {
	ctx, h := testWithCaptureParallel(t)

	Debug(ctx, "1: no dropped logs")
	droppedLogs.Store(0)
	reportDroppedLogs(ctx)

	Debug(ctx, "2: some dropped logs")
	droppedLogs.Store(10)
	reportDroppedLogs(ctx)

	Debug(ctx, "3: no dropped logs")
	reportDroppedLogs(ctx)

	want := []string{
		"debug: 1: no dropped logs",

		"debug: 2: some dropped logs",
		"debug: dropped logs",

		"debug: 3: no dropped logs",
	}
	if diff := cmp.Diff(h.Logs(), want, formatLogs(simple)); diff != "" {
		t.Errorf("logs (-got +want):\n%s", diff)
	}
}
