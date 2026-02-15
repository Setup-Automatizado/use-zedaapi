package nats

import "time"

// Config holds NATS connection and stream configuration.
type Config struct {
	// Connection
	URL            string        `json:"url"`
	Token          string        `json:"-"`
	ConnectTimeout time.Duration `json:"connect_timeout"`
	ReconnectWait  time.Duration `json:"reconnect_wait"`
	MaxReconnects  int           `json:"max_reconnects"`

	// Publish
	PublishTimeout time.Duration `json:"publish_timeout"`

	// Drain
	DrainTimeout time.Duration `json:"drain_timeout"`

	// Stream names (configurable for testing)
	StreamMessageQueue    string `json:"stream_message_queue"`
	StreamWhatsAppEvents  string `json:"stream_whatsapp_events"`
	StreamMediaProcessing string `json:"stream_media_processing"`
	StreamDLQ             string `json:"stream_dlq"`
}

// DefaultConfig returns a Config with production defaults.
func DefaultConfig() Config {
	return Config{
		URL:                   "nats://localhost:4222",
		ConnectTimeout:        10 * time.Second,
		ReconnectWait:         2 * time.Second,
		MaxReconnects:         -1, // unlimited
		PublishTimeout:        5 * time.Second,
		DrainTimeout:          30 * time.Second,
		StreamMessageQueue:    "MESSAGE_QUEUE",
		StreamWhatsAppEvents:  "WHATSAPP_EVENTS",
		StreamMediaProcessing: "MEDIA_PROCESSING",
		StreamDLQ:             "DLQ",
	}
}

// Validate checks that the config has required fields.
func (c Config) Validate() error {
	if c.URL == "" {
		return ErrInvalidConfig
	}
	return nil
}
