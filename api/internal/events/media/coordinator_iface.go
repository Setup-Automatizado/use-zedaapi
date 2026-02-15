package media

import (
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow"
)

// MediaCoordinatorProvider defines the interface for media coordinators.
// Both the PostgreSQL-based MediaCoordinator and NATSMediaCoordinator implement this.
type MediaCoordinatorProvider interface {
	RegisterInstance(instanceID uuid.UUID, client *whatsmeow.Client) error
	UnregisterInstance(instanceID uuid.UUID) error
}

// Compile-time checks.
var (
	_ MediaCoordinatorProvider = (*MediaCoordinator)(nil)
	_ MediaCoordinatorProvider = (*NATSMediaCoordinator)(nil)
)
