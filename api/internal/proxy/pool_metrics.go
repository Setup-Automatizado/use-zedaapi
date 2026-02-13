package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
)

// PoolMetrics holds Prometheus metrics for the proxy pool system.
type PoolMetrics struct {
	syncTotal        *prometheus.CounterVec
	syncDuration     *prometheus.HistogramVec
	assignmentsTotal *prometheus.CounterVec
	releasesTotal    *prometheus.CounterVec
	swapsTotal       *prometheus.CounterVec
	poolSize         *prometheus.GaugeVec
	healingTotal     *prometheus.CounterVec
	healingDuration  *prometheus.HistogramVec
}

// NewPoolMetrics creates and registers pool-related Prometheus metrics.
func NewPoolMetrics(namespace string, reg prometheus.Registerer) *PoolMetrics {
	m := &PoolMetrics{
		syncTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "proxy_pool",
			Name:      "sync_total",
			Help:      "Total number of provider sync operations.",
		}, []string{"provider_id", "status"}),

		syncDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "proxy_pool",
			Name:      "sync_duration_seconds",
			Help:      "Duration of provider sync operations.",
			Buckets:   []float64{1, 5, 10, 30, 60, 120},
		}, []string{"provider_id"}),

		assignmentsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "proxy_pool",
			Name:      "assignments_total",
			Help:      "Total number of proxy assignments from pool.",
		}, []string{"provider_id", "status"}),

		releasesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "proxy_pool",
			Name:      "releases_total",
			Help:      "Total number of proxy releases back to pool.",
		}, []string{"reason"}),

		swapsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "proxy_pool",
			Name:      "swaps_total",
			Help:      "Total number of pool proxy swaps.",
		}, []string{"provider_id", "status"}),

		poolSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "proxy_pool",
			Name:      "size",
			Help:      "Current number of proxies in the pool by status.",
		}, []string{"provider_id", "status"}),

		healingTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "proxy_pool",
			Name:      "healing_total",
			Help:      "Total number of auto-healing operations.",
		}, []string{"instance_id", "status"}),

		healingDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "proxy_pool",
			Name:      "healing_duration_seconds",
			Help:      "Duration of auto-healing operations.",
			Buckets:   []float64{0.5, 1, 2.5, 5, 10, 30},
		}, []string{"instance_id"}),
	}

	reg.MustRegister(
		m.syncTotal,
		m.syncDuration,
		m.assignmentsTotal,
		m.releasesTotal,
		m.swapsTotal,
		m.poolSize,
		m.healingTotal,
		m.healingDuration,
	)

	return m
}

// RecordSync records a provider sync operation.
func (m *PoolMetrics) RecordSync(providerID, status string, duration float64) {
	m.syncTotal.WithLabelValues(providerID, status).Inc()
	m.syncDuration.WithLabelValues(providerID).Observe(duration)
}

// RecordAssignment records a proxy assignment from the pool.
func (m *PoolMetrics) RecordAssignment(providerID, status string) {
	m.assignmentsTotal.WithLabelValues(providerID, status).Inc()
}

// RecordRelease records a proxy release back to the pool.
func (m *PoolMetrics) RecordRelease(reason string) {
	m.releasesTotal.WithLabelValues(reason).Inc()
}

// RecordSwap records a pool proxy swap.
func (m *PoolMetrics) RecordSwap(providerID, status string) {
	m.swapsTotal.WithLabelValues(providerID, status).Inc()
}

// SetPoolSize sets the current pool size gauge for a provider and status.
func (m *PoolMetrics) SetPoolSize(providerID, status string, count float64) {
	m.poolSize.WithLabelValues(providerID, status).Set(count)
}

// RecordHealing records an auto-healing operation.
func (m *PoolMetrics) RecordHealing(instanceID, status string, duration float64) {
	m.healingTotal.WithLabelValues(instanceID, status).Inc()
	m.healingDuration.WithLabelValues(instanceID).Observe(duration)
}
