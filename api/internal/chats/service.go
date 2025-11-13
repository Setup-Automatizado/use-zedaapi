package chats

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	watypes "go.mau.fi/whatsmeow/types"
)

// Client defines the interface for interacting with WhatsApp chats.
// It abstracts the whatsmeow client operations needed for chat listing.
type Client interface {
	// GetAllContacts retrieves all stored contacts
	GetAllContacts(ctx context.Context) (map[watypes.JID]watypes.ContactInfo, error)

	// GetJoinedGroups retrieves all groups the user has joined
	GetJoinedGroups(ctx context.Context) ([]*watypes.GroupInfo, error)

	// GetChatSettings retrieves local settings for a specific chat
	GetChatSettings(ctx context.Context, chat watypes.JID) (watypes.LocalChatSettings, error)
}

// ClientProvider defines the interface for providing WhatsApp clients by instance ID.
type ClientProvider interface {
	Get(ctx context.Context, instanceID uuid.UUID) (Client, error)
}

// Service handles business logic for chat operations.
type Service struct {
	provider ClientProvider
	log      *slog.Logger
	now      func() time.Time
}

// NewService creates a new chats service.
func NewService(provider ClientProvider, log *slog.Logger) *Service {
	return &Service{
		provider: provider,
		log:      log,
		now:      time.Now,
	}
}

// List retrieves all chats (contacts and groups) with pagination.
// It combines contacts and groups, fetches settings for each, and returns a paginated result.
func (s *Service) List(ctx context.Context, instanceID uuid.UUID, params ListParams) (ListResult, error) {
	logger := s.log.With(
		slog.String("instance_id", instanceID.String()),
		slog.String("component", "chats_service"),
		slog.String("operation", "list"),
		slog.Int("page", params.Page),
		slog.Int("page_size", params.PageSize),
	)

	logger.InfoContext(ctx, "listing chats")

	// Get WhatsApp client
	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get whatsapp client",
			slog.String("error", err.Error()))
		return ListResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	// Fetch all contacts
	contactsMap, err := client.GetAllContacts(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get contacts from store",
			slog.String("error", err.Error()))
		return ListResult{}, fmt.Errorf("get all contacts: %w", err)
	}

	logger.DebugContext(ctx, "fetched contacts from store",
		slog.Int("count", len(contactsMap)))

	// Fetch all groups
	groupInfos, err := client.GetJoinedGroups(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to fetch joined groups",
			slog.String("error", err.Error()))
		return ListResult{}, fmt.Errorf("get joined groups: %w", err)
	}

	logger.DebugContext(ctx, "fetched joined groups",
		slog.Int("count", len(groupInfos)))

	// Combine contacts and groups into chats
	chats := make([]Chat, 0, len(contactsMap)+len(groupInfos))

	// Convert contacts to chats
	for jid, contactInfo := range contactsMap {
		// Fetch chat settings
		settings, err := client.GetChatSettings(ctx, jid)
		if err != nil {
			logger.WarnContext(ctx, "failed to get chat settings for contact, using defaults",
				slog.String("jid", jid.String()),
				slog.String("error", err.Error()))
			settings = watypes.LocalChatSettings{} // Use defaults
		}

		chat := fromContactInfo(jid, contactInfo, settings)
		chats = append(chats, chat)
	}

	// Convert groups to chats
	for _, groupInfo := range groupInfos {
		// Fetch chat settings
		settings, err := client.GetChatSettings(ctx, groupInfo.JID)
		if err != nil {
			logger.WarnContext(ctx, "failed to get chat settings for group, using defaults",
				slog.String("jid", groupInfo.JID.String()),
				slog.String("error", err.Error()))
			settings = watypes.LocalChatSettings{} // Use defaults
		}

		chat := fromGroupInfo(groupInfo, settings)
		chats = append(chats, chat)
	}

	totalCount := len(chats)
	logger.InfoContext(ctx, "combined chats",
		slog.Int("total_count", totalCount),
		slog.Int("contacts", len(contactsMap)),
		slog.Int("groups", len(groupInfos)))

	// Apply pagination
	start := (params.Page - 1) * params.PageSize
	if start > totalCount {
		start = totalCount
	}

	end := start + params.PageSize
	if end > totalCount {
		end = totalCount
	}

	paginatedChats := chats[start:end]

	logger.InfoContext(ctx, "paginated chats",
		slog.Int("returned", len(paginatedChats)),
		slog.Int("start", start),
		slog.Int("end", end))

	return ListResult{
		Chats:      paginatedChats,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}, nil
}
