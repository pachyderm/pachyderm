package gc

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// PrometheusPort is the port the aggregated metrics are served on for scraping
	PrometheusPort = 9090
)

// Client:
// reserve_chunks_results
// reserve_chunks_time
// update_references_results
// update_references_time
//
// Server:
// flush_delete_results
// flush_delete_time
// delete_chunks_results
// delete_chunks_time
// mark_deleted_results
// mark_deleted_time
// delete_results
// delete_time
// remove_chunk_row_results
// remove_chunk_row_time

var (
	summaryObjectives = map[float64]float64{
		0.5:  0.05,
		0.9:  0.01,
		0.99: 0.001,
	}

	requestResults = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "gc",
			Name:      "request_results",
			Help:      "garbage collector requests, count by request and result type",
		},
		[]string{"request", "result"},
	)

	requestTime = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "pachyderm",
			Subsystem:  "gc",
			Name:       "request_time",
			Help:       "time spent in request by type, histogram by duration (seconds)",
			Objectives: summaryObjectives,
		},
		[]string{"request"},
	)

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

	deleteResults = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "gc",
			Name:      "delete_results",
			Help:      "garbage collector batch deletes to object storage, count by result type",
		},
		[]string{"result"},
	)

	deleteTime = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace:  "pachyderm",
			Subsystem:  "gc",
			Name:       "delete_time",
			Help:       "time spent deleting objects from object storage, histogram by duration (seconds)",
			Objectives: summaryObjectives,
		},
	)
)

func initPrometheus(registry prometheus.Registerer) {
	metrics := []prometheus.Collector{
		requestResults,
		requestTime,
		sqlResults,
		sqlTime,
		deleteResults,
		deleteTime,
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

func applyRequestStats(request string, err error, start time.Time) {
	var result string
	switch err {
	//TODO: add more resolution here
	case nil:
		result = "success"
	default:
		result = "error"
	}
	requestResults.WithLabelValues(request, result).Inc()
	requestTime.WithLabelValues(request).Observe(float64(time.Since(start).Seconds()))
}

func applySQLStats(operation string, err error, start time.Time) {
	var result string
	switch x := err.(type) {
	case nil:
		result = "success"
	case *pq.Error:
		result = x.Code.Name()
	default:
		fmt.Printf("SQL error in %s: %v\n", operation, err)
		result = "unknown"
	}
	sqlResults.WithLabelValues(operation, result).Inc()
	sqlTime.WithLabelValues(operation).Observe(float64(time.Since(start).Seconds()))
}

func applyDeleteStats(err error, start time.Time) {
	var result string
	switch err {
	// TODO: add more resolution here
	case nil:
		result = "success"
	default:
		result = "unknown"
	}
	deleteResults.WithLabelValues(result).Inc()
	deleteTime.Observe(float64(time.Since(start).Seconds()))
}
