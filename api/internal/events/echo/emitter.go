package echo

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"

	"go.mau.fi/whatsmeow/api/internal/events/capture"
	eventtypes "go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// APIEchoMarker is used to identify API echo events in the event pipeline
const APIEchoMarker = "__api_echo__"

// Emitter emits API echo events for messages sent via the API
// This ensures that when a message is sent via API, a webhook is dispatched
// to notify the partner with fromMe=true and fromApi=true
type Emitter struct {
	instanceID  uuid.UUID
	eventRouter *capture.EventRouter
	log         *slog.Logger
	metrics     *observability.Metrics
	enabled     bool
	storeJID    string
}

// EmitterConfig holds configuration for the API echo emitter
type EmitterConfig struct {
	InstanceID  uuid.UUID
	EventRouter *capture.EventRouter
	Metrics     *observability.Metrics
	Enabled     bool
	StoreJID    string
}

// EchoRequest contains the information needed to emit an API echo event
type EchoRequest struct {
	// Instance identification (required)
	InstanceID uuid.UUID // WhatsApp instance UUID (from SendMessageArgs)
	StoreJID   string    // Connected phone JID (optional, for connectedPhone field)

	// Message identification
	WhatsAppMessageID string
	RecipientJID      types.JID
	Message           *waE2E.Message
	Timestamp         time.Time
	MessageType       string // text, image, audio, video, document, sticker, etc.
	MediaType         string // image, video, audio, document, sticker (for media messages)
	ZaapID            string // External ID for correlation
	HasMedia          bool
	MediaKey          string
	DirectPath        string
	FileSHA256        *string
	FileEncSHA256     *string
	MimeType          *string
	FileLength        *int64
}

// NewEmitter creates a new API echo emitter
func NewEmitter(ctx context.Context, cfg *EmitterConfig) *Emitter {
	log := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "api_echo_emitter"),
		slog.String("instance_id", cfg.InstanceID.String()),
	)

	return &Emitter{
		instanceID:  cfg.InstanceID,
		eventRouter: cfg.EventRouter,
		log:         log,
		metrics:     cfg.Metrics,
		enabled:     cfg.Enabled,
		storeJID:    cfg.StoreJID,
	}
}

// EmitEcho creates and routes an API echo event for a successfully sent message
// This should be called AFTER client.SendMessage() returns successfully
func (e *Emitter) EmitEcho(ctx context.Context, req *EchoRequest) error {
	if !e.enabled {
		e.log.Debug("API echo disabled, skipping",
			slog.String("zaap_id", req.ZaapID))
		return nil
	}

	if e.eventRouter == nil {
		e.log.Warn("no event router configured, skipping API echo",
			slog.String("zaap_id", req.ZaapID))
		return nil
	}

	// Build metadata for the echo event
	metadata := map[string]string{
		"message_id":      req.WhatsAppMessageID,
		"from_me":         "true",
		"from_api":        "true",
		"api_echo":        "true",
		"chat":            req.RecipientJID.String(),
		"phone":           extractPhone(req.RecipientJID),
		"is_group":        strconv.FormatBool(req.RecipientJID.Server == types.GroupServer),
		"timestamp":       strconv.FormatInt(req.Timestamp.Unix(), 10),
		"zaap_id":         req.ZaapID,
		"message_type":    req.MessageType,
		"api_echo_marker": APIEchoMarker,
	}

	// Add store JID if available (connected phone number)
	// Prefer req.StoreJID (instance-specific) over e.storeJID (global fallback)
	storeJID := req.StoreJID
	if storeJID == "" {
		storeJID = e.storeJID
	}
	if storeJID != "" {
		metadata["store_jid"] = storeJID
		// Extract phone number from store JID for connectedPhone field
		if jid, err := types.ParseJID(storeJID); err == nil {
			metadata["connected_phone"] = jid.User
		}
	}

	// Add media type if applicable
	if req.MediaType != "" {
		metadata["media_type"] = req.MediaType
	}

	// Determine instance ID - use req.InstanceID (required) or fallback to e.instanceID
	instanceID := req.InstanceID
	if instanceID == uuid.Nil {
		instanceID = e.instanceID
	}

	// Create the internal event
	event := &eventtypes.InternalEvent{
		InstanceID: instanceID,
		EventID:    uuid.New(),
		EventType:  "message",
		SourceLib:  eventtypes.SourceLibAPI,
		Metadata:   metadata,
		RawPayload: req.Message,
		CapturedAt: time.Now(),
		HasMedia:   req.HasMedia,
	}

	// Add media information if present
	if req.HasMedia {
		event.MediaKey = req.MediaKey
		event.DirectPath = req.DirectPath
		event.FileSHA256 = req.FileSHA256
		event.FileEncSHA256 = req.FileEncSHA256
		event.MimeType = req.MimeType
		event.FileLength = req.FileLength
		event.MediaType = req.MediaType
	}

	// Route the event through the event pipeline
	if err := e.eventRouter.RouteEvent(ctx, event); err != nil {
		e.log.Warn("failed to route API echo event",
			slog.String("error", err.Error()),
			slog.String("zaap_id", req.ZaapID),
			slog.String("whatsapp_message_id", req.WhatsAppMessageID),
			slog.String("instance_id", instanceID.String()))

		if e.metrics != nil {
			e.metrics.EventsFailed.WithLabelValues(
				instanceID.String(),
				"message",
				"api_echo_route_failed",
			).Inc()
		}

		return err
	}

	e.log.Info("API echo event emitted",
		slog.String("zaap_id", req.ZaapID),
		slog.String("whatsapp_message_id", req.WhatsAppMessageID),
		slog.String("message_type", req.MessageType),
		slog.String("recipient", req.RecipientJID.String()),
		slog.String("instance_id", instanceID.String()),
		slog.Bool("has_media", req.HasMedia))

	if e.metrics != nil {
		e.metrics.EventsCaptured.WithLabelValues(
			instanceID.String(),
			"message",
			string(eventtypes.SourceLibAPI),
		).Inc()
	}

	return nil
}

// SetEnabled enables or disables the emitter
func (e *Emitter) SetEnabled(enabled bool) {
	e.enabled = enabled
}

// SetStoreJID updates the store JID (connected phone number)
func (e *Emitter) SetStoreJID(storeJID string) {
	e.storeJID = storeJID
}

// IsEnabled returns whether the emitter is enabled
func (e *Emitter) IsEnabled() bool {
	return e.enabled
}

// extractPhone extracts the phone number from a JID
func extractPhone(jid types.JID) string {
	user := jid.User
	// Remove any device suffix if present
	if idx := strings.Index(user, ":"); idx != -1 {
		user = user[:idx]
	}
	return user
}
