package queue

import (
	"context"
	"fmt"
	"log/slog"

	wameow "go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"

	"go.mau.fi/whatsmeow/api/internal/events/echo"
)

// TextProcessor handles text message sending via WhatsApp
type TextProcessor struct {
	presenceHelper *PresenceHelper
	linkPreviewGen *LinkPreviewGenerator
	echoEmitter    *echo.Emitter
	log            *slog.Logger
}

// NewTextProcessor creates a new text message processor
func NewTextProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *TextProcessor {
	return &TextProcessor{
		presenceHelper: NewPresenceHelper(),
		echoEmitter:    echoEmitter,
		log:            log.With(slog.String("processor", "text")),
	}
}

// Process sends a text message via WhatsApp with full feature support
func (p *TextProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.TextContent == nil {
		return fmt.Errorf("text_content is required for text messages")
	}

	// Parse recipient JID
	recipientJID, err := types.ParseJID(args.Phone)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Simulate typing indicator
	if args.DelayTyping > 0 {
		if err := p.presenceHelper.SimulateTyping(client, recipientJID, args.DelayTyping); err != nil {
			p.log.Warn("failed to send typing indicator",
				slog.String("error", err.Error()),
				slog.String("phone", args.Phone))
		}
	}

	// Build ContextInfo using helper
	contextBuilder := NewContextInfoBuilder(client, recipientJID, args, p.log)
	contextInfo, err := contextBuilder.Build(ctx)
	if err != nil {
		p.log.Warn("failed to build context info",
			slog.String("error", err.Error()))
		contextInfo = nil
	}

	// Generate link preview if URL detected and not replying
	var extendedText *waProto.ExtendedTextMessage
	if args.ReplyToMessageID == "" {
		// Initialize link preview generator lazily
		if p.linkPreviewGen == nil {
			p.linkPreviewGen = NewLinkPreviewGenerator(client, p.log)
		}

		// Pass LinkPreviewOverride for custom metadata (e.g., from /send-link endpoint)
		extendedText, err = p.linkPreviewGen.Generate(ctx, args.TextContent.Message, args.LinkPreview, args.LinkPreviewOverride)
		if err != nil {
			p.log.Warn("failed to generate link preview",
				slog.String("error", err.Error()))
		}
	}

	// Build message
	msg := p.buildMessage(args, contextInfo, extendedText)

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send text message: %w", err)
	}

	p.log.Info("text message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.Bool("has_mentions", contextInfo != nil && len(contextInfo.MentionedJID) > 0),
		slog.Bool("has_link_preview", extendedText != nil))

	args.WhatsAppMessageID = resp.ID

	// Emit API echo event for webhook notification
	if p.echoEmitter != nil {
		echoReq := &echo.EchoRequest{
			InstanceID:        args.InstanceID,
			WhatsAppMessageID: resp.ID,
			RecipientJID:      recipientJID,
			Message:           msg,
			Timestamp:         resp.Timestamp,
			MessageType:       "text",
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

// buildMessage constructs the appropriate WhatsApp message proto
func (p *TextProcessor) buildMessage(args *SendMessageArgs, contextInfo *waProto.ContextInfo, extendedText *waProto.ExtendedTextMessage) *waProto.Message {
	// If we have extended text from link preview, use it
	if extendedText != nil {
		// Add context info if available
		if contextInfo != nil {
			if extendedText.ContextInfo == nil {
				extendedText.ContextInfo = contextInfo
			}
		}
		return &waProto.Message{
			ExtendedTextMessage: extendedText,
		}
	}

	// If we have context info, use ExtendedTextMessage
	if contextInfo != nil {
		text := args.TextContent.Message
		return &waProto.Message{
			ExtendedTextMessage: &waProto.ExtendedTextMessage{
				Text:        &text,
				ContextInfo: contextInfo,
			},
		}
	}

	// Simple conversation message
	conversation := args.TextContent.Message
	return &waProto.Message{
		Conversation: &conversation,
	}
}
