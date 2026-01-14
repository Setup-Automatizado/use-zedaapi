package queue

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/protobuf/proto"

	wameow "go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"

	"go.mau.fi/whatsmeow/api/internal/events/echo"
)

// ImageProcessor handles image message sending via WhatsApp
type ImageProcessor struct {
	log             *slog.Logger
	mediaDownloader *MediaDownloader
	thumbGenerator  *ThumbnailGenerator
	presenceHelper  *PresenceHelper
	echoEmitter     *echo.Emitter
}

// NewImageProcessor creates a new image message processor
func NewImageProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *ImageProcessor {
	return &ImageProcessor{
		log:             log.With(slog.String("processor", "image")),
		mediaDownloader: NewMediaDownloader(100), // 100MB max for images
		presenceHelper:  NewPresenceHelper(),
		echoEmitter:     echoEmitter,
	}
}

// Process sends an image message via WhatsApp
func (p *ImageProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.ImageContent == nil {
		return fmt.Errorf("image_content is required for image messages")
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

	// Download or decode image data using helper
	imageData, mimeType, err := p.mediaDownloader.Download(args.ImageContent.MediaURL)
	if err != nil {
		return fmt.Errorf("download image: %w", err)
	}

	// Validate media type
	if err := p.mediaDownloader.ValidateMediaType(mimeType, MediaTypeImage); err != nil {
		return fmt.Errorf("invalid media type: %w", err)
	}

	// Detect image dimensions
	width, height, err := GetMediaDimensions(imageData, mimeType)
	if err != nil {
		p.log.Warn("failed to detect image dimensions, sending without dimensions",
			slog.String("error", err.Error()),
			slog.String("mime_type", mimeType))
		width, height = 0, 0 // Will not set Width/Height fields
	} else {
		p.log.Debug("detected image dimensions",
			slog.Int("width", width),
			slog.Int("height", height))
	}

	// Generate and upload thumbnail (lazy initialization)
	var thumbnail *ThumbnailResult
	if p.thumbGenerator == nil {
		p.thumbGenerator = NewThumbnailGenerator(client, p.log)
	}
	thumbnail, err = p.thumbGenerator.GenerateAndUploadImageThumbnail(ctx, imageData, mimeType)
	if err != nil {
		p.log.Warn("failed to generate thumbnail, sending without it",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Upload image to WhatsApp servers
	uploaded, err := client.Upload(ctx, imageData, wameow.MediaImage)
	if err != nil {
		return fmt.Errorf("upload image: %w", err)
	}

	// Build ContextInfo using helper
	contextBuilder := NewContextInfoBuilder(client, recipientJID, args, p.log)
	contextInfo, err := contextBuilder.Build(ctx)
	if err != nil {
		p.log.Warn("failed to build context info, sending without it",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Build image message
	msg := p.buildMessage(args, uploaded, mimeType, width, height, thumbnail, contextInfo)

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send image message: %w", err)
	}

	p.log.Info("image message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.Int("width", width),
		slog.Int("height", height),
		slog.Int64("file_size", int64(uploaded.FileLength)),
		slog.Bool("has_thumbnail", thumbnail != nil),
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
			MessageType:       "image",
			MediaType:         "image",
			ZaapID:            args.ZaapID,
			HasMedia:          true,
		}
		if err := p.echoEmitter.EmitEcho(ctx, echoReq); err != nil {
			p.log.Warn("failed to emit API echo",
				slog.String("error", err.Error()),
				slog.String("zaap_id", args.ZaapID))
		}
	}

	return nil
}

// buildMessage constructs the image message proto
func (p *ImageProcessor) buildMessage(
	args *SendMessageArgs,
	uploaded wameow.UploadResponse,
	mimeType string,
	width, height int,
	thumbnail *ThumbnailResult,
	contextInfo *waProto.ContextInfo,
) *waProto.Message {
	imageMsg := &waProto.ImageMessage{
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      proto.String(mimeType),
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uploaded.FileLength),
		ViewOnce:      proto.Bool(args.ViewOnce),
		ContextInfo:   contextInfo,
	}

	// Add image dimensions if detected
	if width > 0 && height > 0 {
		imageMsg.Width = proto.Uint32(uint32(width))
		imageMsg.Height = proto.Uint32(uint32(height))
	}

	// Add caption if provided
	if args.ImageContent.Caption != nil && *args.ImageContent.Caption != "" {
		imageMsg.Caption = args.ImageContent.Caption
	}

	// Add thumbnail if generated successfully
	if thumbnail != nil {
		imageMsg.JPEGThumbnail = thumbnail.Data
		imageMsg.ThumbnailDirectPath = proto.String(thumbnail.DirectPath)
		imageMsg.ThumbnailSHA256 = thumbnail.FileSha256
		imageMsg.ThumbnailEncSHA256 = thumbnail.FileEncSha256
	}

	return &waProto.Message{
		ImageMessage: imageMsg,
	}
}
