package zapi

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	eventctx "go.mau.fi/whatsmeow/api/internal/events/eventctx"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	watypes "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (t *Transformer) transformCallEvent(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*ReceivedCallback, error) {
	basicMeta, notification, missed, err := t.extractCallMetadata(event)
	if err != nil {
		return nil, err
	}

	phone := conversationIdentifierFromJID(basicMeta.From)
	if pn := strings.TrimSpace(event.Metadata["chat_pn"]); pn != "" {
		if parsed, err := watypes.ParseJID(pn); err == nil {
			phone = conversationIdentifierFromJID(parsed)
		} else if fallback := sanitizeConversationFallback(pn); fallback != "" {
			phone = fallback
		}
	}
	if phone == "" {
		phone = sanitizeConversationFallback(basicMeta.From.String())
	}

	callback := &ReceivedCallback{
		Type:           "ReceivedCallback",
		InstanceID:     event.InstanceID.String(),
		MessageID:      event.Metadata["message_id"],
		Phone:          phone,
		FromMe:         isConnectedJID(basicMeta.CallCreator, t.connectedPhone),
		ConnectedPhone: t.connectedPhone,
		Momment:        eventTimestampMillis(event),
		Status:         "RECEIVED",
		IsGroup:        !basicMeta.GroupJID.IsEmpty(),
		Notification:   notification,
		CallID:         strings.ToUpper(basicMeta.CallID),
		RequestMethod:  strings.TrimSpace(event.Metadata["call_event_kind"]),
	}

	if callback.MessageID == "" {
		callback.MessageID = event.EventID.String()
	}

	if reason := strings.TrimSpace(event.Metadata["call_reason"]); reason != "" {
		callback.NotificationParameters = []string{reason}
	}
	if missed {
		// Reserve first parameter for call stage when available
		if kind := strings.TrimSpace(event.Metadata["call_event_kind"]); kind != "" {
			callback.NotificationParameters = append([]string{strings.ToUpper(kind)}, callback.NotificationParameters...)
		}
	}

	if !basicMeta.GroupJID.IsEmpty() {
		callback.ChatName = sanitizeConversationFallback(basicMeta.GroupJID.String())
	}

	provider := eventctx.ContactProvider(ctx)
	if provider != nil {
		target := basicMeta.From
		if callback.IsGroup && !basicMeta.GroupJID.IsEmpty() {
			target = basicMeta.GroupJID
		}
		if callback.ChatName == "" {
			if name := provider.ContactName(ctx, target); name != "" {
				callback.ChatName = name
			}
		}
		if callback.Photo == "" {
			if photo := provider.ContactPhoto(ctx, target); photo != "" {
				callback.Photo = photo
			}
		}
		if callback.SenderName == "" {
			if name := provider.ContactName(ctx, basicMeta.CallCreator); name != "" {
				callback.SenderName = name
			}
		}
		if callback.SenderPhoto == "" {
			if photo := provider.ContactPhoto(ctx, basicMeta.CallCreator); photo != "" {
				callback.SenderPhoto = photo
			}
		}
	}

	if phone := participantPhoneFromJID(basicMeta.CallCreator); phone != "" && !callback.IsGroup {
		// For direct calls expose participant phone
		callback.ParticipantPhone = phone
	}
	if lid := participantLIDFromJID(basicMeta.CallCreator); lid != "" && !callback.IsGroup {
		callback.ParticipantLid = lid
	}
	if callback.ParticipantPhone == "" && !callback.IsGroup {
		if phone := extractPhoneNumber(event.Metadata["call_creator_pn"]); phone != "" {
			callback.ParticipantPhone = phone
		}
	}

	return callback, nil
}

func (t *Transformer) extractCallMetadata(event *types.InternalEvent) (watypes.BasicCallMeta, string, bool, error) {
	var (
		basic  watypes.BasicCallMeta
		missed bool
	)

	switch payload := event.RawPayload.(type) {
	case *events.CallOffer:
		basic = payload.BasicCallMeta
	case *events.CallOfferNotice:
		basic = payload.BasicCallMeta
	case *events.CallRelayLatency:
		basic = payload.BasicCallMeta
	case *events.CallTransport:
		basic = payload.BasicCallMeta
	case *events.CallTerminate:
		basic = payload.BasicCallMeta
		missed = true
	case *events.CallReject:
		basic = payload.BasicCallMeta
		missed = true
	default:
		return watypes.BasicCallMeta{}, "", false, fmt.Errorf("unsupported call payload type %T", event.RawPayload)
	}

	media := strings.ToLower(strings.TrimSpace(event.Metadata["call_media"]))
	if media == "" {
		media = "audio"
	}
	notification := callNotificationName(media, missed)

	// Propagate message_id from metadata if present
	if event.Metadata != nil {
		if id := strings.TrimSpace(event.Metadata["call_id"]); id != "" {
			basic.CallID = id
		}
	}

	return basic, notification, missed, nil
}

func callNotificationName(media string, missed bool) string {
	base := "CALL_"
	if missed {
		base += "MISSED_"
	}
	if strings.Contains(media, "video") {
		return base + "VIDEO"
	}
	return base + "VOICE"
}

func isConnectedJID(jid watypes.JID, connected string) bool {
	connected = sanitizeUserComponent(connected)
	if connected == "" {
		return false
	}
	return strings.EqualFold(connected, userPhoneFromJID(jid))
}
