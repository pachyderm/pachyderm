package log

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/camelcase"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	bucketFactor = 2.0
	bucketCount  = 20 // Which makes the max bucket 2^20 seconds or ~12 days in size
)

// This needs to be a global var, not a field on the logger, because multiple servers
// create new loggers, and the prometheus registration uses a global namespace
var reportMetricGauge prometheus.Gauge
var reportMetricsOnce sync.Once

// Logger is a helper for emitting our grpc API logs
type Logger interface {
	Log(request interface{}, response interface{}, err error, duration time.Duration)
	LogAtLevelFromDepth(request interface{}, response interface{}, err error, duration time.Duration, level logrus.Level, depth int)
}

type logger struct {
	*logrus.Entry
	histogram   map[string]*prometheus.HistogramVec
	counter     map[string]prometheus.Counter
	mutex       *sync.Mutex
	exportStats bool
	service     string
}

// NewLogger creates a new logger
func NewLogger(service string) Logger {
	return newLogger(service, true)
}

// NewLocalLogger creates a new logger for local testing (which does not report prometheus metrics)
func NewLocalLogger(service string) Logger {
	return newLogger(service, false)
}

func newLogger(service string, exportStats bool) Logger {
	l := logrus.New()
	l.Formatter = new(prettyFormatter)
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
				entry := newLogger.WithFields(logrus.Fields{"method": "NewLogger"})
				newLogger.LogAtLevel(entry, logrus.WarnLevel, fmt.Sprintf("error registering prometheus metric: %v", newReportMetricGauge), err)
			} else {
				reportMetricGauge = newReportMetricGauge
			}
		})
	}
	return newLogger
}

// Helper function used to log requests and responses from our GRPC method
// implementations
func (l *logger) Log(request interface{}, response interface{}, err error, duration time.Duration) {
	if err != nil {
		l.LogAtLevelFromDepth(request, response, err, duration, logrus.ErrorLevel, 4)
	} else {
		l.LogAtLevelFromDepth(request, response, err, duration, logrus.InfoLevel, 4)
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
		runTimeName := fmt.Sprintf("%v_time", rootStatName)
		runTime, ok := l.histogram[runTimeName]
		if !ok {
			runTime = prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace: "pachyderm",
					Subsystem: l.service,
					Name:      runTimeName,
					Help:      fmt.Sprintf("Run time of %v", method),
					Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
				},
				[]string{
					"state", // Since both finished and errored API calls can have run times
				},
			)
			if err := prometheus.Register(runTime); err != nil {
				l.LogAtLevel(entry, logrus.WarnLevel, fmt.Sprintf("error registering prometheus metric: %v", runTimeName), err)
			} else {
				l.histogram[runTimeName] = runTime
			}
		}
		if hist, err := runTime.GetMetricWithLabelValues(state); err != nil {
			l.LogAtLevel(entry, logrus.WarnLevel, "failed to get histogram w labels: state (%v) with error %v", state, err)
		} else {
			hist.Observe(duration.Seconds())
		}
	}

	secondsCountName := fmt.Sprintf("%v_seconds_count", rootStatName)
	secondsCount, ok := l.counter[secondsCountName]
	if !ok {
		secondsCount = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: "pachyderm",
				Subsystem: l.service,
				Name:      secondsCountName,
				Help:      fmt.Sprintf("cumulative number of seconds spent in %v", method),
			},
		)
		if err := prometheus.Register(secondsCount); err != nil {
			l.LogAtLevel(entry, logrus.WarnLevel, fmt.Sprintf("error registering prometheus metric: %v", secondsCountName), err)
		} else {
			l.counter[secondsCountName] = secondsCount
		}
	}
	secondsCount.Add(duration.Seconds())

}

func (l *logger) LogAtLevel(entry *logrus.Entry, level logrus.Level, args ...interface{}) {
	switch level {
	case logrus.PanicLevel:
		entry.Panic(args)
	case logrus.FatalLevel:
		entry.Fatal(args)
	case logrus.ErrorLevel:
		entry.Error(args)
	case logrus.WarnLevel:
		entry.Warn(args)
	case logrus.InfoLevel:
		entry.Info(args)
	case logrus.DebugLevel:
		entry.Debug(args)
	}
}

func (l *logger) LogAtLevelFromDepth(request interface{}, response interface{}, err error, duration time.Duration, level logrus.Level, depth int) {
	pc := make([]uintptr, depth)
	runtime.Callers(depth, pc)
	split := strings.Split(runtime.FuncForPC(pc[0]).Name(), ".")
	method := split[len(split)-1]

	fields := logrus.Fields{
		"method":  method,
		"request": request,
	}
	if response != nil {
		fields["response"] = response
	}
	if err != nil {
		// "err" itself might be a code or even an empty struct
		fields["error"] = err.Error()
	}
	if duration > 0 {
		fields["duration"] = duration
	}
	l.LogAtLevel(l.WithFields(fields), level)
}

type prettyFormatter struct{}

func (f *prettyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	serialized := []byte(
		fmt.Sprintf(
			"%v %v ",
			entry.Time.Format(logrus.DefaultTimestampFormat),
			strings.ToUpper(entry.Level.String()),
		),
	)
	if entry.Data["service"] != nil {
		serialized = append(serialized, []byte(fmt.Sprintf("%v.%v ", entry.Data["service"], entry.Data["method"]))...)
	}
	if len(entry.Data) > 2 {
		delete(entry.Data, "service")
		delete(entry.Data, "method")
		if entry.Data["duration"] != nil {
			entry.Data["duration"] = entry.Data["duration"].(time.Duration).Seconds()
		}
		data, err := json.Marshal(entry.Data)
		if err != nil {
			return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
		}
		serialized = append(serialized, []byte(string(data))...)
		serialized = append(serialized, ' ')
	}

	serialized = append(serialized, []byte(entry.Message)...)
	serialized = append(serialized, '\n')
	return serialized, nil
}
