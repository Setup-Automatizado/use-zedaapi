package dispatch

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/config"
	eventsnats "go.mau.fi/whatsmeow/api/internal/events/nats"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/events/pollstore"
	"go.mau.fi/whatsmeow/api/internal/events/transport"
	natsclient "go.mau.fi/whatsmeow/api/internal/nats"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// NATSDispatchCoordinator manages NATS-based event dispatch workers.
// It replaces the PostgreSQL-polling Coordinator with push-based NATS consumers.
type NATSDispatchCoordinator struct {
	natsClient        *natsclient.Client
	cfg               *config.Config
	eventCfg          eventsnats.NATSEventConfig
	transportRegistry *transport.Registry
	webhookResolver   WebhookResolver
	pollStore         pollstore.Store
	mediaResults      MediaResultLookup
	metrics           *observability.Metrics
	natsMetrics       *natsclient.NATSMetrics
	dlqHandler        *NATSEventDLQHandler
	statusInterceptor StatusInterceptor
	log               *slog.Logger

	// Worker management
	mu      sync.RWMutex
	workers map[uuid.UUID]*NATSDispatchWorker
	running bool

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// NATSDispatchCoordinatorConfig holds dependencies for creating a NATSDispatchCoordinator.
type NATSDispatchCoordinatorConfig struct {
	NATSClient        *natsclient.Client
	Config            *config.Config
	EventConfig       eventsnats.NATSEventConfig
	TransportRegistry *transport.Registry
	WebhookResolver   WebhookResolver
	PollStore         pollstore.Store
	MediaResults      MediaResultLookup
	Metrics           *observability.Metrics
	NATSMetrics       *natsclient.NATSMetrics
	Logger            *slog.Logger
}

// NewNATSDispatchCoordinator creates a new NATS-based dispatch coordinator.
func NewNATSDispatchCoordinator(cfg *NATSDispatchCoordinatorConfig) *NATSDispatchCoordinator {
	log := cfg.Logger
	if log == nil {
		log = slog.Default()
	}

	dlqHandler := NewNATSEventDLQHandler(cfg.NATSClient, log)

	return &NATSDispatchCoordinator{
		natsClient:        cfg.NATSClient,
		cfg:               cfg.Config,
		eventCfg:          cfg.EventConfig,
		transportRegistry: cfg.TransportRegistry,
		webhookResolver:   cfg.WebhookResolver,
		pollStore:         cfg.PollStore,
		mediaResults:      cfg.MediaResults,
		metrics:           cfg.Metrics,
		natsMetrics:       cfg.NATSMetrics,
		dlqHandler:        dlqHandler,
		log:               log.With(slog.String("component", "nats_dispatch_coordinator")),
		workers:           make(map[uuid.UUID]*NATSDispatchWorker),
	}
}

// Start initializes the coordinator.
func (c *NATSDispatchCoordinator) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("NATS dispatch coordinator already running")
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.running = true

	c.log.Info("NATS dispatch coordinator started")
	return nil
}

// RegisterInstance creates and starts a NATS dispatch worker for an instance.
func (c *NATSDispatchCoordinator) RegisterInstance(ctx context.Context, instanceID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return fmt.Errorf("NATS dispatch coordinator not running")
	}

	if _, exists := c.workers[instanceID]; exists {
		return nil
	}

	worker := NewNATSDispatchWorker(NATSDispatchWorkerConfig{
		InstanceID:        instanceID,
		Client:            c.natsClient,
		Config:            c.cfg,
		EventConfig:       c.eventCfg,
		TransportRegistry: c.transportRegistry,
		WebhookResolver:   c.webhookResolver,
		PollStore:         c.pollStore,
		MediaResults:      c.mediaResults,
		Metrics:           c.metrics,
		NATSMetrics:       c.natsMetrics,
		DLQHandler:        c.dlqHandler,
		Logger:            c.log,
	})

	if c.statusInterceptor != nil {
		worker.SetStatusInterceptor(c.statusInterceptor)
	}

	if err := worker.Start(c.ctx); err != nil {
		return fmt.Errorf("start dispatch worker for %s: %w", instanceID, err)
	}

	c.workers[instanceID] = worker

	c.log.Info("NATS dispatch worker registered",
		slog.String("instance_id", instanceID.String()),
		slog.Int("total_workers", len(c.workers)))

	return nil
}

// UnregisterInstance stops and removes a NATS dispatch worker.
func (c *NATSDispatchCoordinator) UnregisterInstance(ctx context.Context, instanceID uuid.UUID) error {
	c.mu.Lock()
	worker, exists := c.workers[instanceID]
	if !exists {
		c.mu.Unlock()
		return nil
	}
	delete(c.workers, instanceID)
	c.mu.Unlock()

	if err := worker.Stop(ctx); err != nil {
		return fmt.Errorf("stop dispatch worker %s: %w", instanceID, err)
	}

	c.log.Info("NATS dispatch worker unregistered",
		slog.String("instance_id", instanceID.String()))

	return nil
}

// SetDLQRepository sets the PostgreSQL DLQ repository on the DLQ handler for dual-write.
func (c *NATSDispatchCoordinator) SetDLQRepository(repo persistence.DLQRepository) {
	if c.dlqHandler != nil {
		c.dlqHandler.SetDLQRepository(repo)
	}
}

// SetStatusInterceptor sets the status cache interceptor for all workers.
func (c *NATSDispatchCoordinator) SetStatusInterceptor(interceptor StatusInterceptor) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statusInterceptor = interceptor

	for _, worker := range c.workers {
		worker.SetStatusInterceptor(interceptor)
	}
}

// GetWorkerCount returns the number of active dispatch workers.
func (c *NATSDispatchCoordinator) GetWorkerCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.workers)
}

// IsInstanceRegistered returns whether an instance has a dispatch worker.
func (c *NATSDispatchCoordinator) IsInstanceRegistered(instanceID uuid.UUID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.workers[instanceID]
	return exists
}

// Stop gracefully stops all dispatch workers.
func (c *NATSDispatchCoordinator) Stop(ctx context.Context) error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = false

	c.log.Info("stopping NATS dispatch coordinator",
		slog.Int("active_workers", len(c.workers)))

	workers := make([]*NATSDispatchWorker, 0, len(c.workers))
	for _, w := range c.workers {
		workers = append(workers, w)
	}
	c.workers = make(map[uuid.UUID]*NATSDispatchWorker)
	c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	var wg sync.WaitGroup
	for _, w := range workers {
		wg.Add(1)
		go func(worker *NATSDispatchWorker) {
			defer wg.Done()
			if err := worker.Stop(ctx); err != nil {
				c.log.Error("failed to stop dispatch worker",
					slog.String("instance_id", worker.instanceID.String()),
					slog.String("error", err.Error()))
			}
		}(w)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.log.Info("all NATS dispatch workers stopped", slog.Int("count", len(workers)))
	case <-time.After(30 * time.Second):
		c.log.Warn("NATS dispatch worker stop timeout")
	}

	return nil
}

// Verify NATSDispatchCoordinator implements DispatchCoordinator at compile time.
var _ DispatchCoordinator = (*NATSDispatchCoordinator)(nil)
