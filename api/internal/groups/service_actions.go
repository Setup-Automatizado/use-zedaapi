package groups

import (
	"context"
	"fmt"
	"strings"

	"log/slog"

	"github.com/google/uuid"

	whatsmeow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/types"
)

// Create provisions a new group on WhatsApp and returns the Z-API compliant identifiers.
func (s *Service) Create(ctx context.Context, instanceID uuid.UUID, params CreateParams) (CreateResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", "create"),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_name", params.GroupName),
	)

	if err := ensureGroupName(params.GroupName); err != nil {
		logger.Warn("invalid group name", slog.String("error", err.Error()))
		return CreateResult{}, err
	}

	participantJIDs, err := PhonesToJIDs(params.Phones)
	if err != nil {
		logger.Warn("invalid participant phones", slog.String("error", err.Error()))
		return CreateResult{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return CreateResult{}, err
		}
		logger.Error("failed to obtain whatsapp client", slog.String("error", err.Error()))
		return CreateResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	info, err := client.CreateGroup(ctx, whatsmeow.ReqCreateGroup{
		Name:         params.GroupName,
		Participants: participantJIDs,
	})
	if err != nil {
		logger.Error("failed to create group", slog.String("error", err.Error()))
		return CreateResult{}, fmt.Errorf("create group: %w", err)
	}

	groupJID := info.JID
	inviteLink, inviteErr := client.GetGroupInviteLink(groupJID, false)
	if inviteErr != nil {
		logger.Warn("failed to get invite link post creation", slog.String("error", inviteErr.Error()))
	}

	if params.AutoInvite {
		failures := failedParticipants(participantJIDs, info.Participants)
		if len(failures) > 0 && inviteErr == nil {
			logger.Info("sending auto-invites for failed participants",
				slog.Int("failed_participants", len(failures)))
			sendAutoInvites(ctx, client, logger, failures, inviteLink)
		}
	}

	result := CreateResult{
		Phone:          FormatGroupID(groupJID),
		InvitationLink: inviteLink,
	}
	logger.Info("group created successfully",
		slog.String("phone", result.Phone),
		slog.Bool("auto_invite", params.AutoInvite),
		slog.Bool("invitation_available", inviteLink != ""))

	return result, nil
}

// UpdateName renames an existing group.
func (s *Service) UpdateName(ctx context.Context, instanceID uuid.UUID, params UpdateNameParams) (ValueResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", "update_name"),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", params.GroupID),
	)

	if err := ensureGroupName(params.GroupName); err != nil {
		logger.Warn("invalid group name", slog.String("error", err.Error()))
		return ValueResult{}, err
	}

	groupJID, err := ParseGroupID(params.GroupID)
	if err != nil {
		logger.Warn("invalid group id", slog.String("error", err.Error()))
		return ValueResult{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return ValueResult{}, err
		}
		logger.Error("failed to obtain whatsapp client", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	if err := client.SetGroupName(groupJID, params.GroupName); err != nil {
		logger.Error("failed to set group name", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("set group name: %w", err)
	}

	logger.Info("group name updated successfully")
	return ValueResult{Value: true}, nil
}

// UpdatePhoto sets a new photo for the group.
func (s *Service) UpdatePhoto(ctx context.Context, instanceID uuid.UUID, params UpdatePhotoParams) (ValueResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", "update_photo"),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", params.GroupID),
	)

	groupJID, err := ParseGroupID(params.GroupID)
	if err != nil {
		logger.Warn("invalid group id", slog.String("error", err.Error()))
		return ValueResult{}, err
	}

	imageBytes, err := imageBytesFromInput(ctx, params.GroupPhoto)
	if err != nil {
		logger.Warn("invalid group photo input", slog.String("error", err.Error()))
		return ValueResult{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return ValueResult{}, err
		}
		logger.Error("failed to obtain whatsapp client", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	if _, err := client.SetGroupPhoto(groupJID, imageBytes); err != nil {
		logger.Error("failed to set group photo", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("set group photo: %w", err)
	}

	logger.Info("group photo updated successfully")
	return ValueResult{Value: true}, nil
}

func (s *Service) modifyParticipants(ctx context.Context, instanceID uuid.UUID, params ModifyParticipantsParams, action whatsmeow.ParticipantChange, operation string) (ValueResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", operation),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", params.GroupID),
	)

	groupJID, err := ParseGroupID(params.GroupID)
	if err != nil {
		logger.Warn("invalid group id", slog.String("error", err.Error()))
		return ValueResult{}, err
	}

	participantJIDs, err := PhonesToJIDs(params.Phones)
	if err != nil {
		logger.Warn("invalid participant phones", slog.String("error", err.Error()))
		return ValueResult{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return ValueResult{}, err
		}
		logger.Error("failed to obtain whatsapp client", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	result, err := client.UpdateGroupParticipants(groupJID, participantJIDs, action)
	if err != nil {
		logger.Error("failed to modify participants", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("update participants: %w", err)
	}

	if params.AutoInvite && action == whatsmeow.ParticipantChangeAdd {
		inviteLink, inviteErr := client.GetGroupInviteLink(groupJID, false)
		if inviteErr == nil && inviteLink != "" {
			failures := failedParticipants(participantJIDs, result)
			if len(failures) > 0 {
				logger.Info("sending auto-invites after add participants",
					slog.Int("failed_participants", len(failures)))
				sendAutoInvites(ctx, client, logger, failures, inviteLink)
			}
		} else if inviteErr != nil {
			logger.Warn("failed to get invite link after adding participants",
				slog.String("error", inviteErr.Error()))
		}
	}

	logger.Info("participants modification completed",
		slog.Int("requested", len(participantJIDs)))
	return ValueResult{Value: true}, nil
}

// AddParticipants adds new members to the group.
func (s *Service) AddParticipants(ctx context.Context, instanceID uuid.UUID, params ModifyParticipantsParams) (ValueResult, error) {
	return s.modifyParticipants(ctx, instanceID, params, whatsmeow.ParticipantChangeAdd, "add_participants")
}

// RemoveParticipants removes members from the group.
func (s *Service) RemoveParticipants(ctx context.Context, instanceID uuid.UUID, params ModifyParticipantsParams) (ValueResult, error) {
	return s.modifyParticipants(ctx, instanceID, params, whatsmeow.ParticipantChangeRemove, "remove_participants")
}

// AddAdmins promotes group participants to admin role.
func (s *Service) AddAdmins(ctx context.Context, instanceID uuid.UUID, params ModifyParticipantsParams) (ValueResult, error) {
	return s.modifyParticipants(ctx, instanceID, params, whatsmeow.ParticipantChangePromote, "add_admin")
}

// RemoveAdmins demotes group admins back to regular participants.
func (s *Service) RemoveAdmins(ctx context.Context, instanceID uuid.UUID, params ModifyParticipantsParams) (ValueResult, error) {
	return s.modifyParticipants(ctx, instanceID, params, whatsmeow.ParticipantChangeDemote, "remove_admin")
}

func (s *Service) updateRequestParticipants(ctx context.Context, instanceID uuid.UUID, params ModifyParticipantsParams, action whatsmeow.ParticipantRequestChange, operation string) (ValueResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", operation),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", params.GroupID),
	)

	groupJID, err := ParseGroupID(params.GroupID)
	if err != nil {
		logger.Warn("invalid group id", slog.String("error", err.Error()))
		return ValueResult{}, err
	}

	participantJIDs, err := PhonesToJIDs(params.Phones)
	if err != nil {
		logger.Warn("invalid participant phones", slog.String("error", err.Error()))
		return ValueResult{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return ValueResult{}, err
		}
		logger.Error("failed to obtain whatsapp client", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	if _, err := client.UpdateGroupRequestParticipants(groupJID, participantJIDs, action); err != nil {
		logger.Error("failed to update request participants", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("update request participants: %w", err)
	}

	logger.Info("group request participants updated", slog.Int("processed", len(participantJIDs)))
	return ValueResult{Value: true}, nil
}

// ApproveParticipants approves pending participant requests.
func (s *Service) ApproveParticipants(ctx context.Context, instanceID uuid.UUID, params ModifyParticipantsParams) (ValueResult, error) {
	return s.updateRequestParticipants(ctx, instanceID, params, whatsmeow.ParticipantChangeApprove, "approve_participants")
}

// RejectParticipants rejects pending participant requests.
func (s *Service) RejectParticipants(ctx context.Context, instanceID uuid.UUID, params ModifyParticipantsParams) (ValueResult, error) {
	return s.updateRequestParticipants(ctx, instanceID, params, whatsmeow.ParticipantChangeReject, "reject_participants")
}

// Leave causes the current instance to leave the group.
func (s *Service) Leave(ctx context.Context, instanceID uuid.UUID, params ModifyParticipantsParams) (ValueResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", "leave_group"),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", params.GroupID),
	)

	groupJID, err := ParseGroupID(params.GroupID)
	if err != nil {
		logger.Warn("invalid group id", slog.String("error", err.Error()))
		return ValueResult{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return ValueResult{}, err
		}
		logger.Error("failed to obtain whatsapp client", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	if err := client.LeaveGroup(groupJID); err != nil {
		logger.Error("failed to leave group", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("leave group: %w", err)
	}

	logger.Info("left group successfully")
	return ValueResult{Value: true}, nil
}

// Metadata returns the full group metadata including invite link.
func (s *Service) Metadata(ctx context.Context, instanceID uuid.UUID, groupID string) (Metadata, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", "metadata"),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", groupID),
	)

	groupJID, err := ParseGroupID(groupID)
	if err != nil {
		logger.Warn("invalid group id", slog.String("error", err.Error()))
		return Metadata{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return Metadata{}, err
		}
		logger.Error("failed to obtain whatsapp client", slog.String("error", err.Error()))
		return Metadata{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	info, err := client.GetGroupInfo(groupJID)
	if err != nil {
		logger.Error("failed to fetch group info", slog.String("error", err.Error()))
		return Metadata{}, fmt.Errorf("get group info: %w", err)
	}

	inviteLink, inviteErr := client.GetGroupInviteLink(groupJID, false)
	if inviteErr != nil {
		logger.Warn("failed to fetch invite link", slog.String("error", inviteErr.Error()))
	}

	meta := metadataFromGroup(info, true, inviteLink, inviteErr)
	logger.Info("group metadata retrieved", slog.Int("participants", len(meta.Participants)))
	return meta, nil
}

// LightMetadata returns the group metadata without attempting to fetch the invite link.
func (s *Service) LightMetadata(ctx context.Context, instanceID uuid.UUID, groupID string) (Metadata, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", "light_metadata"),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", groupID),
	)

	groupJID, err := ParseGroupID(groupID)
	if err != nil {
		logger.Warn("invalid group id", slog.String("error", err.Error()))
		return Metadata{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return Metadata{}, err
		}
		logger.Error("failed to obtain whatsapp client", slog.String("error", err.Error()))
		return Metadata{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	info, err := client.GetGroupInfo(groupJID)
	if err != nil {
		logger.Error("failed to fetch group info", slog.String("error", err.Error()))
		return Metadata{}, fmt.Errorf("get group info: %w", err)
	}

	meta := metadataFromGroup(info, false, "", nil)
	logger.Info("light group metadata retrieved", slog.Int("participants", len(meta.Participants)))
	return meta, nil
}

// InvitationMetadata resolves metadata for an invite link.
func (s *Service) InvitationMetadata(ctx context.Context, instanceID uuid.UUID, inviteURL string) (InvitationMetadata, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", "invitation_metadata"),
		slog.String("instance_id", instanceID.String()),
	)

	code, err := normalizeInviteURL(inviteURL)
	if err != nil {
		logger.Warn("invalid invite url", slog.String("error", err.Error()))
		return InvitationMetadata{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return InvitationMetadata{}, err
		}
		logger.Error("failed to obtain whatsapp client", slog.String("error", err.Error()))
		return InvitationMetadata{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	info, err := client.GetGroupInfoFromLink(code)
	if err != nil {
		logger.Error("failed to fetch invite metadata", slog.String("error", err.Error()))
		return InvitationMetadata{}, fmt.Errorf("get group info from link: %w", err)
	}

	meta := invitationMetadataFromGroup(info, whatsmeow.InviteLinkPrefix+code)
	logger.Info("group invitation metadata retrieved", slog.Int("participants", len(meta.Participants)))
	return meta, nil
}

// InvitationLink retrieves the current invite link for a group.
func (s *Service) InvitationLink(ctx context.Context, instanceID uuid.UUID, groupID string) (InvitationLinkResult, error) {
	return s.obtainInvite(ctx, instanceID, groupID, false)
}

// RedefineInvitationLink regenerates the invite link for a group.
func (s *Service) RedefineInvitationLink(ctx context.Context, instanceID uuid.UUID, groupID string) (InvitationLinkResult, error) {
	return s.obtainInvite(ctx, instanceID, groupID, true)
}

func (s *Service) obtainInvite(ctx context.Context, instanceID uuid.UUID, groupID string, reset bool) (InvitationLinkResult, error) {
	operation := "get_invite"
	if reset {
		operation = "reset_invite"
	}
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", operation),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", groupID),
	)

	groupJID, err := ParseGroupID(groupID)
	if err != nil {
		logger.Warn("invalid group id", slog.String("error", err.Error()))
		return InvitationLinkResult{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return InvitationLinkResult{}, err
		}
		logger.Error("failed to obtain whatsapp client", slog.String("error", err.Error()))
		return InvitationLinkResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	link, err := client.GetGroupInviteLink(groupJID, reset)
	if err != nil {
		logger.Error("failed to retrieve invite link", slog.String("error", err.Error()))
		return InvitationLinkResult{}, fmt.Errorf("get invite link: %w", err)
	}

	result := InvitationLinkResult{
		Phone:          FormatGroupID(groupJID),
		InvitationLink: link,
	}
	logger.Info("invite link processed successfully")
	return result, nil
}

// UpdateSettings toggles the WhatsApp group settings.
func (s *Service) UpdateSettings(ctx context.Context, instanceID uuid.UUID, params UpdateSettingsParams) (ValueResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", "update_settings"),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", params.Phone),
	)

	groupJID, err := ParseGroupID(params.Phone)
	if err != nil {
		logger.Warn("invalid group id", slog.String("error", err.Error()))
		return ValueResult{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return ValueResult{}, err
		}
		logger.Error("failed to obtain whatsapp client", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	if err := client.SetGroupAnnounce(groupJID, params.AdminOnlyMessage); err != nil {
		logger.Error("failed to update announce mode", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("set group announce: %w", err)
	}
	if err := client.SetGroupLocked(groupJID, params.AdminOnlySettings); err != nil {
		logger.Error("failed to update locked mode", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("set group locked: %w", err)
	}
	if err := client.SetGroupJoinApprovalMode(groupJID, params.RequireAdminApproval); err != nil {
		logger.Error("failed to update approval mode", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("set group join approval: %w", err)
	}
	mode := types.GroupMemberAddModeAllMember
	if params.AdminOnlyAddMember {
		mode = types.GroupMemberAddModeAdmin
	}
	if err := client.SetGroupMemberAddMode(groupJID, mode); err != nil {
		logger.Error("failed to update member add mode", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("set group member add mode: %w", err)
	}

	logger.Info("group settings updated successfully")
	return ValueResult{Value: true}, nil
}

// UpdateDescription updates the textual description of a group.
func (s *Service) UpdateDescription(ctx context.Context, instanceID uuid.UUID, params UpdateDescriptionParams) (ValueResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", "update_description"),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", params.GroupID),
	)

	groupJID, err := ParseGroupID(params.GroupID)
	if err != nil {
		logger.Warn("invalid group id", slog.String("error", err.Error()))
		return ValueResult{}, err
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return ValueResult{}, err
		}
		logger.Error("failed to obtain whatsapp client", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	if err := client.SetGroupDescription(groupJID, params.GroupDescription); err != nil {
		logger.Error("failed to set group description", slog.String("error", err.Error()))
		return ValueResult{}, fmt.Errorf("set group description: %w", err)
	}

	logger.Info("group description updated successfully")
	return ValueResult{Value: true}, nil
}

// AcceptInvite joins a group using an invite link.
func (s *Service) AcceptInvite(ctx context.Context, instanceID uuid.UUID, inviteURL string) (AcceptInviteResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "groups_service"),
		slog.String("operation", "accept_invite"),
		slog.String("instance_id", instanceID.String()),
	)

	if strings.TrimSpace(inviteURL) == "" {
		logger.Warn("empty invite url provided")
		return AcceptInviteResult{}, ErrInvalidInviteURL
	}

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return AcceptInviteResult{}, err
		}
		logger.Error("failed to obtain whatsapp client", slog.String("error", err.Error()))
		return AcceptInviteResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	if _, err := client.JoinGroupWithLink(inviteURL); err != nil {
		logger.Error("failed to join group with invite", slog.String("error", err.Error()))
		return AcceptInviteResult{}, fmt.Errorf("join group with link: %w", err)
	}

	logger.Info("invite accepted successfully")
	return AcceptInviteResult{Success: true}, nil
}
