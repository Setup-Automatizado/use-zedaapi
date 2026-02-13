package communities

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"log/slog"

	"github.com/google/uuid"

	whatsmeowclient "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/groups"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/types"
)

const (
	maxCommunityNameLength = 25
)

func ensureCommunityName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len([]rune(trimmed)) > maxCommunityNameLength {
		return "", ErrInvalidCommunityName
	}
	return trimmed, nil
}

func parseCommunityID(value string) (types.JID, error) {
	id := strings.TrimSpace(value)
	if id == "" {
		return types.EmptyJID, ErrInvalidCommunityID
	}
	jid, err := groups.ParseGroupID(id)
	if err != nil {
		return types.EmptyJID, ErrInvalidCommunityID
	}
	return jid, nil
}

func mapSubGroups(targets []*types.GroupLinkTarget) ([]SubGroup, string) {
	if len(targets) == 0 {
		return []SubGroup{}, ""
	}
	subGroups := make([]SubGroup, 0, len(targets))
	var announcementID string
	for _, target := range targets {
		if target == nil {
			continue
		}
		name := strings.TrimSpace(target.GroupName.Name)
		if name == "" {
			name = target.JID.User
		}
		if name == "" {
			name = target.JID.String()
		}
		phone := groups.FormatGroupID(target.JID)
		subGroups = append(subGroups, SubGroup{
			Name:                name,
			Phone:               phone,
			IsGroupAnnouncement: target.IsDefaultSubGroup,
		})
		if target.IsDefaultSubGroup && announcementID == "" {
			announcementID = phone
		}
	}
	return subGroups, announcementID
}

func communityIdentifier(jid types.JID) string {
	if jid.User != "" {
		return jid.User
	}
	return jid.String()
}

// Create provisions a new community on WhatsApp.
func (s *Service) Create(ctx context.Context, instanceID uuid.UUID, params CreateParams) (CreateResult, error) {
	name, err := ensureCommunityName(params.Name)
	if err != nil {
		return CreateResult{}, err
	}

	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "communities_service"),
		slog.String("operation", "create"),
		slog.String("instance_id", instanceID.String()),
		slog.String("name", name),
	)

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return CreateResult{}, err
		}
		logger.Error("failed to obtain whatsapp client",
			slog.String("error", err.Error()))
		return CreateResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	req := whatsmeowclient.ReqCreateGroup{
		Name: name,
	}
	req.IsParent = true
	if req.DefaultMembershipApprovalMode == "" {
		req.DefaultMembershipApprovalMode = "request_required"
	}

	info, err := client.CreateGroup(ctx, req)
	if err != nil {
		logger.Error("failed to create community",
			slog.String("error", err.Error()))
		return CreateResult{}, fmt.Errorf("create community: %w", err)
	}

	description := strings.TrimSpace(params.Description)
	if description != "" {
		if descErr := client.SetGroupTopic(ctx, info.JID, "", "", description); descErr != nil {
			logger.Warn("failed to set community description",
				slog.String("error", descErr.Error()))
		}
	}

	inviteLink, inviteErr := client.GetGroupInviteLink(info.JID, false)
	if inviteErr != nil {
		logger.Warn("failed to retrieve community invite link",
			slog.String("error", inviteErr.Error()))
	}

	rawSubGroups, subErr := client.GetSubGroups(info.JID)
	if subErr != nil {
		logger.Warn("failed to fetch community sub groups",
			slog.String("error", subErr.Error()))
	}
	subGroups, announcementGroup := mapSubGroups(rawSubGroups)
	if announcementGroup == "" {
		announcementGroup = groups.FormatGroupID(info.JID)
	}

	participantsCount := 0
	if members, countErr := client.GetLinkedGroupsParticipants(info.JID); countErr != nil {
		logger.Warn("failed to fetch linked groups participants",
			slog.String("error", countErr.Error()))
	} else {
		participantsCount = len(members)
	}

	var invitationPtr *string
	if inviteErr == nil && inviteLink != "" {
		invitationPtr = &inviteLink
	}

	result := CreateResult{
		ID:                  communityIdentifier(info.JID),
		InvitationLink:      invitationPtr,
		AnnouncementGroupID: announcementGroup,
		SubGroups:           subGroups,
	}

	logger.Info("community created successfully",
		slog.String("community_id", result.ID),
		slog.Int("subgroup_count", len(subGroups)),
		slog.Int("participants_count", participantsCount))

	return result, nil
}

// Link associates existing groups to the provided community.
func (s *Service) Link(ctx context.Context, instanceID uuid.UUID, params LinkParams) (OperationResult, error) {
	if len(params.GroupIDs) == 0 {
		return OperationResult{}, ErrInvalidGroupList
	}

	maskedCommunityID := params.CommunityID
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "communities_service"),
		slog.String("operation", "link"),
		slog.String("instance_id", instanceID.String()),
		slog.String("community_id", maskedCommunityID),
		slog.Int("groups_count", len(params.GroupIDs)),
	)

	communityJID, err := parseCommunityID(params.CommunityID)
	if err != nil {
		logger.Warn("invalid community id")
		return OperationResult{}, err
	}

	groupJIDs := make([]types.JID, 0, len(params.GroupIDs))
	for _, raw := range params.GroupIDs {
		groupJID, parseErr := groups.ParseGroupID(strings.TrimSpace(raw))
		if parseErr != nil {
			logger.Warn("invalid group id in link payload",
				slog.String("group_id", raw),
				slog.String("error", parseErr.Error()))
			return OperationResult{}, ErrInvalidGroupList
		}
		groupJIDs = append(groupJIDs, groupJID)
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return OperationResult{}, err
		}
		logger.Error("failed to obtain whatsapp client",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	for _, groupJID := range groupJIDs {
		if linkErr := client.LinkGroup(communityJID, groupJID); linkErr != nil {
			logger.Error("failed to link group",
				slog.String("group_jid", groupJID.String()),
				slog.String("error", linkErr.Error()))
			return OperationResult{}, fmt.Errorf("link group: %w", linkErr)
		}
	}

	logger.Info("groups linked successfully")
	return NewSuccessResult(true), nil
}

// Unlink removes the association between groups and the community.
func (s *Service) Unlink(ctx context.Context, instanceID uuid.UUID, params LinkParams) (OperationResult, error) {
	if len(params.GroupIDs) == 0 {
		return OperationResult{}, ErrInvalidGroupList
	}

	maskedCommunityID := params.CommunityID
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "communities_service"),
		slog.String("operation", "unlink"),
		slog.String("instance_id", instanceID.String()),
		slog.String("community_id", maskedCommunityID),
		slog.Int("groups_count", len(params.GroupIDs)),
	)

	communityJID, err := parseCommunityID(params.CommunityID)
	if err != nil {
		logger.Warn("invalid community id")
		return OperationResult{}, err
	}

	groupJIDs := make([]types.JID, 0, len(params.GroupIDs))
	for _, raw := range params.GroupIDs {
		groupJID, parseErr := groups.ParseGroupID(strings.TrimSpace(raw))
		if parseErr != nil {
			logger.Warn("invalid group id in unlink payload",
				slog.String("group_id", raw),
				slog.String("error", parseErr.Error()))
			return OperationResult{}, ErrInvalidGroupList
		}
		groupJIDs = append(groupJIDs, groupJID)
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return OperationResult{}, err
		}
		logger.Error("failed to obtain whatsapp client",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	for _, groupJID := range groupJIDs {
		if unlinkErr := client.UnlinkGroup(communityJID, groupJID); unlinkErr != nil {
			logger.Error("failed to unlink group",
				slog.String("group_jid", groupJID.String()),
				slog.String("error", unlinkErr.Error()))
			return OperationResult{}, fmt.Errorf("unlink group: %w", unlinkErr)
		}
	}

	logger.Info("groups unlinked successfully")
	return NewSuccessResult(true), nil
}

// Metadata aggregates detailed information about the provided community.
func (s *Service) Metadata(ctx context.Context, instanceID uuid.UUID, communityID string) (Metadata, error) {
	maskedCommunityID := communityID
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "communities_service"),
		slog.String("operation", "metadata"),
		slog.String("instance_id", instanceID.String()),
		slog.String("community_id", maskedCommunityID),
	)

	communityJID, err := parseCommunityID(communityID)
	if err != nil {
		logger.Warn("invalid community id")
		return Metadata{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return Metadata{}, err
		}
		logger.Error("failed to obtain whatsapp client",
			slog.String("error", err.Error()))
		return Metadata{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	info, err := client.GetGroupInfo(communityJID)
	if err != nil {
		logger.Error("failed to load community info",
			slog.String("error", err.Error()))
		return Metadata{}, fmt.Errorf("get community info: %w", err)
	}

	rawSubGroups, subErr := client.GetSubGroups(communityJID)
	if subErr != nil {
		logger.Warn("failed to fetch community sub groups",
			slog.String("error", subErr.Error()))
	}
	subGroups, announcementGroup := mapSubGroups(rawSubGroups)
	if announcementGroup == "" {
		announcementGroup = groups.FormatGroupID(communityJID)
	}

	participantsCount := 0
	if members, countErr := client.GetLinkedGroupsParticipants(communityJID); countErr != nil {
		logger.Warn("failed to fetch linked groups participants",
			slog.String("error", countErr.Error()))
	} else {
		participantsCount = len(members)
	}

	name := strings.TrimSpace(info.GroupName.Name)
	if name == "" {
		name = communityIdentifier(info.JID)
	}
	description := strings.TrimSpace(info.GroupTopic.Topic)
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}

	result := Metadata{
		ID:                  communityIdentifier(info.JID),
		Name:                name,
		Description:         descriptionPtr,
		AnnouncementGroupID: announcementGroup,
		ParticipantsCount:   participantsCount,
		SubGroups:           subGroups,
	}

	logger.Info("community metadata retrieved",
		slog.Int("subgroup_count", len(subGroups)),
		slog.Int("participants_count", participantsCount))

	return result, nil
}

// RegenerateInvitationLink resets the community invitation link.
func (s *Service) RegenerateInvitationLink(ctx context.Context, instanceID uuid.UUID, communityID string) (InvitationResult, error) {
	maskedCommunityID := communityID
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "communities_service"),
		slog.String("operation", "regenerate_invite"),
		slog.String("instance_id", instanceID.String()),
		slog.String("community_id", maskedCommunityID),
	)

	communityJID, err := parseCommunityID(communityID)
	if err != nil {
		logger.Warn("invalid community id")
		return InvitationResult{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return InvitationResult{}, err
		}
		logger.Error("failed to obtain whatsapp client",
			slog.String("error", err.Error()))
		return InvitationResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	link, linkErr := client.GetGroupInviteLink(communityJID, true)
	if linkErr != nil {
		logger.Error("failed to regenerate community invite link",
			slog.String("error", linkErr.Error()))
		return InvitationResult{}, fmt.Errorf("regenerate invite link: %w", linkErr)
	}

	logger.Info("community invite link regenerated successfully")
	return InvitationResult{InvitationLink: link}, nil
}

// UpdateSettings adjusts community-level preferences.
func (s *Service) UpdateSettings(ctx context.Context, instanceID uuid.UUID, params SettingsParams) (OperationResult, error) {
	maskedCommunityID := params.CommunityID
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "communities_service"),
		slog.String("operation", "update_settings"),
		slog.String("instance_id", instanceID.String()),
		slog.String("community_id", maskedCommunityID),
	)

	communityJID, err := parseCommunityID(params.CommunityID)
	if err != nil {
		logger.Warn("invalid community id")
		return OperationResult{}, err
	}

	mode, modeErr := mapAddGroupMode(params.WhoCanAddNewGroups)
	if modeErr != nil {
		logger.Warn("invalid community settings payload",
			slog.String("error", modeErr.Error()))
		return OperationResult{}, modeErr
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return OperationResult{}, err
		}
		logger.Error("failed to obtain whatsapp client",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	if setErr := client.SetGroupMemberAddMode(communityJID, mode); setErr != nil {
		logger.Error("failed to update community settings",
			slog.String("error", setErr.Error()))
		return OperationResult{}, fmt.Errorf("update community settings: %w", setErr)
	}

	logger.Info("community settings updated successfully",
		slog.String("who_can_add_new_groups", params.WhoCanAddNewGroups))
	return NewSuccessResult(true), nil
}

// UpdateDescription updates the community description text.
func (s *Service) UpdateDescription(ctx context.Context, instanceID uuid.UUID, params UpdateDescriptionParams) (OperationResult, error) {
	maskedCommunityID := params.CommunityID
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "communities_service"),
		slog.String("operation", "update_description"),
		slog.String("instance_id", instanceID.String()),
		slog.String("community_id", maskedCommunityID),
	)

	communityJID, err := parseCommunityID(params.CommunityID)
	if err != nil {
		logger.Warn("invalid community id")
		return OperationResult{}, err
	}

	description := strings.TrimSpace(params.Description)
	if description == "" {
		logger.Warn("invalid community description")
		return OperationResult{}, ErrInvalidCommunityDescription
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return OperationResult{}, err
		}
		logger.Error("failed to obtain whatsapp client",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	if err := client.SetGroupTopic(ctx, communityJID, "", "", description); err != nil {
		logger.Error("failed to update community description",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("update community description: %w", err)
	}

	logger.Info("community description updated")
	return NewValueResult(true), nil
}

func mapAddGroupMode(value string) (types.GroupMemberAddMode, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "admins", "admin", "only_admins":
		return types.GroupMemberAddModeAdmin, nil
	case "members", "all", "everyone", "all_members":
		return types.GroupMemberAddModeAllMember, nil
	default:
		return "", fmt.Errorf("unknown whoCanAddNewGroups value: %s", value)
	}
}

func convertGroupError(err error) error {
	switch {
	case errors.Is(err, groups.ErrInvalidPhoneList):
		return ErrInvalidPhoneList
	case errors.Is(err, groups.ErrInvalidGroupID):
		return ErrInvalidCommunityID
	case errors.Is(err, groups.ErrOperationFailed):
		return ErrOperationFailed
	default:
		return err
	}
}

func (s *Service) AddParticipants(ctx context.Context, instanceID uuid.UUID, params ParticipantsParams) (OperationResult, error) {
	if len(params.Phones) == 0 {
		return OperationResult{}, ErrInvalidPhoneList
	}
	announcementID, err := s.ResolveAnnouncementGroup(ctx, instanceID, params.CommunityID)
	if err != nil {
		return OperationResult{}, err
	}
	if s.groupsService == nil {
		return OperationResult{}, ErrOperationFailed
	}
	result, groupErr := s.groupsService.AddParticipants(ctx, instanceID, groups.ModifyParticipantsParams{
		GroupID:    announcementID,
		Phones:     params.Phones,
		AutoInvite: params.AutoInvite,
	})
	if groupErr != nil {
		return OperationResult{}, convertGroupError(groupErr)
	}
	return NewValueResult(result.Value), nil
}

func (s *Service) RemoveParticipants(ctx context.Context, instanceID uuid.UUID, params ParticipantsParams) (OperationResult, error) {
	if len(params.Phones) == 0 {
		return OperationResult{}, ErrInvalidPhoneList
	}
	announcementID, err := s.ResolveAnnouncementGroup(ctx, instanceID, params.CommunityID)
	if err != nil {
		return OperationResult{}, err
	}
	if s.groupsService == nil {
		return OperationResult{}, ErrOperationFailed
	}
	result, groupErr := s.groupsService.RemoveParticipants(ctx, instanceID, groups.ModifyParticipantsParams{
		GroupID: announcementID,
		Phones:  params.Phones,
	})
	if groupErr != nil {
		return OperationResult{}, convertGroupError(groupErr)
	}
	return NewValueResult(result.Value), nil
}

func (s *Service) AddAdmins(ctx context.Context, instanceID uuid.UUID, params ParticipantsParams) (OperationResult, error) {
	if len(params.Phones) == 0 {
		return OperationResult{}, ErrInvalidPhoneList
	}
	announcementID, err := s.ResolveAnnouncementGroup(ctx, instanceID, params.CommunityID)
	if err != nil {
		return OperationResult{}, err
	}
	if s.groupsService == nil {
		return OperationResult{}, ErrOperationFailed
	}
	result, groupErr := s.groupsService.AddAdmins(ctx, instanceID, groups.ModifyParticipantsParams{
		GroupID: announcementID,
		Phones:  params.Phones,
	})
	if groupErr != nil {
		return OperationResult{}, convertGroupError(groupErr)
	}
	return NewValueResult(result.Value), nil
}

func (s *Service) RemoveAdmins(ctx context.Context, instanceID uuid.UUID, params ParticipantsParams) (OperationResult, error) {
	if len(params.Phones) == 0 {
		return OperationResult{}, ErrInvalidPhoneList
	}
	announcementID, err := s.ResolveAnnouncementGroup(ctx, instanceID, params.CommunityID)
	if err != nil {
		return OperationResult{}, err
	}
	if s.groupsService == nil {
		return OperationResult{}, ErrOperationFailed
	}
	result, groupErr := s.groupsService.RemoveAdmins(ctx, instanceID, groups.ModifyParticipantsParams{
		GroupID: announcementID,
		Phones:  params.Phones,
	})
	if groupErr != nil {
		return OperationResult{}, convertGroupError(groupErr)
	}
	return NewValueResult(result.Value), nil
}

// ResolveAnnouncementGroup returns the group identifier of the community announcement group.
func (s *Service) ResolveAnnouncementGroup(ctx context.Context, instanceID uuid.UUID, communityID string) (string, error) {
	maskedCommunityID := communityID
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "communities_service"),
		slog.String("operation", "resolve_announcement_group"),
		slog.String("instance_id", instanceID.String()),
		slog.String("community_id", maskedCommunityID),
	)

	communityJID, err := parseCommunityID(communityID)
	if err != nil {
		logger.Warn("invalid community id")
		return "", err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return "", err
		}
		logger.Error("failed to obtain whatsapp client",
			slog.String("error", err.Error()))
		return "", fmt.Errorf("get whatsapp client: %w", err)
	}

	subGroups, subErr := client.GetSubGroups(communityJID)
	if subErr != nil {
		logger.Error("failed to fetch community sub groups",
			slog.String("error", subErr.Error()))
		return "", fmt.Errorf("get sub groups: %w", subErr)
	}

	_, announcementGroup := mapSubGroups(subGroups)
	if announcementGroup == "" {
		logger.Warn("announcement group not found, falling back to parent id")
		announcementGroup = groups.FormatGroupID(communityJID)
	}

	return announcementGroup, nil
}

// Delete removes the caller from the community and all linked subgroups.
func (s *Service) Delete(ctx context.Context, instanceID uuid.UUID, communityID string) (OperationResult, error) {
	maskedCommunityID := communityID
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "communities_service"),
		slog.String("operation", "delete"),
		slog.String("instance_id", instanceID.String()),
		slog.String("community_id", maskedCommunityID),
	)

	communityJID, err := parseCommunityID(communityID)
	if err != nil {
		logger.Warn("invalid community id")
		return OperationResult{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return OperationResult{}, err
		}
		logger.Error("failed to obtain whatsapp client",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	subGroups, subErr := client.GetSubGroups(communityJID)
	if subErr != nil {
		logger.Warn("failed to fetch community sub groups before delete",
			slog.String("error", subErr.Error()))
	}

	for _, subgroup := range subGroups {
		if subgroup == nil || subgroup.JID.IsEmpty() {
			continue
		}
		if leaveErr := client.LeaveGroup(subgroup.JID); leaveErr != nil {
			logger.Error("failed to leave subgroup",
				slog.String("group_jid", subgroup.JID.String()),
				slog.String("error", leaveErr.Error()))
			return OperationResult{}, fmt.Errorf("leave subgroup %s: %w", subgroup.JID.String(), leaveErr)
		}
	}

	if leaveErr := client.LeaveGroup(communityJID); leaveErr != nil {
		logger.Error("failed to leave community",
			slog.String("error", leaveErr.Error()))
		return OperationResult{}, fmt.Errorf("leave community: %w", leaveErr)
	}

	logger.Info("community deleted successfully")
	return NewValueResult(true), nil
}
