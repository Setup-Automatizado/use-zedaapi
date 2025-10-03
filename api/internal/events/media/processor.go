package media

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// MediaProcessor orchestrates the download → upload pipeline for WhatsApp media
type MediaProcessor struct {
	downloader      *MediaDownloader
	uploader        *S3Uploader
	localStorage    *LocalMediaStorage
	mediaRepo       persistence.MediaRepository
	outboxRepo      persistence.OutboxRepository
	metrics         *observability.Metrics
	logger          *slog.Logger
	maxRetries      int
	downloadTimeout time.Duration
	uploadTimeout   time.Duration
}

// ProcessResult contains the result of media processing
type ProcessResult struct {
	S3Bucket    string
	S3Key       string
	S3URL       string
	ContentType string
	FileName    string
	FileSize    int64
	SHA256      string
	MediaType   string
}

// NewMediaProcessor creates a new media processor
func NewMediaProcessor(
	ctx context.Context,
	cfg *config.Config,
	mediaRepo persistence.MediaRepository,
	outboxRepo persistence.OutboxRepository,
	metrics *observability.Metrics,
) (*MediaProcessor, error) {
	logger := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "media_processor"),
	)

	// Create downloader with generic type support
	downloader := NewMediaDownloader(metrics, cfg.Events.MediaDownloadTimeout)

	// Create S3 uploader
	uploader, err := NewS3Uploader(ctx, cfg, metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 uploader: %w", err)
	}

	// Create local storage fallback
	localStorage, err := NewLocalMediaStorage(ctx, cfg, metrics)
	if err != nil {
		logger.Warn("failed to initialize local media storage, fallback will be disabled",
			slog.String("error", err.Error()))
		// Continue without local storage fallback
		localStorage = nil
	}

	logger.Info("media processor initialized",
		slog.Int("max_retries", cfg.Events.MediaMaxRetries),
		slog.Duration("download_timeout", cfg.Events.MediaDownloadTimeout),
		slog.Duration("upload_timeout", cfg.Events.MediaUploadTimeout),
		slog.Bool("local_storage_enabled", localStorage != nil))

	return &MediaProcessor{
		downloader:      downloader,
		uploader:        uploader,
		localStorage:    localStorage,
		mediaRepo:       mediaRepo,
		outboxRepo:      outboxRepo,
		metrics:         metrics,
		logger:          logger,
		maxRetries:      cfg.Events.MediaMaxRetries,
		downloadTimeout: cfg.Events.MediaDownloadTimeout,
		uploadTimeout:   cfg.Events.MediaUploadTimeout,
	}, nil
}

// Process downloads WhatsApp media, uploads to S3, and updates metadata
func (p *MediaProcessor) Process(
	ctx context.Context,
	client *whatsmeow.Client,
	instanceID uuid.UUID,
	eventID uuid.UUID,
	msg proto.Message,
	mediaKey string,
) (*ProcessResult, error) {
	logger := logging.ContextLogger(ctx, p.logger).With(
		slog.String("instance_id", instanceID.String()),
		slog.String("event_id", eventID.String()),
		slog.String("media_key", mediaKey))

	start := time.Now()

	logger.Info("starting media processing")

	// Update status to downloading
	err := p.mediaRepo.UpdateDownloadStatus(ctx, eventID, persistence.MediaStatusDownloading, 1, nil, nil)
	if err != nil {
		logger.Error("failed to update download status",
			slog.String("error", err.Error()))
		// Continue anyway - this is not critical
	}

	// Step 1: Download media from WhatsApp
	downloadResult, err := p.downloader.Download(ctx, client, instanceID, eventID, msg)
	if err != nil {
		logger.Error("media download failed",
			slog.String("error", err.Error()))

		// Update status to failed
		errMsg := err.Error()
		_ = p.mediaRepo.UpdateDownloadStatus(ctx, eventID, persistence.MediaStatusFailed, 1, nil, &errMsg)

		p.metrics.MediaFailures.WithLabelValues(instanceID.String(), "unknown", "download").Inc()
		return nil, fmt.Errorf("download failed: %w", err)
	}

	logger.Info("media downloaded successfully",
		slog.Int64("file_size", downloadResult.FileSize),
		slog.String("content_type", downloadResult.ContentType))

	// Update status to downloaded
	err = p.mediaRepo.UpdateDownloadStatus(ctx, eventID, persistence.MediaStatusDownloaded, 1, nil, nil)
	if err != nil {
		logger.Error("failed to update download status",
			slog.String("error", err.Error()))
		// Continue anyway
	}

	// Step 2: Upload to S3
	s3Key, presignedURL, err := p.uploader.Upload(
		ctx,
		instanceID,
		eventID,
		downloadResult.MediaType,
		downloadResult.Data,
		downloadResult.ContentType,
		downloadResult.FileSize,
	)
	if err != nil {
		logger.Error("s3 upload failed",
			slog.String("error", err.Error()))

		p.metrics.MediaFailures.WithLabelValues(instanceID.String(), downloadResult.MediaType, "upload").Inc()
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	logger.Info("media uploaded to s3",
		slog.String("s3_key", s3Key),
		slog.String("s3_url", presignedURL))

	// TODO: Move expiry midia upload to config .env
	// Step 3: Update media metadata with S3 info
	// Calculate URL expiry (30 days from now)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	err = p.mediaRepo.UpdateUploadInfo(ctx, eventID, "funnelchat-media", s3Key, presignedURL, persistence.S3URLPresigned, &expiresAt)
	if err != nil {
		logger.Error("failed to update upload info",
			slog.String("error", err.Error()))
		// This is critical - return error
		return nil, fmt.Errorf("failed to update upload info: %w", err)
	}

	duration := time.Since(start)

	logger.Info("media processing completed",
		slog.Duration("total_duration", duration),
		slog.String("s3_key", s3Key))

	return &ProcessResult{
		S3Bucket:    "funnelchat-media",
		S3Key:       s3Key,
		S3URL:       presignedURL,
		ContentType: downloadResult.ContentType,
		FileName:    downloadResult.FileName,
		FileSize:    downloadResult.FileSize,
		SHA256:      downloadResult.SHA256,
		MediaType:   downloadResult.MediaType,
	}, nil
}

// ProcessWithRetry processes media with automatic retries and fallback chain
// Fallback chain: S3 → Local Storage → NULL
// ALWAYS sets media_processed=true to ensure webhook delivery
func (p *MediaProcessor) ProcessWithRetry(
	ctx context.Context,
	client *whatsmeow.Client,
	instanceID uuid.UUID,
	eventID uuid.UUID,
	msg proto.Message,
	mediaKey string,
) (*ProcessResult, error) {
	logger := logging.ContextLogger(ctx, p.logger).With(
		slog.String("instance_id", instanceID.String()),
		slog.String("event_id", eventID.String()))

	// Step 1: Download media from WhatsApp (this always happens first)
	downloadResult, downloadErr := p.downloadWithRetry(ctx, client, instanceID, eventID, msg)
	if downloadErr != nil {
		// Download failed permanently - mark as processed with NULL URL
		logger.Error("download failed permanently, marking as processed with NULL URL",
			slog.String("error", downloadErr.Error()))

		errMsg := fmt.Sprintf("download failed: %v", downloadErr)
		_ = p.updateOutboxMediaInfo(ctx, eventID, nil, &errMsg, true)

		p.metrics.MediaFallbackFailure.WithLabelValues(instanceID.String(), "unknown", "download_failed").Inc()
		return nil, downloadErr
	}

	// Download succeeded - now try upload with fallback chain

	// Step 2: Try S3 upload with retries
	logger.Info("attempting S3 upload")
	p.metrics.MediaFallbackAttempts.WithLabelValues(instanceID.String(), downloadResult.MediaType, "s3").Inc()

	s3Result, s3Err := p.uploadToS3WithRetry(ctx, instanceID, eventID, downloadResult)
	if s3Err == nil {
		// ✅ S3 success - update outbox and return
		logger.Info("media uploaded to S3 successfully",
			slog.String("s3_url", s3Result.S3URL))

		_ = p.updateOutboxMediaInfo(ctx, eventID, &s3Result.S3URL, nil, true)
		p.metrics.MediaFallbackSuccess.WithLabelValues(instanceID.String(), downloadResult.MediaType, "s3").Inc()

		return s3Result, nil
	}

	// S3 failed - try local storage fallback
	logger.Warn("S3 upload failed, attempting local storage fallback",
		slog.String("s3_error", s3Err.Error()))

	if p.localStorage == nil {
		// Local storage not configured - mark as processed with NULL URL
		logger.Error("local storage not configured, marking as processed with NULL URL")

		errMsg := fmt.Sprintf("s3 failed: %v; local storage: not configured", s3Err)
		_ = p.updateOutboxMediaInfo(ctx, eventID, nil, &errMsg, true)

		p.metrics.MediaFallbackFailure.WithLabelValues(instanceID.String(), downloadResult.MediaType, "no_fallback").Inc()
		return nil, fmt.Errorf("all storage methods failed: %w", s3Err)
	}

	// Step 3: Try local storage upload
	logger.Info("attempting local storage upload")
	p.metrics.MediaFallbackAttempts.WithLabelValues(instanceID.String(), downloadResult.MediaType, "local").Inc()

	localResult, localErr := p.uploadToLocalStorage(ctx, instanceID, eventID, downloadResult)
	if localErr == nil {
		// ✅ Local storage success - update outbox and return
		logger.Info("media uploaded to local storage successfully",
			slog.String("local_url", localResult.S3URL))

		_ = p.updateOutboxMediaInfo(ctx, eventID, &localResult.S3URL, nil, true)
		p.metrics.MediaFallbackSuccess.WithLabelValues(instanceID.String(), downloadResult.MediaType, "local").Inc()

		return localResult, nil
	}

	// Both S3 and local storage failed - mark as processed with NULL URL
	logger.Error("both S3 and local storage failed, marking as processed with NULL URL",
		slog.String("s3_error", s3Err.Error()),
		slog.String("local_error", localErr.Error()))

	// ✅ ALWAYS mark as processed to ensure webhook delivery
	errMsg := fmt.Sprintf("s3 failed: %v; local failed: %v", s3Err, localErr)
	_ = p.updateOutboxMediaInfo(ctx, eventID, nil, &errMsg, true)

	p.metrics.MediaFallbackFailure.WithLabelValues(instanceID.String(), downloadResult.MediaType, "all_failed").Inc()

	// Return error for logging, but webhook WILL be sent with media_url=NULL
	return nil, fmt.Errorf("all storage methods failed (s3: %v, local: %v)", s3Err, localErr)
}

// downloadWithRetry downloads media from WhatsApp with retries
func (p *MediaProcessor) downloadWithRetry(
	ctx context.Context,
	client *whatsmeow.Client,
	instanceID uuid.UUID,
	eventID uuid.UUID,
	msg proto.Message,
) (*DownloadResult, error) {
	logger := logging.ContextLogger(ctx, p.logger).With(
		slog.String("instance_id", instanceID.String()),
		slog.String("event_id", eventID.String()))

	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * 2 * time.Second
			logger.Warn("retrying download after backoff",
				slog.Int("attempt", attempt),
				slog.Duration("backoff", backoff))

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		result, err := p.downloader.Download(ctx, client, instanceID, eventID, msg)
		if err == nil {
			if attempt > 0 {
				logger.Info("download succeeded after retry", slog.Int("attempt", attempt))
			}
			return result, nil
		}

		lastErr = err

		if !isRetryableError(err) {
			logger.Error("non-retryable download error", slog.String("error", err.Error()))
			return nil, err
		}
	}

	return nil, fmt.Errorf("max download retries exceeded: %w", lastErr)
}

// uploadToS3WithRetry uploads to S3 with retries
func (p *MediaProcessor) uploadToS3WithRetry(
	ctx context.Context,
	instanceID uuid.UUID,
	eventID uuid.UUID,
	downloadResult *DownloadResult,
) (*ProcessResult, error) {
	logger := logging.ContextLogger(ctx, p.logger)

	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * 2 * time.Second
			logger.Warn("retrying S3 upload after backoff",
				slog.Int("attempt", attempt),
				slog.Duration("backoff", backoff))

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		s3Key, presignedURL, err := p.uploader.Upload(
			ctx,
			instanceID,
			eventID,
			downloadResult.MediaType,
			downloadResult.Data,
			downloadResult.ContentType,
			downloadResult.FileSize,
		)

		if err == nil {
			// Update media_metadata with S3 info
			expiresAt := time.Now().Add(30 * 24 * time.Hour)
			_ = p.mediaRepo.UpdateUploadInfoWithStorage(
				ctx, eventID,
				"funnelchat-media", s3Key, presignedURL,
				persistence.S3URLPresigned,
				persistence.StorageTypeS3,
				&expiresAt,
			)

			if attempt > 0 {
				logger.Info("S3 upload succeeded after retry", slog.Int("attempt", attempt))
			}

			return &ProcessResult{
				S3Bucket:    "funnelchat-media",
				S3Key:       s3Key,
				S3URL:       presignedURL,
				ContentType: downloadResult.ContentType,
				FileName:    downloadResult.FileName,
				FileSize:    downloadResult.FileSize,
				SHA256:      downloadResult.SHA256,
				MediaType:   downloadResult.MediaType,
			}, nil
		}

		lastErr = err

		if !isRetryableError(err) {
			logger.Error("non-retryable S3 error", slog.String("error", err.Error()))
			return nil, err
		}
	}

	return nil, fmt.Errorf("max S3 upload retries exceeded: %w", lastErr)
}

// uploadToLocalStorage uploads to local storage (no retries - single attempt)
func (p *MediaProcessor) uploadToLocalStorage(
	ctx context.Context,
	instanceID uuid.UUID,
	eventID uuid.UUID,
	downloadResult *DownloadResult,
) (*ProcessResult, error) {
	logger := logging.ContextLogger(ctx, p.logger)

	// Read data from io.Reader
	data, err := io.ReadAll(downloadResult.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to read download data: %w", err)
	}

	// Store locally
	storeResult, err := p.localStorage.StoreMedia(
		ctx,
		instanceID,
		eventID,
		downloadResult.MediaType,
		data,
		downloadResult.ContentType,
	)

	if err != nil {
		return nil, fmt.Errorf("local storage failed: %w", err)
	}

	// Update media_metadata with local URL
	_ = p.mediaRepo.UpdateUploadInfoWithStorage(
		ctx, eventID,
		"local", storeResult.LocalPath, storeResult.PublicURL,
		persistence.S3URLPresigned, // Reuse existing type
		persistence.StorageTypeLocal,
		&storeResult.ExpiresAt,
	)

	// Mark fallback as attempted and successful
	_ = p.mediaRepo.UpdateFallbackStatus(ctx, eventID, true, nil)

	logger.Info("media stored locally",
		slog.String("local_path", storeResult.LocalPath),
		slog.String("public_url", storeResult.PublicURL))

	return &ProcessResult{
		S3Bucket:    "local",
		S3Key:       storeResult.LocalPath,
		S3URL:       storeResult.PublicURL,
		ContentType: downloadResult.ContentType,
		FileName:    downloadResult.FileName,
		FileSize:    downloadResult.FileSize,
		SHA256:      downloadResult.SHA256,
		MediaType:   downloadResult.MediaType,
	}, nil
}

// updateOutboxMediaInfo updates event_outbox with media processing result
func (p *MediaProcessor) updateOutboxMediaInfo(
	ctx context.Context,
	eventID uuid.UUID,
	mediaURL *string,
	mediaError *string,
	processed bool,
) error {
	return p.outboxRepo.UpdateMediaInfo(ctx, eventID, mediaURL, mediaError, processed)
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	errStr := err.Error()

	// Network errors and timeouts are retryable
	switch {
	case err == context.DeadlineExceeded:
		return true
	case errStr == "timeout":
		return true
	case errStr == "connection":
		return true
	case errStr == "network":
		return true
	case errStr == "media_conn_refresh_failed":
		return true
	// Client errors are not retryable
	case errStr == "not_logged_in":
		return false
	case errStr == "no_url":
		return false
	case errStr == "not_downloadable":
		return false
	case errStr == "invalid_message":
		return false
	case errStr == "http_403":
		return false
	case errStr == "http_404":
		return false
	case errStr == "http_410":
		return false
	// Validation errors are not retryable
	case errStr == "invalid_hmac":
		return false
	case errStr == "invalid_enc_hash":
		return false
	case errStr == "invalid_hash":
		return false
	default:
		// Default to retrying unknown errors
		return true
	}
}
