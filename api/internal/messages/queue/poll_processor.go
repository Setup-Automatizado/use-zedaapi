package queue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	wameow "go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

// PollProcessor handles sending poll messages through WhatsApp
type PollProcessor struct {
	log *slog.Logger
}

// NewPollProcessor creates a new poll processor instance
func NewPollProcessor(log *slog.Logger) *PollProcessor {
	return &PollProcessor{
		log: log.With(slog.String("processor", "poll")),
	}
}

// Process sends a poll message via WhatsApp
func (p *PollProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	// Validate poll content
	if args.PollContent == nil {
		return fmt.Errorf("poll_content is required for poll messages")
	}

	if args.PollContent.Question == "" {
		return fmt.Errorf("poll question is required")
	}

	if len(args.PollContent.Options) < 2 {
		return fmt.Errorf("poll must have at least 2 options")
	}

	if len(args.PollContent.Options) > 12 {
		return fmt.Errorf("poll cannot have more than 12 options")
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

	// Build poll message
	msg, err := p.buildPollMessage(args.PollContent)
	if err != nil {
		return fmt.Errorf("build poll message: %w", err)
	}

	// Handle reply to message
	if args.ReplyToMessageID != "" {
		contextInfo := &waProto.ContextInfo{
			StanzaID:      &args.ReplyToMessageID,
			Participant:   &args.Phone,
			QuotedMessage: &waProto.Message{},
		}
		msg.PollCreationMessage.ContextInfo = contextInfo
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send poll message: %w", err)
	}

	p.log.Info("poll message sent successfully",
		slog.String("message_id", resp.ID),
		slog.String("phone", args.Phone),
		slog.String("question", args.PollContent.Question),
		slog.Int("options_count", len(args.PollContent.Options)))

	// Store WhatsApp message ID for tracking
	args.WhatsAppMessageID = resp.ID

	return nil
}

// buildPollMessage constructs a WhatsApp poll creation message
func (p *PollProcessor) buildPollMessage(content *PollMessage) (*waProto.Message, error) {
	// Build poll options
	options := make([]*waProto.PollCreationMessage_Option, 0, len(content.Options))
	for _, opt := range content.Options {
		optionName := opt
		options = append(options, &waProto.PollCreationMessage_Option{
			OptionName: &optionName,
		})
	}

	// Set max selections (default to 0 for single selection)
	maxSelections := uint32(0)
	if content.MaxSelections > 0 {
		maxSelections = uint32(content.MaxSelections)
	}

	// Build poll creation message
	question := content.Question
	msg := &waProto.Message{
		PollCreationMessage: &waProto.PollCreationMessage{
			Name:                   &question,
			Options:                options,
			SelectableOptionsCount: &maxSelections,
		},
	}

	return msg, nil
}

// simulateTyping sends typing indicators to make the interaction feel more natural
func (p *PollProcessor) simulateTyping(client *wameow.Client, jid types.JID, delayMs int64) error {
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
