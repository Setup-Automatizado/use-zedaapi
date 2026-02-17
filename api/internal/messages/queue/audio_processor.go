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

// AudioProcessor handles audio message sending via WhatsApp
// Automatically converts non-Opus audio formats to WhatsApp native Opus/OGG format
// for optimal voice note experience with waveform display
type AudioProcessor struct {
	log             *slog.Logger
	mediaDownloader *MediaDownloader
	audioConverter  *AudioConverter
	presenceHelper  *PresenceHelper
	echoEmitter     *echo.Emitter
}

// NewAudioProcessor creates a new audio message processor
func NewAudioProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *AudioProcessor {
	return &AudioProcessor{
		log:             log.With(slog.String("processor", "audio")),
		mediaDownloader: NewMediaDownloader(100), // 100MB max for audio
		audioConverter:  NewAudioConverter(log),
		presenceHelper:  NewPresenceHelper(),
		echoEmitter:     echoEmitter,
	}
}

// Process sends an audio message via WhatsApp
// Supports all audio formats - automatically converts to Opus/OGG for native voice notes
func (p *AudioProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.AudioContent == nil {
		return fmt.Errorf("audio_content is required for audio messages")
	}

	// Parse recipient JID
	recipientJID, err := types.ParseJID(args.Phone)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Simulate recording indicator if DelayTyping is set
	// For audio messages, DelayTyping represents "recording audio" duration
	if args.DelayTyping > 0 {
		if err := p.presenceHelper.SimulateRecording(client, recipientJID, args.DelayTyping); err != nil {
			p.log.Warn("failed to send recording indicator",
				slog.String("error", err.Error()),
				slog.String("phone", args.Phone))
		}
	}

	// Download or decode audio data using helper
	audioData, mimeType, err := p.mediaDownloader.Download(args.AudioContent.MediaURL)
	if err != nil {
		return fmt.Errorf("download audio: %w", err)
	}

	// Validate media type
	if err := p.mediaDownloader.ValidateMediaType(mimeType, MediaTypeAudio); err != nil {
		return fmt.Errorf("invalid media type: %w", err)
	}

	// Convert audio to WhatsApp native Opus/OGG format if needed
	// This ensures audio appears as voice note with waveform display
	var audioDataToUpload []byte
	var finalMimeType string
	var duration int64
	var waveform []byte

	if p.audioConverter.IsOpusFormat(mimeType) {
		// Already in Opus/OGG format - still need to generate waveform and get duration
		audioDataToUpload = audioData
		finalMimeType = "audio/ogg; codecs=opus"
		// Generate waveform for existing Opus audio
		waveform = p.audioConverter.GenerateWaveformFromData(audioData, mimeType)
		duration = p.audioConverter.GetDurationFromData(audioData, mimeType)
		p.log.Debug("audio already in native Opus/OGG format",
			slog.String("mime_type", mimeType),
			slog.Int("waveform_samples", len(waveform)),
			slog.Int64("duration_seconds", duration))
	} else {
		// Convert to Opus/OGG for native voice note experience
		p.log.Info("converting audio to native Opus/OGG format",
			slog.String("original_mime", mimeType),
			slog.Int("original_size_bytes", len(audioData)))

		converted, err := p.audioConverter.Convert(audioData, mimeType)
		if err != nil {
			// Conversion failed, send original audio as regular audio file
			p.log.Warn("failed to convert audio to Opus/OGG, sending as regular audio file",
				slog.String("error", err.Error()),
				slog.String("original_mime", mimeType))
			audioDataToUpload = audioData
			finalMimeType = mimeType
			// Try to generate waveform anyway
			waveform = p.audioConverter.GenerateWaveformFromData(audioData, mimeType)
		} else {
			// Conversion successful
			audioDataToUpload = converted.Data
			finalMimeType = converted.MimeType
			duration = converted.Duration
			waveform = converted.Waveform

			p.log.Info("audio converted to Opus/OGG successfully",
				slog.String("original_mime", mimeType),
				slog.Int("original_size_bytes", len(audioData)),
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

	// Build ContextInfo using helper
	contextBuilder := NewContextInfoBuilder(client, recipientJID, args, p.log)
	contextInfo, err := contextBuilder.Build(ctx)
	if err != nil {
		p.log.Warn("failed to build context info, sending without it",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Determine if this is a voice note (PTT - Push To Talk)
	// All Opus/OGG audio should be sent as PTT (voice note)
	isPTT := p.audioConverter.IsOpusFormat(finalMimeType)

	// Build audio message with waveform for voice note visualization
	msg := p.buildMessage(args, uploaded, finalMimeType, isPTT, duration, waveform, contextInfo)

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg, BuildSendExtra(args))
	if err != nil {
		return fmt.Errorf("send audio message: %w", err)
	}

	p.log.Info("audio message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.String("mime_type", finalMimeType),
		slog.Bool("is_voice_note", isPTT),
		slog.Bool("was_converted", mimeType != finalMimeType),
		slog.Int64("file_size", int64(uploaded.FileLength)),
		slog.Int64("duration_seconds", duration),
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
			MessageType:       "audio",
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

// buildMessage constructs the audio message proto
func (p *AudioProcessor) buildMessage(
	args *SendMessageArgs,
	uploaded wameow.UploadResponse,
	mimeType string,
	isPTT bool,
	duration int64,
	waveform []byte,
	contextInfo *waProto.ContextInfo,
) *waProto.Message {
	audioMsg := &waProto.AudioMessage{
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      proto.String(mimeType),
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uploaded.FileLength),
		PTT:           proto.Bool(isPTT), // Push To Talk (voice note with waveform)
		ContextInfo:   contextInfo,
	}

	// Add duration if available (helps with voice note display)
	if duration > 0 {
		audioMsg.Seconds = proto.Uint32(uint32(duration))
	}

	// Add waveform data for voice note visualization (the wave bars in WhatsApp)
	if len(waveform) > 0 {
		audioMsg.Waveform = waveform
	}

	return &waProto.Message{
		AudioMessage: audioMsg,
	}
}
