package worker

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// The port prometheus metrics are exposed on
	PrometheusPort = int32(9090)

	datumCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "user",
			Name:      "datum_count",
			Help:      "Number of datums counted by pipeline and state",
		},
		[]string{
			"pipeline",
			"state",
		},
	)

	bucketFactor  = 2.0
	bucketCount   = 20 // Which makes the max bucket 2^20 seconds or ~12 days in size
	datumProcTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "user",
			Name:      "datum_proc_time",
			Help:      "Time processing user code",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"pipeline",
			"state", // Since both finished and errored datums can have proc times
		},
	)

	datumDownloadTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "user",
			Name:      "datum_download_time",
			Help:      "Time to download input data",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"pipeline",
		},
	)

	datumUploadTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "user",
			Name:      "datum_upload_time",
			Help:      "Time to upload output data",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"pipeline",
		},
	)

	datumDownloadSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "user",
			Name:      "datum_download_size",
			Help:      "Size of downloaded input data",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"pipeline",
		},
	)

	datumUploadSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "user",
			Name:      "datum_upload_size",
			Help:      "Size of uploaded output data",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"pipeline",
		},
	)
)

func initPrometheus() {
	prometheus.MustRegister(datumCount)
	prometheus.MustRegister(datumProcTime)
	prometheus.MustRegister(datumDownloadTime)
	prometheus.MustRegister(datumUploadTime)
	prometheus.MustRegister(datumDownloadSize)
	prometheus.MustRegister(datumUploadSize)
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", PrometheusPort), nil))
	}()
}
