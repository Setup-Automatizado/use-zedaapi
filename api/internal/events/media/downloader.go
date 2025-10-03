package media

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"

	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// MediaDownloader handles downloading encrypted media from WhatsApp with generic type support
type MediaDownloader struct {
	metrics *observability.Metrics
	logger  *slog.Logger
	timeout time.Duration
}

// DownloadResult contains the downloaded media data and metadata
type DownloadResult struct {
	Data        io.Reader
	ContentType string
	FileName    string
	FileSize    int64
	SHA256      string
	MediaType   string
}

// NewMediaDownloader creates a new media downloader with complete media type support
func NewMediaDownloader(metrics *observability.Metrics, timeout time.Duration) *MediaDownloader {
	logger := slog.Default().With(
		slog.String("component", "media_downloader"),
	)

	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	logger.Info("media downloader initialized",
		slog.Duration("timeout", timeout),
		slog.String("features", "generic_type_support,mime_stdlib,reflection_extraction"))

	return &MediaDownloader{
		metrics: metrics,
		logger:  logger,
		timeout: timeout,
	}
}

// Download downloads and decrypts media from WhatsApp using generic type detection
func (d *MediaDownloader) Download(ctx context.Context, client *whatsmeow.Client, instanceID uuid.UUID, eventID uuid.UUID, msg proto.Message) (*DownloadResult, error) {
	logger := logging.ContextLogger(ctx, d.logger).With(
		slog.String("instance_id", instanceID.String()),
		slog.String("event_id", eventID.String()))

	// Create context with timeout
	downloadCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	start := time.Now()

	// Extract downloadable message and metadata using generic detection
	downloadable, mediaType, contentType, fileName, err := d.extractMediaInfoGeneric(msg)
	if err != nil {
		logger.Error("failed to extract media info",
			slog.String("error", err.Error()))
		d.metrics.MediaDownloadsTotal.WithLabelValues(instanceID.String(), "unknown", "failure").Inc()
		d.metrics.MediaDownloadErrors.WithLabelValues("invalid_message").Inc()
		return nil, fmt.Errorf("invalid media message: %w", err)
	}

	logger.Debug("downloading media",
		slog.String("media_type", mediaType),
		slog.String("content_type", contentType),
		slog.String("file_name", fileName))

	// Download and decrypt media using whatsmeow client
	data, err := client.Download(downloadCtx, downloadable)
	if err != nil {
		duration := time.Since(start)
		logger.Error("whatsapp media download failed",
			slog.String("error", err.Error()),
			slog.Duration("duration", duration))

		d.metrics.MediaDownloadsTotal.WithLabelValues(instanceID.String(), mediaType, "failure").Inc()
		d.metrics.MediaDownloadErrors.WithLabelValues(classifyDownloadError(err)).Inc()

		return nil, fmt.Errorf("whatsapp download failed: %w", err)
	}

	duration := time.Since(start)
	fileSize := int64(len(data))

	logger.Info("media download succeeded",
		slog.Int64("file_size", fileSize),
		slog.Duration("duration", duration))

	// Update metrics
	d.metrics.MediaDownloadsTotal.WithLabelValues(instanceID.String(), mediaType, "success").Inc()
	d.metrics.MediaDownloadDuration.WithLabelValues(instanceID.String(), mediaType).Observe(duration.Seconds())
	d.metrics.MediaDownloadSize.WithLabelValues(instanceID.String(), mediaType).Observe(float64(fileSize))

	// Get SHA256 from message (already validated by whatsmeow)
	sha256Hash := ""
	if fileSHA := downloadable.GetFileSHA256(); len(fileSHA) > 0 {
		sha256Hash = fmt.Sprintf("%x", fileSHA)
	}

	return &DownloadResult{
		Data:        bytes.NewReader(data),
		ContentType: contentType,
		FileName:    fileName,
		FileSize:    fileSize,
		SHA256:      sha256Hash,
		MediaType:   mediaType,
	}, nil
}

// extractMediaInfoGeneric extracts media info using whatsmeow's type system + reflection
func (d *MediaDownloader) extractMediaInfoGeneric(msg proto.Message) (whatsmeow.DownloadableMessage, string, string, string, error) {
	// Check if message implements DownloadableMessage interface
	downloadable, ok := msg.(whatsmeow.DownloadableMessage)
	if !ok {
		return nil, "", "", "", fmt.Errorf("message does not implement DownloadableMessage interface: %T", msg)
	}

	// Use whatsmeow's GetMediaType() to determine WhatsApp media type
	waMediaType := whatsmeow.GetMediaType(downloadable)
	if waMediaType == "" {
		return nil, "", "", "", fmt.Errorf("unknown WhatsApp media type for message: %T", msg)
	}

	// Map WhatsApp media type to our persistence media type
	persistenceType := mapWhatsAppMediaType(waMediaType)

	// Extract content type using reflection (works for all message types)
	contentType := extractContentTypeReflection(downloadable, waMediaType)

	// Extract filename using reflection
	fileName := extractFileNameReflection(downloadable, contentType)

	return downloadable, string(persistenceType), contentType, fileName, nil
}

// mapWhatsAppMediaType maps WhatsApp MediaType to our persistence.MediaType
func mapWhatsAppMediaType(waType whatsmeow.MediaType) persistence.MediaType {
	switch waType {
	case whatsmeow.MediaImage, whatsmeow.MediaLinkThumbnail:
		return persistence.MediaTypeImage
	case whatsmeow.MediaVideo:
		return persistence.MediaTypeVideo
	case whatsmeow.MediaAudio:
		return persistence.MediaTypeAudio
	case whatsmeow.MediaDocument, whatsmeow.MediaHistory, whatsmeow.MediaAppState, whatsmeow.MediaStickerPack:
		return persistence.MediaTypeDocument
	default:
		// Safe fallback for any future WhatsApp media types
		return persistence.MediaTypeDocument
	}
}

// extractContentTypeReflection extracts MIME type using reflection
func extractContentTypeReflection(msg whatsmeow.DownloadableMessage, waType whatsmeow.MediaType) string {
	// Try to get Mimetype field using reflection
	v := reflect.ValueOf(msg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Most WhatsApp messages have "Mimetype" field
	mimeField := v.FieldByName("Mimetype")
	if mimeField.IsValid() && !mimeField.IsNil() {
		if mimePtr, ok := mimeField.Interface().(*string); ok && mimePtr != nil {
			return *mimePtr
		}
	}

	// Fallback to default content type based on WhatsApp media type
	return getDefaultContentType(waType)
}

// extractFileNameReflection extracts filename using reflection
func extractFileNameReflection(msg whatsmeow.DownloadableMessage, contentType string) string {
	v := reflect.ValueOf(msg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Try "FileName" field (DocumentMessage)
	fileNameField := v.FieldByName("FileName")
	if fileNameField.IsValid() && !fileNameField.IsNil() {
		if namePtr, ok := fileNameField.Interface().(*string); ok && namePtr != nil && *namePtr != "" {
			return *namePtr
		}
	}

	// Try "Caption" field (may contain filename hint)
	captionField := v.FieldByName("Caption")
	if captionField.IsValid() && !captionField.IsNil() {
		if captionPtr, ok := captionField.Interface().(*string); ok && captionPtr != nil && *captionPtr != "" {
			return generateFileNameFromCaption(*captionPtr, contentType)
		}
	}

	// Fallback: generate generic filename from content type
	return generateFileName("media", contentType)
}

// getDefaultContentType returns default MIME type for WhatsApp media types
func getDefaultContentType(waType whatsmeow.MediaType) string {
	switch waType {
	case whatsmeow.MediaImage, whatsmeow.MediaLinkThumbnail:
		return "image/jpeg"
	case whatsmeow.MediaVideo:
		return "video/mp4"
	case whatsmeow.MediaAudio:
		return "audio/ogg; codecs=opus"
	case whatsmeow.MediaDocument:
		return "application/octet-stream"
	case whatsmeow.MediaHistory:
		return "application/x-protobuf"
	case whatsmeow.MediaAppState:
		return "application/x-protobuf"
	case whatsmeow.MediaStickerPack:
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

// generateFileName creates filename from base name and MIME type
func generateFileName(baseName string, mimeType string) string {
	if baseName == "" {
		baseName = "media"
	}

	ext := getExtensionFromMIME(mimeType)
	if ext != "" {
		return fmt.Sprintf("%s.%s", baseName, ext)
	}

	return baseName
}

// generateFileNameFromCaption creates filename from caption
func generateFileNameFromCaption(caption string, mimeType string) string {
	// Clean caption for use as filename
	name := strings.TrimSpace(caption)
	if len(name) > 100 {
		name = name[:100]
	}

	// Remove problematic characters
	name = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`<>:"/\|?*`, r) {
			return '_'
		}
		return r
	}, name)

	if name == "" {
		return generateFileName("media", mimeType)
	}

	return generateFileName(name, mimeType)
}

// getExtensionFromMIME extracts file extension using stdlib + fallback
func getExtensionFromMIME(mimeType string) string {
	// Use standard library mime.ExtensionsByType() - supports 100+ types
	exts, err := mime.ExtensionsByType(mimeType)
	if err == nil && len(exts) > 0 {
		return strings.TrimPrefix(exts[0], ".") // Remove leading dot
	}

	// Fallback for special WhatsApp types
	fallbackMap := map[string]string{
		"audio/ogg; codecs=opus":   "ogg",
		"image/webp":               "webp",
		"application/x-protobuf":   "pb",
		"application/octet-stream": "bin",
	}

	if ext, ok := fallbackMap[strings.ToLower(mimeType)]; ok {
		return ext
	}

	// Try extracting from MIME type (e.g., "image/jpeg" -> "jpeg")
	parts := strings.Split(mimeType, "/")
	if len(parts) == 2 {
		ext := strings.TrimSpace(parts[1])
		// Remove codec info if present
		if idx := strings.Index(ext, ";"); idx != -1 {
			ext = ext[:idx]
		}
		if ext != "" && ext != "plain" {
			return ext
		}
	}

	return "bin" // Final fallback
}

// classifyDownloadError categorizes download errors for metrics
func classifyDownloadError(err error) string {
	errStr := err.Error()

	switch {
	case errStr == "client is not logged in":
		return "not_logged_in"
	case errStr == "failed to refresh media connections":
		return "media_conn_refresh_failed"
	case errStr == "no url present":
		return "no_url"
	case errStr == "nothing downloadable found":
		return "not_downloadable"
	case errStr == "file too short":
		return "file_too_short"
	case errStr == "invalid media hmac":
		return "invalid_hmac"
	case errStr == "invalid media enc sha256":
		return "invalid_enc_hash"
	case errStr == "invalid media sha256":
		return "invalid_hash"
	case errStr == "file length mismatch":
		return "length_mismatch"
	case errStr == "media download failed with status 403":
		return "http_403"
	case errStr == "media download failed with status 404":
		return "http_404"
	case errStr == "media download failed with status 410":
		return "http_410"
	case err == context.DeadlineExceeded:
		return "timeout"
	case err == context.Canceled:
		return "canceled"
	default:
		return "unknown"
	}
}
