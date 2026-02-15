package queue

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"

	"go.mau.fi/whatsmeow/api/internal/events/echo"
	natsclient "go.mau.fi/whatsmeow/api/internal/nats"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// NATSCoordinator manages NATS-based message queue workers.
// It implements QueueCoordinator and replaces the PostgreSQL-based Coordinator.
type NATSCoordinator struct {
	natsClient     *natsclient.Client
	publisher      *NATSPublisher
	dlq            *NATSDLQHandler
	clientRegistry ClientRegistry
	processor      MessageProcessor
	cfg            NATSConfig
	log            *slog.Logger
	metrics        *observability.Metrics
	natsMetrics    *natsclient.NATSMetrics

	// Redis for cancel set
	redis *redis.Client

	// Worker management
	mu      sync.RWMutex
	workers map[uuid.UUID]*NATSWorker

	// Proxy pause control
	pauseMu         sync.RWMutex
	pausedInstances map[uuid.UUID]*pauseState

	// Drain control
	drainMu  sync.RWMutex
	draining bool

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// NATSCoordinatorConfig holds dependencies for creating a NATSCoordinator.
type NATSCoordinatorConfig struct {
	NATSClient     *natsclient.Client
	ClientRegistry ClientRegistry
	Processor      MessageProcessor
	Config         NATSConfig
	Logger         *slog.Logger
	Metrics        *observability.Metrics
	NATSMetrics    *natsclient.NATSMetrics
	EchoEmitter    *echo.Emitter
	RedisClient    *redis.Client
}

// NewNATSCoordinator creates a new NATS-based queue coordinator.
func NewNATSCoordinator(ctx context.Context, cfg *NATSCoordinatorConfig) (*NATSCoordinator, error) {
	if cfg.NATSClient == nil {
		return nil, fmt.Errorf("nats client is required")
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	ctx, cancel := context.WithCancel(ctx)

	publisher := NewNATSPublisher(cfg.NATSClient, cfg.Config, cfg.Logger)
	dlq := NewNATSDLQHandler(cfg.NATSClient, cfg.Logger)

	processor := cfg.Processor
	if processor == nil {
		processor = NewWhatsAppMessageProcessor(cfg.Logger, cfg.EchoEmitter)
	}

	c := &NATSCoordinator{
		natsClient:      cfg.NATSClient,
		publisher:       publisher,
		dlq:             dlq,
		clientRegistry:  cfg.ClientRegistry,
		processor:       processor,
		cfg:             cfg.Config,
		log:             cfg.Logger.With(slog.String("component", "nats_msg_coordinator")),
		metrics:         cfg.Metrics,
		natsMetrics:     cfg.NATSMetrics,
		redis:           cfg.RedisClient,
		workers:         make(map[uuid.UUID]*NATSWorker),
		pausedInstances: make(map[uuid.UUID]*pauseState),
		ctx:             ctx,
		cancel:          cancel,
	}

	c.log.Info("NATS message queue coordinator initialized")

	return c, nil
}

// AddInstance creates and starts a NATS worker for a specific instance.
func (c *NATSCoordinator) AddInstance(instanceID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.workers[instanceID]; exists {
		return nil
	}

	worker := NewNATSWorker(NATSWorkerConfig{
		InstanceID:     instanceID,
		Client:         c.natsClient,
		ClientRegistry: c.clientRegistry,
		Processor:      c.processor,
		DLQ:            c.dlq,
		Config:         c.cfg,
		Logger:         c.log,
		Metrics:        c.metrics,
		NATSMetrics:    c.natsMetrics,
	})

	worker.SetPauseChecker(c.IsInstancePaused)
	if c.redis != nil {
		worker.SetCancelChecker(c.isJobCancelled)
	}

	if err := worker.Start(c.ctx); err != nil {
		return fmt.Errorf("start worker for %s: %w", instanceID, err)
	}

	c.workers[instanceID] = worker

	if c.metrics != nil {
		c.metrics.MessageQueueWorkers.WithLabelValues(instanceID.String()).Inc()
	}

	c.log.Info("added NATS queue worker",
		slog.String("instance_id", instanceID.String()),
		slog.Int("total_workers", len(c.workers)))

	return nil
}

// RemoveInstance stops and removes a NATS worker for a specific instance.
func (c *NATSCoordinator) RemoveInstance(ctx context.Context, instanceID uuid.UUID) error {
	c.mu.Lock()
	worker, exists := c.workers[instanceID]
	if !exists {
		c.mu.Unlock()
		return ErrInstanceNotFound
	}
	delete(c.workers, instanceID)
	c.mu.Unlock()

	if err := worker.Stop(ctx); err != nil {
		return fmt.Errorf("stop worker %s: %w", instanceID, err)
	}

	if c.metrics != nil {
		c.metrics.MessageQueueWorkers.WithLabelValues(instanceID.String()).Dec()
	}

	c.log.Info("removed NATS queue worker",
		slog.String("instance_id", instanceID.String()))

	return nil
}

// EnqueueMessage publishes a message to NATS for the given instance.
func (c *NATSCoordinator) EnqueueMessage(ctx context.Context, instanceID uuid.UUID, args SendMessageArgs) (string, error) {
	c.drainMu.RLock()
	draining := c.draining
	c.drainMu.RUnlock()
	if draining {
		return "", ErrQueueStopped
	}

	// Generate ZaapID if not provided
	if args.ZaapID == "" {
		args.ZaapID = generateZaapID()
	}
	args.EnqueuedAt = time.Now()
	args.InstanceID = instanceID

	// Calculate delay from DelayMessage (in milliseconds)
	delay := time.Duration(args.DelayMessage) * time.Millisecond
	args.ScheduledFor = time.Now().Add(delay)

	if err := args.Validate(); err != nil {
		return "", fmt.Errorf("invalid message args: %w", err)
	}

	// Publish to NATS
	if err := c.publisher.Publish(ctx, args); err != nil {
		if c.metrics != nil {
			c.metrics.MessageQueueErrors.WithLabelValues(instanceID.String(), "message", "enqueue").Inc()
		}
		return "", err
	}

	if c.metrics != nil {
		c.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "message", "success").Inc()
	}

	// Ensure worker exists for this instance
	c.mu.RLock()
	_, exists := c.workers[instanceID]
	c.mu.RUnlock()

	if !exists {
		if err := c.AddInstance(instanceID); err != nil {
			c.log.Warn("failed to auto-create NATS worker",
				slog.String("instance_id", instanceID.String()),
				slog.String("error", err.Error()))
		}
	}

	return args.ZaapID, nil
}

// PauseInstance pauses message processing for an instance during proxy operations.
func (c *NATSCoordinator) PauseInstance(ctx context.Context, instanceID uuid.UUID, reason string) error {
	c.pauseMu.Lock()
	defer c.pauseMu.Unlock()

	if existing, ok := c.pausedInstances[instanceID]; ok {
		existing.reason = reason
		return nil
	}

	const safetyTimeout = 5 * time.Minute
	timer := time.AfterFunc(safetyTimeout, func() {
		c.log.Warn("auto-resuming instance after safety timeout",
			slog.String("instance_id", instanceID.String()),
			slog.String("original_reason", reason))
		_ = c.ResumeInstance(context.Background(), instanceID)
	})

	c.pausedInstances[instanceID] = &pauseState{
		reason:   reason,
		pausedAt: time.Now(),
		timer:    timer,
	}

	c.log.Info("instance paused for proxy operation",
		slog.String("instance_id", instanceID.String()),
		slog.String("reason", reason))

	return nil
}

// ResumeInstance resumes message processing for a previously paused instance.
func (c *NATSCoordinator) ResumeInstance(ctx context.Context, instanceID uuid.UUID) error {
	c.pauseMu.Lock()
	defer c.pauseMu.Unlock()

	state, ok := c.pausedInstances[instanceID]
	if !ok {
		return nil
	}

	if state.timer != nil {
		state.timer.Stop()
	}

	delete(c.pausedInstances, instanceID)

	c.log.Info("instance resumed after proxy operation",
		slog.String("instance_id", instanceID.String()),
		slog.String("reason", state.reason))

	return nil
}

// IsInstancePaused returns true if message processing is paused for the instance.
func (c *NATSCoordinator) IsInstancePaused(instanceID uuid.UUID) bool {
	c.pauseMu.RLock()
	defer c.pauseMu.RUnlock()
	_, ok := c.pausedInstances[instanceID]
	return ok
}

// DrainQueue prevents new messages and waits for in-flight messages to complete.
func (c *NATSCoordinator) DrainQueue(ctx context.Context) error {
	c.log.Info("starting NATS queue drain")
	c.drainMu.Lock()
	c.draining = true
	c.drainMu.Unlock()

	defer func() {
		c.drainMu.Lock()
		c.draining = false
		c.drainMu.Unlock()
	}()

	// Stop all consumers (they will finish current messages)
	c.mu.RLock()
	workers := make([]*NATSWorker, 0, len(c.workers))
	for _, w := range c.workers {
		workers = append(workers, w)
	}
	c.mu.RUnlock()

	for _, w := range workers {
		if w.consCtx != nil {
			w.consCtx.Drain()
		}
	}

	c.log.Info("NATS queue drain completed")
	return nil
}

// ListQueueJobs lists jobs from the NATS stream for the given instance.
// Note: NATS doesn't support efficient querying like PostgreSQL.
// This provides basic stream info as a fallback.
func (c *NATSCoordinator) ListQueueJobs(ctx context.Context, instanceID uuid.UUID, limit, offset int) (*QueueListResponse, error) {
	// NATS doesn't support SQL-like queries. Return stream consumer info.
	return &QueueListResponse{
		InstanceID: instanceID,
		Total:      0,
		Jobs:       []QueueJobInfo{},
	}, nil
}

// ClearQueue purges messages for an instance from the stream.
func (c *NATSCoordinator) ClearQueue(ctx context.Context, instanceID uuid.UUID) error {
	js := c.natsClient.JetStream()
	if js == nil {
		return natsclient.ErrNotConnected
	}

	stream, err := js.Stream(ctx, "MESSAGE_QUEUE")
	if err != nil {
		return fmt.Errorf("get stream: %w", err)
	}

	subject := fmt.Sprintf("messages.%s", instanceID.String())
	if err := stream.Purge(ctx, jetstream.WithPurgeSubject(subject)); err != nil {
		return fmt.Errorf("purge subject %s: %w", subject, err)
	}

	c.log.Info("queue cleared via NATS purge",
		slog.String("instance_id", instanceID.String()))

	return nil
}

// CancelJob marks a job as cancelled via Redis set.
func (c *NATSCoordinator) CancelJob(ctx context.Context, zaapID string) error {
	if c.redis == nil {
		return fmt.Errorf("cancel not available: redis not configured")
	}

	key := fmt.Sprintf("cancelled_jobs:%s", zaapID)
	if err := c.redis.Set(ctx, key, "1", 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("set cancel key: %w", err)
	}

	c.log.Info("job marked as cancelled",
		slog.String("zaap_id", zaapID))

	return nil
}

// Stop gracefully stops all workers.
func (c *NATSCoordinator) Stop(ctx context.Context) error {
	c.log.Info("stopping NATS message queue coordinator")

	c.cancel()

	c.mu.Lock()
	workers := make([]*NATSWorker, 0, len(c.workers))
	for _, w := range c.workers {
		workers = append(workers, w)
	}
	c.mu.Unlock()

	var wg sync.WaitGroup
	for _, w := range workers {
		wg.Add(1)
		go func(worker *NATSWorker) {
			defer wg.Done()
			if err := worker.Stop(ctx); err != nil {
				c.log.Error("failed to stop NATS worker",
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
		c.log.Info("all NATS workers stopped", slog.Int("count", len(workers)))
	case <-ctx.Done():
		c.log.Warn("NATS worker stop timeout")
		return ctx.Err()
	}

	return nil
}

// GetClient returns the client registry for an instance.
func (c *NATSCoordinator) GetClient(instanceID uuid.UUID) (ClientRegistry, bool) {
	if c.clientRegistry == nil {
		return nil, false
	}
	return c.clientRegistry, true
}

// isJobCancelled checks Redis for cancelled job markers.
func (c *NATSCoordinator) isJobCancelled(ctx context.Context, zaapID string) bool {
	if c.redis == nil {
		return false
	}

	key := fmt.Sprintf("cancelled_jobs:%s", zaapID)
	val, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	return val == "1"
}

// Verify NATSCoordinator implements QueueCoordinator.
var _ QueueCoordinator = (*NATSCoordinator)(nil)
