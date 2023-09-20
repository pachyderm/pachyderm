package stats

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	// PrometheusPort is the port the aggregated metrics are served on for scraping
	PrometheusPort = 9090
)

func JobLabels(job *pps.Job) prometheus.Labels {
	return prometheus.Labels{
		"project":  job.Pipeline.Project.GetName(),
		"pipeline": job.Pipeline.Name,
		"job":      job.Id,
	}
}

func DatumLabels(job *pps.Job, state string) prometheus.Labels {
	return prometheus.Labels{
		"project":  job.Pipeline.Project.GetName(),
		"pipeline": job.Pipeline.Name,
		"job":      job.Id,
		"state":    state,
	}
}

var (
	bucketFactor = 2.0
	bucketCount  = 20 // Which makes the max bucket 2^20 seconds or ~12 days in size

	// DatumCount is a counter tracking the number of datums processed by a pipeline
	DatumCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_count",
			Help:      "Number of datums processed by pipeline ID and state (started|errored|finished)",
		},
		[]string{
			"project",
			"pipeline",
			"job",
			"state",
		},
	)

	// DatumProcTime is a histogram tracking the time spent in user code for datums processed by a pipeline
	DatumProcTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_proc_time",
			Help:      "Time running user code",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"project",
			"pipeline",
			"job",
			"state", // Since both finished and errored datums can have proc times
		},
	)

	// DatumProcSecondsCount is a counter tracking the total time spent in user code by a pipeline
	DatumProcSecondsCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_proc_seconds_count",
			Help:      "Cumulative number of seconds spent processing",
		},
		[]string{
			"project",
			"pipeline",
			"job",
			"state",
		},
	)

	// DatumDownloadTime is a histogram tracking the time spent downloading input data by a pipeline
	DatumDownloadTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_download_time",
			Help:      "Time to download input data",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"project",
			"pipeline",
			"job",
		},
	)

	// DatumDownloadSecondsCount is a counter tracking the total time spent downloading input data by a pipeline
	DatumDownloadSecondsCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_download_seconds_count",
			Help:      "Cumulative number of seconds spent downloading",
		},
		[]string{
			"project",
			"pipeline",
			"job",
		},
	)

	// DatumUploadTime is a histogram tracking the time spent uploading output data by a pipeline
	DatumUploadTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_upload_time",
			Help:      "Time to upload output data",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"project",
			"pipeline",
			"job",
		},
	)

	// DatumUploadSecondsCount is a counter tracking the total time spent uploading output data by a pipeline
	DatumUploadSecondsCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_upload_seconds_count",
			Help:      "Cumulative number of seconds spent uploading",
		},
		[]string{
			"project",
			"pipeline",
			"job",
		},
	)

	// DatumDownloadSize is a histogram tracking the size of input data downloaded by a pipeline
	DatumDownloadSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_download_size",
			Help:      "Size of downloaded input data",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"project",
			"pipeline",
			"job",
		},
	)

	// DatumDownloadBytesCount is a counter tracking the total size of input data downloaded by a pipeline
	DatumDownloadBytesCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_download_bytes_count",
			Help:      "Cumulative number of bytes downloaded",
		},
		[]string{
			"project",
			"pipeline",
			"job",
		},
	)

	// DatumUploadSize is a histogram tracking the size of output data uploaded by a pipeline
	DatumUploadSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_upload_size",
			Help:      "Size of uploaded output data",
			Buckets:   prometheus.ExponentialBuckets(1.0, bucketFactor, bucketCount),
		},
		[]string{
			"project",
			"pipeline",
			"job",
		},
	)

	// DatumUploadBytesCount is a counter tracking the total size of output data uploaded by a pipeline
	DatumUploadBytesCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pachyderm",
			Subsystem: "worker",
			Name:      "datum_upload_bytes_count",
			Help:      "Cumulative number of bytes uploaded",
		},
		[]string{
			"project",
			"pipeline",
			"job",
		},
	)
)

// InitPrometheus sets up the default datum stats collectors for use by worker
// code, and exposes the stats on an http endpoint.
func InitPrometheus(ctx context.Context) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", PrometheusPort),
		Handler: mux,
	}
	log.AddLoggerToHTTPServer(ctx, "http", server)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Error(ctx, "error serving prometheus metrics", zap.Error(err))
		}
	}()
}
