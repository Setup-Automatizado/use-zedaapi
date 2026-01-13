package queue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	wameow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/events/echo"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// EventProcessor handles sending event/calendar messages through WhatsApp
type EventProcessor struct {
	log         *slog.Logger
	echoEmitter *echo.Emitter
}

// NewEventProcessor creates a new event processor instance
func NewEventProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *EventProcessor {
	return &EventProcessor{
		log:         log.With(slog.String("processor", "event")),
		echoEmitter: echoEmitter,
	}
}

// Process sends an event message via WhatsApp
func (p *EventProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	// Validate event content
	if args.EventContent == nil {
		return fmt.Errorf("event_content is required for event messages")
	}

	if args.EventContent.Name == "" {
		return fmt.Errorf("event name is required")
	}

	if args.EventContent.Description == "" {
		return fmt.Errorf("event description is required")
	}

	if args.EventContent.StartTime.IsZero() {
		return fmt.Errorf("event start_time is required")
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

	// Build event message
	msg, err := p.buildEventMessage(args.EventContent)
	if err != nil {
		return fmt.Errorf("build event message: %w", err)
	}

	// Handle reply to message
	if args.ReplyToMessageID != "" {
		contextInfo := &waProto.ContextInfo{
			StanzaID:      &args.ReplyToMessageID,
			Participant:   &args.Phone,
			QuotedMessage: &waProto.Message{},
		}
		msg.EventMessage.ContextInfo = contextInfo
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send event message: %w", err)
	}

	p.log.Info("event message sent successfully",
		slog.String("message_id", resp.ID),
		slog.String("phone", args.Phone),
		slog.String("event_name", args.EventContent.Name),
		slog.Time("start_time", args.EventContent.StartTime))

	// Store WhatsApp message ID for tracking
	args.WhatsAppMessageID = resp.ID

	// Emit API echo event for webhook notification
	if p.echoEmitter != nil {
		echoReq := &echo.EchoRequest{
			InstanceID:        args.InstanceID,
			WhatsAppMessageID: resp.ID,
			RecipientJID:      recipientJID,
			Message:           msg,
			Timestamp:         resp.Timestamp,
			MessageType:       "event",
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

// buildEventMessage constructs a WhatsApp event message
func (p *EventProcessor) buildEventMessage(content *EventMessage) (*waProto.Message, error) {
	// Convert start time to Unix timestamp
	startTimestamp := content.StartTime.Unix()

	// Build event message
	name := content.Name
	description := content.Description

	eventMsg := &waProto.EventMessage{
		Name:      proto.String(name),
		StartTime: proto.Int64(startTimestamp),
	}

	if description != "" {
		eventMsg.Description = proto.String(description)
	}

	// Add end time if specified
	if content.EndTime != nil && !content.EndTime.IsZero() {
		endTimestamp := content.EndTime.Unix()
		eventMsg.EndTime = proto.Int64(endTimestamp)
	}

	// Add location if specified
	if content.Location != nil {
		location := &waProto.LocationMessage{}

		if content.Location.Name != "" {
			location.Name = proto.String(content.Location.Name)
		}

		if content.Location.DegreesLatitude != nil {
			location.DegreesLatitude = content.Location.DegreesLatitude
		}

		if content.Location.DegreesLongitude != nil {
			location.DegreesLongitude = content.Location.DegreesLongitude
		}

		if location.Name != nil || location.DegreesLatitude != nil || location.DegreesLongitude != nil {
			eventMsg.Location = location
		}
	}

	if content.Canceled {
		eventMsg.IsCanceled = proto.Bool(true)
	}

	msg := &waProto.Message{
		EventMessage: eventMsg,
	}

	return msg, nil
}

// simulateTyping sends typing indicators to make the interaction feel more natural
func (p *EventProcessor) simulateTyping(client *wameow.Client, jid types.JID, delayMs int64) error {
	// Send "composing" presence
	if err := client.SendChatPresence(context.Background(), jid, types.ChatPresenceComposing, types.ChatPresenceMediaText); err != nil {
		return err
	}

	// Wait for the specified delay
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// Send "paused" presence
	if err := client.SendChatPresence(context.Background(), jid, types.ChatPresencePaused, types.ChatPresenceMediaText); err != nil {
		return err
	}

	return nil
}
