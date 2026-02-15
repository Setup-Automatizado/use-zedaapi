package media

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"

	"go.mau.fi/whatsmeow/api/internal/observability"
)

// FastPathProcessor attempts to process media inline within a timeout.
// If processing completes within the timeout, the media URL is available immediately.
// If it times out, the media task is published to NATS for async processing.
type FastPathProcessor struct {
	processor *MediaProcessor
	publisher *NATSMediaPublisher
	timeout   time.Duration
	metrics   *observability.Metrics
	log       *slog.Logger
}

// NewFastPathProcessor creates a new fast path processor.
func NewFastPathProcessor(
	processor *MediaProcessor,
	publisher *NATSMediaPublisher,
	timeout time.Duration,
	metrics *observability.Metrics,
	log *slog.Logger,
) *FastPathProcessor {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &FastPathProcessor{
		processor: processor,
		publisher: publisher,
		timeout:   timeout,
		metrics:   metrics,
		log:       log.With(slog.String("component", "media_fast_path")),
	}
}

// TryFastPath attempts to process media within the timeout.
// Returns a FastPathResult indicating success or timeout.
// On timeout, the media task is automatically queued for async processing.
func (f *FastPathProcessor) TryFastPath(
	ctx context.Context,
	client *whatsmeow.Client,
	instanceID uuid.UUID,
	eventID uuid.UUID,
	msg proto.Message,
	mediaKey string,
	mediaType string,
	encodedPayload string,
) FastPathResult {
	start := time.Now()

	// Create deadline context for fast path
	fastCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	// Channel for result
	resultCh := make(chan FastPathResult, 1)

	go func() {
		result, err := f.processor.ProcessWithRetry(fastCtx, client, instanceID, eventID, msg, mediaKey)
		if err != nil {
			resultCh <- FastPathResult{Success: false, Error: err}
			return
		}
		resultCh <- FastPathResult{Success: true, MediaURL: result.S3URL}
	}()

	select {
	case result := <-resultCh:
		duration := time.Since(start)
		if result.Success {
			f.log.Debug("media fast path succeeded",
				slog.String("instance_id", instanceID.String()),
				slog.String("event_id", eventID.String()),
				slog.String("media_url", result.MediaURL),
				slog.Duration("duration", duration))
			if f.metrics != nil {
				f.metrics.MediaDownloadsTotal.WithLabelValues(instanceID.String(), mediaType, "success").Inc()
				f.metrics.MediaDownloadDuration.WithLabelValues(instanceID.String(), mediaType).Observe(duration.Seconds())
			}
			return result
		}
		// Fast path failed but within timeout - queue for async
		f.log.Warn("media fast path failed, queuing async",
			slog.String("instance_id", instanceID.String()),
			slog.String("event_id", eventID.String()),
			slog.String("error", result.Error.Error()))

	case <-fastCtx.Done():
		// Timeout - queue for async processing
		f.log.Debug("media fast path timeout, queuing async",
			slog.String("instance_id", instanceID.String()),
			slog.String("event_id", eventID.String()),
			slog.Duration("timeout", f.timeout))
	}

	// Queue for async processing via NATS
	if f.publisher != nil {
		task := MediaTask{
			InstanceID:  instanceID,
			EventID:     eventID,
			MediaKey:    mediaKey,
			MediaType:   mediaType,
			PublishedAt: time.Now(),
			Payload:     encodedPayload,
		}
		if err := f.publisher.PublishTask(ctx, task); err != nil {
			f.log.Error("failed to queue media task",
				slog.String("instance_id", instanceID.String()),
				slog.String("event_id", eventID.String()),
				slog.String("error", err.Error()))
			return FastPathResult{Success: false, Error: fmt.Errorf("fast path timeout and queue failed: %w", err)}
		}
	}

	return FastPathResult{Success: false, Error: fmt.Errorf("fast path timeout, queued for async")}
}
