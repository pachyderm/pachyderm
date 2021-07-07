// Package promutil contains utilities for collecting Prometheus metrics.
package promutil

import (
	"io"
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
	if rt == nil {
		rt = http.DefaultTransport
	}
	ls := prometheus.Labels{"client": name}
	return promhttp.InstrumentRoundTripperInFlight(
		inFlightMetric.With(ls),
		promhttp.InstrumentRoundTripperDuration(
			requestTimeMetric.MustCurryWith(ls),
			promhttp.InstrumentRoundTripperCounter(
				requestCountMetric.MustCurryWith(ls),
				rt)))
}

// Adder is something that can be added to.
type Adder interface {
	Add(float64) // Implemented by prometheus.Counter.
}

// In the event that Prometheus changes their API, you'll be reading this comment.
var _ Adder = prometheus.NewCounter(prometheus.CounterOpts{})

// CountingReader exports a count of bytes read from an underlying io.Reader.
type CountingReader struct {
	io.Reader
	Counter Adder
}

// Read implements io.Reader.
func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.Counter.Add(float64(n))
	return
}

// CountingWriter exports a count of bytes written to an underlying io.Writer.
type CountingWriter struct {
	io.Writer
	Counter Adder
}

// Write implements io.Writer.
func (w *CountingWriter) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	w.Counter.Add(float64(n))
	return
}
