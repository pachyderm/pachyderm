package log

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/camelcase"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// This needs to be a global var, not a field on the logger, because multiple servers
// create new loggers, and the prometheus registration uses a global namespace
var reportMetricGauge prometheus.Gauge
var reportMetricsOnce sync.Once

// Logger is a helper for emitting our grpc API logs
type Logger interface {
	Log(ctx context.Context, request interface{}, response interface{}, err error, duration time.Duration)
	LogAtLevelFromDepth(ctx context.Context, request interface{}, response interface{}, err error, duration time.Duration, level logrus.Level, depth int)
}

type logger struct {
	*logrus.Entry
	histogram   map[string]*prometheus.HistogramVec
	counter     map[string]prometheus.Counter
	mutex       *sync.Mutex // synchronizes access to both histogram and counter maps
	exportStats bool
	service     string
}

// NewLogger creates a new logger
func NewLogger(service string, l *logrus.Logger) Logger {
	return newLogger(service, true, l)
}

// NewLocalLogger creates a new logger for local testing (which does not report prometheus metrics)
func NewLocalLogger(service string, l *logrus.Logger) Logger {
	return newLogger(service, false, l)
}

func newLogger(service string, exportStats bool, l *logrus.Logger) Logger {
	l.Formatter = FormatterFunc(Pretty)
	newLogger := &logger{
		l.WithFields(logrus.Fields{"service": service}),
		make(map[string]*prometheus.HistogramVec),
		make(map[string]prometheus.Counter),
		&sync.Mutex{},
		exportStats,
		service,
	}
	if exportStats {
		reportMetricsOnce.Do(func() {
			newReportMetricGauge := prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace: "pachyderm",
					Subsystem: "pachd",
					Name:      "report_metric",
					Help:      "gauge of number of calls to ReportMetric()",
				},
			)
			if err := prometheus.Register(newReportMetricGauge); err != nil {
				// metrics may be redundantly registered; ignore these errors
				if !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
					entry := newLogger.WithFields(logrus.Fields{"method": "NewLogger"})
					newLogger.LogAtLevel(entry, logrus.WarnLevel, fmt.Sprintf("error registering prometheus metric: %v", newReportMetricGauge), err)
				}
			} else {
				reportMetricGauge = newReportMetricGauge
			}
		})
	}
	return newLogger
}

// Helper function used to log requests and responses from our GRPC method
// implementations
func (l *logger) Log(ctx context.Context, request interface{}, response interface{}, err error, duration time.Duration) {
	if err != nil {
		l.LogAtLevelFromDepth(ctx, request, response, err, duration, logrus.ErrorLevel, 4)
	} else {
		l.LogAtLevelFromDepth(ctx, request, response, err, duration, logrus.InfoLevel, 4)
	}
	// We have to grab the method's name here before we
	// enter the goro's stack
	go l.ReportMetric(getMethodName(), duration, err)
}

func getMethodName() string {
	depth := 4
	pc := make([]uintptr, depth)
	runtime.Callers(depth, pc)
	split := strings.Split(runtime.FuncForPC(pc[0]).Name(), ".")
	return split[len(split)-1]
}

func (l *logger) ReportMetric(method string, duration time.Duration, err error) {
	if !l.exportStats {
		return
	}
	// Count the number of ReportMetric() goros in case we start to leak them
	if reportMetricGauge != nil {
		reportMetricGauge.Inc()
	}
	defer func() {
		if reportMetricGauge != nil {
			reportMetricGauge.Dec()
		}
	}()
	l.mutex.Lock() // for conccurent map access (histogram,counter)
	defer l.mutex.Unlock()
	state := "started"
	if err != nil {
		state = "errored"
	} else {
		if duration.Seconds() > 0 {
			state = "finished"
		}
	}
	entry := l.WithFields(logrus.Fields{"method": method})

	var tokens []string
	for _, token := range camelcase.Split(method) {
		tokens = append(tokens, strings.ToLower(token))
	}
	rootStatName := strings.Join(tokens, "_")

	// Recording the distribution of started times is meaningless
	if state != "started" {
		runTimeName := fmt.Sprintf("%v_seconds", rootStatName)
		runTime, ok := l.histogram[runTimeName]
		if !ok {
			runTime = prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace: "pachyderm",
					Subsystem: fmt.Sprintf("pachd_%v", topLevelService(l.service)),
					Name:      runTimeName,
					Help:      fmt.Sprintf("Run time of %v", method),
					Buckets:   []float64{0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300, 600, 1800, 3600, 86400},
				},
				[]string{
					"state", // Since both finished and errored API calls can have run times
				},
			)
			if err := prometheus.Register(runTime); err != nil {
				// metrics may be redundantly registered; ignore these errors
				if !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
					l.LogAtLevel(entry, logrus.WarnLevel, fmt.Sprintf("error registering prometheus metric %v: %v", runTimeName, err))
				}
			} else {
				l.histogram[runTimeName] = runTime
			}
		}
		if hist, err := runTime.GetMetricWithLabelValues(state); err != nil {
			l.LogAtLevel(entry, logrus.WarnLevel, fmt.Sprintf("failed to get histogram w labels: state (%v) with error %v", state, err))
		} else {
			hist.Observe(duration.Seconds())
		}
	}
}

func (l *logger) LogAtLevel(entry *logrus.Entry, level logrus.Level, args ...interface{}) {
	entry.Log(level, args...)
}

func (l *logger) LogAtLevelFromDepth(ctx context.Context, request interface{}, response interface{}, err error, duration time.Duration, level logrus.Level, depth int) {
	// We're only interested in 1 stack frame, however due to weirdness with
	// inlining sometimes you need to get more than 1 caller so that
	// CallersFrames can resolve the first function. 2 seems to be enough be
	// we've set it to 5 to insulate us more, because this at one point broke
	// due to some compile optimization changes.
	pc := make([]uintptr, 5)
	runtime.Callers(depth, pc)
	frames := runtime.CallersFrames(pc)
	frame, ok := frames.Next()
	fields := logrus.Fields{}
	if ok {
		split := strings.Split(frame.Function, ".")
		method := split[len(split)-1]
		fields["method"] = method
	} else {
		fields["warn"] = "failed to resolve method"
	}
	fields["request"] = request
	if response != nil {
		fields["response"] = response
	}
	if err != nil {
		// "err" itself might be a code or even an empty struct
		fields["error"] = err.Error()
		var frames []string
		errors.ForEachStackFrame(err, func(frame errors.Frame) {
			frames = append(frames, fmt.Sprintf("%+v", frame))
		})
		fields["stack"] = frames
	}
	if duration > 0 {
		fields["duration"] = duration
	}
	if user := auth.GetWhoAmI(ctx); user != "" {
		fields["user"] = user
	}
	l.LogAtLevel(l.WithFields(fields), level)
}

func topLevelService(fullyQualifiedService string) string {
	tokens := strings.Split(fullyQualifiedService, ".")
	return tokens[0]
}

// FormatterFunc is a type alias for a function that satisfies logrus'
// `Formatter` interface
type FormatterFunc func(entry *logrus.Entry) ([]byte, error)

// Format proxies the closure in order to satisfy `logrus.Formatter`'s
// interface.
func (f FormatterFunc) Format(entry *logrus.Entry) ([]byte, error) {
	return f(entry)
}

// Pretty formats a logrus entry like so:
// ```
// 2019-02-11T16:02:02Z INFO pfs.API.InspectRepo {"request":{"repo":{"name":"images"}}} []
// ```
func Pretty(entry *logrus.Entry) ([]byte, error) {
	serialized := []byte(
		fmt.Sprintf(
			"%v %v ",
			entry.Time.Format(time.RFC3339),
			strings.ToUpper(entry.Level.String()),
		),
	)
	if entry.Data["service"] != nil && entry.Data["method"] != nil {
		serialized = append(serialized, []byte(fmt.Sprintf("%v.%v ", entry.Data["service"], entry.Data["method"]))...)
		delete(entry.Data, "service")
		delete(entry.Data, "method")
	}
	if len(entry.Data) > 0 {
		if entry.Data["duration"] != nil {
			entry.Data["duration"] = entry.Data["duration"].(time.Duration).Seconds()
		}
		data, err := json.Marshal(entry.Data)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to marshal fields to JSON")
		}
		serialized = append(serialized, data...)
		serialized = append(serialized, ' ')
	}

	serialized = append(serialized, []byte(entry.Message)...)
	serialized = append(serialized, '\n')
	return serialized, nil
}

// GRPCLogWriter proxies gRPC and etcd-produced log messages to a logrus
// logger. Because it implements `io.Writer`, it could be used anywhere where
// `io.Writer`s are used, but it has some logic specifically designed to
// handle gRPC-formatted logs.
type GRPCLogWriter struct {
	logger *logrus.Logger
	source string
}

// NewGRPCLogWriter creates a new GRPC log writer. `logger` specifies the
// underlying logger, and `source` specifies where these logs are coming from;
// it is added as a entry field for all log messages.
func NewGRPCLogWriter(logger *logrus.Logger, source string) *GRPCLogWriter {
	return &GRPCLogWriter{
		logger: logger,
		source: source,
	}
}

// Write allows `GRPCInfoWriter` to implement the `io.Writer` interface. This
// will take gRPC logs, which look something like this:
// ```
// INFO: 2019/02/18 12:21:54 ClientConn switching balancer to "pick_first"
// ```
// strip out redundant content, and print the message at the appropriate log
// level in logrus. Any parse errors of the log message will be reported in
// logrus as well.
func (l *GRPCLogWriter) Write(p []byte) (int, error) {
	parts := strings.SplitN(string(p), " ", 4)
	entry := l.logger.WithField("source", l.source)

	if len(parts) == 4 {
		// parts[1] and parts[2] contain the date and time, but logrus already
		// adds this under the `time` entry field, so it's not needed (though
		// the time will presumably be marginally ahead of the original log
		// message)
		level := parts[0]
		message := strings.TrimSpace(parts[3])

		if level == "INFO:" {
			entry.Info(message)
		} else if level == "ERROR:" {
			entry.Error(message)
		} else if level == "WARNING:" {
			entry.Warning(message)
		} else if level == "FATAL:" {
			// no need to call fatal ourselves because gRPC will exit the
			// process
			entry.Error(message)
		} else {
			entry.Error(message)
			entry.Errorf("entry had unknown log level prefix: '%s'; this is a bug, please report it along with the previous log entry", level)
		}
	} else {
		// can't format the message -- just display the contents
		entry := l.logger.WithFields(logrus.Fields{
			"source": l.source,
		})
		entry.Error(p)
		entry.Error("entry had unexpected format; this is a bug, please report it along with the previous log entry")
	}

	return len(p), nil
}
