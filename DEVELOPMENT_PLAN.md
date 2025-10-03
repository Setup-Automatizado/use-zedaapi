# WhatsApp API Development Plan - Phase 6 & 7

**Project Status**: 70% Complete (Phases 1-5 ✅) | **Remaining**: 30% (Phases 6-7 📋)
**Timeline**: 13 days development + 5 weeks rollout = **7 weeks to production**
**Last Updated**: 2025-10-03

---

## Executive Summary

This document provides a comprehensive implementation plan for the remaining 30% of the WhatsApp API project, specifically **Phase 6: Media Processing** and **Phase 7: Background Jobs**. The existing foundation (Phases 1-5) includes database infrastructure, WhatsApp client integration, instance management, complete dispatch system, and event pipeline with outbox/DLQ persistence.

### Key Deliverables

**Phase 6: Media Processing (35 hours)**
- Async media download from WhatsApp
- S3 upload with encryption and metadata
- Presigned URL generation for webhook delivery
- Media deduplication and caching
- Worker pool for parallel processing

**Phase 7: Background Jobs (32 hours)**
- DLQ reprocessing (unblock failed webhooks)
- Media retry job (recover failed media)
- Outbox cleanup (prevent table bloat)
- Media cleanup (manage S3 costs)
- Health monitor (detect stale instances)

### Critical Dependencies

- **External**: AWS S3 (or MinIO for dev), go-co-op/gocron library
- **Internal**: Phases 1-5 must be stable (database, ClientRegistry, DispatchCoordinator)
- **Infrastructure**: S3 bucket provisioned, IAM roles configured, monitoring dashboards ready

---

## Architecture Overview

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                     WhatsApp API System                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │  WhatsApp    │───▶│   Event      │───▶│   Webhook    │  │
│  │  Client      │    │   Pipeline   │    │   Delivery   │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│         │                    │                    │          │
│         │              ┌─────▼─────┐       ┌─────▼─────┐   │
│         │              │   Media   │       │  Dispatch │   │
│         │              │ Processing│       │   System  │   │
│         │              └───────────┘       └───────────┘   │
│         │                    │                              │
│         │              ┌─────▼─────┐                        │
│         │              │    S3     │                        │
│         │              │  Storage  │                        │
│         │              └───────────┘                        │
│         │                                                    │
│    ┌────▼──────────────────────────────────────────────┐   │
│    │           Background Jobs Scheduler               │   │
│    │  ┌──────────┐ ┌──────────┐ ┌──────────┐         │   │
│    │  │   DLQ    │ │  Health  │ │ Cleanup  │   ...   │   │
│    │  │Reprocess │ │ Monitor  │ │   Jobs   │         │   │
│    │  └──────────┘ └──────────┘ └──────────┘         │   │
│    └───────────────────────────────────────────────────┘   │
│                                                               │
│    Database: Postgres (api_core + whatsmeow_store)          │
│    Cache/Locks: Redis                                        │
│    Observability: Prometheus + Sentry                        │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow: Message with Media

```
1. WhatsApp Message Received (with media)
   ↓
2. EventHandler captures event → BufferedCapture
   ↓
3. Transform: whatsmeow → internal → webhook format
   ↓
4. EventProcessor detects media, creates media_job
   ↓
5. Outbox status set to 'pending_media'
   ↓
6. MediaWorker polls media_jobs table
   ↓
7. MediaProcessor: download → S3 upload → presigned URL
   ↓
8. Update media_metadata (for deduplication)
   ↓
9. Update media_job status to 'completed'
   ↓
10. Update outbox status back to 'pending'
    ↓
11. InstanceWorker picks up event, delivers webhook with S3 URL
```

---

## Phase 6: Media Processing (35 hours)

### Goals
- Enable WhatsApp media (images, videos, audio, documents, stickers) to be automatically downloaded, uploaded to S3, and delivered via webhooks with presigned URLs
- Implement deduplication to avoid redundant storage
- Provide async processing with retry logic
- Integrate seamlessly with existing dispatch system

### Architecture Components

```
internal/events/media/
├── coordinator.go     // MediaCoordinator - manages worker pool (4h)
├── worker.go          // MediaWorker - polls media_jobs (4h)
├── processor.go       // MediaProcessor - download → upload → URL (5h)
├── downloader.go      // WhatsApp media downloader (3h)
└── uploader.go        // S3 uploader + presigned URLs (4h)

internal/events/persistence/
├── media_jobs.go      // MediaJobsRepository (2h)
└── media_metadata.go  // MediaMetadataRepository (1h)
```

### Implementation Steps

#### Step 6.1: Database Schema (2 hours) ⚠️ CRITICAL

**Files**:
- `api/migrations/000004_create_media_metadata.sql`
- `api/migrations/000006_create_media_jobs.sql`
- Update `api/migrations/000002_create_event_outbox.sql` (add 'pending_media' status)

**Schema: media_jobs**
```sql
CREATE TABLE media_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES webhook_outbox(id) ON DELETE CASCADE,
    media_type VARCHAR(20) NOT NULL, -- image, video, audio, document, sticker
    whatsapp_url TEXT NOT NULL,
    media_key BYTEA NOT NULL,
    file_enc_sha256 BYTEA NOT NULL,
    file_sha256 BYTEA, -- filled after download
    file_length BIGINT NOT NULL,
    mime_type VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    s3_bucket VARCHAR(100),
    s3_key TEXT,
    s3_url TEXT, -- presigned URL
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP
);

CREATE INDEX idx_media_jobs_status ON media_jobs(status, created_at);
CREATE INDEX idx_media_jobs_instance ON media_jobs(instance_id);
CREATE INDEX idx_media_jobs_event ON media_jobs(event_id);
```

**Schema: media_metadata**
```sql
CREATE TABLE media_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    file_sha256 BYTEA NOT NULL,
    s3_bucket VARCHAR(100) NOT NULL,
    s3_key TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    media_type VARCHAR(20) NOT NULL,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_accessed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    access_count INTEGER NOT NULL DEFAULT 0,
    UNIQUE(s3_bucket, s3_key)
);

CREATE UNIQUE INDEX idx_media_dedup ON media_metadata(instance_id, file_sha256);
CREATE INDEX idx_media_s3_key ON media_metadata(s3_key);
CREATE INDEX idx_media_cleanup ON media_metadata(last_accessed_at) WHERE access_count < 5;
```

**Acceptance Criteria**:
- [ ] Migrations run successfully on clean database
- [ ] Foreign key constraints validated
- [ ] Indexes created with correct columns
- [ ] Status enums validated
- [ ] Rollback migration tested

#### Step 6.2: Configuration (1 hour) ⚠️ CRITICAL

**Files**:
- `api/internal/config/config.go`

**Add MediaConfig**:
```go
type Config struct {
    // ... existing fields
    Media MediaConfig
}

type MediaConfig struct {
    Enabled            bool          `env:"MEDIA_PROCESSING_ENABLED" envDefault:"false"`
    S3Bucket           string        `env:"MEDIA_S3_BUCKET" envDefault:"whatsapp-media"`
    S3Region           string        `env:"MEDIA_S3_REGION" envDefault:"us-east-1"`
    S3AccessKey        string        `env:"MEDIA_S3_ACCESS_KEY"`
    S3SecretKey        string        `env:"MEDIA_S3_SECRET_KEY"`
    S3Endpoint         string        `env:"MEDIA_S3_ENDPOINT"` // For MinIO
    PresignedURLExpiry time.Duration `env:"MEDIA_PRESIGNED_EXPIRY" envDefault:"24h"`
    MaxFileSize        int64         `env:"MEDIA_MAX_FILE_SIZE" envDefault:"52428800"` // 50MB
    WorkerCount        int           `env:"MEDIA_WORKER_COUNT" envDefault:"3"`
    DownloadTimeout    time.Duration `env:"MEDIA_DOWNLOAD_TIMEOUT" envDefault:"5m"`
    UploadTimeout      time.Duration `env:"MEDIA_UPLOAD_TIMEOUT" envDefault:"10m"`
    RetryPolicy        RetryConfig
}
```

**Environment Variables**:
```bash
MEDIA_PROCESSING_ENABLED=false  # Feature flag
MEDIA_S3_BUCKET=whatsapp-media
MEDIA_S3_REGION=us-east-1
MEDIA_S3_ACCESS_KEY=<aws-access-key>
MEDIA_S3_SECRET_KEY=<aws-secret-key>
MEDIA_S3_ENDPOINT=http://localhost:9000  # MinIO for dev
MEDIA_PRESIGNED_EXPIRY=24h
MEDIA_MAX_FILE_SIZE=52428800  # 50MB
MEDIA_WORKER_COUNT=3
MEDIA_DOWNLOAD_TIMEOUT=5m
MEDIA_UPLOAD_TIMEOUT=10m
```

**Acceptance Criteria**:
- [ ] Config validation enforces required fields
- [ ] Default values reasonable
- [ ] Feature flag works (enabled/disabled)
- [ ] S3 credentials validated on startup

#### Step 6.3: S3 Integration (4 hours)

**Files**:
- `api/internal/events/media/uploader.go`

**Dependencies**:
```bash
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/feature/s3/manager
```

**Implementation**:
```go
type S3Uploader struct {
    client       *s3.Client
    bucket       string
    urlExpiry    time.Duration
    uploadMgr    *manager.Uploader
    metrics      *observability.Metrics
}

// Upload uploads file to S3 and returns key + presigned URL
func (u *S3Uploader) Upload(ctx context.Context, instanceID uuid.UUID,
    messageID string, mediaType string, reader io.Reader,
    contentType string, fileSize int64) (key string, url string, err error)

// GeneratePresignedURL creates time-limited URL for S3 object
func (u *S3Uploader) GeneratePresignedURL(ctx context.Context,
    key string, expiry time.Duration) (string, error)

// Delete removes file from S3 (for cleanup job)
func (u *S3Uploader) Delete(ctx context.Context, key string) error
```

**S3 Key Structure**:
```
{instance_id}/{year}/{month}/{day}/{message_id}_{file_sha256}.{ext}
Example: 550e8400-e29b-41d4-a716-446655440000/2025/10/03/msg123_abc123.jpg
```

**Metrics**:
```go
media_upload_attempts_total{status="success|failure"}
media_upload_duration_seconds{media_type="image|video|audio|document"}
media_upload_size_bytes{media_type}
media_presigned_url_generated_total
```

**Acceptance Criteria**:
- [ ] Upload succeeds for all media types
- [ ] Presigned URLs generated with correct expiry
- [ ] Multipart upload for files >5MB
- [ ] Error handling for network failures
- [ ] Metrics exported
- [ ] Unit tests with S3 mock (MinIO)

#### Step 6.4: Media Downloader (3 hours)

**Files**:
- `api/internal/events/media/downloader.go`

**Implementation**:
```go
type MediaDownloader struct {
    client  *whatsmeow.Client
    timeout time.Duration
    metrics *observability.Metrics
}

// Download downloads and decrypts media from WhatsApp
func (d *MediaDownloader) Download(ctx context.Context,
    url string, mediaKey []byte, mediaType whatsmeow.MediaType,
    fileLength int, fileSha256 []byte, fileEncSha256 []byte) (io.ReadCloser, error)

// DownloadMediaWithPath wraps client.Download for error handling
func (d *MediaDownloader) DownloadMediaWithPath(ctx context.Context,
    directPath string, encFileHash []byte, fileHash []byte,
    mediaKey []byte, fileLength int, mediaType whatsmeow.MediaType,
    mmsType string) (io.ReadCloser, error)
```

**Error Handling**:
- **MediaURLExpired**: Retry once, then mark as permanent failure
- **DecryptionFailed**: Permanent failure (corrupted MediaKey)
- **NetworkTimeout**: Retry with exponential backoff
- **FileTooLarge**: Permanent failure (exceeds MAX_FILE_SIZE)

**Metrics**:
```go
media_download_attempts_total{status="success|failure", error_type}
media_download_duration_seconds{media_type}
media_download_size_bytes{media_type}
media_decryption_errors_total
```

**Acceptance Criteria**:
- [ ] Downloads all media types successfully
- [ ] Decryption validates file hashes
- [ ] Timeout enforced (5 minutes default)
- [ ] Retries transient failures
- [ ] Returns io.ReadCloser for streaming
- [ ] Unit tests with mock WhatsApp client

#### Step 6.5: Media Processor (5 hours)

**Files**:
- `api/internal/events/media/processor.go`

**Implementation**:
```go
type MediaProcessor struct {
    downloader *MediaDownloader
    uploader   *S3Uploader
    metadataRepo *persistence.MediaMetadataRepository
    jobsRepo     *persistence.MediaJobsRepository
    metrics      *observability.Metrics
}

// ProcessMediaJob orchestrates download → upload → URL generation
func (p *MediaProcessor) ProcessMediaJob(ctx context.Context,
    job *models.MediaJob) error {

    // 1. Check deduplication cache
    if metadata, err := p.metadataRepo.FindByFileHash(ctx, job.InstanceID, job.FileEncSha256); err == nil {
        // Hit! Update access count, return existing URL
        return p.useExistingMedia(ctx, job, metadata)
    }

    // 2. Download from WhatsApp
    reader, err := p.downloader.Download(ctx, job.WhatsAppURL, job.MediaKey, ...)
    if err != nil {
        return p.handleDownloadError(ctx, job, err)
    }
    defer reader.Close()

    // 3. Upload to S3 (streaming)
    s3Key, presignedURL, err := p.uploader.Upload(ctx, job.InstanceID,
        job.EventID.String(), job.MediaType, reader, job.MimeType, job.FileLength)
    if err != nil {
        return p.handleUploadError(ctx, job, err)
    }

    // 4. Store metadata for deduplication
    metadata := &models.MediaMetadata{
        InstanceID: job.InstanceID,
        FileSha256: job.FileSha256,
        S3Bucket:   p.uploader.bucket,
        S3Key:      s3Key,
        FileSize:   job.FileLength,
        MimeType:   job.MimeType,
        MediaType:  job.MediaType,
    }
    if err := p.metadataRepo.Create(ctx, metadata); err != nil {
        // Non-fatal, log warning
    }

    // 5. Update job status
    job.Status = "completed"
    job.S3Key = s3Key
    job.S3URL = presignedURL
    job.ProcessedAt = time.Now()
    return p.jobsRepo.Update(ctx, job)
}
```

**Deduplication Logic**:
```go
// Check if media already exists by file_sha256
metadata, err := metadataRepo.FindByFileHash(ctx, instanceID, fileSha256)
if err == nil {
    // Generate fresh presigned URL
    presignedURL, _ := uploader.GeneratePresignedURL(ctx, metadata.S3Key, expiry)
    // Update access tracking
    metadataRepo.IncrementAccessCount(ctx, metadata.ID)
    // Metrics
    metrics.media_deduplication_hits_total.Inc()
    return presignedURL
}
```

**Metrics**:
```go
media_processing_duration_seconds{media_type, status}
media_deduplication_hits_total
media_deduplication_misses_total
media_processing_errors_total{error_type}
```

**Acceptance Criteria**:
- [ ] Complete flow: download → upload → URL
- [ ] Deduplication works correctly
- [ ] Error handling for all failure modes
- [ ] Streaming prevents memory issues
- [ ] Metrics exported
- [ ] Integration tests with mock S3 + WhatsApp

#### Step 6.6: Media Repositories (3 hours)

**Files**:
- `api/internal/events/persistence/media_jobs.go`
- `api/internal/events/persistence/media_metadata.go`

**MediaJobsRepository**:
```go
type MediaJobsRepository struct {
    db *pgxpool.Pool
}

func (r *MediaJobsRepository) Create(ctx context.Context, job *models.MediaJob) error
func (r *MediaJobsRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.MediaJob, error)
func (r *MediaJobsRepository) FindPendingByInstance(ctx context.Context, instanceID uuid.UUID, limit int) ([]*models.MediaJob, error)
func (r *MediaJobsRepository) Update(ctx context.Context, job *models.MediaJob) error
func (r *MediaJobsRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error
func (r *MediaJobsRepository) IncrementRetryCount(ctx context.Context, id uuid.UUID) error
func (r *MediaJobsRepository) FindFailedForRetry(ctx context.Context, limit int) ([]*models.MediaJob, error)
```

**MediaMetadataRepository**:
```go
type MediaMetadataRepository struct {
    db *pgxpool.Pool
}

func (r *MediaMetadataRepository) Create(ctx context.Context, metadata *models.MediaMetadata) error
func (r *MediaMetadataRepository) FindByFileHash(ctx context.Context, instanceID uuid.UUID, fileSha256 []byte) (*models.MediaMetadata, error)
func (r *MediaMetadataRepository) FindByS3Key(ctx context.Context, s3Key string) (*models.MediaMetadata, error)
func (r *MediaMetadataRepository) IncrementAccessCount(ctx context.Context, id uuid.UUID) error
func (r *MediaMetadataRepository) FindUnused(ctx context.Context, inactiveDays int, limit int) ([]*models.MediaMetadata, error)
func (r *MediaMetadataRepository) Delete(ctx context.Context, id uuid.UUID) error
```

**Acceptance Criteria**:
- [ ] All CRUD operations work
- [ ] Transactions used where appropriate
- [ ] Queries use indexes (verify with EXPLAIN)
- [ ] Error handling for deadlocks, FK violations
- [ ] Unit tests with test database

#### Step 6.7: Media Worker (4 hours)

**Files**:
- `api/internal/events/media/worker.go`

**Implementation** (mirrors InstanceWorker pattern):
```go
type MediaWorker struct {
    instanceID uuid.UUID
    processor  *MediaProcessor
    jobsRepo   *MediaJobsRepository
    pollInterval time.Duration
    batchSize    int
    metrics      *observability.Metrics
    logger       *slog.Logger
    stopChan     chan struct{}
    wg           sync.WaitGroup
}

func (w *MediaWorker) Start(ctx context.Context) error {
    w.logger.Info("starting media worker", "instance_id", w.instanceID)
    w.metrics.media_worker_starts_total.WithLabelValues(w.instanceID.String()).Inc()

    w.wg.Add(1)
    go w.run(ctx)
    return nil
}

func (w *MediaWorker) run(ctx context.Context) {
    defer w.wg.Done()
    ticker := time.NewTicker(w.pollInterval) // 100ms default
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            w.logger.Info("media worker context cancelled", "instance_id", w.instanceID)
            return
        case <-w.stopChan:
            w.logger.Info("media worker stop signal received", "instance_id", w.instanceID)
            return
        case <-ticker.C:
            w.processBatch(ctx)
        }
    }
}

func (w *MediaWorker) processBatch(ctx context.Context) {
    jobs, err := w.jobsRepo.FindPendingByInstance(ctx, w.instanceID, w.batchSize)
    if err != nil {
        w.logger.Error("failed to fetch media jobs", "error", err)
        w.metrics.media_worker_errors_total.WithLabelValues(w.instanceID.String(), "fetch").Inc()
        return
    }

    if len(jobs) == 0 {
        return // No work to do
    }

    w.logger.Debug("processing media batch", "instance_id", w.instanceID, "count", len(jobs))

    for _, job := range jobs {
        if err := w.processor.ProcessMediaJob(ctx, job); err != nil {
            w.logger.Error("failed to process media job",
                "job_id", job.ID,
                "error", err)
            w.metrics.media_worker_jobs_processed_total.WithLabelValues(
                w.instanceID.String(), "failure").Inc()
        } else {
            w.metrics.media_worker_jobs_processed_total.WithLabelValues(
                w.instanceID.String(), "success").Inc()
        }
    }
}

func (w *MediaWorker) Stop(ctx context.Context) error {
    w.logger.Info("stopping media worker", "instance_id", w.instanceID)
    close(w.stopChan)

    // Wait for graceful shutdown with timeout
    done := make(chan struct{})
    go func() {
        w.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        w.logger.Info("media worker stopped gracefully", "instance_id", w.instanceID)
        return nil
    case <-ctx.Done():
        w.logger.Warn("media worker shutdown timeout", "instance_id", w.instanceID)
        return ctx.Err()
    }
}
```

**Metrics**:
```go
media_worker_starts_total{instance_id}
media_worker_stops_total{instance_id}
media_worker_polls_total{instance_id}
media_worker_jobs_processed_total{instance_id, status="success|failure"}
media_worker_errors_total{instance_id, error_type}
```

**Acceptance Criteria**:
- [ ] Polls media_jobs every 100ms
- [ ] Processes batch (default 10 jobs)
- [ ] Graceful shutdown within 30s
- [ ] Metrics exported
- [ ] Logs structured with instance_id
- [ ] Integration tests

#### Step 6.8: Media Coordinator (4 hours)

**Files**:
- `api/internal/events/media/coordinator.go`

**Implementation** (mirrors DispatchCoordinator pattern):
```go
type MediaCoordinator struct {
    cfg          *config.MediaConfig
    db           *pgxpool.Pool
    jobsRepo     *MediaJobsRepository
    metadataRepo *MediaMetadataRepository
    processor    *MediaProcessor
    metrics      *observability.Metrics
    logger       *slog.Logger

    workers      map[uuid.UUID]*MediaWorker
    workersMu    sync.RWMutex
    stopChan     chan struct{}
    wg           sync.WaitGroup
}

func NewMediaCoordinator(
    cfg *config.MediaConfig,
    db *pgxpool.Pool,
    downloader *MediaDownloader,
    uploader *S3Uploader,
    metrics *observability.Metrics,
) *MediaCoordinator

func (c *MediaCoordinator) Start(ctx context.Context) error {
    if !c.cfg.Enabled {
        c.logger.Info("media processing disabled")
        return nil
    }

    c.logger.Info("starting media coordinator")
    c.metrics.media_coordinator_starts_total.Inc()
    return nil
}

func (c *MediaCoordinator) RegisterInstance(ctx context.Context, instanceID uuid.UUID) error {
    c.workersMu.Lock()
    defer c.workersMu.Unlock()

    if _, exists := c.workers[instanceID]; exists {
        return fmt.Errorf("worker already registered for instance %s", instanceID)
    }

    worker := NewMediaWorker(instanceID, c.processor, c.jobsRepo, c.metrics)
    if err := worker.Start(ctx); err != nil {
        return fmt.Errorf("failed to start media worker: %w", err)
    }

    c.workers[instanceID] = worker
    c.metrics.media_coordinator_workers_active.Set(float64(len(c.workers)))
    c.logger.Info("media worker registered", "instance_id", instanceID)
    return nil
}

func (c *MediaCoordinator) UnregisterInstance(ctx context.Context, instanceID uuid.UUID) error {
    c.workersMu.Lock()
    defer c.workersMu.Unlock()

    worker, exists := c.workers[instanceID]
    if !exists {
        return fmt.Errorf("no worker found for instance %s", instanceID)
    }

    shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    if err := worker.Stop(shutdownCtx); err != nil {
        c.logger.Error("failed to stop media worker gracefully",
            "instance_id", instanceID, "error", err)
    }

    delete(c.workers, instanceID)
    c.metrics.media_coordinator_workers_active.Set(float64(len(c.workers)))
    c.logger.Info("media worker unregistered", "instance_id", instanceID)
    return nil
}

func (c *MediaCoordinator) Stop(ctx context.Context) error {
    c.logger.Info("stopping media coordinator")
    close(c.stopChan)

    c.workersMu.Lock()
    instanceIDs := make([]uuid.UUID, 0, len(c.workers))
    for id := range c.workers {
        instanceIDs = append(instanceIDs, id)
    }
    c.workersMu.Unlock()

    // Stop all workers concurrently
    var stopWg sync.WaitGroup
    for _, instanceID := range instanceIDs {
        stopWg.Add(1)
        go func(id uuid.UUID) {
            defer stopWg.Done()
            c.UnregisterInstance(ctx, id)
        }(instanceID)
    }

    stopWg.Wait()
    c.wg.Wait()

    c.logger.Info("media coordinator stopped")
    c.metrics.media_coordinator_stops_total.Inc()
    return nil
}
```

**Metrics**:
```go
media_coordinator_starts_total
media_coordinator_stops_total
media_coordinator_workers_active gauge
media_coordinator_register_total{status="success|failure"}
media_coordinator_unregister_total{status="success|failure"}
```

**Acceptance Criteria**:
- [ ] Registers workers on instance connect
- [ ] Unregisters workers on instance disconnect
- [ ] Tracks active worker count
- [ ] Graceful shutdown all workers
- [ ] Thread-safe worker map access
- [ ] Integration with ClientRegistry events

#### Step 6.9: EventProcessor Integration (3 hours) ⚠️ CRITICAL

**Files**:
- `api/internal/events/dispatch/processor.go` (modify existing)
- `api/internal/events/transform/target.go` (modify existing)

**Modifications to EventProcessor**:
```go
type EventProcessor struct {
    // ... existing fields
    mediaJobsRepo *persistence.MediaJobsRepository
}

func (p *EventProcessor) ProcessEvent(ctx context.Context, event *models.WebhookEvent) error {
    logger := logging.ContextLogger(ctx, nil).With("event_id", event.ID)

    // 1. Transform event
    webhookPayload, err := p.transformPipeline.TransformToWebhook(ctx, event)
    if err != nil {
        return p.handleTransformError(ctx, event, err)
    }

    // 2. Check for media
    mediaInfo := p.extractMediaInfo(webhookPayload)
    if mediaInfo != nil {
        // Check if media already processed (deduplication)
        if existingURL, err := p.checkMediaCache(ctx, event.InstanceID, mediaInfo.FileSha256); err == nil {
            // Use existing URL
            webhookPayload.MediaURL = existingURL
        } else {
            // Create media job for async processing
            if err := p.createMediaJob(ctx, event, mediaInfo); err != nil {
                logger.Error("failed to create media job", "error", err)
                return err
            }

            // Update outbox status to pending_media
            if err := p.outboxRepo.UpdateStatus(ctx, event.ID, "pending_media"); err != nil {
                return err
            }

            logger.Info("media job created, event marked pending_media")
            return nil // Don't deliver yet, wait for media
        }
    }

    // 3. Deliver webhook (existing logic)
    return p.deliverWebhook(ctx, event, webhookPayload)
}

func (p *EventProcessor) createMediaJob(ctx context.Context, event *models.WebhookEvent, media *MediaInfo) error {
    job := &models.MediaJob{
        InstanceID:      event.InstanceID,
        EventID:         event.ID,
        MediaType:       media.Type,
        WhatsAppURL:     media.URL,
        MediaKey:        media.MediaKey,
        FileEncSha256:   media.FileEncSha256,
        FileSha256:      media.FileSha256,
        FileLength:      media.FileLength,
        MimeType:        media.MimeType,
        Status:          "pending",
        MaxRetries:      3,
    }
    return p.mediaJobsRepo.Create(ctx, job)
}
```

**Modifications to InstanceWorker**:
```go
func (w *InstanceWorker) fetchPendingEvents(ctx context.Context) ([]*models.WebhookEvent, error) {
    // Skip events with status 'pending_media' (waiting for media processing)
    events, err := w.outboxRepo.FindPendingByInstance(ctx, w.instanceID, w.batchSize)
    // Filter out pending_media events
    filtered := make([]*models.WebhookEvent, 0, len(events))
    for _, evt := range events {
        if evt.Status != "pending_media" {
            filtered = append(filtered, evt)
        }
    }
    return filtered, err
}
```

**MediaWorker callback** (update outbox when media ready):
```go
func (p *MediaProcessor) ProcessMediaJob(ctx context.Context, job *models.MediaJob) error {
    // ... process media (existing logic)

    // After successful processing, update outbox status back to 'pending'
    if err := p.outboxRepo.UpdateStatus(ctx, job.EventID, "pending"); err != nil {
        return fmt.Errorf("failed to update outbox status: %w", err)
    }

    // Update event with S3 URL
    if err := p.outboxRepo.UpdateMediaURL(ctx, job.EventID, job.S3URL); err != nil {
        return fmt.Errorf("failed to update media URL: %w", err)
    }

    return nil
}
```

**Acceptance Criteria**:
- [ ] Events with media create media_jobs
- [ ] Outbox status set to 'pending_media' correctly
- [ ] InstanceWorker skips 'pending_media' events
- [ ] MediaWorker updates outbox when done
- [ ] Events delivered with S3 URL
- [ ] End-to-end test: message with media → webhook

#### Step 6.10: Testing & Validation (6 hours) ⚠️ CRITICAL

**Unit Tests** (3 hours):
```
api/internal/events/media/
├── downloader_test.go      // Mock WhatsApp client
├── uploader_test.go        // Mock S3 client (MinIO)
├── processor_test.go       // Mock dependencies
├── worker_test.go          // Mock repositories
└── coordinator_test.go     // Mock workers
```

**Integration Tests** (2 hours):
```
api/tests/integration/
└── media_processing_test.go
    - Test end-to-end: WhatsApp message → S3 → webhook
    - Test deduplication: same media twice
    - Test retry logic: failed download
    - Test worker lifecycle: start/stop/failover
```

**Performance Tests** (1 hour):
```
api/tests/performance/
└── media_load_test.go
    - 1000 concurrent media uploads
    - 50MB file upload
    - Worker pool scaling
```

**Test Fixtures**:
- Mock WhatsApp media URLs with valid encryption
- MinIO server for S3 testing
- Test database with media_jobs and media_metadata tables

**Acceptance Criteria**:
- [ ] >80% code coverage for media package
- [ ] All unit tests pass
- [ ] Integration tests cover all critical paths
- [ ] Performance tests validate latency targets
- [ ] Load test: 1000 media uploads succeed
- [ ] Documentation updated (CLAUDE.md, AGENTS.md)

---

## Phase 7: Background Jobs (32 hours)

### Goals
- Implement background job scheduler for system maintenance
- Recover failed webhook deliveries (DLQ reprocessing)
- Retry failed media processing
- Clean up old data (outbox, media)
- Monitor instance health and release stale locks
- Provide operational visibility and alerting

### Architecture Components

```
internal/jobs/
├── scheduler.go          // JobScheduler - manages all jobs (4h)
├── dlq_reprocessor.go    // DLQ → Outbox retry (4h)
├── media_retry.go        // Media job retry (3h)
├── outbox_cleanup.go     // Delete old events (3h)
├── media_cleanup.go      // Delete unused media (4h)
├── health_monitor.go     // Instance health checks (3h)
└── metrics_export.go     // Aggregate metrics (optional, 3h)
```

### Implementation Steps

#### Step 7.1: Jobs Configuration (1 hour) ⚠️ CRITICAL

**Files**:
- `api/internal/config/config.go`

**Add JobsConfig**:
```go
type Config struct {
    // ... existing fields
    Jobs JobsConfig
}

type JobsConfig struct {
    Enabled             bool `env:"JOBS_ENABLED" envDefault:"true"`
    DLQReprocessor      JobConfig
    MediaRetry          JobConfig
    OutboxCleanup       JobConfig
    MediaCleanup        JobConfig
    HealthMonitor       JobConfig
}

type JobConfig struct {
    Enabled         bool          `env:"-"`
    Schedule        string        // Cron expression
    BatchSize       int
    RetentionDays   int           // For cleanup jobs
    Timeout         time.Duration
}

func LoadJobsConfig() JobsConfig {
    return JobsConfig{
        Enabled: getEnvBool("JOBS_ENABLED", true),
        DLQReprocessor: JobConfig{
            Enabled:   getEnvBool("JOB_DLQ_ENABLED", true),
            Schedule:  getEnv("JOB_DLQ_SCHEDULE", "*/5 * * * *"), // Every 5 min
            BatchSize: getEnvInt("JOB_DLQ_BATCH_SIZE", 100),
            Timeout:   getEnvDuration("JOB_DLQ_TIMEOUT", 5*time.Minute),
        },
        MediaRetry: JobConfig{
            Enabled:   getEnvBool("JOB_MEDIA_RETRY_ENABLED", true),
            Schedule:  getEnv("JOB_MEDIA_RETRY_SCHEDULE", "*/10 * * * *"), // Every 10 min
            BatchSize: getEnvInt("JOB_MEDIA_RETRY_BATCH_SIZE", 50),
            Timeout:   getEnvDuration("JOB_MEDIA_RETRY_TIMEOUT", 10*time.Minute),
        },
        OutboxCleanup: JobConfig{
            Enabled:       getEnvBool("JOB_OUTBOX_CLEANUP_ENABLED", true),
            Schedule:      getEnv("JOB_OUTBOX_CLEANUP_SCHEDULE", "0 2 * * *"), // 2 AM daily
            BatchSize:     getEnvInt("JOB_OUTBOX_CLEANUP_BATCH_SIZE", 1000),
            RetentionDays: getEnvInt("JOB_OUTBOX_RETENTION_DAYS", 7),
            Timeout:       getEnvDuration("JOB_OUTBOX_CLEANUP_TIMEOUT", 30*time.Minute),
        },
        MediaCleanup: JobConfig{
            Enabled:       getEnvBool("JOB_MEDIA_CLEANUP_ENABLED", true),
            Schedule:      getEnv("JOB_MEDIA_CLEANUP_SCHEDULE", "0 3 * * *"), // 3 AM daily
            BatchSize:     getEnvInt("JOB_MEDIA_CLEANUP_BATCH_SIZE", 100),
            RetentionDays: getEnvInt("JOB_MEDIA_RETENTION_DAYS", 30),
            Timeout:       getEnvDuration("JOB_MEDIA_CLEANUP_TIMEOUT", 1*time.Hour),
        },
        HealthMonitor: JobConfig{
            Enabled:   getEnvBool("JOB_HEALTH_MONITOR_ENABLED", true),
            Schedule:  getEnv("JOB_HEALTH_MONITOR_SCHEDULE", "*/10 * * * *"), // Every 10 min
            BatchSize: getEnvInt("JOB_HEALTH_MONITOR_BATCH_SIZE", 100),
            Timeout:   getEnvDuration("JOB_HEALTH_MONITOR_TIMEOUT", 2*time.Minute),
        },
    }
}
```

**Environment Variables**:
```bash
JOBS_ENABLED=true

# DLQ Reprocessor
JOB_DLQ_ENABLED=true
JOB_DLQ_SCHEDULE="*/5 * * * *"  # Every 5 minutes
JOB_DLQ_BATCH_SIZE=100
JOB_DLQ_TIMEOUT=5m

# Media Retry
JOB_MEDIA_RETRY_ENABLED=true
JOB_MEDIA_RETRY_SCHEDULE="*/10 * * * *"  # Every 10 minutes
JOB_MEDIA_RETRY_BATCH_SIZE=50
JOB_MEDIA_RETRY_TIMEOUT=10m

# Outbox Cleanup
JOB_OUTBOX_CLEANUP_ENABLED=true
JOB_OUTBOX_CLEANUP_SCHEDULE="0 2 * * *"  # 2 AM daily
JOB_OUTBOX_CLEANUP_BATCH_SIZE=1000
JOB_OUTBOX_RETENTION_DAYS=7
JOB_OUTBOX_CLEANUP_TIMEOUT=30m

# Media Cleanup
JOB_MEDIA_CLEANUP_ENABLED=true
JOB_MEDIA_CLEANUP_SCHEDULE="0 3 * * *"  # 3 AM daily
JOB_MEDIA_CLEANUP_BATCH_SIZE=100
JOB_MEDIA_RETENTION_DAYS=30
JOB_MEDIA_CLEANUP_TIMEOUT=1h

# Health Monitor
JOB_HEALTH_MONITOR_ENABLED=true
JOB_HEALTH_MONITOR_SCHEDULE="*/10 * * * *"  # Every 10 minutes
JOB_HEALTH_MONITOR_BATCH_SIZE=100
JOB_HEALTH_MONITOR_TIMEOUT=2m
```

**Acceptance Criteria**:
- [ ] All job configs loaded from env vars
- [ ] Cron expressions validated on startup
- [ ] Default values reasonable
- [ ] Individual jobs can be disabled

#### Step 7.2: Job Scheduler Framework (4 hours) ⚠️ CRITICAL

**Files**:
- `api/internal/jobs/scheduler.go`

**Dependencies**:
```bash
go get github.com/go-co-op/gocron
```

**Implementation**:
```go
type JobScheduler struct {
    scheduler *gocron.Scheduler
    jobs      []Job
    metrics   *observability.Metrics
    logger    *slog.Logger
    stopChan  chan struct{}
}

type Job interface {
    Name() string
    Schedule() string
    Enabled() bool
    Execute(ctx context.Context) error
}

func NewJobScheduler(jobs []Job, metrics *observability.Metrics) *JobScheduler {
    return &JobScheduler{
        scheduler: gocron.NewScheduler(time.UTC),
        jobs:      jobs,
        metrics:   metrics,
        logger:    slog.Default(),
        stopChan:  make(chan struct{}),
    }
}

func (s *JobScheduler) Start(ctx context.Context) error {
    s.logger.Info("starting job scheduler", "job_count", len(s.jobs))

    for _, job := range s.jobs {
        if !job.Enabled() {
            s.logger.Info("job disabled, skipping", "job_name", job.Name())
            continue
        }

        _, err := s.scheduler.Cron(job.Schedule()).Do(func() {
            s.executeJob(ctx, job)
        })

        if err != nil {
            return fmt.Errorf("failed to schedule job %s: %w", job.Name(), err)
        }

        s.logger.Info("job scheduled", "job_name", job.Name(), "schedule", job.Schedule())
    }

    s.scheduler.StartAsync()
    s.logger.Info("job scheduler started")
    return nil
}

func (s *JobScheduler) executeJob(ctx context.Context, job Job) {
    logger := s.logger.With("job_name", job.Name())
    logger.Info("job execution started")

    start := time.Now()
    err := job.Execute(ctx)
    duration := time.Since(start)

    // Update metrics
    s.metrics.job_duration_seconds.WithLabelValues(job.Name()).Observe(duration.Seconds())

    if err != nil {
        logger.Error("job execution failed", "error", err, "duration", duration)
        s.metrics.job_executions_total.WithLabelValues(job.Name(), "failure").Inc()
        s.metrics.job_errors_total.WithLabelValues(job.Name(), classifyError(err)).Inc()
        // Capture in Sentry
        sentry.CaptureException(err)
    } else {
        logger.Info("job execution completed", "duration", duration)
        s.metrics.job_executions_total.WithLabelValues(job.Name(), "success").Inc()
    }
}

func (s *JobScheduler) Stop(ctx context.Context) error {
    s.logger.Info("stopping job scheduler")
    s.scheduler.Stop()
    close(s.stopChan)

    // Wait for all jobs to complete
    <-time.After(5 * time.Second)

    s.logger.Info("job scheduler stopped")
    return nil
}
```

**Metrics**:
```go
job_executions_total{job_name, status="success|failure"} counter
job_duration_seconds{job_name} histogram
job_errors_total{job_name, error_type} counter
job_last_execution_timestamp{job_name} gauge
```

**Acceptance Criteria**:
- [ ] Scheduler starts and stops gracefully
- [ ] Jobs execute on schedule
- [ ] Metrics exported for all jobs
- [ ] Errors logged and captured in Sentry
- [ ] Job interface easy to implement
- [ ] Cron expressions parsed correctly

#### Step 7.3: Health Monitor Job (3 hours)

**Files**:
- `api/internal/jobs/health_monitor.go`

**Purpose**: Detect stale instances, release Redis locks, update database status

**Implementation**:
```go
type HealthMonitorJob struct {
    cfg         config.JobConfig
    db          *pgxpool.Pool
    redis       *redis.Client
    instanceRepo *persistence.InstanceRepository
    metrics     *observability.Metrics
    logger      *slog.Logger
}

func (j *HealthMonitorJob) Execute(ctx context.Context) error {
    // 1. Find stale instances (no heartbeat >1 hour)
    staleInstances, err := j.instanceRepo.FindStale(ctx, 1*time.Hour)
    if err != nil {
        return fmt.Errorf("failed to find stale instances: %w", err)
    }

    if len(staleInstances) == 0 {
        j.logger.Debug("no stale instances found")
        return nil
    }

    j.logger.Info("stale instances detected", "count", len(staleInstances))
    j.metrics.stale_instances_detected_total.Add(float64(len(staleInstances)))

    // 2. Process each stale instance
    for _, instance := range staleInstances {
        if err := j.processStaleInstance(ctx, instance); err != nil {
            j.logger.Error("failed to process stale instance",
                "instance_id", instance.ID, "error", err)
            continue
        }
    }

    return nil
}

func (j *HealthMonitorJob) processStaleInstance(ctx context.Context, instance *models.Instance) error {
    logger := j.logger.With("instance_id", instance.ID)

    // 1. Release Redis lock
    lockKey := fmt.Sprintf("lock:instance:%s", instance.ID)
    if err := j.redis.Del(ctx, lockKey).Err(); err != nil {
        logger.Error("failed to release redis lock", "error", err)
    } else {
        logger.Info("redis lock released")
        j.metrics.locks_released_total.Inc()
    }

    // 2. Update instance status to 'disconnected'
    instance.Status = "disconnected"
    instance.DisconnectedAt = time.Now()
    if err := j.instanceRepo.Update(ctx, instance); err != nil {
        return fmt.Errorf("failed to update instance status: %w", err)
    }

    logger.Info("instance marked as disconnected")
    return nil
}
```

**Queries**:
```sql
-- Find stale instances
SELECT * FROM instances
WHERE status = 'connected'
  AND last_heartbeat < NOW() - INTERVAL '1 hour'
LIMIT ?;
```

**Metrics**:
```go
instances_checked_total counter
stale_instances_detected_total counter
locks_released_total counter
instance_status_updates_total{status} counter
```

**Acceptance Criteria**:
- [ ] Detects instances with no heartbeat >1 hour
- [ ] Releases Redis locks successfully
- [ ] Updates instance status to 'disconnected'
- [ ] Handles errors gracefully (continue processing)
- [ ] Metrics exported
- [ ] Integration test with stale instance

#### Step 7.4: DLQ Reprocessor Job (4 hours)

**Files**:
- `api/internal/jobs/dlq_reprocessor.go`

**Purpose**: Move retryable events from webhook_dlq back to webhook_outbox

**Implementation**:
```go
type DLQReprocessorJob struct {
    cfg         config.JobConfig
    db          *pgxpool.Pool
    dlqRepo     *persistence.DLQRepository
    outboxRepo  *persistence.OutboxRepository
    metrics     *observability.Metrics
    logger      *slog.Logger
}

func (j *DLQReprocessorJob) Execute(ctx context.Context) error {
    // 1. Find retryable DLQ events (retry_count < max_retries)
    dlqEvents, err := j.dlqRepo.FindRetryable(ctx, j.cfg.BatchSize)
    if err != nil {
        return fmt.Errorf("failed to fetch retryable DLQ events: %w", err)
    }

    if len(dlqEvents) == 0 {
        j.logger.Debug("no retryable DLQ events found")
        return nil
    }

    j.logger.Info("processing DLQ events", "count", len(dlqEvents))

    // 2. Process in batches with transaction
    for i := 0; i < len(dlqEvents); i += 50 {
        batch := dlqEvents[i:min(i+50, len(dlqEvents))]
        if err := j.processBatch(ctx, batch); err != nil {
            j.logger.Error("failed to process DLQ batch", "error", err)
            continue
        }
    }

    return nil
}

func (j *DLQReprocessorJob) processBatch(ctx context.Context, batch []*models.DLQEvent) error {
    tx, err := j.db.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    for _, dlqEvent := range batch {
        // Check if error is retryable
        if !j.isRetryable(dlqEvent) {
            j.logger.Debug("dlq event not retryable", "dlq_id", dlqEvent.ID)
            continue
        }

        // Move back to outbox
        outboxEvent := &models.WebhookEvent{
            ID:          dlqEvent.OriginalEventID,
            InstanceID:  dlqEvent.InstanceID,
            EventType:   dlqEvent.EventType,
            Payload:     dlqEvent.Payload,
            Status:      "pending",
            RetryCount:  dlqEvent.RetryCount + 1,
            MaxRetries:  dlqEvent.MaxRetries,
            CreatedAt:   dlqEvent.OriginalCreatedAt,
        }

        if err := j.outboxRepo.CreateTx(ctx, tx, outboxEvent); err != nil {
            j.logger.Error("failed to create outbox event", "error", err)
            continue
        }

        // Delete from DLQ
        if err := j.dlqRepo.DeleteTx(ctx, tx, dlqEvent.ID); err != nil {
            j.logger.Error("failed to delete DLQ event", "error", err)
            continue
        }

        j.metrics.dlq_events_moved_back_total.Inc()
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    j.metrics.dlq_events_reprocessed_total.WithLabelValues("success").Add(float64(len(batch)))
    return nil
}

func (j *DLQReprocessorJob) isRetryable(dlqEvent *models.DLQEvent) bool {
    // Check retry count
    if dlqEvent.RetryCount >= dlqEvent.MaxRetries {
        return false
    }

    // Check error type
    retryableErrors := []string{
        "network_timeout",
        "connection_refused",
        "temporary_failure",
    }

    for _, retryable := range retryableErrors {
        if strings.Contains(dlqEvent.ErrorType, retryable) {
            return true
        }
    }

    return false
}
```

**Queries**:
```sql
-- Find retryable DLQ events
SELECT * FROM webhook_dlq
WHERE retry_count < max_retries
  AND error_type IN ('network_timeout', 'connection_refused', 'temporary_failure')
  AND moved_to_dlq_at < NOW() - INTERVAL '5 minutes'  -- Cooldown
ORDER BY moved_to_dlq_at ASC
LIMIT ?;
```

**Metrics**:
```go
dlq_events_reprocessed_total{status="success|failure"} counter
dlq_events_moved_back_total counter
dlq_batch_processing_duration_seconds histogram
```

**Acceptance Criteria**:
- [ ] Moves retryable events back to outbox
- [ ] Respects retry limits
- [ ] Transaction safety (atomic move)
- [ ] Batch processing efficient
- [ ] Metrics exported
- [ ] Integration test with DLQ events

#### Step 7.5: Media Retry Job (3 hours)

**Files**:
- `api/internal/jobs/media_retry.go`

**Purpose**: Retry failed media processing jobs

**Implementation**:
```go
type MediaRetryJob struct {
    cfg          config.JobConfig
    db           *pgxpool.Pool
    jobsRepo     *persistence.MediaJobsRepository
    metrics      *observability.Metrics
    logger       *slog.Logger
}

func (j *MediaRetryJob) Execute(ctx context.Context) error {
    // Find failed media jobs eligible for retry
    failedJobs, err := j.jobsRepo.FindFailedForRetry(ctx, j.cfg.BatchSize)
    if err != nil {
        return fmt.Errorf("failed to fetch failed media jobs: %w", err)
    }

    if len(failedJobs) == 0 {
        j.logger.Debug("no failed media jobs found")
        return nil
    }

    j.logger.Info("retrying failed media jobs", "count", len(failedJobs))

    successCount := 0
    for _, job := range failedJobs {
        if job.RetryCount >= job.MaxRetries {
            j.logger.Warn("media job exceeded max retries, moving to permanent failure",
                "job_id", job.ID, "retry_count", job.RetryCount)
            continue
        }

        // Reset status to 'pending' for MediaWorker to pick up
        job.Status = "pending"
        job.RetryCount++

        if err := j.jobsRepo.Update(ctx, job); err != nil {
            j.logger.Error("failed to reset media job", "job_id", job.ID, "error", err)
            continue
        }

        successCount++
        j.metrics.media_jobs_retried_total.Inc()
    }

    j.logger.Info("media jobs reset for retry", "count", successCount)
    return nil
}
```

**Queries**:
```sql
-- Find failed media jobs for retry
SELECT * FROM media_jobs
WHERE status = 'failed'
  AND retry_count < max_retries
  AND updated_at < NOW() - INTERVAL '10 minutes'  -- Cooldown
ORDER BY created_at ASC
LIMIT ?;
```

**Metrics**:
```go
media_jobs_retried_total counter
media_jobs_permanent_failure_total counter
```

**Acceptance Criteria**:
- [ ] Resets failed jobs to 'pending'
- [ ] Respects max retry limits
- [ ] Cooldown period enforced
- [ ] Permanent failures logged
- [ ] Metrics exported
- [ ] Integration test with failed media jobs

#### Step 7.6: Outbox Cleanup Job (3 hours)

**Files**:
- `api/internal/jobs/outbox_cleanup.go`

**Purpose**: Delete old completed/failed events from webhook_outbox

**Implementation**:
```go
type OutboxCleanupJob struct {
    cfg          config.JobConfig
    db           *pgxpool.Pool
    outboxRepo   *persistence.OutboxRepository
    metrics      *observability.Metrics
    logger       *slog.Logger
}

func (j *OutboxCleanupJob) Execute(ctx context.Context) error {
    // 1. Delete completed events older than retention period
    completedDeleted, err := j.deleteOldEvents(ctx, "completed", j.cfg.RetentionDays)
    if err != nil {
        j.logger.Error("failed to delete completed events", "error", err)
    } else {
        j.logger.Info("deleted completed events", "count", completedDeleted)
        j.metrics.outbox_events_deleted_total.WithLabelValues("completed").Add(float64(completedDeleted))
    }

    // 2. Delete failed events older than 30 days (longer retention)
    failedDeleted, err := j.deleteOldEvents(ctx, "failed", 30)
    if err != nil {
        j.logger.Error("failed to delete failed events", "error", err)
    } else {
        j.logger.Info("deleted failed events", "count", failedDeleted)
        j.metrics.outbox_events_deleted_total.WithLabelValues("failed").Add(float64(failedDeleted))
    }

    // 3. Vacuum analyze table
    if err := j.vacuumAnalyze(ctx); err != nil {
        j.logger.Error("failed to vacuum analyze", "error", err)
    }

    return nil
}

func (j *OutboxCleanupJob) deleteOldEvents(ctx context.Context, status string, retentionDays int) (int64, error) {
    cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

    // Delete in batches to avoid long locks
    totalDeleted := int64(0)
    for {
        result, err := j.db.Exec(ctx, `
            DELETE FROM webhook_outbox
            WHERE id IN (
                SELECT id FROM webhook_outbox
                WHERE status = $1 AND created_at < $2
                ORDER BY created_at ASC
                LIMIT $3
            )
        `, status, cutoffDate, j.cfg.BatchSize)

        if err != nil {
            return totalDeleted, err
        }

        deleted := result.RowsAffected()
        totalDeleted += deleted

        if deleted < int64(j.cfg.BatchSize) {
            break // No more rows to delete
        }

        // Short pause between batches
        time.Sleep(100 * time.Millisecond)
    }

    return totalDeleted, nil
}

func (j *OutboxCleanupJob) vacuumAnalyze(ctx context.Context) error {
    _, err := j.db.Exec(ctx, "VACUUM ANALYZE webhook_outbox")
    return err
}
```

**Queries**:
```sql
-- Delete old completed events
DELETE FROM webhook_outbox
WHERE status = 'completed'
  AND created_at < NOW() - INTERVAL '7 days'
LIMIT ?;

-- Delete old failed events
DELETE FROM webhook_outbox
WHERE status = 'failed'
  AND created_at < NOW() - INTERVAL '30 days'
LIMIT ?;
```

**Metrics**:
```go
outbox_events_deleted_total{status="completed|failed"} counter
outbox_cleanup_duration_seconds histogram
```

**Acceptance Criteria**:
- [ ] Deletes events respecting retention
- [ ] Batch deletion prevents long locks
- [ ] Vacuum analyze runs successfully
- [ ] Metrics exported
- [ ] Integration test with old events

#### Step 7.7: Media Cleanup Job (4 hours)

**Files**:
- `api/internal/jobs/media_cleanup.go`

**Purpose**: Delete unused media from S3 and media_metadata

**Implementation**:
```go
type MediaCleanupJob struct {
    cfg          config.JobConfig
    db           *pgxpool.Pool
    metadataRepo *persistence.MediaMetadataRepository
    uploader     *media.S3Uploader
    metrics      *observability.Metrics
    logger       *slog.Logger
}

func (j *MediaCleanupJob) Execute(ctx context.Context) error {
    // Find unused media (last accessed >30 days, low access count)
    unusedMedia, err := j.metadataRepo.FindUnused(ctx, j.cfg.RetentionDays, j.cfg.BatchSize)
    if err != nil {
        return fmt.Errorf("failed to find unused media: %w", err)
    }

    if len(unusedMedia) == 0 {
        j.logger.Debug("no unused media found")
        return nil
    }

    j.logger.Info("cleaning up unused media", "count", len(unusedMedia))

    totalFreed := int64(0)
    successCount := 0

    for _, metadata := range unusedMedia {
        // 1. Delete from S3
        if err := j.uploader.Delete(ctx, metadata.S3Key); err != nil {
            // Log error but continue (file might already be deleted)
            j.logger.Warn("failed to delete S3 object",
                "s3_key", metadata.S3Key, "error", err)
        }

        // 2. Delete from database
        if err := j.metadataRepo.Delete(ctx, metadata.ID); err != nil {
            j.logger.Error("failed to delete media metadata",
                "metadata_id", metadata.ID, "error", err)
            continue
        }

        totalFreed += metadata.FileSize
        successCount++
    }

    j.metrics.media_files_deleted_total.Add(float64(successCount))
    j.metrics.media_storage_freed_bytes.Add(float64(totalFreed))

    j.logger.Info("media cleanup completed",
        "files_deleted", successCount,
        "storage_freed_mb", totalFreed/1024/1024)

    return nil
}
```

**Queries**:
```sql
-- Find unused media
SELECT * FROM media_metadata
WHERE last_accessed_at < NOW() - INTERVAL '30 days'
  AND access_count < 5
ORDER BY last_accessed_at ASC
LIMIT ?;
```

**Metrics**:
```go
media_files_deleted_total counter
media_storage_freed_bytes counter
media_cleanup_duration_seconds histogram
```

**Acceptance Criteria**:
- [ ] Identifies unused media correctly
- [ ] Deletes from S3 successfully
- [ ] Deletes from database successfully
- [ ] Tracks storage freed
- [ ] Handles S3 deletion failures gracefully
- [ ] Integration test with unused media

#### Step 7.8: Metrics Export Job (3 hours) - Optional

**Files**:
- `api/internal/jobs/metrics_export.go`

**Purpose**: Aggregate and export metrics to external systems

**Implementation**:
```go
type MetricsExportJob struct {
    cfg          config.JobConfig
    db           *pgxpool.Pool
    metrics      *observability.Metrics
    logger       *slog.Logger
}

func (j *MetricsExportJob) Execute(ctx context.Context) error {
    // 1. Aggregate instance health metrics
    healthMetrics, err := j.aggregateInstanceHealth(ctx)
    if err != nil {
        return fmt.Errorf("failed to aggregate instance health: %w", err)
    }

    // 2. Aggregate webhook delivery stats
    deliveryStats, err := j.aggregateWebhookStats(ctx)
    if err != nil {
        return fmt.Errorf("failed to aggregate webhook stats: %w", err)
    }

    // 3. Export to external system (example: log structured JSON)
    j.logger.Info("metrics export",
        "instance_health", healthMetrics,
        "webhook_stats", deliveryStats)

    return nil
}
```

**Note**: This job is optional and implementation depends on external monitoring system requirements.

**Acceptance Criteria**:
- [ ] Aggregates key metrics
- [ ] Exports to configured destination
- [ ] Runs without errors

#### Step 7.9: Main Integration (2 hours) ⚠️ CRITICAL

**Files**:
- `api/cmd/server/main.go`

**Implementation**:
```go
func main() {
    // ... existing setup (database, redis, config, etc.)

    // Initialize background jobs
    if cfg.Jobs.Enabled {
        jobs := []jobs.Job{
            jobs.NewHealthMonitorJob(cfg.Jobs.HealthMonitor, pgPool, redisClient,
                instanceRepo, metrics),
            jobs.NewDLQReprocessorJob(cfg.Jobs.DLQReprocessor, pgPool, dlqRepo,
                outboxRepo, metrics),
            jobs.NewMediaRetryJob(cfg.Jobs.MediaRetry, pgPool, mediaJobsRepo, metrics),
            jobs.NewOutboxCleanupJob(cfg.Jobs.OutboxCleanup, pgPool, outboxRepo, metrics),
            jobs.NewMediaCleanupJob(cfg.Jobs.MediaCleanup, pgPool, metadataRepo,
                uploader, metrics),
        }

        jobScheduler := jobs.NewJobScheduler(jobs, metrics)
        if err := jobScheduler.Start(ctx); err != nil {
            log.Fatal("failed to start job scheduler", "error", err)
        }

        // Register shutdown
        shutdownFuncs = append(shutdownFuncs, func(ctx context.Context) error {
            return jobScheduler.Stop(ctx)
        })

        log.Info("background jobs started")
    }

    // ... rest of main (HTTP server, signal handling, etc.)
}
```

**Shutdown Order**:
1. HTTP server (stop accepting new requests)
2. DispatchCoordinator (stop webhook delivery)
3. MediaCoordinator (stop media processing)
4. JobScheduler (stop background jobs)
5. ClientRegistry (disconnect WhatsApp clients)
6. Database connections

**Acceptance Criteria**:
- [ ] Jobs start on application startup
- [ ] Jobs respect enabled flags
- [ ] Graceful shutdown coordinated
- [ ] Logs confirm job scheduler status
- [ ] All jobs executing on schedule

#### Step 7.10: Testing & Documentation (5 hours) ⚠️ CRITICAL

**Unit Tests** (2 hours):
```
api/internal/jobs/
├── scheduler_test.go
├── health_monitor_test.go
├── dlq_reprocessor_test.go
├── media_retry_test.go
├── outbox_cleanup_test.go
└── media_cleanup_test.go
```

**Integration Tests** (2 hours):
```
api/tests/integration/
└── background_jobs_test.go
    - Test job scheduler lifecycle
    - Test DLQ reprocessing end-to-end
    - Test media retry flow
    - Test cleanup jobs with old data
```

**Documentation Updates** (1 hour):
- Update PLAN.md with Phase 7 completion
- Update CLAUDE.md with background jobs section
- Update AGENTS.md with jobs observability patterns
- Create RUNBOOKS.md with operational procedures

**Acceptance Criteria**:
- [ ] >80% code coverage for jobs package
- [ ] All unit tests pass
- [ ] Integration tests cover critical paths
- [ ] Documentation complete and accurate
- [ ] Runbooks validated by operations team

---

## Integration & Dependencies

### Phase 6 ↔ Existing System Integration

**EventProcessor Integration**:
- EventProcessor detects media → creates media_job
- Outbox status set to 'pending_media'
- MediaWorker processes job → updates outbox to 'pending'
- InstanceWorker delivers webhook with S3 URL

**ClientRegistry Integration**:
- Instance Connected → MediaCoordinator.RegisterInstance()
- Instance LoggedOut → MediaCoordinator.UnregisterInstance()
- Worker lifecycle tied to WhatsApp connection

**Database Integration**:
- Foreign keys: media_jobs → instances, media_jobs → webhook_outbox
- Status state machine: pending → processing → completed/failed
- Cascade deletion: instance deleted → media jobs deleted

### Phase 7 ↔ Phase 6 Integration

**Media Retry Job**:
- Queries media_jobs with status='failed'
- Resets to 'pending' for MediaWorker
- Respects retry count limits

**Media Cleanup Job**:
- Queries media_metadata for unused files
- Deletes from S3 via MediaUploader
- Cleans up database records

**DLQ Reprocessor**:
- Handles webhook delivery failures
- Independent of media processing
- Can trigger events with pending media

### External Dependencies

**AWS SDK v2**:
```bash
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/feature/s3/manager
```

**gocron Library**:
```bash
go get github.com/go-co-op/gocron
```

**License Compatibility**:
- AWS SDK v2: Apache 2.0 ✅
- gocron: MIT ✅

---

## Deployment Plan

### Pre-Deployment Checklist

**Infrastructure** (Week before deployment):
- [ ] S3 bucket created: `whatsapp-media-production`
- [ ] IAM role configured with S3 access (GetObject, PutObject, DeleteObject)
- [ ] S3 lifecycle policy configured (optional: transition to Glacier after 90 days)
- [ ] MinIO deployed for development environment
- [ ] Database migrations tested in staging
- [ ] Redis cluster validated (distributed locking)
- [ ] Prometheus/Grafana dashboards created
- [ ] PagerDuty/Opsgenie alert routing configured
- [ ] Sentry project configured for error tracking

**Configuration** (3 days before deployment):
- [ ] All environment variables documented
- [ ] Secrets stored in AWS Secrets Manager/Vault
- [ ] Feature flags configured (MEDIA_PROCESSING_ENABLED=false, JOBS_ENABLED=false)
- [ ] Config validation tests pass in staging
- [ ] Backup/restore procedures documented

**Testing** (Week before deployment):
- [ ] Unit tests: >80% coverage ✅
- [ ] Integration tests: All critical paths ✅
- [ ] Load tests: 1000 concurrent media uploads succeed
- [ ] Chaos tests: Worker failures, network issues
- [ ] Security tests: S3 access controls, presigned URL validation
- [ ] Performance benchmarks: Media latency <30s p99

### Rollout Strategy: Blue-Green Deployment

**Phase 1: Dark Launch (Week 1)**

Goal: Deploy code without activating features

Actions:
- Deploy Phase 6 & 7 code to production
- Keep feature flags OFF: `MEDIA_PROCESSING_ENABLED=false`, `JOBS_ENABLED=false`
- Monitor resource usage (CPU, memory, database connections)
- Validate worker lifecycle (coordinator start/stop)
- Test with internal test instances only

Success Criteria:
- [ ] No degradation in existing functionality
- [ ] Resource usage within baseline ±5%
- [ ] No errors in application logs
- [ ] All health checks passing

Rollback Plan:
- Redeploy previous version (no feature flags, no risk)

**Phase 2: Canary Release (Week 2)**

Goal: Enable media processing for 5% of instances

Actions:
- Enable MEDIA_PROCESSING_ENABLED=true for 5% of instances (whitelist by instance_id)
- Enable JOBS_ENABLED=true globally (all jobs except media cleanup)
- Monitor media job processing metrics
- Validate S3 upload success rate (target: >99%)
- Monitor webhook delivery latency (should not increase >10%)

Success Criteria:
- [ ] Media processing latency <30s p99
- [ ] S3 upload success rate >99%
- [ ] Webhook delivery latency unchanged
- [ ] DLQ size stable
- [ ] No critical errors

Monitoring:
- Dashboard: Media processing metrics (jobs pending, latency, success rate)
- Alerts: Media job backlog >1000, S3 upload failure >5%

Rollback Plan:
- Set MEDIA_PROCESSING_ENABLED=false for canary instances
- DLQ reprocessor will retry failed webhooks

**Phase 3: Gradual Rollout (Weeks 3-4)**

Goal: Increase media processing to 25% → 50% of instances

Week 3 Actions:
- Enable for 25% of instances
- Monitor DLQ size, media job backlog
- Validate cleanup jobs running correctly
- Tune worker count based on load

Week 4 Actions:
- Enable for 50% of instances
- Enable media cleanup job (JOBS_ENABLED=true, all jobs active)
- Verify S3 storage growth rate
- Validate deduplication effectiveness

Success Criteria:
- [ ] Zero critical errors for 7 consecutive days
- [ ] Media processing stable (latency, success rate)
- [ ] Cleanup jobs completing successfully
- [ ] S3 costs within budget

Monitoring:
- Storage growth: <20GB/day
- Deduplication hit rate: >30%
- Cleanup job execution: successful daily runs

Rollback Plan:
- Reduce enabled percentage incrementally
- Pause cleanup jobs if S3 deletions causing issues

**Phase 4: Full Rollout (Week 5)**

Goal: Enable for 100% of instances

Actions:
- Enable MEDIA_PROCESSING_ENABLED=true globally
- All background jobs active and validated
- 24/7 monitoring active
- On-call team briefed with runbooks
- Incident response procedures validated

Success Criteria:
- [ ] System stable for 7 consecutive days
- [ ] All SLAs met (latency, uptime, error rate)
- [ ] No unresolved critical incidents
- [ ] Operational team comfortable with new features

Final Validation:
- Load test: 10,000 media messages processed successfully
- Failover test: Instance disconnect handled gracefully
- Disaster recovery: Database restore and system recovery tested

**Rollback Strategy**

Immediate Rollback Triggers:
- Error rate >1% for 15+ minutes
- API latency >2x baseline for 15+ minutes
- Data loss detected (any amount)
- S3 costs exceeding budget by >50%
- Security incident related to media processing

Rollback Procedure:
1. Set feature flags to false: `MEDIA_PROCESSING_ENABLED=false`, `JOBS_ENABLED=false`
2. Restart application (graceful shutdown)
3. Verify system returns to baseline performance
4. Investigate root cause with logs, metrics, traces
5. Fix issue in staging before re-enabling

Rollback Artifacts:
- Previous deployment tagged in container registry
- Database migration rollback scripts ready
- Configuration rollback in version control
- Communication templates for stakeholders

---

## Operational Readiness

### Runbooks

#### Media Processing Incident Response

**Scenario: Media jobs backing up (>1000 pending)**

Detection:
```
Alert: media_jobs_pending_total > 1000 for 10 minutes
Dashboard: Media Processing > Jobs Pending
```

Diagnosis:
```bash
# Check worker count
psql -c "SELECT instance_id, COUNT(*) FROM media_workers GROUP BY instance_id"

# Check S3 connectivity
aws s3 ls s3://whatsapp-media-production --region us-east-1

# Check WhatsApp API rate limits (application logs)
grep "rate_limit" /var/log/whatsapp-api/app.log

# Check error patterns
psql -c "SELECT error_message, COUNT(*) FROM media_jobs WHERE status='failed' GROUP BY error_message ORDER BY COUNT(*) DESC LIMIT 10"
```

Response:
1. **Increase worker count**: Set `MEDIA_WORKER_COUNT=5` (from 3), rolling restart
2. **S3 issue**: Switch to fallback storage or wait for AWS resolution
3. **WhatsApp rate limit**: Implement rate limiter or backoff strategy
4. **Persistent errors**: Manually investigate failed jobs, fix data issues

Escalation:
- If unresolved in 15 minutes → page infrastructure team
- If S3 unavailable → page AWS support
- If WhatsApp API issues → escalate to WhatsApp support

#### DLQ Growth

**Scenario: DLQ size exceeding threshold (>10,000 events)**

Detection:
```
Alert: webhook_dlq_size_total > 10000
Dashboard: Webhooks > DLQ Size
```

Diagnosis:
```sql
-- Top failing webhooks
SELECT webhook_url, COUNT(*) as failures
FROM webhook_dlq
GROUP BY webhook_url
ORDER BY failures DESC
LIMIT 10;

-- Error patterns
SELECT error_type, COUNT(*)
FROM webhook_dlq
GROUP BY error_type
ORDER BY COUNT(*) DESC;

-- Check webhook endpoint reachability
curl -I https://customer-webhook.com/webhook
```

Response:
1. **Contact webhook owners**: Notify customers with persistent failures
2. **Increase retry attempts**: Temporarily increase max_retries for transient failures
3. **Manual reprocessing**:
   ```sql
   UPDATE webhook_dlq
   SET retry_count = 0
   WHERE webhook_url = 'https://customer-webhook.com/webhook'
     AND error_type = 'network_timeout';
   ```
4. **Trigger DLQ reprocessor manually**: `POST /admin/jobs/dlq-reprocessor/execute`

Escalation:
- If affecting >50% of instances → notify engineering team
- If webhook service-wide issue → escalate to platform team

#### S3 Storage Alert

**Scenario: S3 storage exceeding budget**

Detection:
```
CloudWatch Alert: S3 bucket size > threshold (500GB)
Dashboard: Storage > S3 Usage
```

Diagnosis:
```sql
-- Check media_metadata table size
SELECT COUNT(*), SUM(file_size) / 1024 / 1024 / 1024 as size_gb
FROM media_metadata;

-- Identify largest files
SELECT s3_key, file_size / 1024 / 1024 as size_mb, media_type, uploaded_at
FROM media_metadata
ORDER BY file_size DESC
LIMIT 100;

-- Check cleanup job status
SELECT job_name, last_execution, status, error_message
FROM job_executions
WHERE job_name = 'media_cleanup'
ORDER BY last_execution DESC
LIMIT 10;

-- Check media type distribution
SELECT media_type, COUNT(*), SUM(file_size) / 1024 / 1024 / 1024 as size_gb
FROM media_metadata
GROUP BY media_type;
```

Response:
1. **Manually trigger cleanup job**: `POST /admin/jobs/media-cleanup/execute`
2. **Reduce retention temporarily**: Set `JOB_MEDIA_RETENTION_DAYS=15` (from 30)
3. **Review media type distribution**: Consider compression for videos/images
4. **Check for duplicates**: Verify deduplication working correctly

Escalation:
- If cost projection exceeds budget by >20% → notify management
- If cleanup job failing persistently → page engineering team

#### Instance Failover

**Scenario: Instance appears stale but still has Redis lock**

Detection:
```
Alert: instance_heartbeat_missing > 1 hour
Dashboard: Instances > Health
```

Diagnosis:
```sql
-- Check instance status
SELECT id, partner_id, status, last_heartbeat, connected_at
FROM instances
WHERE last_heartbeat < NOW() - INTERVAL '1 hour';

-- Check Redis lock
redis-cli GET lock:instance:<instance_id>

-- Check if workers still running
ps aux | grep instance_worker | grep <instance_id>
```

Response:
1. **Manual lock release**:
   ```bash
   redis-cli DEL lock:instance:<instance_id>
   ```
2. **Force instance disconnect**: `POST /admin/instances/{id}/disconnect`
3. **Let health monitor job handle automatically** (preferred)

Prevention:
- Improve heartbeat reliability (more frequent, timeout handling)
- Add lock auto-renewal mechanism
- Monitor health monitor job execution

Escalation:
- If multiple instances affected → investigate system-wide issue
- If Redis connection issues → page infrastructure team

### Monitoring Dashboards

**1. Media Processing Dashboard**

Panels:
- Media jobs pending/processing/completed (last 24h) - line chart
- Media processing latency (p50, p95, p99) - gauge
- S3 upload success rate - gauge (target: >99%)
- Media deduplication hit rate - gauge
- Storage usage trend - area chart
- Worker count by instance - table
- Top errors - table

Alerts:
- Critical: Media job backlog >5000, S3 upload failure rate >5%
- High: Worker crash rate >10/hour
- Medium: Media processing latency >60s p99

**2. Background Jobs Dashboard**

Panels:
- Job execution frequency (last 24h) - bar chart
- Job execution duration - heatmap
- DLQ size trend - line chart
- Outbox cleanup metrics (rows deleted) - bar chart
- Media cleanup metrics (storage freed GB) - bar chart
- Health monitor alerts (stale instances detected) - count

Alerts:
- Critical: DLQ size >50000, Job execution failures >3 consecutive
- Medium: Cleanup job missed execution, Worker restart rate >5/hour

**3. System Health Dashboard**

Panels:
- Instance count by status (connected/disconnected/error) - pie chart
- Webhook delivery success rate - gauge
- Database connection pool usage - gauge
- Redis operations latency - line chart
- API endpoint response times (p95) - table

Alerts:
- Critical: Database connections >80%, API latency >1s p95
- High: Redis latency >100ms p95

### Alert Rules

**Critical Alerts** (page on-call immediately):
- Media job backlog >5000 for 10 minutes
- DLQ size >50000
- S3 upload failure rate >5% for 5 minutes
- Database connections >90% for 5 minutes
- API error rate >1% for 15 minutes

**High Alerts** (notify engineering team):
- Worker crash rate >10/hour
- Job execution failures >3 consecutive
- Media processing latency >60s p99 for 10 minutes
- Redis latency >100ms p95 for 10 minutes

**Medium Alerts** (log for review):
- S3 storage growth >20GB/day
- Cleanup job missed execution
- Worker restart rate >5/hour
- Media deduplication hit rate <50%

---

## Performance Optimization

### Media Processing Optimizations

**1. Parallel Processing**

Current: Sequential download → upload
```go
data, err := downloader.Download(ctx, url)
s3URL, err := uploader.Upload(ctx, data)
```

Optimization: Stream download directly to S3
```go
reader, err := downloader.DownloadStream(ctx, url)
s3URL, err := uploader.UploadStream(ctx, reader) // Multipart upload
```

Benefit: 40-60% latency reduction for large files (>5MB)

**2. Deduplication Cache**

Current: Database query for each media check
```go
metadata, err := metadataRepo.FindByFileHash(ctx, instanceID, fileSha256)
```

Optimization: Redis cache for file_sha256 → s3_url mapping
```go
// Check Redis cache first
if cachedURL := redisClient.Get(ctx, cacheKey); cachedURL != "" {
    return cachedURL // Cache hit
}

// Cache miss, query database
metadata, err := metadataRepo.FindByFileHash(ctx, instanceID, fileSha256)
if err == nil {
    // Store in cache with 24h TTL
    redisClient.SetEX(ctx, cacheKey, metadata.S3URL, 24*time.Hour)
}
```

Benefit: 70% reduction in database queries, faster lookups

**3. Worker Pool Tuning**

Start conservative: 3 workers per instance
```yaml
MEDIA_WORKER_COUNT=3
```

Monitor and auto-scale:
- Queue depth >100 → increase to 5 workers
- Queue depth >500 → increase to 10 workers (max)
- Idle >30min → decrease to 2 workers (min)

Benefit: Adaptive resource usage, better throughput

### Background Jobs Optimizations

**1. DLQ Reprocessor Batching**

Batch size: 100 events per iteration
Parallel delivery: 5 concurrent webhook requests
Transaction batching: Commit every 50 moves

Benefit: 3-5x throughput improvement

**2. Cleanup Job Partitioning**

Process 1 week at a time:
```sql
DELETE FROM webhook_outbox
WHERE created_at >= '2025-09-26' AND created_at < '2025-10-03'
  AND status = 'completed'
LIMIT 1000;
```

Benefit: Avoid long-running transactions, prevent table locks

**3. Health Monitor Efficiency**

Single query for all stale instances:
```sql
SELECT * FROM instances
WHERE status = 'connected' AND last_heartbeat < NOW() - INTERVAL '1 hour'
LIMIT 100;
```

Batch Redis lock releases:
```go
pipeline := redisClient.Pipeline()
for _, instance := range staleInstances {
    pipeline.Del(ctx, fmt.Sprintf("lock:instance:%s", instance.ID))
}
pipeline.Exec(ctx)
```

Benefit: Sub-second execution time

### Database Optimizations

**Index Strategy**:
```sql
-- Media jobs polling
CREATE INDEX idx_media_jobs_pending ON media_jobs(status, created_at)
WHERE status = 'pending';

-- Media deduplication
CREATE INDEX idx_media_metadata_sha256 ON media_metadata(instance_id, file_sha256);

-- DLQ reprocessor
CREATE INDEX idx_dlq_retry ON webhook_dlq(retry_count, created_at)
WHERE retry_count < max_retries;

-- Cleanup jobs
CREATE INDEX idx_outbox_cleanup ON webhook_outbox(status, created_at)
WHERE status IN ('completed', 'failed');
```

**Connection Pooling**:
- Media workers: Separate pool (max 20 connections)
- Background jobs: Shared pool (max 10 connections)
- API handlers: Separate pool (max 50 connections)

Benefit: Prevent connection exhaustion, better isolation

**Query Optimization**:
- Use prepared statements (already implemented)
- Avoid N+1 queries (batch loads)
- Use `EXPLAIN ANALYZE` for slow queries
- Set `statement_timeout = 30s`

### Performance Targets

**Media Processing**:
- Latency: <30s p99, <10s p95, <5s p50
- S3 upload success rate: >99%
- Deduplication hit rate: >30%
- Worker restart time: <5s

**Background Jobs**:
- DLQ reprocessor: <10s execution for 100 events
- Cleanup jobs: <60s execution
- Health monitor: <5s execution
- Job execution success rate: >95%

**System Health**:
- API latency: <500ms p95, <1s p99
- Database connections: <80% pool usage
- Redis latency: <10ms p95
- Uptime: 99.9% (8.7 hours downtime/year)

---

## Risk Management

### Critical Success Factors

1. **S3 Integration Stability** (Impact: HIGH)
   - AWS SDK configuration correct
   - Network timeouts properly configured
   - Presigned URL generation reliable
   - Mitigation: Extensive testing with real S3, fallback to alternative storage
   - Validation: Load test 10,000 uploads, chaos test network failures

2. **Media Deduplication Accuracy** (Impact: MEDIUM)
   - file_sha256 uniqueness per instance
   - Race conditions in concurrent uploads
   - Cache invalidation timing
   - Mitigation: Database unique constraints, transaction isolation
   - Validation: Concurrent upload test, duplicate media test

3. **Worker Coordination** (Impact: HIGH)
   - No duplicate processing of same media job
   - Graceful worker shutdown without data loss
   - Instance failover handling
   - Mitigation: Redis locks, status state machine, idempotency
   - Validation: Failover test, crash recovery test

4. **Job Scheduler Reliability** (Impact: MEDIUM)
   - Cron schedules execute at correct times
   - Jobs don't overlap (mutex per job)
   - Graceful shutdown completes inflight jobs
   - Mitigation: Use battle-tested library (gocron), job-level locking
   - Validation: Schedule accuracy test, concurrent execution test

5. **Database Performance** (Impact: HIGH)
   - Outbox queries use indexes efficiently
   - Media jobs polling doesn't overwhelm DB
   - Cleanup operations don't lock tables
   - Mitigation: Query optimization, batch processing, off-peak scheduling
   - Validation: EXPLAIN ANALYZE all queries, load test with 100K events

### Key Risks

**1. S3 Costs** (Risk: MEDIUM)
- Unexpected storage growth (>1TB/month)
- High presigned URL generation costs
- Mitigation: Media cleanup job, monitor storage metrics, set CloudWatch alarms
- Contingency: Reduce retention to 15 days, implement aggressive cleanup

**2. Media Processing Bottleneck** (Risk: MEDIUM)
- Large files (>50MB) slow down workers
- Download timeouts from WhatsApp (>5 min)
- Mitigation: Configurable worker count, per-media-type queues, timeout tuning
- Contingency: Scale workers horizontally, implement priority queues

**3. DLQ Growth** (Risk: LOW)
- Persistent webhook endpoint failures
- Events stuck in DLQ indefinitely
- Mitigation: DLQ reprocessor, alerting on DLQ size, manual intervention runbook
- Contingency: Contact customers, increase retry limits, implement exponential backoff

**4. Instance State Inconsistency** (Risk: MEDIUM)
- Redis lock lost but instance still running
- Health monitor false positives
- Mitigation: Heartbeat mechanism, graceful lock renewal, operator alerts
- Contingency: Manual lock release, instance restart, investigate Redis stability

### Contingency Plans

**S3 Unavailable**:
1. Switch webhook delivery to WhatsApp URLs (temporary)
2. Queue media jobs for later processing
3. Notify customers of temporary degradation
4. Resume media processing when S3 available

**Database Performance Degradation**:
1. Reduce worker count to lower DB load
2. Pause cleanup jobs temporarily
3. Scale database vertically (increase instance size)
4. Optimize slow queries identified in logs

**Worker Pool Exhaustion**:
1. Increase `MEDIA_WORKER_COUNT` dynamically
2. Implement worker autoscaling based on queue depth
3. Add more application instances
4. Investigate and fix memory leaks

---

## Code Review Guidelines

### Phase 6: Media Processing Review

**Security Review**:
- [ ] S3 credentials never logged or exposed in errors
- [ ] Presigned URLs have appropriate expiration (24h default, configurable)
- [ ] Media decryption uses constant-time comparisons (crypto timing attacks)
- [ ] File type validation before upload (prevent malicious files)
- [ ] S3 bucket policies restrict public access (no public reads)
- [ ] No path traversal vulnerabilities in S3 key generation

**Error Handling Review**:
- [ ] All downloader errors categorized (retryable vs fatal)
- [ ] S3 upload failures trigger proper retry logic with exponential backoff
- [ ] Partial downloads/uploads cleaned up (no orphaned files)
- [ ] Worker crashes don't leave orphaned 'processing' jobs
- [ ] All errors logged with structured context (instance_id, job_id, error_type)
- [ ] Critical errors captured in Sentry with tags (component, instance_id)

**Observability Review**:
- [ ] All metrics follow naming conventions (media_*)
- [ ] Duration histograms use appropriate buckets (0.1, 0.5, 1, 5, 10, 30s)
- [ ] Instance_id included in all media logs
- [ ] Media type (image/video/audio) tagged in metrics
- [ ] Download/upload sizes tracked (histogram)
- [ ] Deduplication hits/misses measured (counter)

**Performance Review**:
- [ ] No blocking operations in worker main loop (polling, processing async)
- [ ] Database queries use indexes (EXPLAIN ANALYZE verified)
- [ ] S3 client configured with appropriate timeouts (download: 5min, upload: 10min)
- [ ] Context deadlines propagated correctly (cancel downloads on timeout)
- [ ] No unbounded goroutine spawning (worker pool limits)
- [ ] Memory limits considered for large files (streaming, not loading into memory)

**Data Integrity Review**:
- [ ] file_sha256 uniqueness enforced per instance (database unique constraint)
- [ ] Media job status transitions atomic (transaction-safe updates)
- [ ] No race conditions in concurrent uploads (Redis locks or DB locks)
- [ ] Transaction boundaries correct (commit after complete operation)
- [ ] Foreign key constraints validated (media_jobs → instances, → webhook_outbox)
- [ ] Migration rollback tested (down migrations work)

### Phase 7: Background Jobs Review

**Scheduling Review**:
- [ ] Cron expressions validated and documented (comments explain schedule)
- [ ] Jobs don't overlap (mutex per job if needed, single execution guaranteed)
- [ ] Timezone handling explicit (UTC for all cron schedules)
- [ ] Job execution timeout configured (prevent runaway jobs)
- [ ] Graceful shutdown waits for inflight jobs (30s timeout)
- [ ] Startup delay configured to avoid thundering herd (stagger job starts)

**Batch Processing Review**:
- [ ] Batch sizes configurable and reasonable (100-1000 range)
- [ ] Transaction commits at appropriate intervals (every 50 items)
- [ ] Individual item failures don't fail entire batch (error handling per item)
- [ ] Progress tracking for long-running jobs (log every N items)
- [ ] Pagination used for large result sets (LIMIT/OFFSET or cursor-based)
- [ ] Database locks minimized (use FOR UPDATE SKIP LOCKED)

**Resource Management Review**:
- [ ] Database connections returned to pool (defer close)
- [ ] Redis connections properly closed (defer close)
- [ ] S3 client reused (not created per job execution)
- [ ] Context cancellation checked in loops (select on ctx.Done())
- [ ] Goroutines don't leak on error (defer cleanup)
- [ ] Memory usage bounded (no unbounded slices, stream large data)

**Retry Logic Review**:
- [ ] Exponential backoff implemented correctly (base, multiplier, max delay)
- [ ] Max retries enforced (prevent infinite retries)
- [ ] Retry state persisted to database (retry_count incremented)
- [ ] Idempotency maintained (safe to retry same operation)
- [ ] Retry delays respect job schedule (don't retry too frequently)
- [ ] Permanent failures handled (move to DLQ or mark as final failure)

**Testing Review**:
- [ ] Unit tests for each job's Execute method
- [ ] Mock dependencies (database, Redis, S3, time.Now)
- [ ] Test cron schedule parsing (valid expressions)
- [ ] Test graceful shutdown scenarios (context cancellation)
- [ ] Test batch processing edge cases (empty batch, errors mid-batch)
- [ ] Integration tests with real scheduler (jobs execute on time)

### Cross-Cutting Concerns

- [ ] All new code follows Go conventions (gofmt, golint pass)
- [ ] Context propagation consistent throughout (ctx first parameter)
- [ ] Structured logging with slog, no fmt.Println (use logger.Info/Error)
- [ ] Errors wrapped with context (%w, not %v)
- [ ] No hardcoded values (use config structs)
- [ ] Comments explain "why" not "what" (document intent)
- [ ] Public functions have godoc comments
- [ ] README and CLAUDE.md updated with new features

---

## Timeline Summary

### Development Timeline (13 days)

**Week 1: Phase 6 Foundation (Days 1-3)**
- Day 1: Database schema + configuration
- Day 2: S3 integration + WhatsApp downloader
- Day 3: Media processor core logic

**Week 2: Phase 6 Integration (Days 4-7)**
- Day 4: Media repositories
- Day 5: Media worker + coordinator
- Day 6: EventProcessor integration
- Day 7: Testing and validation

**Week 3: Phase 7 Jobs (Days 8-10)**
- Day 8: Jobs config + scheduler framework
- Day 9: Health monitor + DLQ reprocessor
- Day 10: Media retry job

**Week 4: Phase 7 Completion (Days 11-13)**
- Day 11: Cleanup jobs (outbox + media)
- Day 12: Main integration + optional metrics export
- Day 13: Testing and documentation

### Deployment Timeline (5 weeks)

- Week 5: Dark launch (feature flags OFF, infrastructure validation)
- Week 6: Canary release (5% instances)
- Week 7: Gradual rollout (25%)
- Week 8: Gradual rollout (50%)
- Week 9: Full rollout (100%), stability validation

**Total: 7 weeks from start to production**

---

## Success Metrics

### Development Metrics
- Unit test coverage: >80% ✅
- Integration test coverage: All critical paths ✅
- Code review approval: All reviewers ✅
- Performance benchmarks: Meet targets ✅

### Deployment Metrics
- Rollout duration: 5 weeks (as planned) ✅
- Rollbacks required: 0 ✅
- Critical incidents: 0 ✅
- Customer complaints: 0 ✅

### Production Metrics
- Media processing latency: <30s p99 ✅
- S3 upload success rate: >99% ✅
- Webhook delivery success rate: >99% ✅
- Background job success rate: >95% ✅
- System uptime: 99.9% (8.7h downtime/year) ✅
- DLQ size: <5000 events ✅
- Storage costs: Within budget ✅

---

## Appendices

### Appendix A: Environment Variables Reference

```bash
# Feature Flags
MEDIA_PROCESSING_ENABLED=false  # Enable media processing
JOBS_ENABLED=true              # Enable background jobs

# Media Processing
MEDIA_S3_BUCKET=whatsapp-media
MEDIA_S3_REGION=us-east-1
MEDIA_S3_ACCESS_KEY=<aws-access-key>
MEDIA_S3_SECRET_KEY=<aws-secret-key>
MEDIA_S3_ENDPOINT=http://localhost:9000  # MinIO for dev
MEDIA_PRESIGNED_EXPIRY=24h
MEDIA_MAX_FILE_SIZE=52428800  # 50MB
MEDIA_WORKER_COUNT=3
MEDIA_DOWNLOAD_TIMEOUT=5m
MEDIA_UPLOAD_TIMEOUT=10m

# Background Jobs - DLQ Reprocessor
JOB_DLQ_ENABLED=true
JOB_DLQ_SCHEDULE="*/5 * * * *"  # Every 5 minutes
JOB_DLQ_BATCH_SIZE=100
JOB_DLQ_TIMEOUT=5m

# Background Jobs - Media Retry
JOB_MEDIA_RETRY_ENABLED=true
JOB_MEDIA_RETRY_SCHEDULE="*/10 * * * *"  # Every 10 minutes
JOB_MEDIA_RETRY_BATCH_SIZE=50
JOB_MEDIA_RETRY_TIMEOUT=10m

# Background Jobs - Outbox Cleanup
JOB_OUTBOX_CLEANUP_ENABLED=true
JOB_OUTBOX_CLEANUP_SCHEDULE="0 2 * * *"  # 2 AM daily
JOB_OUTBOX_CLEANUP_BATCH_SIZE=1000
JOB_OUTBOX_RETENTION_DAYS=7
JOB_OUTBOX_CLEANUP_TIMEOUT=30m

# Background Jobs - Media Cleanup
JOB_MEDIA_CLEANUP_ENABLED=true
JOB_MEDIA_CLEANUP_SCHEDULE="0 3 * * *"  # 3 AM daily
JOB_MEDIA_CLEANUP_BATCH_SIZE=100
JOB_MEDIA_RETENTION_DAYS=30
JOB_MEDIA_CLEANUP_TIMEOUT=1h

# Background Jobs - Health Monitor
JOB_HEALTH_MONITOR_ENABLED=true
JOB_HEALTH_MONITOR_SCHEDULE="*/10 * * * *"  # Every 10 minutes
JOB_HEALTH_MONITOR_BATCH_SIZE=100
JOB_HEALTH_MONITOR_TIMEOUT=2m
```

### Appendix B: Database Schema DDL

See migrations directory:
- `api/migrations/000004_create_media_metadata.sql`
- `api/migrations/000006_create_media_jobs.sql`

### Appendix C: Metrics Reference

**Media Processing Metrics**:
```
media_coordinator_starts_total
media_coordinator_stops_total
media_coordinator_workers_active gauge
media_worker_starts_total{instance_id}
media_worker_stops_total{instance_id}
media_worker_polls_total{instance_id}
media_worker_jobs_processed_total{instance_id, status}
media_download_attempts_total{status, error_type}
media_download_duration_seconds{media_type}
media_upload_attempts_total{status}
media_upload_duration_seconds{media_type}
media_processing_duration_seconds{media_type, status}
media_deduplication_hits_total
media_deduplication_misses_total
```

**Background Jobs Metrics**:
```
job_executions_total{job_name, status}
job_duration_seconds{job_name}
job_errors_total{job_name, error_type}
dlq_events_reprocessed_total{status}
dlq_events_moved_back_total
outbox_events_deleted_total{status}
media_files_deleted_total
media_storage_freed_bytes
instances_checked_total
stale_instances_detected_total
locks_released_total
```

---

## Conclusion

This development plan provides a complete roadmap for implementing Phase 6 (Media Processing) and Phase 7 (Background Jobs), representing the final 30% of the WhatsApp API project. The plan emphasizes:

1. **Architectural Consistency**: New components mirror existing patterns (coordinator → worker → processor)
2. **Risk Mitigation**: Gradual rollout with feature flags, comprehensive testing, clear rollback procedures
3. **Operational Readiness**: Runbooks, monitoring dashboards, alert rules prepared before deployment
4. **Performance Focus**: Optimizations identified and prioritized (deduplication caching, streaming uploads)
5. **Quality Standards**: >80% test coverage, code review checklists, acceptance criteria for each step

**Timeline**: 13 days development + 5 weeks rollout = **7 weeks to production**

**Next Steps**:
1. Review and approve this plan with stakeholders
2. Provision infrastructure (S3, monitoring dashboards)
3. Begin Phase 6 Step 6.1 (Database Schema)
4. Follow implementation steps sequentially
5. Deploy with dark launch and gradual rollout

**Success Factors**:
- Clear acceptance criteria for each step
- Continuous integration with existing system
- Proactive monitoring and alerting
- Well-documented operational procedures
- Stakeholder communication throughout rollout

This plan ensures a structured, safe, and successful delivery of the remaining project scope.

---

**Document Version**: 1.0
**Author**: Development Team
**Date**: 2025-10-03
**Status**: Ready for Implementation
