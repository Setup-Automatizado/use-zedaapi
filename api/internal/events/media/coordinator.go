package media

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// MediaCoordinator manages media workers for multiple WhatsApp instances
type MediaCoordinator struct {
	cfg        *config.Config
	mediaRepo  persistence.MediaRepository
	outboxRepo persistence.OutboxRepository
	metrics    *observability.Metrics
	logger     *slog.Logger

	// Worker management
	workers map[uuid.UUID]*MediaWorker
	clients map[uuid.UUID]*whatsmeow.Client
	mu      sync.RWMutex

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewMediaCoordinator creates a new media coordinator
func NewMediaCoordinator(
	cfg *config.Config,
	mediaRepo persistence.MediaRepository,
	outboxRepo persistence.OutboxRepository,
	metrics *observability.Metrics,
) *MediaCoordinator {
	logger := slog.Default().With(
		slog.String("component", "media_coordinator"),
	)

	ctx, cancel := context.WithCancel(context.Background())

	logger.Info("media coordinator initialized",
		slog.Duration("poll_interval", cfg.Events.MediaPollInterval),
		slog.Int("batch_size", cfg.Events.MediaBatchSize))

	return &MediaCoordinator{
		cfg:        cfg,
		mediaRepo:  mediaRepo,
		outboxRepo: outboxRepo,
		metrics:    metrics,
		logger:     logger,
		workers:    make(map[uuid.UUID]*MediaWorker),
		clients:    make(map[uuid.UUID]*whatsmeow.Client),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// RegisterInstance registers a WhatsApp instance for media processing
func (c *MediaCoordinator) RegisterInstance(instanceID uuid.UUID, client *whatsmeow.Client) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already registered
	if _, exists := c.workers[instanceID]; exists {
		c.logger.Warn("instance already registered",
			slog.String("instance_id", instanceID.String()))
		return nil
	}

	logger := logging.ContextLogger(c.ctx, c.logger).With(
		slog.String("instance_id", instanceID.String()))

	// Create worker for instance
	worker, err := NewMediaWorker(
		c.ctx,
		instanceID,
		client,
		c.cfg,
		c.mediaRepo,
		c.outboxRepo,
		c.metrics,
	)
	if err != nil {
		logger.Error("failed to create media worker",
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to create media worker: %w", err)
	}

	// Store worker and client
	c.workers[instanceID] = worker
	c.clients[instanceID] = client

	// Start worker
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		worker.Start(c.ctx)
	}()

	logger.Info("media worker registered and started",
		slog.String("worker_id", worker.workerID))

	return nil
}

// UnregisterInstance unregisters a WhatsApp instance from media processing
func (c *MediaCoordinator) UnregisterInstance(instanceID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	worker, exists := c.workers[instanceID]
	if !exists {
		c.logger.Warn("instance not registered",
			slog.String("instance_id", instanceID.String()))
		return nil
	}

	logger := logging.ContextLogger(c.ctx, c.logger).With(
		slog.String("instance_id", instanceID.String()))

	// Stop worker gracefully with timeout
	stopCtx, cancel := context.WithTimeout(context.Background(), c.cfg.Events.ShutdownGracePeriod)
	defer cancel()

	if err := worker.Stop(stopCtx); err != nil {
		logger.Error("worker stop failed",
			slog.String("error", err.Error()))
		// Continue with cleanup even if stop failed
	}

	// Remove from maps
	delete(c.workers, instanceID)
	delete(c.clients, instanceID)

	logger.Info("media worker unregistered")

	return nil
}

// GetWorker retrieves the media worker for an instance (if registered)
func (c *MediaCoordinator) GetWorker(instanceID uuid.UUID) (*MediaWorker, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	worker, exists := c.workers[instanceID]
	return worker, exists
}

// GetActiveWorkerCount returns the number of active media workers
func (c *MediaCoordinator) GetActiveWorkerCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.workers)
}

// GetRegisteredInstances returns list of registered instance IDs
func (c *MediaCoordinator) GetRegisteredInstances() []uuid.UUID {
	c.mu.RLock()
	defer c.mu.RUnlock()

	instances := make([]uuid.UUID, 0, len(c.workers))
	for instanceID := range c.workers {
		instances = append(instances, instanceID)
	}

	return instances
}

// Stop gracefully stops all media workers
func (c *MediaCoordinator) Stop(ctx context.Context) error {
	c.logger.Info("stopping media coordinator",
		slog.Int("active_workers", len(c.workers)))

	// Cancel context to signal all workers to stop
	c.cancel()

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.logger.Info("media coordinator stopped gracefully")
		return nil
	case <-ctx.Done():
		c.logger.Warn("media coordinator stop timeout",
			slog.String("error", ctx.Err().Error()))
		return ctx.Err()
	}
}

// GetStats returns coordinator statistics
func (c *MediaCoordinator) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"active_workers":      len(c.workers),
		"registered_instances": c.GetRegisteredInstances(),
		"poll_interval":       c.cfg.Events.MediaPollInterval.String(),
		"batch_size":          c.cfg.Events.MediaBatchSize,
	}
}
