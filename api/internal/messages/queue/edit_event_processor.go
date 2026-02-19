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
	"go.mau.fi/whatsmeow/types"
)

// EditEventProcessor handles sending edit event messages through WhatsApp
type EditEventProcessor struct {
	log         *slog.Logger
	echoEmitter *echo.Emitter
}

// NewEditEventProcessor creates a new edit event processor instance
func NewEditEventProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *EditEventProcessor {
	return &EditEventProcessor{
		log:         log.With(slog.String("processor", "edit_event")),
		echoEmitter: echoEmitter,
	}
}

// Process sends an edit event message via WhatsApp
func (p *EditEventProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.EditEventContent == nil {
		return fmt.Errorf("edit_event_content is required for edit_event messages")
	}

	content := args.EditEventContent

	if content.EventID == "" {
		return fmt.Errorf("event_id is required")
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

	// Build the updated event message
	eventMsg := &waProto.EventMessage{
		ContextInfo: &waProto.ContextInfo{
			StanzaID:    proto.String(content.EventID),
			Participant: proto.String(args.Phone),
		},
	}

	if content.Name != "" {
		eventMsg.Name = proto.String(content.Name)
	}
	if content.Description != "" {
		eventMsg.Description = proto.String(content.Description)
	}
	if content.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, content.StartTime)
		if err == nil {
			eventMsg.StartTime = proto.Int64(startTime.Unix())
		}
	}
	if content.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, content.EndTime)
		if err == nil {
			eventMsg.EndTime = proto.Int64(endTime.Unix())
		}
	}
	if content.Location != "" {
		eventMsg.Location = &waProto.LocationMessage{
			Name: proto.String(content.Location),
		}
	}
	if content.Canceled {
		eventMsg.IsCanceled = proto.Bool(true)
	}

	msg := &waProto.Message{
		EventMessage: eventMsg,
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg, BuildSendExtra(args))
	if err != nil {
		return fmt.Errorf("send edit event: %w", err)
	}

	p.log.Info("edit event sent successfully",
		slog.String("message_id", resp.ID),
		slog.String("phone", args.Phone),
		slog.String("event_id", content.EventID))

	args.WhatsAppMessageID = resp.ID

	// Emit API echo event
	if p.echoEmitter != nil {
		echoReq := &echo.EchoRequest{
			InstanceID:        args.InstanceID,
			WhatsAppMessageID: resp.ID,
			RecipientJID:      recipientJID,
			Message:           msg,
			Timestamp:         resp.Timestamp,
			MessageType:       "edit_event",
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

func (p *EditEventProcessor) simulateTyping(client *wameow.Client, jid types.JID, delayMs int64) error {
	if err := client.SendChatPresence(context.Background(), jid, types.ChatPresenceComposing, types.ChatPresenceMediaText); err != nil {
		return err
	}
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	if err := client.SendChatPresence(context.Background(), jid, types.ChatPresencePaused, types.ChatPresenceMediaText); err != nil {
		return err
	}
	return nil
}
