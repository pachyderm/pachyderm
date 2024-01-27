package log

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/adrg/xdg"
	"github.com/fatih/color"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapgrpc"
	"google.golang.org/grpc/grpclog"
	v1 "k8s.io/api/core/v1"
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
	EnvLogLevel           = "PACHYDERM_LOG_LEVEL"
	EnvDevelopmentLogger  = "PACHYDERM_DEVELOPMENT_LOGGER"
	EnvDisableLogSampling = "PACHYDERM_DISABLE_LOG_SAMPLING"
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

// StartupLogConfig is the logging configuration at startup time.
type StartupLogConfig struct {
	LogLevel           zapcore.Level
	DevelopmentLogger  bool
	DisableLogSampling bool
}

// WorkerLogConfig is the configuration that worker_rc.go reads to propagate the startup state of
// the logging subsystem to new workers.  Some of this state can change as pachd runs (the log
// level), so we capture a view at app startup and only propagate that view to the workers.
var WorkerLogConfig = StartupLogConfig{}

// AsKubernetesEnvironment returns environment variables that should be set to propagate the logging
// config.
func (c StartupLogConfig) AsKubernetesEnvironment() []v1.EnvVar {
	result := []v1.EnvVar{
		{
			Name:  EnvLogLevel,
			Value: c.LogLevel.String(),
		},
	}
	if c.DevelopmentLogger {
		result = append(result, v1.EnvVar{
			Name:  EnvDevelopmentLogger,
			Value: "1",
		})
	}
	if c.DisableLogSampling {
		result = append(result, v1.EnvVar{
			Name:  EnvDisableLogSampling,
			Value: "1",
		})
	}
	return result
}

func init() {
	if lvl := os.Getenv(EnvLogLevel); lvl != "" {
		if err := logLevel.UnmarshalText([]byte(lvl)); err != nil {
			addInitWarningf("parse $%s: %v; proceeding at %v level", EnvLogLevel, err, logLevel.Level().String())
		}
	} else if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		if err := logLevel.UnmarshalText([]byte(lvl)); err != nil {
			addInitWarningf("parse $LOG_LEVEL: %v; proceeding at %v level", err, logLevel.Level().String())
		}
		addInitWarningf("$LOG_LEVEL has been renamed to $PACHYDERM_LOG_LEVEL; please set pachd.logLevel in the helm chart rather than passing in LOG_LEVEL as a patch")
	}
	WorkerLogConfig.LogLevel = logLevel.Level()

	if d := os.Getenv(EnvDevelopmentLogger); d != "" {
		if d == "true" || d == "1" {
			developmentLogger = true
			WorkerLogConfig.DevelopmentLogger = true
		} else {
			addInitWarningf("$%s set but unparsable; got %q, want 'true' or '1'", EnvDevelopmentLogger, d)
		}
	}

	if s := os.Getenv(EnvDisableLogSampling); s != "" {
		if s == "true" || s == "1" {
			samplingDisabled = true
			WorkerLogConfig.DisableLogSampling = true
		} else {
			addInitWarningf("$%s set but unparsable; got %q, want 'true' or '1'", EnvDisableLogSampling, s)
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

// InitBatchLogger creates a new logger for command-line tools that need to retain their logs on
// error.  If the returned callback is called with no error, then the log file is deleted.  If it's
// called with an error, the path to the log is printed, the log file is retained, and log.Exit is
// called to log the error (and ensure everything is flushed).  If logFile is non-empty, then the
// log file is always retained.
func InitBatchLogger(logFile string) func(err error) {
	cfg := minimalConsoleEncoder
	if !color.NoColor { // Enable color if it's not disabled via flags.
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	enc := zapcore.NewConsoleEncoder(cfg)
	keepLog := logFile != ""
	close := func() {}
	makeLoggerOnce(enc, os.Stderr, false, []zap.Option{zap.AddCaller(), zap.WrapCore(
		func(c zapcore.Core) zapcore.Core {
			name := "batch"
			if len(os.Args) > 0 {
				name = filepath.Base(os.Args[0])
			}
			if logFile == "" {
				var err error
				// On error, out is set to "" again.
				name := fmt.Sprintf("%s.%s.log", name, time.Now().In(time.UTC).Format("20060102T150405Z"))
				logFile, err = xdg.CacheFile(filepath.Join("pachyderm/log", name))
				if err != nil {
					// When xdg.CacheFile doesn't work, use $PWD.  This works
					// fine for, say, Bazel genrules.
					logFile = name
				}
			} else {
				if stat, err := os.Stat(logFile); err == nil {
					// Any errors here will be handled by Open.
					if stat.IsDir() {
						addInitWarningf("log file %v is a directory; not logging to file", logFile)
						logFile = ""
						keepLog = false
						return c
					}
				}
			}
			ws, closeLog, err := zap.Open(logFile)
			if err != nil {
				addInitWarningf("problem opening log file %v: %v", logFile, err)
				logFile = ""
				keepLog = false
				return c
			}
			close = closeLog
			enc := zapcore.NewJSONEncoder(pachdEncoder)
			fileCore := zapcore.NewCore(enc, ws, zap.DebugLevel)
			return zapcore.NewTee(c, fileCore)
		},
	)})
	if logFile != "" {
		zap.S().Debugf("logging to file %v", logFile)
	}
	return func(err error) {
		if err == nil {
			if keepLog {
				zap.S().Infof("logfile retained at %v", logFile)
			}
			close()
			if !keepLog && logFile != "" {
				if err := os.Remove(logFile); err != nil {
					// The logger is gone at this point, so... we can't log the error.
					fmt.Fprintf(os.Stderr, "unable to delete unwanted logfile at %v: %v", logFile, err)
				}
			}
			return
		}
		// If we're exiting because of a context being canceled, give those goroutines a
		// little time to do IO and notice that the context is dead.  That way, their final
		// errors go to the log file and don't print an error about trying to write to the
		// closed log file.  (Needed 800us in testing.)
		time.Sleep(5 * time.Millisecond)
		if logFile != "" {
			zap.L().Info(fmt.Sprintf("logfile retained at %v", logFile))
		}
		close()
	}
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
			Info(ctx, "dropped log reporting ended", zap.Error(context.Cause(ctx)))
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
