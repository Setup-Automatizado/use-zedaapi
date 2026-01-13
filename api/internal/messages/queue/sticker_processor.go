package queue

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	wameow "go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow/api/internal/events/echo"
)

// StickerProcessor handles sticker message sending via WhatsApp
// Stickers are sent as waProto.StickerMessage with WebP format
// Supports conversion from PNG, JPG, GIF to WebP with proper sizing
type StickerProcessor struct {
	log              *slog.Logger
	mediaDownloader  *MediaDownloader
	thumbGenerator   *ThumbnailGenerator
	stickerConverter *StickerConverter
	presenceHelper   *PresenceHelper
	echoEmitter      *echo.Emitter
}

// NewStickerProcessor creates a new sticker message processor
func NewStickerProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *StickerProcessor {
	return &StickerProcessor{
		log:              log.With(slog.String("processor", "sticker")),
		mediaDownloader:  NewMediaDownloader(5), // 5MB max for stickers (WebP is small)
		stickerConverter: NewStickerConverter(log),
		presenceHelper:   NewPresenceHelper(),
		echoEmitter:      echoEmitter,
	}
}

// Process sends a sticker message via WhatsApp
func (p *StickerProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.StickerContent == nil {
		return fmt.Errorf("sticker_content is required for sticker messages")
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

	// Download or decode sticker data
	stickerData, mimeType, err := p.mediaDownloader.Download(args.StickerContent.MediaURL)
	if err != nil {
		return fmt.Errorf("download sticker: %w", err)
	}

	// Validate that it's an image format
	if !p.isImageFormat(mimeType) {
		return fmt.Errorf("invalid sticker format: %s (expected image)", mimeType)
	}

	// Convert to WebP format if needed
	var finalData []byte
	var width, height int

	conversionResult, err := p.stickerConverter.Convert(stickerData, mimeType)
	if err != nil {
		p.log.Warn("sticker conversion failed, attempting to use original data",
			slog.String("error", err.Error()),
			slog.String("original_mime", mimeType))

		// If conversion fails and it's already WebP, use original
		if p.stickerConverter.IsWebPFormat(mimeType) {
			finalData = stickerData
			width, height = 512, 512 // Default sticker size
		} else {
			return fmt.Errorf("sticker conversion failed: %w", err)
		}
	} else {
		finalData = conversionResult.Data
		width = conversionResult.Width
		height = conversionResult.Height
		mimeType = conversionResult.MimeType
	}

	p.log.Debug("sticker prepared for upload",
		slog.String("mime_type", mimeType),
		slog.Int("size_bytes", len(finalData)),
		slog.Int("width", width),
		slog.Int("height", height))

	// Generate and upload thumbnail (lazy initialization)
	var thumbnail *ThumbnailResult
	if p.thumbGenerator == nil {
		p.thumbGenerator = NewThumbnailGenerator(client, p.log)
	}
	thumbnail, err = p.thumbGenerator.GenerateAndUploadImageThumbnail(ctx, finalData, mimeType)
	if err != nil {
		p.log.Warn("failed to generate sticker thumbnail, sending without it",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Upload sticker to WhatsApp servers
	// Stickers use MediaImage upload type
	uploaded, err := client.Upload(ctx, finalData, wameow.MediaImage)
	if err != nil {
		return fmt.Errorf("upload sticker: %w", err)
	}

	// Build ContextInfo using helper
	contextBuilder := NewContextInfoBuilder(client, recipientJID, args, p.log)
	contextInfo, err := contextBuilder.Build(ctx)
	if err != nil {
		p.log.Warn("failed to build context info, sending without it",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Build sticker message
	msg := p.buildMessage(args, uploaded, width, height, thumbnail, contextInfo)

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send sticker message: %w", err)
	}

	p.log.Info("sticker message sent successfully",
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
			MessageType:       "sticker",
			MediaType:         "sticker",
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

// buildMessage constructs the sticker message proto
func (p *StickerProcessor) buildMessage(
	args *SendMessageArgs,
	uploaded wameow.UploadResponse,
	width, height int,
	thumbnail *ThumbnailResult,
	contextInfo *waProto.ContextInfo,
) *waProto.Message {
	stickerMsg := &waProto.StickerMessage{
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      proto.String("image/webp"),
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uploaded.FileLength),
		ContextInfo:   contextInfo,
		IsAvatar:      proto.Bool(false), // Regular sticker, not avatar
		IsAnimated:    proto.Bool(false), // Static sticker (for now)
		IsLottie:      proto.Bool(false), // Not a Lottie animation
	}

	// Add sticker dimensions
	if width > 0 && height > 0 {
		stickerMsg.Width = proto.Uint32(uint32(width))
		stickerMsg.Height = proto.Uint32(uint32(height))
	}

	// Add thumbnail if generated successfully
	if thumbnail != nil {
		stickerMsg.PngThumbnail = thumbnail.Data
	}

	return &waProto.Message{
		StickerMessage: stickerMsg,
	}
}

// isImageFormat checks if the MIME type is a supported image format for stickers
func (p *StickerProcessor) isImageFormat(mimeType string) bool {
	mimeType = strings.ToLower(mimeType)
	supportedFormats := []string{
		"image/png",
		"image/jpeg",
		"image/jpg",
		"image/gif",
		"image/webp",
		"image/bmp",
		"image/tiff",
	}

	for _, format := range supportedFormats {
		if strings.Contains(mimeType, strings.TrimPrefix(format, "image/")) {
			return true
		}
	}

	return false
}
