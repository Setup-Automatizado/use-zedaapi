package whatsmeow

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	eventctx "go.mau.fi/whatsmeow/api/internal/events/eventctx"
	eventtypes "go.mau.fi/whatsmeow/api/internal/events/types"
	waBinary "go.mau.fi/whatsmeow/binary"
	watypes "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (t *Transformer) transformCallOffer(ctx context.Context, logger *slog.Logger, offer *events.CallOffer) (*eventtypes.InternalEvent, error) {
	event := newInternalEvent(t.instanceID, "call_offer", offer)
	populateCallMetadata(ctx, event, offer.BasicCallMeta, offer.CallRemoteMeta, offer.Data, "offer")
	logger.InfoContext(ctx, "transformed call offer",
		slog.String("event_id", event.EventID.String()),
		slog.String("call_id", offer.CallID),
		slog.String("from", offer.From.String()),
	)
	return event, nil
}

func (t *Transformer) transformCallOfferNotice(ctx context.Context, logger *slog.Logger, notice *events.CallOfferNotice) (*eventtypes.InternalEvent, error) {
	event := newInternalEvent(t.instanceID, "call_offer_notice", notice)
	populateCallMetadata(ctx, event, notice.BasicCallMeta, watypes.CallRemoteMeta{}, notice.Data, "offer_notice")
	if notice.Media != "" {
		event.Metadata["call_media"] = strings.ToLower(notice.Media)
	}
	logger.InfoContext(ctx, "transformed call offer notice",
		slog.String("event_id", event.EventID.String()),
		slog.String("call_id", notice.CallID),
		slog.String("media", notice.Media),
	)
	return event, nil
}

func (t *Transformer) transformCallRelayLatency(ctx context.Context, logger *slog.Logger, relay *events.CallRelayLatency) (*eventtypes.InternalEvent, error) {
	event := newInternalEvent(t.instanceID, "call_relay_latency", relay)
	populateCallMetadata(ctx, event, relay.BasicCallMeta, watypes.CallRemoteMeta{}, relay.Data, "relaylatency")
	logger.DebugContext(ctx, "transformed call relay latency",
		slog.String("event_id", event.EventID.String()),
		slog.String("call_id", relay.CallID),
	)
	return event, nil
}

func (t *Transformer) transformCallTransport(ctx context.Context, logger *slog.Logger, transport *events.CallTransport) (*eventtypes.InternalEvent, error) {
	event := newInternalEvent(t.instanceID, "call_transport", transport)
	populateCallMetadata(ctx, event, transport.BasicCallMeta, transport.CallRemoteMeta, transport.Data, "transport")
	logger.DebugContext(ctx, "transformed call transport",
		slog.String("event_id", event.EventID.String()),
		slog.String("call_id", transport.CallID),
	)
	return event, nil
}

func (t *Transformer) transformCallTerminate(ctx context.Context, logger *slog.Logger, terminate *events.CallTerminate) (*eventtypes.InternalEvent, error) {
	event := newInternalEvent(t.instanceID, "call_terminate", terminate)
	populateCallMetadata(ctx, event, terminate.BasicCallMeta, watypes.CallRemoteMeta{}, terminate.Data, "terminate")
	if strings.TrimSpace(terminate.Reason) != "" {
		event.Metadata["call_reason"] = terminate.Reason
	}
	logger.InfoContext(ctx, "transformed call terminate",
		slog.String("event_id", event.EventID.String()),
		slog.String("call_id", terminate.CallID),
		slog.String("reason", terminate.Reason),
	)
	return event, nil
}

func (t *Transformer) transformCallReject(ctx context.Context, logger *slog.Logger, reject *events.CallReject) (*eventtypes.InternalEvent, error) {
	event := newInternalEvent(t.instanceID, "call_reject", reject)
	populateCallMetadata(ctx, event, reject.BasicCallMeta, watypes.CallRemoteMeta{}, reject.Data, "reject")
	logger.InfoContext(ctx, "transformed call reject",
		slog.String("event_id", event.EventID.String()),
		slog.String("call_id", reject.CallID),
	)
	return event, nil
}

func newInternalEvent(instanceID uuid.UUID, eventType string, payload interface{}) *eventtypes.InternalEvent {
	return &eventtypes.InternalEvent{
		InstanceID: instanceID,
		EventID:    uuid.New(),
		EventType:  eventType,
		SourceLib:  eventtypes.SourceLibWhatsmeow,
		RawPayload: payload,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}
}

func populateCallMetadata(ctx context.Context, event *eventtypes.InternalEvent, meta watypes.BasicCallMeta, remote watypes.CallRemoteMeta, data *waBinary.Node, kind string) {
	if event.Metadata == nil {
		event.Metadata = make(map[string]string)
	}
	event.Metadata["call_event_kind"] = kind
	if meta.CallID != "" {
		event.Metadata["call_id"] = strings.ToUpper(meta.CallID)
	}
	if !meta.From.IsEmpty() {
		event.Metadata["call_from"] = meta.From.String()
	}
	if !meta.CallCreator.IsEmpty() {
		event.Metadata["call_creator"] = meta.CallCreator.String()
	}
	if !meta.CallCreatorAlt.IsEmpty() {
		event.Metadata["call_creator_alt"] = meta.CallCreatorAlt.String()
	}
	if !meta.GroupJID.IsEmpty() {
		event.Metadata["group_id"] = meta.GroupJID.String()
	}
	if event.Metadata["timestamp"] == "" {
		event.Metadata["timestamp"] = fmt.Sprintf("%d", meta.Timestamp.Unix())
	}
	if remote.RemotePlatform != "" {
		event.Metadata["remote_platform"] = remote.RemotePlatform
	}
	if remote.RemoteVersion != "" {
		event.Metadata["remote_version"] = remote.RemoteVersion
	}
	if media := deriveCallMedia(data); media != "" {
		event.Metadata["call_media"] = media
	}

	resolver := eventctx.LIDResolverFromContext(ctx)
	if _, ok := event.Metadata["chat_pn"]; !ok {
		if pn := resolvedPNString(ctx, resolver, meta.From); pn != "" {
			event.Metadata["chat_pn"] = pn
		}
	}
	if _, ok := event.Metadata["call_from_pn"]; !ok {
		if pn := resolvedPNString(ctx, resolver, meta.From); pn != "" {
			event.Metadata["call_from_pn"] = pn
		}
	}
	if _, ok := event.Metadata["sender_pn"]; !ok {
		if pn := resolvedPNString(ctx, resolver, meta.CallCreator); pn != "" {
			event.Metadata["sender_pn"] = pn
		} else if pn := resolvedPNString(ctx, resolver, meta.CallCreatorAlt); pn != "" {
			event.Metadata["sender_pn"] = pn
		}
	}
	if _, ok := event.Metadata["call_creator_pn"]; !ok {
		if pn := resolvedPNString(ctx, resolver, meta.CallCreator); pn != "" {
			event.Metadata["call_creator_pn"] = pn
		} else if pn := resolvedPNString(ctx, resolver, meta.CallCreatorAlt); pn != "" {
			event.Metadata["call_creator_pn"] = pn
		}
	}
}

func deriveCallMedia(node *waBinary.Node) string {
	if node == nil {
		return ""
	}
	ag := node.AttrGetter()
	if media := strings.TrimSpace(ag.OptionalString("media")); media != "" {
		return strings.ToLower(media)
	}
	if typ := strings.TrimSpace(ag.OptionalString("type")); typ != "" {
		return strings.ToLower(typ)
	}
	// Inspect nested transport payloads for media hints
	for _, child := range node.GetChildren() {
		cag := child.AttrGetter()
		if media := strings.TrimSpace(cag.OptionalString("media")); media != "" {
			return strings.ToLower(media)
		}
		if typ := strings.TrimSpace(cag.OptionalString("type")); typ != "" {
			return strings.ToLower(typ)
		}
	}
	return ""
}
