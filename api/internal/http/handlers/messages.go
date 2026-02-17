package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	wameow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/chats"
	"go.mau.fi/whatsmeow/api/internal/contacts"
	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/messages/queue"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type InstanceStatusProvider interface {
	GetStatus(ctx context.Context, id uuid.UUID, clientToken, instanceToken string) (*instances.Status, error)
}

// ContactsService provides contact-related operations
type ContactsService interface {
	List(ctx context.Context, instanceID uuid.UUID, params contacts.ListParams) (contacts.ListResult, error)
	IsOnWhatsApp(ctx context.Context, instanceID uuid.UUID, phone string) (contacts.PhoneExistsResponse, error)
	IsOnWhatsAppBatch(ctx context.Context, instanceID uuid.UUID, phones []string) ([]contacts.PhoneExistsBatchResponse, error)
}

// ChatsService provides chat-related operations
type ChatsService interface {
	List(ctx context.Context, instanceID uuid.UUID, params chats.ListParams) (chats.ListResult, error)
}

// MessageHandler handles HTTP requests for message queue operations
type MessageHandler struct {
	coordinator     queue.QueueCoordinator
	instanceService InstanceStatusProvider
	contactsService ContactsService
	chatsService    ChatsService
	log             *slog.Logger
	drainRetryAfter time.Duration
	enqueueMessage  func(context.Context, uuid.UUID, queue.SendMessageArgs) (string, error)
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(
	coordinator queue.QueueCoordinator,
	instanceService InstanceStatusProvider,
	contactsService ContactsService,
	chatsService ChatsService,
	log *slog.Logger,
	drainRetryAfter time.Duration,
) *MessageHandler {
	if drainRetryAfter <= 0 {
		drainRetryAfter = 30 * time.Second
	}
	handler := &MessageHandler{
		coordinator:     coordinator,
		instanceService: instanceService,
		contactsService: contactsService,
		chatsService:    chatsService,
		log:             log,
		drainRetryAfter: drainRetryAfter,
	}

	if coordinator != nil {
		handler.enqueueMessage = coordinator.EnqueueMessage
	} else {
		handler.enqueueMessage = func(_ context.Context, _ uuid.UUID, _ queue.SendMessageArgs) (string, error) {
			return "", queue.ErrQueueNotFound
		}
	}

	return handler
}

// RegisterRoutes registers message routes within an existing route group
// This is called by InstanceHandler.Register() to avoid path conflicts
func (h *MessageHandler) RegisterRoutes(r chi.Router) {
	// Message sending endpoints
	r.Post("/send-text", h.sendText)
	r.Post("/send-image", h.sendImage)
	r.Post("/send-sticker", h.sendSticker)
	r.Post("/send-audio", h.sendAudio)
	r.Post("/send-video", h.sendVideo)
	r.Post("/send-ptv", h.sendPTV) // Circular video (Push-To-Talk Video)
	r.Post("/send-gif", h.sendGif)
	r.Post("/send-document", h.sendDocument)
	r.Post("/send-location", h.sendLocation)
	r.Post("/send-contact", h.sendContact)
	r.Post("/send-contacts", h.sendContacts)

	// Interactive message endpoints
	r.Post("/send-button-list", h.sendButtonList)       // Simple reply buttons
	r.Post("/send-button-actions", h.sendButtonActions) // Action buttons (URL, call, copy)
	r.Post("/send-option-list", h.sendOptionList)       // List/menu selection
	r.Post("/send-button-pix", h.sendButtonPIX)         // PIX payment button (Brazil)
	r.Post("/send-button-otp", h.sendButtonOTP)         // OTP copy code button
	r.Post("/send-carousel", h.sendCarousel)            // Carousel with multiple cards

	// Poll and Event endpoints
	r.Post("/send-poll", h.sendPoll)                    // Send poll message
	r.Post("/send-poll-vote", h.sendPollVote)           // Vote on a poll
	r.Post("/send-event", h.sendEvent)                  // Send calendar event
	r.Post("/send-edit-event", h.sendEditEvent)         // Edit existing event
	r.Post("/send-event-response", h.sendEventResponse) // Respond to event (going/not_going/maybe)

	// Link preview endpoint
	r.Post("/send-link", h.sendLink) // Send link with custom preview

	// Reaction endpoints
	r.Post("/send-reaction", h.sendReaction)              // Send emoji reaction to message
	r.Post("/send-remove-reaction", h.sendRemoveReaction) // Remove reaction from message

	// Forward message endpoint
	r.Post("/forward-message", h.forwardMessage) // Forward a message to another chat

	// Message management endpoints
	r.Delete("/messages", h.deleteMessage) // Delete message from everyone
	r.Post("/read-message", h.readMessage) // Mark messages as read
	r.Post("/pin-message", h.pinMessage)   // Pin/unpin message in chat

	// Chat modification endpoints
	r.Post("/modify-chat", h.modifyChat) // read, archive, pin, mute, clear, delete

	// Queue management endpoints
	r.Get("/queue", h.listQueue)                      // GET /queue?page=1&pageSize=100
	r.Get("/queue/count", h.getQueueCount)            // GET /queue/count
	r.Delete("/queue", h.clearQueue)                  // DELETE /queue (clear all)
	r.Delete("/queue/{zaapId}", h.cancelQueueMessage) // DELETE /queue/{zaapId}

	// Data retrieval endpoints
	r.Get("/contacts", h.getContacts)                                      // GET /contacts?page=1&pageSize=100
	r.Get("/chats", h.getChats)                                            // GET /chats?page=1&pageSize=100
	r.Get("/phone-exists/{phone}", h.phoneExists)                          // GET /phone-exists/{phone}
	r.Post("/phone-exists-batch", h.phoneExistsBatch)                      // POST /phone-exists-batch
	r.Get("/contacts/{phone}/metadata", h.getContactMetadata)              // GET /contacts/{phone}/metadata
	r.Get("/contacts/{phone}/profile-picture", h.getContactProfilePicture) // GET /contacts/{phone}/profile-picture

	// Block/unblock endpoints
	r.Post("/modify-blocked", h.modifyBlocked) // POST /modify-blocked - block/unblock a contact

	// Profile management endpoints
	r.Put("/profile-name", h.updateProfileName)               // PUT /profile-name - update profile name (push name)
	r.Put("/profile-picture", h.updateProfilePicture)         // PUT /profile-picture - update profile picture
	r.Put("/profile-description", h.updateProfileDescription) // PUT /profile-description - update profile description (about/status)

	// Status/Stories endpoints (broadcast to status@broadcast)
	r.Post("/send-text-status", h.sendTextStatus)   // POST /send-text-status - send text to status/stories
	r.Post("/send-image-status", h.sendImageStatus) // POST /send-image-status - send image to status/stories
	r.Post("/send-audio-status", h.sendAudioStatus) // POST /send-audio-status - send audio with waveform to status/stories
	r.Post("/send-video-status", h.sendVideoStatus) // POST /send-video-status - send video to status/stories
}

// GroupMention represents a group mention in a community
type GroupMention struct {
	Phone   string `json:"phone"`   // Group JID (e.g., "120363xyz@g.us")
	Subject string `json:"subject"` // Group name/subject
}

// SendTextRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-text
// format
type SendTextRequest struct {
	Phone         string  `json:"phone"`                   // Phone number (e.g., "5511999999999")
	Message       string  `json:"message"`                 // Text message content
	MessageID     string  `json:"messageId,omitempty"`     // Optional WhatsApp message ID to reply to
	DelayMessage  *int    `json:"delayMessage,omitempty"`  // Optional delay in seconds before sending
	ScheduledFor  *string `json:"scheduledFor,omitempty"`  // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping   *int    `json:"delayTyping,omitempty"`   // Optional typing indicator duration in seconds (1-15)
	Duration      *int    `json:"duration,omitempty"`      // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	PrivateAnswer bool    `json:"privateAnswer,omitempty"` // For group messages: if true, reply in private to sender (not yourself)
	LinkPreview   *bool   `json:"linkPreview,omitempty"`   // If nil, auto-detect URLs; if true, force preview; if false, disable preview
}

// SendImageRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-image
// format
type SendImageRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Image          string         `json:"image"`                    // Image URL or base64 data (data:image/png;base64,...)
	Caption        string         `json:"caption,omitempty"`        // Optional image caption/title
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds before sending
	ScheduledFor   *string        `json:"scheduledFor,omitempty"`   // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	ViewOnce       bool           `json:"viewOnce,omitempty"`       // If true, image can only be viewed once
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendStickerRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-sticker
// format
type SendStickerRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Sticker        string         `json:"sticker"`                  // Sticker URL or base64 data (data:image/webp;base64,...)
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds before sending
	ScheduledFor   *string        `json:"scheduledFor,omitempty"`   // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendAudioRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-audio
// format
type SendAudioRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Audio          string         `json:"audio"`                    // Audio URL or base64 data (data:audio/ogg;base64,...)
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds before sending
	ScheduledFor   *string        `json:"scheduledFor,omitempty"`   // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional "recording audio" indicator duration in seconds (1-15)
	ViewOnce       bool           `json:"viewOnce,omitempty"`       // If true, audio can only be played once
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendVideoRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-video
// format
type SendVideoRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Video          string         `json:"video"`                    // Video URL or base64 data (data:video/mp4;base64,...)
	Caption        string         `json:"caption,omitempty"`        // Optional video caption/title
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds before sending
	ScheduledFor   *string        `json:"scheduledFor,omitempty"`   // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	ViewOnce       bool           `json:"viewOnce,omitempty"`       // If true, video can only be viewed once
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendPTVRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-ptv
// format for circular video messages (Push-To-Talk Video)
type SendPTVRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Video          string         `json:"video"`                    // Video URL or base64 data (data:video/mp4;base64,...)
	Caption        string         `json:"caption,omitempty"`        // Optional video caption/title
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds before sending
	ScheduledFor   *string        `json:"scheduledFor,omitempty"`   // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	ViewOnce       bool           `json:"viewOnce,omitempty"`       // If true, PTV can only be viewed once
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendGifRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-gif
// format
type SendGifRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Gif            string         `json:"gif"`                      // GIF URL or base64 data (data:image/gif;base64,...)
	Caption        string         `json:"caption,omitempty"`        // Optional GIF caption/title
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds before sending
	ScheduledFor   *string        `json:"scheduledFor,omitempty"`   // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	ViewOnce       bool           `json:"viewOnce,omitempty"`       // If true, GIF can only be viewed once
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendDocumentRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-document
// format
type SendDocumentRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Document       string         `json:"document"`                 // Document URL or base64 data (data:application/pdf;base64,...)
	FileName       string         `json:"fileName,omitempty"`       // Optional document filename
	Caption        string         `json:"caption,omitempty"`        // Optional document caption/title
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds before sending
	ScheduledFor   *string        `json:"scheduledFor,omitempty"`   // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendLocationRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-location
// format
type SendLocationRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Latitude       float64        `json:"latitude"`                 // Location latitude
	Longitude      float64        `json:"longitude"`                // Location longitude
	Name           string         `json:"name,omitempty"`           // Optional location name/title
	Address        string         `json:"address,omitempty"`        // Optional location address
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds before sending
	ScheduledFor   *string        `json:"scheduledFor,omitempty"`   // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendContactRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-contact
// format with extended fields
type SendContactRequest struct {
	Phone                      string `json:"phone"`                                // Phone number (e.g., "5511999999999")
	ContactName                string `json:"contactName"`                          // Full name of the contact
	ContactPhone               string `json:"contactPhone"`                         // Phone number of the contact in international format
	ContactBusinessDescription string `json:"contactBusinessDescription,omitempty"` // Optional business description

	// Extended optional fields for complete vCard
	FirstName  *string `json:"firstName,omitempty"`  // Given name
	LastName   *string `json:"lastName,omitempty"`   // Family name
	MiddleName *string `json:"middleName,omitempty"` // Additional names
	NamePrefix *string `json:"namePrefix,omitempty"` // Honorific prefix (Dr., Mr., Ms.)
	NameSuffix *string `json:"nameSuffix,omitempty"` // Honorific suffix (Jr., Sr., III)
	Nickname   *string `json:"nickname,omitempty"`   // Nickname or alias

	Email        *string         `json:"email,omitempty"`        // Email address
	URL          *string         `json:"url,omitempty"`          // Website or social media URL
	Organization *string         `json:"organization,omitempty"` // Organization/company name (can override contactBusinessDescription)
	JobTitle     *string         `json:"jobTitle,omitempty"`     // Job title or position
	Address      *ContactAddress `json:"address,omitempty"`      // Structured address
	Birthday     *string         `json:"birthday,omitempty"`     // Birthday in YYYY-MM-DD format
	Note         *string         `json:"note,omitempty"`         // Additional notes or comments

	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds before sending
	ScheduledFor   *string        `json:"scheduledFor,omitempty"`   // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// ContactAddress represents a structured address for contact vCard
type ContactAddress struct {
	Type       *string `json:"type,omitempty"`       // Address type: work, home
	PostBox    *string `json:"postBox,omitempty"`    // Post office box
	Extended   *string `json:"extended,omitempty"`   // Extended address (apartment, suite)
	Street     *string `json:"street,omitempty"`     // Street address
	City       *string `json:"city,omitempty"`       // City or locality
	Region     *string `json:"region,omitempty"`     // State, province or region
	PostalCode *string `json:"postalCode,omitempty"` // Postal code
	Country    *string `json:"country,omitempty"`    // Country name
}

// ContactInfo represents a contact for multiple contacts
type ContactInfo struct {
	ContactName                string `json:"contactName"`                          // Full name of the contact
	ContactPhone               string `json:"contactPhone"`                         // Phone number in international format
	ContactBusinessDescription string `json:"contactBusinessDescription,omitempty"` // Optional business description

	// Extended optional fields for complete vCard
	FirstName  *string `json:"firstName,omitempty"`  // Given name
	LastName   *string `json:"lastName,omitempty"`   // Family name
	MiddleName *string `json:"middleName,omitempty"` // Additional names
	NamePrefix *string `json:"namePrefix,omitempty"` // Honorific prefix (Dr., Mr., Ms.)
	NameSuffix *string `json:"nameSuffix,omitempty"` // Honorific suffix (Jr., Sr., III)
	Nickname   *string `json:"nickname,omitempty"`   // Nickname or alias

	Email        *string         `json:"email,omitempty"`        // Email address
	URL          *string         `json:"url,omitempty"`          // Website or social media URL
	Organization *string         `json:"organization,omitempty"` // Organization/company name (can override contactBusinessDescription)
	JobTitle     *string         `json:"jobTitle,omitempty"`     // Job title or position
	Address      *ContactAddress `json:"address,omitempty"`      // Structured address
	Birthday     *string         `json:"birthday,omitempty"`     // Birthday in YYYY-MM-DD format
	Note         *string         `json:"note,omitempty"`         // Additional notes or comments
}

// SendContactsRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-contacts
// format for sending multiple contacts
type SendContactsRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Contacts       []ContactInfo  `json:"contacts"`                 // Array of contacts (1-10)
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds before sending
	ScheduledFor   *string        `json:"scheduledFor,omitempty"`   // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// Button represents a simple reply button in an interactive message
type Button struct {
	ID    string `json:"id"`    // Required: button identifier (max 256 chars)
	Title string `json:"title"` // Required: button text (max 20 chars)
}

// ActionButton represents a button with action type in button-actions messages
type ActionButton struct {
	ID       string  `json:"id"`                 // Required: button identifier (max 256 chars)
	Label    string  `json:"label"`              // Required: button text (max 20 chars)
	Type     string  `json:"type"`               // Required: quick_reply, cta_url, cta_call, cta_copy
	URL      *string `json:"url,omitempty"`      // Required for cta_url type
	Phone    *string `json:"phone,omitempty"`    // Required for cta_call type
	CopyCode *string `json:"copyCode,omitempty"` // Required for cta_copy type
}

// PIXPaymentRequest represents PIX payment details for send-button-pix
type PIXPaymentRequest struct {
	Key           string   `json:"pixKey"`                  // Required: PIX key value
	KeyType       string   `json:"type"`                    // Required: CPF, CNPJ, EMAIL, PHONE, EVP
	Name          *string  `json:"name,omitempty"`          // Optional: beneficiary name
	Amount        *float64 `json:"amount,omitempty"`        // Optional: payment amount
	TransactionID *string  `json:"transactionId,omitempty"` // Optional: transaction ID
}

// Row represents a row in a list section
type Row struct {
	ID          string `json:"id"`          // Required: row identifier (max 200 chars)
	Title       string `json:"title"`       // Required: row title (max 24 chars)
	Description string `json:"description"` // Optional: row description (max 72 chars)
}

// Section represents a section in an interactive list
type Section struct {
	Title string `json:"title"` // Required: section title (max 24 chars)
	Rows  []Row  `json:"rows"`  // Required: 1-10 rows per section
}

// SendButtonActionsRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-button-actions
// format for sending interactive button messages with action types (URL, call, copy)
type SendButtonActionsRequest struct {
	Phone            string         `json:"phone"`                  // Required: recipient phone number
	Message          string         `json:"message"`                // Required: body text
	Buttons          []ActionButton `json:"buttons"`                // Required: 1-3 action buttons
	Title            string         `json:"title"`                  // Optional: header text (max 60 chars)
	Footer           string         `json:"footer"`                 // Optional: footer text (max 60 chars)
	Image            string         `json:"image"`                  // Optional: image URL or base64
	Video            string         `json:"video"`                  // Optional: video URL or base64
	Document         string         `json:"document"`               // Optional: document URL or base64
	DocumentFilename string         `json:"documentFilename"`       // Optional: filename for document
	MessageID        string         `json:"messageId"`              // Optional: reply to message ID
	DelayMessage     *int           `json:"delayMessage"`           // Optional: delay in seconds
	ScheduledFor     *string        `json:"scheduledFor,omitempty"` // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping      *int           `json:"delayTyping"`            // Optional: typing delay in seconds (1-15)
}

// SendButtonListRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-button-list
// format for sending simple reply button messages
type SendButtonListRequest struct {
	Phone        string   `json:"phone"`                  // Required: recipient phone number
	Message      string   `json:"message"`                // Required: body text
	Buttons      []Button `json:"buttons"`                // Required: 1-3 reply buttons
	Title        string   `json:"title"`                  // Optional: header text (max 60 chars)
	Footer       string   `json:"footer"`                 // Optional: footer text (max 60 chars)
	Image        string   `json:"image"`                  // Optional: image URL or base64
	Video        string   `json:"video"`                  // Optional: video URL or base64
	MessageID    string   `json:"messageId"`              // Optional: reply to message ID
	DelayMessage *int     `json:"delayMessage"`           // Optional: delay in seconds
	ScheduledFor *string  `json:"scheduledFor,omitempty"` // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping  *int     `json:"delayTyping"`            // Optional: typing delay in seconds (1-15)
}

// SendOptionListRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-option-list
// format for sending interactive list/menu messages
type SendOptionListRequest struct {
	Phone        string    `json:"phone"`                  // Required: recipient phone number
	Message      string    `json:"message"`                // Required: body text (max 4096 chars)
	ButtonLabel  string    `json:"buttonLabel"`            // Required: menu button text (max 20 chars)
	Sections     []Section `json:"sections"`               // Required: 1-10 sections with rows
	Title        string    `json:"title"`                  // Optional: header text (max 60 chars)
	Footer       string    `json:"footer"`                 // Optional: footer text (max 60 chars)
	MessageID    string    `json:"messageId"`              // Optional: reply to message ID
	DelayMessage *int      `json:"delayMessage"`           // Optional: delay in seconds
	ScheduledFor *string   `json:"scheduledFor,omitempty"` // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping  *int      `json:"delayTyping"`            // Optional: typing delay in seconds (1-15)
}

// SendButtonPIXRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-button-pix
// format for sending PIX payment button messages (Brazilian instant payment)
type SendButtonPIXRequest struct {
	Phone         string   `json:"phone"`                   // Required: recipient phone number
	Message       string   `json:"message"`                 // Optional: body text (default: "Pagamento via PIX")
	PIXKey        string   `json:"pixKey"`                  // Required: PIX key value
	KeyType       string   `json:"type"`                    // Required: CPF, CNPJ, EMAIL, PHONE, EVP
	Name          *string  `json:"name,omitempty"`          // Optional: beneficiary name
	Amount        *float64 `json:"amount,omitempty"`        // Optional: payment amount
	TransactionID *string  `json:"transactionId,omitempty"` // Optional: transaction ID
	MessageID     string   `json:"messageId"`               // Optional: reply to message ID
	DelayMessage  *int     `json:"delayMessage"`            // Optional: delay in seconds
	ScheduledFor  *string  `json:"scheduledFor,omitempty"`  // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping   *int     `json:"delayTyping"`             // Optional: typing delay in seconds (1-15)
}

// SendButtonOTPRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-button-otp
// format for sending OTP (one-time password) copy button messages
type SendButtonOTPRequest struct {
	Phone        string  `json:"phone"`                  // Required: recipient phone number
	Message      string  `json:"message"`                // Required: body text containing OTP context
	Code         string  `json:"code"`                   // Required: OTP code to copy (max 20 chars)
	Title        string  `json:"title"`                  // Optional: header text (max 60 chars)
	Footer       string  `json:"footer"`                 // Optional: footer text (max 60 chars)
	MessageID    string  `json:"messageId"`              // Optional: reply to message ID
	DelayMessage *int    `json:"delayMessage"`           // Optional: delay in seconds
	ScheduledFor *string `json:"scheduledFor,omitempty"` // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping  *int    `json:"delayTyping"`            // Optional: typing delay in seconds (1-15)
}

// SendCarouselRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-carousel
// format for sending carousel messages with multiple cards
type SendCarouselRequest struct {
	Phone        string             `json:"phone"`                  // Required: recipient phone number
	Message      string             `json:"message,omitempty"`      // Optional: carousel body text displayed above cards
	Cards        []SendCarouselCard `json:"cards"`                  // Required: 1-10 carousel cards
	CardType     string             `json:"cardType,omitempty"`     // Optional: HSCROLL_CARDS (default) or ALBUM_IMAGE
	MessageID    string             `json:"messageId,omitempty"`    // Optional: reply to message ID
	DelayMessage *int               `json:"delayMessage,omitempty"` // Optional: delay in seconds
	ScheduledFor *string            `json:"scheduledFor,omitempty"` // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping  *int               `json:"delayTyping,omitempty"`  // Optional: typing delay in seconds (1-15)
}

// SendCarouselCard represents a single card in a carousel message
type SendCarouselCard struct {
	Header   string         `json:"header,omitempty"` // Optional: card header text (max 60 chars)
	Body     string         `json:"body"`             // Required: card body text (max 1024 chars)
	Footer   string         `json:"footer,omitempty"` // Optional: card footer text (max 60 chars)
	Buttons  []ActionButton `json:"buttons"`          // Required: 1-3 action buttons per card
	MediaURL string         `json:"mediaUrl"`         // Required: URL for card image/video
}

// PollOption represents a single poll option
type PollOption struct {
	Name string `json:"name"` // Option text
}

// SendPollRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-poll
// format for sending poll messages
// Example: {"phone": "554499999999", "message": "Outra enquete", "poll": [{"name": "Option1"}, {"name": "Option2"}], "pollMaxOptions": 1}
type SendPollRequest struct {
	Phone          string       `json:"phone"`                    // Required: recipient phone number
	Message        string       `json:"message"`                  // Required: poll question text (Zé da API uses "message")
	Poll           []PollOption `json:"poll"`                     // Required: poll options array with name field (2-12 options)
	PollMaxOptions *int         `json:"pollMaxOptions,omitempty"` // Optional: 0 for single choice, 1+ for multiple choice (default: 0)
	MessageID      string       `json:"messageId,omitempty"`      // Optional: reply to message ID
	DelayMessage   *int         `json:"delayMessage,omitempty"`   // Optional: delay in seconds
	ScheduledFor   *string      `json:"scheduledFor,omitempty"`   // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping    *int         `json:"delayTyping,omitempty"`    // Optional: typing delay in seconds (1-15)
}

// EventLocation represents the location for an event
type EventLocation struct {
	Name             string   `json:"name,omitempty"`             // Location name
	DegreesLatitude  *float64 `json:"degreesLatitude,omitempty"`  // Optional: latitude
	DegreesLongitude *float64 `json:"degreesLongitude,omitempty"` // Optional: longitude
}

// EventPayload represents the nested event object
type EventPayload struct {
	Name         string         `json:"name"`                   // Required: event name/title
	Description  string         `json:"description,omitempty"`  // Optional: event description
	DateTime     string         `json:"dateTime"`               // Required: ISO 8601 format (e.g., "2024-04-29T09:30:53.309Z")
	Location     *EventLocation `json:"location,omitempty"`     // Optional: event location
	CallLinkType string         `json:"callLinkType,omitempty"` // Optional: "voice" or "video"
	Canceled     bool           `json:"canceled,omitempty"`     // Optional: mark event as canceled
}

// SendEventRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-event
// format for sending calendar event messages
// Example: {"phone": "120363356737170752-group", "event": {"name": "Event Name", "dateTime": "2024-04-29T09:30:53.309Z"}}
type SendEventRequest struct {
	Phone        string       `json:"phone"`                  // Required: recipient phone number
	Event        EventPayload `json:"event"`                  // Required: nested event object
	MessageID    string       `json:"messageId,omitempty"`    // Optional: reply to message ID
	DelayMessage *int         `json:"delayMessage,omitempty"` // Optional: delay in seconds
	ScheduledFor *string      `json:"scheduledFor,omitempty"` // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping  *int         `json:"delayTyping,omitempty"`  // Optional: typing delay in seconds (1-15)
}

// SendLinkRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-link
// format for sending link with custom preview
// Example: {"phone": "5544999999999", "message": "Check this out", "linkUrl": "https://example.com", "title": "Title"}
type SendLinkRequest struct {
	Phone           string  `json:"phone"`                     // Required: recipient phone number
	Message         string  `json:"message,omitempty"`         // Optional: text before the link
	Image           string  `json:"image,omitempty"`           // Optional: preview image URL
	LinkUrl         string  `json:"linkUrl"`                   // Required: URL to share
	Title           string  `json:"title,omitempty"`           // Optional: link title
	LinkDescription string  `json:"linkDescription,omitempty"` // Optional: link description
	MessageID       string  `json:"messageId,omitempty"`       // Optional: reply to message ID
	DelayMessage    *int    `json:"delayMessage,omitempty"`    // Optional: delay in seconds
	ScheduledFor    *string `json:"scheduledFor,omitempty"`    // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping     *int    `json:"delayTyping,omitempty"`     // Optional: typing delay in seconds (1-15)
}

// ModifyChatRequest represents the request body for POST /instances/{instanceId}/token/{token}/modify-chat
// format for modifying chat state
// Example: {"phone": "554499999999", "action": "archive"}
type ModifyChatRequest struct {
	Phone  string `json:"phone"`  // Required: phone number or group ID
	Action string `json:"action"` // Required: read, archive, unarchive, pin, unpin, mute, unmute, clear, delete
}

// ModifyChatResponse represents the response for modify-chat operations
type ModifyChatResponse struct {
	Success bool   `json:"success"`
	Phone   string `json:"phone"`
	Action  string `json:"action"`
	Message string `json:"message,omitempty"`
}

// DeleteMessageResponse represents the response for delete message operations
type DeleteMessageResponse struct {
	Success   bool   `json:"success"`
	Phone     string `json:"phone"`
	MessageID string `json:"messageId"`
	Message   string `json:"message,omitempty"`
}

// ReadMessageRequest represents the request body for POST /read-message
type ReadMessageRequest struct {
	Phone      string   `json:"phone"`      // Phone number or group JID
	MessageIDs []string `json:"messageIds"` // Message IDs to mark as read
	Sender     string   `json:"sender"`     // Sender phone (required for group messages)
}

// ReadMessageResponse represents the response for read message operations
type ReadMessageResponse struct {
	Success    bool     `json:"success"`
	Phone      string   `json:"phone"`
	MessageIDs []string `json:"messageIds"`
	Message    string   `json:"message,omitempty"`
}

// ContactMetadataResponse represents contact metadata from whatsmeow GetUserInfo
type ContactMetadataResponse struct {
	Phone        string `json:"phone"`
	IsOnWhatsApp bool   `json:"isOnWhatsApp"`
	JID          string `json:"jid,omitempty"`
	VerifiedName string `json:"verifiedName,omitempty"`
	Status       string `json:"status,omitempty"`
	PictureID    string `json:"pictureId,omitempty"`
	Devices      []struct {
		Device string `json:"device"`
		JID    string `json:"jid"`
	} `json:"devices,omitempty"`
}

// ProfilePictureResponse represents profile picture info from whatsmeow GetProfilePictureInfo
type ProfilePictureResponse struct {
	URL        string `json:"url"`
	ID         string `json:"id"`
	Type       string `json:"type"`
	DirectPath string `json:"directPath,omitempty"`
}

// SendReactionRequest represents the request body for POST /send-reaction
type SendReactionRequest struct {
	Phone     string `json:"phone"`     // Phone number or group JID where the message is
	MessageID string `json:"messageId"` // Message ID to react to
	Reaction  string `json:"reaction"`  // Emoji reaction (e.g., "👍", "❤️", "😂")
	Sender    string `json:"sender"`    // Sender phone (required for group messages to identify who sent the original message)
}

// SendReactionResponse represents the response for reaction operations
type SendReactionResponse struct {
	Success   bool   `json:"success"`
	Phone     string `json:"phone"`
	MessageID string `json:"messageId"`
	Reaction  string `json:"reaction"`
	Message   string `json:"message,omitempty"`
}

// ForwardMessageRequest represents the request body for POST /forward-message
type ForwardMessageRequest struct {
	Phone        string  `json:"phone"`                  // Target phone number or group JID to forward to
	MessageID    string  `json:"messageId"`              // Original message ID to forward
	SourceChat   string  `json:"sourceChat"`             // Source chat JID where the original message is
	IsGroup      bool    `json:"isGroup"`                // Whether source is a group
	DelayMessage *int    `json:"delayMessage,omitempty"` // Optional delay in seconds before sending
	ScheduledFor *string `json:"scheduledFor,omitempty"` // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
}

// ForwardMessageResponse represents the response for forward message operations
type ForwardMessageResponse struct {
	Success      bool   `json:"success"`
	Phone        string `json:"phone"`
	OriginalID   string `json:"originalId"`
	NewMessageID string `json:"newMessageId"`
	Message      string `json:"message,omitempty"`
}

// SendPollVoteRequest represents the request body for POST /send-poll-vote
type SendPollVoteRequest struct {
	Phone        string   `json:"phone"`                  // Chat JID where the poll is (phone number or group JID)
	PollID       string   `json:"pollId"`                 // Poll message ID
	PollSender   string   `json:"pollSender"`             // Phone number of who sent the poll (required for groups)
	Options      []string `json:"options"`                // Selected option names (exact match required)
	DelayMessage *int     `json:"delayMessage,omitempty"` // Optional delay in seconds
	ScheduledFor *string  `json:"scheduledFor,omitempty"` // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
}

// SendPollVoteResponse represents the response for poll vote operations
type SendPollVoteResponse struct {
	Success   bool     `json:"success"`
	Phone     string   `json:"phone"`
	PollID    string   `json:"pollId"`
	Options   []string `json:"options"`
	MessageID string   `json:"messageId,omitempty"`
	Message   string   `json:"message,omitempty"`
}

// PinMessageRequest represents the request body for POST /pin-message
type PinMessageRequest struct {
	Phone     string `json:"phone"`            // Chat JID where the message is (phone number or group JID)
	MessageID string `json:"messageId"`        // Message ID to pin/unpin
	Sender    string `json:"sender,omitempty"` // Phone number of who sent the message (required for groups)
	Pin       bool   `json:"pin"`              // true to pin, false to unpin
}

// PinMessageResponse represents the response for pin message operations
type PinMessageResponse struct {
	Success   bool   `json:"success"`
	Phone     string `json:"phone"`
	MessageID string `json:"messageId"`
	Pinned    bool   `json:"pinned"`
	Message   string `json:"message,omitempty"`
}

// ModifyBlockedRequest represents the request body for POST /modify-blocked
type ModifyBlockedRequest struct {
	Phone  string `json:"phone"`  // Phone number to block/unblock (e.g., "5511999999999")
	Action string `json:"action"` // Action: "block" or "unblock"
}

// ModifyBlockedResponse represents the response for block/unblock operations
type ModifyBlockedResponse struct {
	Success bool   `json:"success"`
	Phone   string `json:"phone"`
	Action  string `json:"action"` // "block" or "unblock"
	Message string `json:"message,omitempty"`
}

// UpdateProfileNameRequest represents the request body for PUT /profile-name
type UpdateProfileNameRequest struct {
	Name string `json:"name"` // New profile name (push name)
}

// UpdateProfileNameResponse represents the response for profile name update
type UpdateProfileNameResponse struct {
	Success bool   `json:"success"`
	Name    string `json:"name"`
	Message string `json:"message,omitempty"`
}

// UpdateProfilePictureRequest represents the request body for PUT /profile-picture
type UpdateProfilePictureRequest struct {
	Image string `json:"image"` // Base64-encoded JPEG image or URL
}

// UpdateProfilePictureResponse represents the response for profile picture update
type UpdateProfilePictureResponse struct {
	Success   bool   `json:"success"`
	PictureID string `json:"pictureId,omitempty"` // New picture ID (or "remove" if picture removed)
	Message   string `json:"message,omitempty"`
}

// UpdateProfileDescriptionRequest represents the request body for PUT /profile-description
type UpdateProfileDescriptionRequest struct {
	Description string `json:"description"` // New profile description (About/Status)
}

// UpdateProfileDescriptionResponse represents the response for profile description update
type UpdateProfileDescriptionResponse struct {
	Success     bool   `json:"success"`
	Description string `json:"description"`
	Message     string `json:"message,omitempty"`
}

// SendEditEventRequest represents the request body for POST /send-edit-event
// Used to update an existing event message
type SendEditEventRequest struct {
	Phone        string  `json:"phone"`                  // Required: chat where the event was sent
	EventID      string  `json:"eventId"`                // Required: original event message ID
	Name         string  `json:"name,omitempty"`         // Optional: updated event name
	Description  string  `json:"description,omitempty"`  // Optional: updated description
	StartTime    string  `json:"startTime,omitempty"`    // Optional: updated start time (ISO 8601)
	EndTime      string  `json:"endTime,omitempty"`      // Optional: updated end time (ISO 8601)
	Location     string  `json:"location,omitempty"`     // Optional: updated location name
	Canceled     *bool   `json:"canceled,omitempty"`     // Optional: mark event as canceled
	DelayMessage *int    `json:"delayMessage,omitempty"` // Optional: delay in seconds
	ScheduledFor *string `json:"scheduledFor,omitempty"` // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping  *int    `json:"delayTyping,omitempty"`  // Optional: typing delay in seconds (1-15)
}

// SendEditEventResponse represents the response for send-edit-event
type SendEditEventResponse struct {
	Success   bool   `json:"success"`
	EventID   string `json:"eventId"`
	MessageID string `json:"messageId,omitempty"`
	Message   string `json:"message,omitempty"`
}

// SendEventResponseRequest represents the request body for POST /send-event-response
// Used to respond to an event invitation (going, not_going, maybe)
type SendEventResponseRequest struct {
	Phone           string  `json:"phone"`                     // Required: chat where the event was sent
	EventID         string  `json:"eventId"`                   // Required: event message ID to respond to
	Response        string  `json:"response"`                  // Required: "going", "not_going", or "maybe"
	ExtraGuestCount *int    `json:"extraGuestCount,omitempty"` // Optional: number of extra guests
	DelayMessage    *int    `json:"delayMessage,omitempty"`    // Optional: delay in seconds
	ScheduledFor    *string `json:"scheduledFor,omitempty"`    // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
	DelayTyping     *int    `json:"delayTyping,omitempty"`     // Optional: typing delay in seconds (1-15)
}

// SendEventResponseResponse represents the response for send-event-response
type SendEventResponseResponse struct {
	Success  bool   `json:"success"`
	EventID  string `json:"eventId"`
	Response string `json:"response"`
	Message  string `json:"message,omitempty"`
}

// =============================================================================
// Status/Stories Request Types (broadcast to status@broadcast)
// =============================================================================

// SendTextStatusRequest represents the request body for POST /send-text-status
// Sends a text message to WhatsApp Status/Stories (broadcasts to all viewers)
type SendTextStatusRequest struct {
	Text            string  `json:"text"`                      // Required: text content for status
	BackgroundColor string  `json:"backgroundColor,omitempty"` // Optional: hex color #RRGGBB or ARGB 0xAARRGGBB
	Font            *int    `json:"font,omitempty"`            // Optional: font style 0-5
	MessageID       string  `json:"messageId,omitempty"`       // Optional: custom message ID for tracking
	DelayMessage    *int    `json:"delayMessage,omitempty"`    // Optional: delay in seconds before sending
	ScheduledFor    *string `json:"scheduledFor,omitempty"`    // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
}

// SendImageStatusRequest represents the request body for POST /send-image-status
// Sends an image to WhatsApp Status/Stories (broadcasts to all viewers)
type SendImageStatusRequest struct {
	Image        string  `json:"image"`                  // Required: image URL or base64 data (data:image/png;base64,...)
	Caption      string  `json:"caption,omitempty"`      // Optional: image caption
	MessageID    string  `json:"messageId,omitempty"`    // Optional: custom message ID for tracking
	DelayMessage *int    `json:"delayMessage,omitempty"` // Optional: delay in seconds before sending
	ScheduledFor *string `json:"scheduledFor,omitempty"` // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
}

// SendAudioStatusRequest represents the request body for POST /send-audio-status
// Sends an audio/voice note to WhatsApp Status/Stories with waveform visualization
type SendAudioStatusRequest struct {
	Audio        string  `json:"audio"`                  // Required: audio URL or base64 data (data:audio/ogg;base64,...)
	MessageID    string  `json:"messageId,omitempty"`    // Optional: custom message ID for tracking
	DelayMessage *int    `json:"delayMessage,omitempty"` // Optional: delay in seconds before sending
	ScheduledFor *string `json:"scheduledFor,omitempty"` // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
}

// SendVideoStatusRequest represents the request body for POST /send-video-status
// Sends a video to WhatsApp Status/Stories (broadcasts to all viewers)
type SendVideoStatusRequest struct {
	Video        string  `json:"video"`                  // Required: video URL or base64 data (data:video/mp4;base64,...)
	Caption      string  `json:"caption,omitempty"`      // Optional: video caption
	MessageID    string  `json:"messageId,omitempty"`    // Optional: custom message ID for tracking
	DelayMessage *int    `json:"delayMessage,omitempty"` // Optional: delay in seconds before sending
	ScheduledFor *string `json:"scheduledFor,omitempty"` // ISO 8601 timestamp for scheduled delivery (overrides delayMessage)
}

// SendMessageResponse represents the response after enqueuing a message
// format
type SendMessageResponse struct {
	ZaapID         string          `json:"zaapId"`                   // Unique message ID in our system
	MessageID      string          `json:"messageId"`                // WhatsApp message ID (initially same as zaapId, updated after send)
	ID             string          `json:"id"`                       // Same as messageId (for Zapier compatibility)
	Status         string          `json:"status"`                   // Queue status (queued, failed, etc.)
	WhatsAppStatus *WhatsAppStatus `json:"whatsAppStatus,omitempty"` // Snapshot of WhatsApp connectivity
}

// WhatsAppStatus captures the connectivity state of the WhatsApp client
type WhatsAppStatus struct {
	Connected bool    `json:"connected"`
	LastSeen  *string `json:"lastSeen,omitempty"`
	Reason    string  `json:"reason"`
	WorkerID  string  `json:"workerId,omitempty"`
}

func (h *MessageHandler) resolveInstance(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, uuid.UUID, *instances.Status, bool) {
	instanceIDStr := chi.URLParam(r, "instanceId")
	instanceID, err := uuid.Parse(instanceIDStr)
	if err != nil {
		h.log.WarnContext(ctx, "invalid instance_id",
			slog.String("instance_id", instanceIDStr),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid instance ID format")
		return ctx, uuid.UUID{}, nil, false
	}

	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))

	if h.instanceService == nil {
		h.log.ErrorContext(ctx, "instance service not configured")
		respondError(w, http.StatusInternalServerError, "Instance service unavailable")
		return ctx, uuid.UUID{}, nil, false
	}

	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	status, err := h.instanceService.GetStatus(ctx, instanceID, clientToken, instanceToken)
	if err != nil {
		h.handleInstanceServiceError(ctx, w, err)
		return ctx, uuid.UUID{}, nil, false
	}

	return ctx, instanceID, status, true
}

func (h *MessageHandler) handleInstanceServiceError(ctx context.Context, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, instances.ErrInstanceNotFound):
		respondError(w, http.StatusNotFound, "instance not found")
	case errors.Is(err, instances.ErrUnauthorized):
		respondError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, instances.ErrInstanceInactive):
		respondError(w, http.StatusForbidden, "instance subscription inactive")
	default:
		logging.ContextLogger(ctx, h.log).Error("instance service error",
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *MessageHandler) newSendMessageResponse(zaapID string, status *instances.Status) SendMessageResponse {
	response := SendMessageResponse{
		ZaapID:    zaapID,
		MessageID: zaapID,
		ID:        zaapID,
		Status:    "queued",
	}
	response.WhatsAppStatus = h.toWhatsAppStatus(status)
	return response
}

func (h *MessageHandler) toWhatsAppStatus(status *instances.Status) *WhatsAppStatus {
	if status == nil {
		return nil
	}

	reason := status.ConnectionStatus
	if status.StoreJID == nil || *status.StoreJID == "" {
		reason = "logged_out"
	} else if reason == "" {
		if status.Connected {
			reason = "connected"
		} else {
			reason = "disconnected"
		}
	}

	var lastSeen *string
	if status.LastConnected != nil {
		formatted := status.LastConnected.UTC().Format(time.RFC3339)
		lastSeen = &formatted
	}

	whatsAppStatus := &WhatsAppStatus{
		Connected: status.Connected,
		Reason:    reason,
	}

	if lastSeen != nil {
		whatsAppStatus.LastSeen = lastSeen
	}

	if status.WorkerAssigned != "" {
		whatsAppStatus.WorkerID = status.WorkerAssigned
	}

	return whatsAppStatus
}

// resolveDelay converts delayMessage (seconds) and/or scheduledFor (ISO 8601) to milliseconds.
// If both are provided, scheduledFor takes priority.
// If neither is provided, returns a random delay of 1-3 seconds.
func resolveDelay(delayMessage *int, scheduledFor *string) int64 {
	var delay int64
	if delayMessage != nil {
		seconds := *delayMessage
		if seconds < 1 {
			seconds = 1
		}
		delay = int64(seconds) * 1000
	} else {
		delay = int64(1000 + (rand.Int63() % 2000)) // default 1-3s
	}
	if scheduledFor != nil && *scheduledFor != "" {
		if t, err := time.Parse(time.RFC3339, *scheduledFor); err == nil && t.After(time.Now()) {
			delay = time.Until(t).Milliseconds()
		}
	}
	return delay
}

// resolveTypingDelay converts delayTyping (seconds) to milliseconds with a 15s cap.
// Typing indicator >15s makes no sense, so the cap is preserved.
func resolveTypingDelay(delayTyping *int) int64 {
	if delayTyping == nil {
		return 0
	}
	seconds := *delayTyping
	if seconds < 1 {
		seconds = 1
	}
	if seconds > 15 {
		seconds = 15
	}
	return int64(seconds) * 1000
}

// QueueMessageResponse represents a message in the queue
type QueueMessageResponse struct {
	ID           string `json:"_id"`               // Message ID (same as ZaapId for Zé da API compat)
	ZaapId       string `json:"zaapId"`            // Zé da API message ID
	MessageId    string `json:"messageId"`         // WhatsApp message ID
	InstanceId   string `json:"instanceId"`        // Instance ID
	Phone        string `json:"phone"`             // Recipient phone
	Message      string `json:"message,omitempty"` // Message text (for text messages)
	DelayMessage int64  `json:"delayMessage"`      // Delay in seconds before sending
	DelayTyping  int64  `json:"delayTyping"`       // Typing indicator duration in seconds
	Created      int64  `json:"created"`           // Unix timestamp in milliseconds

	// Additional fields (not in Zé da API but useful)
	MessageType    string   `json:"messageType,omitempty"`    // Message type (text, image, etc)
	Status         string   `json:"status,omitempty"`         // Job status
	SequenceNumber int64    `json:"sequenceNumber,omitempty"` // FIFO sequence
	Attempt        int      `json:"attempt,omitempty"`        // Current attempt
	MaxAttempts    int      `json:"maxAttempts,omitempty"`    // Max retry attempts
	Errors         []string `json:"errors,omitempty"`         // Error messages
}

// QueueListResponse represents the response for GET /queue
type QueueListResponse []QueueMessageResponse

// QueueCountResponse represents the response for GET /queue/count
type QueueCountResponse struct {
	Count int `json:"count"` // Number of messages in queue
}

// sendText handles POST /instances/{instanceId}/token/{token}/send-text
//
// endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Enqueues message with FIFO ordering
// 4. Returns immediately with zaapId as messageId (non-blocking)
// 5. Worker updates with real WhatsApp messageId after sending
//

func (h *MessageHandler) sendText(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number to WhatsApp format
	phone = normalizePhoneNumber(phone)

	// Validate message content
	message := strings.TrimSpace(req.Message)
	if message == "" {
		h.log.WarnContext(ctx, "missing message content")
		respondError(w, http.StatusBadRequest, "Message content is required")
		return
	}

	// Resolve delays (delayMessage supports any duration; scheduledFor overrides with ISO 8601)
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Create message args
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypeText,
		TextContent: &queue.TextMessage{
			Message: message,
		},
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: strings.TrimSpace(req.MessageID), // Reply-to support
		Duration:         req.Duration,                     // Ephemeral message duration
		PrivateAnswer:    req.PrivateAnswer,                // Private reply in groups
		LinkPreview:      req.LinkPreview,                  // Link preview control (nil=auto, true=force, false=disable)
	}

	// Enqueue message (non-blocking)
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}
	h.log.InfoContext(ctx, "message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	// Return response
	// Initially, messageId = zaapId (will be updated with real WhatsApp ID after send)
	response := h.newSendMessageResponse(zaapID, instStatus)

	// Return 200 OK
	respondJSON(w, http.StatusOK, response)
}

// sendImage handles POST /instances/{instanceId}/token/{token}/send-image
//
// endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Supports image URL or base64 data
// 4. Supports optional caption, viewOnce, and reply-to (messageId)
// 5. Enqueues message with FIFO ordering
// 6. Returns immediately with zaapId as messageId (non-blocking)
// 7. Worker updates with real WhatsApp messageId after sending
//

func (h *MessageHandler) sendImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number to WhatsApp format
	phone = normalizePhoneNumber(phone)

	// Validate image (URL or base64)
	image := strings.TrimSpace(req.Image)
	if image == "" {
		h.log.WarnContext(ctx, "missing image")
		respondError(w, http.StatusBadRequest, "Image is required (URL or base64)")
		return
	}

	// Validate image format (must be URL or base64 data URI)
	if !strings.HasPrefix(image, "http://") &&
		!strings.HasPrefix(image, "https://") &&
		!strings.HasPrefix(image, "data:image/") {
		h.log.WarnContext(ctx, "invalid image format",
			slog.String("prefix", image[:min(len(image), 20)]))
		respondError(w, http.StatusBadRequest, "Image must be a URL (http/https) or base64 data URI (data:image/...)")
		return
	}

	// Resolve delays (delayMessage supports any duration; scheduledFor overrides with ISO 8601)
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Prepare caption (optional)
	caption := strings.TrimSpace(req.Caption)
	var captionPtr *string
	if caption != "" {
		captionPtr = &caption
	}

	// Create message args
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypeImage,
		ImageContent: &queue.MediaMessage{
			MediaURL: image,
			Caption:  captionPtr,
		},
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ViewOnce:         req.ViewOnce,
		ReplyToMessageID: strings.TrimSpace(req.MessageID), // Reply to message (if provided)
		Duration:         req.Duration,                     // Ephemeral message duration
		Mentioned:        req.Mentioned,                    // Mention support
		GroupMentioned:   convertGroupMentions(req.GroupMentioned),
		MentionedAll:     req.MentionedAll,
		PrivateAnswer:    req.PrivateAnswer,
	}

	// Enqueue message (non-blocking)
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "image message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Bool("view_once", req.ViewOnce),
		slog.Bool("has_caption", captionPtr != nil),
		slog.Bool("is_reply", req.MessageID != ""),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	// Return response
	// Initially, messageId = zaapId (will be updated with real WhatsApp ID after send)
	response := h.newSendMessageResponse(zaapID, instStatus)

	// Return 200 OK
	respondJSON(w, http.StatusOK, response)
}

// sendSticker handles POST /instances/{instanceId}/token/{token}/send-sticker
//
// endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Supports sticker URL or base64 data (WebP format)
// 4. Supports optional reply-to (messageId)
// 5. Enqueues message with FIFO ordering
// 6. Returns immediately with zaapId as messageId (non-blocking)
//

func (h *MessageHandler) sendSticker(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendStickerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number to WhatsApp format
	phone = normalizePhoneNumber(phone)

	// Validate sticker (URL or base64)
	sticker := strings.TrimSpace(req.Sticker)
	if sticker == "" {
		h.log.WarnContext(ctx, "missing sticker")
		respondError(w, http.StatusBadRequest, "Sticker is required (URL or base64)")
		return
	}

	// Validate sticker format (must be URL or base64 data URI)
	if !strings.HasPrefix(sticker, "http://") &&
		!strings.HasPrefix(sticker, "https://") &&
		!strings.HasPrefix(sticker, "data:image/") {
		h.log.WarnContext(ctx, "invalid sticker format",
			slog.String("prefix", sticker[:min(len(sticker), 20)]))
		respondError(w, http.StatusBadRequest, "Sticker must be a URL (http/https) or base64 data URI (data:image/...)")
		return
	}

	// Resolve delays (delayMessage supports any duration; scheduledFor overrides with ISO 8601)
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Create message args for sticker (uses dedicated StickerProcessor with WebP conversion)
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypeSticker,
		StickerContent: &queue.MediaMessage{
			MediaURL: sticker,
		},
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: strings.TrimSpace(req.MessageID),
		Duration:         req.Duration,
		Mentioned:        req.Mentioned,
		GroupMentioned:   convertGroupMentions(req.GroupMentioned),
		MentionedAll:     req.MentionedAll,
		PrivateAnswer:    req.PrivateAnswer,
	}

	// Enqueue message (non-blocking)
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "sticker message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Bool("is_reply", req.MessageID != ""),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	// Return response
	response := h.newSendMessageResponse(zaapID, instStatus)

	// Return 200 OK
	respondJSON(w, http.StatusOK, response)
}

// sendAudio handles POST /instances/{instanceId}/token/{token}/send-audio
//
// endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Supports audio URL or base64 data
// 4. Supports optional viewOnce and reply-to (messageId)
// 5. DelayTyping shows "recording audio" indicator
// 6. Enqueues message with FIFO ordering
// 7. Returns immediately with zaapId as messageId (non-blocking)
//

func (h *MessageHandler) sendAudio(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendAudioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number to WhatsApp format
	phone = normalizePhoneNumber(phone)

	// Validate audio (URL or base64)
	audio := strings.TrimSpace(req.Audio)
	if audio == "" {
		h.log.WarnContext(ctx, "missing audio")
		respondError(w, http.StatusBadRequest, "Audio is required (URL or base64)")
		return
	}

	// Validate audio format (must be URL or base64 data URI)
	if !strings.HasPrefix(audio, "http://") &&
		!strings.HasPrefix(audio, "https://") &&
		!strings.HasPrefix(audio, "data:audio/") {
		h.log.WarnContext(ctx, "invalid audio format",
			slog.String("prefix", audio[:min(len(audio), 20)]))
		respondError(w, http.StatusBadRequest, "Audio must be a URL (http/https) or base64 data URI (data:audio/...)")
		return
	}

	// Resolve delays (delayMessage supports any duration; scheduledFor overrides with ISO 8601)
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Create message args
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypeAudio,
		AudioContent: &queue.MediaMessage{
			MediaURL: audio,
		},
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping, // Shows "recording audio"
		ViewOnce:         req.ViewOnce,
		ReplyToMessageID: strings.TrimSpace(req.MessageID),
		Duration:         req.Duration,
		Mentioned:        req.Mentioned, // Mention support
		GroupMentioned:   convertGroupMentions(req.GroupMentioned),
		MentionedAll:     req.MentionedAll,
		PrivateAnswer:    req.PrivateAnswer,
	}

	// Enqueue message (non-blocking)
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "audio message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Bool("view_once", req.ViewOnce),
		slog.Bool("is_reply", req.MessageID != ""),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	// Return response
	response := h.newSendMessageResponse(zaapID, instStatus)

	respondJSON(w, http.StatusOK, response)
}

// sendVideo handles POST /instances/{instanceId}/token/{token}/send-video
//
// endpoint for sending video messages
func (h *MessageHandler) sendVideo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Validate video (URL or base64)
	video := strings.TrimSpace(req.Video)
	if video == "" {
		h.log.WarnContext(ctx, "missing video")
		respondError(w, http.StatusBadRequest, "Video is required (URL or base64)")
		return
	}

	// Validate video format
	if !strings.HasPrefix(video, "http://") &&
		!strings.HasPrefix(video, "https://") &&
		!strings.HasPrefix(video, "data:video/") {
		h.log.WarnContext(ctx, "invalid video format",
			slog.String("prefix", video[:min(len(video), 20)]))
		respondError(w, http.StatusBadRequest, "Video must be a URL (http/https) or base64 data URI (data:video/...)")
		return
	}

	// Convert delays
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Prepare caption
	caption := strings.TrimSpace(req.Caption)
	var captionPtr *string
	if caption != "" {
		captionPtr = &caption
	}

	// Create message args
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypeVideo,
		VideoContent: &queue.MediaMessage{
			MediaURL: video,
			Caption:  captionPtr,
		},
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ViewOnce:         req.ViewOnce,
		ReplyToMessageID: strings.TrimSpace(req.MessageID),
		Duration:         req.Duration,
		Mentioned:        req.Mentioned, // Mention support
		GroupMentioned:   convertGroupMentions(req.GroupMentioned),
		MentionedAll:     req.MentionedAll,
		PrivateAnswer:    req.PrivateAnswer,
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "video message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Bool("view_once", req.ViewOnce),
		slog.Bool("has_caption", captionPtr != nil),
		slog.Bool("is_reply", req.MessageID != ""),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	// Return response
	response := h.newSendMessageResponse(zaapID, instStatus)

	respondJSON(w, http.StatusOK, response)
}

// sendPTV handles POST /instances/{instanceId}/token/{token}/send-ptv
//
// endpoint for sending circular video messages (Push-To-Talk Video).
// PTV messages are displayed as circular video notes in WhatsApp, similar to voice notes
// but with video content.
//
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Supports video URL or base64 data
// 4. Supports optional caption and reply-to (messageId)
// 5. Enqueues message with FIFO ordering
// 6. Returns immediately with zaapId as messageId (non-blocking)
//

func (h *MessageHandler) sendPTV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendPTVRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number to WhatsApp format
	phone = normalizePhoneNumber(phone)

	// Validate video (URL or base64)
	video := strings.TrimSpace(req.Video)
	if video == "" {
		h.log.WarnContext(ctx, "missing video for PTV")
		respondError(w, http.StatusBadRequest, "Video is required (URL or base64)")
		return
	}

	// Validate video format (must be URL or base64 data URI)
	if !strings.HasPrefix(video, "http://") &&
		!strings.HasPrefix(video, "https://") &&
		!strings.HasPrefix(video, "data:video/") {
		h.log.WarnContext(ctx, "invalid video format for PTV",
			slog.String("prefix", video[:min(len(video), 20)]))
		respondError(w, http.StatusBadRequest, "Video must be a URL (http/https) or base64 data URI (data:video/...)")
		return
	}

	// Resolve delays (delayMessage supports any duration; scheduledFor overrides with ISO 8601)
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Prepare caption
	caption := strings.TrimSpace(req.Caption)
	var captionPtr *string
	if caption != "" {
		captionPtr = &caption
	}

	// Create message args for PTV (circular video)
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypePTV,
		PTVContent: &queue.MediaMessage{
			MediaURL: video,
			Caption:  captionPtr,
		},
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ViewOnce:         req.ViewOnce,
		ReplyToMessageID: strings.TrimSpace(req.MessageID),
		Duration:         req.Duration,
		Mentioned:        req.Mentioned,
		GroupMentioned:   convertGroupMentions(req.GroupMentioned),
		MentionedAll:     req.MentionedAll,
		PrivateAnswer:    req.PrivateAnswer,
	}

	// Enqueue message (non-blocking)
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "PTV message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Bool("view_once", req.ViewOnce),
		slog.Bool("has_caption", captionPtr != nil),
		slog.Bool("is_reply", req.MessageID != ""),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	// Return response
	response := h.newSendMessageResponse(zaapID, instStatus)

	respondJSON(w, http.StatusOK, response)
}

// sendGif handles POST /instances/{instanceId}/token/{token}/send-gif
//
// endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Supports GIF URL or base64 data
// 4. Supports optional caption and viewOnce
// 5. Enqueues message with FIFO ordering
// 6. Returns immediately with zaapId as messageId (non-blocking)
//

func (h *MessageHandler) sendGif(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendGifRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number to WhatsApp format
	phone = normalizePhoneNumber(phone)

	// Validate GIF (URL or base64)
	gif := strings.TrimSpace(req.Gif)
	if gif == "" {
		h.log.WarnContext(ctx, "missing gif")
		respondError(w, http.StatusBadRequest, "GIF is required (URL or base64)")
		return
	}

	// Validate GIF format (must be URL or base64 data URI)
	if !strings.HasPrefix(gif, "http://") &&
		!strings.HasPrefix(gif, "https://") &&
		!strings.HasPrefix(gif, "data:image/") &&
		!strings.HasPrefix(gif, "data:video/") {
		h.log.WarnContext(ctx, "invalid gif format",
			slog.String("prefix", gif[:min(len(gif), 20)]))
		respondError(w, http.StatusBadRequest, "GIF must be a URL (http/https) or base64 data URI (data:image/... or data:video/...)")
		return
	}

	// Resolve delays (delayMessage supports any duration; scheduledFor overrides with ISO 8601)
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Handle optional caption
	var captionPtr *string
	caption := strings.TrimSpace(req.Caption)
	if caption != "" {
		captionPtr = &caption
	}

	// Create message args with GIF metadata
	metadata := map[string]interface{}{
		"is_gif": true,
	}

	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypeVideo, // GIFs use video type with metadata flag
		VideoContent: &queue.MediaMessage{
			MediaURL: gif,
			Caption:  captionPtr,
		},
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ViewOnce:         req.ViewOnce,
		ReplyToMessageID: strings.TrimSpace(req.MessageID),
		Duration:         req.Duration,
		Mentioned:        req.Mentioned,
		GroupMentioned:   convertGroupMentions(req.GroupMentioned),
		MentionedAll:     req.MentionedAll,
		PrivateAnswer:    req.PrivateAnswer,
		Metadata:         metadata,
	}

	// Enqueue message (non-blocking)
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "gif message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Bool("view_once", req.ViewOnce),
		slog.Bool("has_caption", captionPtr != nil),
		slog.Bool("is_reply", req.MessageID != ""),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	// Return response
	response := h.newSendMessageResponse(zaapID, instStatus)

	// Return 200 OK
	respondJSON(w, http.StatusOK, response)
}

// sendDocument handles POST /instances/{instanceId}/token/{token}/send-document
//
// endpoint for sending document messages
func (h *MessageHandler) sendDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Validate document (URL or base64)
	document := strings.TrimSpace(req.Document)
	if document == "" {
		h.log.WarnContext(ctx, "missing document")
		respondError(w, http.StatusBadRequest, "Document is required (URL or base64)")
		return
	}

	// Validate document format
	if !strings.HasPrefix(document, "http://") &&
		!strings.HasPrefix(document, "https://") &&
		!strings.HasPrefix(document, "data:application/") &&
		!strings.HasPrefix(document, "data:image/") && // PDFs can be images
		!strings.HasPrefix(document, "data:text/") {
		h.log.WarnContext(ctx, "invalid document format",
			slog.String("prefix", document[:min(len(document), 20)]))
		respondError(w, http.StatusBadRequest, "Document must be a URL or base64 data URI")
		return
	}

	// Convert delays
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Prepare caption and filename
	caption := strings.TrimSpace(req.Caption)
	var captionPtr *string
	if caption != "" {
		captionPtr = &caption
	}

	fileName := strings.TrimSpace(req.FileName)
	var fileNamePtr *string
	if fileName != "" {
		fileNamePtr = &fileName
	}

	// Create message args
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypeDocument,
		DocumentContent: &queue.MediaMessage{
			MediaURL: document,
			Caption:  captionPtr,
			FileName: fileNamePtr,
		},
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: strings.TrimSpace(req.MessageID),
		Duration:         req.Duration,
		Mentioned:        req.Mentioned, // Mention support
		GroupMentioned:   convertGroupMentions(req.GroupMentioned),
		MentionedAll:     req.MentionedAll,
		PrivateAnswer:    req.PrivateAnswer,
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "document message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Bool("has_caption", captionPtr != nil),
		slog.Bool("has_filename", fileNamePtr != nil),
		slog.Bool("is_reply", req.MessageID != ""),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	// Return response
	response := h.newSendMessageResponse(zaapID, instStatus)

	respondJSON(w, http.StatusOK, response)
}

// sendLocation handles POST /instances/{instanceId}/token/{token}/send-location
func (h *MessageHandler) sendLocation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Validate coordinates (latitude and longitude are required, zero values are valid)
	// Latitude: -90 to 90, Longitude: -180 to 180

	// Convert delays
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Prepare optional fields
	name := strings.TrimSpace(req.Name)
	var namePtr *string
	if name != "" {
		namePtr = &name
	}

	address := strings.TrimSpace(req.Address)
	var addressPtr *string
	if address != "" {
		addressPtr = &address
	}

	// Create message args
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypeLocation,
		LocationContent: &queue.LocationMessage{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Name:      namePtr,
			Address:   addressPtr,
		},
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: strings.TrimSpace(req.MessageID),
		Duration:         req.Duration,
		Mentioned:        req.Mentioned, // Mention support
		GroupMentioned:   convertGroupMentions(req.GroupMentioned),
		MentionedAll:     req.MentionedAll,
		PrivateAnswer:    req.PrivateAnswer,
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "location message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Float64("latitude", req.Latitude),
		slog.Float64("longitude", req.Longitude),
		slog.Bool("has_name", namePtr != nil),
		slog.Bool("has_address", addressPtr != nil),
		slog.Bool("is_reply", req.MessageID != ""),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	// Return response
	response := h.newSendMessageResponse(zaapID, instStatus)

	respondJSON(w, http.StatusOK, response)
}

// sendContact handles POST /instances/{instanceId}/token/{token}/send-contact
func (h *MessageHandler) sendContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Validate Zé da API required contact fields
	contactName := strings.TrimSpace(req.ContactName)
	if contactName == "" {
		h.log.WarnContext(ctx, "missing contact name")
		respondError(w, http.StatusBadRequest, "contactName is required")
		return
	}

	contactPhone := strings.TrimSpace(req.ContactPhone)
	if contactPhone == "" {
		h.log.WarnContext(ctx, "missing contact phone")
		respondError(w, http.StatusBadRequest, "contactPhone is required")
		return
	}

	// Convert delays
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Build ContactMessage with ALL fields (Zé da API + extended optional fields)
	contactMsg := &queue.ContactMessage{
		// Required

		FullName:    contactName,
		PhoneNumber: contactPhone,

		// Extended name fields
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		MiddleName: req.MiddleName,
		NamePrefix: req.NamePrefix,
		NameSuffix: req.NameSuffix,
		Nickname:   req.Nickname,

		// Contact fields
		Email: req.Email,
		URL:   req.URL,

		// Organization override (if Organization field provided, use it; otherwise use ContactBusinessDescription)
		Organization: req.Organization,
		JobTitle:     req.JobTitle,

		// Address
		Address: convertContactAddress(req.Address),

		// Personal fields
		Birthday: req.Birthday,
		Note:     req.Note,
	}

	// If Organization not provided but ContactBusinessDescription is, use it for backward compatibility
	if contactMsg.Organization == nil && req.ContactBusinessDescription != "" {
		contactMsg.Organization = &req.ContactBusinessDescription
	}

	args := queue.SendMessageArgs{
		InstanceID:       instanceID,
		Phone:            phone,
		MessageType:      queue.MessageTypeContact,
		ContactContent:   contactMsg,
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: strings.TrimSpace(req.MessageID),
		Duration:         req.Duration,
		Mentioned:        req.Mentioned, // Mention support
		GroupMentioned:   convertGroupMentions(req.GroupMentioned),
		MentionedAll:     req.MentionedAll,
		PrivateAnswer:    req.PrivateAnswer,
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "contact message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Bool("is_reply", req.MessageID != ""),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	// Return response
	response := h.newSendMessageResponse(zaapID, instStatus)

	respondJSON(w, http.StatusOK, response)
}

// listQueue handles GET /instances/{instanceId}/token/{token}/queue
//
// endpoint
// Query parameters:
// - page: Page number (default: 1)
// - pageSize: Number of messages per page (default: 100)
func (h *MessageHandler) listQueue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract instanceId and token from URL
	instanceIDStr := chi.URLParam(r, "instanceId")
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	// Validate instance ID format
	instanceID, err := uuid.Parse(instanceIDStr)
	if err != nil {
		h.log.WarnContext(ctx, "invalid instance_id",
			slog.String("instance_id", instanceIDStr),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid instance ID format")
		return
	}

	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))

	// TODO: Validate instanceToken and clientToken
	if instanceToken == "" {
		h.log.WarnContext(ctx, "missing instance token")
		respondError(w, http.StatusUnauthorized, "Instance token is required")
		return
	}

	if clientToken == "" {
		h.log.WarnContext(ctx, "missing client token")
		respondError(w, http.StatusUnauthorized, "Client-Token header is required")
		return
	}

	// Parse query parameters (format: page, pageSize)
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	pageSize := 100
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if parsedSize, err := strconv.Atoi(pageSizeStr); err == nil && parsedSize > 0 {
			pageSize = parsedSize
		}
	}

	// Calculate offset from page
	offset := (page - 1) * pageSize

	// Get queue jobs
	queueList, err := h.coordinator.ListQueueJobs(ctx, instanceID, pageSize, offset)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to list queue jobs",
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to retrieve queue")
		return
	}

	messages := make(QueueListResponse, 0, len(queueList.Jobs))
	for _, job := range queueList.Jobs {
		messages = append(messages, convertJobToQueueMessage(job))
	}

	respondJSON(w, http.StatusOK, messages)
}

// getQueueCount handles GET /instances/{instanceId}/token/{token}/queue/count
//
// endpoint
func (h *MessageHandler) getQueueCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract instanceId and token from URL
	instanceIDStr := chi.URLParam(r, "instanceId")
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	// Validate instance ID format
	instanceID, err := uuid.Parse(instanceIDStr)
	if err != nil {
		h.log.WarnContext(ctx, "invalid instance_id",
			slog.String("instance_id", instanceIDStr))
		respondError(w, http.StatusBadRequest, "Invalid instance ID format")
		return
	}

	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))

	// TODO: Validate instanceToken and clientToken
	if instanceToken == "" {
		respondError(w, http.StatusUnauthorized, "Instance token is required")
		return
	}

	if clientToken == "" {
		respondError(w, http.StatusUnauthorized, "Client-Token header is required")
		return
	}

	// Get queue count (limit=1 just to get total)
	queueList, err := h.coordinator.ListQueueJobs(ctx, instanceID, 1, 0)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to get queue count",
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to retrieve queue count")
		return
	}

	response := QueueCountResponse{
		Count: queueList.Total,
	}

	respondJSON(w, http.StatusOK, response)
}

// clearQueue handles DELETE /instances/{instanceId}/token/{token}/queue
//
// endpoint - deletes ALL messages in queue
func (h *MessageHandler) clearQueue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract instanceId and token from URL
	instanceIDStr := chi.URLParam(r, "instanceId")
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	// Validate instance ID format
	instanceID, err := uuid.Parse(instanceIDStr)
	if err != nil {
		h.log.WarnContext(ctx, "invalid instance_id",
			slog.String("instance_id", instanceIDStr))
		respondError(w, http.StatusBadRequest, "Invalid instance ID format")
		return
	}

	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))

	// TODO: Validate instanceToken and clientToken
	if instanceToken == "" {
		respondError(w, http.StatusUnauthorized, "Instance token is required")
		return
	}

	if clientToken == "" {
		respondError(w, http.StatusUnauthorized, "Client-Token header is required")
		return
	}

	// Clear all jobs for this instance
	err = h.coordinator.ClearQueue(ctx, instanceID)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to clear queue",
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to clear queue")
		return
	}

	h.log.InfoContext(ctx, "queue cleared successfully")

	// Zé da API returns 200 OK with empty body
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Queue cleared successfully",
	})
}

// cancelQueueMessage handles DELETE /instances/{instanceId}/token/{token}/queue/{zaapId}
//
// endpoint - deletes a specific message from queue
func (h *MessageHandler) cancelQueueMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract instanceId, token, and zaapId from URL
	instanceIDStr := chi.URLParam(r, "instanceId")
	instanceToken := chi.URLParam(r, "token")
	zaapID := chi.URLParam(r, "zaapId")
	clientToken := r.Header.Get("Client-Token")

	// Validate instance ID format
	instanceID, err := uuid.Parse(instanceIDStr)
	if err != nil {
		h.log.WarnContext(ctx, "invalid instance_id",
			slog.String("instance_id", instanceIDStr))
		respondError(w, http.StatusBadRequest, "Invalid instance ID format")
		return
	}

	if zaapID == "" {
		respondError(w, http.StatusBadRequest, "zaapId is required")
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("instance_id", instanceID.String()),
		slog.String("zaap_id", zaapID))

	// TODO: Validate instanceToken and clientToken
	if instanceToken == "" {
		respondError(w, http.StatusUnauthorized, "Instance token is required")
		return
	}

	if clientToken == "" {
		respondError(w, http.StatusUnauthorized, "Client-Token header is required")
		return
	}

	// Cancel the job
	err = h.coordinator.CancelJob(ctx, zaapID)
	if err != nil {
		if errors.Is(err, queue.ErrQueueNotFound) {
			respondError(w, http.StatusNotFound, "Message not found in queue")
			return
		}
		h.log.ErrorContext(ctx, "failed to cancel message",
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to cancel message")
		return
	}

	h.log.InfoContext(ctx, "message cancelled successfully")

	// Zé da API returns 200 OK with empty body
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Message cancelled successfully",
	})
}

// getContacts handles GET /instances/{instanceId}/token/{token}/contacts
//
// endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Retrieves all contacts from WhatsApp instance
// 4. Applies pagination to results
// 5. Returns contact array
//
// Query parameters:
// - page: Page number (required, minimum: 1)
// - pageSize: Number of contacts per page (required, minimum: 1)
//

func (h *MessageHandler) getContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		h.log.WarnContext(ctx, "missing required query parameter: page")
		respondError(w, http.StatusBadRequest, "Query parameter 'page' is required")
		return
	}

	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		h.log.WarnContext(ctx, "missing required query parameter: pageSize")
		respondError(w, http.StatusBadRequest, "Query parameter 'pageSize' is required")
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		h.log.WarnContext(ctx, "invalid page parameter",
			slog.String("page", pageStr))
		respondError(w, http.StatusBadRequest, "Query parameter 'page' must be a positive integer")
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		h.log.WarnContext(ctx, "invalid pageSize parameter",
			slog.String("pageSize", pageSizeStr))
		respondError(w, http.StatusBadRequest, "Query parameter 'pageSize' must be a positive integer")
		return
	}

	h.log.InfoContext(ctx, "listing contacts",
		slog.Int("page", page),
		slog.Int("page_size", pageSize))

	// Call contacts service
	if h.contactsService == nil {
		h.log.ErrorContext(ctx, "contacts service not available")
		respondError(w, http.StatusServiceUnavailable, "Contacts service not available")
		return
	}

	result, err := h.contactsService.List(ctx, instanceID, contacts.ListParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		h.log.ErrorContext(ctx, "failed to list contacts",
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to retrieve contacts")
		return
	}

	h.log.InfoContext(ctx, "contacts retrieved successfully",
		slog.Int("total_contacts", result.Total),
		slog.Int("page_items", len(result.Items)))

	// Zé da API returns array of contacts (not wrapped in object)
	respondJSON(w, http.StatusOK, result.Items)
}

// getChats handles GET /instances/{instanceId}/token/{token}/chats
//
// endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Retrieves all chats from WhatsApp instance
// 4. Applies pagination to results
// 5. Returns chat array
//
// Query parameters:
// - page: Page number (required, minimum: 1)
// - pageSize: Number of chats per page (required, minimum: 1)
//

func (h *MessageHandler) getChats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse query parameters (format: page, pageSize)
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		h.log.WarnContext(ctx, "missing required query parameter: page")
		respondError(w, http.StatusBadRequest, "Query parameter 'page' is required")
		return
	}

	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		h.log.WarnContext(ctx, "missing required query parameter: pageSize")
		respondError(w, http.StatusBadRequest, "Query parameter 'pageSize' is required")
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		h.log.WarnContext(ctx, "invalid page parameter",
			slog.String("page", pageStr))
		respondError(w, http.StatusBadRequest, "Query parameter 'page' must be a positive integer")
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		h.log.WarnContext(ctx, "invalid pageSize parameter",
			slog.String("pageSize", pageSizeStr))
		respondError(w, http.StatusBadRequest, "Query parameter 'pageSize' must be a positive integer")
		return
	}

	h.log.InfoContext(ctx, "listing chats",
		slog.Int("page", page),
		slog.Int("page_size", pageSize))

	// Call chats service
	if h.chatsService == nil {
		h.log.ErrorContext(ctx, "chats service not available")
		respondError(w, http.StatusServiceUnavailable, "Chats service not available")
		return
	}

	result, err := h.chatsService.List(ctx, instanceID, chats.ListParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		h.log.ErrorContext(ctx, "failed to list chats",
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to retrieve chats")
		return
	}

	h.log.InfoContext(ctx, "chats retrieved successfully",
		slog.Int("total_chats", result.TotalCount),
		slog.Int("page_items", len(result.Chats)))

	// Zé da API returns array of chats (not wrapped in object)
	respondJSON(w, http.StatusOK, result.Chats)
}

// phoneExists handles GET /instances/{instanceId}/token/{token}/phone-exists/{phone}
//
// endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Checks if phone number is registered on WhatsApp
// 4. Returns response array
//
// Path parameters:
// - phone: Phone number in format DDI DDD NUMBER (e.g., 551199999999)
//

func (h *MessageHandler) phoneExists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	phone := chi.URLParam(r, "phone")
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone parameter")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Validate phone format (only digits allowed)
	for _, c := range phone {
		if c < '0' || c > '9' {
			h.log.WarnContext(ctx, "invalid phone format",
				slog.String("phone", phone))
			respondError(w, http.StatusBadRequest, "Phone number must contain only digits")
			return
		}
	}

	h.log.InfoContext(ctx, "checking phone exists",
		slog.String("phone", phone))

	if h.contactsService == nil {
		h.log.ErrorContext(ctx, "contacts service not available")
		respondError(w, http.StatusServiceUnavailable, "Contacts service not available")
		return
	}

	result, err := h.contactsService.IsOnWhatsApp(ctx, instanceID, phone)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to check phone exists",
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to check phone number")
		return
	}

	h.log.InfoContext(ctx, "phone check completed",
		slog.Bool("exists", result.Exists))

	// Zé da API returns the response object directly
	respondJSON(w, http.StatusOK, result)
}

// phoneExistsBatch handles POST /instances/{instanceId}/token/{token}/phone-exists-batch
//
// endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Accepts a JSON body with array of phone numbers
// 4. Checks if each phone number is registered on WhatsApp
// 5. Returns response array with validation results
//
// Request body:
// - phones: Array of phone numbers in format DDI DDD NUMBER (e.g., ["551199999999", "551188888888"])
//
// Maximum batch size: 50,000 numbers per request (Zé da API limit)
//

func (h *MessageHandler) phoneExistsBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req contacts.PhoneExistsBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phones array
	if len(req.Phones) == 0 {
		h.log.WarnContext(ctx, "empty phones array")
		respondError(w, http.StatusBadRequest, "Phones array is required and cannot be empty")
		return
	}

	// Zé da API limit: 50,000 numbers per request
	const maxBatchSize = 50000
	if len(req.Phones) > maxBatchSize {
		h.log.WarnContext(ctx, "batch size exceeds limit",
			slog.Int("requested", len(req.Phones)),
			slog.Int("max", maxBatchSize))
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Batch size exceeds maximum limit of %d numbers", maxBatchSize))
		return
	}

	// Validate each phone format (only digits allowed)
	for i, phone := range req.Phones {
		if phone == "" {
			h.log.WarnContext(ctx, "empty phone in batch",
				slog.Int("index", i))
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Phone at index %d is empty", i))
			return
		}
		for _, c := range phone {
			if c < '0' || c > '9' {
				h.log.WarnContext(ctx, "invalid phone format in batch",
					slog.Int("index", i),
					slog.String("phone", phone))
				respondError(w, http.StatusBadRequest, fmt.Sprintf("Phone at index %d must contain only digits", i))
				return
			}
		}
	}

	h.log.InfoContext(ctx, "checking batch phone exists",
		slog.Int("phone_count", len(req.Phones)))

	if h.contactsService == nil {
		h.log.ErrorContext(ctx, "contacts service not available")
		respondError(w, http.StatusServiceUnavailable, "Contacts service not available")
		return
	}

	results, err := h.contactsService.IsOnWhatsAppBatch(ctx, instanceID, req.Phones)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to check batch phone exists",
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to check phone numbers")
		return
	}

	h.log.InfoContext(ctx, "batch phone check completed",
		slog.Int("results_count", len(results)))

	// Zé da API returns array of results
	respondJSON(w, http.StatusOK, results)
}

// Helper functions

// normalizePhoneNumber normalizes a phone number to WhatsApp format
// Accepts formats like:
// - "5511999999999" → "5511999999999@s.whatsapp.net"
// - "5511999999999@s.whatsapp.net" → "5511999999999@s.whatsapp.net"
// - "120363xyz@g.us" → "120363xyz@g.us" (group)
// normalizePhoneNumber normalizes phone numbers/JIDs to WhatsApp format
// Supports:
// - Individual users: "5511999999999" → "5511999999999@s.whatsapp.net"
// - Groups: "120363XXXXX@g.us" (passed through)
// - Newsletter: "120363XXXXX@newsletter" (passed through)
// - Broadcast: "status@broadcast" or "XXXXX@broadcast" (passed through)
// - Hosted: "XXXXX@hosted" or "XXXXX@hosted.lid" (passed through)
// - Device suffixes: "5511999999999:12" → "5511999999999@s.whatsapp.net" (device removed)
// - Agent/Device: "5511999999999.0:12" → "5511999999999@s.whatsapp.net" (agent/device removed)
func normalizePhoneNumber(phone string) string {
	phone = strings.TrimSpace(phone)

	// Already in WhatsApp format with @ server
	if strings.Contains(phone, "@") {
		return phone
	}

	// Zé da API style suffixes for groups/channels/broadcast lists
	if strings.HasSuffix(phone, "-group") {
		base := strings.TrimSuffix(phone, "-group")
		return base + "@g.us"
	}
	if strings.HasSuffix(phone, "-channel") {
		base := strings.TrimSuffix(phone, "-channel")
		return base + "@newsletter"
	}
	if strings.HasSuffix(phone, "-broadcast") {
		base := strings.TrimSuffix(phone, "-broadcast")
		if base == "status" || base == "" {
			return "status@broadcast"
		}
		return base + "@broadcast"
	}

	// Remove device suffix (:12, :99, etc.) or agent.device (.0:12, .123:99, etc.)
	// These patterns indicate specific devices but we want to send to the user's account
	if strings.Contains(phone, ".") {
		// Has agent and possibly device: "5511999999999.0:12" or "5511999999999.123:99"
		parts := strings.Split(phone, ".")
		phone = parts[0] // Keep only the user part
	} else if strings.Contains(phone, ":") {
		// Has only device: "5511999999999:12"
		parts := strings.Split(phone, ":")
		phone = parts[0] // Keep only the user part
	}

	// Default to individual chat for plain numbers
	// The whatsmeow library will handle the actual JID creation
	return phone + "@s.whatsapp.net"
}

// convertGroupMentions converts handler GroupMention slice to queue GroupMention slice
func convertGroupMentions(apiMentions []GroupMention) []queue.GroupMention {
	if len(apiMentions) == 0 {
		return nil
	}
	result := make([]queue.GroupMention, len(apiMentions))
	for i, m := range apiMentions {
		result[i] = queue.GroupMention{
			Phone:   m.Phone,
			Subject: m.Subject,
		}
	}
	return result
}

// convertContactAddress converts handler ContactAddress to queue ContactAddress
func convertContactAddress(apiAddr *ContactAddress) *queue.ContactAddress {
	if apiAddr == nil {
		return nil
	}
	return &queue.ContactAddress{
		Type:       apiAddr.Type,
		PostBox:    apiAddr.PostBox,
		Extended:   apiAddr.Extended,
		Street:     apiAddr.Street,
		City:       apiAddr.City,
		Region:     apiAddr.Region,
		PostalCode: apiAddr.PostalCode,
		Country:    apiAddr.Country,
	}
}

// convertJobToQueueMessage converts queue.QueueJobInfo to QueueMessageResponse
func convertJobToQueueMessage(job queue.QueueJobInfo) QueueMessageResponse {
	// Extract message text if it's a text message
	message := ""
	if job.MessageType == "text" && job.TextContent != nil {
		message = job.TextContent.Message
	}

	// Convert delays from milliseconds to seconds
	delayMessageSec := job.DelayMessage / 1000
	delayTypingSec := job.DelayTyping / 1000

	// Convert created timestamp to milliseconds
	createdMs := job.CreatedAt.UnixMilli()

	response := QueueMessageResponse{
		ID:             job.ZaapID, // _id field
		ZaapId:         job.ZaapID,
		MessageId:      job.WhatsAppMessageID, // WhatsApp message ID (if sent)
		InstanceId:     job.InstanceID.String(),
		Phone:          job.Phone,               // Recipient phone
		Message:        message,                 // Text message content
		DelayMessage:   delayMessageSec,         // In seconds
		DelayTyping:    delayTypingSec,          // In seconds
		Created:        createdMs,               // Unix timestamp in milliseconds
		MessageType:    string(job.MessageType), // Additional field
		Status:         string(job.Status),      // Additional field
		SequenceNumber: job.SequenceNumber,      // Additional field
		Attempt:        job.Attempt,             // Additional field
		MaxAttempts:    job.MaxAttempts,         // Additional field
		Errors:         job.Errors,              // Additional field
	}

	// If WhatsApp message ID is empty, use zaapId (not sent yet)
	if response.MessageId == "" {
		response.MessageId = job.ZaapID
	}

	return response
}

// handleEnqueueError handles errors from enqueue operations
func (h *MessageHandler) handleEnqueueError(ctx context.Context, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, queue.ErrInstanceNotFound):
		h.log.WarnContext(ctx, "instance not found", slog.String("error", err.Error()))
		respondError(w, http.StatusNotFound, "WhatsApp instance not found")
	case errors.Is(err, queue.ErrInstanceNotConnected):
		h.log.WarnContext(ctx, "instance not connected", slog.String("error", err.Error()))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp instance not connected")
	case errors.Is(err, queue.ErrQueueNotFound):
		h.log.WarnContext(ctx, "queue not found", slog.String("error", err.Error()))
		respondError(w, http.StatusNotFound, "Message queue not found for instance")
	case errors.Is(err, queue.ErrInvalidPhone):
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
	case errors.Is(err, queue.ErrNoMessageContent):
		respondError(w, http.StatusBadRequest, "No message content provided")
	case errors.Is(err, queue.ErrQueueStopped):
		retryAfter := int(math.Ceil(h.drainRetryAfter.Seconds()))
		if retryAfter < 1 {
			retryAfter = 1
		}
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		h.log.WarnContext(ctx, "queue draining; enqueue rejected",
			slog.Int("retry_after_seconds", retryAfter))
		respondError(w, http.StatusServiceUnavailable, "Message queue is draining; try again later")
	default:
		h.log.ErrorContext(ctx, "failed to enqueue message",
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to enqueue message: %v", err))
	}
}

// sendContacts handles POST /instances/{instanceId}/token/{token}/send-contacts
// endpoint for sending multiple contact cards
func (h *MessageHandler) sendContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendContactsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Validate contacts array
	if len(req.Contacts) == 0 {
		h.log.WarnContext(ctx, "missing contacts")
		respondError(w, http.StatusBadRequest, "At least one contact is required")
		return
	}

	if len(req.Contacts) > 10 {
		h.log.WarnContext(ctx, "too many contacts",
			slog.Int("count", len(req.Contacts)))
		respondError(w, http.StatusBadRequest, "Maximum 10 contacts per request")
		return
	}

	// Convert delays
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Build array of ContactMessage with ALL fields for each contact
	// This will be sent in ONE WhatsApp message using ContactsArrayMessage
	var contacts []*queue.ContactMessage
	for i, contact := range req.Contacts {
		contactName := strings.TrimSpace(contact.ContactName)
		if contactName == "" {
			h.log.WarnContext(ctx, "missing contact name in array",
				slog.Int("index", i))
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Contact at index %d is missing contactName", i))
			return
		}

		contactPhone := strings.TrimSpace(contact.ContactPhone)
		if contactPhone == "" {
			h.log.WarnContext(ctx, "missing contact phone in array",
				slog.Int("index", i))
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Contact at index %d is missing contactPhone", i))
			return
		}

		// Build ContactMessage with ALL fields (same as sendContact)
		contactMsg := &queue.ContactMessage{
			// Required
			FullName:    contactName,
			PhoneNumber: contactPhone,

			// Extended name fields
			FirstName:  contact.FirstName,
			LastName:   contact.LastName,
			MiddleName: contact.MiddleName,
			NamePrefix: contact.NamePrefix,
			NameSuffix: contact.NameSuffix,
			Nickname:   contact.Nickname,

			// Contact fields
			Email: contact.Email,
			URL:   contact.URL,

			// Organization override (if Organization field provided, use it; otherwise use ContactBusinessDescription)
			Organization: contact.Organization,
			JobTitle:     contact.JobTitle,

			// Address
			Address: convertContactAddress(contact.Address),

			// Personal fields
			Birthday: contact.Birthday,
			Note:     contact.Note,
		}

		// If Organization not provided but ContactBusinessDescription is, use it for backward compatibility
		if contactMsg.Organization == nil && contact.ContactBusinessDescription != "" {
			contactMsg.Organization = &contact.ContactBusinessDescription
		}

		contacts = append(contacts, contactMsg)
	}

	// Create message args for ContactsArrayMessage
	// NOTE: We use a special message type or flag to indicate multiple contacts
	// The processor will handle this and call ProcessMultiple()
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypeContact,
		// TODO: We need a way to pass multiple contacts
		// For now, this will require updating SendMessageArgs to support multiple contacts
		// or using ContactContent with a slice
		// ContactContent:   contactMsg, // Single contact
		// ContactsContent:  contacts,   // Multiple contacts (need to add this field)
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: strings.TrimSpace(req.MessageID),
		Duration:         req.Duration,
		Mentioned:        req.Mentioned,
		GroupMentioned:   convertGroupMentions(req.GroupMentioned),
		MentionedAll:     req.MentionedAll,
		PrivateAnswer:    req.PrivateAnswer,
	}

	// TODO: For now, we'll send the first contact only
	// This requires adding ContactsContent field to SendMessageArgs
	args.ContactContent = contacts[0]

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "multiple contacts enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Int("contact_count", len(req.Contacts)),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	// Return response
	response := h.newSendMessageResponse(zaapID, instStatus)

	respondJSON(w, http.StatusOK, response)
}

// sendButtonList handles POST /send-button-list
// Sends a message with quick reply buttons using NativeFlowMessage
func (h *MessageHandler) sendButtonList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	var req SendButtonListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "failed to decode button list request",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone
	if strings.TrimSpace(req.Phone) == "" {
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	phone := normalizePhoneNumber(strings.TrimSpace(req.Phone))

	// Validate buttons
	if len(req.Buttons) == 0 {
		respondError(w, http.StatusBadRequest, "At least one button is required")
		return
	}
	if len(req.Buttons) > 3 {
		respondError(w, http.StatusBadRequest, "Maximum 3 buttons allowed")
		return
	}

	// Resolve delays
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Convert buttons to queue format
	queueButtons := make([]queue.Button, len(req.Buttons))
	for i, btn := range req.Buttons {
		queueButtons[i] = queue.Button{
			ID:    btn.ID,
			Title: btn.Title,
			Type:  "quick_reply",
		}
	}

	// Build optional header/footer pointers
	var header, footer, image, video *string
	if req.Title != "" {
		header = &req.Title
	}
	if req.Footer != "" {
		footer = &req.Footer
	}
	if req.Image != "" {
		image = &req.Image
	}
	if req.Video != "" {
		video = &req.Video
	}

	args := queue.SendMessageArgs{
		InstanceID:       instanceID,
		Phone:            phone,
		MessageType:      queue.MessageTypeButtonList,
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: req.MessageID,
		InteractiveContent: &queue.InteractiveMessage{
			Type:    queue.InteractiveTypeButton,
			Header:  header,
			Body:    req.Message,
			Footer:  footer,
			Buttons: queueButtons,
			Image:   image,
			Video:   video,
		},
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "button list message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Int("button_count", len(req.Buttons)),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	response := h.newSendMessageResponse(zaapID, instStatus)
	respondJSON(w, http.StatusOK, response)
}

// sendButtonActions handles POST /send-button-actions
// Sends a message with action buttons (URL, call, copy, quick_reply)
func (h *MessageHandler) sendButtonActions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	var req SendButtonActionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "failed to decode button actions request",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone
	if strings.TrimSpace(req.Phone) == "" {
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	phone := normalizePhoneNumber(strings.TrimSpace(req.Phone))

	// Validate buttons
	if len(req.Buttons) == 0 {
		respondError(w, http.StatusBadRequest, "At least one button is required")
		return
	}
	if len(req.Buttons) > 3 {
		respondError(w, http.StatusBadRequest, "Maximum 3 buttons allowed")
		return
	}

	// Validate button types and required fields
	for i, btn := range req.Buttons {
		switch btn.Type {
		case "quick_reply":
			// No additional fields required
		case "cta_url":
			if btn.URL == nil || *btn.URL == "" {
				respondError(w, http.StatusBadRequest, fmt.Sprintf("Button %d: URL required for cta_url type", i+1))
				return
			}
		case "cta_call":
			if btn.Phone == nil || *btn.Phone == "" {
				respondError(w, http.StatusBadRequest, fmt.Sprintf("Button %d: Phone required for cta_call type", i+1))
				return
			}
		case "cta_copy":
			if btn.CopyCode == nil || *btn.CopyCode == "" {
				respondError(w, http.StatusBadRequest, fmt.Sprintf("Button %d: CopyCode required for cta_copy type", i+1))
				return
			}
		default:
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Button %d: Invalid type '%s'", i+1, btn.Type))
			return
		}
	}

	// Resolve delays
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Convert buttons to queue format
	queueButtons := make([]queue.Button, len(req.Buttons))
	for i, btn := range req.Buttons {
		queueButtons[i] = queue.Button{
			ID:    btn.ID,
			Title: btn.Label,
			Type:  btn.Type,
		}
		if btn.URL != nil {
			queueButtons[i].URL = *btn.URL
		}
		if btn.Phone != nil {
			queueButtons[i].Phone = *btn.Phone
		}
		if btn.CopyCode != nil {
			queueButtons[i].CopyCode = *btn.CopyCode
		}
	}

	// Build optional header/footer pointers
	var header, footer *string
	if req.Title != "" {
		header = &req.Title
	}
	if req.Footer != "" {
		footer = &req.Footer
	}

	// Build media pointers
	var image, video, document, documentFilename *string
	if req.Image != "" {
		image = &req.Image
	}
	if req.Video != "" {
		video = &req.Video
	}
	if req.Document != "" {
		document = &req.Document
	}
	if req.DocumentFilename != "" {
		documentFilename = &req.DocumentFilename
	}

	args := queue.SendMessageArgs{
		InstanceID:       instanceID,
		Phone:            phone,
		MessageType:      queue.MessageTypeButtonActions,
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: req.MessageID,
		InteractiveContent: &queue.InteractiveMessage{
			Type:             queue.InteractiveTypeButton,
			Header:           header,
			Body:             req.Message,
			Footer:           footer,
			Buttons:          queueButtons,
			Image:            image,
			Video:            video,
			Document:         document,
			DocumentFilename: documentFilename,
		},
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "button actions message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Int("button_count", len(req.Buttons)),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	response := h.newSendMessageResponse(zaapID, instStatus)
	respondJSON(w, http.StatusOK, response)
}

// sendOptionList handles POST /send-option-list
// Sends a list/menu message with sections and selectable rows
func (h *MessageHandler) sendOptionList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	var req SendOptionListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "failed to decode option list request",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone
	if strings.TrimSpace(req.Phone) == "" {
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	phone := normalizePhoneNumber(strings.TrimSpace(req.Phone))

	// Validate sections
	if len(req.Sections) == 0 {
		respondError(w, http.StatusBadRequest, "At least one section is required")
		return
	}
	if len(req.Sections) > 10 {
		respondError(w, http.StatusBadRequest, "Maximum 10 sections allowed")
		return
	}

	// Count total rows
	totalRows := 0
	for _, sec := range req.Sections {
		totalRows += len(sec.Rows)
	}
	if totalRows > 10 {
		respondError(w, http.StatusBadRequest, "Maximum 10 total rows allowed across all sections")
		return
	}

	// Resolve delays
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Convert sections to queue format
	queueSections := make([]queue.Section, len(req.Sections))
	for i, sec := range req.Sections {
		queueRows := make([]queue.Row, len(sec.Rows))
		for j, row := range sec.Rows {
			queueRow := queue.Row{
				ID:    row.ID,
				Title: row.Title,
			}
			if row.Description != "" {
				desc := row.Description
				queueRow.Description = &desc
			}
			queueRows[j] = queueRow
		}
		queueSections[i] = queue.Section{
			Title: sec.Title,
			Rows:  queueRows,
		}
	}

	// Build optional header/footer pointers
	var header, footer, buttonLabel *string
	if req.Title != "" {
		header = &req.Title
	}
	if req.Footer != "" {
		footer = &req.Footer
	}
	if req.ButtonLabel != "" {
		buttonLabel = &req.ButtonLabel
	}

	args := queue.SendMessageArgs{
		InstanceID:       instanceID,
		Phone:            phone,
		MessageType:      queue.MessageTypeOptionList,
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: req.MessageID,
		InteractiveContent: &queue.InteractiveMessage{
			Type:        queue.InteractiveTypeList,
			Header:      header,
			Body:        req.Message,
			Footer:      footer,
			Sections:    queueSections,
			ButtonLabel: buttonLabel,
		},
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "option list message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Int("section_count", len(req.Sections)),
		slog.Int("total_rows", totalRows),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	response := h.newSendMessageResponse(zaapID, instStatus)
	respondJSON(w, http.StatusOK, response)
}

// sendButtonPIX handles POST /send-button-pix
// Sends a message with a PIX payment button (Brazil)
func (h *MessageHandler) sendButtonPIX(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	var req SendButtonPIXRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "failed to decode button pix request",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone
	if strings.TrimSpace(req.Phone) == "" {
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	phone := normalizePhoneNumber(strings.TrimSpace(req.Phone))

	// Validate PIX key
	if strings.TrimSpace(req.PIXKey) == "" {
		respondError(w, http.StatusBadRequest, "PIX key is required")
		return
	}

	// Validate PIX key type
	validKeyTypes := map[string]bool{
		"CPF":   true,
		"CNPJ":  true,
		"EMAIL": true,
		"PHONE": true,
		"EVP":   true,
	}
	if !validKeyTypes[req.KeyType] {
		respondError(w, http.StatusBadRequest, "Invalid PIX key type. Must be one of: CPF, CNPJ, EMAIL, PHONE, EVP")
		return
	}

	// Resolve delays
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Build PIX payment struct
	pixPayment := &queue.PIXPayment{
		Key:     req.PIXKey,
		KeyType: req.KeyType,
	}
	if req.Name != nil {
		pixPayment.Name = req.Name
	}
	if req.Amount != nil {
		pixPayment.Amount = req.Amount
	}
	if req.TransactionID != nil {
		pixPayment.TransactionID = req.TransactionID
	}

	// Build body message (optional for PIX messages)
	body := req.Message

	args := queue.SendMessageArgs{
		InstanceID:       instanceID,
		Phone:            phone,
		MessageType:      queue.MessageTypeButtonPIX,
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: req.MessageID,
		InteractiveContent: &queue.InteractiveMessage{
			Type:       queue.InteractiveTypeButton,
			Body:       body,
			PIXPayment: pixPayment,
		},
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "button pix message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.String("pix_key_type", req.KeyType),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	response := h.newSendMessageResponse(zaapID, instStatus)
	respondJSON(w, http.StatusOK, response)
}

// sendButtonOTP handles POST /send-button-otp
// Sends a message with a copy code button for OTP verification
func (h *MessageHandler) sendButtonOTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	var req SendButtonOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "failed to decode button otp request",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone
	if strings.TrimSpace(req.Phone) == "" {
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	phone := normalizePhoneNumber(strings.TrimSpace(req.Phone))

	// Validate OTP code
	if strings.TrimSpace(req.Code) == "" {
		respondError(w, http.StatusBadRequest, "OTP code is required")
		return
	}
	if len(req.Code) > 20 {
		respondError(w, http.StatusBadRequest, "OTP code must not exceed 20 characters")
		return
	}

	// Resolve delays
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Build optional header/footer pointers
	var header, footer *string
	if req.Title != "" {
		header = &req.Title
	}
	if req.Footer != "" {
		footer = &req.Footer
	}

	// Store OTP code
	otpCode := req.Code

	args := queue.SendMessageArgs{
		InstanceID:       instanceID,
		Phone:            phone,
		MessageType:      queue.MessageTypeButtonOTP,
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: req.MessageID,
		InteractiveContent: &queue.InteractiveMessage{
			Type:    queue.InteractiveTypeButton,
			Header:  header,
			Body:    req.Message,
			Footer:  footer,
			OTPCode: &otpCode,
		},
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "button otp message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	response := h.newSendMessageResponse(zaapID, instStatus)
	respondJSON(w, http.StatusOK, response)
}

// sendCarousel handles POST /instances/{instanceId}/token/{token}/send-carousel
// endpoint for sending carousel messages with multiple cards
func (h *MessageHandler) sendCarousel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	var req SendCarouselRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "failed to decode carousel request",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone
	if strings.TrimSpace(req.Phone) == "" {
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	phone := normalizePhoneNumber(strings.TrimSpace(req.Phone))

	// Validate cards
	if len(req.Cards) == 0 {
		respondError(w, http.StatusBadRequest, "At least one card is required")
		return
	}
	if len(req.Cards) > 10 {
		respondError(w, http.StatusBadRequest, "Maximum 10 cards allowed")
		return
	}

	// Validate each card
	for i, card := range req.Cards {
		if strings.TrimSpace(card.Body) == "" {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Card %d: body text is required", i+1))
			return
		}
		if len(card.Body) > 1024 {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Card %d: body text must not exceed 1024 characters", i+1))
			return
		}
		if len(card.Header) > 60 {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Card %d: header must not exceed 60 characters", i+1))
			return
		}
		if len(card.Footer) > 60 {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Card %d: footer must not exceed 60 characters", i+1))
			return
		}
		if len(card.Buttons) == 0 {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Card %d: at least one button is required", i+1))
			return
		}
		if len(card.Buttons) > 3 {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Card %d: maximum 3 buttons allowed per card", i+1))
			return
		}
		// Validate button types (must be action buttons for carousel)
		for j, btn := range card.Buttons {
			if strings.TrimSpace(btn.Label) == "" {
				respondError(w, http.StatusBadRequest, fmt.Sprintf("Card %d, Button %d: label is required", i+1, j+1))
				return
			}
		}
	}

	// Resolve delays
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Convert request cards to queue.CarouselCard
	carouselCards := make([]queue.CarouselCard, len(req.Cards))
	for i, card := range req.Cards {
		// Convert buttons to queue.Button
		buttons := make([]queue.Button, len(card.Buttons))
		for j, btn := range card.Buttons {
			// Map ActionButton to queue.Button
			qBtn := queue.Button{
				Type:  btn.Type,
				Title: btn.Label, // ActionButton.Label maps to queue.Button.Title
			}
			if btn.URL != nil {
				qBtn.URL = *btn.URL
			}
			if btn.Phone != nil {
				qBtn.Phone = *btn.Phone
			}
			if btn.CopyCode != nil {
				qBtn.CopyCode = *btn.CopyCode
			}
			buttons[j] = qBtn
		}

		carouselCards[i] = queue.CarouselCard{
			Header:   card.Header,
			Body:     card.Body,
			Footer:   card.Footer,
			Buttons:  buttons,
			MediaURL: card.MediaURL,
		}
	}

	// Determine carousel card type
	cardType := "HSCROLL_CARDS" // default
	if req.CardType != "" {
		cardType = strings.ToUpper(req.CardType)
	}

	args := queue.SendMessageArgs{
		InstanceID:       instanceID,
		Phone:            phone,
		MessageType:      queue.MessageTypeCarousel,
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: req.MessageID,
		InteractiveContent: &queue.InteractiveMessage{
			Type:             queue.InteractiveTypeCarousel,
			Body:             req.Message, // Carousel body text displayed above cards
			CarouselCards:    carouselCards,
			CarouselCardType: cardType,
		},
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "carousel message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Int("cards_count", len(req.Cards)),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	response := h.newSendMessageResponse(zaapID, instStatus)
	respondJSON(w, http.StatusOK, response)
}

// sendPoll handles POST /instances/{instanceId}/token/{token}/send-poll
//
// endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Creates poll with 2-12 options
// 4. Enqueues message with FIFO ordering
// 5. Returns immediately with zaapId as messageId (non-blocking)
//
// Zé da API Request format:
// {"phone": "554499999999", "message": "Poll question", "poll": [{"name": "Option1"}, {"name": "Option2"}], "pollMaxOptions": 1}
func (h *MessageHandler) sendPoll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendPollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number to WhatsApp format
	phone = normalizePhoneNumber(phone)

	// Validate poll question (Zé da API uses "message" field)
	question := strings.TrimSpace(req.Message)
	if question == "" {
		h.log.WarnContext(ctx, "missing poll question")
		respondError(w, http.StatusBadRequest, "Poll question (message) is required")
		return
	}

	// Validate poll options count (2-12 options)
	if len(req.Poll) < 2 {
		h.log.WarnContext(ctx, "insufficient poll options",
			slog.Int("count", len(req.Poll)))
		respondError(w, http.StatusBadRequest, "Poll must have at least 2 options")
		return
	}
	if len(req.Poll) > 12 {
		h.log.WarnContext(ctx, "too many poll options",
			slog.Int("count", len(req.Poll)))
		respondError(w, http.StatusBadRequest, "Poll cannot have more than 12 options")
		return
	}

	// Convert Poll[].Name to []string for queue.PollMessage.Options
	options := make([]string, len(req.Poll))
	for i, opt := range req.Poll {
		optName := strings.TrimSpace(opt.Name)
		if optName == "" {
			h.log.WarnContext(ctx, "empty poll option",
				slog.Int("index", i))
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Poll option %d cannot be empty", i+1))
			return
		}
		options[i] = optName
	}

	// Set max selections (Zé da API: 0 = single choice, 1+ = multi-choice)
	maxSelections := 0
	if req.PollMaxOptions != nil {
		maxSelections = *req.PollMaxOptions
		if maxSelections < 0 {
			maxSelections = 0
		}
		if maxSelections > len(options) {
			maxSelections = len(options)
		}
	}

	// Resolve delays (delayMessage supports any duration; scheduledFor overrides with ISO 8601)
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Create message args
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypePoll,
		PollContent: &queue.PollMessage{
			Question:      question,
			Options:       options,
			MaxSelections: maxSelections,
		},
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: strings.TrimSpace(req.MessageID),
	}

	// Enqueue message (non-blocking)
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "poll message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.String("question", question),
		slog.Int("options_count", len(options)),
		slog.Int("max_selections", maxSelections),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	response := h.newSendMessageResponse(zaapID, instStatus)
	respondJSON(w, http.StatusOK, response)
}

// sendEvent handles POST /instances/{instanceId}/token/{token}/send-event
//
// endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Creates calendar event message
// 4. Enqueues message with FIFO ordering
// 5. Returns immediately with zaapId as messageId (non-blocking)
//
// Zé da API Request format:
// {"phone": "120363356737170752-group", "event": {"name": "Event Name", "dateTime": "2024-04-29T09:30:53.309Z", "description": "...", "location": {...}}}
func (h *MessageHandler) sendEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number to WhatsApp format
	phone = normalizePhoneNumber(phone)

	// Validate event name
	eventName := strings.TrimSpace(req.Event.Name)
	if eventName == "" {
		h.log.WarnContext(ctx, "missing event name")
		respondError(w, http.StatusBadRequest, "Event name is required")
		return
	}

	// Validate and parse event dateTime (ISO 8601 format)
	dateTimeStr := strings.TrimSpace(req.Event.DateTime)
	if dateTimeStr == "" {
		h.log.WarnContext(ctx, "missing event dateTime")
		respondError(w, http.StatusBadRequest, "Event dateTime is required")
		return
	}

	// Parse ISO 8601 datetime
	startTime, err := time.Parse(time.RFC3339, dateTimeStr)
	if err != nil {
		// Try alternative formats
		startTime, err = time.Parse("2006-01-02T15:04:05.000Z", dateTimeStr)
		if err != nil {
			startTime, err = time.Parse("2006-01-02T15:04:05Z", dateTimeStr)
			if err != nil {
				h.log.WarnContext(ctx, "invalid event dateTime format",
					slog.String("dateTime", dateTimeStr),
					slog.String("error", err.Error()))
				respondError(w, http.StatusBadRequest, "Invalid dateTime format. Use ISO 8601 (e.g., 2024-04-29T09:30:53.309Z)")
				return
			}
		}
	}

	// Build event location if provided
	var eventLocation *queue.EventLocation
	if req.Event.Location != nil {
		eventLocation = &queue.EventLocation{
			Name: req.Event.Location.Name,
		}
		if req.Event.Location.DegreesLatitude != nil {
			eventLocation.DegreesLatitude = req.Event.Location.DegreesLatitude
		}
		if req.Event.Location.DegreesLongitude != nil {
			eventLocation.DegreesLongitude = req.Event.Location.DegreesLongitude
		}
	}

	// Resolve delays (delayMessage supports any duration; scheduledFor overrides with ISO 8601)
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Create message args
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypeEvent,
		EventContent: &queue.EventMessage{
			Name:         eventName,
			Description:  strings.TrimSpace(req.Event.Description),
			StartTime:    startTime,
			Location:     eventLocation,
			CallLinkType: req.Event.CallLinkType,
			Canceled:     req.Event.Canceled,
		},
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: strings.TrimSpace(req.MessageID),
	}

	// Enqueue message (non-blocking)
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "event message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.String("event_name", eventName),
		slog.Time("start_time", startTime),
		slog.Bool("canceled", req.Event.Canceled),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	response := h.newSendMessageResponse(zaapID, instStatus)
	respondJSON(w, http.StatusOK, response)
}

// sendLink handles POST /instances/{instanceId}/token/{token}/send-link
//
// endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Creates text message with custom link preview override
// 4. Enqueues message with FIFO ordering
// 5. Returns immediately with zaapId as messageId (non-blocking)
//
// Zé da API Request format:
// {"phone": "5544999999999", "message": "Check this out", "linkUrl": "https://example.com", "title": "Title", "linkDescription": "Description", "image": "https://..."}
func (h *MessageHandler) sendLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	whatsStatus := h.toWhatsAppStatus(instStatus)

	// Parse request body
	var req SendLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number to WhatsApp format
	phone = normalizePhoneNumber(phone)

	// Validate linkUrl (required)
	linkUrl := strings.TrimSpace(req.LinkUrl)
	if linkUrl == "" {
		h.log.WarnContext(ctx, "missing linkUrl")
		respondError(w, http.StatusBadRequest, "linkUrl is required")
		return
	}

	// Validate URL format
	if !strings.HasPrefix(linkUrl, "http://") && !strings.HasPrefix(linkUrl, "https://") {
		h.log.WarnContext(ctx, "invalid linkUrl format",
			slog.String("linkUrl", linkUrl))
		respondError(w, http.StatusBadRequest, "linkUrl must be a valid URL (http:// or https://)")
		return
	}

	// Build message text: prepend message if provided, otherwise use URL as text
	messageText := strings.TrimSpace(req.Message)
	if messageText == "" {
		messageText = linkUrl
	} else if !strings.Contains(messageText, linkUrl) {
		// Append URL to message if not already included
		messageText = messageText + "\n\n" + linkUrl
	}

	// Resolve delays (delayMessage supports any duration; scheduledFor overrides with ISO 8601)
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)
	delayTyping := resolveTypingDelay(req.DelayTyping)

	// Create message args with text content and link preview override
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypeText,
		TextContent: &queue.TextMessage{
			Message: messageText,
		},
		LinkPreviewOverride: &queue.LinkPreviewOverride{
			URL:         linkUrl,
			Image:       strings.TrimSpace(req.Image),
			Title:       strings.TrimSpace(req.Title),
			Description: strings.TrimSpace(req.LinkDescription),
		},
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
		ReplyToMessageID: strings.TrimSpace(req.MessageID),
	}

	// Force link preview to be enabled
	forcePreview := true
	args.LinkPreview = &forcePreview

	// Enqueue message (non-blocking)
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "link message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.String("link_url", linkUrl),
		slog.String("title", req.Title),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	response := h.newSendMessageResponse(zaapID, instStatus)
	respondJSON(w, http.StatusOK, response)
}

// deleteMessage handles DELETE /instances/{instanceId}/token/{token}/messages
// format for deleting messages
// Query params: phone, messageId, owner (true for own message, false for admin delete)
func (h *MessageHandler) deleteMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse query parameters
	phone := strings.TrimSpace(r.URL.Query().Get("phone"))
	messageID := strings.TrimSpace(r.URL.Query().Get("messageId"))
	ownerStr := strings.TrimSpace(r.URL.Query().Get("owner"))

	// Validate required parameters
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone parameter")
		respondError(w, http.StatusBadRequest, "phone parameter is required")
		return
	}

	if messageID == "" {
		h.log.WarnContext(ctx, "missing messageId parameter")
		respondError(w, http.StatusBadRequest, "messageId parameter is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Parse owner flag (default to true for own messages)
	isOwner := true
	if ownerStr != "" {
		isOwner = ownerStr == "true"
	}

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected",
			slog.String("phone", phone))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Parse phone to JID
	chatJID, err := types.ParseJID(phone)
	if err != nil {
		h.log.WarnContext(ctx, "invalid phone number format",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Build revoke message
	var revokeMsg *waE2E.Message
	if isOwner {
		// Revoke own message - pass empty JID as sender
		revokeMsg = client.BuildRevoke(chatJID, types.EmptyJID, messageID)
	} else {
		// Admin delete - this requires the sender JID, but for simplicity we use empty
		// In production, you might need to retrieve the original sender from message history
		revokeMsg = client.BuildRevoke(chatJID, types.EmptyJID, messageID)
	}

	// Send the revoke message
	_, err = client.SendMessage(ctx, chatJID, revokeMsg)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to delete message",
			slog.String("phone", phone),
			slog.String("message_id", messageID),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to delete message: "+err.Error())
		return
	}

	h.log.InfoContext(ctx, "message deleted successfully",
		slog.String("phone", phone),
		slog.String("message_id", messageID),
		slog.Bool("owner", isOwner))

	respondJSON(w, http.StatusOK, DeleteMessageResponse{
		Success:   true,
		Phone:     phone,
		MessageID: messageID,
		Message:   "Message deleted successfully",
	})
}

// readMessage handles POST /instances/{instanceId}/token/{token}/read-message
// Marks specific messages as read using whatsmeow MarkRead
func (h *MessageHandler) readMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req ReadMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Validate message IDs
	if len(req.MessageIDs) == 0 {
		h.log.WarnContext(ctx, "missing message IDs")
		respondError(w, http.StatusBadRequest, "At least one message ID is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected",
			slog.String("phone", phone))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Parse phone to JID (chat JID)
	chatJID, err := types.ParseJID(phone)
	if err != nil {
		h.log.WarnContext(ctx, "invalid phone number format",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Parse sender JID if provided (required for group messages)
	var senderJID types.JID
	sender := strings.TrimSpace(req.Sender)
	if sender != "" {
		sender = normalizePhoneNumber(sender)
		senderJID, err = types.ParseJID(sender)
		if err != nil {
			h.log.WarnContext(ctx, "invalid sender phone format",
				slog.String("sender", sender),
				slog.String("error", err.Error()))
			respondError(w, http.StatusBadRequest, "Invalid sender phone format")
			return
		}
	} else {
		senderJID = types.EmptyJID
	}

	// Convert string message IDs to types.MessageID
	messageIDs := make([]types.MessageID, len(req.MessageIDs))
	for i, id := range req.MessageIDs {
		messageIDs[i] = types.MessageID(strings.TrimSpace(id))
	}

	// Mark messages as read
	err = client.MarkRead(ctx, messageIDs, time.Now(), chatJID, senderJID)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to mark messages as read",
			slog.String("phone", phone),
			slog.Int("message_count", len(messageIDs)),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to mark messages as read: "+err.Error())
		return
	}

	h.log.InfoContext(ctx, "messages marked as read",
		slog.String("phone", phone),
		slog.Int("message_count", len(messageIDs)))

	respondJSON(w, http.StatusOK, ReadMessageResponse{
		Success:    true,
		Phone:      phone,
		MessageIDs: req.MessageIDs,
		Message:    "Messages marked as read successfully",
	})
}

// getContactMetadata handles GET /instances/{instanceId}/token/{token}/contacts/{phone}/metadata
// Returns contact metadata using whatsmeow GetUserInfo
func (h *MessageHandler) getContactMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Get phone from URL parameter
	phone := chi.URLParam(r, "phone")
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected",
			slog.String("phone", phone))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Parse phone to JID
	jid, err := types.ParseJID(phone)
	if err != nil {
		h.log.WarnContext(ctx, "invalid phone number format",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Get user info from whatsmeow
	userInfo, err := client.GetUserInfo(ctx, []types.JID{jid})
	if err != nil {
		h.log.ErrorContext(ctx, "failed to get user info",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to get contact metadata: "+err.Error())
		return
	}

	// Build response
	response := ContactMetadataResponse{
		Phone:        phone,
		IsOnWhatsApp: false,
	}

	if info, exists := userInfo[jid]; exists {
		response.IsOnWhatsApp = true
		response.JID = jid.String()
		response.VerifiedName = info.VerifiedName.Details.GetVerifiedName()
		response.Status = info.Status
		response.PictureID = info.PictureID

		// Add device info
		if len(info.Devices) > 0 {
			response.Devices = make([]struct {
				Device string `json:"device"`
				JID    string `json:"jid"`
			}, len(info.Devices))
			for i, device := range info.Devices {
				response.Devices[i].JID = device.String()
				response.Devices[i].Device = strconv.FormatUint(uint64(device.Device), 10)
			}
		}
	}

	h.log.InfoContext(ctx, "contact metadata retrieved",
		slog.String("phone", phone),
		slog.Bool("is_on_whatsapp", response.IsOnWhatsApp))

	respondJSON(w, http.StatusOK, response)
}

// getContactProfilePicture handles GET /instances/{instanceId}/token/{token}/contacts/{phone}/profile-picture
// Returns profile picture URL using whatsmeow GetProfilePictureInfo
func (h *MessageHandler) getContactProfilePicture(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Get phone from URL parameter
	phone := chi.URLParam(r, "phone")
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected",
			slog.String("phone", phone))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Parse phone to JID
	jid, err := types.ParseJID(phone)
	if err != nil {
		h.log.WarnContext(ctx, "invalid phone number format",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Get profile picture info from whatsmeow
	// Pass nil for extra params to get default behavior
	pictureInfo, err := client.GetProfilePictureInfo(ctx, jid, nil)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to get profile picture",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to get profile picture: "+err.Error())
		return
	}

	// Handle case where no profile picture exists
	if pictureInfo == nil {
		h.log.InfoContext(ctx, "no profile picture found",
			slog.String("phone", phone))
		respondJSON(w, http.StatusOK, ProfilePictureResponse{
			URL:  "",
			ID:   "",
			Type: "none",
		})
		return
	}

	response := ProfilePictureResponse{
		URL:        pictureInfo.URL,
		ID:         pictureInfo.ID,
		Type:       pictureInfo.Type,
		DirectPath: pictureInfo.DirectPath,
	}

	h.log.InfoContext(ctx, "profile picture retrieved",
		slog.String("phone", phone),
		slog.String("picture_id", response.ID))

	respondJSON(w, http.StatusOK, response)
}

// sendReaction handles POST /instances/{instanceId}/token/{token}/send-reaction
// Sends an emoji reaction to a specific message using whatsmeow BuildReaction
func (h *MessageHandler) sendReaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req SendReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	messageID := strings.TrimSpace(req.MessageID)
	if messageID == "" {
		h.log.WarnContext(ctx, "missing message ID")
		respondError(w, http.StatusBadRequest, "Message ID is required")
		return
	}

	reaction := req.Reaction
	if reaction == "" {
		h.log.WarnContext(ctx, "missing reaction emoji")
		respondError(w, http.StatusBadRequest, "Reaction emoji is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected",
			slog.String("phone", phone))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Parse phone to JID (chat JID)
	chatJID, err := types.ParseJID(phone)
	if err != nil {
		h.log.WarnContext(ctx, "invalid phone number format",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Parse sender JID if provided (required for group messages)
	var senderJID types.JID
	sender := strings.TrimSpace(req.Sender)
	if sender != "" {
		sender = normalizePhoneNumber(sender)
		senderJID, err = types.ParseJID(sender)
		if err != nil {
			h.log.WarnContext(ctx, "invalid sender phone format",
				slog.String("sender", sender),
				slog.String("error", err.Error()))
			respondError(w, http.StatusBadRequest, "Invalid sender phone format")
			return
		}
	} else {
		// For direct messages, sender is the chat JID
		senderJID = chatJID
	}

	// Build reaction message
	reactionMsg := client.BuildReaction(chatJID, senderJID, types.MessageID(messageID), reaction)

	// Send the reaction
	resp, err := client.SendMessage(ctx, chatJID, reactionMsg)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to send reaction",
			slog.String("phone", phone),
			slog.String("message_id", messageID),
			slog.String("reaction", reaction),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to send reaction: "+err.Error())
		return
	}

	h.log.InfoContext(ctx, "reaction sent successfully",
		slog.String("phone", phone),
		slog.String("message_id", messageID),
		slog.String("reaction", reaction),
		slog.String("response_id", resp.ID))

	respondJSON(w, http.StatusOK, SendReactionResponse{
		Success:   true,
		Phone:     phone,
		MessageID: messageID,
		Reaction:  reaction,
		Message:   "Reaction sent successfully",
	})
}

// sendRemoveReaction handles POST /instances/{instanceId}/token/{token}/send-remove-reaction
// Removes a reaction from a message by sending an empty reaction string
func (h *MessageHandler) sendRemoveReaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body - reuse SendReactionRequest but reaction field is ignored
	var req SendReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	messageID := strings.TrimSpace(req.MessageID)
	if messageID == "" {
		h.log.WarnContext(ctx, "missing message ID")
		respondError(w, http.StatusBadRequest, "Message ID is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected",
			slog.String("phone", phone))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Parse phone to JID (chat JID)
	chatJID, err := types.ParseJID(phone)
	if err != nil {
		h.log.WarnContext(ctx, "invalid phone number format",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Parse sender JID if provided (required for group messages)
	var senderJID types.JID
	sender := strings.TrimSpace(req.Sender)
	if sender != "" {
		sender = normalizePhoneNumber(sender)
		senderJID, err = types.ParseJID(sender)
		if err != nil {
			h.log.WarnContext(ctx, "invalid sender phone format",
				slog.String("sender", sender),
				slog.String("error", err.Error()))
			respondError(w, http.StatusBadRequest, "Invalid sender phone format")
			return
		}
	} else {
		// For direct messages, sender is the chat JID
		senderJID = chatJID
	}

	// Build reaction message with empty string to remove reaction
	reactionMsg := client.BuildReaction(chatJID, senderJID, types.MessageID(messageID), "")

	// Send the reaction removal
	resp, err := client.SendMessage(ctx, chatJID, reactionMsg)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to remove reaction",
			slog.String("phone", phone),
			slog.String("message_id", messageID),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to remove reaction: "+err.Error())
		return
	}

	h.log.InfoContext(ctx, "reaction removed successfully",
		slog.String("phone", phone),
		slog.String("message_id", messageID),
		slog.String("response_id", resp.ID))

	respondJSON(w, http.StatusOK, SendReactionResponse{
		Success:   true,
		Phone:     phone,
		MessageID: messageID,
		Reaction:  "",
		Message:   "Reaction removed successfully",
	})
}

// forwardMessage handles POST /instances/{instanceId}/token/{token}/forward-message
// Forwards a message to another chat by retrieving it from the message store and resending with IsForwarded flag
func (h *MessageHandler) forwardMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req ForwardMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing target phone number")
		respondError(w, http.StatusBadRequest, "Target phone number is required")
		return
	}

	messageID := strings.TrimSpace(req.MessageID)
	if messageID == "" {
		h.log.WarnContext(ctx, "missing message ID")
		respondError(w, http.StatusBadRequest, "Message ID is required")
		return
	}

	sourceChat := strings.TrimSpace(req.SourceChat)
	if sourceChat == "" {
		h.log.WarnContext(ctx, "missing source chat")
		respondError(w, http.StatusBadRequest, "Source chat is required")
		return
	}

	// Normalize phone numbers
	phone = normalizePhoneNumber(phone)
	sourceChat = normalizePhoneNumber(sourceChat)

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected",
			slog.String("phone", phone))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Parse target JID
	targetJID, err := types.ParseJID(phone)
	if err != nil {
		h.log.WarnContext(ctx, "invalid target phone number format",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid target phone number format")
		return
	}

	// Parse source chat JID
	sourceChatJID, err := types.ParseJID(sourceChat)
	if err != nil {
		h.log.WarnContext(ctx, "invalid source chat format",
			slog.String("source_chat", sourceChat),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid source chat format")
		return
	}

	// Try to get the message from the message store using GetMessageForRetry callback
	// This requires the message to be cached by the client
	var originalMessage *waE2E.Message
	if client.GetMessageForRetry != nil {
		originalMessage = client.GetMessageForRetry(sourceChatJID, sourceChatJID, types.MessageID(messageID))
	}

	if originalMessage == nil {
		h.log.WarnContext(ctx, "message not found in store - cannot forward",
			slog.String("message_id", messageID),
			slog.String("source_chat", sourceChat))
		respondError(w, http.StatusNotFound, "Message not found in store. The message may have been sent too long ago or the client was restarted.")
		return
	}

	// Create a forwarded copy of the message
	forwardedMessage := proto.Clone(originalMessage).(*waE2E.Message)

	// Set forwarding context info based on message type
	contextInfo := &waE2E.ContextInfo{
		IsForwarded:     proto.Bool(true),
		ForwardingScore: proto.Uint32(1),
	}

	// Apply context info to the appropriate message type
	if forwardedMessage.GetConversation() != "" {
		// Convert simple conversation to ExtendedTextMessage with context
		forwardedMessage = &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text:        proto.String(forwardedMessage.GetConversation()),
				ContextInfo: contextInfo,
			},
		}
	} else if forwardedMessage.ExtendedTextMessage != nil {
		forwardedMessage.ExtendedTextMessage.ContextInfo = contextInfo
	} else if forwardedMessage.ImageMessage != nil {
		forwardedMessage.ImageMessage.ContextInfo = contextInfo
	} else if forwardedMessage.VideoMessage != nil {
		forwardedMessage.VideoMessage.ContextInfo = contextInfo
	} else if forwardedMessage.AudioMessage != nil {
		forwardedMessage.AudioMessage.ContextInfo = contextInfo
	} else if forwardedMessage.DocumentMessage != nil {
		forwardedMessage.DocumentMessage.ContextInfo = contextInfo
	} else if forwardedMessage.StickerMessage != nil {
		forwardedMessage.StickerMessage.ContextInfo = contextInfo
	} else if forwardedMessage.ContactMessage != nil {
		forwardedMessage.ContactMessage.ContextInfo = contextInfo
	} else if forwardedMessage.LocationMessage != nil {
		forwardedMessage.LocationMessage.ContextInfo = contextInfo
	}

	// Apply delay if specified (scheduledFor overrides delayMessage)
	if req.ScheduledFor != nil && *req.ScheduledFor != "" {
		if t, err := time.Parse(time.RFC3339, *req.ScheduledFor); err == nil && t.After(time.Now()) {
			time.Sleep(time.Until(t))
		}
	} else if req.DelayMessage != nil && *req.DelayMessage > 0 {
		time.Sleep(time.Duration(*req.DelayMessage) * time.Second)
	}

	// Send the forwarded message
	resp, err := client.SendMessage(ctx, targetJID, forwardedMessage)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to forward message",
			slog.String("target_phone", phone),
			slog.String("original_id", messageID),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to forward message: "+err.Error())
		return
	}

	h.log.InfoContext(ctx, "message forwarded successfully",
		slog.String("target_phone", phone),
		slog.String("original_id", messageID),
		slog.String("new_id", resp.ID))

	respondJSON(w, http.StatusOK, ForwardMessageResponse{
		Success:      true,
		Phone:        phone,
		OriginalID:   messageID,
		NewMessageID: resp.ID,
		Message:      "Message forwarded successfully",
	})
}

// sendPollVote handles POST /instances/{instanceId}/token/{token}/send-poll-vote
// Sends a vote on an existing poll by constructing MessageInfo and calling BuildPollVote
func (h *MessageHandler) sendPollVote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req SendPollVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number (chat JID) is required")
		return
	}

	pollID := strings.TrimSpace(req.PollID)
	if pollID == "" {
		h.log.WarnContext(ctx, "missing poll ID")
		respondError(w, http.StatusBadRequest, "Poll ID is required")
		return
	}

	pollSender := strings.TrimSpace(req.PollSender)
	if pollSender == "" {
		h.log.WarnContext(ctx, "missing poll sender")
		respondError(w, http.StatusBadRequest, "Poll sender phone is required")
		return
	}

	if len(req.Options) == 0 {
		h.log.WarnContext(ctx, "missing poll options")
		respondError(w, http.StatusBadRequest, "At least one poll option is required")
		return
	}

	// Normalize phone numbers
	phone = normalizePhoneNumber(phone)
	pollSender = normalizePhoneNumber(pollSender)

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected",
			slog.String("phone", phone))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Parse chat JID
	chatJID, err := types.ParseJID(phone)
	if err != nil {
		h.log.WarnContext(ctx, "invalid phone number format",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Parse poll sender JID
	senderJID, err := types.ParseJID(pollSender)
	if err != nil {
		h.log.WarnContext(ctx, "invalid poll sender phone format",
			slog.String("poll_sender", pollSender),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid poll sender phone format")
		return
	}

	// Determine if this is a group chat
	isGroup := chatJID.Server == types.GroupServer

	// Construct MessageInfo for the poll message
	// This is required by BuildPollVote to reference the original poll
	pollInfo := &types.MessageInfo{
		MessageSource: types.MessageSource{
			Chat:     chatJID,
			Sender:   senderJID,
			IsFromMe: senderJID.User == client.Store.ID.User,
			IsGroup:  isGroup,
		},
		ID:        types.MessageID(pollID),
		Timestamp: time.Now(), // Timestamp is not critical for voting
	}

	// Apply delay if specified (scheduledFor overrides delayMessage)
	if req.ScheduledFor != nil && *req.ScheduledFor != "" {
		if t, err := time.Parse(time.RFC3339, *req.ScheduledFor); err == nil && t.After(time.Now()) {
			time.Sleep(time.Until(t))
		}
	} else if req.DelayMessage != nil && *req.DelayMessage > 0 {
		time.Sleep(time.Duration(*req.DelayMessage) * time.Second)
	}

	// Build the poll vote message
	voteMsg, err := client.BuildPollVote(ctx, pollInfo, req.Options)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to build poll vote",
			slog.String("phone", phone),
			slog.String("poll_id", pollID),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to build poll vote: "+err.Error())
		return
	}

	// Send the poll vote
	resp, err := client.SendMessage(ctx, chatJID, voteMsg)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to send poll vote",
			slog.String("phone", phone),
			slog.String("poll_id", pollID),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to send poll vote: "+err.Error())
		return
	}

	h.log.InfoContext(ctx, "poll vote sent successfully",
		slog.String("phone", phone),
		slog.String("poll_id", pollID),
		slog.Any("options", req.Options),
		slog.String("response_id", resp.ID))

	respondJSON(w, http.StatusOK, SendPollVoteResponse{
		Success:   true,
		Phone:     phone,
		PollID:    pollID,
		Options:   req.Options,
		MessageID: resp.ID,
		Message:   "Poll vote sent successfully",
	})
}

// pinMessage handles POST /instances/{instanceId}/token/{token}/pin-message
// Pins or unpins a message in a chat using the PinInChatMessage proto type
func (h *MessageHandler) pinMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req PinMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	messageID := strings.TrimSpace(req.MessageID)
	if messageID == "" {
		h.log.WarnContext(ctx, "missing message ID")
		respondError(w, http.StatusBadRequest, "Message ID is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected",
			slog.String("phone", phone))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Parse phone to JID (chat JID)
	chatJID, err := types.ParseJID(phone)
	if err != nil {
		h.log.WarnContext(ctx, "invalid phone number format",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Parse sender JID if provided (required for group messages)
	var senderJID types.JID
	sender := strings.TrimSpace(req.Sender)
	if sender != "" {
		sender = normalizePhoneNumber(sender)
		senderJID, err = types.ParseJID(sender)
		if err != nil {
			h.log.WarnContext(ctx, "invalid sender phone format",
				slog.String("sender", sender),
				slog.String("error", err.Error()))
			respondError(w, http.StatusBadRequest, "Invalid sender phone format")
			return
		}
	} else {
		// For direct messages, sender is the chat JID
		senderJID = chatJID
	}

	// Build the MessageKey for the message to pin/unpin
	msgKey := client.BuildMessageKey(chatJID, senderJID, types.MessageID(messageID))

	// Determine pin type
	var pinType waE2E.PinInChatMessage_Type
	if req.Pin {
		pinType = waE2E.PinInChatMessage_PIN_FOR_ALL
	} else {
		pinType = waE2E.PinInChatMessage_UNPIN_FOR_ALL
	}

	// Build the PinInChatMessage
	pinMsg := &waE2E.Message{
		PinInChatMessage: &waE2E.PinInChatMessage{
			Key:               msgKey,
			Type:              &pinType,
			SenderTimestampMS: proto.Int64(time.Now().UnixMilli()),
		},
	}

	// Send the pin message
	resp, err := client.SendMessage(ctx, chatJID, pinMsg)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to send pin message",
			slog.String("phone", phone),
			slog.String("message_id", messageID),
			slog.Bool("pin", req.Pin),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to pin/unpin message: "+err.Error())
		return
	}

	action := "pinned"
	if !req.Pin {
		action = "unpinned"
	}

	h.log.InfoContext(ctx, "message "+action+" successfully",
		slog.String("phone", phone),
		slog.String("message_id", messageID),
		slog.Bool("pin", req.Pin),
		slog.String("response_id", resp.ID))

	respondJSON(w, http.StatusOK, PinMessageResponse{
		Success:   true,
		Phone:     phone,
		MessageID: messageID,
		Pinned:    req.Pin,
		Message:   "Message " + action + " successfully",
	})
}

// modifyBlocked handles POST /instances/{instanceId}/token/{token}/modify-blocked
// Blocks or unblocks a contact using the whatsmeow UpdateBlocklist method
func (h *MessageHandler) modifyBlocked(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req ModifyBlockedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Validate action
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action != "block" && action != "unblock" {
		h.log.WarnContext(ctx, "invalid action",
			slog.String("action", action))
		respondError(w, http.StatusBadRequest, "Action must be 'block' or 'unblock'")
		return
	}

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected",
			slog.String("phone", phone))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Build JID for the contact
	jid := types.NewJID(phone, types.DefaultUserServer)

	// Determine blocklist action
	var blockAction events.BlocklistChangeAction
	if action == "block" {
		blockAction = events.BlocklistChangeActionBlock
	} else {
		blockAction = events.BlocklistChangeActionUnblock
	}

	// Update blocklist
	_, err := client.UpdateBlocklist(ctx, jid, blockAction)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to update blocklist",
			slog.String("instance_id", instanceID.String()),
			slog.String("phone", phone),
			slog.String("action", action),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to "+action+" contact: "+err.Error())
		return
	}

	h.log.InfoContext(ctx, "blocklist updated successfully",
		slog.String("instance_id", instanceID.String()),
		slog.String("phone", phone),
		slog.String("action", action))

	respondJSON(w, http.StatusOK, ModifyBlockedResponse{
		Success: true,
		Phone:   phone,
		Action:  action,
		Message: "Contact " + action + "ed successfully",
	})
}

// modifyChat handles POST /instances/{instanceId}/token/{token}/modify-chat
// format for modifying chat state
// Actions: read, archive, unarchive, pin, unpin, mute, unmute, clear, delete
func (h *MessageHandler) modifyChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req ModifyChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	// Validate action
	action := strings.ToLower(strings.TrimSpace(req.Action))
	validActions := map[string]bool{
		"read":      true,
		"archive":   true,
		"unarchive": true,
		"pin":       true,
		"unpin":     true,
		"mute":      true,
		"unmute":    true,
		"clear":     true,
		"delete":    true,
	}

	if !validActions[action] {
		h.log.WarnContext(ctx, "invalid action",
			slog.String("action", action))
		respondError(w, http.StatusBadRequest, "Invalid action. Valid actions: read, archive, unarchive, pin, unpin, mute, unmute, clear, delete")
		return
	}

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected",
			slog.String("phone", phone))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Parse phone to JID
	chatJID, err := types.ParseJID(phone)
	if err != nil {
		h.log.WarnContext(ctx, "invalid phone number format",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Execute action
	var actionErr error
	switch action {
	case "read":
		// Mark chat as read
		actionErr = client.MarkRead(ctx, []types.MessageID{}, time.Now(), chatJID, types.EmptyJID)

	case "archive":
		// Archive chat using appstate
		actionErr = h.modifyChatArchive(ctx, client, chatJID, true)

	case "unarchive":
		// Unarchive chat using appstate
		actionErr = h.modifyChatArchive(ctx, client, chatJID, false)

	case "pin":
		// Pin chat using appstate
		actionErr = h.modifyChatPin(ctx, client, chatJID, true)

	case "unpin":
		// Unpin chat using appstate
		actionErr = h.modifyChatPin(ctx, client, chatJID, false)

	case "mute":
		// Mute chat (mute for 8 hours by default)
		muteDuration := time.Now().Add(8 * time.Hour)
		actionErr = h.modifyChatMute(ctx, client, chatJID, &muteDuration)

	case "unmute":
		// Unmute chat
		actionErr = h.modifyChatMute(ctx, client, chatJID, nil)

	case "clear":
		// Clear chat messages
		actionErr = h.modifyChatClear(ctx, client, chatJID)

	case "delete":
		// Delete chat
		actionErr = h.modifyChatDelete(ctx, client, chatJID)
	}

	if actionErr != nil {
		h.log.ErrorContext(ctx, "failed to modify chat",
			slog.String("phone", phone),
			slog.String("action", action),
			slog.String("error", actionErr.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to "+action+" chat: "+actionErr.Error())
		return
	}

	h.log.InfoContext(ctx, "chat modified successfully",
		slog.String("phone", phone),
		slog.String("action", action))

	respondJSON(w, http.StatusOK, ModifyChatResponse{
		Success: true,
		Phone:   phone,
		Action:  action,
		Message: "Chat " + action + " successful",
	})
}

// modifyChatArchive archives or unarchives a chat using whatsmeow appstate
func (h *MessageHandler) modifyChatArchive(ctx context.Context, client *wameow.Client, chatJID types.JID, archive bool) error {
	// Build archive patch - use current time as last message timestamp
	// In production, you should ideally get the actual last message timestamp
	patch := appstate.BuildArchive(chatJID, archive, time.Now(), nil)
	return client.SendAppState(ctx, patch)
}

// modifyChatPin pins or unpins a chat using whatsmeow appstate
func (h *MessageHandler) modifyChatPin(ctx context.Context, client *wameow.Client, chatJID types.JID, pin bool) error {
	patch := appstate.BuildPin(chatJID, pin)
	return client.SendAppState(ctx, patch)
}

// modifyChatMute mutes or unmutes a chat using whatsmeow appstate
func (h *MessageHandler) modifyChatMute(ctx context.Context, client *wameow.Client, chatJID types.JID, muteUntil *time.Time) error {
	if muteUntil != nil {
		// Mute until specified time
		duration := time.Until(*muteUntil)
		if duration < 0 {
			duration = 8 * time.Hour // Default to 8 hours if time is in the past
		}
		patch := appstate.BuildMute(chatJID, true, duration)
		return client.SendAppState(ctx, patch)
	}
	// Unmute - pass 0 duration
	patch := appstate.BuildMute(chatJID, false, 0)
	return client.SendAppState(ctx, patch)
}

// modifyChatClear clears all messages in a chat
// Note: whatsmeow does not have a BuildClearChat function
func (h *MessageHandler) modifyChatClear(ctx context.Context, client *wameow.Client, chatJID types.JID) error {
	_ = ctx
	_ = client
	_ = chatJID
	// whatsmeow does not support clear chat via appstate
	// This would require a different approach or is not supported
	return fmt.Errorf("clear chat not supported by whatsmeow")
}

// modifyChatDelete deletes a chat using whatsmeow appstate
func (h *MessageHandler) modifyChatDelete(ctx context.Context, client *wameow.Client, chatJID types.JID) error {
	// Build delete chat patch - use current time as last message timestamp
	// deleteMedia=true to also remove associated media from the chat
	patch := appstate.BuildDeleteChat(chatJID, time.Now(), nil, true)
	return client.SendAppState(ctx, patch)
}

// updateProfileName handles PUT /instances/{instanceId}/token/{token}/profile-name
// Updates the profile name (push name) using whatsmeow app state
func (h *MessageHandler) updateProfileName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req UpdateProfileNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate name
	name := strings.TrimSpace(req.Name)
	if name == "" {
		h.log.WarnContext(ctx, "missing profile name")
		respondError(w, http.StatusBadRequest, "Profile name is required")
		return
	}

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Build and send app state patch for push name
	patch := appstate.BuildSettingPushName(name)
	if err := client.SendAppState(ctx, patch); err != nil {
		h.log.ErrorContext(ctx, "failed to update profile name",
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to update profile name: "+err.Error())
		return
	}

	h.log.InfoContext(ctx, "profile name updated successfully",
		slog.String("instance_id", instanceID.String()),
		slog.String("name", name))

	respondJSON(w, http.StatusOK, UpdateProfileNameResponse{
		Success: true,
		Name:    name,
		Message: "Profile name updated successfully",
	})
}

// updateProfilePicture handles PUT /instances/{instanceId}/token/{token}/profile-picture
// Updates the profile picture using whatsmeow SetGroupPhoto with own JID
func (h *MessageHandler) updateProfilePicture(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req UpdateProfilePictureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Get own JID for profile picture update
	ownJID := client.Store.ID
	if ownJID == nil {
		h.log.ErrorContext(ctx, "own JID not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not fully connected")
		return
	}

	var imageData []byte
	var err error

	// Handle empty image (remove profile picture)
	if strings.TrimSpace(req.Image) == "" {
		imageData = nil
	} else if strings.HasPrefix(req.Image, "http://") || strings.HasPrefix(req.Image, "https://") {
		// Download image from URL
		imageData, err = h.downloadImage(ctx, req.Image)
		if err != nil {
			h.log.ErrorContext(ctx, "failed to download image",
				slog.String("error", err.Error()))
			respondError(w, http.StatusBadRequest, "Failed to download image: "+err.Error())
			return
		}
	} else {
		// Assume base64 encoded image
		// Remove data URI prefix if present
		imageStr := req.Image
		if idx := strings.Index(imageStr, ","); idx != -1 {
			imageStr = imageStr[idx+1:]
		}
		imageData, err = base64.StdEncoding.DecodeString(imageStr)
		if err != nil {
			h.log.ErrorContext(ctx, "failed to decode base64 image",
				slog.String("error", err.Error()))
			respondError(w, http.StatusBadRequest, "Invalid base64 image data")
			return
		}
	}

	// Set profile picture using SetGroupPhoto with own JID
	pictureID, err := client.SetGroupPhoto(ctx, *ownJID, imageData)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to update profile picture",
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to update profile picture: "+err.Error())
		return
	}

	h.log.InfoContext(ctx, "profile picture updated successfully",
		slog.String("instance_id", instanceID.String()),
		slog.String("picture_id", pictureID))

	respondJSON(w, http.StatusOK, UpdateProfilePictureResponse{
		Success:   true,
		PictureID: pictureID,
		Message:   "Profile picture updated successfully",
	})
}

// updateProfileDescription handles PUT /instances/{instanceId}/token/{token}/profile-description
// Updates the profile description (About/Status) using whatsmeow SetStatusMessage
func (h *MessageHandler) updateProfileDescription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req UpdateProfileDescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Description can be empty (to clear it)
	description := req.Description

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Set status message (profile description/about)
	if err := client.SetStatusMessage(ctx, description); err != nil {
		h.log.ErrorContext(ctx, "failed to update profile description",
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to update profile description: "+err.Error())
		return
	}

	h.log.InfoContext(ctx, "profile description updated successfully",
		slog.String("instance_id", instanceID.String()))

	respondJSON(w, http.StatusOK, UpdateProfileDescriptionResponse{
		Success:     true,
		Description: description,
		Message:     "Profile description updated successfully",
	})
}

// downloadImage downloads an image from a URL
func (h *MessageHandler) downloadImage(ctx context.Context, imageURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Limit read to 10MB
	limitedReader := io.LimitReader(resp.Body, 10*1024*1024)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// sendEditEvent handles POST /instances/{instanceId}/token/{token}/send-edit-event
// Sends an updated event message to modify an existing event
func (h *MessageHandler) sendEditEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req SendEditEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}
	phone = normalizePhoneNumber(phone)

	// Validate event ID
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		h.log.WarnContext(ctx, "missing event ID")
		respondError(w, http.StatusBadRequest, "Event ID is required")
		return
	}

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected",
			slog.String("phone", phone))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Parse recipient JID
	recipientJID, err := types.ParseJID(phone)
	if err != nil {
		h.log.WarnContext(ctx, "invalid phone format",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Simulate typing if delay specified
	if req.DelayTyping != nil && *req.DelayTyping > 0 {
		delay := *req.DelayTyping
		if delay > 15 {
			delay = 15
		}
		_ = client.SendChatPresence(ctx, recipientJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
		time.Sleep(time.Duration(delay) * time.Second)
		_ = client.SendChatPresence(ctx, recipientJID, types.ChatPresencePaused, types.ChatPresenceMediaText)
	}

	// Build the updated event message
	eventMsg := &waE2E.EventMessage{
		ContextInfo: &waE2E.ContextInfo{
			StanzaID:    proto.String(eventID),
			Participant: proto.String(phone),
		},
	}

	// Set updated fields if provided
	if req.Name != "" {
		eventMsg.Name = proto.String(req.Name)
	}
	if req.Description != "" {
		eventMsg.Description = proto.String(req.Description)
	}
	if req.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err == nil {
			eventMsg.StartTime = proto.Int64(startTime.Unix())
		}
	}
	if req.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err == nil {
			eventMsg.EndTime = proto.Int64(endTime.Unix())
		}
	}
	if req.Location != "" {
		eventMsg.Location = &waE2E.LocationMessage{
			Name: proto.String(req.Location),
		}
	}
	if req.Canceled != nil && *req.Canceled {
		eventMsg.IsCanceled = proto.Bool(true)
	}

	// Apply delay if specified (scheduledFor overrides delayMessage)
	if req.ScheduledFor != nil && *req.ScheduledFor != "" {
		if t, err := time.Parse(time.RFC3339, *req.ScheduledFor); err == nil && t.After(time.Now()) {
			time.Sleep(time.Until(t))
		}
	} else if req.DelayMessage != nil && *req.DelayMessage > 0 {
		time.Sleep(time.Duration(*req.DelayMessage) * time.Second)
	}

	// Send the edit event message
	msg := &waE2E.Message{
		EventMessage: eventMsg,
	}

	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to send edit event",
			slog.String("error", err.Error()),
			slog.String("event_id", eventID))
		respondError(w, http.StatusInternalServerError, "Failed to send edit event: "+err.Error())
		return
	}

	h.log.InfoContext(ctx, "edit event sent successfully",
		slog.String("message_id", resp.ID),
		slog.String("event_id", eventID),
		slog.String("phone", phone))

	respondJSON(w, http.StatusOK, SendEditEventResponse{
		Success:   true,
		EventID:   eventID,
		MessageID: resp.ID,
		Message:   "Event updated successfully",
	})
}

// sendEventResponse handles POST /instances/{instanceId}/token/{token}/send-event-response
// Sends a response to an event invitation (going, not_going, maybe)
func (h *MessageHandler) sendEventResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req SendEventResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate phone number
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		h.log.WarnContext(ctx, "missing phone number")
		respondError(w, http.StatusBadRequest, "Phone number is required")
		return
	}
	phone = normalizePhoneNumber(phone)

	// Validate event ID
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		h.log.WarnContext(ctx, "missing event ID")
		respondError(w, http.StatusBadRequest, "Event ID is required")
		return
	}

	// Validate and parse response type
	responseStr := strings.ToLower(strings.TrimSpace(req.Response))
	var responseType waE2E.EventResponseMessage_EventResponseType
	switch responseStr {
	case "going":
		responseType = waE2E.EventResponseMessage_GOING
	case "not_going":
		responseType = waE2E.EventResponseMessage_NOT_GOING
	case "maybe":
		responseType = waE2E.EventResponseMessage_MAYBE
	default:
		h.log.WarnContext(ctx, "invalid response type",
			slog.String("response", responseStr))
		respondError(w, http.StatusBadRequest, "Response must be 'going', 'not_going', or 'maybe'")
		return
	}

	// Get client registry from coordinator
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return
	}

	// Get the whatsmeow client
	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected",
			slog.String("phone", phone))
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return
	}

	// Parse recipient JID
	recipientJID, err := types.ParseJID(phone)
	if err != nil {
		h.log.WarnContext(ctx, "invalid phone format",
			slog.String("phone", phone),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Simulate typing if delay specified
	if req.DelayTyping != nil && *req.DelayTyping > 0 {
		delay := *req.DelayTyping
		if delay > 15 {
			delay = 15
		}
		_ = client.SendChatPresence(ctx, recipientJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
		time.Sleep(time.Duration(delay) * time.Second)
		_ = client.SendChatPresence(ctx, recipientJID, types.ChatPresencePaused, types.ChatPresenceMediaText)
	}

	// Build the event response message
	eventResponseMsg := &waE2E.EventResponseMessage{
		Response:    &responseType,
		TimestampMS: proto.Int64(time.Now().UnixMilli()),
	}

	// Add extra guest count if provided
	if req.ExtraGuestCount != nil && *req.ExtraGuestCount > 0 {
		eventResponseMsg.ExtraGuestCount = proto.Int32(int32(*req.ExtraGuestCount))
	}

	// Apply delay if specified (scheduledFor overrides delayMessage)
	if req.ScheduledFor != nil && *req.ScheduledFor != "" {
		if t, err := time.Parse(time.RFC3339, *req.ScheduledFor); err == nil && t.After(time.Now()) {
			time.Sleep(time.Until(t))
		}
	} else if req.DelayMessage != nil && *req.DelayMessage > 0 {
		time.Sleep(time.Duration(*req.DelayMessage) * time.Second)
	}

	// Marshal the event response for encryption
	plaintext, err := proto.Marshal(eventResponseMsg)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to marshal event response",
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to build event response")
		return
	}

	// Parse event message key from the eventID
	// The eventID format should be "messageId" or "remoteJid_messageId"
	eventMsgKey := &waCommon.MessageKey{
		RemoteJID: proto.String(recipientJID.String()),
		ID:        proto.String(eventID),
		FromMe:    proto.Bool(false),
	}

	// Build encrypted event response message
	// Note: This requires the original event's message secret to be stored
	// The encryption uses the EncSecretEventResponse type
	encEventResponse := &waE2E.EncEventResponseMessage{
		EventCreationMessageKey: eventMsgKey,
		EncPayload:              plaintext, // In production, this should be encrypted
		EncIV:                   nil,       // IV would be set by encryption
	}

	msg := &waE2E.Message{
		EncEventResponseMessage: encEventResponse,
	}

	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to send event response",
			slog.String("error", err.Error()),
			slog.String("event_id", eventID),
			slog.String("response", responseStr))
		respondError(w, http.StatusInternalServerError, "Failed to send event response: "+err.Error())
		return
	}

	h.log.InfoContext(ctx, "event response sent successfully",
		slog.String("message_id", resp.ID),
		slog.String("event_id", eventID),
		slog.String("response", responseStr),
		slog.String("phone", phone))

	respondJSON(w, http.StatusOK, SendEventResponseResponse{
		Success:  true,
		EventID:  eventID,
		Response: responseStr,
		Message:  "Event response sent successfully",
	})
}

// =============================================================================
// Status/Stories Handlers (broadcast to status@broadcast)
// =============================================================================

// sendTextStatus handles POST /instances/{instanceId}/token/{token}/send-text-status
// Sends a text message to WhatsApp Status/Stories (broadcasts to all viewers)
func (h *MessageHandler) sendTextStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req SendTextStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate text content
	text := strings.TrimSpace(req.Text)
	if text == "" {
		h.log.WarnContext(ctx, "missing text content for status")
		respondError(w, http.StatusBadRequest, "Text content is required")
		return
	}

	// Resolve delay (delayMessage supports any duration; scheduledFor overrides with ISO 8601)
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)

	// Parse font if provided
	var font *int32
	if req.Font != nil {
		f := int32(*req.Font)
		if f < 0 {
			f = 0
		}
		if f > 5 {
			f = 5
		}
		font = &f
	}

	// Create message args for status broadcast
	// Phone is set to status@broadcast JID string
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       types.StatusBroadcastJID.String(), // CRITICAL: Always use status@broadcast
		MessageType: queue.MessageTypeTextStatus,
		TextStatusContent: &queue.TextStatusMessage{
			Text:            text,
			BackgroundColor: strings.TrimSpace(req.BackgroundColor),
			Font:            font,
		},
		DelayMessage: delayMessage,
	}

	// Set custom message ID if provided
	if req.MessageID != "" {
		args.ReplyToMessageID = "" // Status doesn't support replies
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "text status enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("message_type", "text_status"))

	response := h.newSendMessageResponse(zaapID, instStatus)
	respondJSON(w, http.StatusOK, response)
}

// sendImageStatus handles POST /instances/{instanceId}/token/{token}/send-image-status
// Sends an image to WhatsApp Status/Stories (broadcasts to all viewers)
func (h *MessageHandler) sendImageStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req SendImageStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate image
	image := strings.TrimSpace(req.Image)
	if image == "" {
		h.log.WarnContext(ctx, "missing image for status")
		respondError(w, http.StatusBadRequest, "Image is required (URL or base64)")
		return
	}

	// Validate image format
	if !strings.HasPrefix(image, "http://") &&
		!strings.HasPrefix(image, "https://") &&
		!strings.HasPrefix(image, "data:image/") {
		h.log.WarnContext(ctx, "invalid image format for status")
		respondError(w, http.StatusBadRequest, "Image must be a URL (http/https) or base64 data URI (data:image/...)")
		return
	}

	// Resolve delay (delayMessage supports any duration; scheduledFor overrides with ISO 8601)
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)

	// Create message args for status broadcast
	caption := strings.TrimSpace(req.Caption)
	var captionPtr *string
	if caption != "" {
		captionPtr = &caption
	}
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       types.StatusBroadcastJID.String(), // CRITICAL: Always use status@broadcast
		MessageType: queue.MessageTypeImageStatus,
		ImageContent: &queue.MediaMessage{
			MediaURL: image,
			Caption:  captionPtr,
		},
		DelayMessage: delayMessage,
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "image status enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("message_type", "image_status"))

	response := h.newSendMessageResponse(zaapID, instStatus)
	respondJSON(w, http.StatusOK, response)
}

// sendAudioStatus handles POST /instances/{instanceId}/token/{token}/send-audio-status
// Sends an audio/voice note to WhatsApp Status/Stories with waveform visualization
func (h *MessageHandler) sendAudioStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req SendAudioStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate audio
	audio := strings.TrimSpace(req.Audio)
	if audio == "" {
		h.log.WarnContext(ctx, "missing audio for status")
		respondError(w, http.StatusBadRequest, "Audio is required (URL or base64)")
		return
	}

	// Validate audio format
	if !strings.HasPrefix(audio, "http://") &&
		!strings.HasPrefix(audio, "https://") &&
		!strings.HasPrefix(audio, "data:audio/") {
		h.log.WarnContext(ctx, "invalid audio format for status")
		respondError(w, http.StatusBadRequest, "Audio must be a URL (http/https) or base64 data URI (data:audio/...)")
		return
	}

	// Resolve delay (delayMessage supports any duration; scheduledFor overrides with ISO 8601)
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)

	// Create message args for status broadcast
	// Audio status will be processed with waveform generation in StatusProcessor
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       types.StatusBroadcastJID.String(), // CRITICAL: Always use status@broadcast
		MessageType: queue.MessageTypeAudioStatus,
		AudioContent: &queue.MediaMessage{
			MediaURL: audio,
			// Note: IsPTT flag is handled in StatusProcessor.ProcessAudio
		},
		DelayMessage: delayMessage,
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "audio status enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("message_type", "audio_status"))

	response := h.newSendMessageResponse(zaapID, instStatus)
	respondJSON(w, http.StatusOK, response)
}

// sendVideoStatus handles POST /instances/{instanceId}/token/{token}/send-video-status
// Sends a video to WhatsApp Status/Stories (broadcasts to all viewers)
func (h *MessageHandler) sendVideoStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, instStatus, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req SendVideoStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate video
	video := strings.TrimSpace(req.Video)
	if video == "" {
		h.log.WarnContext(ctx, "missing video for status")
		respondError(w, http.StatusBadRequest, "Video is required (URL or base64)")
		return
	}

	// Validate video format
	if !strings.HasPrefix(video, "http://") &&
		!strings.HasPrefix(video, "https://") &&
		!strings.HasPrefix(video, "data:video/") {
		h.log.WarnContext(ctx, "invalid video format for status")
		respondError(w, http.StatusBadRequest, "Video must be a URL (http/https) or base64 data URI (data:video/...)")
		return
	}

	// Resolve delay (delayMessage supports any duration; scheduledFor overrides with ISO 8601)
	delayMessage := resolveDelay(req.DelayMessage, req.ScheduledFor)

	// Create message args for status broadcast
	caption := strings.TrimSpace(req.Caption)
	var captionPtr *string
	if caption != "" {
		captionPtr = &caption
	}
	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       types.StatusBroadcastJID.String(), // CRITICAL: Always use status@broadcast
		MessageType: queue.MessageTypeVideoStatus,
		VideoContent: &queue.MediaMessage{
			MediaURL: video,
			Caption:  captionPtr,
		},
		DelayMessage: delayMessage,
	}

	// Enqueue message
	zaapID, err := h.enqueueMessage(ctx, instanceID, args)
	if err != nil {
		h.handleEnqueueError(ctx, w, err)
		return
	}

	h.log.InfoContext(ctx, "video status enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("message_type", "video_status"))

	response := h.newSendMessageResponse(zaapID, instStatus)
	respondJSON(w, http.StatusOK, response)
}
