package capture

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/encoding"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// TransactionalWriter writes events to persistence with batching semantics.
type TransactionalWriter struct {
	log              *slog.Logger
	metrics          *observability.Metrics
	pool             *pgxpool.Pool
	outboxRepo       persistence.OutboxRepository
	mediaRepo        persistence.MediaRepository
	cfg              *config.Config
	resolver         WebhookResolver
	metadataEnricher MetadataEnricher
}

// WebhookResolver resolves transport configuration for an instance.
type WebhookResolver interface {
	Resolve(ctx context.Context, instanceID uuid.UUID) (*ResolvedWebhookConfig, error)
}

// MetadataEnricher allows caller-specific metadata augmentation before persistence.
type MetadataEnricher func(cfg *ResolvedWebhookConfig, event *types.InternalEvent)

// ResolvedWebhookConfig contains the subset of webhook configuration required by the writer.
type ResolvedWebhookConfig struct {
	DeliveryURL         string
	ReceivedURL         string
	ReceivedDeliveryURL string
	MessageStatusURL    string
	DisconnectedURL     string
	ChatPresenceURL     string
	ConnectedURL        string
	NotifySentByMe      bool
	StoreJID            *string
}

// NewTransactionalWriter constructs a TransactionalWriter.
func NewTransactionalWriter(
	ctx context.Context,
	pool *pgxpool.Pool,
	outboxRepo persistence.OutboxRepository,
	mediaRepo persistence.MediaRepository,
	resolver WebhookResolver,
	metadataEnricher MetadataEnricher,
	cfg *config.Config,
	metrics *observability.Metrics,
) *TransactionalWriter {
	log := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "transactional_writer"),
	)

	return &TransactionalWriter{
		log:              log,
		metrics:          metrics,
		pool:             pool,
		outboxRepo:       outboxRepo,
		mediaRepo:        mediaRepo,
		cfg:              cfg,
		resolver:         resolver,
		metadataEnricher: metadataEnricher,
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

	resolvedConfig, err := w.resolver.Resolve(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("resolve webhook config %s: %w", instanceID, err)
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
		if w.metadataEnricher != nil {
			w.metadataEnricher(resolvedConfig, internalEvent)
		}

		// Determine webhook destination.
		transportConfig, shouldDeliver := w.buildTransportConfig(internalEvent, resolvedConfig)
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
func (w *TransactionalWriter) buildTransportConfig(event *types.InternalEvent, cfg *ResolvedWebhookConfig) (json.RawMessage, bool) {
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

func resolveWebhookURL(event *types.InternalEvent, cfg *ResolvedWebhookConfig) (string, string) {
	if cfg == nil {
		return "", ""
	}

	switch event.EventType {
	case "message":
		fromMe := event.Metadata["from_me"] == "true"
		if fromMe && !cfg.NotifySentByMe {
			return "", ""
		}
		return cfg.ReceivedURL, "received"
	case "receipt":
		return cfg.ReceivedDeliveryURL, "receipt"
	case "chat_presence":
		return cfg.ChatPresenceURL, "chat_presence"
	case "presence":
		return cfg.ChatPresenceURL, "presence"
	case "connected":
		return cfg.ConnectedURL, "connected"
	case "disconnected":
		return cfg.DisconnectedURL, "disconnected"
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
