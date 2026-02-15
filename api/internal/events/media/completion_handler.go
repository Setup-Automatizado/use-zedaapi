package media

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"

	natsclient "go.mau.fi/whatsmeow/api/internal/nats"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// CompletionHandler consumes media.done events from NATS.
// It receives notifications when async media processing completes (success or failure).
// This can be used to update webhooks, caches, or trigger downstream actions.
type CompletionHandler struct {
	natsClient *natsclient.Client
	metrics    *observability.Metrics
	log        *slog.Logger
	callback   func(ctx context.Context, result MediaResult)

	consumer jetstream.Consumer
	consCtx  jetstream.ConsumeContext
	cancel   context.CancelFunc
}

// CompletionHandlerConfig holds dependencies for creating a CompletionHandler.
type CompletionHandlerConfig struct {
	NATSClient *natsclient.Client
	Metrics    *observability.Metrics
	Logger     *slog.Logger
	// Callback is invoked for each completed media result.
	// If nil, results are logged but not acted upon.
	Callback func(ctx context.Context, result MediaResult)
}

// NewCompletionHandler creates a new media completion handler.
func NewCompletionHandler(cfg CompletionHandlerConfig) *CompletionHandler {
	return &CompletionHandler{
		natsClient: cfg.NATSClient,
		metrics:    cfg.Metrics,
		log:        cfg.Logger.With(slog.String("component", "media_completion_handler")),
		callback:   cfg.Callback,
	}
}

// Start creates a consumer for media.done.> and begins processing completion events.
func (h *CompletionHandler) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	h.cancel = cancel

	consumerCfg := natsclient.MediaCompletionConsumerConfig()
	consumer, err := h.natsClient.EnsureConsumer(ctx, "MEDIA_PROCESSING", consumerCfg)
	if err != nil {
		cancel()
		return err
	}
	h.consumer = consumer

	consCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		h.handleCompletion(ctx, msg)
	})
	if err != nil {
		cancel()
		return err
	}
	h.consCtx = consCtx

	h.log.Info("media completion handler started")
	return nil
}

// Stop gracefully stops the completion handler.
func (h *CompletionHandler) Stop() {
	if h.consCtx != nil {
		h.consCtx.Stop()
	}
	if h.cancel != nil {
		h.cancel()
	}
	h.log.Info("media completion handler stopped")
}

// handleCompletion processes a single media.done event.
func (h *CompletionHandler) handleCompletion(ctx context.Context, msg jetstream.Msg) {
	var result MediaResult
	if err := json.Unmarshal(msg.Data(), &result); err != nil {
		h.log.Error("failed to unmarshal media result",
			slog.String("error", err.Error()))
		if termErr := msg.Term(); termErr != nil {
			h.log.Error("failed to term malformed media result", slog.String("error", termErr.Error()))
		}
		return
	}

	h.log.Debug("media completion received",
		slog.String("instance_id", result.InstanceID.String()),
		slog.String("event_id", result.EventID.String()),
		slog.Bool("success", result.Success),
		slog.String("media_url", result.MediaURL))

	if h.callback != nil {
		h.callback(ctx, result)
	}

	if err := msg.Ack(); err != nil {
		h.log.Error("failed to ack media completion",
			slog.String("error", err.Error()))
	}
}
