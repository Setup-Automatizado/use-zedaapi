package capture

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// EventBuffer buffers events for an instance before persistence
type EventBuffer struct {
	log        *slog.Logger
	metrics    *observability.Metrics
	writer     EventWriter
	instanceID uuid.UUID

	eventCh chan *types.InternalEvent
	flushCh chan struct{}

	mu            sync.RWMutex
	batchSize     int
	flushInterval time.Duration
	droppedEvents int64
	totalEvents   int64
	stopped       bool
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

// EventWriter defines the interface for writing events to persistence
type EventWriter interface {
	WriteEvents(ctx context.Context, events []*types.InternalEvent) error
}

// BufferConfig contains buffer configuration
type BufferConfig struct {
	InstanceID    uuid.UUID
	BufferSize    int
	BatchSize     int
	FlushInterval time.Duration
}

// NewEventBuffer creates a new EventBuffer for an instance
func NewEventBuffer(
	ctx context.Context,
	config BufferConfig,
	writer EventWriter,
	metrics *observability.Metrics,
) *EventBuffer {
	log := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "event_buffer"),
		slog.String("instance_id", config.InstanceID.String()),
	)

	buffer := &EventBuffer{
		log:           log,
		metrics:       metrics,
		writer:        writer,
		instanceID:    config.InstanceID,
		eventCh:       make(chan *types.InternalEvent, config.BufferSize),
		flushCh:       make(chan struct{}, 1),
		batchSize:     config.BatchSize,
		flushInterval: config.FlushInterval,
		stopCh:        make(chan struct{}),
	}

	buffer.wg.Add(1)
	go buffer.run(ctx)

	return buffer
}

// run is the main buffer loop that batches and flushes events
func (b *EventBuffer) run(ctx context.Context) {
	defer b.wg.Done()

	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	batch := make([]*types.InternalEvent, 0, b.batchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		start := time.Now()
		if err := b.writer.WriteEvents(ctx, batch); err != nil {
			b.log.ErrorContext(ctx, "failed to write events",
				slog.String("error", err.Error()),
				slog.Int("batch_size", len(batch)),
			)
			// TODO: Implement retry logic or move to DLQ
		} else {
			b.log.DebugContext(ctx, "events flushed",
				slog.Int("batch_size", len(batch)),
				slog.Duration("duration", time.Since(start)),
			)

			// Update metrics
			b.metrics.EventsBuffered.Sub(float64(len(batch)))
		}

		// Reset batch
		batch = batch[:0]
	}

	for {
		select {
		case event := <-b.eventCh:
			batch = append(batch, event)
			b.mu.Lock()
			b.totalEvents++
			b.mu.Unlock()

			// Flush if batch is full
			if len(batch) >= b.batchSize {
				flush()
			}

		case <-ticker.C:
			// Periodic flush
			flush()

		case <-b.flushCh:
			// Manual flush requested
			flush()

		case <-b.stopCh:
			// Flush remaining events before stopping
			flush()
			b.log.Info("buffer stopped",
				slog.Int64("total_events", b.totalEvents),
				slog.Int64("dropped_events", b.droppedEvents),
			)
			return

		case <-ctx.Done():
			// Context cancelled, flush and stop
			flush()
			return
		}
	}
}

// Flush manually triggers a flush
func (b *EventBuffer) Flush() {
	select {
	case b.flushCh <- struct{}{}:
	default:
		// Flush already pending
	}
}

// Stop gracefully stops the buffer
func (b *EventBuffer) Stop() {
	b.mu.Lock()
	if b.stopped {
		b.mu.Unlock()
		return
	}
	b.stopped = true
	b.mu.Unlock()

	close(b.stopCh)
	b.wg.Wait()
}

// Stats returns buffer statistics
func (b *EventBuffer) Stats() types.BufferStats {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return types.BufferStats{
		Capacity:      cap(b.eventCh),
		Size:          len(b.eventCh),
		DroppedEvents: b.droppedEvents,
		TotalEvents:   b.totalEvents,
	}
}

// IsStopped returns whether the buffer is stopped
func (b *EventBuffer) IsStopped() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.stopped
}

// Size returns the current buffer size
func (b *EventBuffer) Size() int {
	return len(b.eventCh)
}

// DroppedCount returns the number of dropped events
func (b *EventBuffer) DroppedCount() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.droppedEvents
}

// TotalCount returns the total number of events processed
func (b *EventBuffer) TotalCount() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.totalEvents
}

// WaitForFlush waits for the buffer to be empty (useful for testing)
func (b *EventBuffer) WaitForFlush(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for buffer flush")
		case <-ticker.C:
			if len(b.eventCh) == 0 {
				return nil
			}
		}
	}
}
