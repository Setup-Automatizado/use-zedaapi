package contacts

import (
	"context"
	"fmt"

	"log/slog"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
	internalwhatsmeow "go.mau.fi/whatsmeow/api/internal/whatsmeow"
)

// InstanceRepository provides access to instance data required to materialise WhatsApp clients.
type InstanceRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*instances.Instance, error)
}

// RegistryClientProvider fetches WhatsApp clients from the client registry, ensuring they exist when necessary.
type RegistryClientProvider struct {
	registry *internalwhatsmeow.ClientRegistry
	repo     InstanceRepository
	log      *slog.Logger
}

// NewRegistryClientProvider builds a provider backed by the main WhatsApp client registry.
func NewRegistryClientProvider(registry *internalwhatsmeow.ClientRegistry, repo InstanceRepository, log *slog.Logger) *RegistryClientProvider {
	return &RegistryClientProvider{
		registry: registry,
		repo:     repo,
		log:      log,
	}
}

// Get returns a contacts client for the given instance ID, ensuring a WhatsApp client is available.
func (p *RegistryClientProvider) Get(ctx context.Context, instanceID uuid.UUID) (Client, error) {
	if p.registry == nil {
		return nil, fmt.Errorf("client registry not configured")
	}

	// Try to get existing client first
	if client, ok := p.registry.GetClient(instanceID); ok && client != nil {
		return NewWhatsmeowClientProvider(client), nil
	}

	// Client doesn't exist, need to ensure it
	if p.repo == nil {
		return nil, fmt.Errorf("client not connected")
	}

	logger := logging.ContextLogger(ctx, p.log).With(
		slog.String("component", "contacts_client_provider"),
		slog.String("instance_id", instanceID.String()),
	)

	instance, err := p.repo.GetByID(ctx, instanceID)
	if err != nil {
		if err == instances.ErrInstanceNotFound {
			logger.Warn("instance not found while ensuring whatsapp client")
			return nil, err
		}
		logger.Error("failed to load instance",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("load instance: %w", err)
	}

	client, _, ensureErr := p.registry.EnsureClient(ctx, toInstanceInfo(*instance))
	if ensureErr != nil {
		logger.Error("ensure whatsapp client failed",
			slog.String("error", ensureErr.Error()))
		return nil, fmt.Errorf("ensure client: %w", ensureErr)
	}

	logger.Debug("whatsapp client ensured successfully")

	return NewWhatsmeowClientProvider(client), nil
}

// toInstanceInfo converts an instances.Instance to whatsmeow.InstanceInfo.
func toInstanceInfo(inst instances.Instance) internalwhatsmeow.InstanceInfo {
	return internalwhatsmeow.InstanceInfo{
		ID:            inst.ID,
		Name:          inst.Name,
		SessionName:   inst.SessionName,
		InstanceToken: inst.InstanceToken,
		StoreJID:      inst.StoreJID,
	}
}
