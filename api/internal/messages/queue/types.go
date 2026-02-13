package queue

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// QueueMessage represents a message in the queue system
type QueueMessage struct {
	ID            int64           `db:"id"`
	InstanceID    uuid.UUID       `db:"instance_id"`
	Payload       json.RawMessage `db:"payload"`
	Status        string          `db:"status"`
	ScheduledAt   time.Time       `db:"scheduled_at"`
	CreatedAt     time.Time       `db:"created_at"`
	Attempts      int             `db:"attempts"`
	MaxAttempts   int             `db:"max_attempts"`
	LastError     *string         `db:"last_error"`
	LastAttemptAt *time.Time      `db:"last_attempt_at"`
	ProcessedAt   *time.Time      `db:"processed_at"`
	ProcessingKey string          `db:"processing_key"` // Phone number for FIFO by recipient
}

// MessageStatus represents the current state of a message
type MessageStatus string

const (
	StatusPending    MessageStatus = "pending"
	StatusProcessing MessageStatus = "processing"
	StatusCompleted  MessageStatus = "completed"
	StatusFailed     MessageStatus = "failed"
)

// DLQMessage represents a message that has been moved to the Dead Letter Queue
type DLQMessage struct {
	ID         int64           `db:"id"`
	OriginalID int64           `db:"original_id"`
	InstanceID uuid.UUID       `db:"instance_id"`
	Payload    json.RawMessage `db:"payload"`
	Error      string          `db:"error"`
	Attempts   int             `db:"attempts"`
	CreatedAt  time.Time       `db:"created_at"`
	MovedAt    time.Time       `db:"moved_at"`
}

// Config holds configuration for the message queue system
type Config struct {
	// Polling configuration
	PollInterval time.Duration // How often to check for new messages (default: 100ms)
	BatchSize    int           // Messages to fetch per poll (default: 1 for strict FIFO)

	// Retry configuration
	MaxAttempts       int           // Maximum retry attempts before moving to DLQ (default: 3)
	InitialBackoff    time.Duration // First retry delay (default: 1s)
	MaxBackoff        time.Duration // Maximum retry delay (default: 5m)
	BackoffMultiplier float64       // Exponential backoff multiplier (default: 2.0)

	// WhatsApp disconnection handling
	DisconnectRetryDelay time.Duration // Delay when WhatsApp is offline (default: 30s)
	ProxyRetryDelay      time.Duration // Delay when instance paused for proxy operation (default: 2m)

	// Cleanup configuration
	CompletedRetention time.Duration // How long to keep completed messages (default: 24h)
	FailedRetention    time.Duration // How long to keep DLQ messages (default: 7d)

	// Performance configuration
	WorkersPerInstance int // Number of workers per instance (must be 1 for FIFO)
}

// DefaultConfig returns sensible defaults for the message queue
func DefaultConfig() *Config {
	return &Config{
		// Polling
		PollInterval: 100 * time.Millisecond,
		BatchSize:    1, // CRITICAL: Must be 1 for FIFO ordering

		// Retry
		MaxAttempts:       3,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        5 * time.Minute,
		BackoffMultiplier: 2.0,

		// Disconnection
		DisconnectRetryDelay: 30 * time.Second,
		ProxyRetryDelay:      2 * time.Minute,

		// Cleanup
		CompletedRetention: 24 * time.Hour,
		FailedRetention:    7 * 24 * time.Hour,

		// Performance
		WorkersPerInstance: 1, // CRITICAL: Must be 1 for FIFO ordering
	}
}

// Validate ensures the configuration is valid for FIFO ordering
func (c *Config) Validate() error {
	if c.WorkersPerInstance != 1 {
		return ErrInvalidWorkerCount
	}
	if c.BatchSize != 1 {
		return ErrInvalidBatchSize
	}
	return nil
}

// QueueStats holds statistics for a specific instance queue
type QueueStats struct {
	InstanceID        uuid.UUID
	PendingCount      int
	ProcessingCount   int
	CompletedCount    int
	FailedCount       int
	DLQCount          int
	OldestPendingAge  *time.Duration
	AvgProcessingTime *time.Duration
}

// Error types
var (
	ErrInvalidWorkerCount = &QueueError{Code: "INVALID_WORKER_COUNT", Message: "WorkersPerInstance must be 1 for FIFO ordering guarantee"}
	ErrInvalidBatchSize   = &QueueError{Code: "INVALID_BATCH_SIZE", Message: "BatchSize must be 1 for FIFO ordering guarantee"}
	ErrQueueStopped       = &QueueError{Code: "QUEUE_STOPPED", Message: "queue has been stopped"}
	ErrInstancePaused     = &QueueError{Code: "INSTANCE_PAUSED", Message: "instance is paused for proxy operation"}
	ErrPayloadTooLarge    = &QueueError{Code: "PAYLOAD_TOO_LARGE", Message: "message payload exceeds maximum size"}
)

// QueueError represents a queue-specific error
type QueueError struct {
	Code    string
	Message string
	Err     error
}

func (e *QueueError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *QueueError) Unwrap() error {
	return e.Err
}

// CalculateBackoff calculates the backoff duration for a given attempt
func (c *Config) CalculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return c.InitialBackoff
	}

	// Exponential backoff: InitialBackoff * (Multiplier ^ attempt)
	backoff := float64(c.InitialBackoff) * pow(c.BackoffMultiplier, float64(attempt))

	// Cap at MaxBackoff
	if time.Duration(backoff) > c.MaxBackoff {
		return c.MaxBackoff
	}

	return time.Duration(backoff)
}

// pow is a simple integer power function for exponential backoff
func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}
