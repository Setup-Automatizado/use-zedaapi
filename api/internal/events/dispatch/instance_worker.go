package dispatch

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/events/transport"
	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// InstanceWorker processes events for a single WhatsApp instance
type InstanceWorker struct {
	instanceID        uuid.UUID
	cfg               *config.Config
	outboxRepo        persistence.OutboxRepository
	dlqRepo           persistence.DLQRepository
	transportRegistry *transport.Registry
	instanceRepo      *instances.Repository
	metrics           *observability.Metrics

	processor *EventProcessor

	mu       sync.RWMutex
	stopChan chan struct{}
	running  bool
}

// NewInstanceWorker creates a new instance worker
func NewInstanceWorker(
	instanceID uuid.UUID,
	cfg *config.Config,
	outboxRepo persistence.OutboxRepository,
	dlqRepo persistence.DLQRepository,
	transportRegistry *transport.Registry,
	instanceRepo *instances.Repository,
	metrics *observability.Metrics,
) *InstanceWorker {
	// Create event processor
	processor := NewEventProcessor(
		instanceID,
		cfg,
		outboxRepo,
		dlqRepo,
		transportRegistry,
		instanceRepo,
		metrics,
	)

	return &InstanceWorker{
		instanceID:        instanceID,
		cfg:               cfg,
		outboxRepo:        outboxRepo,
		dlqRepo:           dlqRepo,
		transportRegistry: transportRegistry,
		instanceRepo:      instanceRepo,
		metrics:           metrics,
		processor:         processor,
		stopChan:          make(chan struct{}),
		running:           false,
	}
}

// Run starts the worker's main loop
func (w *InstanceWorker) Run(ctx context.Context) {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.mu.Unlock()

	// Add instance_id to context
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", w.instanceID.String()))
	logger := logging.ContextLogger(ctx, nil)

	logger.Info("instance worker started",
		slog.Duration("poll_interval", w.cfg.Events.PollInterval))

	// Update metrics
	w.metrics.WorkersActive.WithLabelValues(w.instanceID.String()).Inc()
	defer w.metrics.WorkersActive.WithLabelValues(w.instanceID.String()).Dec()

	ticker := time.NewTicker(w.cfg.Events.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker stopped due to context cancellation")
			return

		case <-w.stopChan:
			logger.Info("worker stopped by stop signal")
			return

		case <-ticker.C:
			if !w.isRunning() {
				return
			}

			// Poll and process events
			if err := w.pollAndProcess(ctx); err != nil {
				logger.Error("poll and process failed",
					slog.String("error", err.Error()))
				w.metrics.WorkerErrors.WithLabelValues(w.instanceID.String(), "poll_error").Inc()
			}
		}
	}
}

// Stop gracefully stops the worker
func (w *InstanceWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	w.running = false
	close(w.stopChan)
}

// isRunning checks if worker is still running (thread-safe)
func (w *InstanceWorker) isRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// pollAndProcess polls the outbox and processes pending events
func (w *InstanceWorker) pollAndProcess(ctx context.Context) error {
	start := time.Now()
	logger := logging.ContextLogger(ctx, nil)

	// Get pending events for this instance
	events, err := w.pollEvents(ctx)
	if err != nil {
		return fmt.Errorf("poll events: %w", err)
	}

	if len(events) == 0 {
		return nil // No events to process
	}

	logger.Debug("polled events from outbox",
		slog.Int("count", len(events)),
		slog.Duration("duration", time.Since(start)))

	// Process each event
	successCount := 0
	failureCount := 0

	for _, event := range events {
		if !w.isRunning() {
			logger.Info("worker stopped during processing, breaking loop")
			break
		}

		// Add event_id to context
		eventCtx := logging.WithAttrs(ctx,
			slog.String("event_id", event.EventID.String()),
			slog.String("event_type", event.EventType))

		// Process event with timeout
		processCtx, cancel := context.WithTimeout(eventCtx, w.cfg.Events.ProcessingTimeout)
		err := w.processor.Process(processCtx, event)
		cancel()

		if err != nil {
			logger.Error("event processing failed",
				slog.String("event_id", event.EventID.String()),
				slog.String("error", err.Error()))
			failureCount++
		} else {
			successCount++
		}
	}

	logger.Info("batch processing completed",
		slog.Int("total", len(events)),
		slog.Int("success", successCount),
		slog.Int("failed", failureCount),
		slog.Duration("duration", time.Since(start)))

	// Update batch metrics
	w.metrics.WorkerTaskDuration.WithLabelValues(w.instanceID.String(), "poll_and_process").Observe(time.Since(start).Seconds())

	return nil
}

// pollEvents retrieves pending events from the outbox for this instance
func (w *InstanceWorker) pollEvents(ctx context.Context) ([]*persistence.OutboxEvent, error) {
	// Query pending events that are ready for processing
	// - status = 'pending' OR (status = 'retrying' AND next_attempt_at <= NOW())
	// - instance_id = w.instanceID
	// - ORDER BY sequence_number ASC (maintain order)
	// - LIMIT cfg.Events.BatchSize

	events, err := w.outboxRepo.PollPendingEvents(ctx, w.instanceID, w.cfg.Events.BatchSize)
	if err != nil {
		return nil, fmt.Errorf("poll pending events: %w", err)
	}

	return events, nil
}
