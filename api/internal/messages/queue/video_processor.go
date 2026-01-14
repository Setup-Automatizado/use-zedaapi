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

// VideoProcessor handles video message sending via WhatsApp
type VideoProcessor struct {
	log             *slog.Logger
	mediaDownloader *MediaDownloader
	thumbGenerator  *ThumbnailGenerator
	presenceHelper  *PresenceHelper
	echoEmitter     *echo.Emitter
}

// NewVideoProcessor creates a new video message processor
func NewVideoProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *VideoProcessor {
	return &VideoProcessor{
		log:             log.With(slog.String("processor", "video")),
		mediaDownloader: NewMediaDownloader(100), // 100MB max for videos
		presenceHelper:  NewPresenceHelper(),
		echoEmitter:     echoEmitter,
	}
}

// Process sends a video message via WhatsApp
func (p *VideoProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.VideoContent == nil {
		return fmt.Errorf("video_content is required for video messages")
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

	// Download or decode video data using helper
	videoData, mimeType, err := p.mediaDownloader.Download(args.VideoContent.MediaURL)
	if err != nil {
		return fmt.Errorf("download video: %w", err)
	}

	// Validate media type
	if err := p.mediaDownloader.ValidateMediaType(mimeType, MediaTypeVideo); err != nil {
		return fmt.Errorf("invalid media type: %w", err)
	}

	// Detect video dimensions
	width, height, err := GetMediaDimensions(videoData, mimeType)
	if err != nil {
		p.log.Warn("failed to detect video dimensions, sending without dimensions",
			slog.String("error", err.Error()),
			slog.String("mime_type", mimeType))
		width, height = 0, 0 // Will not set Width/Height fields
	} else {
		p.log.Debug("detected video dimensions",
			slog.Int("width", width),
			slog.Int("height", height))
	}

	// Generate and upload thumbnail (lazy initialization)
	// Note: Video thumbnail generation requires ffmpeg and is not yet implemented
	var thumbnail *ThumbnailResult
	if p.thumbGenerator == nil {
		p.thumbGenerator = NewThumbnailGenerator(client, p.log)
	}
	thumbnail, err = p.thumbGenerator.GenerateAndUploadVideoThumbnail(ctx, videoData)
	if err != nil {
		p.log.Warn("failed to generate video thumbnail (feature not yet implemented)",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Upload video to WhatsApp servers
	uploaded, err := client.Upload(ctx, videoData, wameow.MediaVideo)
	if err != nil {
		return fmt.Errorf("upload video: %w", err)
	}

	// Build ContextInfo using helper
	contextBuilder := NewContextInfoBuilder(client, recipientJID, args, p.log)
	contextInfo, err := contextBuilder.Build(ctx)
	if err != nil {
		p.log.Warn("failed to build context info, sending without it",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Build video message
	msg := p.buildMessage(args, uploaded, mimeType, width, height, thumbnail, contextInfo)

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send video message: %w", err)
	}

	p.log.Info("video message sent successfully",
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
			MessageType:       "video",
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

// buildMessage constructs the video message proto
func (p *VideoProcessor) buildMessage(
	args *SendMessageArgs,
	uploaded wameow.UploadResponse,
	mimeType string,
	width, height int,
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

	// Add video dimensions if detected
	if width > 0 && height > 0 {
		videoMsg.Width = proto.Uint32(uint32(width))
		videoMsg.Height = proto.Uint32(uint32(height))
	}

	// Add caption if provided
	if args.VideoContent.Caption != nil && *args.VideoContent.Caption != "" {
		videoMsg.Caption = args.VideoContent.Caption
	}

	// Add thumbnail if generated successfully
	if thumbnail != nil {
		videoMsg.JPEGThumbnail = thumbnail.Data
		videoMsg.ThumbnailDirectPath = proto.String(thumbnail.DirectPath)
		videoMsg.ThumbnailSHA256 = thumbnail.FileSha256
		videoMsg.ThumbnailEncSHA256 = thumbnail.FileEncSha256
	}

	if args.VideoContent != nil && args.VideoContent.IsPTV {
		return &waProto.Message{PtvMessage: videoMsg}
	}

	return &waProto.Message{VideoMessage: videoMsg}
}
