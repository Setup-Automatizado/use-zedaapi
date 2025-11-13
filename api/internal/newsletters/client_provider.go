package newsletters

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

// Client exposes the WhatsApp APIs required by the newsletters service.
type Client interface {
	GetSubscribedNewsletters(ctx context.Context) ([]*types.NewsletterMetadata, error)
	CreateNewsletter(ctx context.Context, params whatsmeowclient.CreateNewsletterParams) (*types.NewsletterMetadata, error)
	UploadNewsletter(ctx context.Context, data []byte, mediaType whatsmeowclient.MediaType) (whatsmeowclient.UploadResponse, error)
	FollowNewsletter(id types.JID) error
	UnfollowNewsletter(id types.JID) error
	NewsletterToggleMute(id types.JID, mute bool) error
	GetNewsletterInfo(id types.JID) (*types.NewsletterMetadata, error)
	SendNewsletterMex(ctx context.Context, queryID string, variables any) ([]byte, error)
}

// ClientProvider resolves a WhatsApp client for newsletter operations.
type ClientProvider interface {
	Get(ctx context.Context, instanceID uuid.UUID) (Client, error)
}

// InstanceRepository provides access to instance data required to ensure WhatsApp clients exist.
type InstanceRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*instances.Instance, error)
}

// RegistryClientProvider ensures WhatsApp clients are present in the registry before newsletter operations execute.
type RegistryClientProvider struct {
	registry *internalwhatsmeow.ClientRegistry
	repo     InstanceRepository
	log      *slog.Logger
}

// NewRegistryClientProvider constructs a provider backed by the WhatsApp client registry.
func NewRegistryClientProvider(registry *internalwhatsmeow.ClientRegistry, repo InstanceRepository, log *slog.Logger) *RegistryClientProvider {
	return &RegistryClientProvider{
		registry: registry,
		repo:     repo,
		log:      log,
	}
}

// Get fetches or materialises a WhatsApp client for the supplied instance.
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
		slog.String("component", "newsletters_client_provider"),
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

type whatsAppClientAdapter struct {
	client *whatsmeowclient.Client
}

func newWhatsAppClientAdapter(client *whatsmeowclient.Client) *whatsAppClientAdapter {
	return &whatsAppClientAdapter{client: client}
}

func (a *whatsAppClientAdapter) GetSubscribedNewsletters(ctx context.Context) ([]*types.NewsletterMetadata, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.GetSubscribedNewsletters(ctx)
}

func (a *whatsAppClientAdapter) CreateNewsletter(ctx context.Context, params whatsmeowclient.CreateNewsletterParams) (*types.NewsletterMetadata, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.CreateNewsletter(ctx, params)
}

func (a *whatsAppClientAdapter) UploadNewsletter(ctx context.Context, data []byte, mediaType whatsmeowclient.MediaType) (whatsmeowclient.UploadResponse, error) {
	if a.client == nil {
		return whatsmeowclient.UploadResponse{}, ErrClientNotConnected
	}
	return a.client.UploadNewsletter(ctx, data, mediaType)
}

func (a *whatsAppClientAdapter) FollowNewsletter(id types.JID) error {
	if a.client == nil {
		return ErrClientNotConnected
	}
	return a.client.FollowNewsletter(context.Background(), id)
}

func (a *whatsAppClientAdapter) UnfollowNewsletter(id types.JID) error {
	if a.client == nil {
		return ErrClientNotConnected
	}
	return a.client.UnfollowNewsletter(context.Background(), id)
}

func (a *whatsAppClientAdapter) NewsletterToggleMute(id types.JID, mute bool) error {
	if a.client == nil {
		return ErrClientNotConnected
	}
	return a.client.NewsletterToggleMute(context.Background(), id, mute)
}

func (a *whatsAppClientAdapter) GetNewsletterInfo(id types.JID) (*types.NewsletterMetadata, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.GetNewsletterInfo(context.Background(), id)
}

func (a *whatsAppClientAdapter) SendNewsletterMex(ctx context.Context, queryID string, variables any) ([]byte, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.DangerousInternals().SendMexIQ(ctx, queryID, variables)
}

func toInstanceInfo(inst instances.Instance) internalwhatsmeow.InstanceInfo {
	return internalwhatsmeow.InstanceInfo{
		ID:            inst.ID,
		Name:          inst.Name,
		SessionName:   inst.SessionName,
		InstanceToken: inst.InstanceToken,
		StoreJID:      inst.StoreJID,
	}
}
