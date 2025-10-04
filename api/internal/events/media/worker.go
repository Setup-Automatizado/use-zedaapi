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

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/encoding"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
	whatsmeowevents "go.mau.fi/whatsmeow/types/events"
)

type MediaWorker struct {
	instanceID   uuid.UUID
	client       *whatsmeow.Client
	processor    *MediaProcessor
	mediaRepo    persistence.MediaRepository
	outboxRepo   persistence.OutboxRepository
	pollInterval time.Duration
	batchSize    int
	workerID     string
	metrics      *observability.Metrics
	logger       *slog.Logger
	stopCh       chan struct{}
	doneCh       chan struct{}
}

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

func (w *MediaWorker) Start(ctx context.Context) {
	w.logger.Info("media worker starting")

	go w.run(ctx)
}

func (w *MediaWorker) Stop(ctx context.Context) error {
	w.logger.Info("media worker stopping")

	close(w.stopCh)

	select {
	case <-w.doneCh:
		w.logger.Info("media worker stopped gracefully")
		return nil
	case <-ctx.Done():
		w.logger.Warn("media worker stop timeout")
		return ctx.Err()
	}
}

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
			if err := w.processBatch(ctx); err != nil {
				w.logger.Error("batch processing failed",
					slog.String("error", err.Error()))
			}
		}
	}
}

func (w *MediaWorker) processBatch(ctx context.Context) error {
	mediaItems, err := w.mediaRepo.PollPendingDownloads(ctx, w.batchSize)
	if err != nil {
		w.metrics.MediaFailures.WithLabelValues(w.instanceID.String(), "unknown", "poll").Inc()
		return fmt.Errorf("failed to poll pending media: %w", err)
	}

	if len(mediaItems) == 0 {
		return nil
	}

	w.logger.Debug("processing media batch",
		slog.Int("count", len(mediaItems)))

	w.metrics.MediaBacklog.Set(float64(len(mediaItems)))

	for _, media := range mediaItems {
		if err := w.processMedia(ctx, media); err != nil {
			w.logger.Error("media processing failed",
				slog.String("event_id", media.EventID.String()),
				slog.String("error", err.Error()))
			continue
		}
	}

	return nil
}

func (w *MediaWorker) processMedia(ctx context.Context, media *persistence.MediaMetadata) error {
	logger := w.logger.With(
		slog.String("event_id", media.EventID.String()),
		slog.String("media_type", string(media.MediaType)))

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

	defer func() {
		if err != nil {
			_ = w.mediaRepo.ReleaseFromProcessing(ctx, media.EventID, w.workerID)
		}
	}()

	logger.Info("processing media")

	msg, err := w.reconstructMessage(ctx, media)
	if err != nil {
		logger.Error("failed to reconstruct message",
			slog.String("error", err.Error()))
		w.metrics.MediaFailures.WithLabelValues(w.instanceID.String(), string(media.MediaType), "reconstruct").Inc()
		return fmt.Errorf("failed to reconstruct message: %w", err)
	}

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

	if err := w.mediaRepo.MarkComplete(ctx, media.EventID); err != nil {
		logger.Error("failed to mark media as complete",
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to mark complete: %w", err)
	}

	return nil
}

func (w *MediaWorker) reconstructMessage(ctx context.Context, media *persistence.MediaMetadata) (proto.Message, error) {
	outboxEvent, err := w.outboxRepo.GetEventByID(ctx, media.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event from outbox: %w", err)
	}

	var encoded string
	if err := json.Unmarshal(outboxEvent.Payload, &encoded); err != nil {
		return nil, fmt.Errorf("decode payload string: %w", err)
	}

	internalEvent, err := encoding.DecodeInternalEvent(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode internal event: %w", err)
	}

	msgEvent, ok := internalEvent.RawPayload.(*whatsmeowevents.Message)
	if !ok {
		return nil, fmt.Errorf("event raw payload is not a message: %T", internalEvent.RawPayload)
	}

	if msgEvent.Message == nil {
		return nil, fmt.Errorf("message event missing proto payload")
	}

	switch media.MediaType {
	case persistence.MediaTypeImage:
		if img := msgEvent.Message.GetImageMessage(); img != nil {
			return img, nil
		}
	case persistence.MediaTypeVideo:
		if video := msgEvent.Message.GetVideoMessage(); video != nil {
			return video, nil
		}
	case persistence.MediaTypeAudio, persistence.MediaTypeVoice:
		if audio := msgEvent.Message.GetAudioMessage(); audio != nil {
			return audio, nil
		}
	case persistence.MediaTypeDocument:
		if doc := msgEvent.Message.GetDocumentMessage(); doc != nil {
			return doc, nil
		}
	case persistence.MediaTypeSticker:
		if sticker := msgEvent.Message.GetStickerMessage(); sticker != nil {
			return sticker, nil
		}
	default:
		return nil, fmt.Errorf("unsupported media type for reconstruction: %s", media.MediaType)
	}

	return nil, fmt.Errorf("message payload missing %s content", media.MediaType)
}
