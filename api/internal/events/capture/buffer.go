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

type EventWriter interface {
	WriteEvents(ctx context.Context, events []*types.InternalEvent) error
}

type BufferConfig struct {
	InstanceID    uuid.UUID
	BufferSize    int
	BatchSize     int
	FlushInterval time.Duration
}

func NewEventBuffer(
	ctx context.Context,
	config BufferConfig,
	writer EventWriter,
	metrics *observability.Metrics,
) *EventBuffer {
	ctx = context.WithoutCancel(ctx)
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
		err := b.persistWithRetry(ctx, batch)
		if err != nil {
			b.log.ErrorContext(ctx, "failed to persist events after retries",
				slog.String("error", err.Error()),
				slog.Int("batch_size", len(batch)))
			for _, evt := range batch {
				select {
				case b.eventCh <- evt:
					// TODO: this is a hack to ensure that the event is not lost if the persistence fails
					// requeued for future attempt
				default:
					b.droppedEvents++
					b.log.ErrorContext(ctx, "event dropped after retry",
						slog.String("event_id", evt.EventID.String()))
				}
			}
		} else {
			b.log.DebugContext(ctx, "events flushed",
				slog.Int("batch_size", len(batch)),
				slog.Duration("duration", time.Since(start)),
			)

			b.metrics.EventsBuffered.Sub(float64(len(batch)))
		}

		batch = batch[:0]
	}

	for {
		select {
		case event := <-b.eventCh:
			batch = append(batch, event)
			b.mu.Lock()
			b.totalEvents++
			b.mu.Unlock()

			if len(batch) >= b.batchSize {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-b.flushCh:
			flush()

		case <-b.stopCh:
			flush()
			b.log.Info("buffer stopped",
				slog.Int64("total_events", b.totalEvents),
				slog.Int64("dropped_events", b.droppedEvents),
			)
			return

		case <-ctx.Done():
			flush()
			return
		}
	}
}

func (b *EventBuffer) Flush() {
	select {
	case b.flushCh <- struct{}{}:
	default:
	}
}

func (b *EventBuffer) persistWithRetry(ctx context.Context, events []*types.InternalEvent) error {
	if len(events) == 0 {
		return nil
	}

	var err error
	const maxAttempts = 3

	for attempt := 0; attempt < maxAttempts; attempt++ {
		err = b.writer.WriteEvents(ctx, events)
		if err == nil {
			return nil
		}

		backoff := time.Duration(attempt+1) * 200 * time.Millisecond
		b.log.WarnContext(ctx, "batch persistence failed",
			slog.Int("attempt", attempt+1),
			slog.Duration("backoff", backoff),
			slog.String("error", err.Error()))

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}

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

func (b *EventBuffer) IsStopped() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.stopped
}

func (b *EventBuffer) Size() int {
	return len(b.eventCh)
}

func (b *EventBuffer) DroppedCount() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.droppedEvents
}

func (b *EventBuffer) TotalCount() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.totalEvents
}

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
