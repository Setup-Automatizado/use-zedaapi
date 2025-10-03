package media

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// MediaWorker processes media for a specific WhatsApp instance
type MediaWorker struct {
	instanceID   uuid.UUID
	client       *whatsmeow.Client
	processor    *MediaProcessor
	mediaRepo    persistence.MediaRepository
	outboxRepo   persistence.OutboxRepository // Added to fetch original proto messages
	pollInterval time.Duration
	batchSize    int
	workerID     string
	metrics      *observability.Metrics
	logger       *slog.Logger
	stopCh       chan struct{}
	doneCh       chan struct{}
}

// NewMediaWorker creates a new media worker for an instance
func NewMediaWorker(
	ctx context.Context,
	instanceID uuid.UUID,
	client *whatsmeow.Client,
	cfg *config.Config,
	mediaRepo persistence.MediaRepository,
	outboxRepo persistence.OutboxRepository,
	metrics *observability.Metrics,
) (*MediaWorker, error) {
	logger := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "media_worker"),
		slog.String("instance_id", instanceID.String()))

	// Create processor
	processor, err := NewMediaProcessor(ctx, cfg, mediaRepo, outboxRepo, metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to create media processor: %w", err)
	}

	workerID := fmt.Sprintf("worker-%s-%d", instanceID.String()[:8], time.Now().Unix())

	logger.Info("media worker created",
		slog.String("worker_id", workerID),
		slog.Duration("poll_interval", cfg.Events.MediaPollInterval),
		slog.Int("batch_size", cfg.Events.MediaBatchSize))

	return &MediaWorker{
		instanceID:   instanceID,
		client:       client,
		processor:    processor,
		mediaRepo:    mediaRepo,
		outboxRepo:   outboxRepo,
		pollInterval: cfg.Events.MediaPollInterval,
		batchSize:    cfg.Events.MediaBatchSize,
		workerID:     workerID,
		metrics:      metrics,
		logger:       logger,
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
	}, nil
}

// Start begins processing media for this instance
func (w *MediaWorker) Start(ctx context.Context) {
	w.logger.Info("media worker starting")

	go w.run(ctx)
}

// Stop gracefully stops the worker
func (w *MediaWorker) Stop(ctx context.Context) error {
	w.logger.Info("media worker stopping")

	close(w.stopCh)

	// Wait for worker to finish with timeout
	select {
	case <-w.doneCh:
		w.logger.Info("media worker stopped gracefully")
		return nil
	case <-ctx.Done():
		w.logger.Warn("media worker stop timeout")
		return ctx.Err()
	}
}

// run is the main worker loop
func (w *MediaWorker) run(ctx context.Context) {
	defer close(w.doneCh)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.logger.Info("media worker running")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("media worker context cancelled")
			return
		case <-w.stopCh:
			w.logger.Info("media worker stop signal received")
			return
		case <-ticker.C:
			// Poll and process media
			if err := w.processBatch(ctx); err != nil {
				w.logger.Error("batch processing failed",
					slog.String("error", err.Error()))
			}
		}
	}
}

// processBatch polls for pending media and processes them
func (w *MediaWorker) processBatch(ctx context.Context) error {
	// Poll for pending media downloads
	mediaItems, err := w.mediaRepo.PollPendingDownloads(ctx, w.batchSize)
	if err != nil {
		w.metrics.MediaFailures.WithLabelValues(w.instanceID.String(), "unknown", "poll").Inc()
		return fmt.Errorf("failed to poll pending media: %w", err)
	}

	if len(mediaItems) == 0 {
		// No pending media
		return nil
	}

	w.logger.Debug("processing media batch",
		slog.Int("count", len(mediaItems)))

	// Update backlog metric
	w.metrics.MediaBacklog.Set(float64(len(mediaItems)))

	// Process each media item
	for _, media := range mediaItems {
		if err := w.processMedia(ctx, media); err != nil {
			w.logger.Error("media processing failed",
				slog.String("event_id", media.EventID.String()),
				slog.String("error", err.Error()))
			// Continue processing other items
			continue
		}
	}

	return nil
}

// processMedia processes a single media item
func (w *MediaWorker) processMedia(ctx context.Context, media *persistence.MediaMetadata) error {
	logger := w.logger.With(
		slog.String("event_id", media.EventID.String()),
		slog.String("media_type", string(media.MediaType)))

	// Try to acquire lock for processing
	acquired, err := w.mediaRepo.AcquireForProcessing(ctx, media.EventID, w.workerID)
	if err != nil {
		logger.Error("failed to acquire processing lock",
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !acquired {
		logger.Debug("media already being processed by another worker")
		return nil
	}

	// Ensure we release the lock on error
	defer func() {
		if err != nil {
			_ = w.mediaRepo.ReleaseFromProcessing(ctx, media.EventID, w.workerID)
		}
	}()

	logger.Info("processing media")

	// Reconstruct proto message from event_outbox payload
	msg, err := w.reconstructMessage(ctx, media)
	if err != nil {
		logger.Error("failed to reconstruct message",
			slog.String("error", err.Error()))
		w.metrics.MediaFailures.WithLabelValues(w.instanceID.String(), string(media.MediaType), "reconstruct").Inc()
		return fmt.Errorf("failed to reconstruct message: %w", err)
	}

	// Process with retry
	result, err := w.processor.ProcessWithRetry(ctx, w.client, w.instanceID, media.EventID, msg, media.MediaKey)
	if err != nil {
		logger.Error("media processing failed after retries",
			slog.String("error", err.Error()))
		w.metrics.MediaFailures.WithLabelValues(w.instanceID.String(), string(media.MediaType), "process").Inc()
		return fmt.Errorf("processing failed: %w", err)
	}

	logger.Info("media processed successfully",
		slog.String("s3_key", result.S3Key),
		slog.Int64("file_size", result.FileSize))

	// Mark as complete
	if err := w.mediaRepo.MarkComplete(ctx, media.EventID); err != nil {
		logger.Error("failed to mark media as complete",
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to mark complete: %w", err)
	}

	return nil
}

// reconstructMessage reconstructs a proto message from event_outbox payload
// Strategy: Fetch event from outbox → deserialize JSON payload → extract media message
func (w *MediaWorker) reconstructMessage(ctx context.Context, media *persistence.MediaMetadata) (proto.Message, error) {
	// 1. Fetch original event from outbox
	outboxEvent, err := w.outboxRepo.GetEventByID(ctx, media.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event from outbox: %w", err)
	}

	// 2. Deserialize payload to map to identify message structure
	var payload map[string]interface{}
	if err := json.Unmarshal(outboxEvent.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	// 3. Re-marshal specific message field and unmarshal to appropriate proto type
	// The payload structure depends on event type (e.g., events.Message contains Message field)
	var messageData interface{}
	var ok bool

	// Check for common message field names
	if messageData, ok = payload["Message"]; !ok {
		if messageData, ok = payload["message"]; !ok {
			return nil, fmt.Errorf("no message field found in payload for event type: %s", outboxEvent.EventType)
		}
	}

	// Re-serialize message data
	messageJSON, err := json.Marshal(messageData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message data: %w", err)
	}

	// 4. Determine message type from media_type and unmarshal to appropriate proto message
	switch media.MediaType {
	case persistence.MediaTypeImage:
		var msg waE2E.ImageMessage
		if err := json.Unmarshal(messageJSON, &msg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ImageMessage: %w", err)
		}
		return &msg, nil

	case persistence.MediaTypeVideo:
		var msg waE2E.VideoMessage
		if err := json.Unmarshal(messageJSON, &msg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal VideoMessage: %w", err)
		}
		return &msg, nil

	case persistence.MediaTypeAudio, persistence.MediaTypeVoice:
		var msg waE2E.AudioMessage
		if err := json.Unmarshal(messageJSON, &msg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal AudioMessage: %w", err)
		}
		return &msg, nil

	case persistence.MediaTypeDocument:
		var msg waE2E.DocumentMessage
		if err := json.Unmarshal(messageJSON, &msg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal DocumentMessage: %w", err)
		}
		return &msg, nil

	case persistence.MediaTypeSticker:
		var msg waE2E.StickerMessage
		if err := json.Unmarshal(messageJSON, &msg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal StickerMessage: %w", err)
		}
		return &msg, nil

	default:
		return nil, fmt.Errorf("unsupported media type for reconstruction: %s", media.MediaType)
	}
}
