package dispatch

import (
	"context"

	"github.com/google/uuid"
)

// DispatchCoordinator defines the interface for event dispatch coordination.
// Both PostgreSQL-based Coordinator and NATS-based NATSDispatchCoordinator implement this.
type DispatchCoordinator interface {
	// Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Instance management
	RegisterInstance(ctx context.Context, instanceID uuid.UUID) error
	UnregisterInstance(ctx context.Context, instanceID uuid.UUID) error

	// Status interception
	SetStatusInterceptor(interceptor StatusInterceptor)

	// Info
	GetWorkerCount() int
	IsInstanceRegistered(instanceID uuid.UUID) bool
}

// Verify that Coordinator implements DispatchCoordinator at compile time.
var _ DispatchCoordinator = (*Coordinator)(nil)
