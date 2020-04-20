package metrics

import (
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// (bryce) the metrics code should probably be reorganized at some point.
// the current setup provides an easy way to collect metrics for both external and internal pfs/storage apis.

type metrics struct {
	requestCounter                           *prometheus.CounterVec
	requestSummary, requestSummaryThroughput *prometheus.SummaryVec
}

var (
	subsystems = make(map[string]*metrics)
	mu         sync.Mutex
)

const (
	trimPrefix = "github.com/pachyderm/pachyderm/src/"
)

func ReportRequest(f func() error, skip ...int) (retErr error) {
	ci := retrieveCallInfo(skip...)
	ms, err := maybeRegisterSubsystem(ci.packageName)
	if err != nil {
		return err
	}
	operation := ci.funcName
	start := time.Now()
	defer func() {
		result := "success"
		if retErr != nil {
			result = retErr.Error()
		}
		ms.requestCounter.WithLabelValues(operation, result).Inc()
		ms.requestSummary.WithLabelValues(operation).Observe(time.Since(start).Seconds())
	}()
	return f()
}

func ReportRequestWithThroughput(f func() (int64, error)) error {
	ci := retrieveCallInfo()
	ms, err := maybeRegisterSubsystem(ci.packageName)
	if err != nil {
		return err
	}
	operation := ci.funcName
	start := time.Now()
	return ReportRequest(func() error {
		bytesProcessed, err := f()
		throughput := float64(bytesProcessed) / time.Since(start).Seconds()
		ms.requestSummaryThroughput.WithLabelValues(operation).Observe(throughput)
		return err
	}, 1)
}

type callInfo struct {
	packageName string
	fileName    string
	funcName    string
	line        int
}

func retrieveCallInfo(skip ...int) *callInfo {
	skipFrames := 2
	if len(skip) > 0 {
		skipFrames += skip[0]
	}
	pc, file, line, _ := runtime.Caller(skipFrames)
	_, fileName := path.Split(file)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	packageName := ""
	funcName := parts[pl-1]

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
		packageName = strings.Join(parts[0:pl-2], ".")
	} else {
		packageName = strings.Join(parts[0:pl-1], ".")
	}

	return &callInfo{
		packageName: packageName,
		fileName:    fileName,
		funcName:    funcName,
		line:        line,
	}
}

func maybeRegisterSubsystem(packageName string) (*metrics, error) {
	subsystem := strings.ReplaceAll(strings.TrimPrefix(packageName, trimPrefix), "/", "_")
	mu.Lock()
	defer mu.Unlock()
	if ms, ok := subsystems[subsystem]; ok {
		return ms, nil
	}
	err := register(subsystem)
	return subsystems[subsystem], err
}

func register(subsystem string) error {
	ms := &metrics{
		requestCounter:           newRequestCounter(subsystem),
		requestSummary:           newRequestSummary(subsystem),
		requestSummaryThroughput: newRequestSummaryThroughput(subsystem),
	}
	for _, m := range []prometheus.Collector{
		ms.requestCounter,
		ms.requestSummary,
		ms.requestSummaryThroughput,
	} {
		if err := prometheus.Register(m); err != nil {
			return err
		}
	}
	subsystems[subsystem] = ms
	return nil
}

func newRequestCounter(subsystem string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: subsystem,
			Name:      "request_results",
			Help:      subsystem + " operations, count by operation and result type",
		},
		[]string{"operation", "result"},
	)
}

func newRequestSummary(subsystem string) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "pachyderm",
			Subsystem: subsystem,
			Name:      "request_time",
			Help:      "time spent on " + subsystem + " operations, histogram by duration (seconds)",
		},
		[]string{"operation"},
	)
}

func newRequestSummaryThroughput(subsystem string) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "pachyderm",
			Subsystem: subsystem,
			Name:      "request_throughput",
			Help:      "throughput of " + subsystem + " operations, histogram by throughput (MB/s)",
		},
		[]string{"operation"},
	)
}
