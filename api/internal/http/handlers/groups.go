package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/communities"
	"go.mau.fi/whatsmeow/api/internal/groups"
	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// GroupsService defines the required behaviour of the groups domain service.
type GroupsService interface {
	List(ctx context.Context, instanceID uuid.UUID, params groups.ListParams) (groups.ListResult, error)
	Create(ctx context.Context, instanceID uuid.UUID, params groups.CreateParams) (groups.CreateResult, error)
	UpdateName(ctx context.Context, instanceID uuid.UUID, params groups.UpdateNameParams) (groups.ValueResult, error)
	UpdatePhoto(ctx context.Context, instanceID uuid.UUID, params groups.UpdatePhotoParams) (groups.ValueResult, error)
	AddParticipants(ctx context.Context, instanceID uuid.UUID, params groups.ModifyParticipantsParams) (groups.ValueResult, error)
	RemoveParticipants(ctx context.Context, instanceID uuid.UUID, params groups.ModifyParticipantsParams) (groups.ValueResult, error)
	ApproveParticipants(ctx context.Context, instanceID uuid.UUID, params groups.ModifyParticipantsParams) (groups.ValueResult, error)
	RejectParticipants(ctx context.Context, instanceID uuid.UUID, params groups.ModifyParticipantsParams) (groups.ValueResult, error)
	AddAdmins(ctx context.Context, instanceID uuid.UUID, params groups.ModifyParticipantsParams) (groups.ValueResult, error)
	RemoveAdmins(ctx context.Context, instanceID uuid.UUID, params groups.ModifyParticipantsParams) (groups.ValueResult, error)
	Leave(ctx context.Context, instanceID uuid.UUID, params groups.ModifyParticipantsParams) (groups.ValueResult, error)
	Metadata(ctx context.Context, instanceID uuid.UUID, groupID string) (groups.Metadata, error)
	LightMetadata(ctx context.Context, instanceID uuid.UUID, groupID string) (groups.Metadata, error)
	InvitationMetadata(ctx context.Context, instanceID uuid.UUID, inviteURL string) (groups.InvitationMetadata, error)
	InvitationLink(ctx context.Context, instanceID uuid.UUID, groupID string) (groups.InvitationLinkResult, error)
	RedefineInvitationLink(ctx context.Context, instanceID uuid.UUID, groupID string) (groups.InvitationLinkResult, error)
	UpdateSettings(ctx context.Context, instanceID uuid.UUID, params groups.UpdateSettingsParams) (groups.ValueResult, error)
	UpdateDescription(ctx context.Context, instanceID uuid.UUID, params groups.UpdateDescriptionParams) (groups.ValueResult, error)
	AcceptInvite(ctx context.Context, instanceID uuid.UUID, inviteURL string) (groups.AcceptInviteResult, error)
}

// GroupsHandler exposes group-related endpoints matching the Zé da API surface.
type GroupsHandler struct {
	instanceService    InstanceStatusProvider
	service            GroupsService
	communitiesService CommunitiesService
	metrics            *observability.Metrics
	log                *slog.Logger
}

// NewGroupsHandler wires the HTTP handler with the underlying dependencies.
func NewGroupsHandler(
	instanceService InstanceStatusProvider,
	service GroupsService,
	communitiesService CommunitiesService,
	metrics *observability.Metrics,
	log *slog.Logger,
) *GroupsHandler {
	var handlerLogger *slog.Logger
	if log != nil {
		handlerLogger = log.With(slog.String("handler", "groups"))
	}
	return &GroupsHandler{
		instanceService:    instanceService,
		service:            service,
		communitiesService: communitiesService,
		metrics:            metrics,
		log:                handlerLogger,
	}
}

// RegisterRoutes attaches all group routes, combining implemented and placeholder endpoints.
func (h *GroupsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/groups", h.listGroups)
	r.Post("/create-group", h.createGroup)
	r.Post("/update-group-name", h.updateGroupName)
	r.Post("/update-group-photo", h.updateGroupPhoto)
	r.Post("/add-participant", h.addParticipant)
	r.Post("/remove-participant", h.removeParticipant)
	r.Post("/approve-participant", h.approveParticipant)
	r.Post("/reject-participant", h.rejectParticipant)
	r.Post("/add-admin", h.addAdmin)
	r.Post("/remove-admin", h.removeAdmin)
	r.Post("/leave-group", h.leaveGroup)
	r.Get("/group-metadata/{groupId}", h.groupMetadata)
	r.Get("/light-group-metadata/{groupId}", h.lightGroupMetadata)
	r.Get("/group-invitation-metadata", h.groupInvitationMetadata)
	r.Post("/group-invitation-link/{groupId}", h.groupInvitationLink)
	r.Post("/redefine-invitation-link/{groupId}", h.redefineGroupInvitationLink)
	r.Post("/update-group-settings", h.updateGroupSettings)
	r.Post("/update-group-description", h.updateGroupDescription)
	r.Get("/accept-invite-group", h.acceptInviteGroup)
}

func (h *GroupsHandler) startMetrics(operation string) func(string) {
	start := time.Now()
	return func(status string) {
		if h.metrics == nil {
			return
		}
		if h.metrics.GroupsLatency != nil {
			h.metrics.GroupsLatency.WithLabelValues(operation).Observe(time.Since(start).Seconds())
		}
		if h.metrics.GroupsRequests != nil {
			h.metrics.GroupsRequests.WithLabelValues(operation, status).Inc()
		}
	}
}

func decodeRequest(r *http.Request, dst interface{}) error {
	if dst == nil {
		return errors.New("nil decode target")
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func (h *GroupsHandler) handleGroupError(logger *slog.Logger, w http.ResponseWriter, operation string, instanceID uuid.UUID, err error) {
	switch {
	case errors.Is(err, groups.ErrInvalidGroupID):
		respondError(w, http.StatusBadRequest, "invalid group id")
	case errors.Is(err, groups.ErrInvalidGroupName):
		respondError(w, http.StatusBadRequest, "invalid group name")
	case errors.Is(err, groups.ErrInvalidPhoneList):
		respondError(w, http.StatusBadRequest, "phones are required")
	case errors.Is(err, groups.ErrInvalidInviteURL):
		respondError(w, http.StatusBadRequest, "invalid invite url")
	case errors.Is(err, groups.ErrClientNotConnected):
		respondError(w, http.StatusServiceUnavailable, "whatsapp client not connected")
	case errors.Is(err, instances.ErrInstanceNotFound):
		respondError(w, http.StatusNotFound, "instance not found")
	case errors.Is(err, instances.ErrUnauthorized):
		respondError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, instances.ErrInstanceInactive):
		respondError(w, http.StatusForbidden, "instance subscription inactive")
	case errors.Is(err, whatsmeow.ErrGroupNotFound):
		respondError(w, http.StatusNotFound, "group not found")
	case errors.Is(err, whatsmeow.ErrNotInGroup):
		respondError(w, http.StatusForbidden, "not participant of group")
	case errors.Is(err, whatsmeow.ErrGroupInviteLinkUnauthorized):
		respondError(w, http.StatusForbidden, "invitation link unavailable")
	case errors.Is(err, whatsmeow.ErrInviteLinkInvalid):
		respondError(w, http.StatusBadRequest, "invalid invite link")
	case errors.Is(err, whatsmeow.ErrInviteLinkRevoked):
		respondError(w, http.StatusGone, "invite link revoked")
	case errors.Is(err, whatsmeow.ErrInvalidImageFormat):
		respondError(w, http.StatusBadRequest, "invalid image format")
	default:
		logger.Error("group operation failed",
			slog.String("error", err.Error()))
		captureHandlerError("groups", operation, instanceID.String(), err)
		respondError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *GroupsHandler) handleCommunityError(logger *slog.Logger, w http.ResponseWriter, operation string, instanceID uuid.UUID, err error) {
	switch {
	case errors.Is(err, communities.ErrInvalidCommunityID):
		respondError(w, http.StatusBadRequest, "invalid community id")
	case errors.Is(err, communities.ErrInvalidCommunityName):
		respondError(w, http.StatusBadRequest, "invalid community name")
	case errors.Is(err, communities.ErrInvalidGroupList):
		respondError(w, http.StatusBadRequest, "invalid groups list")
	case errors.Is(err, communities.ErrInvalidPhoneList):
		respondError(w, http.StatusBadRequest, "phones are required")
	case errors.Is(err, communities.ErrClientNotConnected):
		respondError(w, http.StatusServiceUnavailable, "whatsapp client not connected")
	case errors.Is(err, instances.ErrInstanceNotFound):
		respondError(w, http.StatusNotFound, "instance not found")
	case errors.Is(err, instances.ErrUnauthorized):
		respondError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, instances.ErrInstanceInactive):
		respondError(w, http.StatusForbidden, "instance subscription inactive")
	default:
		logger.Error("community operation failed",
			slog.String("error", err.Error()))
		captureHandlerError("communities", operation, instanceID.String(), err)
		respondError(w, http.StatusInternalServerError, "internal error")
	}
}

type participantsRequest struct {
	GroupID     string   `json:"groupId"`
	CommunityID string   `json:"communityId"`
	Phones      []string `json:"phones"`
	AutoInvite  bool     `json:"autoInvite"`
}

func (h *GroupsHandler) handleParticipants(
	w http.ResponseWriter,
	r *http.Request,
	operationKey string,
	httpOperation string,
	includeAutoInvite bool,
	action func(context.Context, uuid.UUID, groups.ModifyParticipantsParams) (groups.ValueResult, error),
) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics(operationKey)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "groups_http_handler"),
		slog.String("operation", httpOperation),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload participantsRequest
	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid request payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	maskedCommunityID := payload.CommunityID
	if payload.CommunityID != "" {
		ctx = logging.WithAttrs(ctx, slog.String("community_id", maskedCommunityID))
		logger = logging.ContextLogger(ctx, h.log)
		if h.communitiesService == nil {
			status = "error"
			logger.Error("communities service unavailable for participant operation")
			respondError(w, http.StatusInternalServerError, "communities service unavailable")
			return
		}
		commParams := communities.ParticipantsParams{
			CommunityID: payload.CommunityID,
			Phones:      payload.Phones,
		}
		if includeAutoInvite {
			commParams.AutoInvite = payload.AutoInvite
		}
		var (
			result communities.OperationResult
			err    error
		)
		switch operationKey {
		case "add_participant":
			result, err = h.communitiesService.AddParticipants(ctx, instanceID, commParams)
		case "remove_participant":
			result, err = h.communitiesService.RemoveParticipants(ctx, instanceID, commParams)
		case "add_admin":
			result, err = h.communitiesService.AddAdmins(ctx, instanceID, commParams)
		case "remove_admin":
			result, err = h.communitiesService.RemoveAdmins(ctx, instanceID, commParams)
		default:
			status = "error"
			logger.Warn("operation unsupported for communities",
				slog.String("operation", operationKey))
			respondError(w, http.StatusBadRequest, "operation not available for communities")
			return
		}
		if err != nil {
			status = "error"
			h.handleCommunityError(logger, w, httpOperation, instanceID, err)
			return
		}
		value, _ := result.ValueBool()
		respondJSON(w, http.StatusOK, groups.ValueResult{Value: value})
		logger.Info("community participant operation completed",
			slog.String("community_id", maskedCommunityID),
			slog.Int("phone_count", len(payload.Phones)),
			slog.Bool("auto_invite", commParams.AutoInvite))
		return
	}

	params := groups.ModifyParticipantsParams{
		GroupID: payload.GroupID,
		Phones:  payload.Phones,
	}
	if includeAutoInvite {
		params.AutoInvite = payload.AutoInvite
	}

	result, err := action(ctx, instanceID, params)
	if err != nil {
		status = "error"
		h.handleGroupError(logger, w, httpOperation, instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)

	logger.Info("participant operation completed",
		slog.String("group_id", payload.GroupID),
		slog.Int("phone_count", len(payload.Phones)),
		slog.Bool("auto_invite", params.AutoInvite))
}

func (h *GroupsHandler) listGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operation := "list"
	status := "success"
	observe := h.startMetrics(operation)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "groups_http_handler"),
		slog.String("operation", "GET /groups"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageStr == "" || pageSizeStr == "" {
		status = "error"
		logger.Warn("missing pagination parameters")
		respondError(w, http.StatusBadRequest, "page and pageSize are required")
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		status = "error"
		logger.Warn("invalid page parameter",
			slog.String("value", pageStr),
			slog.String("error", errString(err)))
		respondError(w, http.StatusBadRequest, "invalid page parameter")
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		status = "error"
		logger.Warn("invalid pageSize parameter",
			slog.String("value", pageSizeStr),
			slog.String("error", errString(err)))
		respondError(w, http.StatusBadRequest, "invalid pageSize parameter")
		return
	}

	result, err := h.service.List(ctx, instanceID, groups.ListParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		status = "error"
		switch {
		case errors.Is(err, groups.ErrInvalidPagination):
			respondError(w, http.StatusBadRequest, "invalid pagination parameters")
		case errors.Is(err, groups.ErrClientNotConnected):
			respondError(w, http.StatusServiceUnavailable, "whatsapp client not connected")
		case errors.Is(err, instances.ErrInstanceNotFound):
			respondError(w, http.StatusNotFound, "instance not found")
		default:
			logger.Error("failed to list groups",
				slog.String("error", err.Error()))
			captureHandlerError("groups", "GET /groups", instanceID.String(), err)
			respondError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	w.Header().Set("X-Total-Count", strconv.Itoa(result.Total))
	respondJSON(w, http.StatusOK, result.Items)

	logger.Info("groups listed successfully",
		slog.Int("returned_groups", len(result.Items)),
		slog.Int("total_groups", result.Total),
		slog.Int("page", page),
		slog.Int("page_size", pageSize))
}

func (h *GroupsHandler) createGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operation := "create"
	status := "success"
	observe := h.startMetrics(operation)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "groups_http_handler"),
		slog.String("operation", "POST /create-group"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		AutoInvite bool     `json:"autoInvite"`
		GroupName  string   `json:"groupName"`
		Phones     []string `json:"phones"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid create-group payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Create(ctx, instanceID, groups.CreateParams{
		AutoInvite: payload.AutoInvite,
		GroupName:  payload.GroupName,
		Phones:     payload.Phones,
	})
	if err != nil {
		status = "error"
		h.handleGroupError(logger, w, "POST /create-group", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)

	logger.Info("group created via handler",
		slog.String("phone", result.Phone),
		slog.Int("requested_participants", len(payload.Phones)),
		slog.Bool("auto_invite", payload.AutoInvite))
}

func (h *GroupsHandler) updateGroupName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operation := "update_name"
	status := "success"
	observe := h.startMetrics(operation)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "groups_http_handler"),
		slog.String("operation", "POST /update-group-name"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		GroupID   string `json:"groupId"`
		GroupName string `json:"groupName"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid update-group-name payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.UpdateName(ctx, instanceID, groups.UpdateNameParams{
		GroupID:   payload.GroupID,
		GroupName: payload.GroupName,
	})
	if err != nil {
		status = "error"
		h.handleGroupError(logger, w, "POST /update-group-name", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("group name updated via handler",
		slog.String("group_id", payload.GroupID))
}

func (h *GroupsHandler) updateGroupPhoto(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operation := "update_photo"
	status := "success"
	observe := h.startMetrics(operation)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "groups_http_handler"),
		slog.String("operation", "POST /update-group-photo"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		GroupID    string `json:"groupId"`
		GroupPhoto string `json:"groupPhoto"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid update-group-photo payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.UpdatePhoto(ctx, instanceID, groups.UpdatePhotoParams{
		GroupID:    payload.GroupID,
		GroupPhoto: payload.GroupPhoto,
	})
	if err != nil {
		status = "error"
		h.handleGroupError(logger, w, "POST /update-group-photo", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("group photo updated via handler",
		slog.String("group_id", payload.GroupID))
}

func (h *GroupsHandler) addParticipant(w http.ResponseWriter, r *http.Request) {
	h.handleParticipants(w, r, "add_participant", "POST /add-participant", true, h.service.AddParticipants)
}

func (h *GroupsHandler) removeParticipant(w http.ResponseWriter, r *http.Request) {
	h.handleParticipants(w, r, "remove_participant", "POST /remove-participant", false, h.service.RemoveParticipants)
}

func (h *GroupsHandler) approveParticipant(w http.ResponseWriter, r *http.Request) {
	h.handleParticipants(w, r, "approve_participant", "POST /approve-participant", false, h.service.ApproveParticipants)
}

func (h *GroupsHandler) rejectParticipant(w http.ResponseWriter, r *http.Request) {
	h.handleParticipants(w, r, "reject_participant", "POST /reject-participant", false, h.service.RejectParticipants)
}

func (h *GroupsHandler) addAdmin(w http.ResponseWriter, r *http.Request) {
	h.handleParticipants(w, r, "add_admin", "POST /add-admin", false, h.service.AddAdmins)
}

func (h *GroupsHandler) removeAdmin(w http.ResponseWriter, r *http.Request) {
	h.handleParticipants(w, r, "remove_admin", "POST /remove-admin", false, h.service.RemoveAdmins)
}

func (h *GroupsHandler) leaveGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operation := "leave"
	status := "success"
	observe := h.startMetrics(operation)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "groups_http_handler"),
		slog.String("operation", "POST /leave-group"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		GroupID string `json:"groupId"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid leave-group payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Leave(ctx, instanceID, groups.ModifyParticipantsParams{
		GroupID: payload.GroupID,
	})
	if err != nil {
		status = "error"
		h.handleGroupError(logger, w, "POST /leave-group", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("left group via handler",
		slog.String("group_id", payload.GroupID))
}

func (h *GroupsHandler) groupMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operation := "metadata"
	status := "success"
	observe := h.startMetrics(operation)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	groupID := chi.URLParam(r, "groupId")

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "groups_http_handler"),
		slog.String("operation", "GET /group-metadata/{groupId}"),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", groupID))
	logger := logging.ContextLogger(ctx, h.log)

	result, err := h.service.Metadata(ctx, instanceID, groupID)
	if err != nil {
		status = "error"
		h.handleGroupError(logger, w, "GET /group-metadata/{groupId}", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("group metadata served",
		slog.String("group_id", groupID),
		slog.Int("participants", len(result.Participants)))
}

func (h *GroupsHandler) lightGroupMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operation := "light_metadata"
	status := "success"
	observe := h.startMetrics(operation)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	groupID := chi.URLParam(r, "groupId")

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "groups_http_handler"),
		slog.String("operation", "GET /light-group-metadata/{groupId}"),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", groupID))
	logger := logging.ContextLogger(ctx, h.log)

	result, err := h.service.LightMetadata(ctx, instanceID, groupID)
	if err != nil {
		status = "error"
		h.handleGroupError(logger, w, "GET /light-group-metadata/{groupId}", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("light group metadata served",
		slog.String("group_id", groupID),
		slog.Int("participants", len(result.Participants)))
}

func (h *GroupsHandler) groupInvitationMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operation := "invitation_metadata"
	status := "success"
	observe := h.startMetrics(operation)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	inviteURL := strings.TrimSpace(r.URL.Query().Get("url"))
	ctx = logging.WithAttrs(ctx,
		slog.String("component", "groups_http_handler"),
		slog.String("operation", "GET /group-invitation-metadata"),
		slog.String("instance_id", instanceID.String()),
		slog.String("invite_url", inviteURL))
	logger := logging.ContextLogger(ctx, h.log)

	if inviteURL == "" {
		status = "error"
		logger.Warn("missing invite url parameter")
		respondError(w, http.StatusBadRequest, "url query parameter is required")
		return
	}

	result, err := h.service.InvitationMetadata(ctx, instanceID, inviteURL)
	if err != nil {
		status = "error"
		h.handleGroupError(logger, w, "GET /group-invitation-metadata", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("group invitation metadata served",
		slog.Int("participants", len(result.Participants)))
}

func (h *GroupsHandler) groupInvitationLink(w http.ResponseWriter, r *http.Request) {
	h.handleInviteLink(w, r, "group_invitation_link", "POST /group-invitation-link/{groupId}", false)
}

func (h *GroupsHandler) redefineGroupInvitationLink(w http.ResponseWriter, r *http.Request) {
	h.handleInviteLink(w, r, "redefine_invitation_link", "POST /redefine-invitation-link/{groupId}", true)
}

func (h *GroupsHandler) handleInviteLink(
	w http.ResponseWriter,
	r *http.Request,
	operation string,
	httpOperation string,
	reset bool,
) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics(operation)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	groupID := chi.URLParam(r, "groupId")
	ctx = logging.WithAttrs(ctx,
		slog.String("component", "groups_http_handler"),
		slog.String("operation", httpOperation),
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", groupID))
	logger := logging.ContextLogger(ctx, h.log)

	var (
		result groups.InvitationLinkResult
		err    error
	)
	if reset {
		result, err = h.service.RedefineInvitationLink(ctx, instanceID, groupID)
	} else {
		result, err = h.service.InvitationLink(ctx, instanceID, groupID)
	}
	if err != nil {
		status = "error"
		h.handleGroupError(logger, w, httpOperation, instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("group invitation link processed",
		slog.String("group_id", groupID),
		slog.Bool("reset", reset))
}

func (h *GroupsHandler) updateGroupSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operation := "update_settings"
	status := "success"
	observe := h.startMetrics(operation)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "groups_http_handler"),
		slog.String("operation", "POST /update-group-settings"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload groups.UpdateSettingsParams
	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid update-group-settings payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.UpdateSettings(ctx, instanceID, payload)
	if err != nil {
		status = "error"
		h.handleGroupError(logger, w, "POST /update-group-settings", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("group settings updated via handler",
		slog.String("group_id", payload.Phone),
		slog.Bool("admin_only_message", payload.AdminOnlyMessage),
		slog.Bool("admin_only_settings", payload.AdminOnlySettings),
		slog.Bool("require_admin_approval", payload.RequireAdminApproval),
		slog.Bool("admin_only_add_member", payload.AdminOnlyAddMember))
}

func (h *GroupsHandler) updateGroupDescription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operation := "update_description"
	status := "success"
	observe := h.startMetrics(operation)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "groups_http_handler"),
		slog.String("operation", "POST /update-group-description"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload groups.UpdateDescriptionParams
	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid update-group-description payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.UpdateDescription(ctx, instanceID, payload)
	if err != nil {
		status = "error"
		h.handleGroupError(logger, w, "POST /update-group-description", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("group description updated via handler",
		slog.String("group_id", payload.GroupID))
}

func (h *GroupsHandler) acceptInviteGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operation := "accept_invite"
	status := "success"
	observe := h.startMetrics(operation)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	inviteURL := strings.TrimSpace(r.URL.Query().Get("url"))

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "groups_http_handler"),
		slog.String("operation", "GET /accept-invite-group"),
		slog.String("instance_id", instanceID.String()),
		slog.String("invite_url", inviteURL))
	logger := logging.ContextLogger(ctx, h.log)

	if inviteURL == "" {
		status = "error"
		logger.Warn("missing invite url parameter")
		respondError(w, http.StatusBadRequest, "url query parameter is required")
		return
	}

	result, err := h.service.AcceptInvite(ctx, instanceID, inviteURL)
	if err != nil {
		status = "error"
		h.handleGroupError(logger, w, "GET /accept-invite-group", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("group invite accepted via handler")
}

func (h *GroupsHandler) resolveInstance(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, uuid.UUID, *instances.Status, bool) {
	instanceIDStr := chi.URLParam(r, "instanceId")
	instanceID, err := uuid.Parse(instanceIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid instance ID format")
		return ctx, uuid.UUID{}, nil, false
	}

	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))

	if h.instanceService == nil {
		respondError(w, http.StatusInternalServerError, "Instance service unavailable")
		return ctx, uuid.UUID{}, nil, false
	}

	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	status, err := h.instanceService.GetStatus(ctx, instanceID, clientToken, instanceToken)
	if err != nil {
		h.handleInstanceServiceError(ctx, w, err)
		return ctx, uuid.UUID{}, nil, false
	}

	return ctx, instanceID, status, true
}

func (h *GroupsHandler) handleInstanceServiceError(ctx context.Context, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, instances.ErrInstanceNotFound):
		respondError(w, http.StatusNotFound, "instance not found")
	case errors.Is(err, instances.ErrUnauthorized):
		respondError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, instances.ErrInstanceInactive):
		respondError(w, http.StatusForbidden, "instance subscription inactive")
	default:
		logging.ContextLogger(ctx, h.log).Error("instance service error",
			slog.String("error", err.Error()))
		captureHandlerError("groups", "instance_status", "", err)
		respondError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *GroupsHandler) notImplemented(operation string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.logNotImplemented(r.Context(), operation)
		respondError(w, http.StatusNotImplemented, "not implemented")
	}
}

func (h *GroupsHandler) logNotImplemented(ctx context.Context, operation string) {
	logger := logging.ContextLogger(ctx, h.log)
	logger.Warn("endpoint not implemented",
		slog.String("component", "groups_http_handler"),
		slog.String("operation", operation))
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
