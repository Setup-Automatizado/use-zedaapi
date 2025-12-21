package queue

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	wameow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// ClientRegistry defines the interface for retrieving WhatsApp clients
type ClientRegistry interface {
	GetClient(instanceID string) (*wameow.Client, bool)
}

// MessageProcessor defines the interface for processing messages
// This allows for custom message processing logic
type MessageProcessor interface {
	Process(ctx context.Context, client *wameow.Client, payload []byte) error
}

// Worker processes messages for a specific WhatsApp instance
// It maintains FIFO ordering by processing one message at a time
type Worker struct {
	instanceID     uuid.UUID
	repo           *Repository
	clientRegistry ClientRegistry
	processor      MessageProcessor
	config         *Config
	log            *slog.Logger
	metrics        *observability.Metrics

	// Control channels
	stopCh chan struct{}
	doneCh chan struct{}

	// State
	running bool
}

// NewWorker creates a new worker for a specific instance
func NewWorker(
	instanceID uuid.UUID,
	repo *Repository,
	clientRegistry ClientRegistry,
	processor MessageProcessor,
	config *Config,
	log *slog.Logger,
	metrics *observability.Metrics,
) *Worker {
	return &Worker{
		instanceID:     instanceID,
		repo:           repo,
		clientRegistry: clientRegistry,
		processor:      processor,
		config:         config,
		log:            log.With(slog.String("instance_id", instanceID.String())),
		metrics:        metrics,
		stopCh:         make(chan struct{}),
		doneCh:         make(chan struct{}),
	}
}

// Start begins processing messages for this instance
func (w *Worker) Start(ctx context.Context) {
	w.running = true
	go w.processLoop(ctx)
}

// Stop gracefully stops the worker
func (w *Worker) Stop(ctx context.Context) error {
	if !w.running {
		return nil
	}

	w.log.Info("stopping worker")
	close(w.stopCh)

	// Wait for worker to finish current message (with timeout)
	select {
	case <-w.doneCh:
		w.log.Info("worker stopped gracefully")
		return nil
	case <-ctx.Done():
		w.log.Warn("worker stop timeout")
		return ctx.Err()
	}
}

// processLoop is the main worker loop
// Uses PostgreSQL LISTEN/NOTIFY for <10ms latency with polling fallback
func (w *Worker) processLoop(ctx context.Context) {
	defer close(w.doneCh)

	// Set up LISTEN/NOTIFY
	notificationChan, err := w.repo.ListenForNotifications(ctx)
	if err != nil {
		w.log.Warn("failed to setup LISTEN/NOTIFY, falling back to polling only",
			slog.String("error", err.Error()))
		notificationChan = nil // Disable notifications, use polling only
	}

	// Fallback polling ticker (in case notifications are missed or disabled)
	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	if notificationChan != nil {
		w.log.Info("worker started with LISTEN/NOTIFY + polling fallback",
			slog.Duration("poll_interval", w.config.PollInterval))
	} else {
		w.log.Info("worker started with polling only",
			slog.Duration("poll_interval", w.config.PollInterval))
	}

	for {
		select {
		case <-ctx.Done():
			w.log.Info("worker stopped (context cancelled)")
			return
		case <-w.stopCh:
			w.log.Info("worker stopped (stop signal)")
			return
		case instanceID, ok := <-notificationChan:
			if !ok {
				// Notification channel closed, disable and rely on polling
				w.log.Warn("notification channel closed, falling back to polling only")
				notificationChan = nil
				continue
			}
			// Only process if notification is for this instance
			if instanceID == w.instanceID {
				if err := w.processNext(ctx); err != nil {
					w.log.Error("process message failed (notification)", slog.String("error", err.Error()))
				}
			}
		case <-ticker.C:
			// Fallback polling (catches missed notifications or when LISTEN/NOTIFY disabled)
			if err := w.processNext(ctx); err != nil {
				w.log.Error("process message failed (polling)", slog.String("error", err.Error()))
			}
		}
	}
}

// processNext retrieves and processes the next pending message
func (w *Worker) processNext(ctx context.Context) error {
	// Dequeue next message
	msg, err := w.repo.Dequeue(ctx, w.instanceID)
	if err != nil {
		return fmt.Errorf("dequeue: %w", err)
	}

	// No messages available
	if msg == nil {
		return nil
	}

	w.log.Debug("processing message",
		slog.Int64("message_id", msg.ID),
		slog.Int("attempt", msg.Attempts),
		slog.Int("max_attempts", msg.MaxAttempts))

	// Get WhatsApp client
	client, ok := w.clientRegistry.GetClient(w.instanceID.String())
	if !ok {
		w.log.Warn("whatsapp client not found, rescheduling message",
			slog.Int64("message_id", msg.ID))
		return w.repo.RescheduleOnDisconnect(ctx, msg.ID, w.config.DisconnectRetryDelay)
	}

	// Check if client is connected
	if !client.IsConnected() {
		w.log.Warn("whatsapp not connected, rescheduling message",
			slog.Int64("message_id", msg.ID))
		return w.repo.RescheduleOnDisconnect(ctx, msg.ID, w.config.DisconnectRetryDelay)
	}

	// Process message
	startTime := time.Now()
	err = w.processor.Process(ctx, client, msg.Payload)
	duration := time.Since(startTime)

	if err != nil {
		w.log.Error("message processing failed",
			slog.Int64("message_id", msg.ID),
			slog.String("error", err.Error()),
			slog.Int("attempt", msg.Attempts),
			slog.Duration("duration", duration))

		// Update error metrics
		if w.metrics != nil {
			w.metrics.MessageQueueErrors.WithLabelValues(w.instanceID.String(), "message", "processing_error").Inc()
		}

		// Handle failure with retry logic
		return w.handleFailure(ctx, msg, err.Error())
	}

	// Mark as completed
	if err := w.repo.MarkCompleted(ctx, msg.ID); err != nil {
		w.log.Error("mark completed failed",
			slog.Int64("message_id", msg.ID),
			slog.String("error", err.Error()))
		return err
	}

	// Update success metrics
	if w.metrics != nil {
		w.metrics.MessageQueueProcessed.WithLabelValues(w.instanceID.String(), "message", "success").Inc()
		w.metrics.MessageQueueDuration.WithLabelValues(w.instanceID.String(), "message").Observe(duration.Seconds())
	}

	w.log.Info("message processed successfully",
		slog.Int64("message_id", msg.ID),
		slog.Duration("duration", duration))

	return nil
}

// handleFailure handles a failed message with retry logic
func (w *Worker) handleFailure(ctx context.Context, msg *QueueMessage, errorMsg string) error {
	// Calculate backoff delay
	backoff := w.config.CalculateBackoff(msg.Attempts)

	w.log.Debug("calculating retry backoff",
		slog.Int64("message_id", msg.ID),
		slog.Int("attempt", msg.Attempts),
		slog.Duration("backoff", backoff))

	// Mark failed (will retry or move to DLQ)
	willRetry, err := w.repo.MarkFailed(ctx, msg.ID, errorMsg, backoff)
	if err != nil {
		return fmt.Errorf("mark failed: %w", err)
	}

	if willRetry {
		w.log.Info("message will be retried",
			slog.Int64("message_id", msg.ID),
			slog.Int("attempt", msg.Attempts),
			slog.Int("max_attempts", msg.MaxAttempts),
			slog.Duration("retry_after", backoff))

		// Update retry metrics
		if w.metrics != nil {
			w.metrics.MessageQueueRetries.WithLabelValues(w.instanceID.String(), "message", strconv.Itoa(msg.Attempts+1)).Inc()
		}
	} else {
		w.log.Warn("message moved to DLQ after max retries",
			slog.Int64("message_id", msg.ID),
			slog.Int("final_attempts", msg.Attempts),
			slog.String("final_error", errorMsg))

		// Update DLQ metrics
		if w.metrics != nil {
			w.metrics.MessageQueueDLQSize.Inc()
			w.metrics.MessageQueueProcessed.WithLabelValues(w.instanceID.String(), "message", "failed").Inc()
		}
	}

	return nil
}

// DefaultMessageProcessor is a simple processor that logs the payload
// In production, this should be replaced with actual WhatsApp sending logic
type DefaultMessageProcessor struct {
	log *slog.Logger
}

func NewDefaultMessageProcessor(log *slog.Logger) *DefaultMessageProcessor {
	return &DefaultMessageProcessor{log: log}
}

func (p *DefaultMessageProcessor) Process(ctx context.Context, client *wameow.Client, payload []byte) error {
	// TODO: Implement actual message sending logic here
	// This is a placeholder that should be replaced with:
	// 1. Parse payload to get message details (recipient, content, type)
	// 2. Use client to send message via WhatsApp
	// 3. Handle WhatsApp-specific errors

	p.log.Debug("processing message payload",
		slog.Int("payload_size", len(payload)))

	// Placeholder - replace with actual sending logic
	// Example:
	// var msg types.MessageRequest
	// if err := json.Unmarshal(payload, &msg); err != nil {
	//     return fmt.Errorf("unmarshal payload: %w", err)
	// }
	//
	// resp, err := client.SendMessage(ctx, msg.To, msg.Content)
	// if err != nil {
	//     return fmt.Errorf("send whatsapp message: %w", err)
	// }

	return nil
}
