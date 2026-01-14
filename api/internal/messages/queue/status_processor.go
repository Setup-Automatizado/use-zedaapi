package queue

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"

	wameow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/events/echo"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

// StatusProcessor handles WhatsApp Status/Stories message sending
// Status messages are always sent to status@broadcast JID
// which automatically broadcasts to viewers based on privacy settings
type StatusProcessor struct {
	log             *slog.Logger
	mediaDownloader *MediaDownloader
	thumbGenerator  *ThumbnailGenerator
	audioConverter  *AudioConverter
	echoEmitter     *echo.Emitter
}

// NewStatusProcessor creates a new status message processor
func NewStatusProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *StatusProcessor {
	return &StatusProcessor{
		log:             log.With(slog.String("processor", "status")),
		mediaDownloader: NewMediaDownloader(100), // 100MB max
		audioConverter:  NewAudioConverter(log),
		echoEmitter:     echoEmitter,
	}
}

// ProcessText sends a text status/story to status@broadcast
// Supports optional background color and font styling
func (p *StatusProcessor) ProcessText(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	// CRITICAL: Status messages always go to StatusBroadcastJID
	recipientJID := types.StatusBroadcastJID

	// Get text content - can come from TextStatusContent (styled) or TextContent (plain)
	var text string
	var backgroundColor *uint32
	var font *waProto.ExtendedTextMessage_FontType

	if args.TextStatusContent != nil {
		text = args.TextStatusContent.Text

		// Parse background color (ARGB hex format: "0xFFRRGGBB" or "#RRGGBB")
		if args.TextStatusContent.BackgroundColor != "" {
			bgColor := p.parseBackgroundColor(args.TextStatusContent.BackgroundColor)
			if bgColor != 0 {
				backgroundColor = proto.Uint32(bgColor)
			}
		}

		// Parse font (0-5)
		if args.TextStatusContent.Font != nil {
			fontType := waProto.ExtendedTextMessage_FontType(*args.TextStatusContent.Font)
			font = &fontType
		}
	} else if args.TextContent != nil {
		text = args.TextContent.Message
	} else {
		return fmt.Errorf("text_status_content or text_content is required for text status")
	}

	if text == "" {
		return fmt.Errorf("text content cannot be empty for text status")
	}

	// Build ExtendedTextMessage for status
	extendedText := &waProto.ExtendedTextMessage{
		Text: proto.String(text),
	}

	// Add background color if specified
	if backgroundColor != nil {
		extendedText.BackgroundArgb = backgroundColor
	}

	// Add font if specified
	if font != nil {
		extendedText.Font = font
	}

	msg := &waProto.Message{
		ExtendedTextMessage: extendedText,
	}

	// Send to status@broadcast
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send text status: %w", err)
	}

	p.log.Info("text status sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("recipient", recipientJID.String()),
		slog.String("whatsapp_message_id", resp.ID),
		slog.Bool("has_background_color", backgroundColor != nil),
		slog.Bool("has_custom_font", font != nil),
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
			MessageType:       "status_text",
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

// ProcessImage sends an image status/story to status@broadcast
func (p *StatusProcessor) ProcessImage(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.ImageContent == nil {
		return fmt.Errorf("image_content is required for image status")
	}

	// CRITICAL: Status messages always go to StatusBroadcastJID
	recipientJID := types.StatusBroadcastJID

	// Download or decode image data
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
		p.log.Warn("failed to detect image dimensions",
			slog.String("error", err.Error()))
		width, height = 0, 0
	}

	// Generate thumbnail (lazy initialization)
	var thumbnail *ThumbnailResult
	if p.thumbGenerator == nil {
		p.thumbGenerator = NewThumbnailGenerator(client, p.log)
	}
	thumbnail, err = p.thumbGenerator.GenerateAndUploadImageThumbnail(ctx, imageData, mimeType)
	if err != nil {
		p.log.Warn("failed to generate thumbnail for status",
			slog.String("error", err.Error()))
	}

	// Upload image to WhatsApp servers
	uploaded, err := client.Upload(ctx, imageData, wameow.MediaImage)
	if err != nil {
		return fmt.Errorf("upload image: %w", err)
	}

	// Build image message
	imageMsg := &waProto.ImageMessage{
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      proto.String(mimeType),
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uploaded.FileLength),
	}

	// Add dimensions if detected
	if width > 0 && height > 0 {
		imageMsg.Width = proto.Uint32(uint32(width))
		imageMsg.Height = proto.Uint32(uint32(height))
	}

	// Add caption if provided
	if args.ImageContent.Caption != nil && *args.ImageContent.Caption != "" {
		imageMsg.Caption = args.ImageContent.Caption
	}

	// Add thumbnail if generated
	if thumbnail != nil {
		imageMsg.JPEGThumbnail = thumbnail.Data
		imageMsg.ThumbnailDirectPath = proto.String(thumbnail.DirectPath)
		imageMsg.ThumbnailSHA256 = thumbnail.FileSha256
		imageMsg.ThumbnailEncSHA256 = thumbnail.FileEncSha256
	}

	msg := &waProto.Message{
		ImageMessage: imageMsg,
	}

	// Send to status@broadcast
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send image status: %w", err)
	}

	p.log.Info("image status sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("recipient", recipientJID.String()),
		slog.String("whatsapp_message_id", resp.ID),
		slog.Int("width", width),
		slog.Int("height", height),
		slog.Int64("file_size", int64(uploaded.FileLength)),
		slog.Bool("has_thumbnail", thumbnail != nil),
		slog.Bool("has_caption", args.ImageContent.Caption != nil),
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
			MessageType:       "status_image",
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

// ProcessAudio sends an audio status/story with WAVEFORM visualization to status@broadcast
// The waveform (64 samples) is critical for proper voice note display in WhatsApp Stories
func (p *StatusProcessor) ProcessAudio(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.AudioContent == nil {
		return fmt.Errorf("audio_content is required for audio status")
	}

	// CRITICAL: Status messages always go to StatusBroadcastJID
	recipientJID := types.StatusBroadcastJID

	// Download or decode audio data
	audioData, mimeType, err := p.mediaDownloader.Download(args.AudioContent.MediaURL)
	if err != nil {
		return fmt.Errorf("download audio: %w", err)
	}

	// Validate media type
	if err := p.mediaDownloader.ValidateMediaType(mimeType, MediaTypeAudio); err != nil {
		return fmt.Errorf("invalid media type: %w", err)
	}

	// Convert audio to WhatsApp native Opus/OGG format for voice note experience
	// This is essential for showing the waveform visualization in status
	var audioDataToUpload []byte
	var finalMimeType string
	var duration int64
	var waveform []byte

	if p.audioConverter.IsOpusFormat(mimeType) {
		// Already in Opus/OGG format
		audioDataToUpload = audioData
		finalMimeType = "audio/ogg; codecs=opus"
		waveform = p.audioConverter.GenerateWaveformFromData(audioData, mimeType)
		duration = p.audioConverter.GetDurationFromData(audioData, mimeType)

		p.log.Debug("audio already in Opus/OGG format for status",
			slog.Int("waveform_samples", len(waveform)),
			slog.Int64("duration_seconds", duration))
	} else {
		// Convert to Opus/OGG for native voice note experience with waveform
		p.log.Info("converting audio to Opus/OGG for status",
			slog.String("original_mime", mimeType),
			slog.Int("original_size_bytes", len(audioData)))

		converted, err := p.audioConverter.Convert(audioData, mimeType)
		if err != nil {
			p.log.Warn("failed to convert audio for status, sending original",
				slog.String("error", err.Error()))
			audioDataToUpload = audioData
			finalMimeType = mimeType
			// Try to generate waveform anyway
			waveform = p.audioConverter.GenerateWaveformFromData(audioData, mimeType)
		} else {
			audioDataToUpload = converted.Data
			finalMimeType = converted.MimeType
			duration = converted.Duration
			waveform = converted.Waveform // CRITICAL: 64 samples for visualization

			p.log.Info("audio converted for status successfully",
				slog.Int("converted_size_bytes", len(converted.Data)),
				slog.Int64("duration_seconds", duration),
				slog.Int("waveform_samples", len(waveform)))
		}
	}

	// Upload audio to WhatsApp servers
	uploaded, err := client.Upload(ctx, audioDataToUpload, wameow.MediaAudio)
	if err != nil {
		return fmt.Errorf("upload audio: %w", err)
	}

	// Build audio message with PTT=true for voice note and waveform
	audioMsg := &waProto.AudioMessage{
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      proto.String(finalMimeType),
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uploaded.FileLength),
		PTT:           proto.Bool(true), // CRITICAL: Voice note mode for status
	}

	// Add duration for progress bar
	if duration > 0 {
		audioMsg.Seconds = proto.Uint32(uint32(duration))
	}

	// CRITICAL: Add waveform for visualization (64 samples, 0-100 range)
	if len(waveform) > 0 {
		audioMsg.Waveform = waveform
	}

	msg := &waProto.Message{
		AudioMessage: audioMsg,
	}

	// Send to status@broadcast
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send audio status: %w", err)
	}

	p.log.Info("audio status sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("recipient", recipientJID.String()),
		slog.String("whatsapp_message_id", resp.ID),
		slog.String("mime_type", finalMimeType),
		slog.Int64("duration_seconds", duration),
		slog.Int("waveform_samples", len(waveform)),
		slog.Int64("file_size", int64(uploaded.FileLength)),
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
			MessageType:       "status_audio",
			MediaType:         "audio",
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

// ProcessVideo sends a video status/story to status@broadcast
func (p *StatusProcessor) ProcessVideo(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.VideoContent == nil {
		return fmt.Errorf("video_content is required for video status")
	}

	// CRITICAL: Status messages always go to StatusBroadcastJID
	recipientJID := types.StatusBroadcastJID

	// Download or decode video data
	videoData, mimeType, err := p.mediaDownloader.Download(args.VideoContent.MediaURL)
	if err != nil {
		return fmt.Errorf("download video: %w", err)
	}

	// Validate media type
	if err := p.mediaDownloader.ValidateMediaType(mimeType, MediaTypeVideo); err != nil {
		return fmt.Errorf("invalid media type: %w", err)
	}

	// Generate thumbnail (lazy initialization)
	var thumbnail *ThumbnailResult
	if p.thumbGenerator == nil {
		p.thumbGenerator = NewThumbnailGenerator(client, p.log)
	}
	thumbnail, err = p.thumbGenerator.GenerateAndUploadVideoThumbnail(ctx, videoData)
	if err != nil {
		p.log.Warn("failed to generate video thumbnail for status",
			slog.String("error", err.Error()))
	}

	// Upload video to WhatsApp servers
	uploaded, err := client.Upload(ctx, videoData, wameow.MediaVideo)
	if err != nil {
		return fmt.Errorf("upload video: %w", err)
	}

	// Build video message
	videoMsg := &waProto.VideoMessage{
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      proto.String(mimeType),
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uploaded.FileLength),
	}

	// Add caption if provided
	if args.VideoContent.Caption != nil && *args.VideoContent.Caption != "" {
		videoMsg.Caption = args.VideoContent.Caption
	}

	// Add thumbnail if generated
	if thumbnail != nil {
		videoMsg.JPEGThumbnail = thumbnail.Data
		videoMsg.ThumbnailDirectPath = proto.String(thumbnail.DirectPath)
		videoMsg.ThumbnailSHA256 = thumbnail.FileSha256
		videoMsg.ThumbnailEncSHA256 = thumbnail.FileEncSha256
		if thumbnail.Width > 0 {
			videoMsg.Width = proto.Uint32(thumbnail.Width)
		}
		if thumbnail.Height > 0 {
			videoMsg.Height = proto.Uint32(thumbnail.Height)
		}
	}

	msg := &waProto.Message{
		VideoMessage: videoMsg,
	}

	// Send to status@broadcast
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send video status: %w", err)
	}

	p.log.Info("video status sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("recipient", recipientJID.String()),
		slog.String("whatsapp_message_id", resp.ID),
		slog.Int64("file_size", int64(uploaded.FileLength)),
		slog.Bool("has_thumbnail", thumbnail != nil),
		slog.Bool("has_caption", args.VideoContent.Caption != nil),
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
			MessageType:       "status_video",
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

// parseBackgroundColor parses a color string to ARGB uint32
// Supports formats: "0xFFRRGGBB", "#RRGGBB", "#AARRGGBB"
func (p *StatusProcessor) parseBackgroundColor(color string) uint32 {
	color = strings.TrimSpace(color)

	// Remove "0x" prefix if present
	if strings.HasPrefix(strings.ToLower(color), "0x") {
		color = color[2:]
	}

	// Remove "#" prefix if present
	if strings.HasPrefix(color, "#") {
		color = color[1:]
	}

	// If 6 chars (RRGGBB), add FF for full opacity
	if len(color) == 6 {
		color = "FF" + color
	}

	// Parse as hex
	if len(color) != 8 {
		p.log.Warn("invalid background color format",
			slog.String("color", color))
		return 0
	}

	value, err := strconv.ParseUint(color, 16, 32)
	if err != nil {
		p.log.Warn("failed to parse background color",
			slog.String("color", color),
			slog.String("error", err.Error()))
		return 0
	}

	return uint32(value)
}
