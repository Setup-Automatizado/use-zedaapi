package groups

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/google/uuid"

	whatsmeow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/logging"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
)

// Client exposes the WhatsApp operations required by the groups service.
type Client interface {
	GetJoinedGroups(ctx context.Context) ([]*types.GroupInfo, error)
	GetChatSettings(ctx context.Context, chat types.JID) (types.LocalChatSettings, error)
	CreateGroup(ctx context.Context, req whatsmeow.ReqCreateGroup) (*types.GroupInfo, error)
	SetGroupName(jid types.JID, name string) error
	SetGroupPhoto(jid types.JID, avatar []byte) (string, error)
	UpdateGroupParticipants(jid types.JID, participantChanges []types.JID, action whatsmeow.ParticipantChange) ([]types.GroupParticipant, error)
	UpdateGroupRequestParticipants(jid types.JID, participantChanges []types.JID, action whatsmeow.ParticipantRequestChange) ([]types.GroupParticipant, error)
	LeaveGroup(jid types.JID) error
	GetGroupInfo(jid types.JID) (*types.GroupInfo, error)
	GetGroupInfoWithContext(ctx context.Context, jid types.JID) (*types.GroupInfo, error)
	GetGroupInviteLink(jid types.JID, reset bool) (string, error)
	GetGroupInfoFromLink(code string) (*types.GroupInfo, error)
	JoinGroupWithLink(code string) (types.JID, error)
	SetGroupAnnounce(jid types.JID, announce bool) error
	SetGroupLocked(jid types.JID, locked bool) error
	SetGroupJoinApprovalMode(jid types.JID, mode bool) error
	SetGroupMemberAddMode(jid types.JID, mode types.GroupMemberAddMode) error
	SetGroupTopic(ctx context.Context, jid types.JID, previousID, newID, topic string) error
	SendMessage(ctx context.Context, to types.JID, message *waProto.Message, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error)
}

// ClientProvider resolves a WhatsApp client for a given instance.
type ClientProvider interface {
	Get(ctx context.Context, instanceID uuid.UUID) (Client, error)
}

// Service orchestrates the group listing flow using a WhatsApp client.
type Service struct {
	provider ClientProvider
	log      *slog.Logger
	now      func() time.Time
}

// NewService builds a groups service with the given dependencies.
func NewService(provider ClientProvider, log *slog.Logger) *Service {
	return &Service{
		provider: provider,
		log:      log,
		now:      time.Now,
	}
}

// List returns a paginated list of groups for the given instance.
func (s *Service) List(ctx context.Context, instanceID uuid.UUID, params ListParams) (ListResult, error) {
	if params.Page <= 0 || params.PageSize <= 0 {
		return ListResult{}, ErrInvalidPagination
	}

	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", "list"),
		slog.String("instance_id", instanceID.String()),
		slog.Int("page", params.Page),
		slog.Int("page_size", params.PageSize),
	)

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return ListResult{}, err
		}
		logger.Error("failed to obtain whatsapp client",
			slog.String("error", err.Error()))
		return ListResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	groupInfos, err := client.GetJoinedGroups(ctx)
	if err != nil {
		logger.Error("failed to fetch joined groups",
			slog.String("error", err.Error()))
		return ListResult{}, fmt.Errorf("get joined groups: %w", err)
	}

	summaries := make([]Summary, 0, len(groupInfos))
	now := s.now()
	for _, info := range groupInfos {
		if info == nil {
			continue
		}
		settings, settingsErr := client.GetChatSettings(ctx, info.JID)
		if settingsErr != nil {
			logger.Warn("chat settings unavailable",
				slog.String("group_jid", info.JID.String()),
				slog.String("error", settingsErr.Error()))
		}
		summary := toSummary(info, settings, now)
		summaries = append(summaries, summary)
	}

	total := len(summaries)
	start := (params.Page - 1) * params.PageSize
	if start >= total {
		logger.Info("list request beyond total groups",
			slog.Int("total_groups", total))
		return ListResult{Items: []Summary{}, Total: total}, nil
	}
	end := start + params.PageSize
	if end > total {
		end = total
	}

	result := ListResult{
		Items: append([]Summary(nil), summaries[start:end]...),
		Total: total,
	}

	logger.Info("groups listed successfully",
		slog.Int("total_groups", total),
		slog.Int("returned_groups", len(result.Items)))

	return result, nil
}

func toSummary(info *types.GroupInfo, settings types.LocalChatSettings, now time.Time) Summary {
	// Pending store support for conversation activity timestamps; default to creation time for now.
	lastMessage := formatTimestamp(info.GroupCreated)
	isMuted, muteEnd := evaluateMute(settings, now)

	return Summary{
		IsGroup:         true,
		Name:            info.GroupName.Name,
		Phone:           conversationIdentifierFromJID(info.JID),
		Unread:          "0",
		LastMessageTime: lastMessage,
		IsMuted:         isMuted,
		MuteEndTime:     muteEnd,
		IsMarkedSpam:    false,
		Archived:        settings.Archived,
		Pinned:          settings.Pinned,
		MessagesUnread:  "0",
	}
}

func formatTimestamp(t time.Time) *string {
	if t.IsZero() {
		return nil
	}
	ms := t.UTC().UnixMilli()
	value := strconv.FormatInt(ms, 10)
	return &value
}

func evaluateMute(settings types.LocalChatSettings, now time.Time) (string, *string) {
	if !settings.Found || settings.MutedUntil.IsZero() {
		return "0", nil
	}

	if settings.MutedUntil.Equal(store.MutedForever) {
		value := "-1"
		return "1", &value
	}

	if settings.MutedUntil.After(now) {
		value := strconv.FormatInt(settings.MutedUntil.UTC().UnixMilli(), 10)
		return "1", &value
	}

	return "0", nil
}

func conversationIdentifierFromJID(jid types.JID) string {
	user := sanitizeUserComponent(jid.User)
	if user == "" {
		user = jid.User
	}

	switch jid.Server {
	case types.GroupServer:
		return user + "-group"
	case types.BroadcastServer:
		if user == "" {
			return jid.User + "-broadcast"
		}
		return user + "-broadcast"
	case types.NewsletterServer:
		return user + "-channel"
	case types.HiddenUserServer:
		if user == "" {
			return jid.User + "@" + types.HiddenUserServer
		}
		return user + "@" + types.HiddenUserServer
	default:
		return user
	}
}

func sanitizeUserComponent(user string) string {
	if idx := strings.IndexRune(user, ':'); idx >= 0 {
		user = user[:idx]
	}
	if idx := strings.IndexRune(user, '.'); idx >= 0 {
		user = user[:idx]
	}
	return user
}
