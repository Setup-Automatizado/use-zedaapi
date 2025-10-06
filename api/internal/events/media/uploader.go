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

type S3Uploader struct {
	client        *s3.Client
	uploader      *manager.Uploader
	bucket        string
	urlExpiry     time.Duration
	acl           string
	usePresigned  bool
	publicBaseURL string
	metrics       *observability.Metrics
	logger        *slog.Logger
}

func NewS3Uploader(ctx context.Context, cfg *config.Config, metrics *observability.Metrics) (*S3Uploader, error) {
	logger := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "s3_uploader"),
	)

	awsCfg := aws.Config{
		Region:      cfg.S3.Region,
		Credentials: credentials.NewStaticCredentialsProvider(cfg.S3.AccessKey, cfg.S3.SecretKey, ""),
	}

	if cfg.S3.Endpoint != "" {
		awsCfg.BaseEndpoint = aws.String(cfg.S3.Endpoint)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	uploader := manager.NewUploader(s3Client, func(u *manager.Uploader) {
		u.PartSize = cfg.Events.MediaChunkSize
		u.Concurrency = 100
	})

	logger.Info("s3 uploader initialized",
		slog.String("bucket", cfg.S3.Bucket),
		slog.String("region", cfg.S3.Region),
		slog.Duration("url_expiry", cfg.S3.URLExpiry),
		slog.Bool("use_presigned_urls", cfg.S3.UsePresignedURLs),
		slog.String("public_base_url", cfg.S3.PublicBaseURL),
		slog.String("acl", cfg.S3.ACL))

	if cfg.S3.ACL != "" {
		logger.Warn("using S3 ACL (deprecated AWS pattern - prefer bucket policies)",
			slog.String("acl", cfg.S3.ACL))
	}

	return &S3Uploader{
		client:        s3Client,
		uploader:      uploader,
		bucket:        cfg.S3.Bucket,
		urlExpiry:     cfg.S3.URLExpiry,
		acl:           cfg.S3.ACL,
		usePresigned:  cfg.S3.UsePresignedURLs,
		publicBaseURL: strings.TrimSuffix(cfg.S3.PublicBaseURL, "/"),
		metrics:       metrics,
		logger:        logger,
	}, nil
}

func (u *S3Uploader) Upload(ctx context.Context, instanceID uuid.UUID, eventID uuid.UUID, mediaType string, reader io.Reader, contentType string, fileSize int64) (key string, mediaURL string, err error) {
	logger := logging.ContextLogger(ctx, u.logger).With(
		slog.String("instance_id", instanceID.String()),
		slog.String("event_id", eventID.String()),
		slog.String("media_type", mediaType),
		slog.Int64("file_size", fileSize))

	start := time.Now()

	key = u.generateKey(instanceID, eventID, mediaType, contentType)

	logger.Debug("uploading media to s3", slog.String("s3_key", key))

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

	if u.acl != "" {
		uploadInput.ACL = types.ObjectCannedACL(u.acl)
	}

	uploadResult, err := u.uploader.Upload(ctx, uploadInput)

	duration := time.Since(start)

	if err != nil {
		logger.Error("s3 upload failed",
			slog.String("error", err.Error()),
			slog.Duration("duration", duration))

		u.metrics.MediaUploadsTotal.WithLabelValues(instanceID.String(), mediaType, "failure").Inc()
		u.metrics.MediaUploadAttempts.WithLabelValues("failure").Inc()
		u.metrics.MediaUploadErrors.WithLabelValues(classifyS3Error(err)).Inc()

		return "", "", fmt.Errorf("s3 upload failed: %w", err)
	}

	logger.Info("s3 upload succeeded",
		slog.String("location", uploadResult.Location),
		slog.Duration("duration", duration))

	u.metrics.MediaUploadsTotal.WithLabelValues(instanceID.String(), mediaType, "success").Inc()
	u.metrics.MediaUploadAttempts.WithLabelValues("success").Inc()
	u.metrics.MediaUploadDuration.WithLabelValues(instanceID.String(), mediaType).Observe(duration.Seconds())
	u.metrics.MediaUploadSizeBytes.WithLabelValues(mediaType).Add(float64(fileSize))

	if u.usePresigned {
		mediaURL, err = u.GeneratePresignedURL(ctx, key)
		if err != nil {
			return key, "", fmt.Errorf("failed to generate presigned URL: %w", err)
		}
	} else {
		mediaURL = u.buildPublicURL(key, uploadResult.Location)
	}

	return key, mediaURL, nil
}

func (u *S3Uploader) GeneratePresignedURL(ctx context.Context, key string) (string, error) {
	if !u.usePresigned {
		return "", fmt.Errorf("presigned URLs are disabled")
	}

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

func (u *S3Uploader) buildPublicURL(key string, uploadLocation string) string {
	if u.publicBaseURL != "" {
		return fmt.Sprintf("%s/%s", u.publicBaseURL, key)
	}
	return uploadLocation
}

func (u *S3Uploader) UsesPresignedURLs() bool {
	return u.usePresigned
}

func (u *S3Uploader) Delete(ctx context.Context, key string) error {
	return u.DeleteObject(ctx, u.bucket, key)
}

func (u *S3Uploader) DeleteObject(ctx context.Context, bucket, key string) error {
	if key == "" {
		return fmt.Errorf("s3 delete failed: empty key")
	}
	if bucket == "" {
		bucket = u.bucket
	}

	logger := logging.ContextLogger(ctx, u.logger).With(
		slog.String("s3_bucket", bucket),
		slog.String("s3_key", key))

	logger.Debug("deleting media from s3")

	_, err := u.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
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

func (u *S3Uploader) generateKey(instanceID uuid.UUID, eventID uuid.UUID, mediaType string, contentType string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	ext := u.getExtensionFromContentType(contentType, mediaType)

	return fmt.Sprintf("%s/%s/%s/%s/%s.%s",
		instanceID.String(), year, month, day, eventID.String(), ext)
}

func (u *S3Uploader) getExtensionFromContentType(contentType string, mediaType string) string {
	exts, err := mime.ExtensionsByType(contentType)
	if err == nil && len(exts) > 0 {
		return strings.TrimPrefix(exts[0], ".")
	}

	mimeToExt := map[string]string{
		"audio/ogg; codecs=opus": "ogg",
		"image/webp":             "webp",
	}

	if ext, ok := mimeToExt[strings.ToLower(contentType)]; ok {
		return ext
	}

	parts := strings.Split(contentType, "/")
	if len(parts) == 2 {
		ext := strings.TrimSpace(parts[1])
		if ext != "" && ext != "plain" {
			return ext
		}
	}

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

func (u *S3Uploader) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

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
