package log

import (
	"context"
	"fmt"
	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapgrpc"
	"google.golang.org/grpc/grpclog"
	"k8s.io/klog/v2"
)

// LevelChanger is a log level that can be changed for a period of time.
type LevelChanger interface {
	// See resettableLevel.SetLevelFor.
	SetLevelFor(zapcore.Level, time.Duration, func(string, string))
}

var (
	logLevel  = NewResettableLevelAt(zapcore.InfoLevel)  // Current base logger level.
	grpcLevel = NewResettableLevelAt(zapcore.FatalLevel) // The log level for the GRPC adaptor.

	// These log levels are for the src/server/debug package, which changes log levels at
	// runtime based on an RPC.  The Debug package takes special care to not change LogLevel
	// level to one we don't use; zapcore.Level has more levels than Pachyderm.
	LogLevel, GRPCLevel LevelChanger = logLevel, grpcLevel

	healthCheckLogger *zap.Logger // A logger only for GRPC health checks.

	developmentLogger bool // True if a development logger was requested via the environment.
	samplingDisabled  bool // True if log sampling was disabled via the environment.
	pachctlJSON       bool // True if pachctl should log JSON logs.

	initOnce       sync.Once   // initOnce gates creating the zap global logger
	warningsLogged atomic.Bool // True if startup warnings have already been printed.
	warnings       []string    // Any warnings generated during startup (before the logger was ready).

	// droppedLogs and droppedHealthLogs record how many logs or health check logs we dropped
	// because of log sampling.
	droppedLogs, droppedHealthLogs atomic.Uint64
)

const (
	// EnvLogLevel is the name of the log level environment variable.  It's read by worker_rc.go
	// to propagate our value to any workers the PPS master creates.
	EnvLogLevel = "LOG_LEVEL"
)

// SetLevel changes the global logger level.  It is safe to call at any time from multiple
// goroutines.
func SetLevel(l Level) {
	logLevel.SetLevel(l.coreLevel())
}

// SetGRPCLogLevel changes the grpc logger level.  It is safe to call at any time from multiple
// goroutines.
//
// Note: to see any messages, the overall log level has to be lower than the grpc log level.  (For
// example, if you SetLevel to ERROR and then SetGRPCLogLevel to DEBUG, you will only see ERROR logs
// from GRPC.  But, if you SetLogLevel(DEBUG) and SetGRPCLogLevel(DEBUG), you will see DEBUG grpc
// logs.)
//
// Note: SetGRPCLogLevel takes a zapcore.Level instead of a internal/log.Level.  That's because GRPC
// has more log levels than Pachyderm.
func SetGRPCLogLevel(l zapcore.Level) {
	grpcLevel.SetLevel(l)
}

// addInitWarningf logs a warning at logger initialization time.  The intent is to be able to log
// warnings about logging configuration before the logger exists.
func addInitWarningf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if warningsLogged.Load() {
		zap.L().DPanic(msg, zap.Stack("stack"))
		return
	}
	warnings = append(warnings, msg)
}

func init() {
	// Note: the service enviroment also has a LOG_LEVEL field. It's not used here (servieenv
	// initialization is too late), but it is used to copy the log level to the workers in
	// worker_rc.go.
	if lvl := os.Getenv(EnvLogLevel); lvl != "" {
		if err := logLevel.UnmarshalText([]byte(lvl)); err != nil {
			addInitWarningf("parse $LOG_LEVEL: %v; proceeding at %v level", err.Error(), logLevel.Level().String())
		}
	}
	if d := os.Getenv("DEVELOPMENT_LOGGER"); d != "" {
		if d == "true" || d == "1" {
			developmentLogger = true
		} else {
			addInitWarningf("$DEVELOPMENT_LOGGER set but unparsable; got %q, want 'true' or '1'", d)
		}
	}
	if s := os.Getenv("DISABLE_LOG_SAMPLING"); s != "" {
		if s == "true" || s == "1" {
			samplingDisabled = true
		} else {
			addInitWarningf("$DISABLE_LOG_SAMPLING set but unparsable; got %q, want 'true' or '1'", s)
		}
	}
	if j := os.Getenv("PACHCTL_JSON_LOGS"); j != "" {
		if j == "true" || j == "1" {
			pachctlJSON = true
		} else {
			addInitWarningf("$PACHCTL_JSON_LOGS is set but unparsable; got %q, want 'true' or '1'", j)
		}
	}
}

// InitPachdLogger creates a new global zap logger suitable for use in pachd/enterprise server/etc.
func InitPachdLogger() {
	enc := zapcore.NewJSONEncoder(pachdEncoder)
	makeLoggerOnce(enc, os.Stderr, true, []zap.Option{zap.AddCaller()})
}

// InitWorkerLogger creates a new global zap logger suitable for use in the worker.
func InitWorkerLogger() {
	enc := zapcore.NewJSONEncoder(workerEncoder)
	makeLoggerOnce(enc, os.Stdout, true, []zap.Option{zap.AddCaller()})
}

// InitPachctlLogger creates a new logger for command-line tools like pachctl.
func InitPachctlLogger() {
	if pachctlJSON {
		InitPachdLogger()
		return
	}
	cfg := pachctlEncoder
	if !color.NoColor { // Enable color if it's not disabled via flags.
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	enc := zapcore.NewConsoleEncoder(cfg)
	makeLoggerOnce(enc, os.Stderr, false, []zap.Option{zap.AddCaller()})
}

// makeLoggerOnce sets up the global logger once.
func makeLoggerOnce(enc zapcore.Encoder, w zapcore.WriteSyncer, sample bool, opts []zap.Option) {
	if warningsLogged.Load() {
		zap.L().DPanic("logger already initialized", zap.Stack("stack"))
		// This will then return the already-created logger in production, and panic in dev.
	}
	initOnce.Do(func() {
		defer func() {
			warningsLogged.Store(true)
			warnings = nil
		}()
		w = zapcore.Lock(w)
		if developmentLogger {
			opts = append(opts, zap.Development())
		}

		// Global logger
		sample = sample && !samplingDisabled
		l := makeLogger(enc, w, logLevel, sample, opts)
		zap.ReplaceGlobals(l)
		zap.RedirectStdLog(l)

		// Health check logger
		healthCheckLogger = l.WithOptions(zap.WithCaller(false), zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return newSamplerWithOptions(
				c, true, time.Hour, 1, math.MaxInt,
				samplerHook(func(_ zapcore.Entry, dec zapcore.SamplingDecision) {
					if dec&zapcore.LogDropped > 0 {
						droppedHealthLogs.Add(1)
					}
				}))
		}))

		// Copy gRPC logs to the base logger, at the level controlled by
		// grpcLevel. (Changeable at runtime.)
		grpclog.SetLoggerV2(zapgrpc.NewLogger(l.Named("grpclog").WithOptions(zap.IncreaseLevel(grpcLevel))))

		// Copy client-go logs to the base logger.
		klog.SetLogger(zapr.NewLogger(l.Named("client-go")))

		// Show warnings.
		if len(warnings) > 0 {
			l.DPanic("warnings unexpectedly generated before logger initalization", zap.Strings("warnings", warnings), zap.Stack("stack"))
		}
	})
}

// makeLogger actually builds a global logger, but doesn't make it global.  It is safe to use in
// tests or benchmarks.
func makeLogger(enc zapcore.Encoder, w zapcore.WriteSyncer, lvl zapcore.LevelEnabler, sample bool, opts []zap.Option) *zap.Logger {
	core := zapcore.NewCore(enc, w, lvl)
	core = newSamplerWithOptions(
		core, sample, 3*time.Second, 2, 100,
		samplerHook(func(_ zapcore.Entry, dec zapcore.SamplingDecision) {
			if dec&zapcore.LogDropped > 0 {
				droppedLogs.Add(1)
			}
		}))
	return zap.New(core, opts...)
}

// reportDroppedLogs prints a metric about how many logs have been dropped, if any have been
// dropped.
func reportDroppedLogs(ctx context.Context) {
	if n := droppedLogs.Swap(0); n > 0 {
		Debug(ctx, "dropped logs", zap.Uint64("n", n))
	}
	if n := droppedHealthLogs.Swap(0); n > 0 {
		Debug(ctx, "dropped health check logs", zap.Uint64("n", n))
	}
}

// WatchDroppedLogs periodically prints a report on how many logs were dropped.
func WatchDroppedLogs(ctx context.Context, d time.Duration) {
	t := time.NewTicker(d)
	for {
		select {
		case <-t.C:
			reportDroppedLogs(ctx)
		case <-ctx.Done():
			Info(ctx, "dropped log reporting ended", zap.Error(ctx.Err()))
			t.Stop()
			return
		}
	}
}

// HealthCheckLogger returns a logger to log health checks.  If the gRPC interceptor moves to this
// package, it shouldn't need to be public anymore.
func HealthCheckLogger(ctx context.Context) context.Context {
	if healthCheckLogger == nil {
		zap.L().WithOptions(zap.AddCallerSkip(1)).DPanic("log: internal error: health check logger not yet initialized", zap.Stack("stack"))
	}
	return withLogger(ctx, healthCheckLogger)
}
