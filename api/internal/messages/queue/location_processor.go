package queue

import (
	"context"
	"fmt"
	"log/slog"

	wameow "go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

// LocationProcessor handles location message sending via WhatsApp
type LocationProcessor struct {
	log            *slog.Logger
	presenceHelper *PresenceHelper
}

// NewLocationProcessor creates a new location message processor
func NewLocationProcessor(log *slog.Logger) *LocationProcessor {
	return &LocationProcessor{
		log:            log.With(slog.String("processor", "location")),
		presenceHelper: NewPresenceHelper(),
	}
}

// Process sends a location message via WhatsApp
func (p *LocationProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.LocationContent == nil {
		return fmt.Errorf("location_content is required for location messages")
	}

	// Parse recipient JID
	recipientJID, err := types.ParseJID(args.Phone)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Simulate typing indicator if DelayTyping is set
	if args.DelayTyping > 0 {
		if err := p.presenceHelper.SimulateTyping(client, recipientJID, args.DelayTyping); err != nil {
			p.log.Warn("failed to send typing indicator",
				slog.String("error", err.Error()),
				slog.String("phone", args.Phone))
		}
	}

	// Build ContextInfo using helper
	// This provides support for: mentions, reply-to, ephemeral messages, private answer
	contextBuilder := NewContextInfoBuilder(client, recipientJID, args, p.log)
	contextInfo, err := contextBuilder.Build(ctx)
	if err != nil {
		p.log.Warn("failed to build context info, sending without it",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Build location message
	degreesLatitude := args.LocationContent.Latitude
	degreesLongitude := args.LocationContent.Longitude

	msg := &waProto.Message{
		LocationMessage: &waProto.LocationMessage{
			DegreesLatitude:  &degreesLatitude,
			DegreesLongitude: &degreesLongitude,
			ContextInfo:      contextInfo, // Can be nil if no context features
		},
	}

	// Add name if provided
	if args.LocationContent.Name != nil && *args.LocationContent.Name != "" {
		msg.LocationMessage.Name = args.LocationContent.Name
	}

	// Add address if provided
	if args.LocationContent.Address != nil && *args.LocationContent.Address != "" {
		msg.LocationMessage.Address = args.LocationContent.Address
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send location message: %w", err)
	}

	p.log.Info("location message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.Float64("latitude", degreesLatitude),
		slog.Float64("longitude", degreesLongitude),
		slog.Bool("has_name", args.LocationContent.Name != nil),
		slog.Bool("has_address", args.LocationContent.Address != nil),
		slog.Bool("has_context", contextInfo != nil),
		slog.Time("timestamp", resp.Timestamp))

	args.WhatsAppMessageID = resp.ID

	return nil
}
