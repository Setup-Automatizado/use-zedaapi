package nats

import (
	"time"

	"github.com/google/uuid"
)

// NATS header keys for event metadata.
const (
	HeaderInstanceID = "X-Instance-ID"
	HeaderEventID    = "X-Event-ID"
	HeaderEventType  = "X-Event-Type"
	HeaderSourceLib  = "X-Source-Lib"
	HeaderCapturedAt = "X-Captured-At"
	HeaderHasMedia   = "X-Has-Media"
)

// NATSEventEnvelope wraps an InternalEvent for NATS transport.
type NATSEventEnvelope struct {
	// Event identification
	EventID    uuid.UUID `json:"event_id"`
	InstanceID uuid.UUID `json:"instance_id"`
	EventType  string    `json:"event_type"`
	SourceLib  string    `json:"source_lib"`

	// Timing
	CapturedAt  time.Time `json:"captured_at"`
	PublishedAt time.Time `json:"published_at"`

	// Media info (for fast-path / async media processing)
	HasMedia  bool       `json:"has_media"`
	MediaInfo *MediaInfo `json:"media_info,omitempty"`

	// The encoded internal event payload (base64 string)
	Payload string `json:"payload"`

	// Metadata from capture
	Metadata map[string]string `json:"metadata,omitempty"`
}

// MediaInfo carries media-related fields from InternalEvent.
type MediaInfo struct {
	MediaKey      string  `json:"media_key,omitempty"`
	DirectPath    string  `json:"direct_path,omitempty"`
	FileSHA256    *string `json:"file_sha256,omitempty"`
	FileEncSHA256 *string `json:"file_enc_sha256,omitempty"`
	MediaType     string  `json:"media_type,omitempty"`
	MimeType      *string `json:"mime_type,omitempty"`
	FileLength    *int64  `json:"file_length,omitempty"`
}

// NATSEventConfig holds NATS-specific event pipeline configuration.
type NATSEventConfig struct {
	// MaxAttempts before DLQ for event delivery
	MaxAttempts int

	// Retry delays for delivery failures
	RetryDelays []time.Duration

	// Webhook delivery timeout
	WebhookTimeout time.Duration

	// Webhook URL cache TTL
	WebhookCacheTTL time.Duration
}

// DefaultNATSEventConfig returns sensible defaults.
func DefaultNATSEventConfig() NATSEventConfig {
	return NATSEventConfig{
		MaxAttempts:     6,
		RetryDelays:     []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second},
		WebhookTimeout:  30 * time.Second,
		WebhookCacheTTL: 30 * time.Second,
	}
}
