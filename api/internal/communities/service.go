package communities

import (
	"context"
	"fmt"
	"sort"

	"log/slog"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/groups"
	"go.mau.fi/whatsmeow/api/internal/logging"
)

// GroupMutations abstracts the groups service operations communities rely on.
type GroupMutations interface {
	AddParticipants(ctx context.Context, instanceID uuid.UUID, params groups.ModifyParticipantsParams) (groups.ValueResult, error)
	RemoveParticipants(ctx context.Context, instanceID uuid.UUID, params groups.ModifyParticipantsParams) (groups.ValueResult, error)
	AddAdmins(ctx context.Context, instanceID uuid.UUID, params groups.ModifyParticipantsParams) (groups.ValueResult, error)
	RemoveAdmins(ctx context.Context, instanceID uuid.UUID, params groups.ModifyParticipantsParams) (groups.ValueResult, error)
}

// Service orchestrates the community listing flow on top of a WhatsApp client.
type Service struct {
	provider      ClientProvider
	groupsService GroupMutations
	log           *slog.Logger
}

// NewService builds a community service with the supplied dependencies.
func NewService(provider ClientProvider, groupsService GroupMutations, log *slog.Logger) *Service {
	return &Service{
		provider:      provider,
		groupsService: groupsService,
		log:           log,
	}
}

// List returns a paginated slice of communities for a given instance.
func (s *Service) List(ctx context.Context, instanceID uuid.UUID, params ListParams) (ListResult, error) {
	if params.Page <= 0 || params.PageSize <= 0 {
		return ListResult{}, ErrInvalidPagination
	}

	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "communities_service"),
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
	seen := make(map[string]struct{})
	for _, info := range groupInfos {
		if info == nil || !info.IsParent {
			continue
		}
		id := info.JID.User
		if id == "" {
			id = info.JID.String()
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}

		name := info.GroupName.Name
		if name == "" {
			name = id
		}
		summaries = append(summaries, Summary{
			ID:   id,
			Name: name,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Name == summaries[j].Name {
			return summaries[i].ID < summaries[j].ID
		}
		return summaries[i].Name < summaries[j].Name
	})

	total := len(summaries)
	start := (params.Page - 1) * params.PageSize
	if start >= total {
		logger.Info("list request beyond total communities",
			slog.Int("total_communities", total))
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

	logger.Info("communities listed successfully",
		slog.Int("total_communities", total),
		slog.Int("returned_communities", len(result.Items)))

	return result, nil
}
