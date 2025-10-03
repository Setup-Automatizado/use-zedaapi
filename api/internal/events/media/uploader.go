package media

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// S3Uploader handles media uploads to S3-compatible storage
type S3Uploader struct {
	client    *s3.Client
	uploader  *manager.Uploader
	bucket    string
	urlExpiry time.Duration
	acl       string // Optional ACL (e.g., "public-read"). Empty = use bucket policy
	metrics   *observability.Metrics
	logger    *slog.Logger
}

// NewS3Uploader creates a new S3 uploader with AWS SDK v2
func NewS3Uploader(ctx context.Context, cfg *config.Config, metrics *observability.Metrics) (*S3Uploader, error) {
	logger := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "s3_uploader"),
	)

	// Create AWS config
	awsCfg := aws.Config{
		Region:      cfg.S3.Region,
		Credentials: credentials.NewStaticCredentialsProvider(cfg.S3.AccessKey, cfg.S3.SecretKey, ""),
	}

	// Set custom endpoint for MinIO or other S3-compatible services
	if cfg.S3.Endpoint != "" {
		awsCfg.BaseEndpoint = aws.String(cfg.S3.Endpoint)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true // Required for MinIO
	})

	// Create uploader with custom part size for large files
	uploader := manager.NewUploader(s3Client, func(u *manager.Uploader) {
		u.PartSize = cfg.Events.MediaChunkSize // 5MB default
		u.Concurrency = 100                    // Upload 100 parts concurrently
	})

	logger.Info("s3 uploader initialized",
		slog.String("bucket", cfg.S3.Bucket),
		slog.String("region", cfg.S3.Region),
		slog.Duration("url_expiry", cfg.S3.URLExpiry),
		slog.String("acl", cfg.S3.ACL))

	if cfg.S3.ACL != "" {
		logger.Warn("using S3 ACL (deprecated AWS pattern - prefer bucket policies)",
			slog.String("acl", cfg.S3.ACL))
	}

	return &S3Uploader{
		client:    s3Client,
		uploader:  uploader,
		bucket:    cfg.S3.Bucket,
		urlExpiry: cfg.S3.URLExpiry,
		acl:       cfg.S3.ACL,
		metrics:   metrics,
		logger:    logger,
	}, nil
}

// Upload uploads media to S3 and returns the key and presigned URL
func (u *S3Uploader) Upload(ctx context.Context, instanceID uuid.UUID, eventID uuid.UUID, mediaType string, reader io.Reader, contentType string, fileSize int64) (key string, presignedURL string, err error) {
	logger := logging.ContextLogger(ctx, u.logger).With(
		slog.String("instance_id", instanceID.String()),
		slog.String("event_id", eventID.String()),
		slog.String("media_type", mediaType),
		slog.Int64("file_size", fileSize))

	start := time.Now()

	// Generate S3 key with organized structure: {instance_id}/{year}/{month}/{day}/{event_id}.{ext}
	key = u.generateKey(instanceID, eventID, mediaType, contentType)

	logger.Debug("uploading media to s3", slog.String("s3_key", key))

	// Build upload input
	uploadInput := &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"instance-id": instanceID.String(),
			"event-id":    eventID.String(),
			"media-type":  mediaType,
		},
	}

	// Only set ACL if configured (modern pattern: use bucket policy instead)
	if u.acl != "" {
		uploadInput.ACL = types.ObjectCannedACL(u.acl)
	}

	// Upload to S3
	uploadResult, err := u.uploader.Upload(ctx, uploadInput)

	duration := time.Since(start)

	if err != nil {
		logger.Error("s3 upload failed",
			slog.String("error", err.Error()),
			slog.Duration("duration", duration))

		u.metrics.MediaUploadAttempts.WithLabelValues("failure").Inc()
		u.metrics.MediaUploadErrors.WithLabelValues(classifyS3Error(err)).Inc()

		return "", "", fmt.Errorf("s3 upload failed: %w", err)
	}

	logger.Info("s3 upload succeeded",
		slog.String("location", uploadResult.Location),
		slog.Duration("duration", duration))

	// Update metrics
	u.metrics.MediaUploadAttempts.WithLabelValues("success").Inc()
	u.metrics.MediaUploadDuration.WithLabelValues(mediaType).Observe(duration.Seconds())
	u.metrics.MediaUploadSizeBytes.WithLabelValues(mediaType).Add(float64(fileSize))

	// Generate presigned URL
	presignedURL, err = u.GeneratePresignedURL(ctx, key)
	if err != nil {
		return key, "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return key, presignedURL, nil
}

// GeneratePresignedURL creates a time-limited presigned URL for an S3 object
func (u *S3Uploader) GeneratePresignedURL(ctx context.Context, key string) (string, error) {
	logger := logging.ContextLogger(ctx, u.logger).With(
		slog.String("s3_key", key))

	presignClient := s3.NewPresignClient(u.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = u.urlExpiry
	})

	if err != nil {
		logger.Error("failed to generate presigned URL",
			slog.String("error", err.Error()))
		return "", fmt.Errorf("presign URL generation failed: %w", err)
	}

	logger.Debug("presigned URL generated",
		slog.String("url", req.URL),
		slog.Time("expires_at", time.Now().Add(u.urlExpiry)))

	u.metrics.MediaPresignedURLGenerated.Inc()

	return req.URL, nil
}

// Delete removes a file from S3 (used by cleanup job)
func (u *S3Uploader) Delete(ctx context.Context, key string) error {
	logger := logging.ContextLogger(ctx, u.logger).With(
		slog.String("s3_key", key))

	logger.Debug("deleting media from s3")

	_, err := u.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		logger.Error("s3 delete failed",
			slog.String("error", err.Error()))

		u.metrics.MediaDeleteAttempts.WithLabelValues("failure").Inc()
		return fmt.Errorf("s3 delete failed: %w", err)
	}

	logger.Info("s3 delete succeeded")
	u.metrics.MediaDeleteAttempts.WithLabelValues("success").Inc()

	return nil
}

// generateKey creates an organized S3 key structure
func (u *S3Uploader) generateKey(instanceID uuid.UUID, eventID uuid.UUID, mediaType string, contentType string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	// Extract file extension from content type
	ext := u.getExtensionFromContentType(contentType, mediaType)

	// Format: {instance_id}/{year}/{month}/{day}/{event_id}.{ext}
	return fmt.Sprintf("%s/%s/%s/%s/%s.%s",
		instanceID.String(), year, month, day, eventID.String(), ext)
}

// getExtensionFromContentType returns file extension based on content type
func (u *S3Uploader) getExtensionFromContentType(contentType string, mediaType string) string {
	// Use standard library mime.ExtensionsByType() first - supports 100+ types automatically
	exts, err := mime.ExtensionsByType(contentType)
	if err == nil && len(exts) > 0 {
		// Remove leading dot and return first (preferred) extension
		return strings.TrimPrefix(exts[0], ".")
	}

	// Fallback: Common MIME type to extension mapping for special cases
	mimeToExt := map[string]string{
		"audio/ogg; codecs=opus": "ogg", // WhatsApp voice notes
		"image/webp":             "webp",
	}

	// Try fallback mapping
	if ext, ok := mimeToExt[strings.ToLower(contentType)]; ok {
		return ext
	}

	// Try extracting from content type (e.g., "image/jpeg" -> "jpeg")
	parts := strings.Split(contentType, "/")
	if len(parts) == 2 {
		ext := strings.TrimSpace(parts[1])
		if ext != "" && ext != "plain" {
			return ext
		}
	}

	// Fall back to media type
	switch mediaType {
	case "image":
		return "jpg"
	case "video":
		return "mp4"
	case "audio":
		return "mp3"
	case "document":
		return "pdf"
	case "sticker":
		return "webp"
	case "voice":
		return "ogg"
	default:
		return "bin"
	}
}

// classifyS3Error categorizes S3 errors for metrics
func classifyS3Error(err error) string {
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "connection"):
		return "connection"
	case strings.Contains(errStr, "access denied"):
		return "access_denied"
	case strings.Contains(errStr, "no such bucket"):
		return "bucket_not_found"
	case strings.Contains(errStr, "file too large"):
		return "file_too_large"
	case strings.Contains(errStr, "network"):
		return "network"
	default:
		return "unknown"
	}
}

// ObjectExists checks if an object exists in S3 (for deduplication)
func (u *S3Uploader) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		// Object not found is not an error, return false
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetObjectSize returns the size of an object in S3
func (u *S3Uploader) GetObjectSize(ctx context.Context, key string) (int64, error) {
	resp, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return 0, err
	}

	return *resp.ContentLength, nil
}
