package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// Coordinator manages the lifecycle of queue workers
// It creates workers dynamically as instances connect/disconnect
type Coordinator struct {
	repo           *Repository
	clientRegistry ClientRegistry
	processor      MessageProcessor
	config         *Config
	log            *slog.Logger
	metrics        *observability.Metrics

	// Worker management
	mu      sync.RWMutex
	workers map[uuid.UUID]*Worker

	// Cleanup job
	cleanupTicker  *time.Ticker
	cleanupStop    chan struct{}
	cleanupTimeout time.Duration

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Drain control
	drainMu  sync.RWMutex
	draining bool
}

// CoordinatorConfig holds configuration for the coordinator
type CoordinatorConfig struct {
	Pool           *pgxpool.Pool
	ClientRegistry ClientRegistry
	Processor      MessageProcessor
	Config         *Config
	Logger         *slog.Logger
	Metrics        *observability.Metrics

	// Cleanup configuration
	CleanupInterval time.Duration // How often to run cleanup (default: 1h)
	CleanupTimeout  time.Duration // Timeout for stuck messages (default: 5m)
}

// NewCoordinator creates a new queue coordinator
func NewCoordinator(ctx context.Context, cfg *CoordinatorConfig) (*Coordinator, error) {
	// Validate configuration
	if cfg.Config == nil {
		cfg.Config = DefaultConfig()
	}
	if err := cfg.Config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	if cfg.CleanupInterval == 0 {
		cfg.CleanupInterval = 1 * time.Hour
	}

	if cfg.CleanupTimeout == 0 {
		cfg.CleanupTimeout = 5 * time.Minute
	}

	// Create repository
	repo := NewRepository(cfg.Pool)

	// Create processor if not provided
	processor := cfg.Processor
	if processor == nil {
		processor = NewDefaultMessageProcessor(cfg.Logger)
	}

	ctx, cancel := context.WithCancel(ctx)

	c := &Coordinator{
		repo:           repo,
		clientRegistry: cfg.ClientRegistry,
		processor:      processor,
		config:         cfg.Config,
		log:            cfg.Logger,
		metrics:        cfg.Metrics,
		workers:        make(map[uuid.UUID]*Worker),
		cleanupStop:    make(chan struct{}),
		cleanupTimeout: cfg.CleanupTimeout,
		ctx:            ctx,
		cancel:         cancel,
	}

	// Start cleanup job
	c.cleanupTicker = time.NewTicker(cfg.CleanupInterval)
	c.wg.Add(1)
	go c.cleanupLoop()

	c.log.Info("message queue coordinator initialized",
		slog.Int("max_workers_per_instance", cfg.Config.WorkersPerInstance),
		slog.Duration("poll_interval", cfg.Config.PollInterval),
		slog.Duration("cleanup_interval", cfg.CleanupInterval))

	return c, nil
}

func (c *Coordinator) setDraining(active bool) {
	c.drainMu.Lock()
	c.draining = active
	c.drainMu.Unlock()
}

func (c *Coordinator) isDraining() bool {
	c.drainMu.RLock()
	defer c.drainMu.RUnlock()
	return c.draining
}

// AddInstance creates and starts a worker for a specific instance
func (c *Coordinator) AddInstance(instanceID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if worker already exists
	if _, exists := c.workers[instanceID]; exists {
		c.log.Debug("worker already exists for instance",
			slog.String("instance_id", instanceID.String()))
		return nil
	}

	// Create new worker
	worker := NewWorker(
		instanceID,
		c.repo,
		c.clientRegistry,
		c.processor,
		c.config,
		c.log,
	)

	// Start worker
	worker.Start(c.ctx)

	// Track worker
	c.workers[instanceID] = worker

	c.log.Info("added queue worker for instance",
		slog.String("instance_id", instanceID.String()),
		slog.Int("total_workers", len(c.workers)))

	return nil
}

// RemoveInstance stops and removes a worker for a specific instance
func (c *Coordinator) RemoveInstance(ctx context.Context, instanceID uuid.UUID) error {
	c.mu.Lock()
	worker, exists := c.workers[instanceID]
	if !exists {
		c.mu.Unlock()
		return ErrInstanceNotFound
	}
	delete(c.workers, instanceID)
	c.mu.Unlock()

	// Stop worker
	if err := worker.Stop(ctx); err != nil {
		c.log.Error("failed to stop worker",
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()))
		return err
	}

	c.log.Info("removed queue worker for instance",
		slog.String("instance_id", instanceID.String()),
		slog.Int("total_workers", len(c.workers)))

	return nil
}

// Enqueue adds a message to the queue
// This is the main entry point for adding messages from HTTP handlers
func (c *Coordinator) Enqueue(ctx context.Context, instanceID uuid.UUID, payload interface{}, delay time.Duration) (int64, error) {
	if c.isDraining() {
		return 0, ErrQueueStopped
	}

	// Calculate scheduled time
	scheduledAt := time.Now().Add(delay)

	// Enqueue message
	id, err := c.repo.Enqueue(ctx, instanceID, payload, scheduledAt, c.config.MaxAttempts)
	if err != nil {
		return 0, fmt.Errorf("enqueue message: %w", err)
	}

	c.log.Debug("message enqueued",
		slog.String("instance_id", instanceID.String()),
		slog.Int64("message_id", id),
		slog.Duration("delay", delay))

	// Ensure worker exists for this instance
	c.mu.RLock()
	_, exists := c.workers[instanceID]
	c.mu.RUnlock()

	if !exists {
		// Auto-create worker if it doesn't exist
		if err := c.AddInstance(instanceID); err != nil {
			c.log.Warn("failed to auto-create worker",
				slog.String("instance_id", instanceID.String()),
				slog.String("error", err.Error()))
		}
	}

	return id, nil
}

// DrainQueue waits for all pending and in-flight messages to be processed.
// It prevents new messages from being enqueued while draining is active.
func (c *Coordinator) DrainQueue(ctx context.Context) error {
	if c == nil {
		return nil
	}

	c.log.Info("starting queue drain")
	c.setDraining(true)
	defer c.setDraining(false)

	start := time.Now()
	initial, err := c.repo.CountActiveMessages(ctx)
	if err != nil {
		c.log.Error("queue drain count failed", slog.String("error", err.Error()))
		return fmt.Errorf("count active messages: %w", err)
	}

	if initial == 0 {
		c.log.Info("queue drain skipped - no messages pending")
		if c.metrics != nil {
			c.metrics.QueueDrainDuration.Observe(time.Since(start).Seconds())
		}
		return nil
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	remaining := initial
	lastRemaining := initial

	for {
		if ctx.Err() != nil {
			c.log.Warn("queue drain cancelled", slog.Int("remaining", remaining), slog.String("error", ctx.Err().Error()))
			if c.metrics != nil {
				c.metrics.QueueDrainTimeouts.Inc()
			}
			return ctx.Err()
		}

		if remaining == 0 {
			break
		}

		select {
		case <-ticker.C:
			count, countErr := c.repo.CountActiveMessages(ctx)
			if countErr != nil {
				c.log.Error("queue drain count failed", slog.String("error", countErr.Error()))
				return fmt.Errorf("count active messages: %w", countErr)
			}
			remaining = count
			if remaining != lastRemaining {
				c.log.Info("queue drain progress",
					slog.Int("remaining", remaining))
				lastRemaining = remaining
			}
		case <-ctx.Done():
			c.log.Warn("queue drain deadline reached", slog.Int("remaining", remaining))
			if c.metrics != nil {
				c.metrics.QueueDrainTimeouts.Inc()
			}
			return ctx.Err()
		}
	}

	duration := time.Since(start)
	c.log.Info("queue drain completed",
		slog.Duration("duration", duration),
		slog.Int("messages", initial))

	if c.metrics != nil {
		c.metrics.QueueDrainDuration.Observe(duration.Seconds())
		c.metrics.QueueDrainedMessages.Add(float64(initial))
	}

	return nil
}

// GetStats returns statistics for a specific instance queue
func (c *Coordinator) GetStats(ctx context.Context, instanceID uuid.UUID) (*QueueStats, error) {
	return c.repo.GetStats(ctx, instanceID)
}

// ListInstances returns all instances with active workers
func (c *Coordinator) ListInstances() []uuid.UUID {
	c.mu.RLock()
	defer c.mu.RUnlock()

	instances := make([]uuid.UUID, 0, len(c.workers))
	for id := range c.workers {
		instances = append(instances, id)
	}

	return instances
}

// Stop gracefully stops all workers and cleanup jobs
func (c *Coordinator) Stop(ctx context.Context) error {
	c.log.Info("stopping message queue coordinator")

	// Stop cleanup job
	close(c.cleanupStop)
	c.cleanupTicker.Stop()

	// Cancel context to stop all workers
	c.cancel()

	// Wait for cleanup job to finish
	c.wg.Wait()

	// Stop all workers
	c.mu.Lock()
	workerCount := len(c.workers)
	workers := make([]*Worker, 0, workerCount)
	for _, worker := range c.workers {
		workers = append(workers, worker)
	}
	c.mu.Unlock()

	// Stop workers concurrently with timeout
	var wg sync.WaitGroup
	for _, worker := range workers {
		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()
			if err := w.Stop(ctx); err != nil {
				c.log.Error("failed to stop worker",
					slog.String("instance_id", w.instanceID.String()),
					slog.String("error", err.Error()))
			}
		}(worker)
	}

	// Wait for all workers to stop
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.log.Info("all workers stopped successfully",
			slog.Int("count", workerCount))
	case <-ctx.Done():
		c.log.Warn("worker stop timeout",
			slog.Int("count", workerCount))
		return ctx.Err()
	}

	return nil
}

// cleanupLoop runs periodic maintenance tasks
func (c *Coordinator) cleanupLoop() {
	defer c.wg.Done()

	interval := 1 * time.Hour
	c.cleanupTicker.Reset(interval)
	c.log.Info("cleanup job started",
		slog.Duration("interval", interval))

	for {
		select {
		case <-c.ctx.Done():
			c.log.Info("cleanup job stopped (context cancelled)")
			return
		case <-c.cleanupStop:
			c.log.Info("cleanup job stopped (stop signal)")
			return
		case <-c.cleanupTicker.C:
			c.runCleanup()
		}
	}
}

// runCleanup performs maintenance tasks
func (c *Coordinator) runCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Reset stuck messages
	stuck, err := c.repo.ResetStuckMessages(ctx, c.cleanupTimeout)
	if err != nil {
		c.log.Error("failed to reset stuck messages",
			slog.String("error", err.Error()))
	} else if stuck > 0 {
		c.log.Warn("reset stuck messages",
			slog.Int("count", stuck),
			slog.Duration("timeout", c.cleanupTimeout))
	}

	// Cleanup old messages
	cleaned, err := c.repo.CleanupOldMessages(
		ctx,
		c.config.CompletedRetention,
		c.config.FailedRetention,
	)
	if err != nil {
		c.log.Error("failed to cleanup old messages",
			slog.String("error", err.Error()))
	} else if cleaned > 0 {
		c.log.Info("cleaned up old messages",
			slog.Int("count", cleaned))
	}
}

// Health returns health status of the coordinator
func (c *Coordinator) Health(ctx context.Context) error {
	// Check if context is still valid
	select {
	case <-c.ctx.Done():
		return ErrQueueStopped
	default:
	}

	// Try a simple database query
	var count int
	err := c.repo.pool.QueryRow(ctx, "SELECT COUNT(*) FROM message_queue WHERE status = 'pending'").Scan(&count)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// EnqueueMessage is a wrapper method that accepts SendMessageArgs
// This maintains compatibility with existing HTTP handlers
func (c *Coordinator) EnqueueMessage(ctx context.Context, instanceID uuid.UUID, args SendMessageArgs) (string, error) {
	// Generate ZaapID if not provided
	if args.ZaapID == "" {
		args.ZaapID = generateZaapID()
	}

	// Set timestamps
	args.EnqueuedAt = time.Now()

	// Calculate delay from DelayMessage (in milliseconds)
	delay := time.Duration(args.DelayMessage) * time.Millisecond

	// Validate args
	if err := args.Validate(); err != nil {
		return "", fmt.Errorf("invalid message args: %w", err)
	}

	// Enqueue using existing method
	_, err := c.Enqueue(ctx, instanceID, args, delay)
	if err != nil {
		return "", err
	}

	return args.ZaapID, nil
}

// ListQueueJobs retrieves messages from the queue with pagination
// Returns QueueListResponse compatible with Z-API format
func (c *Coordinator) ListQueueJobs(ctx context.Context, instanceID uuid.UUID, limit, offset int) (*QueueListResponse, error) {
	// Get messages from repository
	messages, total, err := c.repo.ListMessages(ctx, instanceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}

	// Convert to QueueJobInfo format
	jobs := make([]QueueJobInfo, 0, len(messages))
	for _, msg := range messages {
		// Unmarshal payload to get SendMessageArgs
		var args SendMessageArgs
		if err := json.Unmarshal(msg.Payload, &args); err != nil {
			c.log.Warn("failed to unmarshal message payload",
				slog.Int64("message_id", msg.ID),
				slog.String("error", err.Error()))
			continue
		}

		// Build job info
		job := QueueJobInfo{
			ID:                msg.ID,
			ZaapID:            args.ZaapID,
			InstanceID:        msg.InstanceID,
			MessageType:       args.MessageType,
			Phone:             args.Phone,
			Status:            convertStatus(msg.Status),
			SequenceNumber:    msg.ID, // Use ID as sequence for FIFO
			ScheduledFor:      msg.ScheduledAt,
			CreatedAt:         msg.CreatedAt,
			Attempt:           msg.Attempts,
			MaxAttempts:       msg.MaxAttempts,
			WhatsAppMessageID: args.WhatsAppMessageID,
			DelayMessage:      args.DelayMessage,
			DelayTyping:       args.DelayTyping,
			TextContent:       args.TextContent,
		}

		// Add attempt timestamp if available
		if msg.LastAttemptAt != nil {
			job.AttemptedAt = msg.LastAttemptAt
		}

		// Add processed timestamp if available
		if msg.ProcessedAt != nil {
			job.FinalizedAt = msg.ProcessedAt
		}

		// Add errors if available
		if msg.LastError != nil {
			job.Errors = []string{*msg.LastError}
		}

		jobs = append(jobs, job)
	}

	return &QueueListResponse{
		InstanceID: instanceID,
		Total:      total,
		Jobs:       jobs,
	}, nil
}

// ClearQueue removes all pending/processing messages for an instance
func (c *Coordinator) ClearQueue(ctx context.Context, instanceID uuid.UUID) error {
	if err := c.repo.DeleteByInstance(ctx, instanceID); err != nil {
		return fmt.Errorf("clear queue: %w", err)
	}

	c.log.Info("queue cleared",
		slog.String("instance_id", instanceID.String()))

	return nil
}

// CancelJob cancels a specific message by its ZaapID
func (c *Coordinator) CancelJob(ctx context.Context, zaapID string) error {
	// Get message by ZaapID
	msg, err := c.repo.GetByZaapID(ctx, zaapID)
	if err != nil {
		return fmt.Errorf("get message: %w", err)
	}

	// Delete message
	if err := c.repo.DeleteByID(ctx, msg.ID); err != nil {
		return fmt.Errorf("delete message: %w", err)
	}

	c.log.Info("message cancelled",
		slog.String("zaap_id", zaapID),
		slog.Int64("message_id", msg.ID))

	return nil
}

// Helper functions

// convertStatus converts internal status to JobStatus
func convertStatus(status string) JobStatus {
	switch status {
	case "pending":
		return JobStatusAvailable
	case "processing":
		return JobStatusRunning
	case "completed":
		return JobStatusCompleted
	case "failed":
		return JobStatusDiscarded
	default:
		return JobStatusAvailable
	}
}

// generateZaapID generates a unique message ID
// Format: {timestamp}{random}
func generateZaapID() string {
	timestamp := time.Now().UnixNano()
	random := uuid.New().String()[:8]
	return fmt.Sprintf("%d%s", timestamp, random)
}
