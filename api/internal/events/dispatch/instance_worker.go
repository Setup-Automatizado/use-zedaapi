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
	"go.mau.fi/whatsmeow/api/internal/events/pollstore"
	"go.mau.fi/whatsmeow/api/internal/events/transport"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

type InstanceWorker struct {
	instanceID        uuid.UUID
	cfg               *config.Config
	outboxRepo        persistence.OutboxRepository
	dlqRepo           persistence.DLQRepository
	transportRegistry *transport.Registry
	metrics           *observability.Metrics
	pollStore         pollstore.Store

	processor *EventProcessor

	mu       sync.RWMutex
	stopChan chan struct{}
	running  bool
}

func NewInstanceWorker(
	instanceID uuid.UUID,
	cfg *config.Config,
	outboxRepo persistence.OutboxRepository,
	dlqRepo persistence.DLQRepository,
	transportRegistry *transport.Registry,
	lookup InstanceLookup,
	pollStore pollstore.Store,
	metrics *observability.Metrics,
) *InstanceWorker {
	processor := NewEventProcessor(
		instanceID,
		cfg,
		outboxRepo,
		dlqRepo,
		transportRegistry,
		lookup,
		pollStore,
		metrics,
	)

	return &InstanceWorker{
		instanceID:        instanceID,
		cfg:               cfg,
		outboxRepo:        outboxRepo,
		dlqRepo:           dlqRepo,
		transportRegistry: transportRegistry,
		metrics:           metrics,
		processor:         processor,
		pollStore:         pollStore,
		stopChan:          make(chan struct{}),
		running:           false,
	}
}

func (w *InstanceWorker) Run(ctx context.Context) {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.mu.Unlock()

	ctx = logging.WithAttrs(ctx, slog.String("instance_id", w.instanceID.String()))
	logger := logging.ContextLogger(ctx, nil)

	logger.Info("instance worker started",
		slog.Duration("poll_interval", w.cfg.Events.PollInterval))

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

			if err := w.pollAndProcess(ctx); err != nil {
				logger.Error("poll and process failed",
					slog.String("error", err.Error()))
				w.metrics.WorkerErrors.WithLabelValues(w.instanceID.String(), "poll_error").Inc()
			}
		}
	}
}

func (w *InstanceWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	w.running = false
	close(w.stopChan)
}

func (w *InstanceWorker) isRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

func (w *InstanceWorker) pollAndProcess(ctx context.Context) error {
	start := time.Now()
	logger := logging.ContextLogger(ctx, nil)

	events, err := w.pollEvents(ctx)
	if err != nil {
		return fmt.Errorf("poll events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	logger.Debug("polled events from outbox",
		slog.Int("count", len(events)),
		slog.Duration("duration", time.Since(start)))

	successCount := 0
	failureCount := 0

	for _, event := range events {
		if !w.isRunning() {
			logger.Info("worker stopped during processing, breaking loop")
			break
		}

		eventCtx := logging.WithAttrs(ctx,
			slog.String("event_id", event.EventID.String()),
			slog.String("event_type", event.EventType))

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
