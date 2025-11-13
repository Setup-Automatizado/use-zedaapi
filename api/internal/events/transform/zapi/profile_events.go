package zapi

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	eventctx "go.mau.fi/whatsmeow/api/internal/events/eventctx"
	eventtypes "go.mau.fi/whatsmeow/api/internal/events/types"
	watypes "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (t *Transformer) transformProfileEvent(ctx context.Context, logger *slog.Logger, event *eventtypes.InternalEvent) (*ReceivedCallback, error) {
	switch payload := event.RawPayload.(type) {
	case *events.PushName:
		return t.profileNameCallback(ctx, event, payload.JID, payload.NewPushName)
	case *events.BusinessName:
		return t.profileNameCallback(ctx, event, payload.JID, payload.NewBusinessName)
	case *events.UserAbout:
		return t.profileAboutCallback(ctx, event, payload)
	default:
		return nil, fmt.Errorf("unsupported profile payload %T", event.RawPayload)
	}
}

func (t *Transformer) profileNameCallback(ctx context.Context, event *eventtypes.InternalEvent, jid watypes.JID, name string) (*ReceivedCallback, error) {
	phone := conversationIdentifierFromJID(jid)
	if phone == "" {
		phone = sanitizeConversationFallback(jid.String())
	}
	fromMe := strings.EqualFold(userPhoneFromJID(jid), sanitizeUserComponent(t.connectedPhone))

	callback := &ReceivedCallback{
		Type:           "ReceivedCallback",
		InstanceID:     event.InstanceID.String(),
		MessageID:      event.EventID.String(),
		Phone:          phone,
		FromMe:         fromMe,
		ConnectedPhone: t.connectedPhone,
		Momment:        eventTimestampMillis(event),
		Status:         "RECEIVED",
		Notification:   "PROFILE_NAME_UPDATED",
		ProfileName:    name,
	}

	provider := eventctx.ContactProvider(ctx)
	if provider != nil {
		if callback.ChatName == "" {
			if n := provider.ContactName(ctx, jid); n != "" {
				callback.ChatName = n
			}
		}
		if callback.Photo == "" {
			if photo := provider.ContactPhoto(ctx, jid); photo != "" {
				callback.Photo = photo
			}
		}
	}

	return callback, nil
}

func (t *Transformer) profileAboutCallback(ctx context.Context, event *eventtypes.InternalEvent, about *events.UserAbout) (*ReceivedCallback, error) {
	phone := conversationIdentifierFromJID(about.JID)
	if phone == "" {
		phone = sanitizeConversationFallback(about.JID.String())
	}
	fromMe := strings.EqualFold(userPhoneFromJID(about.JID), sanitizeUserComponent(t.connectedPhone))

	callback := &ReceivedCallback{
		Type:           "ReceivedCallback",
		InstanceID:     event.InstanceID.String(),
		MessageID:      event.EventID.String(),
		Phone:          phone,
		FromMe:         fromMe,
		ConnectedPhone: t.connectedPhone,
		Momment:        eventTimestampMillis(event),
		Status:         "RECEIVED",
		Notification:   "PROFILE_STATUS_UPDATED",
		Text: &TextContent{
			Message: about.Status,
		},
	}

	provider := eventctx.ContactProvider(ctx)
	if provider != nil {
		if callback.ChatName == "" {
			if n := provider.ContactName(ctx, about.JID); n != "" {
				callback.ChatName = n
			}
		}
		if callback.Photo == "" {
			if photo := provider.ContactPhoto(ctx, about.JID); photo != "" {
				callback.Photo = photo
			}
		}
	}

	return callback, nil
}
