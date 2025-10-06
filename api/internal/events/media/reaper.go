package media

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/locks"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

const (
	mediaCleanupLockKey = "media:cleanup:lock"
)

type MediaReaperConfig struct {
	Interval       time.Duration
	BatchSize      int
	S3Retention    time.Duration
	LocalRetention time.Duration
	LockTTL        time.Duration
}

type MediaReaper struct {
	repo         persistence.MediaRepository
	s3Uploader   *S3Uploader
	localStorage *LocalMediaStorage
	metrics      *observability.Metrics
	logger       *slog.Logger
	lockManager  locks.Manager

	interval       time.Duration
	batchSize      int
	s3Retention    time.Duration
	localRetention time.Duration
	lockTTL        time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewMediaReaper(
	repo persistence.MediaRepository,
	s3Uploader *S3Uploader,
	localStorage *LocalMediaStorage,
	metrics *observability.Metrics,
	logger *slog.Logger,
	lockManager locks.Manager,
	cfg MediaReaperConfig,
) (*MediaReaper, error) {
	if repo == nil {
		return nil, errors.New("media repository is required for reaper")
	}
	if metrics == nil {
		return nil, errors.New("metrics collector is required for reaper")
	}
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.Interval <= 0 {
		return nil, errors.New("cleanup interval must be greater than zero")
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 200
	}
	lockTTL := cfg.LockTTL
	if lockTTL <= 0 {
		// give enough time for cleanup to finish; at least twice the interval
		lockTTL = cfg.Interval * 2
		if lockTTL < time.Minute {
			lockTTL = time.Minute
		}
	}

	reaper := &MediaReaper{
		repo:         repo,
		s3Uploader:   s3Uploader,
		localStorage: localStorage,
		metrics:      metrics,
		logger: logger.With(
			slog.String("component", "media_reaper"),
		),
		lockManager:    lockManager,
		interval:       cfg.Interval,
		batchSize:      cfg.BatchSize,
		s3Retention:    cfg.S3Retention,
		localRetention: cfg.LocalRetention,
		lockTTL:        lockTTL,
	}

	return reaper, nil
}

func (r *MediaReaper) Start(parent context.Context) {
	if parent == nil {
		parent = context.Background()
	}
	if r.cancel != nil {
		return
	}
	r.ctx, r.cancel = context.WithCancel(parent)
	r.logger.Info("media reaper started",
		slog.Duration("interval", r.interval),
		slog.Int("batch_size", r.batchSize),
		slog.Duration("s3_retention", r.s3Retention),
		slog.Duration("local_retention", r.localRetention))

	r.wg.Add(1)
	go r.loop()
}

func (r *MediaReaper) Stop(ctx context.Context) error {
	if r.cancel == nil {
		return nil
	}

	r.cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		r.wg.Wait()
	}()

	select {
	case <-done:
		r.logger.Info("media reaper stopped")
		return nil
	case <-ctx.Done():
		r.logger.Warn("media reaper stop timeout",
			slog.String("error", ctx.Err().Error()))
		return ctx.Err()
	}
}

func (r *MediaReaper) loop() {
	defer r.wg.Done()

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.runOnce(r.ctx, "startup")

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.runOnce(r.ctx, "scheduled")
		}
	}
}

func (r *MediaReaper) runOnce(ctx context.Context, trigger string) {
	start := time.Now()
	logger := logging.ContextLogger(ctx, r.logger).With(
		slog.String("trigger", trigger),
	)

	lock, acquired := r.acquireLock(ctx, logger)
	if lock != nil {
		defer func() {
			if err := lock.Release(ctx); err != nil {
				logger.Warn("failed to release cleanup lock",
					slog.String("error", err.Error()))
			}
		}()
	}

	if !acquired {
		logger.Debug("media cleanup skipped (lock not acquired)")
		r.metrics.MediaCleanupRuns.WithLabelValues("skipped").Inc()
		return
	}

	stats := r.performCleanup(ctx, logger)

	duration := time.Since(start)

	resultLabel := "empty"
	switch {
	case stats.Errors > 0 && stats.DeletedItems > 0:
		resultLabel = "partial"
	case stats.Errors > 0:
		resultLabel = "error"
	case stats.DeletedItems > 0:
		resultLabel = "success"
	}

	r.metrics.MediaCleanupRuns.WithLabelValues(resultLabel).Inc()
	r.metrics.MediaCleanupDuration.Observe(duration.Seconds())

	if stats.DeletedItems > 0 {
		r.metrics.MediaCleanupTotal.WithLabelValues("expired_files").Add(float64(stats.DeletedItems))
	}
	if stats.DeletedBytes[persistence.StorageTypeS3] > 0 {
		r.metrics.MediaCleanupDeletedBytes.WithLabelValues(string(persistence.StorageTypeS3)).Add(float64(stats.DeletedBytes[persistence.StorageTypeS3]))
	}
	if stats.DeletedBytes[persistence.StorageTypeLocal] > 0 {
		r.metrics.MediaCleanupDeletedBytes.WithLabelValues(string(persistence.StorageTypeLocal)).Add(float64(stats.DeletedBytes[persistence.StorageTypeLocal]))
	}

	logger.Info("media cleanup completed",
		slog.String("result", resultLabel),
		slog.Int("deleted_items", stats.DeletedItems),
		slog.Int("errors", stats.Errors),
		slog.Duration("duration", duration))
}

type cleanupStats struct {
	DeletedItems int
	Errors       int
	DeletedBytes map[persistence.StorageType]int64
}

func (r *MediaReaper) performCleanup(ctx context.Context, logger *slog.Logger) cleanupStats {
	stats := cleanupStats{DeletedBytes: make(map[persistence.StorageType]int64)}

	s3Cutoff := r.computeCutoff(r.s3Retention)
	localCutoff := r.computeCutoff(r.localRetention)

	for {
		select {
		case <-ctx.Done():
			return stats
		default:
		}

		records, err := r.repo.ListExpiredMedia(ctx, s3Cutoff, localCutoff, r.batchSize)
		if err != nil {
			stats.Errors++
			r.metrics.MediaCleanupErrors.WithLabelValues("repository", "list_failed").Inc()
			logger.Error("failed to list expired media",
				slog.String("error", err.Error()))
			return stats
		}

		if len(records) == 0 {
			return stats
		}

		var successfulIDs []int64

		for _, record := range records {
			if ctx.Err() != nil {
				return stats
			}

			switch record.StorageType {
			case persistence.StorageTypeS3:
				if r.deleteS3Object(ctx, record, logger) {
					successfulIDs = append(successfulIDs, record.ID)
					stats.DeletedItems++
					stats.DeletedBytes[persistence.StorageTypeS3] += record.SizeBytes
				} else {
					stats.Errors++
				}
			case persistence.StorageTypeLocal:
				if ok, size := r.deleteLocalObject(ctx, record, logger); ok {
					successfulIDs = append(successfulIDs, record.ID)
					stats.DeletedItems++
					if size > 0 {
						stats.DeletedBytes[persistence.StorageTypeLocal] += size
					} else {
						stats.DeletedBytes[persistence.StorageTypeLocal] += record.SizeBytes
					}
				} else {
					stats.Errors++
				}
			default:
				// unknown storage type, skip but count as error so we can investigate
				stats.Errors++
				r.metrics.MediaCleanupErrors.WithLabelValues(string(record.StorageType), "unsupported_storage").Inc()
				logger.Warn("skipping media cleanup for unsupported storage type",
					slog.Int64("media_id", record.ID),
					slog.String("storage_type", string(record.StorageType)))
			}
		}

		if len(successfulIDs) > 0 {
			if err := r.repo.DeleteMediaByIDs(ctx, successfulIDs); err != nil {
				stats.Errors++
				r.metrics.MediaCleanupErrors.WithLabelValues("repository", "delete_failed").Inc()
				logger.Error("failed to delete media metadata",
					slog.String("error", err.Error()))
				// do not remove ids from count; entries will be retried later
			}
		}

		if len(records) < r.batchSize {
			return stats
		}
	}
}

func (r *MediaReaper) deleteS3Object(ctx context.Context, record *persistence.ExpiredMediaRecord, logger *slog.Logger) bool {
	if r.s3Uploader == nil {
		r.metrics.MediaCleanupErrors.WithLabelValues(string(persistence.StorageTypeS3), "s3_uploader_unavailable").Inc()
		logger.Error("s3 uploader unavailable for media cleanup")
		return false
	}

	bucket := ""
	if record.S3Bucket != nil {
		bucket = *record.S3Bucket
	}
	if bucket == "" {
		bucket = r.s3Uploader.bucket
	}
	if record.S3Key == nil || *record.S3Key == "" {
		r.metrics.MediaCleanupErrors.WithLabelValues(string(persistence.StorageTypeS3), "missing_key").Inc()
		logger.Warn("skipping s3 cleanup due to missing key",
			slog.Int64("media_id", record.ID))
		return false
	}

	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := r.s3Uploader.DeleteObject(deleteCtx, bucket, *record.S3Key); err != nil {
		r.metrics.MediaCleanupErrors.WithLabelValues(string(persistence.StorageTypeS3), "delete_failed").Inc()
		logger.Error("failed to delete media from s3",
			slog.Int64("media_id", record.ID),
			slog.String("bucket", bucket),
			slog.String("key", *record.S3Key),
			slog.String("error", err.Error()))
		return false
	}

	logger.Debug("deleted media from s3",
		slog.Int64("media_id", record.ID),
		slog.String("bucket", bucket),
		slog.String("key", *record.S3Key))
	return true
}

func (r *MediaReaper) deleteLocalObject(ctx context.Context, record *persistence.ExpiredMediaRecord, logger *slog.Logger) (bool, int64) {
	if r.localStorage == nil {
		r.metrics.MediaCleanupErrors.WithLabelValues(string(persistence.StorageTypeLocal), "local_storage_unavailable").Inc()
		logger.Error("local storage not configured; skipping local media cleanup",
			slog.Int64("media_id", record.ID))
		return false, 0
	}

	if record.S3Key == nil || *record.S3Key == "" {
		r.metrics.MediaCleanupErrors.WithLabelValues(string(persistence.StorageTypeLocal), "missing_path").Inc()
		logger.Warn("skipping local cleanup due to missing path",
			slog.Int64("media_id", record.ID))
		return false, 0
	}

	size, err := r.localStorage.DeleteMedia(ctx, *record.S3Key)
	if err != nil {
		r.metrics.MediaCleanupErrors.WithLabelValues(string(persistence.StorageTypeLocal), "delete_failed").Inc()
		logger.Error("failed to delete local media",
			slog.Int64("media_id", record.ID),
			slog.String("path", *record.S3Key),
			slog.String("error", err.Error()))
		return false, 0
	}

	if size > 0 {
		logger.Debug("deleted local media",
			slog.Int64("media_id", record.ID),
			slog.Int64("size", size))
	}

	return true, size
}

func (r *MediaReaper) acquireLock(ctx context.Context, logger *slog.Logger) (locks.Lock, bool) {
	if r.lockManager == nil {
		return nil, true
	}

	ttlSeconds := int(r.lockTTL.Seconds())
	if ttlSeconds <= 0 {
		ttlSeconds = 60
	}

	lock, acquired, err := r.lockManager.Acquire(ctx, mediaCleanupLockKey, ttlSeconds)
	if err != nil {
		logger.Error("failed to acquire cleanup lock",
			slog.String("error", err.Error()))
		r.metrics.MediaCleanupErrors.WithLabelValues("lock", "acquire_failed").Inc()
		return nil, false
	}
	if !acquired {
		return nil, false
	}

	return lock, true
}

func (r *MediaReaper) computeCutoff(retention time.Duration) time.Time {
	if retention <= 0 {
		return time.Unix(0, 0)
	}
	return time.Now().Add(-retention)
}
