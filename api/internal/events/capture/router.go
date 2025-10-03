package capture

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// EventRouter routes events from handlers to instance-specific buffers
type EventRouter struct {
	log     *slog.Logger
	metrics *observability.Metrics

	mu      sync.RWMutex
	buffers map[uuid.UUID]*EventBuffer
}

// NewEventRouter creates a new EventRouter
func NewEventRouter(ctx context.Context, metrics *observability.Metrics) *EventRouter {
	log := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "event_router"),
	)

	return &EventRouter{
		log:     log,
		metrics: metrics,
		buffers: make(map[uuid.UUID]*EventBuffer),
	}
}

// RegisterBuffer registers a buffer for an instance
func (r *EventRouter) RegisterBuffer(instanceID uuid.UUID, buffer *EventBuffer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.buffers[instanceID] = buffer
	r.log.Info("buffer registered",
		slog.String("instance_id", instanceID.String()),
	)
}

// UnregisterBuffer removes a buffer for an instance
func (r *EventRouter) UnregisterBuffer(instanceID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if buffer, ok := r.buffers[instanceID]; ok {
		buffer.Stop()
		delete(r.buffers, instanceID)
		r.log.Info("buffer unregistered",
			slog.String("instance_id", instanceID.String()),
		)
	}
}

// RouteEvent routes an event to the appropriate instance buffer
func (r *EventRouter) RouteEvent(ctx context.Context, event *types.InternalEvent) error {
	r.mu.RLock()
	buffer, ok := r.buffers[event.InstanceID]
	r.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no buffer registered for instance %s", event.InstanceID)
	}

	select {
	case buffer.eventCh <- event:
		r.metrics.EventsBuffered.Inc()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Buffer full, event dropped
		buffer.mu.Lock()
		buffer.droppedEvents++
		buffer.mu.Unlock()

		r.log.WarnContext(ctx, "event dropped: buffer full",
			slog.String("instance_id", event.InstanceID.String()),
			slog.String("event_type", event.EventType),
			slog.String("event_id", event.EventID.String()),
		)

		return fmt.Errorf("buffer full for instance %s", event.InstanceID)
	}
}

// GetBuffer returns the buffer for an instance
func (r *EventRouter) GetBuffer(instanceID uuid.UUID) (*EventBuffer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	buffer, ok := r.buffers[instanceID]
	return buffer, ok
}

// GetAllBuffers returns all registered buffers
func (r *EventRouter) GetAllBuffers() map[uuid.UUID]*EventBuffer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to avoid concurrent map access
	buffersCopy := make(map[uuid.UUID]*EventBuffer, len(r.buffers))
	for id, buffer := range r.buffers {
		buffersCopy[id] = buffer
	}

	return buffersCopy
}

// Stop stops all buffers
func (r *EventRouter) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for instanceID, buffer := range r.buffers {
		buffer.Stop()
		r.log.Info("buffer stopped",
			slog.String("instance_id", instanceID.String()),
		)
	}

	r.buffers = make(map[uuid.UUID]*EventBuffer)
}
