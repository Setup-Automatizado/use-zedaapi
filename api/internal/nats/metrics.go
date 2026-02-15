package nats

import "github.com/prometheus/client_golang/prometheus"

// NATSMetrics holds Prometheus metrics for NATS operations.
type NATSMetrics struct {
	// Publish metrics
	PublishTotal    *prometheus.CounterVec
	PublishDuration *prometheus.HistogramVec
	PublishErrors   *prometheus.CounterVec

	// Consume metrics
	ConsumeTotal    *prometheus.CounterVec
	ConsumeDuration *prometheus.HistogramVec
	ConsumeErrors   *prometheus.CounterVec

	// Ack metrics
	AckTotal *prometheus.CounterVec
	NakTotal *prometheus.CounterVec

	// Stream metrics
	StreamMessages *prometheus.GaugeVec
	StreamBytes    *prometheus.GaugeVec

	// Connection metrics
	ConnectionStatus     prometheus.Gauge
	ReconnectionTotal    prometheus.Counter
	DisconnectionTotal   prometheus.Counter
	ConnectionErrorTotal prometheus.Counter
}

// NewNATSMetrics creates and registers NATS Prometheus metrics.
func NewNATSMetrics(namespace string, registerer prometheus.Registerer) *NATSMetrics {
	m := &NATSMetrics{
		PublishTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "publish_total",
			Help:      "Total number of messages published to NATS",
		}, []string{"stream", "status"}),

		PublishDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "publish_duration_seconds",
			Help:      "Duration of NATS publish operations",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		}, []string{"stream"}),

		PublishErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "publish_errors_total",
			Help:      "Total number of NATS publish errors",
		}, []string{"stream", "error_type"}),

		ConsumeTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "consume_total",
			Help:      "Total number of messages consumed from NATS",
		}, []string{"stream", "consumer", "status"}),

		ConsumeDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "consume_duration_seconds",
			Help:      "Duration of NATS message processing",
			Buckets:   []float64{.001, .01, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"stream", "consumer"}),

		ConsumeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "consume_errors_total",
			Help:      "Total number of NATS consume errors",
		}, []string{"stream", "consumer", "error_type"}),

		AckTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "ack_total",
			Help:      "Total number of message acknowledgments",
		}, []string{"stream", "consumer"}),

		NakTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "nak_total",
			Help:      "Total number of message negative acknowledgments",
		}, []string{"stream", "consumer"}),

		StreamMessages: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "stream_messages",
			Help:      "Current number of messages in stream",
		}, []string{"stream"}),

		StreamBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "stream_bytes",
			Help:      "Current bytes in stream",
		}, []string{"stream"}),

		ConnectionStatus: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "connection_status",
			Help:      "NATS connection status (1=connected, 0=disconnected)",
		}),

		ReconnectionTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "reconnection_total",
			Help:      "Total number of NATS reconnections",
		}),

		DisconnectionTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "disconnection_total",
			Help:      "Total number of NATS disconnections",
		}),

		ConnectionErrorTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "nats",
			Name:      "connection_error_total",
			Help:      "Total number of NATS connection errors",
		}),
	}

	registerer.MustRegister(
		m.PublishTotal,
		m.PublishDuration,
		m.PublishErrors,
		m.ConsumeTotal,
		m.ConsumeDuration,
		m.ConsumeErrors,
		m.AckTotal,
		m.NakTotal,
		m.StreamMessages,
		m.StreamBytes,
		m.ConnectionStatus,
		m.ReconnectionTotal,
		m.DisconnectionTotal,
		m.ConnectionErrorTotal,
	)

	return m
}
