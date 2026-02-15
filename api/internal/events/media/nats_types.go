package media

import (
	"time"

	"github.com/google/uuid"
)

// NATS header keys for media tasks.
const (
	NATSHeaderInstanceID = "X-Instance-ID"
	NATSHeaderEventID    = "X-Event-ID"
	NATSHeaderMediaType  = "X-Media-Type"
	NATSHeaderMediaKey   = "X-Media-Key"
)

// MediaTask represents a media processing task published to NATS.
type MediaTask struct {
	InstanceID  uuid.UUID `json:"instance_id"`
	EventID     uuid.UUID `json:"event_id"`
	MediaKey    string    `json:"media_key"`
	DirectPath  string    `json:"direct_path"`
	MediaType   string    `json:"media_type"`
	MimeType    string    `json:"mime_type,omitempty"`
	FileLength  int64     `json:"file_length,omitempty"`
	PublishedAt time.Time `json:"published_at"`
	Payload     string    `json:"payload"` // Encoded InternalEvent (base64)
}

// MediaResult represents the result of media processing.
type MediaResult struct {
	InstanceID  uuid.UUID `json:"instance_id"`
	EventID     uuid.UUID `json:"event_id"`
	Success     bool      `json:"success"`
	MediaURL    string    `json:"media_url,omitempty"`
	S3Key       string    `json:"s3_key,omitempty"`
	ContentType string    `json:"content_type,omitempty"`
	FileSize    int64     `json:"file_size,omitempty"`
	Error       string    `json:"error,omitempty"`
	ProcessedAt time.Time `json:"processed_at"`
}

// FastPathResult represents the outcome of a fast-path media processing attempt.
type FastPathResult struct {
	Success  bool
	MediaURL string
	Error    error
}

// NATSMediaConfig holds NATS-specific media processing configuration.
type NATSMediaConfig struct {
	// FastPathTimeout is the maximum time for fast-path inline processing.
	FastPathTimeout time.Duration

	// MaxAttempts before DLQ for media tasks.
	MaxAttempts int
}

// DefaultNATSMediaConfig returns sensible defaults.
func DefaultNATSMediaConfig() NATSMediaConfig {
	return NATSMediaConfig{
		FastPathTimeout: 5 * time.Second,
		MaxAttempts:     5,
	}
}
