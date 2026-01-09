package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SendMessageArgs represents the arguments for a send message job.
// This struct is serialized to JSON and stored in message_queue.payload column.
//
// Queue Requirements:
// - Must be JSON serializable
// - All fields must be exported (capitalized)
//
// FIFO Ordering:
// - Guaranteed by BIGSERIAL id column (auto-incrementing)
// - ScheduledFor: Calculated timestamp considering DelayMessage
// - FOR UPDATE SKIP LOCKED ensures safe concurrent processing
type SendMessageArgs struct {
	// Message identification
	ZaapID     string    `json:"zaap_id"`     // Unique message ID returned to client
	InstanceID uuid.UUID `json:"instance_id"` // WhatsApp instance UUID

	// Recipient information
	Phone string `json:"phone"` // Phone number in format: 5511999999999@s.whatsapp.net, @g.us, @newsletter, @broadcast, @lid

	// Message content (only one should be set)
	MessageType         MessageType          `json:"message_type"` // text, image, audio, video, document, location, contact, interactive, sticker, ptv
	TextContent         *TextMessage         `json:"text_content,omitempty"`
	ImageContent        *MediaMessage        `json:"image_content,omitempty"`
	AudioContent        *MediaMessage        `json:"audio_content,omitempty"`
	VideoContent        *MediaMessage        `json:"video_content,omitempty"`
	DocumentContent     *MediaMessage        `json:"document_content,omitempty"`
	StickerContent      *MediaMessage        `json:"sticker_content,omitempty"` // WebP sticker (converted from image if needed)
	PTVContent          *MediaMessage        `json:"ptv_content,omitempty"`     // Circular video (Push-To-Talk Video)
	LocationContent     *LocationMessage     `json:"location_content,omitempty"`
	ContactContent      *ContactMessage      `json:"contact_content,omitempty"`
	ContactsContent     []*ContactMessage    `json:"contacts_content,omitempty"` // Multiple contacts sent in ONE message (ContactsArrayMessage)
	InteractiveContent  *InteractiveMessage  `json:"interactive_content,omitempty"`
	LinkPreviewOverride *LinkPreviewOverride `json:"link_preview_override,omitempty"`
	PollContent         *PollMessage         `json:"poll_content,omitempty"`
	EventContent        *EventMessage        `json:"event_content,omitempty"`

	// Timing configuration
	DelayMessage int64 `json:"delay_message"` // Delay in milliseconds BEFORE enqueue (affects scheduled_at)
	DelayTyping  int64 `json:"delay_typing"`  // Typing indicator duration in milliseconds BEFORE send (or "recording audio" for audio)

	// Message options
	ViewOnce         bool   `json:"view_once,omitempty"`           // If true, message can only be viewed once (image, audio, video only)
	ReplyToMessageID string `json:"reply_to_message_id,omitempty"` // WhatsApp message ID to reply to
	Duration         *int   `json:"duration,omitempty"`            // Ephemeral message duration in seconds (0=off, 86400=24h, 604800=7d, 7776000=90d)
	PrivateAnswer    bool   `json:"private_answer,omitempty"`      // For group messages: if true, reply in private to sender (not yourself)

	// Mention options
	Mentioned      []string       `json:"mentioned,omitempty"`       // Array of phone numbers to mention (e.g., ["5511999999999"])
	GroupMentioned []GroupMention `json:"group_mentioned,omitempty"` // Array of groups to mention in communities
	MentionedAll   bool           `json:"mentioned_all,omitempty"`   // If true, mentions all members in a group (without listing them)

	// Link preview option
	LinkPreview *bool `json:"link_preview,omitempty"` // If nil, auto-detect; if true, force preview; if false, disable preview

	// FIFO ordering fields
	SequenceNumber int64     `json:"sequence_number"` // Monotonic sequence for FIFO ordering
	ScheduledFor   time.Time `json:"scheduled_for"`   // When this message should be processed

	// Metadata for tracking and debugging
	EnqueuedAt        time.Time              `json:"enqueued_at"`                   // When the job was created
	WhatsAppMessageID string                 `json:"whatsapp_message_id,omitempty"` // Real WhatsApp message ID (filled after send)
	Metadata          map[string]interface{} `json:"metadata,omitempty"`            // Additional custom metadata
}

// Kind returns the message type identifier.
// This is used for logging and debugging purposes.
func (SendMessageArgs) Kind() string {
	return "send_message"
}

// MessageType defines the type of WhatsApp message
type MessageType string

const (
	MessageTypeText        MessageType = "text"
	MessageTypeImage       MessageType = "image"
	MessageTypeAudio       MessageType = "audio"
	MessageTypeVideo       MessageType = "video"
	MessageTypeDocument    MessageType = "document"
	MessageTypeLocation    MessageType = "location"
	MessageTypeContact     MessageType = "contact"
	MessageTypeInteractive MessageType = "interactive"
	MessageTypePoll        MessageType = "poll"
	MessageTypeEvent       MessageType = "event"
	MessageTypeSticker     MessageType = "sticker" // WebP sticker message
	MessageTypePTV         MessageType = "ptv"     // Circular video (Push-To-Talk Video)

	// interactive message types
	MessageTypeButtonList    MessageType = "button_list"
	MessageTypeButtonActions MessageType = "button_actions"
	MessageTypeOptionList    MessageType = "option_list"
	MessageTypeButtonPIX     MessageType = "button_pix"
	MessageTypeButtonOTP     MessageType = "button_otp"
	MessageTypeCarousel      MessageType = "carousel"
)

// TextMessage represents a text message
type TextMessage struct {
	Message string `json:"message"` // Message text (supports WhatsApp formatting)
}

// GroupMention represents a group to be mentioned in a message (for communities)
// format
type GroupMention struct {
	Phone   string `json:"phone"`   // Group JID (e.g., "120363XXXXX@g.us")
	Subject string `json:"subject"` // Group name/subject for display
}

// MediaMessage represents a media message (image, audio, video, document)
type MediaMessage struct {
	MediaURL string  `json:"media_url"`           // URL or base64 data
	Caption  *string `json:"caption,omitempty"`   // Optional caption for image/video/document
	FileName *string `json:"file_name,omitempty"` // Optional filename for document
	MimeType *string `json:"mime_type,omitempty"` // Optional MIME type
	IsPTV    bool    `json:"is_ptv,omitempty"`
}

// LocationMessage represents a location message
type LocationMessage struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      *string `json:"name,omitempty"`    // Location name
	Address   *string `json:"address,omitempty"` // Location address
}

// ContactMessage represents a contact message
// Supports both pre-formatted vCard (FUNNELCHAT) and structured fields
type ContactMessage struct {
	// Option 1: Pre-formatted vCard string
	// If provided, this takes precedence over individual fields
	VCard *string `json:"vcard,omitempty"`

	// Option 2: Individual fields (will be converted to vCard if VCard is nil)

	// Name fields (RFC 6350)
	FullName   string  `json:"full_name"`             // FN: Full formatted name (required)
	FirstName  *string `json:"first_name,omitempty"`  // N: Given name
	LastName   *string `json:"last_name,omitempty"`   // N: Family name
	MiddleName *string `json:"middle_name,omitempty"` // N: Additional names
	NamePrefix *string `json:"name_prefix,omitempty"` // N: Honorific prefix (Dr., Mr., Ms.)
	NameSuffix *string `json:"name_suffix,omitempty"` // N: Honorific suffix (Jr., Sr., III)
	Nickname   *string `json:"nickname,omitempty"`    // NICKNAME: Nickname or alias

	// Contact fields
	PhoneNumber string  `json:"phone_number"`    // TEL: Phone number in international format (required)
	Email       *string `json:"email,omitempty"` // EMAIL: Email address
	URL         *string `json:"url,omitempty"`   // URL: Website or social media URL

	// Professional fields
	Organization *string `json:"organization,omitempty"` // ORG: Organization/company name
	JobTitle     *string `json:"job_title,omitempty"`    // TITLE: Job title or position

	// Address field
	Address *ContactAddress `json:"address,omitempty"` // ADR: Structured address

	// Personal fields
	Birthday *string `json:"birthday,omitempty"` // BDAY: Birthday in YYYY-MM-DD format
	Note     *string `json:"note,omitempty"`     // NOTE: Additional notes or comments
}

// ContactAddress represents a structured address for vCard (RFC 6350)
// Format: ADR;TYPE=work:;;street;city;state;postal;country
type ContactAddress struct {
	Type       *string `json:"type,omitempty"`        // Address type: work, home
	PostBox    *string `json:"post_box,omitempty"`    // Post office box
	Extended   *string `json:"extended,omitempty"`    // Extended address (apartment, suite)
	Street     *string `json:"street,omitempty"`      // Street address
	City       *string `json:"city,omitempty"`        // City or locality
	Region     *string `json:"region,omitempty"`      // State, province or region
	PostalCode *string `json:"postal_code,omitempty"` // Postal code
	Country    *string `json:"country,omitempty"`     // Country name
}

// InteractiveMessage represents an interactive message (buttons, lists, carousel)
type InteractiveMessage struct {
	Type     InteractiveType `json:"type"` // button, list, carousel
	Header   *string         `json:"header,omitempty"`
	Body     string          `json:"body"`
	Footer   *string         `json:"footer,omitempty"`
	Buttons  []Button        `json:"buttons,omitempty"`  // For button type
	Sections []Section       `json:"sections,omitempty"` // For list type

	// fields
	ButtonLabel      *string `json:"button_label,omitempty"`      // For list messages - button text
	Image            *string `json:"image,omitempty"`             // URL or base64 for header image
	Video            *string `json:"video,omitempty"`             // URL or base64 for header video
	Document         *string `json:"document,omitempty"`          // URL or base64 for header document (PDF, etc.)
	DocumentFilename *string `json:"document_filename,omitempty"` // Filename for document

	// PIX-specific fields
	PIXPayment *PIXPayment `json:"pix_payment,omitempty"` // For button_pix type

	// OTP-specific fields
	OTPCode *string `json:"otp_code,omitempty"` // For button_otp type

	// Carousel-specific fields
	CarouselCards    []CarouselCard `json:"carousel_cards,omitempty"`     // For carousel type - array of cards
	CarouselCardType string         `json:"carousel_card_type,omitempty"` // HSCROLL_CARDS or ALBUM_IMAGE
}

// CarouselCard represents a single card in a carousel message
type CarouselCard struct {
	Header   string   `json:"header,omitempty"`    // Card header text (max 60 chars)
	Body     string   `json:"body"`                // Card body text (required, max 1024 chars)
	Footer   string   `json:"footer,omitempty"`    // Card footer text (max 60 chars)
	Buttons  []Button `json:"buttons"`             // Card buttons (1-3 buttons)
	MediaURL string   `json:"media_url,omitempty"` // URL for card image/video
}

// InteractiveType defines the type of interactive message
type InteractiveType string

const (
	InteractiveTypeButton   InteractiveType = "button"
	InteractiveTypeList     InteractiveType = "list"
	InteractiveTypeCarousel InteractiveType = "carousel"
)

// Button represents a button in an interactive message
type Button struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Type     string `json:"type,omitempty"`
	URL      string `json:"url,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Payload  string `json:"payload,omitempty"`
	CopyCode string `json:"copyCode,omitempty"`
}

// Section represents a section in a list message
type Section struct {
	Title string `json:"title"`
	Rows  []Row  `json:"rows"`
}

// Row represents a row in a list section
type Row struct {
	ID          string  `json:"id"`                    // Row identifier
	Title       string  `json:"title"`                 // Row title (max 24 chars)
	Description *string `json:"description,omitempty"` // Row description (max 72 chars)
}

// PIXPayment represents PIX payment data for Brazilian payment messages
// Used with button_pix message type
type PIXPayment struct {
	Key           string   `json:"key"`                      // PIX key (CPF, CNPJ, email, phone, or EVP)
	KeyType       string   `json:"key_type"`                 // Key type: CPF, CNPJ, EMAIL, PHONE, EVP
	Name          *string  `json:"name,omitempty"`           // Beneficiary name
	Amount        *float64 `json:"amount,omitempty"`         // Amount in BRL
	TransactionID *string  `json:"transaction_id,omitempty"` // PIX transaction ID
}

// LinkPreviewOverride mirrors the explicit metadata provided by /send-link
type LinkPreviewOverride struct {
	URL         string `json:"url"`
	Image       string `json:"image,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	LinkType    string `json:"link_type,omitempty"`
}

// PollMessage represents a poll creation payload
type PollMessage struct {
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	MaxSelections int      `json:"max_selections,omitempty"`
}

// EventMessage encapsulates the scheduled event payload
type EventMessage struct {
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      *time.Time     `json:"end_time,omitempty"`
	Location     *EventLocation `json:"location,omitempty"`
	CallLinkType string         `json:"call_link_type,omitempty"`
	Canceled     bool           `json:"canceled,omitempty"`
}

// EventLocation describes the physical place of the event
type EventLocation struct {
	Name             string   `json:"name,omitempty"`
	DegreesLongitude *float64 `json:"degrees_longitude,omitempty"`
	DegreesLatitude  *float64 `json:"degrees_latitude,omitempty"`
}

// InsertOptions holds configuration for job insertion
type InsertOptions struct {
	// Priority: Lower values = higher priority
	// For FIFO ordering, this is set to -SequenceNumber
	Priority int

	// ScheduledAt: When the job should become available for processing
	// For FIFO with delays, this is calculated as:
	// PreviousScheduledAt + DelayMessage + RandomJitter(1-3s)
	ScheduledAt time.Time

	// Queue: The queue name for this job
	// Format: "instance-{uuid}"
	Queue string

	// MaxAttempts: Maximum number of retry attempts
	MaxAttempts int

	// Metadata: Custom metadata stored with the job
	// Used for quick lookups without deserializing args
	Metadata map[string]interface{}
}

// JobStatus represents the current status of a message in the queue
type JobStatus string

const (
	JobStatusAvailable JobStatus = "available" // Ready to be picked up
	JobStatusRunning   JobStatus = "running"   // Currently being processed
	JobStatusCompleted JobStatus = "completed" // Successfully finished
	JobStatusCancelled JobStatus = "cancelled" // Manually cancelled
	JobStatusDiscarded JobStatus = "discarded" // Failed permanently (max retries exceeded)
	JobStatusRetryable JobStatus = "retryable" // Failed but will retry
	JobStatusScheduled JobStatus = "scheduled" // Waiting for scheduled_at time
)

// QueueJobInfo represents a job in the queue with metadata
type QueueJobInfo struct {
	ID                int64        `json:"id"`
	ZaapID            string       `json:"zaap_id"`
	InstanceID        uuid.UUID    `json:"instance_id"`
	MessageType       MessageType  `json:"message_type"`
	Phone             string       `json:"phone"`
	Status            JobStatus    `json:"status"`
	SequenceNumber    int64        `json:"sequence_number"`
	ScheduledFor      time.Time    `json:"scheduled_for"`
	CreatedAt         time.Time    `json:"created_at"`
	AttemptedAt       *time.Time   `json:"attempted_at,omitempty"`
	FinalizedAt       *time.Time   `json:"finalized_at,omitempty"`
	Errors            []string     `json:"errors,omitempty"`
	Attempt           int          `json:"attempt"`
	MaxAttempts       int          `json:"max_attempts"`
	WhatsAppMessageID string       `json:"whatsapp_message_id,omitempty"` // Real WhatsApp message ID
	DelayMessage      int64        `json:"delay_message"`                 // Delay in milliseconds
	DelayTyping       int64        `json:"delay_typing"`                  // Typing indicator in milliseconds
	TextContent       *TextMessage `json:"text_content,omitempty"`        // Text message content (for text messages)
}

// SendMessageResponse represents the response after enqueuing a message
type SendMessageResponse struct {
	ZaapID         string    `json:"zaap_id"`
	Status         string    `json:"status"` // "queued"
	SequenceNumber int64     `json:"sequence_number"`
	ScheduledFor   time.Time `json:"scheduled_for"`
	QueuePosition  int       `json:"queue_position"` // Position in queue (0 = next)
}

// QueueListResponse represents a list of jobs in a queue
type QueueListResponse struct {
	InstanceID uuid.UUID      `json:"instance_id"`
	Total      int            `json:"total"`
	Jobs       []QueueJobInfo `json:"jobs"`
}

// QueueStatsResponse represents statistics for a queue
type QueueStatsResponse struct {
	InstanceID        uuid.UUID `json:"instance_id"`
	QueueName         string    `json:"queue_name"`
	AvailableJobs     int       `json:"available_jobs"`
	RunningJobs       int       `json:"running_jobs"`
	CompletedJobs     int       `json:"completed_jobs"`
	FailedJobs        int       `json:"failed_jobs"`
	TotalProcessed    int       `json:"total_processed"`
	AvgProcessingTime float64   `json:"avg_processing_time_ms"`
}

// LastScheduledTime represents the last scheduled time for an instance
type LastScheduledTime struct {
	InstanceID  uuid.UUID `json:"instance_id"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

// MarshalJSON implements custom JSON marshaling for SendMessageArgs
// This ensures proper serialization for PostgreSQL JSONB storage
func (s SendMessageArgs) MarshalJSON() ([]byte, error) {
	type Alias SendMessageArgs
	return json.Marshal(&struct {
		*Alias
		ScheduledFor string `json:"scheduled_for"`
		EnqueuedAt   string `json:"enqueued_at"`
	}{
		Alias:        (*Alias)(&s),
		ScheduledFor: s.ScheduledFor.Format(time.RFC3339Nano),
		EnqueuedAt:   s.EnqueuedAt.Format(time.RFC3339Nano),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for SendMessageArgs
func (s *SendMessageArgs) UnmarshalJSON(data []byte) error {
	type Alias SendMessageArgs
	aux := &struct {
		*Alias
		ScheduledFor string `json:"scheduled_for"`
		EnqueuedAt   string `json:"enqueued_at"`
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if aux.ScheduledFor != "" {
		scheduledFor, err := time.Parse(time.RFC3339Nano, aux.ScheduledFor)
		if err != nil {
			return err
		}
		s.ScheduledFor = scheduledFor
	}

	if aux.EnqueuedAt != "" {
		enqueuedAt, err := time.Parse(time.RFC3339Nano, aux.EnqueuedAt)
		if err != nil {
			return err
		}
		s.EnqueuedAt = enqueuedAt
	}

	return nil
}

// Validate checks if the SendMessageArgs is valid
func (s *SendMessageArgs) Validate() error {
	if s.ZaapID == "" {
		return ErrInvalidZaapID
	}

	if s.InstanceID == uuid.Nil {
		return ErrInvalidInstanceID
	}

	if s.Phone == "" {
		return ErrInvalidPhone
	}

	// Validate that exactly one message content is set
	contentCount := 0
	if s.TextContent != nil {
		contentCount++
	}
	if s.ImageContent != nil {
		contentCount++
	}
	if s.AudioContent != nil {
		contentCount++
	}
	if s.VideoContent != nil {
		contentCount++
	}
	if s.DocumentContent != nil {
		contentCount++
	}
	if s.StickerContent != nil {
		contentCount++
	}
	if s.PTVContent != nil {
		contentCount++
	}
	if s.LocationContent != nil {
		contentCount++
	}
	if s.ContactContent != nil {
		contentCount++
	}
	if len(s.ContactsContent) > 0 {
		contentCount++
	}
	if s.InteractiveContent != nil {
		contentCount++
	}
	if s.PollContent != nil {
		contentCount++
	}
	if s.EventContent != nil {
		contentCount++
	}

	if contentCount == 0 {
		return ErrNoMessageContent
	}
	if contentCount > 1 {
		return ErrMultipleMessageContents
	}

	return nil
}

// GetContentPreview returns a short preview of the message content for logging
func (s *SendMessageArgs) GetContentPreview() string {
	switch s.MessageType {
	case MessageTypeText:
		if s.TextContent != nil {
			msg := s.TextContent.Message
			if len(msg) > 50 {
				return msg[:50] + "..."
			}
			return msg
		}
	case MessageTypeImage:
		return "[Image]"
	case MessageTypeAudio:
		return "[Audio]"
	case MessageTypeVideo:
		return "[Video]"
	case MessageTypeDocument:
		return "[Document]"
	case MessageTypeSticker:
		return "[Sticker]"
	case MessageTypePTV:
		return "[PTV Video]"
	case MessageTypeLocation:
		return "[Location]"
	case MessageTypeContact:
		if len(s.ContactsContent) > 1 {
			return fmt.Sprintf("[%d Contacts]", len(s.ContactsContent))
		}
		return "[Contact]"
	case MessageTypeInteractive:
		return "[Interactive]"
	case MessageTypeButtonList:
		return "[Button List]"
	case MessageTypeButtonActions:
		return "[Button Actions]"
	case MessageTypeOptionList:
		return "[Option List]"
	case MessageTypeButtonPIX:
		return "[Button PIX]"
	case MessageTypeButtonOTP:
		return "[Button OTP]"
	case MessageTypeCarousel:
		if s.InteractiveContent != nil {
			return fmt.Sprintf("[Carousel] %d cards", len(s.InteractiveContent.CarouselCards))
		}
		return "[Carousel]"
	case MessageTypePoll:
		if s.PollContent != nil {
			return fmt.Sprintf("[Poll] %s", s.PollContent.Question)
		}
	case MessageTypeEvent:
		if s.EventContent != nil {
			return fmt.Sprintf("[Event] %s", s.EventContent.Name)
		}
	}
	return "[Unknown]"
}

// Common errors
var (
	ErrInvalidZaapID           = errors.New("invalid zaap_id")
	ErrInvalidInstanceID       = errors.New("invalid instance_id")
	ErrInvalidPhone            = errors.New("invalid phone number")
	ErrNoMessageContent        = errors.New("no message content provided")
	ErrMultipleMessageContents = errors.New("multiple message contents provided")
	ErrInstanceNotConnected    = errors.New("instance not connected")
	ErrInstanceNotFound        = errors.New("instance not found")
	ErrQueueNotFound           = errors.New("queue not found")
)
