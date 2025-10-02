package observability

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	HTTPRequests *prometheus.CounterVec
	HTTPDuration *prometheus.HistogramVec
	WebhookQueue prometheus.Gauge

	LockAcquisitions           *prometheus.CounterVec
	LockReacquisitionAttempts  *prometheus.CounterVec
	LockReacquisitionFallbacks *prometheus.CounterVec
	CircuitBreakerState        prometheus.Gauge
	SplitBrainDetected         prometheus.Counter
	SplitBrainInvalidLocks     *prometheus.CounterVec

	HealthChecks *prometheus.CounterVec
}

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

	lockAcquisitions := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "lock_acquisitions_total",
		Help:      "Total lock acquisition attempts, labeled by status (success/failure).",
	}, []string{"status"})

	circuitBreakerState := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "circuit_breaker_state",
		Help:      "Current circuit breaker state (0=CLOSED, 1=OPEN, 2=HALF_OPEN).",
	})

	splitBrainDetected := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "split_brain_detected_total",
		Help:      "Total number of split-brain conditions detected.",
	})

	lockReacquisitionAttempts := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "lock_reacquisition_attempts_total",
		Help:      "Total lock reacquisition attempts after Redis recovery, labeled by instance_id and result (success/failure/fallback).",
	}, []string{"instance_id", "result"})

	lockReacquisitionFallbacks := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "lock_reacquisition_fallbacks_total",
		Help:      "Lock reacquisition attempts that returned fallback locks, labeled by instance_id and circuit_state.",
	}, []string{"instance_id", "circuit_state"})

	splitBrainInvalidLocks := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "split_brain_invalid_locks_total",
		Help:      "Locks marked as redis mode but with empty tokens (prevents false positives), labeled by instance_id.",
	}, []string{"instance_id"})

	healthChecks := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "health_checks_total",
		Help:      "Total health check attempts, labeled by component and status.",
	}, []string{"component", "status"})

	reg.MustRegister(
		requests, duration, queue,
		lockAcquisitions, lockReacquisitionAttempts, lockReacquisitionFallbacks,
		circuitBreakerState, splitBrainDetected, splitBrainInvalidLocks,
		healthChecks,
	)

	return &Metrics{
		HTTPRequests: requests,
		HTTPDuration: duration,
		WebhookQueue: queue,

		LockAcquisitions:           lockAcquisitions,
		LockReacquisitionAttempts:  lockReacquisitionAttempts,
		LockReacquisitionFallbacks: lockReacquisitionFallbacks,
		CircuitBreakerState:        circuitBreakerState,
		SplitBrainDetected:         splitBrainDetected,
		SplitBrainInvalidLocks:     splitBrainInvalidLocks,

		HealthChecks: healthChecks,
	}
}
