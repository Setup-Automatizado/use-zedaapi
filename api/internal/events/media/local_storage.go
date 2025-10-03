package media

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// LocalMediaStorage manages local filesystem storage for media fallback
type LocalMediaStorage struct {
	cfg           *config.Config
	metrics       *observability.Metrics
	logger        *slog.Logger
	basePath      string
	urlExpiry     time.Duration
	secretKey     []byte
	publicBaseURL string
}

// StoreResult contains the result of storing media locally
type StoreResult struct {
	LocalPath  string
	PublicURL  string
	FileSize   int64
	ExpiresAt  time.Time
	ContentType string
}

// NewLocalMediaStorage creates a new local media storage manager
func NewLocalMediaStorage(
	ctx context.Context,
	cfg *config.Config,
	metrics *observability.Metrics,
) (*LocalMediaStorage, error) {
	logger := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "local_media_storage"),
	)

	// Validate configuration
	basePath := cfg.Media.LocalStoragePath
	if basePath == "" {
		return nil, fmt.Errorf("MEDIA_LOCAL_STORAGE_PATH is required for local media storage")
	}

	secretKey := cfg.Media.LocalSecretKey
	if secretKey == "" {
		return nil, fmt.Errorf("MEDIA_LOCAL_SECRET_KEY is required for URL signing")
	}

	publicBaseURL := cfg.Media.LocalPublicBaseURL
	if publicBaseURL == "" {
		return nil, fmt.Errorf("MEDIA_LOCAL_PUBLIC_BASE_URL is required (e.g., https://api.example.com)")
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	logger.Info("local media storage initialized",
		slog.String("base_path", basePath),
		slog.Duration("url_expiry", cfg.Media.LocalURLExpiry),
		slog.String("public_base_url", publicBaseURL))

	return &LocalMediaStorage{
		cfg:           cfg,
		metrics:       metrics,
		logger:        logger,
		basePath:      basePath,
		urlExpiry:     cfg.Media.LocalURLExpiry,
		secretKey:     []byte(secretKey),
		publicBaseURL: strings.TrimSuffix(publicBaseURL, "/"),
	}, nil
}

// StoreMedia stores media data to local filesystem and generates a signed public URL
func (s *LocalMediaStorage) StoreMedia(
	ctx context.Context,
	instanceID uuid.UUID,
	eventID uuid.UUID,
	mediaType string,
	data []byte,
	contentType string,
) (*StoreResult, error) {
	logger := logging.ContextLogger(ctx, s.logger).With(
		slog.String("instance_id", instanceID.String()),
		slog.String("event_id", eventID.String()),
		slog.String("media_type", mediaType))

	start := time.Now()

	// Generate file path: {instance_id}/{year}/{month}/{day}/{event_id}.{ext}
	now := time.Now()
	extension := s.getExtensionFromContentType(contentType)

	relativePath := filepath.Join(
		instanceID.String(),
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", int(now.Month())),
		fmt.Sprintf("%02d", now.Day()),
		fmt.Sprintf("%s%s", eventID.String(), extension),
	)

	fullPath := filepath.Join(s.basePath, relativePath)

	// Create directory structure
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Error("failed to create directory",
			slog.String("error", err.Error()),
			slog.String("dir", dir))
		s.metrics.MediaFallbackFailure.WithLabelValues(instanceID.String(), mediaType, "mkdir_error").Inc()
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file to disk
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		logger.Error("failed to write file",
			slog.String("error", err.Error()),
			slog.String("path", fullPath))
		s.metrics.MediaFallbackFailure.WithLabelValues(instanceID.String(), mediaType, "write_error").Inc()
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	fileSize := int64(len(data))

	logger.Info("media stored locally",
		slog.String("path", relativePath),
		slog.Int64("size", fileSize),
		slog.Duration("duration", time.Since(start)))

	// Generate signed public URL
	expiresAt := time.Now().Add(s.urlExpiry)
	publicURL, err := s.GenerateSignedURL(relativePath, expiresAt)
	if err != nil {
		logger.Error("failed to generate signed URL",
			slog.String("error", err.Error()))
		// Não faz rollback do arquivo - pode ser útil para debug
		return nil, fmt.Errorf("failed to generate signed URL: %w", err)
	}

	// Update metrics
	s.metrics.MediaFallbackSuccess.WithLabelValues(instanceID.String(), mediaType, "local").Inc()
	s.metrics.MediaLocalStorageSize.Add(float64(fileSize))
	s.metrics.MediaLocalStorageFiles.Inc()

	return &StoreResult{
		LocalPath:   relativePath,
		PublicURL:   publicURL,
		FileSize:    fileSize,
		ExpiresAt:   expiresAt,
		ContentType: contentType,
	}, nil
}

// GenerateSignedURL generates a signed URL for serving media
func (s *LocalMediaStorage) GenerateSignedURL(relativePath string, expiresAt time.Time) (string, error) {
	// URL format: https://api.example.com/v1/media/{path}?expires={timestamp}&signature={hmac}
	expiresTimestamp := expiresAt.Unix()

	// Create signature: HMAC-SHA256(path + expires, secretKey)
	message := fmt.Sprintf("%s:%d", relativePath, expiresTimestamp)
	signature := s.generateHMAC(message)

	// Build URL
	url := fmt.Sprintf("%s/v1/media/%s?expires=%d&signature=%s",
		s.publicBaseURL,
		relativePath,
		expiresTimestamp,
		signature)

	return url, nil
}

// ValidateSignedURL validates a signed URL and returns the relative path if valid
func (s *LocalMediaStorage) ValidateSignedURL(relativePath string, expiresStr, signature string) error {
	// Parse expiration timestamp
	expiresTimestamp, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid expires parameter: %w", err)
	}

	expiresAt := time.Unix(expiresTimestamp, 0)

	// Check if URL expired
	if time.Now().After(expiresAt) {
		return fmt.Errorf("URL expired at %s", expiresAt.Format(time.RFC3339))
	}

	// Validate signature
	expectedMessage := fmt.Sprintf("%s:%d", relativePath, expiresTimestamp)
	expectedSignature := s.generateHMAC(expectedMessage)

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// ServeMedia serves media file with validation
func (s *LocalMediaStorage) ServeMedia(
	ctx context.Context,
	relativePath string,
	expiresStr, signature string,
) ([]byte, string, error) {
	logger := logging.ContextLogger(ctx, s.logger).With(
		slog.String("path", relativePath))

	// Validate URL signature
	if err := s.ValidateSignedURL(relativePath, expiresStr, signature); err != nil {
		logger.Warn("invalid signed URL",
			slog.String("error", err.Error()))
		return nil, "", fmt.Errorf("validation failed: %w", err)
	}

	// Prevent path traversal
	cleanPath := filepath.Clean(relativePath)
	if strings.Contains(cleanPath, "..") {
		logger.Warn("path traversal attempt detected",
			slog.String("clean_path", cleanPath))
		return nil, "", fmt.Errorf("invalid path: path traversal detected")
	}

	// Build full path
	fullPath := filepath.Join(s.basePath, cleanPath)

	// Check if file exists
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("file not found",
				slog.String("full_path", fullPath))
			return nil, "", fmt.Errorf("file not found")
		}
		return nil, "", fmt.Errorf("failed to stat file: %w", err)
	}

	// Check if it's a file (not directory)
	if fileInfo.IsDir() {
		logger.Warn("attempted to serve directory",
			slog.String("full_path", fullPath))
		return nil, "", fmt.Errorf("path is a directory")
	}

	// Read file
	data, err := os.ReadFile(fullPath)
	if err != nil {
		logger.Error("failed to read file",
			slog.String("error", err.Error()),
			slog.String("full_path", fullPath))
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	// Detect content type from file extension
	contentType := mime.TypeByExtension(filepath.Ext(cleanPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	logger.Info("media served successfully",
		slog.Int("size", len(data)),
		slog.String("content_type", contentType))

	return data, contentType, nil
}

// CleanupExpired removes expired media files
func (s *LocalMediaStorage) CleanupExpired(ctx context.Context) (int, error) {
	logger := logging.ContextLogger(ctx, s.logger)

	start := time.Now()
	cutoffTime := time.Now().Add(-s.urlExpiry)

	var filesRemoved int
	var bytesFreed int64

	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is older than expiry
		if info.ModTime().Before(cutoffTime) {
			logger.Debug("removing expired file",
				slog.String("path", path),
				slog.Time("modified", info.ModTime()))

			if err := os.Remove(path); err != nil {
				logger.Error("failed to remove file",
					slog.String("error", err.Error()),
					slog.String("path", path))
				return nil // Continue with other files
			}

			filesRemoved++
			bytesFreed += info.Size()

			// Update metrics
			s.metrics.MediaLocalStorageSize.Sub(float64(info.Size()))
			s.metrics.MediaLocalStorageFiles.Dec()
		}

		return nil
	})

	if err != nil {
		logger.Error("cleanup walk failed",
			slog.String("error", err.Error()))
		return filesRemoved, fmt.Errorf("cleanup walk failed: %w", err)
	}

	// Remove empty directories
	_ = s.removeEmptyDirs(s.basePath)

	logger.Info("cleanup completed",
		slog.Int("files_removed", filesRemoved),
		slog.Int64("bytes_freed", bytesFreed),
		slog.Duration("duration", time.Since(start)))

	// Update cleanup metrics
	s.metrics.MediaCleanupTotal.WithLabelValues("expired_files").Add(float64(filesRemoved))

	return filesRemoved, nil
}

// GetStats returns storage statistics
func (s *LocalMediaStorage) GetStats(ctx context.Context) (map[string]interface{}, error) {
	var totalFiles int
	var totalBytes int64

	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalFiles++
			totalBytes += info.Size()
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return map[string]interface{}{
		"total_files":      totalFiles,
		"total_bytes":      totalBytes,
		"base_path":        s.basePath,
		"url_expiry":       s.urlExpiry.String(),
		"public_base_url":  s.publicBaseURL,
	}, nil
}

// generateHMAC generates HMAC-SHA256 signature
func (s *LocalMediaStorage) generateHMAC(message string) string {
	h := hmac.New(sha256.New, s.secretKey)
	h.Write([]byte(message))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// getExtensionFromContentType returns file extension based on content type
func (s *LocalMediaStorage) getExtensionFromContentType(contentType string) string {
	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil || len(extensions) == 0 {
		// Fallback to common extensions
		switch contentType {
		case "image/jpeg":
			return ".jpg"
		case "image/png":
			return ".png"
		case "image/webp":
			return ".webp"
		case "video/mp4":
			return ".mp4"
		case "video/mpeg":
			return ".mpeg"
		case "audio/mpeg":
			return ".mp3"
		case "audio/ogg":
			return ".ogg"
		case "audio/opus":
			return ".opus"
		case "application/pdf":
			return ".pdf"
		default:
			return ".bin"
		}
	}
	return extensions[0]
}

// removeEmptyDirs recursively removes empty directories
func (s *LocalMediaStorage) removeEmptyDirs(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() || path == root {
			return nil
		}

		// Try to remove directory (will fail if not empty)
		_ = os.Remove(path)

		return nil
	})
}

// CopyToWriter copies media file to io.Writer (for HTTP response)
func (s *LocalMediaStorage) CopyToWriter(
	ctx context.Context,
	relativePath string,
	expiresStr, signature string,
	w io.Writer,
) (string, int64, error) {
	logger := logging.ContextLogger(ctx, s.logger).With(
		slog.String("path", relativePath))

	// Validate URL signature
	if err := s.ValidateSignedURL(relativePath, expiresStr, signature); err != nil {
		return "", 0, fmt.Errorf("validation failed: %w", err)
	}

	// Prevent path traversal
	cleanPath := filepath.Clean(relativePath)
	if strings.Contains(cleanPath, "..") {
		return "", 0, fmt.Errorf("invalid path: path traversal detected")
	}

	fullPath := filepath.Join(s.basePath, cleanPath)

	// Open file
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", 0, fmt.Errorf("file not found")
		}
		return "", 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return "", 0, fmt.Errorf("failed to stat file: %w", err)
	}

	// Detect content type
	contentType := mime.TypeByExtension(filepath.Ext(cleanPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Copy to writer
	written, err := io.Copy(w, file)
	if err != nil {
		logger.Error("failed to copy file to writer",
			slog.String("error", err.Error()))
		return "", 0, fmt.Errorf("failed to copy file: %w", err)
	}

	logger.Info("media copied to writer",
		slog.Int64("bytes_written", written),
		slog.String("content_type", contentType))

	return contentType, fileInfo.Size(), nil
}
