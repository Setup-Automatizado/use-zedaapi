package dispatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/encoding"
	eventsnats "go.mau.fi/whatsmeow/api/internal/events/nats"
	"go.mau.fi/whatsmeow/api/internal/events/pollstore"
	transformzapi "go.mau.fi/whatsmeow/api/internal/events/transform/zapi"
	"go.mau.fi/whatsmeow/api/internal/events/transport"
	natsclient "go.mau.fi/whatsmeow/api/internal/nats"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// permanentDeliveryError signals that a webhook delivery failed permanently
// and should NOT be retried (e.g., HTTP 404, 403, 400).
type permanentDeliveryError struct {
	msg        string
	statusCode int
}

func (e *permanentDeliveryError) Error() string {
	return e.msg
}

// WebhookResolver resolves webhook URLs for an instance at consume time.
// Using an interface allows the NATSDispatchWorker to be testable.
type WebhookResolver interface {
	ResolveWebhook(ctx context.Context, instanceID uuid.UUID) (*ResolvedWebhook, error)
}

// ResolvedWebhook holds the resolved webhook configuration for an instance.
type ResolvedWebhook struct {
	DeliveryURL         string
	ReceivedURL         string
	ReceivedDeliveryURL string
	MessageStatusURL    string
	DisconnectedURL     string
	ChatPresenceURL     string
	ConnectedURL        string
	HistorySyncURL      string
	NotifySentByMe      bool
	StoreJID            *string
	ClientToken         string
	IsBusiness          bool
}

// MediaResultLookup checks if media processing has completed for an event.
// Used by the dispatch worker to wait for media URLs before delivering webhooks.
type MediaResultLookup interface {
	LookupMediaURL(ctx context.Context, eventID string) (url string, done bool, err error)
}

// NATSDispatchWorker processes events from a NATS consumer and delivers webhooks.
// It consolidates logic from both TransactionalWriter (URL resolution, filtering)
// and EventProcessor (Z-API transform, HTTP delivery, retry/DLQ).
type NATSDispatchWorker struct {
	instanceID        uuid.UUID
	client            *natsclient.Client
	cfg               *config.Config
	eventCfg          eventsnats.NATSEventConfig
	transportRegistry *transport.Registry
	webhookResolver   WebhookResolver
	pollStore         pollstore.Store
	mediaResults      MediaResultLookup
	metrics           *observability.Metrics
	natsMetrics       *natsclient.NATSMetrics
	statusInterceptor StatusInterceptor
	dlqHandler        *NATSEventDLQHandler
	log               *slog.Logger

	// Consumer management
	consumer jetstream.Consumer
	consCtx  jetstream.ConsumeContext
	cancel   context.CancelFunc
}

// NATSDispatchWorkerConfig holds dependencies for creating a NATSDispatchWorker.
type NATSDispatchWorkerConfig struct {
	InstanceID        uuid.UUID
	Client            *natsclient.Client
	Config            *config.Config
	EventConfig       eventsnats.NATSEventConfig
	TransportRegistry *transport.Registry
	WebhookResolver   WebhookResolver
	PollStore         pollstore.Store
	MediaResults      MediaResultLookup
	Metrics           *observability.Metrics
	NATSMetrics       *natsclient.NATSMetrics
	DLQHandler        *NATSEventDLQHandler
	Logger            *slog.Logger
}

// NewNATSDispatchWorker creates a new NATS-based event dispatch worker.
func NewNATSDispatchWorker(cfg NATSDispatchWorkerConfig) *NATSDispatchWorker {
	return &NATSDispatchWorker{
		instanceID:        cfg.InstanceID,
		client:            cfg.Client,
		cfg:               cfg.Config,
		eventCfg:          cfg.EventConfig,
		transportRegistry: cfg.TransportRegistry,
		webhookResolver:   cfg.WebhookResolver,
		pollStore:         cfg.PollStore,
		mediaResults:      cfg.MediaResults,
		metrics:           cfg.Metrics,
		natsMetrics:       cfg.NATSMetrics,
		dlqHandler:        cfg.DLQHandler,
		log:               cfg.Logger.With(slog.String("instance_id", cfg.InstanceID.String()), slog.String("component", "nats_dispatch_worker")),
	}
}

// SetStatusInterceptor sets the status cache interceptor.
func (w *NATSDispatchWorker) SetStatusInterceptor(interceptor StatusInterceptor) {
	w.statusInterceptor = interceptor
}

// Start creates the consumer and begins processing events.
func (w *NATSDispatchWorker) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	// Create/update the consumer for this instance
	consumerCfg := natsclient.EventConsumerConfig(w.instanceID.String())
	consumer, err := w.client.EnsureConsumer(ctx, "WHATSAPP_EVENTS", consumerCfg)
	if err != nil {
		cancel()
		return fmt.Errorf("ensure event consumer for %s: %w", w.instanceID, err)
	}
	w.consumer = consumer

	consCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		w.handleMessage(ctx, msg)
	})
	if err != nil {
		cancel()
		return fmt.Errorf("start event consume for %s: %w", w.instanceID, err)
	}
	w.consCtx = consCtx

	w.log.Info("NATS dispatch worker started",
		slog.String("consumer", consumerCfg.Durable),
		slog.String("filter", consumerCfg.FilterSubject))

	return nil
}

// Stop gracefully stops the worker.
func (w *NATSDispatchWorker) Stop(ctx context.Context) error {
	if w.consCtx != nil {
		w.consCtx.Stop()
	}
	if w.cancel != nil {
		w.cancel()
	}
	w.log.Info("NATS dispatch worker stopped")
	return nil
}

// handleMessage processes a single NATS event message.
func (w *NATSDispatchWorker) handleMessage(ctx context.Context, msg jetstream.Msg) {
	start := time.Now()

	// Decode envelope
	var envelope eventsnats.NATSEventEnvelope
	if err := json.Unmarshal(msg.Data(), &envelope); err != nil {
		w.log.Error("failed to unmarshal event envelope",
			slog.String("error", err.Error()))
		if termErr := msg.Term(); termErr != nil {
			w.log.Error("failed to term malformed event", slog.String("error", termErr.Error()))
		}
		return
	}

	logFields := []any{
		slog.String("event_id", envelope.EventID.String()),
		slog.String("event_type", envelope.EventType),
	}

	// Resolve webhook configuration at consume time
	webhook, err := w.webhookResolver.ResolveWebhook(ctx, w.instanceID)
	if err != nil {
		w.log.Error("failed to resolve webhook config",
			append(logFields, slog.String("error", err.Error()))...)
		// NAK for retry - webhook config might become available
		if nakErr := msg.NakWithDelay(10 * time.Second); nakErr != nil {
			w.log.Error("failed to nak event", slog.String("error", nakErr.Error()))
		}
		return
	}

	// Determine webhook URL based on event type
	webhookURL := w.resolveWebhookURL(webhook, envelope.EventType, envelope.Metadata)
	if webhookURL == "" {
		// Check if status cache should handle this
		if w.statusInterceptor != nil && w.statusInterceptor.ShouldIntercept(envelope.EventType) {
			w.handleStatusCacheEvent(ctx, msg, envelope, webhook, start)
			return
		}

		// No webhook URL configured for this event type
		w.log.Debug("no webhook URL for event type, acknowledging",
			append(logFields, slog.String("event_type", envelope.EventType))...)
		if err := msg.Ack(); err != nil {
			w.log.Error("failed to ack no-url event", slog.String("error", err.Error()))
		}
		return
	}

	// Decode the internal event from payload
	internalEvent, err := encoding.DecodeInternalEvent(envelope.Payload)
	if err != nil {
		w.log.Error("failed to decode internal event",
			append(logFields, slog.String("error", err.Error()))...)
		if termErr := msg.Term(); termErr != nil {
			w.log.Error("failed to term decode-failed event", slog.String("error", termErr.Error()))
		}
		return
	}

	// Wait for media result if event has media, then inject media_url into metadata.
	// This mirrors the PG path where events are held until media_processed=true.
	if envelope.HasMedia && w.mediaResults != nil {
		mediaURL, done := w.waitForMediaResult(ctx, msg, envelope.EventID.String())
		if done && mediaURL != "" {
			if internalEvent.Metadata == nil {
				internalEvent.Metadata = make(map[string]string)
			}
			internalEvent.Metadata["media_url"] = mediaURL
			w.log.Debug("media URL injected into event",
				slog.String("event_id", envelope.EventID.String()),
				slog.String("media_url", mediaURL))
		} else if !done {
			// waitForMediaResult already handled NAK/retry
			return
		}
	}

	// Enrich with storeJID metadata
	if webhook.StoreJID != nil {
		if internalEvent.Metadata == nil {
			internalEvent.Metadata = make(map[string]string)
		}
		internalEvent.Metadata["store_jid"] = *webhook.StoreJID
	}

	// Transform to Z-API format
	connectedPhone := ""
	if webhook.StoreJID != nil {
		connectedPhone = *webhook.StoreJID
	}
	zapiTransformer := transformzapi.NewTransformer(connectedPhone, webhook.IsBusiness, false, "", w.pollStore)
	webhookPayload, err := zapiTransformer.Transform(ctx, internalEvent)
	if err != nil {
		w.log.Error("failed to transform event",
			append(logFields, slog.String("error", err.Error()))...)
		if termErr := msg.Term(); termErr != nil {
			w.log.Error("failed to term transform-failed event", slog.String("error", termErr.Error()))
		}
		return
	}

	// Status cache interception
	if w.statusInterceptor != nil && w.statusInterceptor.ShouldIntercept(envelope.EventType) {
		suppress, interceptErr := w.statusInterceptor.InterceptAndCache(ctx, w.instanceID.String(), envelope.EventType, webhookPayload)
		if interceptErr != nil {
			w.log.Warn("status cache intercept error, continuing with delivery",
				slog.String("error", interceptErr.Error()))
		} else if suppress {
			if err := msg.Ack(); err != nil {
				w.log.Error("failed to ack cached event", slog.String("error", err.Error()))
			}
			if w.metrics != nil {
				w.metrics.EventsProcessed.WithLabelValues(w.instanceID.String(), envelope.EventType, "cached").Inc()
			}
			return
		}
	}

	// Get delivery attempt count from NATS metadata
	attempt := 1
	meta, metaErr := msg.Metadata()
	if metaErr == nil {
		attempt = int(meta.NumDelivered)
	}

	// Deliver webhook via HTTP transport
	err = w.deliverWebhook(ctx, webhookURL, webhook.ClientToken, webhookPayload, envelope, attempt)
	duration := time.Since(start)

	if err != nil {
		w.log.Error("webhook delivery failed",
			append(logFields,
				slog.String("error", err.Error()),
				slog.Int("attempt", attempt),
				slog.Duration("duration", duration))...)

		if w.metrics != nil {
			w.metrics.EventsProcessed.WithLabelValues(w.instanceID.String(), envelope.EventType, "failed").Inc()
		}

		// Check if this is a permanent (non-retryable) error - immediately DLQ + Term
		var permErr *permanentDeliveryError
		if errors.As(err, &permErr) {
			w.log.Warn("permanent delivery error, sending to DLQ immediately",
				append(logFields,
					slog.Int("status_code", permErr.statusCode),
					slog.Int("attempt", attempt))...)
			if w.dlqHandler != nil {
				dlqErr := w.dlqHandler.SendToDLQ(ctx, w.instanceID.String(), msg.Data(), envelope.EventID.String(), attempt, err.Error())
				if dlqErr != nil {
					w.log.Error("failed to send event to DLQ", slog.String("error", dlqErr.Error()))
				}
			}
			if termErr := msg.Term(); termErr != nil {
				w.log.Error("failed to term permanently failed event", slog.String("error", termErr.Error()))
			}
			return
		}

		// Retryable error - check if max attempts reached
		if metaErr == nil && attempt >= w.eventCfg.MaxAttempts {
			w.log.Warn("max delivery attempts reached, sending to DLQ",
				append(logFields, slog.Int("attempts", attempt))...)
			if w.dlqHandler != nil {
				dlqErr := w.dlqHandler.SendToDLQ(ctx, w.instanceID.String(), msg.Data(), envelope.EventID.String(), attempt, err.Error())
				if dlqErr != nil {
					w.log.Error("failed to send event to DLQ", slog.String("error", dlqErr.Error()))
				}
			}
			if termErr := msg.Term(); termErr != nil {
				w.log.Error("failed to term event after DLQ", slog.String("error", termErr.Error()))
			}
			return
		}

		// NAK for retry
		if nakErr := msg.Nak(); nakErr != nil {
			w.log.Error("failed to nak event", slog.String("error", nakErr.Error()))
		}
		if w.natsMetrics != nil {
			w.natsMetrics.NakTotal.WithLabelValues("WHATSAPP_EVENTS", fmt.Sprintf("evt-%s", w.instanceID.String())).Inc()
		}
		return
	}

	// Success - acknowledge
	if err := msg.Ack(); err != nil {
		w.log.Error("failed to ack event", slog.String("error", err.Error()))
		return
	}

	if w.metrics != nil {
		w.metrics.EventsProcessed.WithLabelValues(w.instanceID.String(), envelope.EventType, "delivered").Inc()
		w.metrics.EventDeliveryDuration.WithLabelValues(w.instanceID.String(), envelope.EventType, "nats").Observe(duration.Seconds())
	}
	if w.natsMetrics != nil {
		w.natsMetrics.AckTotal.WithLabelValues("WHATSAPP_EVENTS", fmt.Sprintf("evt-%s", w.instanceID.String())).Inc()
	}

	w.log.Debug("event delivered successfully",
		append(logFields, slog.Duration("duration", duration))...)
}

// resolveWebhookURL determines the correct webhook URL based on event type.
// This logic MUST match capture/writer.go resolveWebhookURL exactly to ensure
// identical webhook routing behavior between PG and NATS modes.
func (w *NATSDispatchWorker) resolveWebhookURL(webhook *ResolvedWebhook, eventType string, metadata map[string]string) string {
	if webhook == nil {
		return ""
	}

	switch eventType {
	case "message":
		fromMe := metadata["from_me"] == "true"
		fromAPI := metadata["from_api"] == "true"

		// API echo events are ALWAYS routed regardless of NotifySentByMe setting.
		// This ensures the partner always receives confirmation of their API calls.
		if fromAPI {
			if webhook.ReceivedDeliveryURL != "" {
				return webhook.ReceivedDeliveryURL
			}
			return webhook.ReceivedURL
		}

		// When NotifySentByMe is enabled:
		// - If receivedAndDeliveryCallbackUrl is configured, send ALL messages there
		// - If NOT configured, fall back to individual webhooks
		if webhook.NotifySentByMe {
			if webhook.ReceivedDeliveryURL != "" {
				return webhook.ReceivedDeliveryURL
			}
			if fromMe {
				return webhook.DeliveryURL
			}
			return webhook.ReceivedURL
		}

		// When NotifySentByMe is disabled, use SEPARATE routing:
		// - Messages SENT by me -> delivery_url, fall back to received_delivery_url
		// - Messages RECEIVED from others -> received_delivery_url, fall back to received_url
		if fromMe {
			if webhook.DeliveryURL != "" {
				return webhook.DeliveryURL
			}
			if webhook.ReceivedDeliveryURL != "" {
				return webhook.ReceivedDeliveryURL
			}
			return ""
		}

		if webhook.ReceivedDeliveryURL != "" {
			return webhook.ReceivedDeliveryURL
		}
		return webhook.ReceivedURL

	case "receipt":
		return webhook.MessageStatusURL

	case "undecryptable", "group_info", "group_joined", "picture":
		if webhook.ReceivedDeliveryURL != "" {
			return webhook.ReceivedDeliveryURL
		}
		return webhook.ReceivedURL

	case "chat_presence", "presence":
		return webhook.ChatPresenceURL

	case "connected":
		return webhook.ConnectedURL

	case "disconnected":
		return webhook.DisconnectedURL

	case "history_sync":
		return webhook.HistorySyncURL

	default:
		return ""
	}
}

// deliverWebhook sends the webhook payload via HTTP transport.
// Returns a *permanentDeliveryError for non-retryable failures (e.g., HTTP 404, 403)
// so that callers can distinguish permanent vs transient failures.
func (w *NATSDispatchWorker) deliverWebhook(ctx context.Context, url, clientToken string, payload []byte, envelope eventsnats.NATSEventEnvelope, attempt int) error {
	trans, err := w.transportRegistry.GetTransport(transport.TransportTypeHTTP)
	if err != nil {
		return fmt.Errorf("get transport: %w", err)
	}

	headers := map[string]string{}
	if clientToken != "" {
		headers["Client-Token"] = clientToken
	}

	request := &transport.DeliveryRequest{
		Endpoint:    url,
		Payload:     payload,
		Headers:     headers,
		EventID:     envelope.EventID.String(),
		EventType:   envelope.EventType,
		InstanceID:  w.instanceID.String(),
		Attempt:     attempt,
		MaxAttempts: w.eventCfg.MaxAttempts,
	}

	result, err := trans.Deliver(ctx, request)
	if err != nil {
		return fmt.Errorf("deliver: %w", err)
	}

	if !result.Success {
		errMsg := result.ErrorMessage
		if errMsg == "" {
			errMsg = fmt.Sprintf("HTTP %d", result.StatusCode)
		}
		if !result.Retryable {
			// Permanent error - capture in Sentry and return typed error
			sentry.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("instance_id", w.instanceID.String())
				scope.SetTag("event_type", envelope.EventType)
				scope.SetTag("status_code", fmt.Sprintf("%d", result.StatusCode))
				sentry.CaptureMessage(fmt.Sprintf("permanent webhook delivery failure: %s", errMsg))
			})
			return &permanentDeliveryError{msg: errMsg, statusCode: result.StatusCode}
		}
		return fmt.Errorf("delivery failed (retryable): %s", errMsg)
	}

	return nil
}

// handleStatusCacheEvent processes events that should go to status cache only.
func (w *NATSDispatchWorker) handleStatusCacheEvent(ctx context.Context, msg jetstream.Msg, envelope eventsnats.NATSEventEnvelope, webhook *ResolvedWebhook, start time.Time) {
	// Decode and transform for status cache
	internalEvent, err := encoding.DecodeInternalEvent(envelope.Payload)
	if err != nil {
		w.log.Error("failed to decode event for status cache",
			slog.String("error", err.Error()))
		if termErr := msg.Term(); termErr != nil {
			w.log.Error("failed to term decode-failed event", slog.String("error", termErr.Error()))
		}
		return
	}

	connectedPhone := ""
	if webhook.StoreJID != nil {
		connectedPhone = *webhook.StoreJID
	}
	zapiTransformer := transformzapi.NewTransformer(connectedPhone, webhook.IsBusiness, false, "", w.pollStore)
	webhookPayload, err := zapiTransformer.Transform(ctx, internalEvent)
	if err != nil {
		w.log.Error("failed to transform for status cache", slog.String("error", err.Error()))
		if termErr := msg.Term(); termErr != nil {
			w.log.Error("failed to term transform-failed event", slog.String("error", termErr.Error()))
		}
		return
	}

	suppress, interceptErr := w.statusInterceptor.InterceptAndCache(ctx, w.instanceID.String(), envelope.EventType, webhookPayload)
	if interceptErr != nil {
		w.log.Warn("status cache intercept error", slog.String("error", interceptErr.Error()))
	}
	_ = suppress // Whether suppressed or not, we ack since there's no webhook URL

	if err := msg.Ack(); err != nil {
		w.log.Error("failed to ack status-cache event", slog.String("error", err.Error()))
	}
	if w.metrics != nil {
		w.metrics.EventsProcessed.WithLabelValues(w.instanceID.String(), envelope.EventType, "cached").Inc()
	}
}

// waitForMediaResult polls the media result KV store until the media processing
// result is available or a timeout is reached. It calls msg.InProgress() periodically
// to extend the ack deadline so the message is not redelivered while waiting.
//
// Returns (url, true) when the result is found (url may be empty if processing failed).
// Returns ("", false) when the wait should be aborted (ctx cancelled or unrecoverable error).
// Returns ("", true) on timeout - the event should be delivered without media URL (degraded mode).
func (w *NATSDispatchWorker) waitForMediaResult(ctx context.Context, msg jetstream.Msg, eventID string) (string, bool) {
	const maxWait = 2 * time.Minute
	const pollInterval = 3 * time.Second
	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		url, done, err := w.mediaResults.LookupMediaURL(ctx, eventID)
		if err != nil {
			w.log.Error("media result lookup error",
				slog.String("event_id", eventID),
				slog.String("error", err.Error()))
			// Transient error - keep retrying within the loop
		}
		if done {
			if url != "" {
				w.log.Info("media result found",
					slog.String("event_id", eventID),
					slog.String("media_url", url))
			} else {
				w.log.Warn("media processing failed, delivering without media URL",
					slog.String("event_id", eventID))
			}
			return url, true
		}

		// Extend ack deadline to prevent auto-redelivery while waiting
		if progressErr := msg.InProgress(); progressErr != nil {
			w.log.Error("failed to extend ack deadline",
				slog.String("event_id", eventID),
				slog.String("error", progressErr.Error()))
			return "", false
		}

		select {
		case <-ctx.Done():
			return "", false
		case <-time.After(pollInterval):
		}
	}

	w.log.Warn("media result timeout, delivering without media URL",
		slog.String("event_id", eventID),
		slog.Duration("waited", maxWait))
	return "", true
}

// NATSEventDLQHandler publishes failed events to the DLQ stream.
type NATSEventDLQHandler struct {
	client *natsclient.Client
	log    *slog.Logger
}

// NewNATSEventDLQHandler creates a new event DLQ handler.
func NewNATSEventDLQHandler(client *natsclient.Client, log *slog.Logger) *NATSEventDLQHandler {
	return &NATSEventDLQHandler{
		client: client,
		log:    log.With(slog.String("component", "nats_event_dlq")),
	}
}

// NATSEventDLQEntry represents a failed event.
type NATSEventDLQEntry struct {
	EventID    string          `json:"event_id"`
	InstanceID string          `json:"instance_id"`
	Envelope   json.RawMessage `json:"envelope"`
	Error      string          `json:"error"`
	Attempts   int             `json:"attempts"`
	FailedAt   time.Time       `json:"failed_at"`
}

// SendToDLQ publishes a failed event to dlq.events.{instance_id}.
func (h *NATSEventDLQHandler) SendToDLQ(ctx context.Context, instanceID string, envelope json.RawMessage, eventID string, attempts int, errorMsg string) error {
	entry := NATSEventDLQEntry{
		EventID:    eventID,
		InstanceID: instanceID,
		Envelope:   envelope,
		Error:      errorMsg,
		Attempts:   attempts,
		FailedAt:   time.Now(),
	}

	data, merr := json.Marshal(entry)
	if merr != nil {
		return fmt.Errorf("marshal event DLQ entry: %w", merr)
	}

	subject := fmt.Sprintf("dlq.events.%s", instanceID)
	msg := &natsgo.Msg{
		Subject: subject,
		Data:    data,
	}

	// Stable MsgID based on eventID ensures idempotent DLQ publish on retry
	dlqMsgID := fmt.Sprintf("dlq-evt-%s", eventID)
	_, err := h.client.PublishMsg(ctx, msg, jetstream.WithMsgID(dlqMsgID))
	if err != nil {
		return fmt.Errorf("publish to event DLQ %s: %w", subject, err)
	}

	h.log.Warn("event sent to DLQ",
		slog.String("instance_id", instanceID),
		slog.String("event_id", eventID),
		slog.Int("attempts", attempts),
		slog.String("error", errorMsg))

	return nil
}
