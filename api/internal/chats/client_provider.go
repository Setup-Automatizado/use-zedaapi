package chats

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	whatsmeowclient "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
	internalwhatsmeow "go.mau.fi/whatsmeow/api/internal/whatsmeow"
	"go.mau.fi/whatsmeow/types"
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

// Get returns a chats client for the given instance ID, ensuring a WhatsApp client is available.
func (p *RegistryClientProvider) Get(ctx context.Context, instanceID uuid.UUID) (Client, error) {
	if p.registry == nil {
		return nil, fmt.Errorf("client registry not configured")
	}

	if client, ok := p.registry.GetClient(instanceID); ok && client != nil {
		return newWhatsAppClientAdapter(client), nil
	}

	if p.repo == nil {
		return nil, ErrClientNotConnected
	}

	logger := logging.ContextLogger(ctx, p.log).With(
		slog.String("component", "chats_client_provider"),
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

	return newWhatsAppClientAdapter(client), nil
}

// whatsAppClientAdapter wraps a whatsmeow.Client to implement the chats.Client interface.
type whatsAppClientAdapter struct {
	client *whatsmeowclient.Client
}

// newWhatsAppClientAdapter creates a new adapter for the whatsmeow client.
func newWhatsAppClientAdapter(client *whatsmeowclient.Client) *whatsAppClientAdapter {
	return &whatsAppClientAdapter{client: client}
}

// GetAllContacts retrieves all contacts from the whatsmeow client's contact store.
func (a *whatsAppClientAdapter) GetAllContacts(ctx context.Context) (map[types.JID]types.ContactInfo, error) {
	if a.client == nil || a.client.Store == nil || a.client.Store.Contacts == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.Store.Contacts.GetAllContacts(ctx)
}

// GetJoinedGroups retrieves all groups the user has joined.
func (a *whatsAppClientAdapter) GetJoinedGroups(ctx context.Context) ([]*types.GroupInfo, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.GetJoinedGroups(ctx)
}

// GetChatSettings retrieves local settings for a specific chat.
func (a *whatsAppClientAdapter) GetChatSettings(ctx context.Context, chat types.JID) (types.LocalChatSettings, error) {
	if a.client == nil || a.client.Store == nil || a.client.Store.ChatSettings == nil {
		return types.LocalChatSettings{}, nil // Return empty settings rather than error
	}
	return a.client.Store.ChatSettings.GetChatSettings(ctx, chat)
}

// toInstanceInfo converts an instances.Instance to internalwhatsmeow.InstanceInfo.
func toInstanceInfo(inst instances.Instance) internalwhatsmeow.InstanceInfo {
	return internalwhatsmeow.InstanceInfo{
		ID:            inst.ID,
		Name:          inst.Name,
		SessionName:   inst.SessionName,
		InstanceToken: inst.InstanceToken,
		StoreJID:      inst.StoreJID,
	}
}
