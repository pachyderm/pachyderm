package gc

import (
	"fmt"

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
	bucketFactor = 2.0
	bucketCount  = 20 // Which makes the max bucket 2^20 milliseconds

	requestResults = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "gc",
			Name:      "request_results",
			Help:      "garbage collector requests, count by request and result type",
		},
		[]string{"request", "result"},
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

	requestTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "gc",
			Name:      "request_time",
			Help:      "time spent in request by type, histogram by duration (milliseconds)",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{"request"},
	)

	deleteTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "gc",
			Name:      "delete_time",
			Help:      "time spent deleting objects from object storage, histogram by duration (milliseconds)",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
	)
)

func initPrometheus(registry prometheus.Registerer) {
	metrics := []prometheus.Collector{
		requestResults,
		sqlResults,
		requestTime,
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
