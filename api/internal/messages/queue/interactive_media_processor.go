package queue

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	wameow "go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

// InteractiveMediaProcessor handles media processing for interactive messages
// (button list, button actions, carousel) with full thumbnail and upload support
type InteractiveMediaProcessor struct {
	log             *slog.Logger
	mediaDownloader *MediaDownloader
	thumbGenerator  *ThumbnailGenerator
	client          *wameow.Client
}

// ProcessedMediaHeader contains the uploaded media information ready for proto
type ProcessedMediaHeader struct {
	// Media type
	MediaType MediaType // image or video

	// Original media data (for backup/debugging)
	MimeType string

	// Uploaded media fields
	URL           string
	DirectPath    string
	MediaKey      []byte
	FileEncSHA256 []byte
	FileSHA256    []byte
	FileLength    uint64

	// Dimensions
	Width  uint32
	Height uint32

	// Thumbnail data (JPEG bytes for JPEGThumbnail field)
	JPEGThumbnail       []byte
	ThumbnailDirectPath string
	ThumbnailSHA256     []byte
	ThumbnailEncSHA256  []byte
	ThumbnailWidth      uint32
	ThumbnailHeight     uint32

	// For documents
	FileName  string
	PageCount uint32
}

// NewInteractiveMediaProcessor creates a new interactive media processor
func NewInteractiveMediaProcessor(client *wameow.Client, log *slog.Logger) *InteractiveMediaProcessor {
	return &InteractiveMediaProcessor{
		log:             log.With(slog.String("component", "interactive_media")),
		mediaDownloader: NewMediaDownloader(100), // 100MB max
		thumbGenerator:  NewThumbnailGenerator(client, log),
		client:          client,
	}
}

// ProcessMediaURL downloads, validates, generates thumbnail, and uploads media
// Returns ProcessedMediaHeader ready to be used in proto construction
func (p *InteractiveMediaProcessor) ProcessMediaURL(ctx context.Context, mediaURL string) (*ProcessedMediaHeader, error) {
	if mediaURL == "" {
		return nil, fmt.Errorf("media URL is empty")
	}

	p.log.Debug("processing media for interactive message",
		slog.String("url_prefix", truncateURL(mediaURL, 50)))

	// Download media
	mediaData, mimeType, err := p.mediaDownloader.Download(mediaURL)
	if err != nil {
		return nil, fmt.Errorf("download media: %w", err)
	}

	p.log.Debug("downloaded media",
		slog.Int("size_bytes", len(mediaData)),
		slog.String("mime_type", mimeType))

	// Determine media type
	var mediaType MediaType
	var waMediaType wameow.MediaType

	if strings.HasPrefix(mimeType, "image/") {
		mediaType = MediaTypeImage
		waMediaType = wameow.MediaImage
	} else if strings.HasPrefix(mimeType, "video/") {
		mediaType = MediaTypeVideo
		waMediaType = wameow.MediaVideo
	} else if strings.HasPrefix(mimeType, "application/pdf") || strings.HasPrefix(mimeType, "application/") {
		mediaType = MediaTypeDocument
		waMediaType = wameow.MediaDocument
	} else {
		return nil, fmt.Errorf("unsupported media type for interactive messages: %s", mimeType)
	}

	// Get dimensions
	width, height := 0, 0
	if mediaType == MediaTypeImage || mediaType == MediaTypeVideo {
		w, h, err := GetMediaDimensions(mediaData, mimeType)
		if err != nil {
			p.log.Warn("failed to detect dimensions",
				slog.String("error", err.Error()),
				slog.String("mime_type", mimeType))
		} else {
			width, height = w, h
		}
	}

	// Generate thumbnail
	var thumbnail *ThumbnailResult
	switch mediaType {
	case MediaTypeImage:
		thumbnail, err = p.thumbGenerator.GenerateAndUploadImageThumbnail(ctx, mediaData, mimeType)
	case MediaTypeVideo:
		thumbnail, err = p.thumbGenerator.GenerateAndUploadVideoThumbnail(ctx, mediaData)
	case MediaTypeDocument:
		thumbnail, err = p.thumbGenerator.GenerateAndUploadDocumentThumbnail(ctx, mediaData, mimeType)
	}

	if err != nil {
		p.log.Warn("failed to generate thumbnail, continuing without it",
			slog.String("error", err.Error()),
			slog.String("media_type", string(mediaType)))
	}

	// Upload media to WhatsApp servers
	uploaded, err := p.client.Upload(ctx, mediaData, waMediaType)
	if err != nil {
		return nil, fmt.Errorf("upload media to WhatsApp: %w", err)
	}

	p.log.Info("processed and uploaded media for interactive message",
		slog.String("media_type", string(mediaType)),
		slog.String("mime_type", mimeType),
		slog.Int("width", width),
		slog.Int("height", height),
		slog.Uint64("file_length", uploaded.FileLength),
		slog.Bool("has_thumbnail", thumbnail != nil))

	result := &ProcessedMediaHeader{
		MediaType:     mediaType,
		MimeType:      mimeType,
		URL:           uploaded.URL,
		DirectPath:    uploaded.DirectPath,
		MediaKey:      uploaded.MediaKey,
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    uploaded.FileLength,
		Width:         uint32(width),
		Height:        uint32(height),
	}

	// Add thumbnail if generated
	if thumbnail != nil {
		result.JPEGThumbnail = thumbnail.Data
		result.ThumbnailDirectPath = thumbnail.DirectPath
		result.ThumbnailSHA256 = thumbnail.FileSha256
		result.ThumbnailEncSHA256 = thumbnail.FileEncSha256
		result.ThumbnailWidth = thumbnail.Width
		result.ThumbnailHeight = thumbnail.Height
	}

	return result, nil
}

// BuildImageMessageProto creates a fully populated ImageMessage proto from processed media
func (p *InteractiveMediaProcessor) BuildImageMessageProto(media *ProcessedMediaHeader) *waProto.ImageMessage {
	if media == nil || media.MediaType != MediaTypeImage {
		return nil
	}

	imageMsg := &waProto.ImageMessage{
		URL:           proto.String(media.URL),
		DirectPath:    proto.String(media.DirectPath),
		MediaKey:      media.MediaKey,
		Mimetype:      proto.String(media.MimeType),
		FileEncSHA256: media.FileEncSHA256,
		FileSHA256:    media.FileSHA256,
		FileLength:    proto.Uint64(media.FileLength),
	}

	if media.Width > 0 && media.Height > 0 {
		imageMsg.Width = proto.Uint32(media.Width)
		imageMsg.Height = proto.Uint32(media.Height)
	}

	if len(media.JPEGThumbnail) > 0 {
		imageMsg.JPEGThumbnail = media.JPEGThumbnail
		if media.ThumbnailDirectPath != "" {
			imageMsg.ThumbnailDirectPath = proto.String(media.ThumbnailDirectPath)
		}
		if len(media.ThumbnailSHA256) > 0 {
			imageMsg.ThumbnailSHA256 = media.ThumbnailSHA256
		}
		if len(media.ThumbnailEncSHA256) > 0 {
			imageMsg.ThumbnailEncSHA256 = media.ThumbnailEncSHA256
		}
	}

	return imageMsg
}

// BuildVideoMessageProto creates a fully populated VideoMessage proto from processed media
func (p *InteractiveMediaProcessor) BuildVideoMessageProto(media *ProcessedMediaHeader) *waProto.VideoMessage {
	if media == nil || media.MediaType != MediaTypeVideo {
		return nil
	}

	videoMsg := &waProto.VideoMessage{
		URL:           proto.String(media.URL),
		DirectPath:    proto.String(media.DirectPath),
		MediaKey:      media.MediaKey,
		Mimetype:      proto.String(media.MimeType),
		FileEncSHA256: media.FileEncSHA256,
		FileSHA256:    media.FileSHA256,
		FileLength:    proto.Uint64(media.FileLength),
	}

	if media.Width > 0 && media.Height > 0 {
		videoMsg.Width = proto.Uint32(media.Width)
		videoMsg.Height = proto.Uint32(media.Height)
	}

	if len(media.JPEGThumbnail) > 0 {
		videoMsg.JPEGThumbnail = media.JPEGThumbnail
		if media.ThumbnailDirectPath != "" {
			videoMsg.ThumbnailDirectPath = proto.String(media.ThumbnailDirectPath)
		}
		if len(media.ThumbnailSHA256) > 0 {
			videoMsg.ThumbnailSHA256 = media.ThumbnailSHA256
		}
		if len(media.ThumbnailEncSHA256) > 0 {
			videoMsg.ThumbnailEncSHA256 = media.ThumbnailEncSHA256
		}
	}

	return videoMsg
}

// BuildDocumentMessageProto creates a fully populated DocumentMessage proto from processed media
func (p *InteractiveMediaProcessor) BuildDocumentMessageProto(media *ProcessedMediaHeader, fileName string) *waProto.DocumentMessage {
	if media == nil || media.MediaType != MediaTypeDocument {
		return nil
	}

	docMsg := &waProto.DocumentMessage{
		URL:           proto.String(media.URL),
		DirectPath:    proto.String(media.DirectPath),
		MediaKey:      media.MediaKey,
		Mimetype:      proto.String(media.MimeType),
		FileEncSHA256: media.FileEncSHA256,
		FileSHA256:    media.FileSHA256,
		FileLength:    proto.Uint64(media.FileLength),
	}

	if fileName != "" {
		docMsg.FileName = proto.String(fileName)
	}

	if len(media.JPEGThumbnail) > 0 {
		docMsg.JPEGThumbnail = media.JPEGThumbnail
		if media.ThumbnailDirectPath != "" {
			docMsg.ThumbnailDirectPath = proto.String(media.ThumbnailDirectPath)
		}
		if len(media.ThumbnailSHA256) > 0 {
			docMsg.ThumbnailSHA256 = media.ThumbnailSHA256
		}
		if len(media.ThumbnailEncSHA256) > 0 {
			docMsg.ThumbnailEncSHA256 = media.ThumbnailEncSHA256
		}
	}

	if media.PageCount > 0 {
		docMsg.PageCount = proto.Uint32(media.PageCount)
	}

	return docMsg
}

// ProcessedMediaToHeader creates an InteractiveMessage_Header from processed media
// This is the key function that converts processed media into proto header format
func (p *InteractiveMediaProcessor) ProcessedMediaToHeader(media *ProcessedMediaHeader, title string) *waProto.InteractiveMessage_Header {
	if media == nil {
		// Return text-only header if no media
		if title != "" {
			return &waProto.InteractiveMessage_Header{
				Title:              proto.String(title),
				HasMediaAttachment: proto.Bool(false),
			}
		}
		return nil
	}

	header := &waProto.InteractiveMessage_Header{
		HasMediaAttachment: proto.Bool(true),
	}

	if title != "" {
		header.Title = proto.String(title)
	}

	// Set appropriate media type
	switch media.MediaType {
	case MediaTypeImage:
		header.Media = &waProto.InteractiveMessage_Header_ImageMessage{
			ImageMessage: p.BuildImageMessageProto(media),
		}
	case MediaTypeVideo:
		header.Media = &waProto.InteractiveMessage_Header_VideoMessage{
			VideoMessage: p.BuildVideoMessageProto(media),
		}
	case MediaTypeDocument:
		header.Media = &waProto.InteractiveMessage_Header_DocumentMessage{
			DocumentMessage: p.BuildDocumentMessageProto(media, ""),
		}
	default:
		// If we only have thumbnail bytes but no full media, use JpegThumbnail
		if len(media.JPEGThumbnail) > 0 {
			header.Media = &waProto.InteractiveMessage_Header_JpegThumbnail{
				JPEGThumbnail: media.JPEGThumbnail,
			}
		}
	}

	return header
}

// truncateURL truncates a URL for logging (avoiding huge base64 strings in logs)
func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	if strings.HasPrefix(url, "data:") {
		// For base64 data URIs, show type and truncate
		if idx := strings.Index(url, ","); idx > 0 && idx < 50 {
			return url[:idx] + ",<base64 data truncated>"
		}
		return url[:maxLen] + "..."
	}
	return url[:maxLen] + "..."
}
