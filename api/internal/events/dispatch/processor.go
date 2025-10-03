package dispatch

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/encoding"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	transformzapi "go.mau.fi/whatsmeow/api/internal/events/transform/zapi"
	"go.mau.fi/whatsmeow/api/internal/events/transport"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// EventProcessor processes individual events through the transform → transport pipeline
type EventProcessor struct {
	instanceID        uuid.UUID
	cfg               *config.Config
	outboxRepo        persistence.OutboxRepository
	dlqRepo           persistence.DLQRepository
	transportRegistry *transport.Registry
	lookup            InstanceLookup
	metrics           *observability.Metrics
}

type transportConfig struct {
	URL     string            `json:"url"`
	Type    string            `json:"type"`
	Headers map[string]string `json:"headers,omitempty"`
}

// InstanceLookup provides per-instance metadata required during dispatch.
type InstanceLookup interface {
	StoreJID(ctx context.Context, instanceID uuid.UUID) (string, error)
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(
	instanceID uuid.UUID,
	cfg *config.Config,
	outboxRepo persistence.OutboxRepository,
	dlqRepo persistence.DLQRepository,
	transportRegistry *transport.Registry,
	lookup InstanceLookup,
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
	}
}

// Process handles the complete processing of a single event
func (p *EventProcessor) Process(ctx context.Context, event *persistence.OutboxEvent) error {
	start := time.Now()
	logger := logging.ContextLogger(ctx, nil)

	logger.Debug("processing event",
		slog.String("event_id", event.EventID.String()),
		slog.String("event_type", event.EventType),
		slog.Int("attempt", event.Attempts+1))

	// Update metrics
	p.metrics.EventsProcessed.WithLabelValues(p.instanceID.String(), event.EventType, "started").Inc()

	// Step 1: Update status to processing
	if err := p.outboxRepo.UpdateEventStatus(ctx, event.EventID, persistence.EventStatusProcessing, event.Attempts, event.NextAttemptAt, event.LastError); err != nil {
		logger.Error("failed to update status to processing",
			slog.String("error", err.Error()))
		return fmt.Errorf("update status: %w", err)
	}

	// Step 2: Transform event (if needed - may already have transformed payload in transport_response)
	var webhookPayload []byte
	var err error

	// Need to transform (we'll use the Payload field which contains the raw event)
	webhookPayload, err = p.transformEvent(ctx, event)
	if err != nil {
		return p.handleTransformError(ctx, event, err)
	}

	// Step 3: Deliver webhook
	deliveryResult, err := p.deliverWebhook(ctx, event, webhookPayload)
	if err != nil {
		return p.handleDeliveryError(ctx, event, err)
	}

	// Step 4: Handle delivery result
	if deliveryResult.Success {
		return p.handleSuccess(ctx, event, start)
	}

	// Delivery failed - determine if retryable
	if deliveryResult.Retryable {
		return p.handleRetryableError(ctx, event, deliveryResult)
	}

	return p.handlePermanentError(ctx, event, deliveryResult)
}

// transformEvent transforms internal event to webhook payload
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

	if event.MediaURL != nil && *event.MediaURL != "" {
		if internalEvent.Metadata == nil {
			internalEvent.Metadata = make(map[string]string)
		}
		internalEvent.Metadata["media_url"] = *event.MediaURL
	}

	connectedPhone := p.connectedPhone(ctx)
	zapiTransformer := transformzapi.NewTransformer(connectedPhone)

	payload, err := zapiTransformer.Transform(ctx, internalEvent)
	if err != nil {
		return nil, err
	}

	logger.Debug("event transformed successfully",
		slog.Int("payload_size", len(payload)))

	return payload, nil
}

// deliverWebhook delivers the webhook to the configured endpoint
func (p *EventProcessor) deliverWebhook(ctx context.Context, event *persistence.OutboxEvent, payload []byte) (*transport.DeliveryResult, error) {
	logger := logging.ContextLogger(ctx, nil)

	// Get transport (default to HTTP)
	trans, err := p.transportRegistry.GetTransport(transport.TransportTypeHTTP)
	if err != nil {
		return nil, fmt.Errorf("get transport: %w", err)
	}

	// Extract webhook URL from transport config
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

	// Prepare delivery request
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

	// Deliver
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

// handleSuccess handles successful event delivery
func (p *EventProcessor) handleSuccess(ctx context.Context, event *persistence.OutboxEvent, startTime time.Time) error {
	logger := logging.ContextLogger(ctx, nil)

	duration := time.Since(startTime)

	// Mark as delivered
	responseJSON, _ := json.Marshal(map[string]interface{}{"status": "delivered", "timestamp": time.Now()})
	if err := p.outboxRepo.MarkDelivered(ctx, event.EventID, responseJSON); err != nil {
		logger.Error("failed to mark as delivered",
			slog.String("error", err.Error()))
		return fmt.Errorf("mark delivered: %w", err)
	}

	logger.Info("event delivered successfully",
		slog.Duration("duration", duration),
		slog.Int("attempt", event.Attempts+1))

	// Update metrics
	p.metrics.EventsProcessed.WithLabelValues(p.instanceID.String(), event.EventType, "success").Inc()
	p.metrics.EventsDelivered.WithLabelValues(p.instanceID.String(), event.EventType).Inc()
	p.metrics.EventDeliveryDuration.WithLabelValues(p.instanceID.String(), event.EventType).Observe(duration.Seconds())

	return nil
}

// handleRetryableError handles retryable delivery failures
func (p *EventProcessor) handleRetryableError(ctx context.Context, event *persistence.OutboxEvent, result *transport.DeliveryResult) error {
	logger := logging.ContextLogger(ctx, nil)

	newAttemptCount := event.Attempts + 1

	// Check if we've exceeded max attempts
	if newAttemptCount >= event.MaxAttempts {
		logger.Warn("max retry attempts exceeded, moving to DLQ",
			slog.Int("attempts", newAttemptCount),
			slog.Int("max_attempts", event.MaxAttempts))

		return p.moveToDLQ(ctx, event, fmt.Sprintf("max retries exceeded: %s", result.ErrorMessage))
	}

	// Calculate next attempt time
	nextAttempt := CalculateNextAttempt(newAttemptCount, p.cfg.Events.RetryDelays)
	errorMsg := result.ErrorMessage

	logger.Warn("delivery failed, will retry",
		slog.Int("attempt", newAttemptCount),
		slog.Time("next_attempt", nextAttempt),
		slog.String("error", errorMsg))

	// Update for retry
	if err := p.outboxRepo.UpdateEventStatus(ctx, event.EventID, persistence.EventStatusRetrying, newAttemptCount, &nextAttempt, &errorMsg); err != nil {
		logger.Error("failed to update for retry",
			slog.String("error", err.Error()))
		return fmt.Errorf("update for retry: %w", err)
	}

	// Update metrics
	p.metrics.EventRetries.WithLabelValues(p.instanceID.String(), event.EventType).Inc()
	p.metrics.EventsProcessed.WithLabelValues(p.instanceID.String(), event.EventType, "retrying").Inc()

	return fmt.Errorf("retryable error: %s", errorMsg)
}

// handlePermanentError handles permanent delivery failures
func (p *EventProcessor) handlePermanentError(ctx context.Context, event *persistence.OutboxEvent, result *transport.DeliveryResult) error {
	logger := logging.ContextLogger(ctx, nil)

	logger.Error("permanent delivery failure, moving to DLQ",
		slog.String("error", result.ErrorMessage),
		slog.Int("status_code", result.StatusCode))

	// Move to DLQ
	if err := p.moveToDLQ(ctx, event, fmt.Sprintf("permanent error: %s (status %d)", result.ErrorMessage, result.StatusCode)); err != nil {
		return err
	}

	// Update metrics
	p.metrics.EventsFailed.WithLabelValues(p.instanceID.String(), event.EventType, "permanent").Inc()

	// Capture in Sentry
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
	return jid
}

// handleTransformError handles transformation failures
func (p *EventProcessor) handleTransformError(ctx context.Context, event *persistence.OutboxEvent, err error) error {
	logger := logging.ContextLogger(ctx, nil)

	logger.Error("transformation failed",
		slog.String("error", err.Error()))

	// Transform errors are typically permanent (malformed event)
	if dlqErr := p.moveToDLQ(ctx, event, fmt.Sprintf("transform error: %v", err)); dlqErr != nil {
		return dlqErr
	}

	// Update metrics
	p.metrics.EventsFailed.WithLabelValues(p.instanceID.String(), event.EventType, "transform_error").Inc()

	// Capture in Sentry
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

// handleDeliveryError handles transport-level delivery errors
func (p *EventProcessor) handleDeliveryError(ctx context.Context, event *persistence.OutboxEvent, err error) error {
	logger := logging.ContextLogger(ctx, nil)

	logger.Error("delivery error",
		slog.String("error", err.Error()))

	// Transport errors are typically retryable (network issues)
	newAttemptCount := event.Attempts + 1

	if newAttemptCount >= event.MaxAttempts {
		return p.moveToDLQ(ctx, event, fmt.Sprintf("max retries exceeded: %v", err))
	}

	nextAttempt := CalculateNextAttempt(newAttemptCount, p.cfg.Events.RetryDelays)
	errorMsg := err.Error()

	if updateErr := p.outboxRepo.UpdateEventStatus(ctx, event.EventID, persistence.EventStatusRetrying, newAttemptCount, &nextAttempt, &errorMsg); updateErr != nil {
		logger.Error("failed to update for retry",
			slog.String("error", updateErr.Error()))
		return fmt.Errorf("update for retry: %w", updateErr)
	}

	// Update metrics
	p.metrics.EventRetries.WithLabelValues(p.instanceID.String(), event.EventType).Inc()

	return fmt.Errorf("delivery error (will retry): %w", err)
}

// moveToDLQ moves an event to the Dead Letter Queue
func (p *EventProcessor) moveToDLQ(ctx context.Context, event *persistence.OutboxEvent, reason string) error {
	logger := logging.ContextLogger(ctx, nil)

	// Create attempt history JSON
	attemptHistory, _ := json.Marshal(map[string]interface{}{
		"total_attempts": event.Attempts,
		"last_error":     event.LastError,
		"moved_at":       time.Now(),
	})

	// Insert into DLQ using existing outbox event
	if err := p.dlqRepo.InsertFromOutbox(ctx, event, reason, attemptHistory); err != nil {
		logger.Error("failed to insert into DLQ",
			slog.String("error", err.Error()))
		return fmt.Errorf("insert DLQ: %w", err)
	}

	// Update outbox status to failed
	errorMsg := reason
	if err := p.outboxRepo.UpdateEventStatus(ctx, event.EventID, persistence.EventStatusFailed, event.Attempts, nil, &errorMsg); err != nil {
		logger.Error("failed to update outbox status to failed",
			slog.String("error", err.Error()))
		// Non-fatal, DLQ entry created
	}

	logger.Info("event moved to DLQ",
		slog.String("reason", reason))

	// Update metrics
	p.metrics.DLQEventsTotal.WithLabelValues(p.instanceID.String(), event.EventType).Inc()
	p.metrics.DLQBacklog.Inc()

	return nil
}
