package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SourceLib identifies the source library that generated the event
type SourceLib string

const (
	SourceLibWhatsmeow  SourceLib = "whatsmeow"
	SourceLibCloudAPI   SourceLib = "cloud_api"
)

// InternalEvent represents a captured event before persistence
type InternalEvent struct {
	// Identity
	InstanceID uuid.UUID
	EventID    uuid.UUID
	EventType  string
	SourceLib  SourceLib

	// Payload
	RawPayload interface{}       // Original event from source library
	Metadata   map[string]string // Additional metadata

	// Timing
	CapturedAt time.Time

	// Media information
	HasMedia    bool
	MediaKey    string
	DirectPath  string
	FileSHA256  *string
	FileEncSHA256 *string
	MediaType   string
	MimeType    *string
	FileLength  *int64

	// Media metadata
	MediaIsGIF     bool
	MediaIsAnimated bool
	MediaWidth     int
	MediaHeight    int
	MediaWaveform  []byte

	// ContextInfo (quotes, mentions, forwards, ephemeral)
	QuotedMessageID   string
	QuotedSender      string
	QuotedRemoteJID   string
	MentionedJIDs     []string
	IsForwarded       bool
	EphemeralExpiry   int64

	// Transport configuration
	TransportType   string
	TransportConfig json.RawMessage
}

// EventMetadata contains additional event metadata
type EventMetadata struct {
	InstanceID   uuid.UUID         `json:"instance_id"`
	CapturedAt   time.Time         `json:"captured_at"`
	SourceLib    string            `json:"source_lib"`
	EventVersion string            `json:"event_version,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"`
}

// TransformResult represents the result of event transformation
type TransformResult struct {
	Success        bool
	TransformedPayload json.RawMessage
	Metadata       EventMetadata
	Error          error
}

// MediaInfo contains media-specific information extracted from events
type MediaInfo struct {
	MediaKey      string
	DirectPath    string
	FileSHA256    *string
	FileEncSHA256 *string
	MediaType     string
	MimeType      *string
	FileLength    *int64
}

// EventWithMedia represents an event that contains media
type EventWithMedia struct {
	Event     *InternalEvent
	MediaInfo *MediaInfo
}

// BufferStats represents buffer statistics
type BufferStats struct {
	Capacity    int
	Size        int
	DroppedEvents int64
	TotalEvents int64
}

// ProcessingResult represents the result of event processing
type ProcessingResult struct {
	EventID       uuid.UUID
	InstanceID    uuid.UUID
	Success       bool
	Error         error
	ProcessedAt   time.Time
	Duration      time.Duration
}
