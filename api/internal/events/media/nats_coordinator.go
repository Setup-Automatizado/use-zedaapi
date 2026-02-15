package media

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow"

	"go.mau.fi/whatsmeow/api/internal/config"
	natsclient "go.mau.fi/whatsmeow/api/internal/nats"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// NATSMediaCoordinator manages NATS-based media processing workers.
// It replaces the PostgreSQL-polling MediaCoordinator with push-based NATS consumers.
type NATSMediaCoordinator struct {
	natsClient     *natsclient.Client
	clientProvider ClientProvider
	processor      *MediaProcessor
	publisher      *NATSMediaPublisher
	cfg            *config.Config
	mediaCfg       NATSMediaConfig
	metrics        *observability.Metrics
	natsMetrics    *natsclient.NATSMetrics
	log            *slog.Logger

	// Worker management
	mu      sync.RWMutex
	workers map[uuid.UUID]*NATSMediaWorker

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// NATSMediaCoordinatorConfig holds dependencies for creating a NATSMediaCoordinator.
type NATSMediaCoordinatorConfig struct {
	NATSClient     *natsclient.Client
	ClientProvider ClientProvider
	Config         *config.Config
	MediaConfig    NATSMediaConfig
	Metrics        *observability.Metrics
	NATSMetrics    *natsclient.NATSMetrics
	Logger         *slog.Logger
}

// NewNATSMediaCoordinator creates a new NATS-based media coordinator.
func NewNATSMediaCoordinator(ctx context.Context, coordCfg *NATSMediaCoordinatorConfig) (*NATSMediaCoordinator, error) {
	log := coordCfg.Logger
	if log == nil {
		log = slog.Default()
	}

	// Create shared media processor (download + upload pipeline)
	processor, err := NewMediaProcessor(ctx, coordCfg.Config, nil, nil, coordCfg.Metrics)
	if err != nil {
		return nil, fmt.Errorf("create media processor: %w", err)
	}

	publisher := NewNATSMediaPublisher(coordCfg.NATSClient, log, coordCfg.Metrics)

	ctx, cancel := context.WithCancel(ctx)

	c := &NATSMediaCoordinator{
		natsClient:     coordCfg.NATSClient,
		clientProvider: coordCfg.ClientProvider,
		processor:      processor,
		publisher:      publisher,
		cfg:            coordCfg.Config,
		mediaCfg:       coordCfg.MediaConfig,
		metrics:        coordCfg.Metrics,
		natsMetrics:    coordCfg.NATSMetrics,
		log:            log.With(slog.String("component", "nats_media_coordinator")),
		workers:        make(map[uuid.UUID]*NATSMediaWorker),
		ctx:            ctx,
		cancel:         cancel,
	}

	c.log.Info("NATS media coordinator initialized")
	return c, nil
}

// RegisterInstance creates and starts a NATS media worker for an instance.
func (c *NATSMediaCoordinator) RegisterInstance(instanceID uuid.UUID, client *whatsmeow.Client) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.workers[instanceID]; exists {
		return nil
	}

	worker := NewNATSMediaWorker(NATSMediaWorkerConfig{
		InstanceID:     instanceID,
		NATSClient:     c.natsClient,
		ClientProvider: c.clientProvider,
		Processor:      c.processor,
		Publisher:      c.publisher,
		Config:         c.cfg,
		MediaConfig:    c.mediaCfg,
		Metrics:        c.metrics,
		NATSMetrics:    c.natsMetrics,
		Logger:         c.log,
	})

	if err := worker.Start(c.ctx); err != nil {
		return fmt.Errorf("start media worker for %s: %w", instanceID, err)
	}

	c.workers[instanceID] = worker

	c.log.Info("NATS media worker registered",
		slog.String("instance_id", instanceID.String()),
		slog.Int("total_workers", len(c.workers)))

	return nil
}

// UnregisterInstance stops and removes a NATS media worker.
func (c *NATSMediaCoordinator) UnregisterInstance(instanceID uuid.UUID) error {
	c.mu.Lock()
	worker, exists := c.workers[instanceID]
	if !exists {
		c.mu.Unlock()
		return nil
	}
	delete(c.workers, instanceID)
	c.mu.Unlock()

	if err := worker.Stop(c.ctx); err != nil {
		return fmt.Errorf("stop media worker %s: %w", instanceID, err)
	}

	c.log.Info("NATS media worker unregistered",
		slog.String("instance_id", instanceID.String()))

	return nil
}

// UpdateClient updates the WhatsApp client reference for an instance.
// This is called when a client reconnects.
func (c *NATSMediaCoordinator) UpdateClient(instanceID uuid.UUID, client *whatsmeow.Client) {
	// ClientProvider interface handles client lookup dynamically,
	// so no per-worker update is needed.
	c.log.Debug("client update noted for media coordinator",
		slog.String("instance_id", instanceID.String()))
}

// GetPublisher returns the media publisher for external use (e.g., fast path).
func (c *NATSMediaCoordinator) GetPublisher() *NATSMediaPublisher {
	return c.publisher
}

// Stop gracefully stops all media workers.
// Workers are stopped BEFORE canceling the context to allow in-flight operations
// to complete gracefully instead of failing with context.Canceled.
func (c *NATSMediaCoordinator) Stop(ctx context.Context) error {
	c.log.Info("stopping NATS media coordinator")

	c.mu.Lock()
	workers := make([]*NATSMediaWorker, 0, len(c.workers))
	for _, w := range c.workers {
		workers = append(workers, w)
	}
	c.workers = make(map[uuid.UUID]*NATSMediaWorker)
	c.mu.Unlock()

	var wg sync.WaitGroup
	for _, w := range workers {
		wg.Add(1)
		go func(worker *NATSMediaWorker) {
			defer wg.Done()
			if err := worker.Stop(ctx); err != nil {
				c.log.Error("failed to stop media worker",
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
		c.log.Info("all NATS media workers stopped", slog.Int("count", len(workers)))
	case <-time.After(30 * time.Second):
		c.log.Warn("NATS media worker stop timeout")
	}

	// Cancel context AFTER workers have stopped to avoid interrupting in-flight operations
	c.cancel()

	return nil
}
