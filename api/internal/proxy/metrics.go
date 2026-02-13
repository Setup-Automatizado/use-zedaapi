package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds Prometheus metrics for proxy health checking.
type Metrics struct {
	healthChecksTotal   *prometheus.CounterVec
	healthStatus        *prometheus.GaugeVec
	healthCheckDuration *prometheus.HistogramVec
	proxySwapTotal      *prometheus.CounterVec
	messagesDLQTotal    *prometheus.CounterVec
}

// NewMetrics creates and registers proxy-related Prometheus metrics.
func NewMetrics(namespace string, reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		healthChecksTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "proxy",
			Name:      "health_checks_total",
			Help:      "Total number of proxy health checks.",
		}, []string{"instance_id", "status"}),

		healthStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "proxy",
			Name:      "health_status",
			Help:      "Current proxy health status (1=healthy, 0=unhealthy).",
		}, []string{"instance_id"}),

		healthCheckDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "proxy",
			Name:      "health_check_duration_seconds",
			Help:      "Duration of proxy health checks.",
			Buckets:   []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}, []string{"instance_id"}),

		proxySwapTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "proxy",
			Name:      "swap_total",
			Help:      "Total number of proxy swaps.",
		}, []string{"instance_id", "status"}),

		messagesDLQTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "proxy",
			Name:      "messages_dlq_total",
			Help:      "Messages moved to DLQ due to proxy failures.",
		}, []string{"instance_id"}),
	}

	reg.MustRegister(
		m.healthChecksTotal,
		m.healthStatus,
		m.healthCheckDuration,
		m.proxySwapTotal,
		m.messagesDLQTotal,
	)

	return m
}

// RecordHealthCheck records a single proxy health check result.
func (m *Metrics) RecordHealthCheck(instanceID, status string, duration float64) {
	m.healthChecksTotal.WithLabelValues(instanceID, status).Inc()
	m.healthCheckDuration.WithLabelValues(instanceID).Observe(duration)
	val := 1.0
	if status != "healthy" {
		val = 0
	}
	m.healthStatus.WithLabelValues(instanceID).Set(val)
}

// RecordSwap records a proxy swap event.
func (m *Metrics) RecordSwap(instanceID, status string) {
	m.proxySwapTotal.WithLabelValues(instanceID, status).Inc()
}

// RecordDLQ records a message moved to the DLQ due to proxy failure.
func (m *Metrics) RecordDLQ(instanceID string) {
	m.messagesDLQTotal.WithLabelValues(instanceID).Inc()
}
