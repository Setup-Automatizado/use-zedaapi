package zapi

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	eventctx "go.mau.fi/whatsmeow/api/internal/events/eventctx"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	watypes "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (t *Transformer) transformNewsletterAdminEvent(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*ReceivedCallback, error) {
	rawJID := strings.TrimSpace(event.Metadata["newsletter_jid"])
	if rawJID == "" {
		return nil, fmt.Errorf("missing newsletter id in metadata")
	}
	jid, err := parseJID(rawJID)
	if err != nil {
		jid = mustParseJIDOrFallback(rawJID)
	}
	phone := conversationIdentifierFromJID(jid)
	if phone == "" {
		phone = sanitizeConversationFallback(rawJID)
	}

	callback := &ReceivedCallback{
		Type:           "ReceivedCallback",
		InstanceID:     event.InstanceID.String(),
		MessageID:      event.EventID.String(),
		Phone:          phone,
		FromMe:         false,
		ConnectedPhone: t.connectedPhone,
		Momment:        eventTimestampMillis(event),
		Status:         "RECEIVED",
		IsNewsletter:   true,
		Notification:   mapNewsletterNotification(event),
	}

	role := strings.ToUpper(strings.TrimSpace(event.Metadata["newsletter_role"]))
	switch event.EventType {
	case "newsletter_join", "newsletter_leave":
		callback.NotificationParameters = []string{t.connectedPhone, role}
		callback.ParticipantPhone = t.connectedPhone
	case "newsletter_mute_change":
		if mute := strings.ToUpper(strings.TrimSpace(event.Metadata["newsletter_mute"])); mute != "" {
			callback.NotificationParameters = []string{mute}
		}
	default:
		callback.NotificationParameters = nil
	}

	if name := strings.TrimSpace(event.Metadata["newsletter_name"]); name != "" {
		callback.ChatName = name
	}
	if photo := strings.TrimSpace(event.Metadata["newsletter_photo"]); photo != "" {
		callback.Photo = photo
	}

	return callback, nil
}

func (t *Transformer) transformNewsletterLiveUpdateEvent(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*ReceivedCallback, error) {
	live, ok := event.RawPayload.(*events.NewsletterLiveUpdate)
	if !ok {
		return nil, fmt.Errorf("invalid newsletter live payload type")
	}
	if len(live.Messages) == 0 {
		return nil, fmt.Errorf("newsletter live update without messages")
	}

	msg := live.Messages[0]
	phone := conversationIdentifierFromJID(live.JID)
	if phone == "" {
		phone = sanitizeConversationFallback(live.JID.String())
	}

	callback := &ReceivedCallback{
		Type:                   "ReceivedCallback",
		InstanceID:             event.InstanceID.String(),
		MessageID:              coalesce(event.Metadata["message_id"], msg.MessageID, event.EventID.String()),
		Phone:                  phone,
		FromMe:                 false,
		ConnectedPhone:         t.connectedPhone,
		Momment:                eventTimestampMillis(event),
		Status:                 "RECEIVED",
		IsNewsletter:           true,
		Notification:           fmt.Sprintf("NEWSLETTER_MESSAGE_%s", strings.ToUpper(strings.TrimSpace(msg.Type))),
		NotificationParameters: []string{msg.MessageID, strings.ToUpper(strings.TrimSpace(msg.Type)), strconv.FormatInt(msg.Timestamp.Unix(), 10)},
	}

	if provider := eventctx.ContactProvider(ctx); provider != nil {
		if callback.ChatName == "" {
			if name := provider.ContactName(ctx, live.JID); name != "" {
				callback.ChatName = name
			}
		}
		if callback.Photo == "" {
			if photo := provider.ContactPhoto(ctx, live.JID); photo != "" {
				callback.Photo = photo
			}
		}
	}

	return callback, nil
}

func mapNewsletterNotification(event *types.InternalEvent) string {
	role := strings.ToUpper(strings.TrimSpace(event.Metadata["newsletter_role"]))
	switch event.EventType {
	case "newsletter_join":
		if role == "ADMIN" || role == "OWNER" {
			return "NEWSLETTER_ADMIN_PROMOTE"
		}
		return "NEWSLETTER_MEMBER_JOIN"
	case "newsletter_leave":
		if role == "ADMIN" || role == "OWNER" {
			return "NEWSLETTER_ADMIN_DEMOTE"
		}
		return "NEWSLETTER_MEMBER_LEAVE"
	case "newsletter_mute_change":
		mute := strings.ToUpper(strings.TrimSpace(event.Metadata["newsletter_mute"]))
		if mute == "" {
			mute = "UNKNOWN"
		}
		return "NEWSLETTER_MUTE_" + mute
	default:
		return event.EventType
	}
}

func mustParseJIDOrFallback(value string) watypes.JID {
	if jid, err := parseJID(value); err == nil {
		return jid
	}
	return watypes.JID{User: value}
}

func coalesce(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}
