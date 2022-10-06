// Package promutil contains utilities for collecting Prometheus metrics.
package promutil

import (
	"io"
	"net/http"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
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

type loggingRT struct {
	name       string
	underlying http.RoundTripper
}

func (rt *loggingRT) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	log := logrus.WithFields(logrus.Fields{
		"name":   rt.name,
		"method": req.Method,
		"uri":    req.URL.String(),
	})

	// Log the start of long HTTP requests.
	timer := time.AfterFunc(10*time.Second, func() {
		l := log
		if dl, ok := req.Context().Deadline(); ok {
			l = l.WithField("deadline", time.Until(dl))
		}
		l.WithField("duration", time.Since(start)).Info("ongoing long http request")
	})
	defer timer.Stop()

	res, err := rt.underlying.RoundTrip(req)
	if err != nil {
		log.WithError(err).Info("outgoing http request completed with error")
		return res, errors.EnsureStack(err)
	}
	if res != nil {
		log.WithFields(logrus.Fields{
			"duration": time.Since(start),
			"status":   res.Status,
		}).Debugf("outgoing http request complete")
	}
	return res, errors.EnsureStack(err)
}

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
				&loggingRT{name: name, underlying: rt})))
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
