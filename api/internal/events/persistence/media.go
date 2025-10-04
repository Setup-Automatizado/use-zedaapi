package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrMediaNotFound = errors.New("media metadata not found")
)

type MediaDownloadStatus string

const (
	MediaStatusPending     MediaDownloadStatus = "pending"
	MediaStatusDownloading MediaDownloadStatus = "downloading"
	MediaStatusDownloaded  MediaDownloadStatus = "downloaded"
	MediaStatusFailed      MediaDownloadStatus = "failed"
	MediaStatusCompleted   MediaDownloadStatus = "completed"
)

type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
	MediaTypeSticker  MediaType = "sticker"
	MediaTypeVoice    MediaType = "voice"
)

type S3URLType string

const (
	S3URLPresigned S3URLType = "presigned"
	S3URLPublic    S3URLType = "public"
	S3URLCDN       S3URLType = "cdn"
)

type StorageType string

const (
	StorageTypeS3    StorageType = "s3"
	StorageTypeLocal StorageType = "local"
	StorageTypeNull  StorageType = "null"
)

type MediaMetadata struct {
	ID                  int64
	EventID             uuid.UUID
	InstanceID          uuid.UUID
	MediaKey            string
	FileSHA256          *string
	FileEncSHA256       *string
	DirectPath          string
	MediaType           MediaType
	MimeType            *string
	FileLength          *int64
	DownloadStatus      MediaDownloadStatus
	DownloadAttempts    int
	DownloadStartedAt   *time.Time
	DownloadedAt        *time.Time
	DownloadDurationMS  *int
	DownloadedSizeBytes *int64
	DownloadError       *string
	S3Bucket            *string
	S3Key               *string
	S3URL               *string
	S3URLType           S3URLType
	URLExpiresAt        *time.Time
	UploadStartedAt     *time.Time
	UploadedAt          *time.Time
	UploadDurationMS    *int
	UploadError         *string
	ProcessingWorkerID  *string
	ProcessingStartedAt *time.Time
	CompletedAt         *time.Time
	NextRetryAt         *time.Time
	MaxRetries          int
	StorageType         StorageType
	FallbackAttempted   bool
	FallbackError       *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type MediaRepository interface {
	InsertMedia(ctx context.Context, media *MediaMetadata) error

	PollPendingDownloads(ctx context.Context, limit int) ([]*MediaMetadata, error)

	UpdateDownloadStatus(ctx context.Context, eventID uuid.UUID, status MediaDownloadStatus, attempts int, nextRetry *time.Time, downloadError *string) error

	UpdateDownloadComplete(ctx context.Context, eventID uuid.UUID, durationMS int, sizeBytes int64) error

	UpdateUploadInfo(ctx context.Context, eventID uuid.UUID, bucket, key, url string, urlType S3URLType, expiresAt *time.Time) error

	UpdateUploadInfoWithStorage(ctx context.Context, eventID uuid.UUID, bucket, key, url string, urlType S3URLType, storageType StorageType, expiresAt *time.Time) error

	UpdateFallbackStatus(ctx context.Context, eventID uuid.UUID, attempted bool, fallbackError *string) error

	UpdateUploadComplete(ctx context.Context, eventID uuid.UUID, durationMS int) error

	MarkComplete(ctx context.Context, eventID uuid.UUID) error

	GetByEventID(ctx context.Context, eventID uuid.UUID) (*MediaMetadata, error)

	GetByInstanceID(ctx context.Context, instanceID uuid.UUID, limit, offset int) ([]*MediaMetadata, error)

	CountPendingByInstance(ctx context.Context, instanceID uuid.UUID) (int, error)

	CountFailedByInstance(ctx context.Context, instanceID uuid.UUID) (int, error)

	GetMediaTypeStats(ctx context.Context, since time.Time) (map[MediaType]int, error)

	AcquireForProcessing(ctx context.Context, eventID uuid.UUID, workerID string) (bool, error)

	ReleaseFromProcessing(ctx context.Context, eventID uuid.UUID, workerID string) error
}

type mediaRepository struct {
	pool *pgxpool.Pool
}

func NewMediaRepository(pool *pgxpool.Pool) MediaRepository {
	return &mediaRepository{pool: pool}
}

func (r *mediaRepository) InsertMedia(ctx context.Context, media *MediaMetadata) error {
	query := `
		INSERT INTO media_metadata (
			event_id, instance_id, media_key, file_sha256, file_enc_sha256,
			direct_path, media_type, mime_type, file_length,
			download_status, download_attempts, max_retries, s3_url_type
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(
		ctx, query,
		media.EventID, media.InstanceID, media.MediaKey, media.FileSHA256, media.FileEncSHA256,
		media.DirectPath, media.MediaType, media.MimeType, media.FileLength,
		media.DownloadStatus, media.DownloadAttempts, media.MaxRetries, media.S3URLType,
	).Scan(&media.ID, &media.CreatedAt, &media.UpdatedAt)

	return err
}

func (r *mediaRepository) PollPendingDownloads(ctx context.Context, limit int) ([]*MediaMetadata, error) {
	query := `
		SELECT
			id, event_id, instance_id, media_key, file_sha256, file_enc_sha256,
			direct_path, media_type, mime_type, file_length,
			download_status, download_attempts, download_started_at, downloaded_at,
			download_duration_ms, downloaded_size_bytes, download_error,
			s3_bucket, s3_key, s3_url, s3_url_type, url_expires_at,
			upload_started_at, uploaded_at, upload_duration_ms, upload_error,
			processing_worker_id, processing_started_at, completed_at,
			next_retry_at, max_retries, created_at, updated_at
		FROM media_metadata
		WHERE download_status IN ('pending', 'failed')
		  AND download_attempts < max_retries
		  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		  AND processing_worker_id IS NULL
		ORDER BY created_at ASC
		LIMIT $1`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medias []*MediaMetadata
	for rows.Next() {
		media := &MediaMetadata{}
		err := rows.Scan(
			&media.ID, &media.EventID, &media.InstanceID, &media.MediaKey, &media.FileSHA256, &media.FileEncSHA256,
			&media.DirectPath, &media.MediaType, &media.MimeType, &media.FileLength,
			&media.DownloadStatus, &media.DownloadAttempts, &media.DownloadStartedAt, &media.DownloadedAt,
			&media.DownloadDurationMS, &media.DownloadedSizeBytes, &media.DownloadError,
			&media.S3Bucket, &media.S3Key, &media.S3URL, &media.S3URLType, &media.URLExpiresAt,
			&media.UploadStartedAt, &media.UploadedAt, &media.UploadDurationMS, &media.UploadError,
			&media.ProcessingWorkerID, &media.ProcessingStartedAt, &media.CompletedAt,
			&media.NextRetryAt, &media.MaxRetries, &media.CreatedAt, &media.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		medias = append(medias, media)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return medias, nil
}

func (r *mediaRepository) UpdateDownloadStatus(ctx context.Context, eventID uuid.UUID, status MediaDownloadStatus, attempts int, nextRetry *time.Time, downloadError *string) error {
	query := `
		UPDATE media_metadata
		SET download_status = $2,
		    download_attempts = $3,
		    next_retry_at = $4,
		    download_error = $5
		WHERE event_id = $1`

	commandTag, err := r.pool.Exec(ctx, query, eventID, status, attempts, nextRetry, downloadError)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrMediaNotFound
	}

	return nil
}

func (r *mediaRepository) UpdateDownloadComplete(ctx context.Context, eventID uuid.UUID, durationMS int, sizeBytes int64) error {
	query := `
		UPDATE media_metadata
		SET download_status = 'downloaded',
		    downloaded_at = NOW(),
		    download_duration_ms = $2,
		    downloaded_size_bytes = $3
		WHERE event_id = $1`

	commandTag, err := r.pool.Exec(ctx, query, eventID, durationMS, sizeBytes)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrMediaNotFound
	}

	return nil
}

func (r *mediaRepository) UpdateUploadInfo(ctx context.Context, eventID uuid.UUID, bucket, key, url string, urlType S3URLType, expiresAt *time.Time) error {
	query := `
		UPDATE media_metadata
		SET s3_bucket = $2,
		    s3_key = $3,
		    s3_url = $4,
		    s3_url_type = $5,
		    url_expires_at = $6
		WHERE event_id = $1`

	commandTag, err := r.pool.Exec(ctx, query, eventID, bucket, key, url, urlType, expiresAt)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrMediaNotFound
	}

	return nil
}

func (r *mediaRepository) UpdateUploadInfoWithStorage(ctx context.Context, eventID uuid.UUID, bucket, key, url string, urlType S3URLType, storageType StorageType, expiresAt *time.Time) error {
	query := `
		UPDATE media_metadata
		SET s3_bucket = $2,
		    s3_key = $3,
		    s3_url = $4,
		    s3_url_type = $5,
		    url_expires_at = $6,
		    storage_type = $7
		WHERE event_id = $1`

	commandTag, err := r.pool.Exec(ctx, query, eventID, bucket, key, url, urlType, expiresAt, storageType)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrMediaNotFound
	}

	return nil
}

func (r *mediaRepository) UpdateFallbackStatus(ctx context.Context, eventID uuid.UUID, attempted bool, fallbackError *string) error {
	query := `
		UPDATE media_metadata
		SET fallback_attempted = $2,
		    fallback_error = $3
		WHERE event_id = $1`

	commandTag, err := r.pool.Exec(ctx, query, eventID, attempted, fallbackError)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrMediaNotFound
	}

	return nil
}

func (r *mediaRepository) UpdateUploadComplete(ctx context.Context, eventID uuid.UUID, durationMS int) error {
	query := `
		UPDATE media_metadata
		SET uploaded_at = NOW(),
		    upload_duration_ms = $2
		WHERE event_id = $1`

	commandTag, err := r.pool.Exec(ctx, query, eventID, durationMS)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrMediaNotFound
	}

	return nil
}

func (r *mediaRepository) MarkComplete(ctx context.Context, eventID uuid.UUID) error {
	query := `
		UPDATE media_metadata
		SET download_status = 'completed',
		    completed_at = NOW()
		WHERE event_id = $1`

	commandTag, err := r.pool.Exec(ctx, query, eventID)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrMediaNotFound
	}

	return nil
}

func (r *mediaRepository) GetByEventID(ctx context.Context, eventID uuid.UUID) (*MediaMetadata, error) {
	query := `
		SELECT
			id, event_id, instance_id, media_key, file_sha256, file_enc_sha256,
			direct_path, media_type, mime_type, file_length,
			download_status, download_attempts, download_started_at, downloaded_at,
			download_duration_ms, downloaded_size_bytes, download_error,
			s3_bucket, s3_key, s3_url, s3_url_type, url_expires_at,
			upload_started_at, uploaded_at, upload_duration_ms, upload_error,
			processing_worker_id, processing_started_at, completed_at,
			next_retry_at, max_retries, created_at, updated_at
		FROM media_metadata
		WHERE event_id = $1`

	media := &MediaMetadata{}
	err := r.pool.QueryRow(ctx, query, eventID).Scan(
		&media.ID, &media.EventID, &media.InstanceID, &media.MediaKey, &media.FileSHA256, &media.FileEncSHA256,
		&media.DirectPath, &media.MediaType, &media.MimeType, &media.FileLength,
		&media.DownloadStatus, &media.DownloadAttempts, &media.DownloadStartedAt, &media.DownloadedAt,
		&media.DownloadDurationMS, &media.DownloadedSizeBytes, &media.DownloadError,
		&media.S3Bucket, &media.S3Key, &media.S3URL, &media.S3URLType, &media.URLExpiresAt,
		&media.UploadStartedAt, &media.UploadedAt, &media.UploadDurationMS, &media.UploadError,
		&media.ProcessingWorkerID, &media.ProcessingStartedAt, &media.CompletedAt,
		&media.NextRetryAt, &media.MaxRetries, &media.CreatedAt, &media.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMediaNotFound
		}
		return nil, err
	}

	return media, nil
}

func (r *mediaRepository) GetByInstanceID(ctx context.Context, instanceID uuid.UUID, limit, offset int) ([]*MediaMetadata, error) {
	query := `
		SELECT
			id, event_id, instance_id, media_key, file_sha256, file_enc_sha256,
			direct_path, media_type, mime_type, file_length,
			download_status, download_attempts, download_started_at, downloaded_at,
			download_duration_ms, downloaded_size_bytes, download_error,
			s3_bucket, s3_key, s3_url, s3_url_type, url_expires_at,
			upload_started_at, uploaded_at, upload_duration_ms, upload_error,
			processing_worker_id, processing_started_at, completed_at,
			next_retry_at, max_retries, created_at, updated_at
		FROM media_metadata
		WHERE instance_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, instanceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medias []*MediaMetadata
	for rows.Next() {
		media := &MediaMetadata{}
		err := rows.Scan(
			&media.ID, &media.EventID, &media.InstanceID, &media.MediaKey, &media.FileSHA256, &media.FileEncSHA256,
			&media.DirectPath, &media.MediaType, &media.MimeType, &media.FileLength,
			&media.DownloadStatus, &media.DownloadAttempts, &media.DownloadStartedAt, &media.DownloadedAt,
			&media.DownloadDurationMS, &media.DownloadedSizeBytes, &media.DownloadError,
			&media.S3Bucket, &media.S3Key, &media.S3URL, &media.S3URLType, &media.URLExpiresAt,
			&media.UploadStartedAt, &media.UploadedAt, &media.UploadDurationMS, &media.UploadError,
			&media.ProcessingWorkerID, &media.ProcessingStartedAt, &media.CompletedAt,
			&media.NextRetryAt, &media.MaxRetries, &media.CreatedAt, &media.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		medias = append(medias, media)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return medias, nil
}

func (r *mediaRepository) CountPendingByInstance(ctx context.Context, instanceID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM media_metadata
		WHERE instance_id = $1
		  AND download_status IN ('pending', 'downloading', 'failed')
		  AND download_attempts < max_retries`

	var count int
	err := r.pool.QueryRow(ctx, query, instanceID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *mediaRepository) CountFailedByInstance(ctx context.Context, instanceID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM media_metadata
		WHERE instance_id = $1
		  AND download_status = 'failed'
		  AND download_attempts >= max_retries`

	var count int
	err := r.pool.QueryRow(ctx, query, instanceID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *mediaRepository) GetMediaTypeStats(ctx context.Context, since time.Time) (map[MediaType]int, error) {
	query := `
		SELECT media_type, download_status, COUNT(*)
		FROM media_metadata
		WHERE created_at >= $1
		GROUP BY media_type, download_status
		ORDER BY media_type, download_status`

	rows, err := r.pool.Query(ctx, query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[MediaType]int)
	for rows.Next() {
		var mediaType MediaType
		var status MediaDownloadStatus
		var count int
		if err := rows.Scan(&mediaType, &status, &count); err != nil {
			return nil, err
		}
		stats[mediaType] += count
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *mediaRepository) AcquireForProcessing(ctx context.Context, eventID uuid.UUID, workerID string) (bool, error) {
	query := `
		UPDATE media_metadata
		SET processing_worker_id = $2,
		    processing_started_at = NOW(),
		    download_started_at = NOW(),
		    download_status = 'downloading'
		WHERE event_id = $1
		  AND processing_worker_id IS NULL
		  AND download_status IN ('pending', 'failed')
		  AND download_attempts < max_retries`

	commandTag, err := r.pool.Exec(ctx, query, eventID, workerID)
	if err != nil {
		return false, err
	}

	return commandTag.RowsAffected() > 0, nil
}

func (r *mediaRepository) ReleaseFromProcessing(ctx context.Context, eventID uuid.UUID, workerID string) error {
	query := `
		UPDATE media_metadata
		SET processing_worker_id = NULL,
		    processing_started_at = NULL
		WHERE event_id = $1
		  AND processing_worker_id = $2`

	_, err := r.pool.Exec(ctx, query, eventID, workerID)
	return err
}
