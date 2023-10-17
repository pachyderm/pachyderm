package metrics

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	units "github.com/docker/go-units"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// TODO The metrics code should probably be reorganized at some point.
// The current setup provides an easy way to collect metrics for both external and internal PFS/Storage APIs.

type metrics struct {
	requestCounter                           *prometheus.CounterVec
	requestSummary, requestSummaryThroughput *prometheus.SummaryVec
}

var (
	subsystems = make(map[string]*metrics)
	mu         sync.Mutex
)

const (
	trimPrefix = "github.com/pachyderm/pachyderm/v2/src/"
)

// ReportRequest reports a request to Prometheus.
// This function automatically registers a metric (if one does not already
// exist) with the default register.
// The calling function's package name is used as the subsystem name and the
// function name is used for the operation label.
// This function also labels the request as successful or not, and records
// the time spent in a separate metric.
func ReportRequest(f func() error, skip ...int) (retErr error) {
	start := time.Now()
	defer func() {
		ci, err := retrieveCallInfo(skip...)
		if err != nil {
			return
		}
		ms, err := maybeRegisterSubsystem(ci.packageName)
		if err != nil {
			return
		}
		operation := ci.funcName
		result := "success"
		if retErr != nil {
			result = retErr.Error()
		}
		ms.requestCounter.WithLabelValues(operation, result).Inc()
		ms.requestSummary.WithLabelValues(operation).Observe(time.Since(start).Seconds())
	}()
	err := f()
	if err != nil {
		fmt.Println("core-2002: error calling ReportRequest():", err.Error())
	}
	return err
}

// ReportRequestWithThroughput functions the same as ReportRequest, but also
// reports the throughput in a separate metric.
func ReportRequestWithThroughput(f func() (int64, error)) error {
	start := time.Now()
	return ReportRequest(func() error {
		bytesProcessed, err := f()
		defer func() {
			ci, err := retrieveCallInfo()
			if err != nil {
				return
			}
			ms, err := maybeRegisterSubsystem(ci.packageName)
			if err != nil {
				return
			}
			operation := ci.funcName
			throughput := float64(bytesProcessed) / units.MB / time.Since(start).Seconds()
			ms.requestSummaryThroughput.WithLabelValues(operation).Observe(throughput)
		}()
		if err != nil {
			fmt.Println("core-2002: error calling ReportRequestWithThroughput()", err.Error())
		}
		return err
	}, 1)
}

type callInfo struct {
	packageName string
	fileName    string
	funcName    string
	line        int
}

func retrieveCallInfo(skip ...int) (*callInfo, error) {
	skipFrames := 2
	if len(skip) > 0 {
		skipFrames += skip[0]
	}
	pc, file, line, ok := runtime.Caller(skipFrames)
	if !ok {
		return nil, errors.Errorf("could not retrieve caller info")
	}
	_, fileName := path.Split(file)
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return nil, errors.Errorf("could not retrieve function for program counter")
	}
	parts := strings.Split(fn.Name(), ".")
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
	}, nil
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
			return errors.EnsureStack(err)
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
