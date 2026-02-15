package queue

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NATS header keys for message metadata.
const (
	HeaderInstanceID    = "X-Instance-ID"
	HeaderZaapID        = "X-Zaap-ID"
	HeaderMessageType   = "X-Message-Type"
	HeaderProcessingKey = "X-Processing-Key"
	HeaderEnqueuedAt    = "X-Enqueued-At"
	HeaderScheduledAt   = "X-Scheduled-At"
	HeaderAttempt       = "X-Attempt"
	HeaderMaxAttempts   = "X-Max-Attempts"
)

// NATSMessageEnvelope wraps a SendMessageArgs for NATS transport.
type NATSMessageEnvelope struct {
	// Message identification
	ZaapID     string    `json:"zaap_id"`
	InstanceID uuid.UUID `json:"instance_id"`

	// Scheduling
	ScheduledAt time.Time `json:"scheduled_at"`
	EnqueuedAt  time.Time `json:"enqueued_at"`

	// Retry tracking
	Attempt     int `json:"attempt"`
	MaxAttempts int `json:"max_attempts"`

	// The actual message payload
	Payload json.RawMessage `json:"payload"`
}

// NATSConfig holds NATS-specific queue configuration.
type NATSConfig struct {
	// MaxAttempts before DLQ
	MaxAttempts int

	// Backoff settings (used for scheduled retries)
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64

	// Disconnect/proxy retry delays
	DisconnectRetryDelay time.Duration
	ProxyRetryDelay      time.Duration
}

// DefaultNATSConfig returns sensible defaults for NATS queue config.
func DefaultNATSConfig() NATSConfig {
	return NATSConfig{
		MaxAttempts:          5,
		InitialBackoff:       1 * time.Second,
		MaxBackoff:           5 * time.Minute,
		BackoffMultiplier:    2.0,
		DisconnectRetryDelay: 30 * time.Second,
		ProxyRetryDelay:      2 * time.Minute,
	}
}
