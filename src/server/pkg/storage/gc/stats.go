package gc

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	summaryObjectives = map[float64]float64{
		0.5:  0.05,
		0.9:  0.01,
		0.99: 0.001,
	}
	sqlResults = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "gc",
			Name:      "sql_results",
			Help:      "garbage collector sql calls, count by operation and result type",
		},
		[]string{"operation", "result"},
	)
	sqlTime = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "pachyderm",
			Subsystem:  "gc",
			Name:       "sql_time",
			Help:       "time spent in sql by operation type, histogram by duration (seconds)",
			Objectives: summaryObjectives,
		},
		[]string{"operation"},
	)
)

func initPrometheus(registry prometheus.Registerer) {
	metrics := []prometheus.Collector{
		sqlResults,
		sqlTime,
	}
	for _, metric := range metrics {
		if err := registry.Register(metric); err != nil {
			// metrics may be redundantly registered; ignore these errors
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				fmt.Printf("error registering prometheus metric: %v", err)
			}
		}
	}
}

func collectSQLStats(operation string, f func() error) (retErr error) {
	start := time.Now()
	defer func() {
		result := "success"
		if retErr != nil {
			result = retErr.Error()
		}
		sqlResults.WithLabelValues(operation, result).Inc()
		sqlTime.WithLabelValues(operation).Observe(float64(time.Since(start).Milliseconds()))
	}()
	return f()
}
