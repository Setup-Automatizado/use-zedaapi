package groups

import (
	"context"
	"fmt"

	"log/slog"

	"github.com/google/uuid"

	whatsmeowclient "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
	internalwhatsmeow "go.mau.fi/whatsmeow/api/internal/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
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

// Get returns a groups client for the given instance ID, ensuring a WhatsApp client is available.
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
		slog.String("component", "groups_client_provider"),
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

func (a *whatsAppClientAdapter) GetChatSettings(ctx context.Context, chat types.JID) (types.LocalChatSettings, error) {
	if a.client == nil || a.client.Store == nil || a.client.Store.ChatSettings == nil {
		return types.LocalChatSettings{}, nil
	}
	return a.client.Store.ChatSettings.GetChatSettings(ctx, chat)
}

func (a *whatsAppClientAdapter) CreateGroup(ctx context.Context, req whatsmeowclient.ReqCreateGroup) (*types.GroupInfo, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.CreateGroup(ctx, req)
}

func (a *whatsAppClientAdapter) SetGroupName(jid types.JID, name string) error {
	if a.client == nil {
		return ErrClientNotConnected
	}
	return a.client.SetGroupName(context.Background(), jid, name)
}

func (a *whatsAppClientAdapter) SetGroupPhoto(jid types.JID, avatar []byte) (string, error) {
	if a.client == nil {
		return "", ErrClientNotConnected
	}
	return a.client.SetGroupPhoto(context.Background(), jid, avatar)
}

func (a *whatsAppClientAdapter) UpdateGroupParticipants(jid types.JID, participantChanges []types.JID, action whatsmeowclient.ParticipantChange) ([]types.GroupParticipant, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.UpdateGroupParticipants(context.Background(), jid, participantChanges, action)
}

func (a *whatsAppClientAdapter) UpdateGroupRequestParticipants(jid types.JID, participantChanges []types.JID, action whatsmeowclient.ParticipantRequestChange) ([]types.GroupParticipant, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.UpdateGroupRequestParticipants(context.Background(), jid, participantChanges, action)
}

func (a *whatsAppClientAdapter) LeaveGroup(jid types.JID) error {
	if a.client == nil {
		return ErrClientNotConnected
	}
	return a.client.LeaveGroup(context.Background(), jid)
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

func (a *whatsAppClientAdapter) GetGroupInfoFromLink(code string) (*types.GroupInfo, error) {
	if a.client == nil {
		return nil, ErrClientNotConnected
	}
	return a.client.GetGroupInfoFromLink(context.Background(), code)
}

func (a *whatsAppClientAdapter) JoinGroupWithLink(code string) (types.JID, error) {
	if a.client == nil {
		return types.EmptyJID, ErrClientNotConnected
	}
	return a.client.JoinGroupWithLink(context.Background(), code)
}

func (a *whatsAppClientAdapter) SetGroupAnnounce(jid types.JID, announce bool) error {
	if a.client == nil {
		return ErrClientNotConnected
	}
	return a.client.SetGroupAnnounce(context.Background(), jid, announce)
}

func (a *whatsAppClientAdapter) SetGroupLocked(jid types.JID, locked bool) error {
	if a.client == nil {
		return ErrClientNotConnected
	}
	return a.client.SetGroupLocked(context.Background(), jid, locked)
}

func (a *whatsAppClientAdapter) SetGroupJoinApprovalMode(jid types.JID, mode bool) error {
	if a.client == nil {
		return ErrClientNotConnected
	}
	return a.client.SetGroupJoinApprovalMode(context.Background(), jid, mode)
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

func (a *whatsAppClientAdapter) SendMessage(ctx context.Context, to types.JID, message *waProto.Message, extra ...whatsmeowclient.SendRequestExtra) (whatsmeowclient.SendResponse, error) {
	if a.client == nil {
		return whatsmeowclient.SendResponse{}, ErrClientNotConnected
	}
	return a.client.SendMessage(ctx, to, message, extra...)
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
