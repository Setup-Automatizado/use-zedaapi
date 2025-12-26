package statuscache

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"go.mau.fi/whatsmeow/api/internal/config"
)

// Interceptor intercepts receipt events and caches them
type Interceptor struct {
	service Service
	cfg     *config.Config
	log     *slog.Logger
}

// NewInterceptor creates a new status cache interceptor
func NewInterceptor(service Service, cfg *config.Config, log *slog.Logger) *Interceptor {
	return &Interceptor{
		service: service,
		cfg:     cfg,
		log:     log.With(slog.String("component", "statuscache_interceptor")),
	}
}

// MessageStatusPayload represents the Z-API message status webhook payload
// This must match the structure from transform/zapi/schemas.go
type MessageStatusPayload struct {
	Status         string   `json:"status"`
	Ids            []string `json:"ids"`
	Momment        int64    `json:"momment"`
	Phone          string   `json:"phone"`
	ChatLid        string   `json:"chatLid,omitempty"`
	ChatPN         string   `json:"chat_pn,omitempty"`
	ParticipantLid string   `json:"participantLid,omitempty"`
	ParticipantPN  string   `json:"participant_pn,omitempty"`
	IsGroup        bool     `json:"isGroup"`
	InstanceId     string   `json:"instanceId"`
	Type           string   `json:"type"`
	FromMe         bool     `json:"fromMe"`
	WaitingMessage bool     `json:"waitingMessage,omitempty"`
	SenderDevice   int      `json:"senderDevice,omitempty"`
}

// IsEnabled returns whether the interceptor is enabled
func (i *Interceptor) IsEnabled() bool {
	return i.cfg.StatusCache.Enabled
}

// ShouldIntercept checks if the event should be intercepted
func (i *Interceptor) ShouldIntercept(eventType string) bool {
	if !i.IsEnabled() {
		return false
	}
	// Only intercept receipt events
	return eventType == "receipt"
}

// InterceptAndCache intercepts a receipt event payload and caches it
// Returns true if webhook should be suppressed, false otherwise
func (i *Interceptor) InterceptAndCache(ctx context.Context, instanceID string, eventType string, payload []byte) (suppress bool, err error) {
	if !i.ShouldIntercept(eventType) {
		return false, nil
	}

	var statusPayload MessageStatusPayload
	if err := json.Unmarshal(payload, &statusPayload); err != nil {
		i.log.Warn("failed to unmarshal status payload",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		return false, nil // Don't suppress on parse error, let webhook continue
	}

	// Check if status type should be cached
	if !i.shouldCacheStatusType(statusPayload.Status) {
		return false, nil
	}

	// Check scope (groups vs direct)
	if !i.shouldCacheScope(statusPayload.IsGroup) {
		return false, nil
	}

	// Build status event
	event := &StatusEvent{
		InstanceID:  instanceID,
		MessageIDs:  statusPayload.Ids,
		Status:      statusPayload.Status,
		Phone:       statusPayload.Phone,
		Participant: statusPayload.ParticipantPN,
		IsGroup:     statusPayload.IsGroup,
		GroupID:     extractGroupID(statusPayload.Phone, statusPayload.IsGroup),
		Timestamp:   statusPayload.Momment,
		Device:      statusPayload.SenderDevice,
	}

	// If no participant PN, use chat PN for direct chats
	if event.Participant == "" && !event.IsGroup {
		event.Participant = statusPayload.ChatPN
	}
	if event.Participant == "" {
		event.Participant = statusPayload.Phone
	}

	// Cache the event
	suppress, err = i.service.CacheStatusEvent(ctx, event, payload)
	if err != nil {
		i.log.Warn("failed to cache status event",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		return false, nil // Don't suppress on cache error
	}

	if suppress {
		i.log.Debug("status event cached and webhook suppressed",
			slog.String("instance_id", instanceID),
			slog.String("status", statusPayload.Status),
			slog.Int("message_count", len(statusPayload.Ids)),
		)
	}

	return suppress, nil
}

// shouldCacheStatusType checks if the status type should be cached
func (i *Interceptor) shouldCacheStatusType(status string) bool {
	for _, t := range i.cfg.StatusCache.Types {
		if t == status {
			return true
		}
	}
	return false
}

// shouldCacheScope checks if the scope (group/direct) should be cached
func (i *Interceptor) shouldCacheScope(isGroup bool) bool {
	for _, scope := range i.cfg.StatusCache.Scope {
		if scope == "groups" && isGroup {
			return true
		}
		if scope == "direct" && !isGroup {
			return true
		}
	}
	return false
}

// extractGroupID extracts the group ID from the phone field for group messages
// WhatsApp group JIDs are in the format: 120363182823169824@g.us or 120363182823169824-1234567890@g.us
func extractGroupID(phone string, isGroup bool) string {
	if !isGroup || phone == "" {
		return ""
	}

	// Remove @g.us suffix if present
	groupID := phone
	if idx := findChar(groupID, '@'); idx >= 0 {
		groupID = groupID[:idx]
	}

	// Remove any prefix before hyphen (owner's JID in some formats)
	// e.g., "120363182823169824-1234567890" -> "120363182823169824"
	if idx := findChar(groupID, '-'); idx >= 0 && idx < len(groupID)-1 {
		// In some cases the format is GROUP_ID-TIMESTAMP, keep the first part
		groupID = groupID[:idx]
	}

	return groupID
}

// findChar returns the index of the first occurrence of char in s, or -1 if not found
func findChar(s string, char byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == char {
			return i
		}
	}
	return -1
}

// DispatchWebhook implements WebhookDispatcher interface
// This is used during flush operations to send suppressed webhooks
type FlushDispatcher struct {
	deliverFn func(ctx context.Context, instanceID string, payload []byte) error
}

// NewFlushDispatcher creates a dispatcher that uses the provided function to deliver webhooks
func NewFlushDispatcher(deliverFn func(ctx context.Context, instanceID string, payload []byte) error) *FlushDispatcher {
	return &FlushDispatcher{deliverFn: deliverFn}
}

// DispatchWebhook sends a webhook payload
func (d *FlushDispatcher) DispatchWebhook(ctx context.Context, instanceID string, payload []byte) error {
	if d.deliverFn == nil {
		return nil
	}
	return d.deliverFn(ctx, instanceID, payload)
}

// CacheHitResult represents the result of checking cache for a message
type CacheHitResult struct {
	Found     bool
	Timestamp time.Time
}
