package queue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	wameow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/events/echo"
	"go.mau.fi/whatsmeow/api/internal/messages/interactive"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// InteractiveZAPIProcessor handles interactive message sending via WhatsApp
// Uses modern waE2E.InteractiveMessage_NativeFlowMessage for buttons and actions
// Now includes full media processing with thumbnail generation for images/videos/documents
type InteractiveZAPIProcessor struct {
	protoBuilder   *interactive.ProtoBuilder
	presenceHelper *PresenceHelper
	mediaProcessor *InteractiveMediaProcessor
	log            *slog.Logger
	echoEmitter    *echo.Emitter
}

// NewInteractiveZAPIProcessor creates a new FUNNELCHAT interactive message processor
func NewInteractiveZAPIProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *InteractiveZAPIProcessor {
	return &InteractiveZAPIProcessor{
		protoBuilder:   interactive.NewProtoBuilder(),
		presenceHelper: NewPresenceHelper(),
		mediaProcessor: nil, // Will be initialized per-client
		log:            log.With(slog.String("processor", "interactive_zapi")),
		echoEmitter:    echoEmitter,
	}
}

// NewInteractiveZAPIProcessorWithClient creates a new FUNNELCHAT interactive message processor with a WhatsApp client
// for media processing capabilities (thumbnail generation, upload, etc.)
func NewInteractiveZAPIProcessorWithClient(client *wameow.Client, log *slog.Logger, echoEmitter *echo.Emitter) *InteractiveZAPIProcessor {
	return &InteractiveZAPIProcessor{
		protoBuilder:   interactive.NewProtoBuilder(),
		presenceHelper: NewPresenceHelper(),
		mediaProcessor: NewInteractiveMediaProcessor(client, log),
		log:            log.With(slog.String("processor", "interactive_zapi")),
		echoEmitter:    echoEmitter,
	}
}

// SetClient sets the WhatsApp client for media processing
// Call this when the client becomes available after initial construction
func (p *InteractiveZAPIProcessor) SetClient(client *wameow.Client) {
	if client != nil {
		p.mediaProcessor = NewInteractiveMediaProcessor(client, p.log)
	}
}

// ProcessButtonList sends a button list message via WhatsApp (FUNNELCHAT /send-button-list)
func (p *InteractiveZAPIProcessor) ProcessButtonList(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.InteractiveContent == nil {
		return fmt.Errorf("interactive_content is required for button_list messages")
	}

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

	// Convert queue model to FUNNELCHAT request format
	req := p.convertToButtonListRequest(args)

	// Process media if present (image or video)
	var mediaHeader *waProto.InteractiveMessage_Header
	mediaURL := p.getMediaURLFromButtonList(args)
	if mediaURL != "" {
		// Initialize media processor with client if needed
		if p.mediaProcessor == nil {
			p.mediaProcessor = NewInteractiveMediaProcessor(client, p.log)
		}

		p.log.Debug("processing media for button list",
			slog.String("zaap_id", args.ZaapID),
			slog.Bool("has_image", args.InteractiveContent.Image != nil),
			slog.Bool("has_video", args.InteractiveContent.Video != nil))

		processedMedia, err := p.mediaProcessor.ProcessMediaURL(ctx, mediaURL)
		if err != nil {
			p.log.Warn("failed to process media for button list, continuing without media",
				slog.String("error", err.Error()),
				slog.String("zaap_id", args.ZaapID))
		} else {
			mediaHeader = p.mediaProcessor.ProcessedMediaToHeader(processedMedia, "")
			p.log.Info("media processed successfully for button list",
				slog.String("zaap_id", args.ZaapID),
				slog.String("media_type", string(processedMedia.MediaType)),
				slog.Bool("has_thumbnail", len(processedMedia.JPEGThumbnail) > 0))
		}
	}

	// Build proto message using our builder with media support
	var msg *waE2E.Message
	if mediaHeader != nil {
		msg, err = p.protoBuilder.BuildButtonListMessageWithMedia(req, mediaHeader)
	} else {
		msg, err = p.protoBuilder.BuildButtonListMessage(req)
	}
	if err != nil {
		return fmt.Errorf("build button list message: %w", err)
	}

	// Add reply context if provided
	if args.ReplyToMessageID != "" {
		p.protoBuilder.AddContextInfo(msg, args.ReplyToMessageID, args.Phone)
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send button list message: %w", err)
	}

	p.log.Info("button list message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.Int("button_count", len(args.InteractiveContent.Buttons)),
		slog.Bool("has_media", mediaHeader != nil),
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
			MessageType:       "button_list",
			ZaapID:            args.ZaapID,
			HasMedia:          mediaHeader != nil,
		}
		if err := p.echoEmitter.EmitEcho(ctx, echoReq); err != nil {
			p.log.Warn("failed to emit API echo",
				slog.String("error", err.Error()),
				slog.String("zaap_id", args.ZaapID))
		}
	}

	return nil
}

// getMediaURLFromButtonList extracts the media URL from button list content
func (p *InteractiveZAPIProcessor) getMediaURLFromButtonList(args *SendMessageArgs) string {
	if args.InteractiveContent == nil {
		return ""
	}
	if args.InteractiveContent.Image != nil && *args.InteractiveContent.Image != "" {
		return *args.InteractiveContent.Image
	}
	if args.InteractiveContent.Video != nil && *args.InteractiveContent.Video != "" {
		return *args.InteractiveContent.Video
	}
	return ""
}

// ProcessButtonActions sends a button actions message via WhatsApp (FUNNELCHAT /send-button-actions)
func (p *InteractiveZAPIProcessor) ProcessButtonActions(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.InteractiveContent == nil {
		return fmt.Errorf("interactive_content is required for button_actions messages")
	}

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

	// Convert queue model to FUNNELCHAT request format
	req := p.convertToButtonActionsRequest(args)

	// Process media if present (image, video, or document)
	var mediaHeader *waProto.InteractiveMessage_Header
	mediaURL := p.getMediaURLFromButtonActions(args)
	if mediaURL != "" {
		// Initialize media processor with client if needed
		if p.mediaProcessor == nil {
			p.mediaProcessor = NewInteractiveMediaProcessor(client, p.log)
		}

		p.log.Debug("processing media for button actions",
			slog.String("zaap_id", args.ZaapID),
			slog.Bool("has_image", args.InteractiveContent.Image != nil),
			slog.Bool("has_video", args.InteractiveContent.Video != nil),
			slog.Bool("has_document", args.InteractiveContent.Document != nil))

		processedMedia, err := p.mediaProcessor.ProcessMediaURL(ctx, mediaURL)
		if err != nil {
			p.log.Warn("failed to process media for button actions, continuing without media",
				slog.String("error", err.Error()),
				slog.String("zaap_id", args.ZaapID))
		} else {
			// Get document filename if applicable
			docFilename := ""
			if args.InteractiveContent.Document != nil && args.InteractiveContent.DocumentFilename != nil {
				docFilename = *args.InteractiveContent.DocumentFilename
			}
			mediaHeader = p.mediaProcessor.ProcessedMediaToHeader(processedMedia, docFilename)
			p.log.Info("media processed successfully for button actions",
				slog.String("zaap_id", args.ZaapID),
				slog.String("media_type", string(processedMedia.MediaType)),
				slog.Bool("has_thumbnail", len(processedMedia.JPEGThumbnail) > 0))
		}
	}

	// Build proto message using our builder with media support
	var msg *waE2E.Message
	if mediaHeader != nil {
		msg, err = p.protoBuilder.BuildButtonActionsMessageWithMedia(req, mediaHeader)
	} else {
		msg, err = p.protoBuilder.BuildButtonActionsMessage(req)
	}
	if err != nil {
		return fmt.Errorf("build button actions message: %w", err)
	}

	// Add reply context if provided
	if args.ReplyToMessageID != "" {
		p.protoBuilder.AddContextInfo(msg, args.ReplyToMessageID, args.Phone)
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send button actions message: %w", err)
	}

	p.log.Info("button actions message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.Int("button_count", len(args.InteractiveContent.Buttons)),
		slog.Bool("has_media", mediaHeader != nil),
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
			MessageType:       "button_actions",
			ZaapID:            args.ZaapID,
			HasMedia:          mediaHeader != nil,
		}
		if err := p.echoEmitter.EmitEcho(ctx, echoReq); err != nil {
			p.log.Warn("failed to emit API echo",
				slog.String("error", err.Error()),
				slog.String("zaap_id", args.ZaapID))
		}
	}

	return nil
}

// getMediaURLFromButtonActions extracts the media URL from button actions content
func (p *InteractiveZAPIProcessor) getMediaURLFromButtonActions(args *SendMessageArgs) string {
	if args.InteractiveContent == nil {
		return ""
	}
	if args.InteractiveContent.Image != nil && *args.InteractiveContent.Image != "" {
		return *args.InteractiveContent.Image
	}
	if args.InteractiveContent.Video != nil && *args.InteractiveContent.Video != "" {
		return *args.InteractiveContent.Video
	}
	if args.InteractiveContent.Document != nil && *args.InteractiveContent.Document != "" {
		return *args.InteractiveContent.Document
	}
	return ""
}

// ProcessOptionList sends an option list message via WhatsApp (FUNNELCHAT /send-option-list)
func (p *InteractiveZAPIProcessor) ProcessOptionList(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.InteractiveContent == nil {
		return fmt.Errorf("interactive_content is required for option_list messages")
	}

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

	// Convert queue model to FUNNELCHAT request format
	req := p.convertToOptionListRequest(args)

	// Build proto message using our builder
	msg, err := p.protoBuilder.BuildOptionListMessage(req)
	if err != nil {
		return fmt.Errorf("build option list message: %w", err)
	}

	// Add reply context if provided
	if args.ReplyToMessageID != "" {
		p.protoBuilder.AddContextInfo(msg, args.ReplyToMessageID, args.Phone)
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send option list message: %w", err)
	}

	p.log.Info("option list message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.Int("section_count", len(args.InteractiveContent.Sections)),
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
			MessageType:       "option_list",
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

// ProcessButtonPIX sends a PIX payment button message via WhatsApp (FUNNELCHAT /send-button-pix)
func (p *InteractiveZAPIProcessor) ProcessButtonPIX(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.InteractiveContent == nil || args.InteractiveContent.PIXPayment == nil {
		return fmt.Errorf("interactive_content with pix_payment is required for button_pix messages")
	}

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

	// Convert queue model to FUNNELCHAT request format
	req := p.convertToButtonPIXRequest(args)

	// Build proto message using our builder
	msg, err := p.protoBuilder.BuildPIXButtonMessage(req)
	if err != nil {
		return fmt.Errorf("build button pix message: %w", err)
	}

	// Add reply context if provided
	if args.ReplyToMessageID != "" {
		p.protoBuilder.AddContextInfo(msg, args.ReplyToMessageID, args.Phone)
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send button pix message: %w", err)
	}

	p.log.Info("button pix message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.String("pix_key_type", args.InteractiveContent.PIXPayment.KeyType),
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
			MessageType:       "button_pix",
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

// ProcessButtonOTP sends an OTP button message via WhatsApp (FUNNELCHAT /send-button-otp)
func (p *InteractiveZAPIProcessor) ProcessButtonOTP(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.InteractiveContent == nil || args.InteractiveContent.OTPCode == nil {
		return fmt.Errorf("interactive_content with otp_code is required for button_otp messages")
	}

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

	// Convert queue model to FUNNELCHAT request format
	req := p.convertToButtonOTPRequest(args)

	// Build proto message using our builder
	msg, err := p.protoBuilder.BuildOTPButtonMessage(req)
	if err != nil {
		return fmt.Errorf("build button otp message: %w", err)
	}

	// Add reply context if provided
	if args.ReplyToMessageID != "" {
		p.protoBuilder.AddContextInfo(msg, args.ReplyToMessageID, args.Phone)
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send button otp message: %w", err)
	}

	p.log.Info("button otp message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
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
			MessageType:       "button_otp",
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

// ProcessCarousel sends a carousel message with multiple cards via WhatsApp (FUNNELCHAT /send-carousel)
func (p *InteractiveZAPIProcessor) ProcessCarousel(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.InteractiveContent == nil || len(args.InteractiveContent.CarouselCards) == 0 {
		return fmt.Errorf("interactive_content with carousel_cards is required for carousel messages")
	}

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

	// Convert queue model to FUNNELCHAT request format
	req := p.convertToCarouselRequest(args)

	// Process media for each card that has media URL
	var cardMediaHeaders []*waProto.InteractiveMessage_Header
	hasAnyMedia := false

	// Initialize media processor with client if needed
	if p.mediaProcessor == nil {
		p.mediaProcessor = NewInteractiveMediaProcessor(client, p.log)
	}

	for i, card := range args.InteractiveContent.CarouselCards {
		if card.MediaURL == "" {
			cardMediaHeaders = append(cardMediaHeaders, nil)
			continue
		}

		p.log.Debug("processing media for carousel card",
			slog.String("zaap_id", args.ZaapID),
			slog.Int("card_index", i),
			slog.String("media_url", card.MediaURL))

		processedMedia, err := p.mediaProcessor.ProcessMediaURL(ctx, card.MediaURL)
		if err != nil {
			p.log.Warn("failed to process media for carousel card, continuing without media",
				slog.String("error", err.Error()),
				slog.String("zaap_id", args.ZaapID),
				slog.Int("card_index", i))
			cardMediaHeaders = append(cardMediaHeaders, nil)
			continue
		}

		mediaHeader := p.mediaProcessor.ProcessedMediaToHeader(processedMedia, "")
		cardMediaHeaders = append(cardMediaHeaders, mediaHeader)
		hasAnyMedia = true

		p.log.Info("media processed successfully for carousel card",
			slog.String("zaap_id", args.ZaapID),
			slog.Int("card_index", i),
			slog.String("media_type", string(processedMedia.MediaType)),
			slog.Bool("has_thumbnail", len(processedMedia.JPEGThumbnail) > 0))
	}

	// Build proto message using our builder with media support
	var msg *waE2E.Message
	if hasAnyMedia {
		msg, err = p.protoBuilder.BuildCarouselMessageWithMedia(req, cardMediaHeaders)
	} else {
		msg, err = p.protoBuilder.BuildCarouselMessage(req)
	}
	if err != nil {
		return fmt.Errorf("build carousel message: %w", err)
	}

	// Add reply context if provided
	if args.ReplyToMessageID != "" {
		p.protoBuilder.AddContextInfo(msg, args.ReplyToMessageID, args.Phone)
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send carousel message: %w", err)
	}

	p.log.Info("carousel message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.Int("card_count", len(args.InteractiveContent.CarouselCards)),
		slog.Bool("has_media", hasAnyMedia),
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
			MessageType:       "carousel",
			ZaapID:            args.ZaapID,
			HasMedia:          hasAnyMedia,
		}
		if err := p.echoEmitter.EmitEcho(ctx, echoReq); err != nil {
			p.log.Warn("failed to emit API echo",
				slog.String("error", err.Error()),
				slog.String("zaap_id", args.ZaapID))
		}
	}

	return nil
}

// Conversion helpers - convert from queue models to FUNNELCHAT request formats

func (p *InteractiveZAPIProcessor) convertToButtonListRequest(args *SendMessageArgs) *interactive.SendButtonListRequest {
	content := args.InteractiveContent

	req := &interactive.SendButtonListRequest{
		Phone:   args.Phone,
		Message: content.Body,
		Title:   content.Header,
		Footer:  content.Footer,
		Image:   content.Image,
		Video:   content.Video,
	}

	// Convert buttons
	req.ButtonList.Buttons = make([]interactive.ButtonListItem, len(content.Buttons))
	for i, btn := range content.Buttons {
		req.ButtonList.Buttons[i] = interactive.ButtonListItem{
			ID:    btn.ID,
			Label: btn.Title,
		}
	}

	return req
}

func (p *InteractiveZAPIProcessor) convertToButtonActionsRequest(args *SendMessageArgs) *interactive.SendButtonActionsRequest {
	content := args.InteractiveContent

	req := &interactive.SendButtonActionsRequest{
		Phone:   args.Phone,
		Message: content.Body,
		Title:   content.Header,
		Footer:  content.Footer,
	}

	// Convert buttons with action types
	req.ButtonActions.Buttons = make([]interactive.ActionButton, len(content.Buttons))
	for i, btn := range content.Buttons {
		actionBtn := interactive.ActionButton{
			ID:    btn.ID,
			Label: btn.Title,
			Type:  interactive.ButtonType(btn.Type),
		}

		// Set type-specific fields
		if btn.URL != "" {
			url := btn.URL
			actionBtn.URL = &url
		}
		if btn.Phone != "" {
			phone := btn.Phone
			actionBtn.Phone = &phone
		}
		if btn.CopyCode != "" {
			copyCode := btn.CopyCode
			actionBtn.CopyCode = &copyCode
		}

		req.ButtonActions.Buttons[i] = actionBtn
	}

	return req
}

func (p *InteractiveZAPIProcessor) convertToOptionListRequest(args *SendMessageArgs) *interactive.SendOptionListRequest {
	content := args.InteractiveContent

	buttonLabel := "Select"
	if content.ButtonLabel != nil {
		buttonLabel = *content.ButtonLabel
	}

	req := &interactive.SendOptionListRequest{
		Phone:       args.Phone,
		Message:     content.Body,
		Title:       content.Header,
		Footer:      content.Footer,
		ButtonLabel: buttonLabel,
	}

	// Convert sections
	req.OptionList.Sections = make([]interactive.OptionSection, len(content.Sections))
	for i, sec := range content.Sections {
		optSection := interactive.OptionSection{
			Title: sec.Title,
			Rows:  make([]interactive.OptionRow, len(sec.Rows)),
		}

		for j, row := range sec.Rows {
			optSection.Rows[j] = interactive.OptionRow{
				ID:          row.ID,
				Title:       row.Title,
				Description: row.Description,
			}
		}

		req.OptionList.Sections[i] = optSection
	}

	return req
}

func (p *InteractiveZAPIProcessor) convertToButtonPIXRequest(args *SendMessageArgs) *interactive.SendButtonPIXRequest {
	content := args.InteractiveContent
	pix := content.PIXPayment

	var message *string
	if content.Body != "" {
		message = &content.Body
	}

	req := &interactive.SendButtonPIXRequest{
		Phone:         args.Phone,
		Message:       message,
		PIXKey:        pix.Key,
		Type:          interactive.PIXKeyType(pix.KeyType),
		Name:          pix.Name,
		Amount:        pix.Amount,
		TransactionID: pix.TransactionID,
	}

	return req
}

func (p *InteractiveZAPIProcessor) convertToButtonOTPRequest(args *SendMessageArgs) *interactive.SendButtonOTPRequest {
	content := args.InteractiveContent

	req := &interactive.SendButtonOTPRequest{
		Phone:   args.Phone,
		Message: content.Body,
		Code:    *content.OTPCode,
		Title:   content.Header,
		Footer:  content.Footer,
	}

	return req
}

func (p *InteractiveZAPIProcessor) convertToCarouselRequest(args *SendMessageArgs) *interactive.SendCarouselRequest {
	content := args.InteractiveContent

	req := &interactive.SendCarouselRequest{
		Phone:    args.Phone,
		Message:  content.Body, // Carousel body text displayed above cards
		Cards:    make([]interactive.CarouselCard, len(content.CarouselCards)),
		CardType: interactive.CarouselCardType(content.CarouselCardType),
	}

	for i, card := range content.CarouselCards {
		carouselCard := interactive.CarouselCard{
			Body: interactive.CarouselBody{
				Text: card.Body,
			},
			Buttons: make([]interactive.ActionButton, len(card.Buttons)),
		}

		// Set header if provided
		if card.Header != "" {
			carouselCard.Header = &interactive.CarouselHeader{
				Text: card.Header,
			}
		}

		// Set footer if provided
		if card.Footer != "" {
			carouselCard.Footer = &interactive.CarouselFooter{
				Text: card.Footer,
			}
		}

		// Set media URL if provided
		if card.MediaURL != "" {
			carouselCard.MediaURL = &card.MediaURL
		}

		// Convert buttons
		for j, btn := range card.Buttons {
			actionBtn := interactive.ActionButton{
				ID:    btn.ID,
				Label: btn.Title,
				Type:  interactive.ButtonType(btn.Type),
			}

			// Set type-specific fields
			if btn.URL != "" {
				url := btn.URL
				actionBtn.URL = &url
			}
			if btn.Phone != "" {
				phone := btn.Phone
				actionBtn.Phone = &phone
			}
			if btn.CopyCode != "" {
				copyCode := btn.CopyCode
				actionBtn.CopyCode = &copyCode
			}

			carouselCard.Buttons[j] = actionBtn
		}

		req.Cards[i] = carouselCard
	}

	return req
}

// simulateTypingIndicator sends typing presence and waits for the specified duration
func (p *InteractiveZAPIProcessor) simulateTypingIndicator(client *wameow.Client, recipientJID types.JID, delayMs int64) error {
	if err := client.SendChatPresence(context.Background(), recipientJID, types.ChatPresenceComposing, types.ChatPresenceMediaText); err != nil {
		return err
	}

	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	if err := client.SendChatPresence(context.Background(), recipientJID, types.ChatPresencePaused, types.ChatPresenceMediaText); err != nil {
		return err
	}

	return nil
}

// Ensure proto message type is compatible - SendMessage expects *waE2E.Message
var _ *waE2E.Message = (*waE2E.Message)(nil)
