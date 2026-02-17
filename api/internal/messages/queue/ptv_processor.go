package queue

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/protobuf/proto"

	wameow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/events/echo"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

// PTVProcessor handles Push-To-Talk Video (circular video) message sending via WhatsApp
// PTV messages are short circular video clips, similar to video notes in other apps
// They are displayed as circular playable thumbnails in WhatsApp chats
type PTVProcessor struct {
	log             *slog.Logger
	mediaDownloader *MediaDownloader
	thumbGenerator  *ThumbnailGenerator
	presenceHelper  *PresenceHelper
	echoEmitter     *echo.Emitter
}

// NewPTVProcessor creates a new PTV (circular video) message processor
func NewPTVProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *PTVProcessor {
	return &PTVProcessor{
		log:             log.With(slog.String("processor", "ptv")),
		mediaDownloader: NewMediaDownloader(16), // 16MB max for PTV (short clips)
		presenceHelper:  NewPresenceHelper(),
		echoEmitter:     echoEmitter,
	}
}

// Process sends a PTV (circular video) message via WhatsApp
func (p *PTVProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.PTVContent == nil {
		return fmt.Errorf("ptv_content is required for PTV messages")
	}

	// Parse recipient JID
	recipientJID, err := types.ParseJID(args.Phone)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Simulate recording indicator if DelayTyping is set
	// For PTV, we use recording audio indicator (similar behavior)
	if args.DelayTyping > 0 {
		if err := p.presenceHelper.SimulateRecording(client, recipientJID, args.DelayTyping); err != nil {
			p.log.Warn("failed to send recording indicator",
				slog.String("error", err.Error()),
				slog.String("phone", args.Phone))
		}
	}

	// Download or decode video data
	videoData, mimeType, err := p.mediaDownloader.Download(args.PTVContent.MediaURL)
	if err != nil {
		return fmt.Errorf("download PTV video: %w", err)
	}

	// Validate media type
	if err := p.mediaDownloader.ValidateMediaType(mimeType, MediaTypeVideo); err != nil {
		return fmt.Errorf("invalid media type for PTV: %w", err)
	}

	// Detect video dimensions
	width, height, err := GetMediaDimensions(videoData, mimeType)
	if err != nil {
		p.log.Warn("failed to detect PTV dimensions, using default square",
			slog.String("error", err.Error()),
			slog.String("mime_type", mimeType))
		// PTV videos are typically square for circular display
		width, height = 240, 240
	} else {
		p.log.Debug("detected PTV dimensions",
			slog.Int("width", width),
			slog.Int("height", height))
	}

	// Extract video duration
	duration, err := GetVideoDuration(videoData)
	if err != nil {
		p.log.Warn("failed to detect PTV duration",
			slog.String("error", err.Error()),
			slog.String("mime_type", mimeType))
		duration = 0
	}

	// Generate and upload thumbnail (lazy initialization)
	var thumbnail *ThumbnailResult
	if p.thumbGenerator == nil {
		p.thumbGenerator = NewThumbnailGenerator(client, p.log)
	}
	thumbnail, err = p.thumbGenerator.GenerateAndUploadVideoThumbnail(ctx, videoData)
	if err != nil {
		p.log.Warn("failed to generate PTV thumbnail",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Upload video to WhatsApp servers
	uploaded, err := client.Upload(ctx, videoData, wameow.MediaVideo)
	if err != nil {
		return fmt.Errorf("upload PTV video: %w", err)
	}

	// Build ContextInfo using helper
	contextBuilder := NewContextInfoBuilder(client, recipientJID, args, p.log)
	contextInfo, err := contextBuilder.Build(ctx)
	if err != nil {
		p.log.Warn("failed to build context info, sending without it",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Build PTV message
	msg := p.buildMessage(args, uploaded, mimeType, width, height, duration, thumbnail, contextInfo)

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg, BuildSendExtra(args))
	if err != nil {
		return fmt.Errorf("send PTV message: %w", err)
	}

	p.log.Info("PTV message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.Int("width", width),
		slog.Int("height", height),
		slog.Int64("duration_seconds", duration),
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
			MessageType:       "ptv",
			MediaType:         "video",
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

// buildMessage constructs the PTV message proto
// PTV uses VideoMessage wrapped in PtvMessage field
func (p *PTVProcessor) buildMessage(
	args *SendMessageArgs,
	uploaded wameow.UploadResponse,
	mimeType string,
	width, height int,
	duration int64,
	thumbnail *ThumbnailResult,
	contextInfo *waProto.ContextInfo,
) *waProto.Message {
	videoMsg := &waProto.VideoMessage{
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

	// Add video duration
	if duration > 0 {
		videoMsg.Seconds = proto.Uint32(uint32(duration))
	}

	// Add video dimensions
	if width > 0 && height > 0 {
		videoMsg.Width = proto.Uint32(uint32(width))
		videoMsg.Height = proto.Uint32(uint32(height))
	}

	// PTV videos typically don't have captions, but support it if provided
	if args.PTVContent.Caption != nil && *args.PTVContent.Caption != "" {
		videoMsg.Caption = args.PTVContent.Caption
	}

	// Add thumbnail if generated successfully
	if thumbnail != nil {
		videoMsg.JPEGThumbnail = thumbnail.Data
		if thumbnail.DirectPath != "" {
			videoMsg.ThumbnailDirectPath = proto.String(thumbnail.DirectPath)
		}
		if len(thumbnail.FileSha256) > 0 {
			videoMsg.ThumbnailSHA256 = thumbnail.FileSha256
		}
		if len(thumbnail.FileEncSha256) > 0 {
			videoMsg.ThumbnailEncSHA256 = thumbnail.FileEncSha256
		}
	}

	// PTV uses PtvMessage wrapper instead of VideoMessage
	return &waProto.Message{
		PtvMessage: videoMsg,
	}
}
