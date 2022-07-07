package logging

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/camelcase"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type config struct {
	// valid values are only "text" and "json", where "json" is default
	format string
}

type Option func(*config)

func WithLogFormat(fmt string) func(*config) {
	return func(cfg *config) {
		if fmt != "text" {
			cfg.format = "json"
		}
		cfg.format = fmt
	}
}

// This needs to be a global var, not a field on the logger, because multiple servers
// create new loggers, and the prometheus registration uses a global namespace
var reportDurationGauge prometheus.Gauge
var reportDurationsOnce sync.Once

type LoggingInterceptor struct {
	logger    *logrus.Logger
	histogram map[string]*prometheus.HistogramVec
	counter   map[string]prometheus.Counter
	mutex     *sync.Mutex // synchronizes access to both histogram and counter maps
}

// NewLoggingInterceptor creates a new interceptor that logs method start and end
func NewLoggingInterceptor(logger *logrus.Logger, opts ...Option) *LoggingInterceptor {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}
	logger.Formatter = log.FormatterFunc(log.JSONPretty)
	if cfg.format == "text" {
		logger.Formatter = log.FormatterFunc(log.Pretty)
	}

	interceptor := &LoggingInterceptor{
		logger,
		make(map[string]*prometheus.HistogramVec),
		make(map[string]prometheus.Counter),
		&sync.Mutex{},
	}

	reportDurationsOnce.Do(func() {
		newReportMetricGauge := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "pachyderm",
				Subsystem: "pachd",
				Name:      "report_metric",
				Help:      "gauge of number of calls to reportDuration()",
			},
		)
		if err := prometheus.Register(newReportMetricGauge); err != nil {
			// metrics may be redundantly registered; ignore these errors
			if !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
				entry := logger.WithFields(logrus.Fields{"method": "NewLogger"})
				entry.Logf(logrus.WarnLevel, "error registering prometheus metric: %v", err)
			}
		} else {
			reportDurationGauge = newReportMetricGauge
		}
	})
	return interceptor
}

func parseMethod(fullMethod string) (string, string) {
	fullMethod = strings.Trim(fullMethod, "/")
	parts := strings.SplitN(fullMethod, "/", 2)
	service := strings.Replace(parts[0], "_v2.", ".", 1)
	if len(parts) < 2 {
		return service, ""
	}
	return service, parts[1]
}

func makeLogFields(ctx context.Context, request interface{}, fullMethod string, response interface{}, err error, duration time.Duration) logrus.Fields {
	fields := logrus.Fields{}
	fields["service"], fields["method"] = parseMethod(fullMethod)
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
	return fields
}

func (li *LoggingInterceptor) logUnaryBefore(ctx context.Context, level logrus.Level, req interface{}, fullMethod string, start time.Time) {
	fields := makeLogFields(ctx, req, fullMethod, nil, nil, 0)
	li.logger.WithFields(fields).Log(level)
}

func (li *LoggingInterceptor) logUnaryAfter(ctx context.Context, level logrus.Level, req interface{}, fullMethod string, start time.Time, resp interface{}, err error) {
	duration := time.Since(start)
	fields := makeLogFields(ctx, req, fullMethod, resp, err, duration)
	go li.reportDuration(fields["service"].(string), fields["method"].(string), duration, err)
	li.logger.WithFields(fields).Log(level)
}

func topLevelService(fullyQualifiedService string) string {
	tokens := strings.Split(fullyQualifiedService, ".")
	return tokens[0]
}

func (li *LoggingInterceptor) getHistogram(service string, method string) *prometheus.HistogramVec {
	fullyQualifiedName := fmt.Sprintf("%v/%v", service, method)
	histVec, ok := li.histogram[fullyQualifiedName]
	if !ok {
		var tokens []string
		for _, token := range camelcase.Split(method) {
			tokens = append(tokens, strings.ToLower(token))
		}
		rootStatName := strings.Join(tokens, "_")

		histogramName := fmt.Sprintf("%v_seconds", rootStatName)
		histVec = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "pachyderm",
				Subsystem: fmt.Sprintf("pachd_%v", topLevelService(service)),
				Name:      histogramName,
				Help:      fmt.Sprintf("Run time of %v", method),
				Buckets:   []float64{0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300, 600, 1800, 3600, 86400},
			},
			[]string{
				"state", // Since both finished and errored API calls can have run times
			},
		)
		if err := prometheus.Register(histVec); err != nil {
			// metrics may be redundantly registered; ignore these errors
			if !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
				entry := li.logger.WithFields(logrus.Fields{"method": method})
				entry.Logf(logrus.WarnLevel, "error registering prometheus metric %v: %v", histogramName, err)
			}
		} else {
			li.histogram[fullyQualifiedName] = histVec
		}
	}
	return histVec
}

func (li *LoggingInterceptor) reportDuration(service string, method string, duration time.Duration, err error) {
	// Count the number of reportDuration() goros in case we start to leak them
	if reportDurationGauge != nil {
		reportDurationGauge.Inc()
	}
	defer func() {
		if reportDurationGauge != nil {
			reportDurationGauge.Dec()
		}
	}()
	li.mutex.Lock() // for concurrent map access (histogram,counter)
	defer li.mutex.Unlock()

	state := "finished"
	if err != nil {
		state = "errored"
	}
	if hist, err := li.getHistogram(service, method).GetMetricWithLabelValues(state); err != nil {
		entry := li.logger.WithFields(logrus.Fields{"method": method})
		entry.Logf(logrus.WarnLevel, "failed to get histogram for state %v: %v", state, err)
	} else {
		hist.Observe(duration.Seconds())
	}
}
