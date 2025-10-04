package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SourceLib string

const (
	SourceLibWhatsmeow SourceLib = "whatsmeow"
	SourceLibCloudAPI  SourceLib = "cloud_api"
)

type InternalEvent struct {
	InstanceID      uuid.UUID
	EventID         uuid.UUID
	EventType       string
	SourceLib       SourceLib
	RawPayload      interface{}
	Metadata        map[string]string
	CapturedAt      time.Time
	HasMedia        bool
	MediaKey        string
	DirectPath      string
	FileSHA256      *string
	FileEncSHA256   *string
	MediaType       string
	MimeType        *string
	FileLength      *int64
	MediaIsGIF      bool
	MediaIsAnimated bool
	MediaWidth      int
	MediaHeight     int
	MediaWaveform   []byte
	QuotedMessageID string
	QuotedSender    string
	QuotedRemoteJID string
	MentionedJIDs   []string
	IsForwarded     bool
	EphemeralExpiry int64
	TransportType   string
	TransportConfig json.RawMessage
}

type EventMetadata struct {
	InstanceID   uuid.UUID         `json:"instance_id"`
	CapturedAt   time.Time         `json:"captured_at"`
	SourceLib    string            `json:"source_lib"`
	EventVersion string            `json:"event_version,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"`
}

type TransformResult struct {
	Success            bool
	TransformedPayload json.RawMessage
	Metadata           EventMetadata
	Error              error
}

type MediaInfo struct {
	MediaKey      string
	DirectPath    string
	FileSHA256    *string
	FileEncSHA256 *string
	MediaType     string
	MimeType      *string
	FileLength    *int64
}

type EventWithMedia struct {
	Event     *InternalEvent
	MediaInfo *MediaInfo
}

type BufferStats struct {
	Capacity      int
	Size          int
	DroppedEvents int64
	TotalEvents   int64
}

type ProcessingResult struct {
	EventID     uuid.UUID
	InstanceID  uuid.UUID
	Success     bool
	Error       error
	ProcessedAt time.Time
	Duration    time.Duration
}
