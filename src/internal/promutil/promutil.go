// Package promutil contains utilities for collecting Prometheus metrics.
package promutil

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	inFlightMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_client_in_flight_requests",
		Help: "A gauge of in-flight requests being made against an HTTP API, by client.",
	}, []string{"client"})

	requestCountMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_client_requests_total",
		Help: "A summary of requests made against an HTTP API, by client, status code, and request method.",
	}, []string{"client", "code", "method"})

	requestTimeMetric = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_client_request_duration_seconds",
		Help:    "A histogram of request timing against an HTTP API, by client and request method.",
		Buckets: []float64{0.001, 0.005, 0.01, 0.1, 1, 10, 30, 60, 300, 600, 1800, 3600},
	}, []string{"client", "method"})
)

// InstrumentRoundTripper returns an http.RoundTripper that collects Prometheus metrics; delegating
// to the underlying RoundTripper to actually make requests.
func InstrumentRoundTripper(name string, rt http.RoundTripper) http.RoundTripper {
	ls := prometheus.Labels{"client": name}
	return promhttp.InstrumentRoundTripperInFlight(
		inFlightMetric.With(ls),
		promhttp.InstrumentRoundTripperDuration(
			requestTimeMetric.MustCurryWith(ls),
			promhttp.InstrumentRoundTripperCounter(
				requestCountMetric.MustCurryWith(ls),
				rt)))
}
