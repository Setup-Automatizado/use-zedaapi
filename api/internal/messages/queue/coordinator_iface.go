package queue

import (
	"context"

	"github.com/google/uuid"
)

// QueueCoordinator defines the interface for message queue coordination.
// Both PostgreSQL-based Coordinator and NATS-based NATSCoordinator implement this.
type QueueCoordinator interface {
	// Instance lifecycle
	AddInstance(instanceID uuid.UUID) error
	RemoveInstance(ctx context.Context, instanceID uuid.UUID) error

	// Message operations
	EnqueueMessage(ctx context.Context, instanceID uuid.UUID, args SendMessageArgs) (string, error)

	// Pause/Resume for proxy operations
	PauseInstance(ctx context.Context, instanceID uuid.UUID, reason string) error
	ResumeInstance(ctx context.Context, instanceID uuid.UUID) error
	IsInstancePaused(instanceID uuid.UUID) bool

	// Queue management
	DrainQueue(ctx context.Context) error
	ListQueueJobs(ctx context.Context, instanceID uuid.UUID, limit, offset int) (*QueueListResponse, error)
	ClearQueue(ctx context.Context, instanceID uuid.UUID) error
	CancelJob(ctx context.Context, zaapID string) error

	// Lifecycle
	Stop(ctx context.Context) error

	// Client access
	GetClient(instanceID uuid.UUID) (ClientRegistry, bool)
}

// Verify that Coordinator implements QueueCoordinator at compile time.
var _ QueueCoordinator = (*Coordinator)(nil)
