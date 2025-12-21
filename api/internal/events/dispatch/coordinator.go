package dispatch

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/events/pollstore"
	"go.mau.fi/whatsmeow/api/internal/events/transport"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

type Coordinator struct {
	cfg               *config.Config
	pool              *pgxpool.Pool
	outboxRepo        persistence.OutboxRepository
	dlqRepo           persistence.DLQRepository
	transportRegistry *transport.Registry
	lookup            InstanceLookup
	metrics           *observability.Metrics
	pollStore         pollstore.Store

	mu       sync.RWMutex
	workers  map[uuid.UUID]*InstanceWorker
	stopChan chan struct{}
	wg       sync.WaitGroup
	running  bool
}

func NewCoordinator(
	cfg *config.Config,
	pool *pgxpool.Pool,
	outboxRepo persistence.OutboxRepository,
	dlqRepo persistence.DLQRepository,
	transportRegistry *transport.Registry,
	lookup InstanceLookup,
	pollStore pollstore.Store,
	metrics *observability.Metrics,
) *Coordinator {
	return &Coordinator{
		cfg:               cfg,
		pool:              pool,
		outboxRepo:        outboxRepo,
		dlqRepo:           dlqRepo,
		transportRegistry: transportRegistry,
		lookup:            lookup,
		metrics:           metrics,
		pollStore:         pollStore,
		workers:           make(map[uuid.UUID]*InstanceWorker),
		stopChan:          make(chan struct{}),
		running:           false,
	}
}

func (c *Coordinator) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("coordinator already running")
	}

	c.running = true

	// Start periodic metrics update routine
	c.wg.Add(1)
	go c.metricsUpdateLoop(ctx)

	logger := logging.ContextLogger(ctx, nil)
	logger.Info("dispatch coordinator started",
		slog.Int("workers", len(c.workers)))

	return nil
}

// metricsUpdateLoop periodically updates outbox backlog metrics
func (c *Coordinator) metricsUpdateLoop(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.updateBacklogMetrics(ctx)
		}
	}
}

// updateBacklogMetrics updates the EventOutboxBacklog gauge for all registered instances
func (c *Coordinator) updateBacklogMetrics(ctx context.Context) {
	if c.metrics == nil {
		return
	}

	c.mu.RLock()
	instanceIDs := make([]uuid.UUID, 0, len(c.workers))
	for id := range c.workers {
		instanceIDs = append(instanceIDs, id)
	}
	c.mu.RUnlock()

	for _, instanceID := range instanceIDs {
		count, err := c.outboxRepo.CountPendingByInstance(ctx, instanceID)
		if err != nil {
			continue
		}
		c.metrics.EventOutboxBacklog.WithLabelValues(instanceID.String()).Set(float64(count))
	}
}

func (c *Coordinator) RegisterInstance(ctx context.Context, instanceID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return fmt.Errorf("coordinator not running")
	}

	if _, exists := c.workers[instanceID]; exists {
		return nil
	}

	logger := logging.ContextLogger(ctx, nil)
	logger.Info("registering dispatch worker for instance",
		slog.String("instance_id", instanceID.String()))

	worker := NewInstanceWorker(
		instanceID,
		c.cfg,
		c.outboxRepo,
		c.dlqRepo,
		c.transportRegistry,
		c.lookup,
		c.pollStore,
		c.metrics,
	)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		worker.Run(c.createWorkerContext())
	}()

	c.workers[instanceID] = worker

	logger.Info("dispatch worker registered and started",
		slog.String("instance_id", instanceID.String()),
		slog.Int("total_workers", len(c.workers)))

	return nil
}

func (c *Coordinator) UnregisterInstance(ctx context.Context, instanceID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	worker, exists := c.workers[instanceID]
	if !exists {
		return nil
	}

	logger := logging.ContextLogger(ctx, nil)
	logger.Info("unregistering dispatch worker for instance",
		slog.String("instance_id", instanceID.String()))

	worker.Stop()

	delete(c.workers, instanceID)

	logger.Info("dispatch worker unregistered",
		slog.String("instance_id", instanceID.String()),
		slog.Int("remaining_workers", len(c.workers)))

	return nil
}

func (c *Coordinator) Stop(ctx context.Context) error {
	c.mu.Lock()

	if !c.running {
		c.mu.Unlock()
		return nil
	}

	c.running = false

	logger := logging.ContextLogger(ctx, nil)
	logger.Info("stopping dispatch coordinator",
		slog.Int("active_workers", len(c.workers)))

	for instanceID, worker := range c.workers {
		logger.Info("stopping worker",
			slog.String("instance_id", instanceID.String()))
		worker.Stop()
	}

	c.workers = make(map[uuid.UUID]*InstanceWorker)

	c.mu.Unlock()

	close(c.stopChan)

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("all dispatch workers stopped gracefully")
	case <-time.After(c.cfg.Events.ShutdownGracePeriod):
		logger.Warn("dispatch workers shutdown timeout exceeded, forcing stop")
	}

	return nil
}

func (c *Coordinator) GetWorkerCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.workers)
}

func (c *Coordinator) IsInstanceRegistered(instanceID uuid.UUID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.workers[instanceID]
	return exists
}

func (c *Coordinator) createWorkerContext() context.Context {
	ctx := context.Background()
	return ctx
}
