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

// Orchestrator manages the entire event system lifecycle
type Orchestrator struct {
	log     *slog.Logger
	metrics *observability.Metrics
	config  config.Config

	pool       *pgxpool.Pool
	outboxRepo persistence.OutboxRepository
	dlqRepo    persistence.DLQRepository
	mediaRepo  persistence.MediaRepository

	router *capture.EventRouter
	writer *capture.TransactionalWriter

	mu       sync.RWMutex
	handlers map[uuid.UUID]*capture.EventHandler
	buffers  map[uuid.UUID]*capture.EventBuffer
	stopped  bool
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewOrchestrator creates a new event system orchestrator
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

	// Initialize repositories
	outboxRepo := persistence.NewOutboxRepository(pool)
	dlqRepo := persistence.NewDLQRepository(pool)
	mediaRepo := persistence.NewMediaRepository(pool)

	// Initialize router
	router := capture.NewEventRouter(ctx, metrics)

	// Initialize transactional writer
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

// RegisterInstance registers an instance with the event system
func (o *Orchestrator) RegisterInstance(ctx context.Context, instanceID uuid.UUID) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.stopped {
		return fmt.Errorf("orchestrator stopped")
	}

	// Check if already registered
	if _, exists := o.handlers[instanceID]; exists {
		return fmt.Errorf("instance %s already registered", instanceID)
	}

	// Create buffer for instance
	bufferConfig := capture.BufferConfig{
		InstanceID:    instanceID,
		BufferSize:    o.config.Events.BufferSize,
		BatchSize:     o.config.Events.BatchSize,
		FlushInterval: o.config.Events.PollInterval,
	}

	buffer := capture.NewEventBuffer(ctx, bufferConfig, o.writer, o.metrics)
	o.buffers[instanceID] = buffer

	// Register buffer with router
	o.router.RegisterBuffer(instanceID, buffer)

	// Create event handler
	handler := capture.NewEventHandler(
		ctx,
		instanceID,
		o.router,
		o.metrics,
	)
	o.handlers[instanceID] = handler

	o.log.InfoContext(ctx, "instance registered",
		slog.String("instance_id", instanceID.String()),
	)

	return nil
}

// UnregisterInstance unregisters an instance from the event system
func (o *Orchestrator) UnregisterInstance(ctx context.Context, instanceID uuid.UUID) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	handler, handlerExists := o.handlers[instanceID]
	buffer, bufferExists := o.buffers[instanceID]

	if !handlerExists && !bufferExists {
		return fmt.Errorf("instance %s not registered", instanceID)
	}

	// Stop handler
	if handlerExists {
		handler.Stop()
		delete(o.handlers, instanceID)
	}

	// Stop and flush buffer
	if bufferExists {
		buffer.Flush()
		// Wait for buffer to flush with timeout
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

	// Unregister from router
	o.router.UnregisterBuffer(instanceID)

	o.log.InfoContext(ctx, "instance unregistered",
		slog.String("instance_id", instanceID.String()),
	)

	return nil
}

// GetHandler returns the event handler for an instance
func (o *Orchestrator) GetHandler(instanceID uuid.UUID) (*capture.EventHandler, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	handler, ok := o.handlers[instanceID]
	return handler, ok
}

// GetBuffer returns the event buffer for an instance
func (o *Orchestrator) GetBuffer(instanceID uuid.UUID) (*capture.EventBuffer, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	buffer, ok := o.buffers[instanceID]
	return buffer, ok
}

// GetBufferStats returns statistics for an instance's buffer
func (o *Orchestrator) GetBufferStats(instanceID uuid.UUID) (types.BufferStats, error) {
	buffer, ok := o.GetBuffer(instanceID)
	if !ok {
		return types.BufferStats{}, fmt.Errorf("no buffer for instance %s", instanceID)
	}

	return buffer.Stats(), nil
}

// FlushInstance manually flushes an instance's buffer
func (o *Orchestrator) FlushInstance(instanceID uuid.UUID) error {
	buffer, ok := o.GetBuffer(instanceID)
	if !ok {
		return fmt.Errorf("no buffer for instance %s", instanceID)
	}

	buffer.Flush()
	return nil
}

// FlushAll manually flushes all buffers
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

// Stop gracefully stops the orchestrator
func (o *Orchestrator) Stop(ctx context.Context) error {
	o.mu.Lock()
	if o.stopped {
		o.mu.Unlock()
		return nil
	}
	o.stopped = true
	o.mu.Unlock()

	o.log.Info("stopping event orchestrator")

	// Get all instance IDs
	o.mu.RLock()
	instanceIDs := make([]uuid.UUID, 0, len(o.handlers))
	for id := range o.handlers {
		instanceIDs = append(instanceIDs, id)
	}
	o.mu.RUnlock()

	// Unregister all instances
	for _, instanceID := range instanceIDs {
		if err := o.UnregisterInstance(ctx, instanceID); err != nil {
			o.log.ErrorContext(ctx, "failed to unregister instance",
				slog.String("instance_id", instanceID.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	// Stop router
	o.router.Stop()

	close(o.stopCh)
	o.wg.Wait()

	o.log.Info("event orchestrator stopped")

	return nil
}

// IsInstanceRegistered checks if an instance is registered
func (o *Orchestrator) IsInstanceRegistered(instanceID uuid.UUID) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()

	_, exists := o.handlers[instanceID]
	return exists
}

// GetRegisteredInstances returns all registered instance IDs
func (o *Orchestrator) GetRegisteredInstances() []uuid.UUID {
	o.mu.RLock()
	defer o.mu.RUnlock()

	instanceIDs := make([]uuid.UUID, 0, len(o.handlers))
	for id := range o.handlers {
		instanceIDs = append(instanceIDs, id)
	}

	return instanceIDs
}

// Stats returns overall orchestrator statistics
type OrchestratorStats struct {
	RegisteredInstances int
	TotalBuffered       int64
	TotalDropped        int64
	TotalProcessed      int64
}

// GetStats returns overall orchestrator statistics
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

// HandleEvent is a convenience method to handle an event for an instance
// This is the method that should be called by ClientRegistry.wrapEventHandler
func (o *Orchestrator) HandleEvent(ctx context.Context, instanceID uuid.UUID, rawEvent interface{}) error {
	handler, ok := o.GetHandler(instanceID)
	if !ok {
		return fmt.Errorf("no handler registered for instance %s", instanceID)
	}

	return handler.HandleEvent(ctx, rawEvent)
}
