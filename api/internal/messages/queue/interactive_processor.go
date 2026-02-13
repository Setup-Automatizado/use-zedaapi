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
)

// InteractiveProcessor handles interactive message sending via WhatsApp (buttons, lists)
type InteractiveProcessor struct {
	log         *slog.Logger
	echoEmitter *echo.Emitter
}

// NewInteractiveProcessor creates a new interactive message processor
func NewInteractiveProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *InteractiveProcessor {
	return &InteractiveProcessor{
		log:         log.With(slog.String("processor", "interactive")),
		echoEmitter: echoEmitter,
	}
}

// Process sends an interactive message via WhatsApp
func (p *InteractiveProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.InteractiveContent == nil {
		return fmt.Errorf("interactive_content is required for interactive messages")
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

	// Build interactive message based on type
	var msg *waProto.Message
	switch args.InteractiveContent.Type {
	case InteractiveTypeButton:
		msg, err = p.buildButtonMessage(args.InteractiveContent)
	case InteractiveTypeList:
		msg, err = p.buildListMessage(args.InteractiveContent)
	default:
		return fmt.Errorf("unsupported interactive type: %s", args.InteractiveContent.Type)
	}

	if err != nil {
		return fmt.Errorf("build interactive message: %w", err)
	}

	// Handle reply to message
	if args.ReplyToMessageID != "" {
		// Add context info to the appropriate message type
		contextInfo := &waProto.ContextInfo{
			StanzaID:      &args.ReplyToMessageID,
			Participant:   &args.Phone,
			QuotedMessage: &waProto.Message{},
		}

		if msg.ButtonsMessage != nil {
			msg.ButtonsMessage.ContextInfo = contextInfo
		} else if msg.ListMessage != nil {
			msg.ListMessage.ContextInfo = contextInfo
		}
	}

	// Handle ephemeral (disappearing) messages - not implemented in current whatsmeow version
	// if args.Duration != nil && *args.Duration > 0 {
	// 	// Ephemeral messages require additional setup
	// }

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send interactive message: %w", err)
	}

	p.log.Info("interactive message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.String("interactive_type", string(args.InteractiveContent.Type)),
		slog.Time("timestamp", resp.Timestamp))

	args.WhatsAppMessageID = resp.ID

	// Emit API echo event for webhook notification
	if p.echoEmitter != nil {
		echoReq := &echo.EchoRequest{
			InstanceID:        args.InstanceID,
			WhatsAppMessageID: resp.ID,
			RecipientJID:      recipientJID,
			Message:           msg,
			Timestamp:         resp.Timestamp,
			MessageType:       "interactive",
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

// buildButtonMessage builds a button message
func (p *InteractiveProcessor) buildButtonMessage(content *InteractiveMessage) (*waProto.Message, error) {
	if len(content.Buttons) == 0 {
		return nil, fmt.Errorf("buttons are required for button type")
	}

	if len(content.Buttons) > 3 {
		return nil, fmt.Errorf("maximum 3 buttons allowed")
	}

	// Build buttons
	buttons := make([]*waProto.ButtonsMessage_Button, 0, len(content.Buttons))
	for _, btn := range content.Buttons {
		buttonID := btn.ID
		displayText := btn.Title

		buttons = append(buttons, &waProto.ButtonsMessage_Button{
			ButtonID: &buttonID,
			ButtonText: &waProto.ButtonsMessage_Button_ButtonText{
				DisplayText: &displayText,
			},
			Type: waProto.ButtonsMessage_Button_RESPONSE.Enum(),
		})
	}

	// Build message
	msg := &waProto.Message{
		ButtonsMessage: &waProto.ButtonsMessage{
			ContentText: &content.Body,
			Buttons:     buttons,
			HeaderType:  waProto.ButtonsMessage_EMPTY.Enum(),
		},
	}

	// Add header if provided (note: ButtonsMessage doesn't have Text field in current whatsmeow)
	// Header support may require different approach depending on whatsmeow version

	// Add footer if provided
	if content.Footer != nil && *content.Footer != "" {
		msg.ButtonsMessage.FooterText = content.Footer
	}

	return msg, nil
}

// buildListMessage builds a list message
func (p *InteractiveProcessor) buildListMessage(content *InteractiveMessage) (*waProto.Message, error) {
	if len(content.Sections) == 0 {
		return nil, fmt.Errorf("sections are required for list type")
	}

	// Build sections
	sections := make([]*waProto.ListMessage_Section, 0, len(content.Sections))
	for _, section := range content.Sections {
		// Build rows
		rows := make([]*waProto.ListMessage_Row, 0, len(section.Rows))
		for _, row := range section.Rows {
			rowID := row.ID
			rowTitle := row.Title

			listRow := &waProto.ListMessage_Row{
				RowID: &rowID,
				Title: &rowTitle,
			}

			// Add description if provided
			if row.Description != nil && *row.Description != "" {
				listRow.Description = row.Description
			}

			rows = append(rows, listRow)
		}

		sectionTitle := section.Title
		sections = append(sections, &waProto.ListMessage_Section{
			Title: &sectionTitle,
			Rows:  rows,
		})
	}

	// Build message
	buttonText := "Select an option" // Default button text
	msg := &waProto.Message{
		ListMessage: &waProto.ListMessage{
			Title:       content.Header,
			Description: &content.Body,
			ButtonText:  &buttonText,
			Sections:    sections,
			ListType:    waProto.ListMessage_SINGLE_SELECT.Enum(),
		},
	}

	// Add footer if provided
	if content.Footer != nil && *content.Footer != "" {
		msg.ListMessage.FooterText = content.Footer
	}

	return msg, nil
}

// simulateTyping sends typing presence and waits for the specified duration
func (p *InteractiveProcessor) simulateTyping(client *wameow.Client, recipientJID types.JID, delayMs int64) error {
	if err := client.SendChatPresence(context.Background(), recipientJID, types.ChatPresenceComposing, types.ChatPresenceMediaText); err != nil {
		return err
	}

	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	if err := client.SendChatPresence(context.Background(), recipientJID, types.ChatPresencePaused, types.ChatPresenceMediaText); err != nil {
		return err
	}

	return nil
}
