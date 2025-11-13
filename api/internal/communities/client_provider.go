package communities

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

// Client exposes the WhatsApp APIs required by the communities service.
type Client interface {
	GetJoinedGroups(ctx context.Context) ([]*types.GroupInfo, error)
	CreateGroup(ctx context.Context, req whatsmeowclient.ReqCreateGroup) (*types.GroupInfo, error)
	GetGroupInfo(jid types.JID) (*types.GroupInfo, error)
	GetGroupInviteLink(jid types.JID, reset bool) (string, error)
	GetSubGroups(community types.JID) ([]*types.GroupLinkTarget, error)
	GetLinkedGroupsParticipants(community types.JID) ([]types.JID, error)
	LinkGroup(parent, child types.JID) error
	UnlinkGroup(parent, child types.JID) error
	UpdateGroupParticipants(jid types.JID, participantChanges []types.JID, action whatsmeowclient.ParticipantChange) ([]types.GroupParticipant, error)
	SetGroupMemberAddMode(jid types.JID, mode types.GroupMemberAddMode) error
	SetGroupDescription(jid types.JID, description string) error
	LeaveGroup(jid types.JID) error
}

// ClientProvider resolves a WhatsApp client that supports community operations for a given instance.
type ClientProvider interface {
	Get(ctx context.Context, instanceID uuid.UUID) (Client, error)
}

// InstanceRepository abstracts the persistence necessary to materialise WhatsApp clients.
type InstanceRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*instances.Instance, error)
}

// RegistryClientProvider ensures WhatsApp clients exist in the registry before community operations run.
type RegistryClientProvider struct {
	registry *internalwhatsmeow.ClientRegistry
	repo     InstanceRepository
	log      *slog.Logger
}

// NewRegistryClientProvider constructs a provider backed by the global client registry.
func NewRegistryClientProvider(registry *internalwhatsmeow.ClientRegistry, repo InstanceRepository, log *slog.Logger) *RegistryClientProvider {
	return &RegistryClientProvider{
		registry: registry,
		repo:     repo,
		log:      log,
	}
}

// Get materialises a WhatsApp client for a given instance ID and returns it wrapped as a communities client.
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
		slog.String("component", "communities_client_provider"),
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

func (a *whatsAppClientAdapter) GetJoinedGroups(ctx context.Context) ([]*types.GroupInfo, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.GetJoinedGroups(ctx)
}

func (a *whatsAppClientAdapter) CreateGroup(ctx context.Context, req whatsmeowclient.ReqCreateGroup) (*types.GroupInfo, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.CreateGroup(ctx, req)
}

func (a *whatsAppClientAdapter) GetGroupInfo(jid types.JID) (*types.GroupInfo, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.GetGroupInfo(context.Background(), jid)
}

func (a *whatsAppClientAdapter) GetGroupInviteLink(jid types.JID, reset bool) (string, error) {
	if a.client == nil {
		return "", ErrClientNotConnected
	}
	return a.client.GetGroupInviteLink(context.Background(), jid, reset)
}

func (a *whatsAppClientAdapter) GetSubGroups(community types.JID) ([]*types.GroupLinkTarget, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.GetSubGroups(context.Background(), community)
}

func (a *whatsAppClientAdapter) GetLinkedGroupsParticipants(community types.JID) ([]types.JID, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.GetLinkedGroupsParticipants(context.Background(), community)
}

func (a *whatsAppClientAdapter) LinkGroup(parent, child types.JID) error {
	if a.client == nil {
		return ErrClientNotConnected
	}
	return a.client.LinkGroup(context.Background(), parent, child)
}

func (a *whatsAppClientAdapter) UnlinkGroup(parent, child types.JID) error {
	if a.client == nil {
		return ErrClientNotConnected
	}
	return a.client.UnlinkGroup(context.Background(), parent, child)
}

func (a *whatsAppClientAdapter) UpdateGroupParticipants(jid types.JID, participantChanges []types.JID, action whatsmeowclient.ParticipantChange) ([]types.GroupParticipant, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.UpdateGroupParticipants(context.Background(), jid, participantChanges, action)
}

func (a *whatsAppClientAdapter) SetGroupMemberAddMode(jid types.JID, mode types.GroupMemberAddMode) error {
	if a.client == nil {
		return ErrClientNotConnected
	}
	return a.client.SetGroupMemberAddMode(context.Background(), jid, mode)
}

func (a *whatsAppClientAdapter) SetGroupDescription(jid types.JID, description string) error {
	if a.client == nil {
		return ErrClientNotConnected
	}
	return a.client.SetGroupDescription(context.Background(), jid, description)
}

func (a *whatsAppClientAdapter) LeaveGroup(jid types.JID) error {
	if a.client == nil {
		return ErrClientNotConnected
	}
	return a.client.LeaveGroup(context.Background(), jid)
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
