package worker

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const (
	// PrometheusPort is the port the aggregated metrics are served on for scraping
	PrometheusPort = 9090
)

var (
	datumCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_count",
			Help:      "Number of datums processed by pipeline ID and state (started|errored|finished)",
		},
		[]string{
			"pipeline",
			"job",
			"state",
		},
	)

	bucketFactor  = 2.0
	bucketCount   = 20 // Which makes the max bucket 2^20 seconds or ~12 days in size
	datumProcTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_proc_time",
			Help:      "Time running user code",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"pipeline",
			"job",
			"state", // Since both finished and errored datums can have proc times
		},
	)
	datumProcSecondsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_proc_seconds_count",
			Help:      "Cumulative number of seconds spent processing",
		},
		[]string{
			"pipeline",
			"job",
		},
	)

	datumDownloadTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_download_time",
			Help:      "Time to download input data",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"pipeline",
			"job",
		},
	)
	datumDownloadSecondsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_download_seconds_count",
			Help:      "Cumulative number of seconds spent downloading",
		},
		[]string{
			"pipeline",
			"job",
		},
	)

	datumUploadTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_upload_time",
			Help:      "Time to upload output data",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"pipeline",
			"job",
		},
	)
	datumUploadSecondsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_upload_seconds_count",
			Help:      "Cumulative number of seconds spent uploading",
		},
		[]string{
			"pipeline",
			"job",
		},
	)

	datumDownloadSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_download_size",
			Help:      "Size of downloaded input data",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"pipeline",
			"job",
		},
	)
	datumDownloadBytesCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_download_bytes_count",
			Help:      "Cumulative number of bytes downloaded",
		},
		[]string{
			"pipeline",
			"job",
		},
	)

	datumUploadSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_upload_size",
			Help:      "Size of uploaded output data",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"pipeline",
			"job",
		},
	)
	datumUploadBytesCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_upload_bytes_count",
			Help:      "Cumulative number of bytes uploaded",
		},
		[]string{
			"pipeline",
			"job",
		},
	)
)

func initPrometheus() {
	metrics := []prometheus.Collector{
		datumCount,
		datumProcTime,
		datumProcSecondsCount,
		datumDownloadTime,
		datumDownloadSecondsCount,
		datumUploadTime,
		datumUploadSecondsCount,
		datumDownloadSize,
		datumDownloadBytesCount,
		datumUploadSize,
		datumUploadBytesCount,
	}
	for _, metric := range metrics {
		if err := prometheus.Register(metric); err != nil {
			// metrics may be redundantly registered; ignore these errors
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				logrus.Errorf("error registering prometheus metric: %v", err)
			}
		}
	}
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%v", PrometheusPort), nil); err != nil {
			logrus.Errorf("error serving prometheus metrics: %v", err)
		}
	}()
}
