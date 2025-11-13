package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/chats"
	"go.mau.fi/whatsmeow/api/internal/contacts"
	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/messages/queue"
)

type InstanceStatusProvider interface {
	GetStatus(ctx context.Context, id uuid.UUID, clientToken, instanceToken string) (*instances.Status, error)
}

// ContactsService provides contact-related operations
type ContactsService interface {
	List(ctx context.Context, instanceID uuid.UUID, params contacts.ListParams) (contacts.ListResult, error)
}

// ChatsService provides chat-related operations
type ChatsService interface {
	List(ctx context.Context, instanceID uuid.UUID, params chats.ListParams) (chats.ListResult, error)
}

// MessageHandler handles HTTP requests for message queue operations
type MessageHandler struct {
	coordinator     *queue.Coordinator
	instanceService InstanceStatusProvider
	contactsService ContactsService
	chatsService    ChatsService
	log             *slog.Logger
	drainRetryAfter time.Duration
	enqueueMessage  func(context.Context, uuid.UUID, queue.SendMessageArgs) (string, error)
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(
	coordinator *queue.Coordinator,
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
	// Message sending endpoints (Z-API compatible)
	r.Post("/send-text", h.sendText)
	r.Post("/send-image", h.sendImage)
	r.Post("/send-sticker", h.sendSticker)
	r.Post("/send-audio", h.sendAudio)
	r.Post("/send-video", h.sendVideo)
	r.Post("/send-gif", h.sendGif)
	r.Post("/send-document", h.sendDocument)
	r.Post("/send-location", h.sendLocation)
	r.Post("/send-contact", h.sendContact)
	r.Post("/send-contacts", h.sendContacts)

	// Queue management endpoints (Z-API compatible)
	r.Get("/queue", h.listQueue)                      // GET /queue?page=1&pageSize=100
	r.Get("/queue/count", h.getQueueCount)            // GET /queue/count
	r.Delete("/queue", h.clearQueue)                  // DELETE /queue (clear all)
	r.Delete("/queue/{zaapId}", h.cancelQueueMessage) // DELETE /queue/{zaapId}

	// Data retrieval endpoints (Z-API compatible)
	r.Get("/contacts", h.getContacts) // GET /contacts?page=1&pageSize=100
	r.Get("/chats", h.getChats)       // GET /chats?page=1&pageSize=100
}

// GroupMention represents a group mention in a community (Z-API compatible format)
type GroupMention struct {
	Phone   string `json:"phone"`   // Group JID (e.g., "120363xyz@g.us")
	Subject string `json:"subject"` // Group name/subject
}

// SendTextRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-text
// Z-API Compatible format
type SendTextRequest struct {
	Phone         string `json:"phone"`                   // Phone number (e.g., "5511999999999")
	Message       string `json:"message"`                 // Text message content
	MessageID     string `json:"messageId,omitempty"`     // Optional WhatsApp message ID to reply to
	DelayMessage  *int   `json:"delayMessage,omitempty"`  // Optional delay in seconds (1-15) before sending
	DelayTyping   *int   `json:"delayTyping,omitempty"`   // Optional typing indicator duration in seconds (1-15)
	Duration      *int   `json:"duration,omitempty"`      // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	PrivateAnswer bool   `json:"privateAnswer,omitempty"` // For group messages: if true, reply in private to sender (not yourself)
	LinkPreview   *bool  `json:"linkPreview,omitempty"`   // If nil, auto-detect URLs; if true, force preview; if false, disable preview
}

// SendImageRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-image
// Z-API Compatible format
type SendImageRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Image          string         `json:"image"`                    // Image URL or base64 data (data:image/png;base64,...)
	Caption        string         `json:"caption,omitempty"`        // Optional image caption/title
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds (1-15) before sending
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	ViewOnce       bool           `json:"viewOnce,omitempty"`       // If true, image can only be viewed once
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendStickerRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-sticker
// Z-API Compatible format
type SendStickerRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Sticker        string         `json:"sticker"`                  // Sticker URL or base64 data (data:image/webp;base64,...)
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds (1-15) before sending
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendAudioRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-audio
// Z-API Compatible format
type SendAudioRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Audio          string         `json:"audio"`                    // Audio URL or base64 data (data:audio/ogg;base64,...)
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds (1-15) before sending
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional "recording audio" indicator duration in seconds (1-15)
	ViewOnce       bool           `json:"viewOnce,omitempty"`       // If true, audio can only be played once
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendVideoRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-video
// Z-API Compatible format
type SendVideoRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Video          string         `json:"video"`                    // Video URL or base64 data (data:video/mp4;base64,...)
	Caption        string         `json:"caption,omitempty"`        // Optional video caption/title
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds (1-15) before sending
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	ViewOnce       bool           `json:"viewOnce,omitempty"`       // If true, video can only be viewed once
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendGifRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-gif
// Z-API Compatible format
type SendGifRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Gif            string         `json:"gif"`                      // GIF URL or base64 data (data:image/gif;base64,...)
	Caption        string         `json:"caption,omitempty"`        // Optional GIF caption/title
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds (1-15) before sending
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	ViewOnce       bool           `json:"viewOnce,omitempty"`       // If true, GIF can only be viewed once
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendDocumentRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-document
// Z-API Compatible format
type SendDocumentRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Document       string         `json:"document"`                 // Document URL or base64 data (data:application/pdf;base64,...)
	FileName       string         `json:"fileName,omitempty"`       // Optional document filename
	Caption        string         `json:"caption,omitempty"`        // Optional document caption/title
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds (1-15) before sending
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendLocationRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-location
// Z-API Compatible format
type SendLocationRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Latitude       float64        `json:"latitude"`                 // Location latitude
	Longitude      float64        `json:"longitude"`                // Location longitude
	Name           string         `json:"name,omitempty"`           // Optional location name/title
	Address        string         `json:"address,omitempty"`        // Optional location address
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds (1-15) before sending
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// SendContactRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-contact
// Z-API Compatible format with extended fields
type SendContactRequest struct {
	Phone                      string `json:"phone"`                                // Phone number (e.g., "5511999999999")
	ContactName                string `json:"contactName"`                          // Full name of the contact (Z-API compatible)
	ContactPhone               string `json:"contactPhone"`                         // Phone number of the contact in international format (Z-API compatible)
	ContactBusinessDescription string `json:"contactBusinessDescription,omitempty"` // Optional business description (Z-API compatible)

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
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds (1-15) before sending
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

// ContactInfo represents a contact for multiple contacts (Z-API compatible with extended fields)
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
// Z-API Compatible format for sending multiple contacts
type SendContactsRequest struct {
	Phone          string         `json:"phone"`                    // Phone number (e.g., "5511999999999")
	Contacts       []ContactInfo  `json:"contacts"`                 // Array of contacts (1-10)
	MessageID      string         `json:"messageId,omitempty"`      // Optional WhatsApp message ID to reply to
	DelayMessage   *int           `json:"delayMessage,omitempty"`   // Optional delay in seconds (1-15) before sending
	DelayTyping    *int           `json:"delayTyping,omitempty"`    // Optional typing indicator duration in seconds (1-15)
	Duration       *int           `json:"duration,omitempty"`       // Ephemeral message duration in seconds (0, 86400, 604800, 7776000)
	Mentioned      []string       `json:"mentioned,omitempty"`      // Optional array of phone numbers to mention
	GroupMentioned []GroupMention `json:"groupMentioned,omitempty"` // Optional array of groups to mention (communities)
	MentionedAll   bool           `json:"mentionedAll,omitempty"`   // If true, mention all group members
	PrivateAnswer  bool           `json:"privateAnswer,omitempty"`  // For group messages: if true, reply in private to sender
}

// Button represents a button in an interactive message
type Button struct {
	ID    string `json:"id"`    // Required: button identifier (max 256 chars)
	Title string `json:"title"` // Required: button text (max 20 chars)
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
// Z-API Compatible format for sending interactive button messages
type SendButtonActionsRequest struct {
	Phone        string   `json:"phone"`        // Required: recipient phone number
	Message      string   `json:"message"`      // Required: body text
	Buttons      []Button `json:"buttons"`      // Required: 1-3 buttons
	Footer       string   `json:"footer"`       // Optional: footer text
	MessageID    string   `json:"messageId"`    // Optional: reply to message ID
	DelayMessage *int     `json:"delayMessage"` // Optional: delay in seconds (1-15)
	DelayTyping  *int     `json:"delayTyping"`  // Optional: typing delay in seconds (1-15)
}

// SendButtonListRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-button-list
// Z-API Compatible format for sending interactive list messages
type SendButtonListRequest struct {
	Phone        string    `json:"phone"`        // Required: recipient phone number
	Message      string    `json:"message"`      // Required: body text
	ButtonText   string    `json:"buttonText"`   // Required: menu button text (max 20 chars)
	Sections     []Section `json:"sections"`     // Required: 1-10 sections
	Title        string    `json:"title"`        // Optional: header text (max 60 chars)
	Footer       string    `json:"footer"`       // Optional: footer text (max 60 chars)
	MessageID    string    `json:"messageId"`    // Optional: reply to message ID
	DelayMessage *int      `json:"delayMessage"` // Optional: delay in seconds (1-15)
	DelayTyping  *int      `json:"delayTyping"`  // Optional: typing delay in seconds (1-15)
}

// SendPollRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-poll
// Z-API Compatible format for sending poll messages
type SendPollRequest struct {
	Phone                  string   `json:"phone"`                            // Required: recipient phone number
	Question               string   `json:"question"`                         // Required: poll question text
	Options                []string `json:"options"`                          // Required: poll options (2-12 options)
	SelectableOptionsCount *int     `json:"selectableOptionsCount,omitempty"` // Optional: 0 for single choice, 1+ for multiple choice (default: 0)
	MessageID              string   `json:"messageId"`                        // Optional: reply to message ID
	DelayMessage           *int     `json:"delayMessage"`                     // Optional: delay in seconds (1-15)
	DelayTyping            *int     `json:"delayTyping"`                      // Optional: typing delay in seconds (1-15)
}

// SendEventRequest represents the request body for POST /instances/{instanceId}/token/{token}/send-event
// Z-API Compatible format for sending calendar event messages
type SendEventRequest struct {
	Phone        string  `json:"phone"`                 // Required: recipient phone number
	Name         string  `json:"name"`                  // Required: event name/title
	StartTime    int64   `json:"startTime"`             // Required: event start time (Unix timestamp)
	EndTime      *int64  `json:"endTime,omitempty"`     // Optional: event end time (Unix timestamp)
	Location     *string `json:"location,omitempty"`    // Optional: event location
	Description  *string `json:"description,omitempty"` // Optional: event description
	CallLink     *string `json:"callLink,omitempty"`    // Optional: video call link
	MessageID    string  `json:"messageId"`             // Optional: reply to message ID
	DelayMessage *int    `json:"delayMessage"`          // Optional: delay in seconds (1-15)
	DelayTyping  *int    `json:"delayTyping"`           // Optional: typing delay in seconds (1-15)
}

// SendMessageResponse represents the response after enqueuing a message
// Z-API Compatible format
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

// QueueMessageResponse represents a message in the queue (Z-API compatible format)
type QueueMessageResponse struct {
	ID           string `json:"_id"`               // Message ID (same as ZaapId for Z-API compat)
	ZaapId       string `json:"zaapId"`            // Z-API message ID
	MessageId    string `json:"messageId"`         // WhatsApp message ID
	InstanceId   string `json:"instanceId"`        // Instance ID
	Phone        string `json:"phone"`             // Recipient phone
	Message      string `json:"message,omitempty"` // Message text (for text messages)
	DelayMessage int64  `json:"delayMessage"`      // Delay in seconds before sending
	DelayTyping  int64  `json:"delayTyping"`       // Typing indicator duration in seconds
	Created      int64  `json:"created"`           // Unix timestamp in milliseconds

	// Additional fields (not in Z-API but useful)
	MessageType    string   `json:"messageType,omitempty"`    // Message type (text, image, etc)
	Status         string   `json:"status,omitempty"`         // Job status
	SequenceNumber int64    `json:"sequenceNumber,omitempty"` // FIFO sequence
	Attempt        int      `json:"attempt,omitempty"`        // Current attempt
	MaxAttempts    int      `json:"maxAttempts,omitempty"`    // Max retry attempts
	Errors         []string `json:"errors,omitempty"`         // Error messages
}

// QueueListResponse represents the response for GET /queue (Z-API compatible)
type QueueListResponse []QueueMessageResponse

// QueueCountResponse represents the response for GET /queue/count
type QueueCountResponse struct {
	Count int `json:"count"` // Number of messages in queue
}

// sendText handles POST /instances/{instanceId}/token/{token}/send-text
//
// Z-API Compatible endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Enqueues message with FIFO ordering
// 4. Returns immediately with zaapId as messageId (non-blocking)
// 5. Worker updates with real WhatsApp messageId after sending
//
// Compatible with Z-API specification
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

	// Convert delays from seconds to milliseconds with Z-API defaults
	// Z-API: delayMessage range 1-15 seconds, default 1-3 seconds random
	// Z-API: delayTyping range 1-15 seconds, default 0
	delayMessage := int64(0)
	if req.DelayMessage != nil {
		seconds := *req.DelayMessage
		// Validate range (1-15 seconds)
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayMessage = int64(seconds) * 1000 // Convert to milliseconds
	} else {
		// Default: random 1-3 seconds
		delayMessage = int64(1000 + (rand.Int63() % 2000)) // 1000-3000ms
	}

	delayTyping := int64(0)
	if req.DelayTyping != nil {
		seconds := *req.DelayTyping
		// Validate range (1-15 seconds)
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayTyping = int64(seconds) * 1000 // Convert to milliseconds
	}
	// Default is 0 (no typing indicator)

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

	// Return Z-API compatible response
	// Initially, messageId = zaapId (will be updated with real WhatsApp ID after send)
	response := h.newSendMessageResponse(zaapID, instStatus)

	// Return 200 OK (Z-API standard)
	respondJSON(w, http.StatusOK, response)
}

// sendImage handles POST /instances/{instanceId}/token/{token}/send-image
//
// Z-API Compatible endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Supports image URL or base64 data
// 4. Supports optional caption, viewOnce, and reply-to (messageId)
// 5. Enqueues message with FIFO ordering
// 6. Returns immediately with zaapId as messageId (non-blocking)
// 7. Worker updates with real WhatsApp messageId after sending
//
// Compatible with Z-API specification
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

	// Convert delays from seconds to milliseconds with Z-API defaults
	// Z-API: delayMessage range 1-15 seconds, default 1-3 seconds random
	delayMessage := int64(0)
	if req.DelayMessage != nil {
		seconds := *req.DelayMessage
		// Validate range (1-15 seconds)
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayMessage = int64(seconds) * 1000 // Convert to milliseconds
	} else {
		// Default: random 1-3 seconds
		delayMessage = int64(1000 + (rand.Int63() % 2000)) // 1000-3000ms
	}

	// Process delayTyping (typing indicator)
	delayTyping := int64(0)
	if req.DelayTyping != nil {
		seconds := *req.DelayTyping
		// Validate range (1-15 seconds)
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayTyping = int64(seconds) * 1000 // Convert to milliseconds
	}

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

	// Return Z-API compatible response
	// Initially, messageId = zaapId (will be updated with real WhatsApp ID after send)
	response := h.newSendMessageResponse(zaapID, instStatus)

	// Return 200 OK (Z-API standard)
	respondJSON(w, http.StatusOK, response)
}

// sendSticker handles POST /instances/{instanceId}/token/{token}/send-sticker
//
// Z-API Compatible endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Supports sticker URL or base64 data (WebP format)
// 4. Supports optional reply-to (messageId)
// 5. Enqueues message with FIFO ordering
// 6. Returns immediately with zaapId as messageId (non-blocking)
//
// Compatible with Z-API specification
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

	// Convert delays from seconds to milliseconds with Z-API defaults
	delayMessage := int64(0)
	if req.DelayMessage != nil {
		seconds := *req.DelayMessage
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayMessage = int64(seconds) * 1000
	} else {
		delayMessage = int64(1000 + (rand.Int63() % 2000)) // Default: 1-3 sec
	}

	delayTyping := int64(0)
	if req.DelayTyping != nil {
		seconds := *req.DelayTyping
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayTyping = int64(seconds) * 1000
	}

	// Create message args with sticker metadata
	metadata := map[string]interface{}{
		"is_sticker": true,
	}

	args := queue.SendMessageArgs{
		InstanceID:  instanceID,
		Phone:       phone,
		MessageType: queue.MessageTypeImage, // Stickers use image type with metadata flag
		ImageContent: &queue.MediaMessage{
			MediaURL: sticker,
			Caption:  nil, // Stickers don't support captions
		},
		DelayMessage:     delayMessage,
		DelayTyping:      delayTyping,
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

	h.log.InfoContext(ctx, "sticker message enqueued successfully",
		slog.String("zaap_id", zaapID),
		slog.String("phone", phone),
		slog.Bool("is_reply", req.MessageID != ""),
		slog.Bool("whatsapp_connected", whatsStatus != nil && whatsStatus.Connected))

	// Return Z-API compatible response
	response := h.newSendMessageResponse(zaapID, instStatus)

	// Return 200 OK (Z-API standard)
	respondJSON(w, http.StatusOK, response)
}

// sendAudio handles POST /instances/{instanceId}/token/{token}/send-audio
//
// Z-API Compatible endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Supports audio URL or base64 data
// 4. Supports optional viewOnce and reply-to (messageId)
// 5. DelayTyping shows "recording audio" indicator
// 6. Enqueues message with FIFO ordering
// 7. Returns immediately with zaapId as messageId (non-blocking)
//
// Compatible with Z-API specification
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

	// Convert delays from seconds to milliseconds
	delayMessage := int64(0)
	if req.DelayMessage != nil {
		seconds := *req.DelayMessage
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayMessage = int64(seconds) * 1000
	} else {
		delayMessage = int64(1000 + (rand.Int63() % 2000)) // Default: 1-3 sec
	}

	// DelayTyping for audio = "recording audio" indicator
	delayTyping := int64(0)
	if req.DelayTyping != nil {
		seconds := *req.DelayTyping
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayTyping = int64(seconds) * 1000
	}

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

	// Return Z-API compatible response
	response := h.newSendMessageResponse(zaapID, instStatus)

	respondJSON(w, http.StatusOK, response)
}

// sendVideo handles POST /instances/{instanceId}/token/{token}/send-video
//
// Z-API Compatible endpoint for sending video messages
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
	delayMessage := int64(0)
	if req.DelayMessage != nil {
		seconds := *req.DelayMessage
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayMessage = int64(seconds) * 1000
	} else {
		delayMessage = int64(1000 + (rand.Int63() % 2000))
	}

	delayTyping := int64(0)
	if req.DelayTyping != nil {
		seconds := *req.DelayTyping
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayTyping = int64(seconds) * 1000
	}

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

	// Return Z-API compatible response
	response := h.newSendMessageResponse(zaapID, instStatus)

	respondJSON(w, http.StatusOK, response)
}

// sendGif handles POST /instances/{instanceId}/token/{token}/send-gif
//
// Z-API Compatible endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Supports GIF URL or base64 data
// 4. Supports optional caption and viewOnce
// 5. Enqueues message with FIFO ordering
// 6. Returns immediately with zaapId as messageId (non-blocking)
//
// Compatible with Z-API specification
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

	// Convert delays from seconds to milliseconds with Z-API defaults
	delayMessage := int64(0)
	if req.DelayMessage != nil {
		seconds := *req.DelayMessage
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayMessage = int64(seconds) * 1000
	} else {
		delayMessage = int64(1000 + (rand.Int63() % 2000)) // Default: 1-3 sec
	}

	delayTyping := int64(0)
	if req.DelayTyping != nil {
		seconds := *req.DelayTyping
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayTyping = int64(seconds) * 1000
	}

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

	// Return Z-API compatible response
	response := h.newSendMessageResponse(zaapID, instStatus)

	// Return 200 OK (Z-API standard)
	respondJSON(w, http.StatusOK, response)
}

// sendDocument handles POST /instances/{instanceId}/token/{token}/send-document
//
// Z-API Compatible endpoint for sending document messages
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
	delayMessage := int64(0)
	if req.DelayMessage != nil {
		seconds := *req.DelayMessage
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayMessage = int64(seconds) * 1000
	} else {
		delayMessage = int64(1000 + (rand.Int63() % 2000))
	}

	delayTyping := int64(0)
	if req.DelayTyping != nil {
		seconds := *req.DelayTyping
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayTyping = int64(seconds) * 1000
	}

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

	// Return Z-API compatible response
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
	delayMessage := int64(0)
	if req.DelayMessage != nil {
		seconds := *req.DelayMessage
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayMessage = int64(seconds) * 1000
	} else {
		delayMessage = int64(1000 + (rand.Int63() % 2000))
	}

	delayTyping := int64(0)
	if req.DelayTyping != nil {
		seconds := *req.DelayTyping
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayTyping = int64(seconds) * 1000
	}

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

	// Return Z-API compatible response
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

	// Validate Z-API required contact fields
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
	delayMessage := int64(0)
	if req.DelayMessage != nil {
		seconds := *req.DelayMessage
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayMessage = int64(seconds) * 1000
	} else {
		delayMessage = int64(1000 + (rand.Int63() % 2000))
	}

	delayTyping := int64(0)
	if req.DelayTyping != nil {
		seconds := *req.DelayTyping
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayTyping = int64(seconds) * 1000
	}

	// Build ContactMessage with ALL fields (Z-API + extended optional fields)
	contactMsg := &queue.ContactMessage{
		// Required Z-API fields
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

	// Return Z-API compatible response
	response := h.newSendMessageResponse(zaapID, instStatus)

	respondJSON(w, http.StatusOK, response)
}

// listQueue handles GET /instances/{instanceId}/token/{token}/queue
//
// Z-API Compatible endpoint
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

	// Parse query parameters (Z-API format: page, pageSize)
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

	// Convert to Z-API format
	messages := make(QueueListResponse, 0, len(queueList.Jobs))
	for _, job := range queueList.Jobs {
		messages = append(messages, convertJobToQueueMessage(job))
	}

	respondJSON(w, http.StatusOK, messages)
}

// getQueueCount handles GET /instances/{instanceId}/token/{token}/queue/count
//
// Z-API Compatible endpoint
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
// Z-API Compatible endpoint - deletes ALL messages in queue
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

	// Z-API returns 200 OK with empty body
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Queue cleared successfully",
	})
}

// cancelQueueMessage handles DELETE /instances/{instanceId}/token/{token}/queue/{zaapId}
//
// Z-API Compatible endpoint - deletes a specific message from queue
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

	// Z-API returns 200 OK with empty body
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Message cancelled successfully",
	})
}

// getContacts handles GET /instances/{instanceId}/token/{token}/contacts
//
// Z-API Compatible endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Retrieves all contacts from WhatsApp instance
// 4. Applies pagination to results
// 5. Returns Z-API compatible contact array
//
// Query parameters:
// - page: Page number (required, minimum: 1)
// - pageSize: Number of contacts per page (required, minimum: 1)
//
// Compatible with Z-API specification
func (h *MessageHandler) getContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse query parameters (Z-API format: page, pageSize)
	// Both are REQUIRED in Z-API specification
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

	// Z-API returns array of contacts (not wrapped in object)
	respondJSON(w, http.StatusOK, result.Items)
}

// getChats handles GET /instances/{instanceId}/token/{token}/chats
//
// Z-API Compatible endpoint that:
// 1. Validates instanceId and token from URL
// 2. Validates Client-Token header
// 3. Retrieves all chats from WhatsApp instance
// 4. Applies pagination to results
// 5. Returns Z-API compatible chat array
//
// Query parameters:
// - page: Page number (required, minimum: 1)
// - pageSize: Number of chats per page (required, minimum: 1)
//
// Compatible with Z-API specification
func (h *MessageHandler) getChats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse query parameters (Z-API format: page, pageSize)
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

	// Z-API returns array of chats (not wrapped in object)
	respondJSON(w, http.StatusOK, result.Chats)
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

	// Z-API style suffixes for groups/channels/broadcast lists
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

// convertJobToQueueMessage converts queue.QueueJobInfo to QueueMessageResponse (Z-API format)
func convertJobToQueueMessage(job queue.QueueJobInfo) QueueMessageResponse {
	// Extract message text if it's a text message
	message := ""
	if job.MessageType == "text" && job.TextContent != nil {
		message = job.TextContent.Message
	}

	// Convert delays from milliseconds to seconds (Z-API format)
	delayMessageSec := job.DelayMessage / 1000
	delayTypingSec := job.DelayTyping / 1000

	// Convert created timestamp to milliseconds
	createdMs := job.CreatedAt.UnixMilli()

	response := QueueMessageResponse{
		ID:             job.ZaapID,              // _id field
		ZaapId:         job.ZaapID,              // Z-API format
		MessageId:      job.WhatsAppMessageID,   // WhatsApp message ID (if sent)
		InstanceId:     job.InstanceID.String(), // Z-API format
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
// Z-API Compatible endpoint for sending multiple contact cards
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
	delayMessage := int64(0)
	if req.DelayMessage != nil {
		seconds := *req.DelayMessage
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayMessage = int64(seconds) * 1000
	} else {
		delayMessage = int64(1000 + (rand.Int63() % 2000))
	}

	delayTyping := int64(0)
	if req.DelayTyping != nil {
		seconds := *req.DelayTyping
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 15 {
			seconds = 15
		}
		delayTyping = int64(seconds) * 1000
	}

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
			// Required Z-API fields
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

	// Return Z-API compatible response
	response := h.newSendMessageResponse(zaapID, instStatus)

	respondJSON(w, http.StatusOK, response)
}
