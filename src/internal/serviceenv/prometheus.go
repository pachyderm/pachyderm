package serviceenv

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	k8sInFlightMetric = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kubernetes_in_flight_requests",
		Help: "A gauge of in-flight requests being made against the k8s API",
	})
	k8sRequestCountMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kubernetes_requests_total",
		Help: "A summary of requests made against the k8s API",
	}, []string{"code", "method"})
	k8sRequestTimeMetric = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubernetes_request_duration_seconds",
		Help:    "A histogram of request timing against the k8s API",
		Buckets: []float64{0.001, 0.005, 0.01, 0.1, 1, 10, 30, 60, 300, 600, 1800, 3600},
	}, []string{"code", "method"})
)

func wrapK8sTransport(rt http.RoundTripper) http.RoundTripper {
	return promhttp.InstrumentRoundTripperInFlight(k8sInFlightMetric,
		promhttp.InstrumentRoundTripperDuration(k8sRequestTimeMetric,
			promhttp.InstrumentRoundTripperCounter(
				k8sRequestCountMetric, rt)))
}
