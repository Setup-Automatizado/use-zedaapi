package capture

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/encoding"
	eventsnats "go.mau.fi/whatsmeow/api/internal/events/nats"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// MediaTaskPublisher publishes media processing tasks for async handling.
// This interface decouples capture from media package, allowing the NATSEventWriter
// to queue media tasks without importing the media package directly.
type MediaTaskPublisher interface {
	PublishMediaTask(ctx context.Context, info MediaTaskInfo) error
}

// MediaTaskInfo holds the information needed to publish a media processing task.
type MediaTaskInfo struct {
	InstanceID uuid.UUID
	EventID    uuid.UUID
	MediaKey   string
	DirectPath string
	MediaType  string
	MimeType   string
	FileLength int64
	Payload    string // Encoded InternalEvent
}

// NATSEventWriter implements EventWriter by publishing events to NATS JetStream.
// It replaces TransactionalWriter for the NATS-based event pipeline.
//
// Key difference from TransactionalWriter: webhook URL resolution is deferred
// to consume time (in NATSDispatchWorker), not write time. This means webhook
// URL changes take effect immediately without reprocessing events.
//
// When a MediaTaskPublisher is set, events with HasMedia=true also get a media
// task published for async processing by NATSMediaWorker.
type NATSEventWriter struct {
	log            *slog.Logger
	metrics        *observability.Metrics
	publisher      *eventsnats.EventPublisher
	cfg            *config.Config
	mediaPublisher MediaTaskPublisher
}

// NewNATSEventWriter creates a new NATS-based event writer.
func NewNATSEventWriter(
	publisher *eventsnats.EventPublisher,
	cfg *config.Config,
	metrics *observability.Metrics,
	log *slog.Logger,
) *NATSEventWriter {
	return &NATSEventWriter{
		log:       log.With(slog.String("component", "nats_event_writer")),
		metrics:   metrics,
		publisher: publisher,
		cfg:       cfg,
	}
}

// SetMediaTaskPublisher sets an optional media task publisher.
// When set, events with HasMedia=true will also have a media task published
// for async processing by the NATS media worker.
func (w *NATSEventWriter) SetMediaTaskPublisher(publisher MediaTaskPublisher) {
	w.mediaPublisher = publisher
}

// WriteEvents encodes and publishes InternalEvents to NATS.
// Implements the EventWriter interface used by EventBuffer.
func (w *NATSEventWriter) WriteEvents(ctx context.Context, events []*types.InternalEvent) error {
	if len(events) == 0 {
		return nil
	}

	instanceID := events[0].InstanceID
	start := time.Now()
	successCount := 0
	skippedCount := 0
	failureCount := 0

	for _, event := range events {
		// Apply event filters (same as TransactionalWriter)
		if w.cfg != nil && w.cfg.EventFilters.FilterWaitingMessage && event.Metadata["waiting_message"] == "true" {
			skippedCount++
			w.log.Debug("filtering waitingMessage=true event",
				slog.String("instance_id", instanceID.String()),
				slog.String("event_id", event.EventID.String()),
				slog.String("event_type", event.EventType))
			if w.metrics != nil {
				w.metrics.EventsInserted.WithLabelValues(
					instanceID.String(),
					event.EventType,
					"filtered_waiting_message",
				).Inc()
			}
			continue
		}

		if w.cfg != nil && w.cfg.EventFilters.FilterSecondaryDeviceReceipts && event.EventType == "receipt" && event.Metadata["sender_device"] != "" {
			skippedCount++
			w.log.Debug("filtering secondary device receipt",
				slog.String("instance_id", instanceID.String()),
				slog.String("event_id", event.EventID.String()))
			if w.metrics != nil {
				w.metrics.EventsInserted.WithLabelValues(
					instanceID.String(),
					event.EventType,
					"filtered_secondary_device",
				).Inc()
			}
			continue
		}

		// Encode the internal event
		encodedPayload, err := encoding.EncodeInternalEvent(event)
		if err != nil {
			failureCount++
			w.log.Error("encode internal event",
				slog.String("event_id", event.EventID.String()),
				slog.String("event_type", event.EventType),
				slog.String("error", err.Error()))
			if w.metrics != nil {
				w.metrics.EventsInserted.WithLabelValues(
					instanceID.String(),
					event.EventType,
					"encode_failed",
				).Inc()
			}
			continue
		}

		// Build media info if present
		var mediaInfo *eventsnats.MediaInfo
		if event.HasMedia {
			mediaInfo = &eventsnats.MediaInfo{
				MediaKey:      event.MediaKey,
				DirectPath:    event.DirectPath,
				FileSHA256:    event.FileSHA256,
				FileEncSHA256: event.FileEncSHA256,
				MediaType:     event.MediaType,
				MimeType:      event.MimeType,
				FileLength:    event.FileLength,
			}
		}

		// Build envelope
		envelope := eventsnats.NATSEventEnvelope{
			EventID:     event.EventID,
			InstanceID:  event.InstanceID,
			EventType:   event.EventType,
			SourceLib:   string(event.SourceLib),
			CapturedAt:  event.CapturedAt,
			PublishedAt: time.Now(),
			HasMedia:    event.HasMedia,
			MediaInfo:   mediaInfo,
			Payload:     encodedPayload,
			Metadata:    event.Metadata,
		}

		if err := w.publisher.Publish(ctx, envelope); err != nil {
			failureCount++
			w.log.Error("publish event to NATS",
				slog.String("event_id", event.EventID.String()),
				slog.String("event_type", event.EventType),
				slog.String("error", err.Error()))
			if w.metrics != nil {
				w.metrics.EventsInserted.WithLabelValues(
					instanceID.String(),
					event.EventType,
					"nats_publish_failed",
				).Inc()
			}
			continue
		}

		successCount++
		if w.metrics != nil {
			w.metrics.EventsInserted.WithLabelValues(
				instanceID.String(),
				event.EventType,
				"nats_published",
			).Inc()
		}

		// Publish media task for async processing if event has media
		if event.HasMedia && w.mediaPublisher != nil {
			mimeType := ""
			if event.MimeType != nil {
				mimeType = *event.MimeType
			}
			var fileLength int64
			if event.FileLength != nil {
				fileLength = *event.FileLength
			}

			// DEBUG: Log all media fields at capture time
			w.log.Info("publishing media task - capture debug",
				slog.String("event_id", event.EventID.String()),
				slog.String("instance_id", event.InstanceID.String()),
				slog.String("media_type", event.MediaType),
				slog.String("media_key", event.MediaKey),
				slog.String("direct_path", event.DirectPath),
				slog.String("mime_type", mimeType),
				slog.Int64("file_length", fileLength),
				slog.Bool("has_file_sha256", event.FileSHA256 != nil && *event.FileSHA256 != ""),
				slog.Bool("has_file_enc_sha256", event.FileEncSHA256 != nil && *event.FileEncSHA256 != ""),
				slog.Int("payload_len", len(encodedPayload)),
			)

			if err := w.mediaPublisher.PublishMediaTask(ctx, MediaTaskInfo{
				InstanceID: event.InstanceID,
				EventID:    event.EventID,
				MediaKey:   event.MediaKey,
				DirectPath: event.DirectPath,
				MediaType:  event.MediaType,
				MimeType:   mimeType,
				FileLength: fileLength,
				Payload:    encodedPayload,
			}); err != nil {
				w.log.Error("failed to publish media task (event still published)",
					slog.String("event_id", event.EventID.String()),
					slog.String("media_type", event.MediaType),
					slog.String("error", err.Error()))
			} else {
				w.log.Debug("media task published for async processing",
					slog.String("event_id", event.EventID.String()),
					slog.String("media_type", event.MediaType))
			}
		}
	}

	duration := time.Since(start)
	w.log.Debug("batch write complete",
		slog.String("instance_id", instanceID.String()),
		slog.Int("total", len(events)),
		slog.Int("success", successCount),
		slog.Int("skipped", skippedCount),
		slog.Int("failed", failureCount),
		slog.Duration("duration", duration),
	)

	// Return error to trigger retry in buffer if any events failed.
	// Already-published events are safe to retry thanks to JetStream MsgID deduplication.
	if failureCount > 0 {
		return fmt.Errorf("%d of %d events failed to publish", failureCount, len(events))
	}

	return nil
}

// Verify NATSEventWriter implements EventWriter at compile time.
var _ EventWriter = (*NATSEventWriter)(nil)
