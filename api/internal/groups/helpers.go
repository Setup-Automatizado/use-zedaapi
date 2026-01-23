package groups

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

const (
	groupIDSuffix = "-group"
	// TODO: export defaultInviteMessage para .env
	defaultInviteMessage = "Has sido invitado a unirte al grupo: %s"
	maxGroupNameLength   = 25
	maxPhotoBytes        = 1 << 20 // 1 MiB safety limit
)

func ParseGroupID(id string) (types.JID, error) {
	if id == "" {
		return types.EmptyJID, ErrInvalidGroupID
	}

	id = strings.TrimSpace(id)

	// 120363420897515281@g.us
	if strings.HasSuffix(id, "@g.us") {
		return types.ParseJID(id)
	}

	// 120363420897515281-group
	if strings.HasSuffix(id, groupIDSuffix) {
		raw := strings.TrimSuffix(id, groupIDSuffix)
		if raw == "" {
			return types.EmptyJID, ErrInvalidGroupID
		}
		return types.ParseJID(raw + "@g.us")
	}

	return types.EmptyJID, ErrInvalidGroupID
}

func FormatGroupID(jid types.JID) string {
	if jid.User == "" {
		return ""
	}
	return jid.User + groupIDSuffix
}

func PhoneToUserJID(phone string) (types.JID, error) {
	if phone == "" {
		return types.EmptyJID, fmt.Errorf("%w: empty phone", ErrInvalidPhoneList)
	}
	if strings.Contains(phone, "@") {
		return types.ParseJID(phone)
	}
	return types.ParseJID(phone + "@s.whatsapp.net")
}

func PhonesToJIDs(phones []string) ([]types.JID, error) {
	if len(phones) == 0 {
		return nil, ErrInvalidPhoneList
	}
	jids := make([]types.JID, 0, len(phones))
	for _, p := range phones {
		jid, err := PhoneToUserJID(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("parse phone %q: %w", p, err)
		}
		jids = append(jids, jid)
	}
	return jids, nil
}

func participantSummaries(participants []types.GroupParticipant) []ParticipantSummary {
	summaries := make([]ParticipantSummary, 0, len(participants))
	for _, part := range participants {
		phone := part.PhoneNumber.User
		if phone == "" {
			phone = part.JID.User
		}
		summary := ParticipantSummary{
			Phone:        phone,
			IsAdmin:      part.IsAdmin,
			IsSuperAdmin: part.IsSuperAdmin,
		}
		if !part.LID.IsEmpty() {
			summary.LID = part.LID.String()
		}
		if part.DisplayName != "" {
			name := part.DisplayName
			summary.Name = &name
		}
		summaries = append(summaries, summary)
	}
	return summaries
}

func selectCommunityID(info *types.GroupInfo) *string {
	if info == nil || info.LinkedParentJID.IsEmpty() {
		return nil
	}
	value := info.LinkedParentJID.User
	if value == "" {
		value = info.LinkedParentJID.String()
	}
	return &value
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func metadataFromGroup(info *types.GroupInfo, includeInvite bool, inviteLink string, inviteErr error) Metadata {
	var invitationLink *string
	var invitationLinkErr *string
	if includeInvite {
		if inviteErr == nil && inviteLink != "" {
			invitationLink = &inviteLink
		} else if inviteErr != nil {
			msg := inviteErr.Error()
			invitationLinkErr = &msg
		}
	}

	description := info.GroupTopic.Topic
	owner := info.OwnerPN.User
	if owner == "" {
		owner = info.OwnerJID.User
	}
	subjectOwner := info.GroupName.NameSetBy.User
	subjectTime := int64(0)
	if !info.GroupName.NameSetAt.IsZero() {
		subjectTime = info.GroupName.NameSetAt.UTC().UnixMilli()
	}

	return Metadata{
		Phone:                FormatGroupID(info.JID),
		Description:          description,
		Owner:                owner,
		Subject:              info.GroupName.Name,
		Creation:             info.GroupCreated.UTC().UnixMilli(),
		InvitationLink:       invitationLink,
		InvitationLinkError:  invitationLinkErr,
		CommunityID:          selectCommunityID(info),
		AdminOnlyMessage:     info.IsAnnounce,
		AdminOnlySettings:    info.IsLocked,
		RequireAdminApproval: info.IsJoinApprovalRequired,
		IsGroupAnnouncement:  info.IsAnnounce,
		IsCommunity:          info.IsParent,
		Participants:         participantSummaries(info.Participants),
		SubjectTime:          subjectTime,
		SubjectOwner:         subjectOwner,
	}
}

func invitationMetadataFromGroup(info *types.GroupInfo, inviteLink string) InvitationMetadata {
	owner := info.OwnerPN.User
	if owner == "" {
		owner = info.OwnerJID.User
	}
	subjectOwner := info.GroupName.NameSetBy.User
	subjectTime := int64(0)
	if !info.GroupName.NameSetAt.IsZero() {
		subjectTime = info.GroupName.NameSetAt.UTC().UnixMilli()
	}

	return InvitationMetadata{
		Phone:             FormatGroupID(info.JID),
		Owner:             owner,
		Subject:           info.GroupName.Name,
		Description:       info.GroupTopic.Topic,
		Creation:          info.GroupCreated.UTC().UnixMilli(),
		InvitationLink:    inviteLink,
		Participants:      participantSummaries(info.Participants),
		ParticipantsCount: len(info.Participants),
		ContactsCount:     len(info.Participants),
		SubjectTime:       subjectTime,
		SubjectOwner:      subjectOwner,
	}
}

func imageBytesFromInput(ctx context.Context, src string) ([]byte, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return nil, ErrOperationFailed
	}

	if strings.HasPrefix(strings.ToLower(src), "data:") {
		parts := strings.SplitN(src, ",", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%w: malformed data URI", ErrOperationFailed)
		}
		payload := parts[1]
		if decoded, err := base64.StdEncoding.DecodeString(payload); err == nil {
			return decoded, nil
		}
		return nil, fmt.Errorf("decode base64 image: %w", ErrOperationFailed)
	}

	if strings.HasPrefix(strings.ToLower(src), "http://") || strings.HasPrefix(strings.ToLower(src), "https://") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
		if err != nil {
			return nil, fmt.Errorf("build http request: %w", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("download image: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("download image: status %d", resp.StatusCode)
		}
		reader := io.LimitReader(resp.Body, maxPhotoBytes+1)
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("read image: %w", err)
		}
		if len(data) > maxPhotoBytes {
			return nil, fmt.Errorf("image exceeds %d bytes", maxPhotoBytes)
		}
		return data, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return nil, fmt.Errorf("decode base64 image: %w", err)
	}
	return decoded, nil
}

func sendAutoInvites(ctx context.Context, client Client, logger *slog.Logger, phones []types.JID, inviteLink string) {
	if len(phones) == 0 || inviteLink == "" {
		return
	}
	messageText := fmt.Sprintf(defaultInviteMessage, inviteLink)
	msg := &waProto.Message{
		Conversation: &messageText,
	}
	for _, phone := range phones {
		if _, err := client.SendMessage(ctx, phone, msg); err != nil && logger != nil {
			logger.Warn("failed to send auto invite",
				slog.String("participant", phone.String()),
				slog.String("error", err.Error()))
		}
		time.Sleep(50 * time.Millisecond) // avoid flooding in rapid loops
	}
}

func failedParticipants(requested []types.JID, response []types.GroupParticipant) []types.JID {
	if len(requested) == 0 {
		return nil
	}
	failed := make([]types.JID, 0)
	requestedMap := make(map[string]types.JID, len(requested))
	for _, jid := range requested {
		requestedMap[jid.User] = jid
	}
	for _, part := range response {
		if part.Error != 0 {
			if jid, ok := requestedMap[part.JID.User]; ok {
				failed = append(failed, jid)
				delete(requestedMap, part.JID.User)
				continue
			}
		}
		delete(requestedMap, part.JID.User)
	}
	for _, remaining := range requestedMap {
		failed = append(failed, remaining)
	}
	return failed
}

func pruneFailedPhones(failed []types.JID) []string {
	if len(failed) == 0 {
		return nil
	}
	result := make([]string, 0, len(failed))
	for _, f := range failed {
		if f.User != "" {
			result = append(result, f.User)
		}
	}
	return result
}

func deriveInviteLink(client Client, groupJID types.JID, reset bool) (string, error) {
	link, err := client.GetGroupInviteLink(groupJID, reset)
	if err != nil {
		return "", err
	}
	return link, nil
}

func ensureGroupName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidGroupName
	}
	if len([]rune(name)) > maxGroupNameLength {
		return fmt.Errorf("%w: exceeds %d characters", ErrInvalidGroupName, maxGroupNameLength)
	}
	return nil
}

func normalizeInviteURL(url string) (string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", ErrInvalidInviteURL
	}
	if strings.Contains(url, "chat.whatsapp.com/") {
		parts := strings.Split(url, "chat.whatsapp.com/")
		return parts[len(parts)-1], nil
	}
	return url, nil
}

// IsValidDescriptionID checks if a description/topic ID is a valid WhatsApp message ID.
// Valid IDs are hexadecimal strings (typically 16+ chars starting with "3EB0" for web clients).
// Invalid values like "undefined", "null", or empty strings return false.
func IsValidDescriptionID(id string) bool {
	if id == "" || id == "undefined" || id == "null" {
		return false
	}
	// Valid message IDs are hex strings, minimum 16 characters
	if len(id) < 16 {
		return false
	}
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}
