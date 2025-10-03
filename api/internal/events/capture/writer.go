package capture

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/encoding"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// TransactionalWriter writes events to persistence with batching semantics.
type TransactionalWriter struct {
	log          *slog.Logger
	metrics      *observability.Metrics
	pool         *pgxpool.Pool
	outboxRepo   persistence.OutboxRepository
	mediaRepo    persistence.MediaRepository
	instanceRepo *instances.Repository
	cfg          *config.Config
}

// NewTransactionalWriter constructs a TransactionalWriter.
func NewTransactionalWriter(
	ctx context.Context,
	pool *pgxpool.Pool,
	outboxRepo persistence.OutboxRepository,
	mediaRepo persistence.MediaRepository,
	instanceRepo *instances.Repository,
	cfg *config.Config,
	metrics *observability.Metrics,
) *TransactionalWriter {
	log := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "transactional_writer"),
	)

	return &TransactionalWriter{
		log:          log,
		metrics:      metrics,
		pool:         pool,
		outboxRepo:   outboxRepo,
		mediaRepo:    mediaRepo,
		instanceRepo: instanceRepo,
		cfg:          cfg,
	}
}

// WriteEvents persists the provided internal events. Events that do not have a
// destination webhook configured are skipped gracefully.
func (w *TransactionalWriter) WriteEvents(ctx context.Context, events []*types.InternalEvent) error {
	if len(events) == 0 {
		return nil
	}

	start := time.Now()
	instanceID := events[0].InstanceID

	// Pre-load instance + webhook configuration once per batch.
	instance, err := w.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("load instance %s: %w", instanceID, err)
	}

	webhookConfig, err := w.instanceRepo.GetWebhookConfig(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("load webhook config %s: %w", instanceID, err)
	}

	// Begin transaction scope to keep behaviour similar to previous implementation.
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	successCount := 0
	skippedCount := 0
	failureCount := 0

	maxAttempts := w.cfg.Events.MaxRetryAttempts
	if maxAttempts <= 0 {
		maxAttempts = 6
	}

	for _, internalEvent := range events {
		if instance.StoreJID != nil {
			if internalEvent.Metadata == nil {
				internalEvent.Metadata = make(map[string]string)
			}
			if _, ok := internalEvent.Metadata["store_jid"]; !ok {
				internalEvent.Metadata["store_jid"] = *instance.StoreJID
			}
		}

		// Determine webhook destination.
		transportConfig, shouldDeliver := w.buildTransportConfig(internalEvent, webhookConfig)
		if !shouldDeliver {
			skippedCount++
			w.metrics.EventsInserted.WithLabelValues(
				instanceID.String(),
				internalEvent.EventType,
				"skipped",
			).Inc()
			continue
		}

		encodedPayload, err := encoding.EncodeInternalEvent(internalEvent)
		if err != nil {
			failureCount++
			w.log.ErrorContext(ctx, "encode internal event",
				slog.String("event_id", internalEvent.EventID.String()),
				slog.String("event_type", internalEvent.EventType),
				slog.String("error", err.Error()))
			w.metrics.EventsInserted.WithLabelValues(
				instanceID.String(),
				internalEvent.EventType,
				"encode_failed",
			).Inc()
			continue
		}

		payloadJSON, err := json.Marshal(encodedPayload)
		if err != nil {
			failureCount++
			w.log.ErrorContext(ctx, "marshal encoded payload",
				slog.String("event_id", internalEvent.EventID.String()),
				slog.String("error", err.Error()))
			w.metrics.EventsInserted.WithLabelValues(
				instanceID.String(),
				internalEvent.EventType,
				"encode_failed",
			).Inc()
			continue
		}

		metadataJSON, err := json.Marshal(types.EventMetadata{
			InstanceID: instanceID,
			CapturedAt: internalEvent.CapturedAt,
			SourceLib:  string(internalEvent.SourceLib),
			Extra:      cloneStringMap(internalEvent.Metadata),
		})
		if err != nil {
			failureCount++
			w.log.ErrorContext(ctx, "marshal metadata",
				slog.String("event_id", internalEvent.EventID.String()),
				slog.String("error", err.Error()))
			w.metrics.EventsInserted.WithLabelValues(
				instanceID.String(),
				internalEvent.EventType,
				"metadata_failed",
			).Inc()
			continue
		}

		outboxEvent := &persistence.OutboxEvent{
			InstanceID:        instanceID,
			EventID:           internalEvent.EventID,
			EventType:         internalEvent.EventType,
			SourceLib:         string(internalEvent.SourceLib),
			Payload:           payloadJSON,
			Metadata:          metadataJSON,
			Status:            persistence.EventStatusPending,
			Attempts:          0,
			MaxAttempts:       maxAttempts,
			HasMedia:          internalEvent.HasMedia,
			MediaProcessed:    !internalEvent.HasMedia,
			TransportType:     persistence.TransportWebhook,
			TransportConfig:   transportConfig,
			TransportResponse: nil,
		}

		if err := w.outboxRepo.InsertEvent(ctx, outboxEvent); err != nil {
			failureCount++
			w.log.ErrorContext(ctx, "insert outbox event",
				slog.String("event_id", internalEvent.EventID.String()),
				slog.String("error", err.Error()))
			w.metrics.EventsInserted.WithLabelValues(
				instanceID.String(),
				internalEvent.EventType,
				"insert_failed",
			).Inc()
			continue
		}

		if internalEvent.HasMedia {
			mediaMetadata := &persistence.MediaMetadata{
				EventID:        internalEvent.EventID,
				InstanceID:     internalEvent.InstanceID,
				MediaKey:       internalEvent.MediaKey,
				FileSHA256:     internalEvent.FileSHA256,
				FileEncSHA256:  internalEvent.FileEncSHA256,
				DirectPath:     internalEvent.DirectPath,
				MediaType:      persistence.MediaType(internalEvent.MediaType),
				MimeType:       internalEvent.MimeType,
				FileLength:     internalEvent.FileLength,
				DownloadStatus: persistence.MediaStatusPending,
				MaxRetries:     w.cfg.Events.MediaMaxRetries,
				S3URLType:      persistence.S3URLPresigned,
			}
			if mediaMetadata.MaxRetries <= 0 {
				mediaMetadata.MaxRetries = 3
			}

			if err := w.mediaRepo.InsertMedia(ctx, mediaMetadata); err != nil {
				w.log.ErrorContext(ctx, "insert media metadata",
					slog.String("event_id", internalEvent.EventID.String()),
					slog.String("error", err.Error()))
			}
		}

		successCount++
		w.metrics.EventsInserted.WithLabelValues(
			instanceID.String(),
			internalEvent.EventType,
			"success",
		).Inc()
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	w.log.InfoContext(ctx, "events persisted",
		slog.String("instance_id", instanceID.String()),
		slog.Int("success", successCount),
		slog.Int("skipped", skippedCount),
		slog.Int("failed", failureCount),
		slog.Duration("duration", time.Since(start)))

	w.metrics.EventProcessingDuration.WithLabelValues(
		instanceID.String(),
		"batch_write",
	).Observe(time.Since(start).Seconds())

	return nil
}

// buildTransportConfig returns a JSON encoded transport configuration for the event.
func (w *TransactionalWriter) buildTransportConfig(event *types.InternalEvent, cfg *instances.WebhookConfig) (json.RawMessage, bool) {
	if cfg == nil {
		return nil, false
	}

	url, category := resolveWebhookURL(event, cfg)
	if url == "" {
		return nil, false
	}

	config := struct {
		URL     string            `json:"url"`
		Type    string            `json:"type"`
		Headers map[string]string `json:"headers,omitempty"`
	}{
		URL:  url,
		Type: category,
	}

	raw, err := json.Marshal(config)
	if err != nil {
		w.log.Error("marshal transport config",
			slog.String("event_id", event.EventID.String()),
			slog.String("error", err.Error()))
		return nil, false
	}

	return raw, true
}

func resolveWebhookURL(event *types.InternalEvent, cfg *instances.WebhookConfig) (string, string) {
	if cfg == nil {
		return "", ""
	}

	deref := func(ptr *string) string {
		if ptr == nil {
			return ""
		}
		return *ptr
	}

	switch event.EventType {
	case "message":
		fromMe := event.Metadata["from_me"] == "true"
		if fromMe && cfg.NotifySentByMe == false {
			return "", ""
		}
		return deref(cfg.ReceivedURL), "received"
	case "receipt":
		return deref(cfg.ReceivedDeliveryURL), "receipt"
	case "chat_presence":
		return deref(cfg.ChatPresenceURL), "chat_presence"
	case "presence":
		return deref(cfg.ChatPresenceURL), "presence"
	case "connected":
		return deref(cfg.ConnectedURL), "connected"
	case "disconnected":
		return deref(cfg.DisconnectedURL), "disconnected"
	default:
		return "", ""
	}
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	clone := make(map[string]string, len(input))
	for k, v := range input {
		clone[k] = v
	}
	return clone
}
