package observability

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics bundles Prometheus collectors used across the service.
type Metrics struct {
	HTTPRequests *prometheus.CounterVec
	HTTPDuration *prometheus.HistogramVec
	WebhookQueue prometheus.Gauge
}

// NewMetrics registers collectors with the provided namespace.
func NewMetrics(namespace string, reg prometheus.Registerer) *Metrics {
	labels := []string{"method", "path", "status"}
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_total",
		Help:      "Total HTTP requests processed.",
	}, labels)
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_duration_seconds",
		Help:      "Duration of HTTP requests in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, labels)
	queue := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "webhook_outbox_backlog",
		Help:      "Number of webhook events pending delivery.",
	})

	reg.MustRegister(requests, duration, queue)

	return &Metrics{
		HTTPRequests: requests,
		HTTPDuration: duration,
		WebhookQueue: queue,
	}
}
