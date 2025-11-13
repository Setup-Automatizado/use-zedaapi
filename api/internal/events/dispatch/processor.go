package dispatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/encoding"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/events/pollstore"
	transformzapi "go.mau.fi/whatsmeow/api/internal/events/transform/zapi"
	"go.mau.fi/whatsmeow/api/internal/events/transport"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

const (
	historySyncEventType  = "history_sync"
	skipReasonHistorySync = "history_sync_temporarily_ignored"
)

type skippedEventError struct {
	Reason string
}

func (e *skippedEventError) Error() string {
	return fmt.Sprintf("event skipped: %s", e.Reason)
}

type EventProcessor struct {
	instanceID        uuid.UUID
	cfg               *config.Config
	outboxRepo        persistence.OutboxRepository
	dlqRepo           persistence.DLQRepository
	transportRegistry *transport.Registry
	lookup            InstanceLookup
	metrics           *observability.Metrics
	pollStore         pollstore.Store
}

type transportConfig struct {
	URL     string            `json:"url"`
	Type    string            `json:"type"`
	Headers map[string]string `json:"headers,omitempty"`
}

type InstanceLookup interface {
	StoreJID(ctx context.Context, instanceID uuid.UUID) (string, error)
}

func NewEventProcessor(
	instanceID uuid.UUID,
	cfg *config.Config,
	outboxRepo persistence.OutboxRepository,
	dlqRepo persistence.DLQRepository,
	transportRegistry *transport.Registry,
	lookup InstanceLookup,
	pollStore pollstore.Store,
	metrics *observability.Metrics,
) *EventProcessor {
	return &EventProcessor{
		instanceID:        instanceID,
		cfg:               cfg,
		outboxRepo:        outboxRepo,
		dlqRepo:           dlqRepo,
		transportRegistry: transportRegistry,
		lookup:            lookup,
		metrics:           metrics,
		pollStore:         pollStore,
	}
}

func (p *EventProcessor) Process(ctx context.Context, event *persistence.OutboxEvent) error {
	start := time.Now()
	logger := logging.ContextLogger(ctx, nil)

	logger.Debug("processing event",
		slog.String("event_id", event.EventID.String()),
		slog.String("event_type", event.EventType),
		slog.Int("attempt", event.Attempts+1))

	p.metrics.EventsProcessed.WithLabelValues(p.instanceID.String(), event.EventType, "started").Inc()

	if err := p.outboxRepo.UpdateEventStatus(ctx, event.EventID, persistence.EventStatusProcessing, event.Attempts, event.NextAttemptAt, event.LastError); err != nil {
		logger.Error("failed to update status to processing",
			slog.String("error", err.Error()))
		return fmt.Errorf("update status: %w", err)
	}

	var webhookPayload []byte
	var err error

	webhookPayload, err = p.transformEvent(ctx, event)
	if err != nil {
		var skippedErr *skippedEventError
		if errors.As(err, &skippedErr) {
			return p.handleSkippedEvent(ctx, event, skippedErr.Reason)
		}
		return p.handleTransformError(ctx, event, err)
	}

	deliveryResult, err := p.deliverWebhook(ctx, event, webhookPayload)
	if err != nil {
		return p.handleDeliveryError(ctx, event, err)
	}

	if deliveryResult.Success {
		return p.handleSuccess(ctx, event, start)
	}

	if deliveryResult.Retryable {
		return p.handleRetryableError(ctx, event, deliveryResult)
	}

	return p.handlePermanentError(ctx, event, deliveryResult)
}

func (p *EventProcessor) transformEvent(ctx context.Context, event *persistence.OutboxEvent) ([]byte, error) {
	logger := logging.ContextLogger(ctx, nil)

	var encoded string
	if err := json.Unmarshal(event.Payload, &encoded); err != nil {
		return nil, fmt.Errorf("decode payload string: %w", err)
	}

	internalEvent, err := encoding.DecodeInternalEvent(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode internal event: %w", err)
	}

	// TODO: Temporary debug logging for schema verification - remove after validation
	if debugPayload, _ := json.Marshal(internalEvent); len(debugPayload) > 0 {
		logger.Debug("debug internal event payload (temporary)", slog.String("payload", string(debugPayload)))
	}

	if internalEvent.EventType == historySyncEventType {
		logger.Warn("skipping history sync event",
			slog.String("event_id", event.EventID.String()),
			slog.String("reason", skipReasonHistorySync))
		return nil, &skippedEventError{Reason: skipReasonHistorySync}
	}

	if event.MediaURL != nil && *event.MediaURL != "" {
		if internalEvent.Metadata == nil {
			internalEvent.Metadata = make(map[string]string)
		}
		internalEvent.Metadata["media_url"] = *event.MediaURL
	}

	connectedPhone := p.connectedPhone(ctx)
	debugRaw := false
	dumpDir := ""
	if p.cfg != nil {
		debugRaw = p.cfg.Events.DebugRawPayload
		dumpDir = p.cfg.Events.DebugDumpDir
	}
	zapiTransformer := transformzapi.NewTransformer(connectedPhone, debugRaw, dumpDir, p.pollStore)

	payload, err := zapiTransformer.Transform(ctx, internalEvent)
	if err != nil {
		return nil, err
	}

	// TODO: Temporary debug logging for schema verification - remove after validation
	logger.Debug("debug zapi payload (temporary)", slog.String("payload", string(payload)))

	logger.Debug("event transformed successfully",
		slog.Int("payload_size", len(payload)))

	return payload, nil
}

func (p *EventProcessor) deliverWebhook(ctx context.Context, event *persistence.OutboxEvent, payload []byte) (*transport.DeliveryResult, error) {
	logger := logging.ContextLogger(ctx, nil)

	trans, err := p.transportRegistry.GetTransport(transport.TransportTypeHTTP)
	if err != nil {
		return nil, fmt.Errorf("get transport: %w", err)
	}

	var cfg transportConfig
	if len(event.TransportConfig) == 0 {
		return nil, fmt.Errorf("missing transport configuration")
	}
	if err := json.Unmarshal(event.TransportConfig, &cfg); err != nil {
		return nil, fmt.Errorf("parse transport config: %w", err)
	}
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, fmt.Errorf("empty webhook url")
	}

	request := &transport.DeliveryRequest{
		Endpoint:    cfg.URL,
		Payload:     payload,
		Headers:     cfg.Headers,
		EventID:     event.EventID.String(),
		EventType:   event.EventType,
		InstanceID:  p.instanceID.String(),
		Attempt:     event.Attempts + 1,
		MaxAttempts: event.MaxAttempts,
	}

	logger.Debug("delivering webhook",
		slog.String("endpoint", cfg.URL),
		slog.Int("attempt", request.Attempt))

	result, err := trans.Deliver(ctx, request)
	if err != nil {
		logger.Error("delivery transport error",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("deliver: %w", err)
	}

	return result, nil
}

func (p *EventProcessor) handleSuccess(ctx context.Context, event *persistence.OutboxEvent, startTime time.Time) error {
	logger := logging.ContextLogger(ctx, nil)

	duration := time.Since(startTime)

	responseJSON, _ := json.Marshal(map[string]interface{}{"status": "delivered", "timestamp": time.Now()})
	if err := p.outboxRepo.MarkDelivered(ctx, event.EventID, responseJSON); err != nil {
		logger.Error("failed to mark as delivered",
			slog.String("error", err.Error()))
		return fmt.Errorf("mark delivered: %w", err)
	}

	transportLabel := string(event.TransportType)
	if transportLabel == "" {
		transportLabel = string(persistence.TransportWebhook)
	}

	logger.Info("event delivered successfully",
		slog.Duration("duration", duration),
		slog.Int("attempt", event.Attempts+1),
		slog.String("transport", transportLabel))

	// Update metrics
	p.metrics.EventsProcessed.WithLabelValues(p.instanceID.String(), event.EventType, "success").Inc()
	p.metrics.EventsDelivered.WithLabelValues(p.instanceID.String(), event.EventType, transportLabel).Inc()
	p.metrics.EventDeliveryDuration.WithLabelValues(p.instanceID.String(), event.EventType, transportLabel).Observe(duration.Seconds())

	return nil
}

func (p *EventProcessor) handleRetryableError(ctx context.Context, event *persistence.OutboxEvent, result *transport.DeliveryResult) error {
	logger := logging.ContextLogger(ctx, nil)

	newAttemptCount := event.Attempts + 1

	if newAttemptCount >= event.MaxAttempts {
		logger.Warn("max retry attempts exceeded, moving to DLQ",
			slog.Int("attempts", newAttemptCount),
			slog.Int("max_attempts", event.MaxAttempts))

		return p.moveToDLQ(ctx, event, fmt.Sprintf("max retries exceeded: %s", result.ErrorMessage), "max_retries_exceeded")
	}

	nextAttempt := CalculateNextAttempt(newAttemptCount, p.cfg.Events.RetryDelays)
	errorMsg := result.ErrorMessage

	logger.Warn("delivery failed, will retry",
		slog.Int("attempt", newAttemptCount),
		slog.Time("next_attempt", nextAttempt),
		slog.String("error", errorMsg))

	if err := p.outboxRepo.UpdateEventStatus(ctx, event.EventID, persistence.EventStatusRetrying, newAttemptCount, &nextAttempt, &errorMsg); err != nil {
		logger.Error("failed to update for retry",
			slog.String("error", err.Error()))
		return fmt.Errorf("update for retry: %w", err)
	}

	p.metrics.EventRetries.WithLabelValues(p.instanceID.String(), event.EventType).Inc()
	p.metrics.EventsProcessed.WithLabelValues(p.instanceID.String(), event.EventType, "retrying").Inc()

	return fmt.Errorf("retryable error: %s", errorMsg)
}

func (p *EventProcessor) handlePermanentError(ctx context.Context, event *persistence.OutboxEvent, result *transport.DeliveryResult) error {
	logger := logging.ContextLogger(ctx, nil)

	logger.Error("permanent delivery failure, moving to DLQ",
		slog.String("error", result.ErrorMessage),
		slog.Int("status_code", result.StatusCode))

	if err := p.moveToDLQ(ctx, event, fmt.Sprintf("permanent error: %s (status %d)", result.ErrorMessage, result.StatusCode), "permanent_error"); err != nil {
		return err
	}

	p.metrics.EventsFailed.WithLabelValues(p.instanceID.String(), event.EventType, "permanent").Inc()

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("component", "dispatch")
		scope.SetTag("instance_id", p.instanceID.String())
		scope.SetTag("event_type", event.EventType)
		scope.SetContext("delivery", map[string]interface{}{
			"event_id":    event.EventID.String(),
			"status_code": result.StatusCode,
		})
		sentry.CaptureMessage(fmt.Sprintf("Permanent webhook delivery failure: %s", result.ErrorMessage))
	})

	return fmt.Errorf("permanent error: %s", result.ErrorMessage)
}

func (p *EventProcessor) connectedPhone(ctx context.Context) string {
	if p.lookup == nil {
		return ""
	}
	storeJID, err := p.lookup.StoreJID(ctx, p.instanceID)
	if err != nil {
		return ""
	}
	if storeJID == "" {
		return ""
	}
	jid := storeJID
	if idx := strings.IndexRune(jid, '@'); idx >= 0 {
		jid = jid[:idx]
	}
	if colon := strings.IndexRune(jid, ':'); colon >= 0 {
		jid = jid[:colon]
	}
	return jid
}

func (p *EventProcessor) handleTransformError(ctx context.Context, event *persistence.OutboxEvent, err error) error {
	logger := logging.ContextLogger(ctx, nil)

	logger.Error("transformation failed",
		slog.String("error", err.Error()))

	if dlqErr := p.moveToDLQ(ctx, event, fmt.Sprintf("transform error: %v", err), "transform_error"); dlqErr != nil {
		return dlqErr
	}

	p.metrics.EventsFailed.WithLabelValues(p.instanceID.String(), event.EventType, "transform_error").Inc()

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("component", "dispatch")
		scope.SetTag("instance_id", p.instanceID.String())
		scope.SetTag("event_type", event.EventType)
		scope.SetContext("event", map[string]interface{}{
			"event_id": event.EventID.String(),
		})
		sentry.CaptureException(err)
	})

	return fmt.Errorf("transform error: %w", err)
}

func (p *EventProcessor) handleDeliveryError(ctx context.Context, event *persistence.OutboxEvent, err error) error {
	logger := logging.ContextLogger(ctx, nil)

	logger.Error("delivery error",
		slog.String("error", err.Error()))

	newAttemptCount := event.Attempts + 1

	if newAttemptCount >= event.MaxAttempts {
		return p.moveToDLQ(ctx, event, fmt.Sprintf("max retries exceeded: %v", err), "max_retries_exceeded")
	}

	nextAttempt := CalculateNextAttempt(newAttemptCount, p.cfg.Events.RetryDelays)
	errorMsg := err.Error()

	if updateErr := p.outboxRepo.UpdateEventStatus(ctx, event.EventID, persistence.EventStatusRetrying, newAttemptCount, &nextAttempt, &errorMsg); updateErr != nil {
		logger.Error("failed to update for retry",
			slog.String("error", updateErr.Error()))
		return fmt.Errorf("update for retry: %w", updateErr)
	}

	p.metrics.EventRetries.WithLabelValues(p.instanceID.String(), event.EventType).Inc()

	return fmt.Errorf("delivery error (will retry): %w", err)
}

func dlqReasonLabel(reason string) string {
	lowerReason := strings.ToLower(reason)
	if strings.HasPrefix(lowerReason, "transform error") {
		return "transform_error"
	}
	if strings.HasPrefix(lowerReason, "permanent error") {
		return "permanent_error"
	}
	if strings.HasPrefix(lowerReason, "max retries exceeded") {
		return "max_retries_exceeded"
	}
	if strings.HasPrefix(lowerReason, "delivery error") {
		return "delivery_error"
	}
	return "unknown"
}

func (p *EventProcessor) moveToDLQ(ctx context.Context, event *persistence.OutboxEvent, reason string, reasonLabel string) error {
	logger := logging.ContextLogger(ctx, nil)

	if reasonLabel == "" {
		reasonLabel = dlqReasonLabel(reason)
	}

	attemptHistory, _ := json.Marshal(map[string]interface{}{
		"total_attempts": event.Attempts,
		"last_error":     event.LastError,
		"moved_at":       time.Now(),
	})

	if err := p.dlqRepo.InsertFromOutbox(ctx, event, reason, attemptHistory); err != nil {
		logger.Error("failed to insert into DLQ",
			slog.String("error", err.Error()))
		return fmt.Errorf("insert DLQ: %w", err)
	}

	errorMsg := reason
	if err := p.outboxRepo.UpdateEventStatus(ctx, event.EventID, persistence.EventStatusFailed, event.Attempts, nil, &errorMsg); err != nil {
		logger.Error("failed to update outbox status to failed",
			slog.String("error", err.Error()))
	}

	logger.Info("event moved to DLQ",
		slog.String("reason", reason))

	p.metrics.DLQEventsTotal.WithLabelValues(p.instanceID.String(), event.EventType, reasonLabel).Inc()
	p.metrics.DLQBacklog.Inc()

	return nil
}

func (p *EventProcessor) handleSkippedEvent(ctx context.Context, event *persistence.OutboxEvent, reason string) error {
	logger := logging.ContextLogger(ctx, nil)

	logger.Info("event skipped",
		slog.String("event_id", event.EventID.String()),
		slog.String("event_type", event.EventType),
		slog.String("reason", reason))

	responseJSON, _ := json.Marshal(map[string]interface{}{
		"status":    "skipped",
		"reason":    reason,
		"timestamp": time.Now(),
	})

	if err := p.outboxRepo.MarkDelivered(ctx, event.EventID, responseJSON); err != nil {
		logger.Error("failed to mark skipped event as delivered",
			slog.String("error", err.Error()))
		return fmt.Errorf("mark skipped event as delivered: %w", err)
	}

	p.metrics.EventsProcessed.WithLabelValues(p.instanceID.String(), event.EventType, "skipped").Inc()

	return nil
}
