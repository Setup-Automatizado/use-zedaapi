package capture

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/events/transform"
	whatsmeowtransform "go.mau.fi/whatsmeow/api/internal/events/transform/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

type EventHandler struct {
	log         *slog.Logger
	metrics     *observability.Metrics
	router      *EventRouter
	instanceID  uuid.UUID
	transformer transform.SourceTransformer

	mu      sync.RWMutex
	stopped bool
	stopCh  chan struct{}
}

func NewEventHandler(
	ctx context.Context,
	instanceID uuid.UUID,
	router *EventRouter,
	metrics *observability.Metrics,
) *EventHandler {
	log := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "event_handler"),
		slog.String("instance_id", instanceID.String()),
	)

	sourceTransformer := whatsmeowtransform.NewTransformer(instanceID)

	return &EventHandler{
		log:         log,
		metrics:     metrics,
		router:      router,
		instanceID:  instanceID,
		transformer: sourceTransformer,
		stopCh:      make(chan struct{}),
	}
}

func (h *EventHandler) HandleEvent(ctx context.Context, rawEvent interface{}) error {
	h.mu.RLock()
	if h.stopped {
		h.mu.RUnlock()
		return fmt.Errorf("event handler stopped")
	}
	h.mu.RUnlock()

	internalEvent, err := h.transformer.Transform(ctx, rawEvent)
	if errors.Is(err, transform.ErrUnsupportedEvent) {
		h.log.DebugContext(ctx, "unsupported event type skipped",
			slog.String("type", fmt.Sprintf("%T", rawEvent)))
		return nil
	}
	if err != nil {
		h.metrics.EventsCaptured.WithLabelValues(
			h.instanceID.String(),
			"unsupported",
			string(types.SourceLibWhatsmeow),
		).Inc()
		return fmt.Errorf("transform event: %w", err)
	}

	h.metrics.EventsCaptured.WithLabelValues(
		h.instanceID.String(),
		internalEvent.EventType,
		string(internalEvent.SourceLib),
	).Inc()

	h.log.DebugContext(ctx, "event captured",
		slog.String("event_id", internalEvent.EventID.String()),
		slog.String("event_type", internalEvent.EventType),
		slog.Bool("has_media", internalEvent.HasMedia),
	)

	return h.router.RouteEvent(ctx, internalEvent)
}

func (h *EventHandler) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.stopped {
		h.stopped = true
		close(h.stopCh)
	}
}

func (h *EventHandler) IsStopped() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.stopped
}
