package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"

	natsclient "go.mau.fi/whatsmeow/api/internal/nats"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// NATSWorker processes messages from a NATS consumer for a specific instance.
// It replaces the PostgreSQL-polling Worker with a push-based NATS consumer.
type NATSWorker struct {
	instanceID     uuid.UUID
	client         *natsclient.Client
	clientRegistry ClientRegistry
	processor      MessageProcessor
	dlq            *NATSDLQHandler
	cfg            NATSConfig
	log            *slog.Logger
	metrics        *observability.Metrics
	natsMetrics    *natsclient.NATSMetrics

	// Pause check function (for proxy operations)
	isPaused func(uuid.UUID) bool

	// Cancel check function (for Redis-based cancel set)
	isCancelled func(ctx context.Context, zaapID string) bool

	// Consumer management
	consumer jetstream.Consumer
	consCtx  jetstream.ConsumeContext
	cancel   context.CancelFunc
}

// NATSWorkerConfig holds dependencies for creating a NATSWorker.
type NATSWorkerConfig struct {
	InstanceID     uuid.UUID
	Client         *natsclient.Client
	ClientRegistry ClientRegistry
	Processor      MessageProcessor
	DLQ            *NATSDLQHandler
	Config         NATSConfig
	Logger         *slog.Logger
	Metrics        *observability.Metrics
	NATSMetrics    *natsclient.NATSMetrics
}

// NewNATSWorker creates a new NATS-based worker for a specific instance.
func NewNATSWorker(cfg NATSWorkerConfig) *NATSWorker {
	return &NATSWorker{
		instanceID:     cfg.InstanceID,
		client:         cfg.Client,
		clientRegistry: cfg.ClientRegistry,
		processor:      cfg.Processor,
		dlq:            cfg.DLQ,
		cfg:            cfg.Config,
		log:            cfg.Logger.With(slog.String("instance_id", cfg.InstanceID.String()), slog.String("component", "nats_msg_worker")),
		metrics:        cfg.Metrics,
		natsMetrics:    cfg.NATSMetrics,
	}
}

// SetPauseChecker sets the function used to check if the instance is paused.
func (w *NATSWorker) SetPauseChecker(fn func(uuid.UUID) bool) {
	w.isPaused = fn
}

// SetCancelChecker sets the function used to check if a message was cancelled.
func (w *NATSWorker) SetCancelChecker(fn func(ctx context.Context, zaapID string) bool) {
	w.isCancelled = fn
}

// Start creates the consumer and begins processing messages.
func (w *NATSWorker) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	// Create/update the consumer for this instance
	consumerCfg := natsclient.MessageConsumerConfig(w.instanceID.String())
	consumer, err := w.client.EnsureConsumer(ctx, "MESSAGE_QUEUE", consumerCfg)
	if err != nil {
		cancel()
		return fmt.Errorf("ensure consumer for %s: %w", w.instanceID, err)
	}
	w.consumer = consumer

	// Start consuming with callback
	consCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		w.handleMessage(ctx, msg)
	})
	if err != nil {
		cancel()
		return fmt.Errorf("start consume for %s: %w", w.instanceID, err)
	}
	w.consCtx = consCtx

	w.log.Info("NATS message worker started",
		slog.String("consumer", consumerCfg.Durable),
		slog.String("filter", consumerCfg.FilterSubject),
	)

	return nil
}

// Stop gracefully stops the worker.
func (w *NATSWorker) Stop(ctx context.Context) error {
	if w.consCtx != nil {
		w.consCtx.Stop()
	}
	if w.cancel != nil {
		w.cancel()
	}
	w.log.Info("NATS message worker stopped")
	return nil
}

// handleMessage processes a single NATS message.
func (w *NATSWorker) handleMessage(ctx context.Context, msg jetstream.Msg) {
	start := time.Now()

	// Decode envelope
	var envelope NATSMessageEnvelope
	if err := json.Unmarshal(msg.Data(), &envelope); err != nil {
		w.log.Error("failed to unmarshal envelope",
			slog.String("error", err.Error()))
		// Can't process malformed messages, terminate delivery
		if termErr := msg.Term(); termErr != nil {
			w.log.Error("failed to term malformed message", slog.String("error", termErr.Error()))
		}
		return
	}

	logFields := []any{
		slog.String("zaap_id", envelope.ZaapID),
		slog.String("message_type", msg.Headers().Get(HeaderMessageType)),
	}

	// Check if message is cancelled
	if w.isCancelled != nil && w.isCancelled(ctx, envelope.ZaapID) {
		w.log.Info("message cancelled, acknowledging", logFields...)
		if err := msg.Ack(); err != nil {
			w.log.Error("failed to ack cancelled message", slog.String("error", err.Error()))
		}
		return
	}

	// Check if message is scheduled for the future
	if !envelope.ScheduledAt.IsZero() && time.Now().Before(envelope.ScheduledAt) {
		// NAK with delay until scheduled time
		delay := time.Until(envelope.ScheduledAt)
		w.log.Debug("message scheduled for future, NAK with delay",
			append(logFields, slog.Duration("delay", delay))...)
		if err := msg.NakWithDelay(delay); err != nil {
			w.log.Error("failed to nak scheduled message", slog.String("error", err.Error()))
		}
		return
	}

	// Check if instance is paused for proxy operation
	if w.isPaused != nil && w.isPaused(w.instanceID) {
		w.log.Debug("instance paused, NAK with proxy retry delay", logFields...)
		if err := msg.NakWithDelay(w.cfg.ProxyRetryDelay); err != nil {
			w.log.Error("failed to nak paused message", slog.String("error", err.Error()))
		}
		return
	}

	// Get WhatsApp client
	client, ok := w.clientRegistry.GetClient(w.instanceID.String())
	if !ok {
		// Client not found in this replica's registry. In multi-replica deployments,
		// the client may be on another replica. Use a short NAK (no explicit delay)
		// so NATS BackOff handles escalation: 1s → 5s → 30s → 2m → 5m.
		// This avoids the 30s DisconnectRetryDelay blocking messages on the wrong replica.
		w.log.Debug("whatsapp client not in local registry, NAK for redelivery", logFields...)
		if err := msg.Nak(); err != nil {
			w.log.Error("failed to nak message for redelivery", slog.String("error", err.Error()))
		}
		return
	}

	// Check if client is connected
	if !client.IsConnected() {
		w.log.Warn("whatsapp client disconnected, NAK with disconnect delay", logFields...)
		if err := msg.NakWithDelay(w.cfg.DisconnectRetryDelay); err != nil {
			w.log.Error("failed to nak disconnected message", slog.String("error", err.Error()))
		}
		return
	}

	// Process the message
	err := w.processor.Process(ctx, client, envelope.Payload)
	duration := time.Since(start)

	if err != nil {
		w.log.Error("message processing failed",
			append(logFields,
				slog.String("error", err.Error()),
				slog.Duration("duration", duration),
			)...)

		if w.metrics != nil {
			w.metrics.MessageQueueErrors.WithLabelValues(w.instanceID.String(), "message", "processing_error").Inc()
		}

		// Check delivery count from message metadata
		meta, metaErr := msg.Metadata()
		if metaErr == nil && int(meta.NumDelivered) >= w.cfg.MaxAttempts {
			// Max attempts reached, send to DLQ and terminate
			if w.dlq != nil {
				dlqErr := w.dlq.SendToDLQ(ctx, w.instanceID.String(), msg.Data(), envelope.ZaapID, int(meta.NumDelivered), err.Error())
				if dlqErr != nil {
					w.log.Error("failed to send to DLQ", slog.String("error", dlqErr.Error()))
				}
			}
			if termErr := msg.Term(); termErr != nil {
				w.log.Error("failed to term message after DLQ", slog.String("error", termErr.Error()))
			}
			if w.metrics != nil {
				w.metrics.MessageQueueDLQSize.Inc()
				w.metrics.MessageQueueProcessed.WithLabelValues(w.instanceID.String(), "message", "failed").Inc()
			}
			return
		}

		// NAK for retry (JetStream handles backoff via consumer config)
		if nakErr := msg.Nak(); nakErr != nil {
			w.log.Error("failed to nak message", slog.String("error", nakErr.Error()))
		}
		if w.natsMetrics != nil {
			w.natsMetrics.NakTotal.WithLabelValues("MESSAGE_QUEUE", fmt.Sprintf("msg-%s", w.instanceID.String())).Inc()
		}
		return
	}

	// Success - acknowledge
	if err := msg.Ack(); err != nil {
		w.log.Error("failed to ack message", slog.String("error", err.Error()))
		return
	}

	if w.metrics != nil {
		w.metrics.MessageQueueProcessed.WithLabelValues(w.instanceID.String(), "message", "success").Inc()
		w.metrics.MessageQueueDuration.WithLabelValues(w.instanceID.String(), "message").Observe(duration.Seconds())
	}
	if w.natsMetrics != nil {
		w.natsMetrics.AckTotal.WithLabelValues("MESSAGE_QUEUE", fmt.Sprintf("msg-%s", w.instanceID.String())).Inc()
	}

	w.log.Info("message processed successfully",
		append(logFields, slog.Duration("duration", duration))...)
}
