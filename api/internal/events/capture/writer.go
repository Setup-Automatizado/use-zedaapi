package capture

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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

type TransactionalWriter struct {
	log              *slog.Logger
	metrics          *observability.Metrics
	pool             *pgxpool.Pool
	outboxRepo       persistence.OutboxRepository
	mediaRepo        persistence.MediaRepository
	cfg              *config.Config
	resolver         WebhookResolver
	metadataEnricher MetadataEnricher
	debugDumpDir     string
}

type WebhookResolver interface {
	Resolve(ctx context.Context, instanceID uuid.UUID) (*ResolvedWebhookConfig, error)
}

type MetadataEnricher func(cfg *ResolvedWebhookConfig, event *types.InternalEvent)

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

	dumpDir := "./tmp/debug-events"
	if cfg != nil && strings.TrimSpace(cfg.Events.DebugDumpDir) != "" {
		dumpDir = strings.TrimSpace(cfg.Events.DebugDumpDir)
	}

	return &TransactionalWriter{
		log:              log,
		metrics:          metrics,
		pool:             pool,
		outboxRepo:       outboxRepo,
		mediaRepo:        mediaRepo,
		cfg:              cfg,
		resolver:         resolver,
		metadataEnricher: metadataEnricher,
		debugDumpDir:     dumpDir,
	}
}

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

		encodedPayload, err := encoding.EncodeInternalEvent(internalEvent)
		if err != nil {
			failureCount++
			w.log.ErrorContext(ctx, "encode internal event",
				slog.String("event_id", internalEvent.EventID.String()),
				slog.String("event_type", internalEvent.EventType),
				slog.String("error", err.Error()))
			w.dumpEventDebug(internalEvent, "encode_internal_event", nil, map[string]string{"error": err.Error()})
			w.metrics.EventsInserted.WithLabelValues(
				instanceID.String(),
				internalEvent.EventType,
				"encode_failed",
			).Inc()
			continue
		}

		decodedPayload, _ := base64.StdEncoding.DecodeString(encodedPayload)
		w.debugLogInternalEvent(ctx, internalEvent, encodedPayload, decodedPayload)
		if w.cfg != nil && w.cfg.Events.DebugRawPayload {
			extra := map[string]string{"encoded_payload": encodedPayload}
			w.dumpEventDebug(internalEvent, "internal_event_captured", decodedPayload, extra)
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

		transportConfig, shouldDeliver := w.buildTransportConfig(internalEvent, resolvedConfig)

		// When StatusCache is enabled, receipt events should be persisted even without webhook
		// The StatusCache interceptor in the processor will handle caching and suppress webhook delivery
		statusCacheWillHandle := w.cfg != nil && w.cfg.StatusCache.Enabled && internalEvent.EventType == "receipt"

		if !shouldDeliver && !statusCacheWillHandle {
			skippedCount++
			w.log.WarnContext(ctx, "skipping event without webhook destination",
				slog.String("instance_id", instanceID.String()),
				slog.String("event_id", internalEvent.EventID.String()),
				slog.String("event_type", internalEvent.EventType))
			w.dumpEventDebug(internalEvent, "no_webhook_destination", decodedPayload, nil)
			w.metrics.EventsInserted.WithLabelValues(
				instanceID.String(),
				internalEvent.EventType,
				"skipped",
			).Inc()
			continue
		}

		// For StatusCache-handled receipts without webhook, create a minimal transport config
		if statusCacheWillHandle && !shouldDeliver {
			w.log.DebugContext(ctx, "persisting receipt for status cache (no webhook configured)",
				slog.String("instance_id", instanceID.String()),
				slog.String("event_id", internalEvent.EventID.String()))
			// Create transport config that marks this for status cache only
			config := struct {
				URL        string `json:"url"`
				Type       string `json:"type"`
				NoDelivery bool   `json:"no_delivery"`
			}{
				URL:        "",
				Type:       "status_cache_only",
				NoDelivery: true,
			}
			var err error
			transportConfig, err = json.Marshal(config)
			if err != nil {
				w.log.Error("marshal status cache transport config",
					slog.String("event_id", internalEvent.EventID.String()),
					slog.String("error", err.Error()))
				continue
			}
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

	if w.cfg != nil && strings.TrimSpace(w.cfg.Client.AuthToken) != "" {
		config.Headers = map[string]string{
			"Client-Token": w.cfg.Client.AuthToken,
		}
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
		fromAPI := event.Metadata["from_api"] == "true"

		// API echo events are ALWAYS routed regardless of NotifySentByMe setting
		// This ensures the partner always receives confirmation of their API calls
		if fromAPI {
			// Prefer ReceivedDeliveryURL for API echo, fall back to ReceivedURL
			if cfg.ReceivedDeliveryURL != "" {
				return cfg.ReceivedDeliveryURL, "received"
			}
			return cfg.ReceivedURL, "received"
		}

		// When NotifySentByMe is enabled, ALL messages go to combined endpoint
		// This matches Z-API behavior: messages (received + sent by me) go to receivedAndDeliveryCallbackUrl
		if cfg.NotifySentByMe {
			if cfg.ReceivedDeliveryURL != "" {
				return cfg.ReceivedDeliveryURL, "received"
			}
			// Fallback to ReceivedURL for Z-API compatibility
			return cfg.ReceivedURL, "received"
		}

		// When NotifySentByMe is disabled, use SEPARATE routing:
		// - Messages SENT by me -> delivery_url
		// - Messages RECEIVED from others -> received_url
		if fromMe {
			// Messages sent by me go to delivery_url
			if cfg.DeliveryURL != "" {
				return cfg.DeliveryURL, "delivery"
			}
			// If delivery_url not configured, filter the event
			return "", ""
		}

		// Messages received from others go to received_url
		return cfg.ReceivedURL, "received"

	case "receipt":
		// Receipt events (message status: sent, delivered, read, played) go ONLY to message_status_url
		// If not configured, the event is discarded (no fallback to other webhooks)
		if cfg.MessageStatusURL != "" {
			return cfg.MessageStatusURL, "message_status"
		}
		return "", ""
	case "undecryptable":
		return cfg.ReceivedURL, "received"
	case "group_info":
		return cfg.ReceivedURL, "received"
	case "group_joined":
		return cfg.ReceivedURL, "received"
	case "picture":
		return cfg.ReceivedURL, "received"
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

func (w *TransactionalWriter) debugLogInternalEvent(ctx context.Context, event *types.InternalEvent, encoded string, decoded []byte) {
	if w.cfg == nil || !w.cfg.Events.DebugRawPayload {
		return
	}

	attrs := []any{
		slog.String("instance_id", event.InstanceID.String()),
		slog.String("event_id", event.EventID.String()),
		slog.String("event_type", event.EventType),
		slog.String("payload_base64", encoded),
	}

	if len(decoded) > 0 {
		attrs = append(attrs, slog.String("payload_json", string(decoded)))
	}

	w.log.DebugContext(ctx, "internal event payload", attrs...)
}

func (w *TransactionalWriter) dumpEventDebug(event *types.InternalEvent, reason string, payloadJSON []byte, extra map[string]string) {
	if w.debugDumpDir == "" {
		return
	}

	if err := os.MkdirAll(w.debugDumpDir, 0o755); err != nil {
		w.log.Warn("failed to create debug dump directory",
			slog.String("dir", w.debugDumpDir),
			slog.String("error", err.Error()))
		return
	}

	dump := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339Nano),
		"reason":      reason,
		"event_id":    event.EventID.String(),
		"event_type":  event.EventType,
		"instance_id": event.InstanceID.String(),
		"metadata":    cloneStringMap(event.Metadata),
	}

	if payloadJSON != nil {
		dump["payload_json"] = json.RawMessage(payloadJSON)
	}

	if event.RawPayload != nil {
		dump["raw_payload_type"] = fmt.Sprintf("%T", event.RawPayload)
		dump["raw_payload_repr"] = fmt.Sprintf("%+v", event.RawPayload)
	}

	if extra != nil {
		for k, v := range extra {
			dump[k] = v
		}
	}

	data, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		data = []byte(fmt.Sprintf("%+v", dump))
	}

	fileName := fmt.Sprintf("%s_%s_%d.json", event.EventType, event.EventID.String(), time.Now().UnixNano())
	filePath := filepath.Join(w.debugDumpDir, fileName)

	if writeErr := os.WriteFile(filePath, data, 0o644); writeErr != nil {
		w.log.Warn("failed to write debug dump",
			slog.String("path", filePath),
			slog.String("error", writeErr.Error()))
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
