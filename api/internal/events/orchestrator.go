package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/capture"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

type Orchestrator struct {
	log        *slog.Logger
	metrics    *observability.Metrics
	config     config.Config
	pool       *pgxpool.Pool
	outboxRepo persistence.OutboxRepository
	dlqRepo    persistence.DLQRepository
	mediaRepo  persistence.MediaRepository
	router     *capture.EventRouter
	writer     *capture.TransactionalWriter
	mu         sync.RWMutex
	handlers   map[uuid.UUID]*capture.EventHandler
	buffers    map[uuid.UUID]*capture.EventBuffer
	stopped    bool
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

func NewOrchestrator(
	ctx context.Context,
	cfg config.Config,
	pool *pgxpool.Pool,
	resolver capture.WebhookResolver,
	metadataEnricher capture.MetadataEnricher,
	metrics *observability.Metrics,
) (*Orchestrator, error) {
	log := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "event_orchestrator"),
	)

	outboxRepo := persistence.NewOutboxRepository(pool)
	dlqRepo := persistence.NewDLQRepository(pool)
	mediaRepo := persistence.NewMediaRepository(pool)

	router := capture.NewEventRouter(ctx, metrics)

	writer := capture.NewTransactionalWriter(
		ctx,
		pool,
		outboxRepo,
		mediaRepo,
		resolver,
		metadataEnricher,
		&cfg,
		metrics,
	)

	orchestrator := &Orchestrator{
		log:        log,
		metrics:    metrics,
		config:     cfg,
		pool:       pool,
		outboxRepo: outboxRepo,
		dlqRepo:    dlqRepo,
		mediaRepo:  mediaRepo,
		router:     router,
		writer:     writer,
		handlers:   make(map[uuid.UUID]*capture.EventHandler),
		buffers:    make(map[uuid.UUID]*capture.EventBuffer),
		stopCh:     make(chan struct{}),
	}

	log.Info("event orchestrator initialized",
		slog.Int("buffer_size", cfg.Events.BufferSize),
		slog.Int("batch_size", cfg.Events.BatchSize),
		slog.Duration("poll_interval", cfg.Events.PollInterval),
	)

	return orchestrator, nil
}

func (o *Orchestrator) RegisterInstance(ctx context.Context, instanceID uuid.UUID) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.stopped {
		return fmt.Errorf("orchestrator stopped")
	}

	if _, exists := o.handlers[instanceID]; exists {
		return fmt.Errorf("instance %s already registered", instanceID)
	}

	bufferConfig := capture.BufferConfig{
		InstanceID:    instanceID,
		BufferSize:    o.config.Events.BufferSize,
		BatchSize:     o.config.Events.BatchSize,
		FlushInterval: o.config.Events.PollInterval,
	}

	buffer := capture.NewEventBuffer(ctx, bufferConfig, o.writer, o.metrics)
	o.buffers[instanceID] = buffer

	o.router.RegisterBuffer(instanceID, buffer)

	handler := capture.NewEventHandler(
		ctx,
		instanceID,
		o.router,
		o.metrics,
		o.config.Events.DebugRawPayload,
		o.config.Events.DebugDumpDir,
	)
	o.handlers[instanceID] = handler

	o.log.InfoContext(ctx, "instance registered",
		slog.String("instance_id", instanceID.String()),
	)

	return nil
}

func (o *Orchestrator) UnregisterInstance(ctx context.Context, instanceID uuid.UUID) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	handler, handlerExists := o.handlers[instanceID]
	buffer, bufferExists := o.buffers[instanceID]

	if !handlerExists && !bufferExists {
		return fmt.Errorf("instance %s not registered", instanceID)
	}

	if handlerExists {
		handler.Stop()
		delete(o.handlers, instanceID)
	}

	if bufferExists {
		buffer.Flush()
		flushCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := buffer.WaitForFlush(flushCtx, 5*time.Second); err != nil {
			o.log.WarnContext(ctx, "buffer flush timeout",
				slog.String("instance_id", instanceID.String()),
				slog.String("error", err.Error()),
			)
		}
		buffer.Stop()
		delete(o.buffers, instanceID)
	}

	o.router.UnregisterBuffer(instanceID)

	o.log.InfoContext(ctx, "instance unregistered",
		slog.String("instance_id", instanceID.String()),
	)

	return nil
}

func (o *Orchestrator) GetHandler(instanceID uuid.UUID) (*capture.EventHandler, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	handler, ok := o.handlers[instanceID]
	return handler, ok
}

func (o *Orchestrator) GetBuffer(instanceID uuid.UUID) (*capture.EventBuffer, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	buffer, ok := o.buffers[instanceID]
	return buffer, ok
}

func (o *Orchestrator) GetBufferStats(instanceID uuid.UUID) (types.BufferStats, error) {
	buffer, ok := o.GetBuffer(instanceID)
	if !ok {
		return types.BufferStats{}, fmt.Errorf("no buffer for instance %s", instanceID)
	}

	return buffer.Stats(), nil
}

func (o *Orchestrator) FlushInstance(instanceID uuid.UUID) error {
	buffer, ok := o.GetBuffer(instanceID)
	if !ok {
		return fmt.Errorf("no buffer for instance %s", instanceID)
	}

	buffer.Flush()
	return nil
}

func (o *Orchestrator) FlushAll() {
	o.mu.RLock()
	buffers := make([]*capture.EventBuffer, 0, len(o.buffers))
	for _, buffer := range o.buffers {
		buffers = append(buffers, buffer)
	}
	o.mu.RUnlock()

	for _, buffer := range buffers {
		buffer.Flush()
	}
}

func (o *Orchestrator) Stop(ctx context.Context) error {
	o.mu.Lock()
	if o.stopped {
		o.mu.Unlock()
		return nil
	}
	o.stopped = true
	o.mu.Unlock()

	o.log.Info("stopping event orchestrator")

	o.mu.RLock()
	instanceIDs := make([]uuid.UUID, 0, len(o.handlers))
	for id := range o.handlers {
		instanceIDs = append(instanceIDs, id)
	}
	o.mu.RUnlock()

	for _, instanceID := range instanceIDs {
		if err := o.UnregisterInstance(ctx, instanceID); err != nil {
			o.log.ErrorContext(ctx, "failed to unregister instance",
				slog.String("instance_id", instanceID.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	o.router.Stop()

	close(o.stopCh)
	o.wg.Wait()

	o.log.Info("event orchestrator stopped")

	return nil
}

func (o *Orchestrator) IsInstanceRegistered(instanceID uuid.UUID) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()

	_, exists := o.handlers[instanceID]
	return exists
}

func (o *Orchestrator) GetRegisteredInstances() []uuid.UUID {
	o.mu.RLock()
	defer o.mu.RUnlock()

	instanceIDs := make([]uuid.UUID, 0, len(o.handlers))
	for id := range o.handlers {
		instanceIDs = append(instanceIDs, id)
	}

	return instanceIDs
}

type OrchestratorStats struct {
	RegisteredInstances int
	TotalBuffered       int64
	TotalDropped        int64
	TotalProcessed      int64
}

func (o *Orchestrator) GetStats() OrchestratorStats {
	o.mu.RLock()
	defer o.mu.RUnlock()

	stats := OrchestratorStats{
		RegisteredInstances: len(o.handlers),
	}

	for _, buffer := range o.buffers {
		bufferStats := buffer.Stats()
		stats.TotalBuffered += int64(bufferStats.Size)
		stats.TotalDropped += bufferStats.DroppedEvents
		stats.TotalProcessed += bufferStats.TotalEvents
	}

	return stats
}

func (o *Orchestrator) HandleEvent(ctx context.Context, instanceID uuid.UUID, rawEvent interface{}) error {
	handler, ok := o.GetHandler(instanceID)
	if !ok {
		return fmt.Errorf("no handler registered for instance %s", instanceID)
	}

	return handler.HandleEvent(ctx, rawEvent)
}
