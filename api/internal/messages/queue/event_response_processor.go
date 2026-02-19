package queue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/protobuf/proto"

	wameow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/events/echo"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	waCommon "go.mau.fi/whatsmeow/proto/waCommon"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// EventResponseProcessor handles sending event response messages through WhatsApp
type EventResponseProcessor struct {
	log         *slog.Logger
	echoEmitter *echo.Emitter
}

// NewEventResponseProcessor creates a new event response processor instance
func NewEventResponseProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *EventResponseProcessor {
	return &EventResponseProcessor{
		log:         log.With(slog.String("processor", "event_response")),
		echoEmitter: echoEmitter,
	}
}

// Process sends an event response via WhatsApp
func (p *EventResponseProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.EventResponseContent == nil {
		return fmt.Errorf("event_response_content is required for event_response messages")
	}

	content := args.EventResponseContent

	if content.EventID == "" {
		return fmt.Errorf("event_id is required")
	}

	// Parse and validate response type
	var responseType waE2E.EventResponseMessage_EventResponseType
	switch content.Response {
	case "going":
		responseType = waE2E.EventResponseMessage_GOING
	case "not_going":
		responseType = waE2E.EventResponseMessage_NOT_GOING
	case "maybe":
		responseType = waE2E.EventResponseMessage_MAYBE
	default:
		return fmt.Errorf("response must be 'going', 'not_going', or 'maybe', got: %s", content.Response)
	}

	// Parse recipient JID
	recipientJID, err := types.ParseJID(args.Phone)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Simulate typing indicator if DelayTyping is set
	if args.DelayTyping > 0 {
		if err := p.simulateTyping(client, recipientJID, args.DelayTyping); err != nil {
			p.log.Warn("failed to send typing indicator",
				slog.String("error", err.Error()),
				slog.String("phone", args.Phone))
		}
	}

	// Build the event response message
	eventResponseMsg := &waE2E.EventResponseMessage{
		Response:    &responseType,
		TimestampMS: proto.Int64(time.Now().UnixMilli()),
	}

	if content.ExtraGuestCount != nil && *content.ExtraGuestCount > 0 {
		eventResponseMsg.ExtraGuestCount = proto.Int32(int32(*content.ExtraGuestCount))
	}

	// Marshal the event response for encryption
	plaintext, err := proto.Marshal(eventResponseMsg)
	if err != nil {
		return fmt.Errorf("marshal event response: %w", err)
	}

	// Build encrypted event response message
	eventMsgKey := &waCommon.MessageKey{
		RemoteJID: proto.String(recipientJID.String()),
		ID:        proto.String(content.EventID),
		FromMe:    proto.Bool(false),
	}

	encEventResponse := &waE2E.EncEventResponseMessage{
		EventCreationMessageKey: eventMsgKey,
		EncPayload:              plaintext,
		EncIV:                   nil,
	}

	msg := &waProto.Message{
		EncEventResponseMessage: encEventResponse,
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg, BuildSendExtra(args))
	if err != nil {
		return fmt.Errorf("send event response: %w", err)
	}

	p.log.Info("event response sent successfully",
		slog.String("message_id", resp.ID),
		slog.String("phone", args.Phone),
		slog.String("event_id", content.EventID),
		slog.String("response", content.Response))

	args.WhatsAppMessageID = resp.ID

	// Emit API echo event
	if p.echoEmitter != nil {
		echoReq := &echo.EchoRequest{
			InstanceID:        args.InstanceID,
			WhatsAppMessageID: resp.ID,
			RecipientJID:      recipientJID,
			Message:           msg,
			Timestamp:         resp.Timestamp,
			MessageType:       "event_response",
			ZaapID:            args.ZaapID,
			HasMedia:          false,
		}
		if err := p.echoEmitter.EmitEcho(ctx, echoReq); err != nil {
			p.log.Warn("failed to emit API echo",
				slog.String("error", err.Error()),
				slog.String("zaap_id", args.ZaapID))
		}
	}

	return nil
}

func (p *EventResponseProcessor) simulateTyping(client *wameow.Client, jid types.JID, delayMs int64) error {
	if err := client.SendChatPresence(context.Background(), jid, types.ChatPresenceComposing, types.ChatPresenceMediaText); err != nil {
		return err
	}
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	if err := client.SendChatPresence(context.Background(), jid, types.ChatPresencePaused, types.ChatPresenceMediaText); err != nil {
		return err
	}
	return nil
}
