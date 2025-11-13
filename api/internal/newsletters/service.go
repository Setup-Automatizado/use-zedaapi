package newsletters

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/types"
)

// Service orchestrates newsletter listings using the WhatsApp client.
type Service struct {
	provider ClientProvider
	log      *slog.Logger
}

const maxNewsletterPictureBytes = 1 << 20

const (
	mutationUpdateNewsletter            = "7150902998257522"
	mutationDeleteNewsletter            = "6285734628148226"
	mutationNewsletterAdminInvite       = "24943748628557365"
	mutationNewsletterAdminInviteRevoke = "6550386328343169"
	mutationNewsletterAdminDemote       = "7220922401252829"
	mutationNewsletterAcceptAdminInvite = "6179636105471882"
	mutationNewsletterChangeOwner       = "6951013521615265"
	queryNewslettersDirectorySearch     = "8422355807877290"
)

type newsletterSearchResponse struct {
	Search struct {
		Result   []*types.NewsletterMetadata `json:"result"`
		PageInfo struct {
			EndCursor   *string `json:"endCursor"`
			StartCursor *string `json:"startCursor"`
		} `json:"page_info"`
	} `json:"xwa2_newsletters_directory_search"`
}

// NewService builds a newsletters service with the supplied dependencies.
func NewService(provider ClientProvider, log *slog.Logger) *Service {
	return &Service{
		provider: provider,
		log:      log,
	}
}

// List returns a paginated slice of newsletters available to the instance.
func (s *Service) List(ctx context.Context, instanceID uuid.UUID, params ListParams) (ListResult, error) {
	if params.Page <= 0 || params.PageSize <= 0 {
		return ListResult{}, ErrInvalidPagination
	}

	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
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

	metas, err := client.GetSubscribedNewsletters(ctx)
	if err != nil {
		logger.Error("failed to fetch subscribed newsletters",
			slog.String("error", err.Error()))
		return ListResult{}, fmt.Errorf("get subscribed newsletters: %w", err)
	}

	summaries := make([]Summary, 0, len(metas))
	for _, meta := range metas {
		if meta == nil {
			continue
		}
		summaries = append(summaries, toSummary(meta))
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
		logger.Info("list request beyond total newsletters",
			slog.Int("total_newsletters", total))
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

	logger.Info("newsletters listed successfully",
		slog.Int("total_newsletters", total),
		slog.Int("returned_newsletters", len(result.Items)))

	return result, nil
}

// Create provisions a new newsletter.
func (s *Service) Create(ctx context.Context, instanceID uuid.UUID, params CreateParams) (CreateResult, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		return CreateResult{}, ErrInvalidName
	}

	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
		slog.String("operation", "create"),
		slog.String("instance_id", instanceID.String()),
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

	createParams := whatsmeow.CreateNewsletterParams{
		Name: name,
	}
	if params.Description != nil {
		createParams.Description = *params.Description
	}
	if strings.TrimSpace(params.Picture) != "" {
		pictureBytes, err := imageBytesFromInput(ctx, params.Picture)
		if err != nil {
			logger.Warn("invalid picture payload",
				slog.String("error", err.Error()))
			return CreateResult{}, fmt.Errorf("%w: %v", ErrInvalidPicture, err)
		}
		createParams.Picture = pictureBytes
	}

	meta, err := client.CreateNewsletter(ctx, createParams)
	if err != nil {
		logger.Error("failed to create newsletter",
			slog.String("error", err.Error()))
		return CreateResult{}, fmt.Errorf("create newsletter: %w", err)
	}

	logger.Info("newsletter created successfully",
		slog.String("newsletter_id", meta.ID.String()))

	return CreateResult{ID: meta.ID.String()}, nil
}

// UpdatePicture sets a new picture for the newsletter.
func (s *Service) UpdatePicture(ctx context.Context, instanceID uuid.UUID, params UpdatePictureParams) (OperationResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
		slog.String("operation", "update_picture"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", params.ID),
	)

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

	jid, err := parseNewsletterID(params.ID)
	if err != nil {
		logger.Warn("invalid newsletter id",
			slog.String("error", err.Error()))
		return OperationResult{}, err
	}

	pictureBytes, err := imageBytesFromInput(ctx, params.Picture)
	if err != nil {
		logger.Warn("invalid picture payload",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("%w: %v", ErrInvalidPicture, err)
	}

	uploadResp, err := client.UploadNewsletter(ctx, pictureBytes, whatsmeow.MediaImage)
	if err != nil {
		logger.Error("failed to upload newsletter picture",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("upload picture: %w", err)
	}

	updates := map[string]any{
		"thread_metadata": map[string]any{
			"picture": map[string]any{
				"handle":      uploadResp.Handle,
				"direct_path": uploadResp.DirectPath,
				"id":          uploadResp.ObjectID,
				"url":         uploadResp.URL,
				"type":        "image",
			},
		},
	}

	if err := sendNewsletterUpdate(ctx, client, jid, updates); err != nil {
		logger.Error("failed to apply newsletter picture",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("update picture: %w", err)
	}

	logger.Info("newsletter picture updated successfully")
	return OperationResult{Value: true}, nil
}

// UpdateName changes the display name of a newsletter.
func (s *Service) UpdateName(ctx context.Context, instanceID uuid.UUID, params UpdateNameParams) (OperationResult, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		return OperationResult{}, ErrInvalidName
	}

	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
		slog.String("operation", "update_name"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", params.ID),
	)

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

	jid, err := parseNewsletterID(params.ID)
	if err != nil {
		logger.Warn("invalid newsletter id",
			slog.String("error", err.Error()))
		return OperationResult{}, err
	}

	updates := map[string]any{
		"thread_metadata": map[string]any{
			"name": map[string]any{
				"text": name,
			},
		},
	}

	if err := sendNewsletterUpdate(ctx, client, jid, updates); err != nil {
		logger.Error("failed to update newsletter name",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("update name: %w", err)
	}

	logger.Info("newsletter name updated successfully")
	return OperationResult{Value: true}, nil
}

// UpdateDescription updates the textual description of a newsletter.
func (s *Service) UpdateDescription(ctx context.Context, instanceID uuid.UUID, params UpdateDescriptionParams) (OperationResult, error) {
	description := strings.TrimSpace(params.Description)

	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
		slog.String("operation", "update_description"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", params.ID),
	)

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

	jid, err := parseNewsletterID(params.ID)
	if err != nil {
		logger.Warn("invalid newsletter id",
			slog.String("error", err.Error()))
		return OperationResult{}, err
	}

	updates := map[string]any{
		"thread_metadata": map[string]any{
			"description": map[string]any{
				"text": description,
			},
		},
	}

	if err := sendNewsletterUpdate(ctx, client, jid, updates); err != nil {
		logger.Error("failed to update newsletter description",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("update description: %w", err)
	}

	logger.Info("newsletter description updated successfully")
	return OperationResult{Value: true}, nil
}

// Follow subscribes the instance to a newsletter.
func (s *Service) Follow(ctx context.Context, instanceID uuid.UUID, params IDParams) (OperationResult, error) {
	return s.toggleFollow(ctx, instanceID, params.ID, true)
}

// Unfollow unsubscribes the instance from a newsletter.
func (s *Service) Unfollow(ctx context.Context, instanceID uuid.UUID, params IDParams) (OperationResult, error) {
	return s.toggleFollow(ctx, instanceID, params.ID, false)
}

// Mute silences notifications for a newsletter.
func (s *Service) Mute(ctx context.Context, instanceID uuid.UUID, params IDParams) (OperationResult, error) {
	return s.toggleMute(ctx, instanceID, params.ID, true)
}

// Unmute re-enables notifications for a newsletter.
func (s *Service) Unmute(ctx context.Context, instanceID uuid.UUID, params IDParams) (OperationResult, error) {
	return s.toggleMute(ctx, instanceID, params.ID, false)
}

// Delete removes a newsletter owned by the account.
func (s *Service) Delete(ctx context.Context, instanceID uuid.UUID, params IDParams) (OperationResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
		slog.String("operation", "delete"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", params.ID),
	)

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

	jid, err := parseNewsletterID(params.ID)
	if err != nil {
		logger.Warn("invalid newsletter id",
			slog.String("error", err.Error()))
		return OperationResult{}, err
	}

	if _, err := sendNewsletterMutation(ctx, client, mutationDeleteNewsletter, map[string]any{
		"newsletter_id": jid.String(),
	}); err != nil {
		logger.Error("failed to delete newsletter",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("delete newsletter: %w", err)
	}

	logger.Info("newsletter deleted successfully")
	return OperationResult{Value: true}, nil
}

// Metadata returns detailed information about a newsletter.
func (s *Service) Metadata(ctx context.Context, instanceID uuid.UUID, id string) (MetadataResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
		slog.String("operation", "metadata"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", id),
	)

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return MetadataResult{}, err
		}
		logger.Error("failed to obtain whatsapp client",
			slog.String("error", err.Error()))
		return MetadataResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	jid, err := parseNewsletterID(id)
	if err != nil {
		logger.Warn("invalid newsletter id",
			slog.String("error", err.Error()))
		return MetadataResult{}, err
	}

	meta, err := client.GetNewsletterInfo(jid)
	if err != nil {
		logger.Error("failed to fetch newsletter metadata",
			slog.String("error", err.Error()))
		return MetadataResult{}, fmt.Errorf("get newsletter info: %w", err)
	}
	if meta == nil {
		return MetadataResult{}, fmt.Errorf("newsletter not found")
	}

	logger.Info("newsletter metadata fetched successfully")
	result := MetadataResult(toSummary(meta))
	return result, nil
}

// Search lists public newsletters using the directory search endpoint.
func (s *Service) Search(ctx context.Context, instanceID uuid.UUID, params SearchParams) (SearchResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
		slog.String("operation", "search"),
		slog.String("instance_id", instanceID.String()),
	)

	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		if err == ErrClientNotConnected {
			logger.Warn("whatsapp client not connected")
			return SearchResult{}, err
		}
		logger.Error("failed to obtain whatsapp client",
			slog.String("error", err.Error()))
		return SearchResult{}, fmt.Errorf("get whatsapp client: %w", err)
	}

	input := map[string]any{}
	if params.Limit > 0 {
		input["limit"] = params.Limit
	}
	if params.View != "" {
		input["view"] = params.View
	}
	if len(params.CountryCodes) > 0 {
		input["filters"] = map[string]any{
			"country_codes": params.CountryCodes,
		}
	}
	if params.SearchText != nil {
		input["search_text"] = *params.SearchText
	}

	payload := map[string]any{"input": input}
	raw, err := sendNewsletterMutation(ctx, client, queryNewslettersDirectorySearch, payload)
	if err != nil {
		logger.Error("failed to search newsletters",
			slog.String("error", err.Error()))
		return SearchResult{}, fmt.Errorf("search newsletters: %w", err)
	}

	var resp newsletterSearchResponse
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &resp); err != nil {
			logger.Error("failed to decode search response",
				slog.String("error", err.Error()))
			return SearchResult{}, fmt.Errorf("decode search response: %w", err)
		}
	}

	items := make([]SearchItem, 0, len(resp.Search.Result))
	for _, meta := range resp.Search.Result {
		items = append(items, toSearchItem(meta))
	}

	logger.Info("newsletter search completed",
		slog.Int("results", len(items)))

	return SearchResult{Cursor: resp.Search.PageInfo.EndCursor, Data: items}, nil
}

// UpdateSettings applies settings mutations to a newsletter.
func (s *Service) UpdateSettings(ctx context.Context, instanceID uuid.UUID, params SettingsParams) (OperationResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
		slog.String("operation", "settings"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", params.ID),
	)

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

	jid, err := parseNewsletterID(params.ID)
	if err != nil {
		logger.Warn("invalid newsletter id",
			slog.String("error", err.Error()))
		return OperationResult{}, err
	}

	mode := strings.ToLower(strings.TrimSpace(params.ReactionCodes))
	if mode == "" {
		return OperationResult{}, ErrInvalidReactionCodes
	}
	reactionMode := types.NewsletterReactionsMode(mode)
	switch reactionMode {
	case types.NewsletterReactionsModeAll, types.NewsletterReactionsModeBasic, types.NewsletterReactionsModeNone, types.NewsletterReactionsModeBlocklist:
	default:
		return OperationResult{}, ErrInvalidReactionCodes
	}
	updates := map[string]any{
		"settings": map[string]any{
			"reaction_codes": map[string]any{
				"value": string(reactionMode),
			},
		},
	}

	if err := sendNewsletterUpdate(ctx, client, jid, updates); err != nil {
		logger.Error("failed to update newsletter settings",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("update settings: %w", err)
	}

	logger.Info("newsletter settings updated successfully")
	return OperationResult{Value: true}, nil
}

// SendAdminInvite dispatches an administrator invitation to a phone number.
func (s *Service) SendAdminInvite(ctx context.Context, instanceID uuid.UUID, params AdminActionParams) (OperationResult, error) {
	return s.adminMutation(ctx, instanceID, params, "send_admin_invite", mutationNewsletterAdminInvite)
}

// AcceptAdminInvite accepts a pending administrator invitation.
func (s *Service) AcceptAdminInvite(ctx context.Context, instanceID uuid.UUID, params IDParams) (OperationResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
		slog.String("operation", "accept_admin_invite"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", params.ID),
	)

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

	jid, err := parseNewsletterID(params.ID)
	if err != nil {
		logger.Warn("invalid newsletter id",
			slog.String("error", err.Error()))
		return OperationResult{}, err
	}

	if _, err := sendNewsletterMutation(ctx, client, mutationNewsletterAcceptAdminInvite, map[string]any{
		"newsletter_id": jid.String(),
	}); err != nil {
		logger.Error("failed to accept admin invite",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("accept admin invite: %w", err)
	}

	logger.Info("newsletter admin invite accepted")
	return OperationResult{Value: true}, nil
}

// RemoveAdmin revokes administrator privileges.
func (s *Service) RemoveAdmin(ctx context.Context, instanceID uuid.UUID, params AdminActionParams) (OperationResult, error) {
	return s.adminMutation(ctx, instanceID, params, "remove_admin", mutationNewsletterAdminDemote)
}

// RevokeAdminInvite cancels a previously sent administrator invitation.
func (s *Service) RevokeAdminInvite(ctx context.Context, instanceID uuid.UUID, params AdminActionParams) (OperationResult, error) {
	return s.adminMutation(ctx, instanceID, params, "revoke_admin_invite", mutationNewsletterAdminInviteRevoke)
}

// TransferOwnership hands ownership to another administrator.
func (s *Service) TransferOwnership(ctx context.Context, instanceID uuid.UUID, params TransferOwnershipParams) (OperationResult, error) {
	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
		slog.String("operation", "transfer_ownership"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", params.ID),
		slog.String("target_phone", params.Phone),
	)

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

	jid, err := parseNewsletterID(params.ID)
	if err != nil {
		logger.Warn("invalid newsletter id",
			slog.String("error", err.Error()))
		return OperationResult{}, err
	}

	if _, err := sendNewsletterMutation(ctx, client, mutationNewsletterChangeOwner, map[string]any{
		"newsletter_id": jid.String(),
		"phone":         strings.TrimSpace(params.Phone),
		"quit_admin":    params.QuitAdmin,
	}); err != nil {
		logger.Error("failed to transfer ownership",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("transfer ownership: %w", err)
	}

	logger.Info("newsletter ownership transferred successfully")
	return OperationResult{Value: true}, nil
}

func toSummary(meta *types.NewsletterMetadata) Summary {
	thread := meta.ThreadMeta
	viewMeta := meta.ViewerMeta

	summary := Summary{
		ID:               meta.ID.String(),
		CreationTime:     formatUnix(thread.CreationTime.Time),
		State:            strings.ToUpper(string(meta.State.Type)),
		Name:             thread.Name.Text,
		Description:      thread.Description.Text,
		SubscribersCount: strconv.Itoa(thread.SubscriberCount),
		InviteLink:       buildInviteLink(thread.InviteCode),
		Verification:     strings.ToUpper(string(thread.VerificationState)),
		Picture:          optionalURL(thread.Picture),
		Preview:          optionalPreviewURL(thread.Preview),
	}

	if viewMeta != nil {
		summary.ViewMetadata = &ViewMetadata{
			Mute: strings.ToUpper(string(viewMeta.Mute)),
			Role: strings.ToUpper(string(viewMeta.Role)),
		}
	}

	return summary
}

func buildInviteLink(code string) string {
	if code == "" {
		return ""
	}
	code = strings.TrimPrefix(code, whatsmeow.NewsletterLinkPrefix)
	return whatsmeow.NewsletterLinkPrefix + code
}

func optionalURL(info *types.ProfilePictureInfo) *string {
	if info == nil || info.URL == "" {
		return nil
	}
	url := info.URL
	return &url
}

func optionalPreviewURL(info types.ProfilePictureInfo) *string {
	if info.URL == "" {
		return nil
	}
	url := info.URL
	return &url
}

func formatUnix(t time.Time) string {
	if t.IsZero() {
		return "0"
	}
	return strconv.FormatInt(t.Unix(), 10)
}

func (s *Service) toggleFollow(ctx context.Context, instanceID uuid.UUID, id string, follow bool) (OperationResult, error) {
	operation := "follow"
	if !follow {
		operation = "unfollow"
	}

	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
		slog.String("operation", operation),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", id),
	)

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

	jid, err := parseNewsletterID(id)
	if err != nil {
		logger.Warn("invalid newsletter id",
			slog.String("error", err.Error()))
		return OperationResult{}, err
	}

	if follow {
		err = client.FollowNewsletter(jid)
	} else {
		err = client.UnfollowNewsletter(jid)
	}
	if err != nil {
		logger.Error("failed to toggle follow state",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("%s newsletter: %w", operation, err)
	}

	logger.Info("newsletter follow state updated",
		slog.Bool("follow", follow))
	return OperationResult{Value: true}, nil
}

func (s *Service) toggleMute(ctx context.Context, instanceID uuid.UUID, id string, mute bool) (OperationResult, error) {
	operation := "mute"
	if !mute {
		operation = "unmute"
	}

	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
		slog.String("operation", operation),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", id),
	)

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

	jid, err := parseNewsletterID(id)
	if err != nil {
		logger.Warn("invalid newsletter id",
			slog.String("error", err.Error()))
		return OperationResult{}, err
	}

	if err := client.NewsletterToggleMute(jid, mute); err != nil {
		logger.Error("failed to toggle mute state",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("%s newsletter: %w", operation, err)
	}

	logger.Info("newsletter mute state updated",
		slog.Bool("mute", mute))
	return OperationResult{Value: true}, nil
}

func (s *Service) adminMutation(ctx context.Context, instanceID uuid.UUID, params AdminActionParams, operation string, mutationID string) (OperationResult, error) {
	phone := strings.TrimSpace(params.Phone)
	if phone == "" {
		return OperationResult{}, ErrInvalidPhone
	}

	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "newsletters_service"),
		slog.String("operation", operation),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", params.ID),
		slog.String("phone", phone),
	)

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

	jid, err := parseNewsletterID(params.ID)
	if err != nil {
		logger.Warn("invalid newsletter id",
			slog.String("error", err.Error()))
		return OperationResult{}, err
	}

	if _, err := sendNewsletterMutation(ctx, client, mutationID, map[string]any{
		"newsletter_id": jid.String(),
		"phone":         phone,
	}); err != nil {
		logger.Error("admin mutation failed",
			slog.String("error", err.Error()))
		return OperationResult{}, fmt.Errorf("%s: %w", operation, err)
	}

	logger.Info("newsletter admin mutation executed")
	return OperationResult{Value: true}, nil
}

func sendNewsletterUpdate(ctx context.Context, client Client, jid types.JID, updates map[string]any) error {
	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}
	_, err := client.SendNewsletterMex(ctx, mutationUpdateNewsletter, map[string]any{
		"newsletter_id": jid.String(),
		"updates":       updates,
	})
	return err
}

func sendNewsletterMutation(ctx context.Context, client Client, queryID string, variables map[string]any) ([]byte, error) {
	return client.SendNewsletterMex(ctx, queryID, variables)
}

func parseNewsletterID(id string) (types.JID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return types.EmptyJID, errors.New("newsletter id is required")
	}
	if !strings.Contains(id, "@") {
		id = id + "@newsletter"
	}
	jid, err := types.ParseJID(id)
	if err != nil {
		return types.EmptyJID, fmt.Errorf("parse newsletter id: %w", err)
	}
	if jid.Server != types.NewsletterServer {
		return types.EmptyJID, fmt.Errorf("invalid newsletter server: %s", jid.Server)
	}
	return jid, nil
}

func imageBytesFromInput(ctx context.Context, src string) ([]byte, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return nil, errors.New("empty image source")
	}

	lower := strings.ToLower(src)
	if strings.HasPrefix(lower, "data:") {
		parts := strings.SplitN(src, ",", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed data uri")
		}
		raw, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, fmt.Errorf("decode data uri: %w", err)
		}
		if len(raw) > maxNewsletterPictureBytes {
			return nil, fmt.Errorf("image exceeds %d bytes", maxNewsletterPictureBytes)
		}
		return raw, nil
	}

	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
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
		reader := io.LimitReader(resp.Body, maxNewsletterPictureBytes+1)
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("read image: %w", err)
		}
		if len(data) > maxNewsletterPictureBytes {
			return nil, fmt.Errorf("image exceeds %d bytes", maxNewsletterPictureBytes)
		}
		return data, nil
	}

	data, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return nil, fmt.Errorf("decode base64 image: %w", err)
	}
	if len(data) > maxNewsletterPictureBytes {
		return nil, fmt.Errorf("image exceeds %d bytes", maxNewsletterPictureBytes)
	}
	return data, nil
}

func toSearchItem(meta *types.NewsletterMetadata) SearchItem {
	if meta == nil {
		return SearchItem{}
	}
	description := strings.TrimSpace(meta.ThreadMeta.Description.Text)
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}
	return SearchItem{
		ID:               meta.ID.String(),
		Name:             meta.ThreadMeta.Name.Text,
		Description:      descriptionPtr,
		SubscribersCount: strconv.Itoa(meta.ThreadMeta.SubscriberCount),
		Picture:          optionalURL(meta.ThreadMeta.Picture),
	}
}
