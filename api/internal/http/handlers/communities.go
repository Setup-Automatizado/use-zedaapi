package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/communities"
	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// CommunitiesService defines the expected behaviour of the communities domain service.
type CommunitiesService interface {
	List(ctx context.Context, instanceID uuid.UUID, params communities.ListParams) (communities.ListResult, error)
	Create(ctx context.Context, instanceID uuid.UUID, params communities.CreateParams) (communities.CreateResult, error)
	Link(ctx context.Context, instanceID uuid.UUID, params communities.LinkParams) (communities.OperationResult, error)
	Unlink(ctx context.Context, instanceID uuid.UUID, params communities.LinkParams) (communities.OperationResult, error)
	Metadata(ctx context.Context, instanceID uuid.UUID, communityID string) (communities.Metadata, error)
	RegenerateInvitationLink(ctx context.Context, instanceID uuid.UUID, communityID string) (communities.InvitationResult, error)
	UpdateSettings(ctx context.Context, instanceID uuid.UUID, params communities.SettingsParams) (communities.OperationResult, error)
	UpdateDescription(ctx context.Context, instanceID uuid.UUID, params communities.UpdateDescriptionParams) (communities.OperationResult, error)
	ResolveAnnouncementGroup(ctx context.Context, instanceID uuid.UUID, communityID string) (string, error)
	Delete(ctx context.Context, instanceID uuid.UUID, communityID string) (communities.OperationResult, error)
	AddParticipants(ctx context.Context, instanceID uuid.UUID, params communities.ParticipantsParams) (communities.OperationResult, error)
	RemoveParticipants(ctx context.Context, instanceID uuid.UUID, params communities.ParticipantsParams) (communities.OperationResult, error)
	AddAdmins(ctx context.Context, instanceID uuid.UUID, params communities.ParticipantsParams) (communities.OperationResult, error)
	RemoveAdmins(ctx context.Context, instanceID uuid.UUID, params communities.ParticipantsParams) (communities.OperationResult, error)
}

// CommunitiesHandler exposes community-related endpoints matching the Zé da API surface.
type CommunitiesHandler struct {
	instanceService InstanceStatusProvider
	service         CommunitiesService
	metrics         *observability.Metrics
	log             *slog.Logger
}

// NewCommunitiesHandler wires the handler with the required dependencies.
func NewCommunitiesHandler(
	instanceService InstanceStatusProvider,
	service CommunitiesService,
	metrics *observability.Metrics,
	log *slog.Logger,
) *CommunitiesHandler {
	var handlerLogger *slog.Logger
	if log != nil {
		handlerLogger = log.With(slog.String("handler", "communities"))
	}
	return &CommunitiesHandler{
		instanceService: instanceService,
		service:         service,
		metrics:         metrics,
		log:             handlerLogger,
	}
}

// RegisterRoutes attaches all community-specific routes.
func (h *CommunitiesHandler) RegisterRoutes(r chi.Router) {
	r.Post("/communities", h.createCommunity)
	r.Get("/communities", h.listCommunities)
	r.Post("/communities/link", h.linkGroups)
	r.Post("/communities/unlink", h.unlinkGroups)
	r.Get("/communities-metadata/{communityId}", h.communityMetadata)
	r.Post("/communities/settings", h.updateCommunitySettings)
	r.Post("/update-community-description", h.updateCommunityDescription)
	r.Delete("/communities/{communityId}", h.deleteCommunity)
	r.Post("/redefine-invitation-link/{communityId}", h.redefineInvitationLink)
}

func (h *CommunitiesHandler) startMetrics(operation string) func(string) {
	start := time.Now()
	return func(status string) {
		if h.metrics == nil {
			return
		}
		if h.metrics.CommunitiesLatency != nil {
			h.metrics.CommunitiesLatency.WithLabelValues(operation).Observe(time.Since(start).Seconds())
		}
		if h.metrics.CommunitiesRequests != nil {
			h.metrics.CommunitiesRequests.WithLabelValues(operation, status).Inc()
		}
	}
}

func (h *CommunitiesHandler) listCommunities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("list")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "communities_http_handler"),
		slog.String("operation", "GET /communities"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	page, pageSize, valid := h.parsePagination(r, logger)
	if !valid {
		status = "error"
		respondError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	result, err := h.service.List(ctx, instanceID, communities.ListParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		status = "error"
		switch {
		case errors.Is(err, communities.ErrInvalidPagination):
			respondError(w, http.StatusBadRequest, "invalid pagination parameters")
		case errors.Is(err, communities.ErrClientNotConnected):
			respondError(w, http.StatusServiceUnavailable, "whatsapp client not connected")
		case errors.Is(err, instances.ErrInstanceNotFound):
			respondError(w, http.StatusNotFound, "instance not found")
		default:
			logger.Error("failed to list communities",
				slog.String("error", err.Error()))
			captureHandlerError("communities", "GET /communities", instanceID.String(), err)
			respondError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	w.Header().Set("X-Total-Count", strconv.Itoa(result.Total))
	respondJSON(w, http.StatusOK, result.Items)

	logger.Info("communities listed successfully",
		slog.Int("returned_communities", len(result.Items)),
		slog.Int("total_communities", result.Total),
		slog.Int("page", page),
		slog.Int("page_size", pageSize))
}

func (h *CommunitiesHandler) parsePagination(r *http.Request, logger *slog.Logger) (int, int, bool) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := 1
	pageSize := 50
	var err error

	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			logger.Warn("invalid page parameter",
				slog.String("value", pageStr),
				slog.String("error", errString(err)))
			return 0, 0, false
		}
	}

	if pageSizeStr != "" {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize <= 0 {
			logger.Warn("invalid pageSize parameter",
				slog.String("value", pageSizeStr),
				slog.String("error", errString(err)))
			return 0, 0, false
		}
	}

	return page, pageSize, true
}

func (h *CommunitiesHandler) createCommunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("create")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "communities_http_handler"),
		slog.String("operation", "POST /communities"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid create-community payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Create(ctx, instanceID, communities.CreateParams{
		Name:        payload.Name,
		Description: payload.Description,
	})
	if err != nil {
		status = "error"
		h.handleCommunityError(logger, w, "POST /communities", instanceID, err)
		return
	}

	maskedCommunityID := result.ID
	respondJSON(w, http.StatusOK, result)

	logger.Info("community created via handler",
		slog.String("community_id", maskedCommunityID))
}

func (h *CommunitiesHandler) linkGroups(w http.ResponseWriter, r *http.Request) {
	h.handleLinkOperation(w, r, "link", "POST /communities/link", h.service.Link)
}

func (h *CommunitiesHandler) unlinkGroups(w http.ResponseWriter, r *http.Request) {
	h.handleLinkOperation(w, r, "unlink", "POST /communities/unlink", h.service.Unlink)
}

func (h *CommunitiesHandler) handleLinkOperation(
	w http.ResponseWriter,
	r *http.Request,
	metric string,
	operation string,
	action func(context.Context, uuid.UUID, communities.LinkParams) (communities.OperationResult, error),
) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics(metric)
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "communities_http_handler"),
		slog.String("operation", operation),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		CommunityID  string   `json:"communityId"`
		GroupsPhones []string `json:"groupsPhones"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid community link payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	maskedCommunityID := payload.CommunityID
	ctx = logging.WithAttrs(ctx, slog.String("community_id", maskedCommunityID))
	logger = logging.ContextLogger(ctx, h.log)

	result, err := action(ctx, instanceID, communities.LinkParams{
		CommunityID: payload.CommunityID,
		GroupIDs:    payload.GroupsPhones,
	})
	if err != nil {
		status = "error"
		h.handleCommunityError(logger, w, operation, instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)

	logger.Info("community link operation completed",
		slog.String("community_id", maskedCommunityID),
		slog.Int("groups_count", len(payload.GroupsPhones)))
}

func (h *CommunitiesHandler) communityMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("metadata")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	communityID := chi.URLParam(r, "communityId")
	maskedCommunityID := communityID

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "communities_http_handler"),
		slog.String("operation", "GET /communities-metadata/{communityId}"),
		slog.String("instance_id", instanceID.String()),
		slog.String("community_id", maskedCommunityID))
	logger := logging.ContextLogger(ctx, h.log)

	result, err := h.service.Metadata(ctx, instanceID, communityID)
	if err != nil {
		status = "error"
		h.handleCommunityError(logger, w, "GET /communities-metadata/{communityId}", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)

	logger.Info("community metadata served",
		slog.String("community_id", maskedCommunityID))
}

func (h *CommunitiesHandler) redefineInvitationLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("redefine_invitation_link")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	communityID := chi.URLParam(r, "communityId")
	maskedCommunityID := communityID

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "communities_http_handler"),
		slog.String("operation", "POST /redefine-invitation-link/{communityId}"),
		slog.String("instance_id", instanceID.String()),
		slog.String("community_id", maskedCommunityID))
	logger := logging.ContextLogger(ctx, h.log)

	result, err := h.service.RegenerateInvitationLink(ctx, instanceID, communityID)
	if err != nil {
		status = "error"
		h.handleCommunityError(logger, w, "POST /redefine-invitation-link/{communityId}", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)

	logger.Info("community invitation link redefined",
		slog.String("community_id", maskedCommunityID))
}

func (h *CommunitiesHandler) updateCommunitySettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("settings")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "communities_http_handler"),
		slog.String("operation", "POST /communities/settings"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		CommunityID        string `json:"communityId"`
		WhoCanAddNewGroups string `json:"whoCanAddNewGroups"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid community settings payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	maskedCommunityID := payload.CommunityID
	ctx = logging.WithAttrs(ctx, slog.String("community_id", maskedCommunityID))
	logger = logging.ContextLogger(ctx, h.log)

	result, err := h.service.UpdateSettings(ctx, instanceID, communities.SettingsParams{
		CommunityID:        payload.CommunityID,
		WhoCanAddNewGroups: payload.WhoCanAddNewGroups,
	})
	if err != nil {
		status = "error"
		h.handleCommunityError(logger, w, "POST /communities/settings", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)

	logger.Info("community settings updated",
		slog.String("community_id", maskedCommunityID),
		slog.String("who_can_add_new_groups", payload.WhoCanAddNewGroups))
}

func (h *CommunitiesHandler) updateCommunityDescription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("update_description")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "communities_http_handler"),
		slog.String("operation", "POST /update-community-description"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		CommunityID          string `json:"communityId"`
		CommunityDescription string `json:"communityDescription"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid community description payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	maskedCommunityID := payload.CommunityID
	ctx = logging.WithAttrs(ctx, slog.String("community_id", maskedCommunityID))
	logger = logging.ContextLogger(ctx, h.log)

	result, err := h.service.UpdateDescription(ctx, instanceID, communities.UpdateDescriptionParams{
		CommunityID: payload.CommunityID,
		Description: payload.CommunityDescription,
	})
	if err != nil {
		status = "error"
		h.handleCommunityError(logger, w, "POST /update-community-description", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)

	logger.Info("community description updated",
		slog.String("community_id", maskedCommunityID))
}

func (h *CommunitiesHandler) deleteCommunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("delete")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	communityID := chi.URLParam(r, "communityId")
	maskedCommunityID := communityID

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "communities_http_handler"),
		slog.String("operation", "DELETE /communities/{communityId}"),
		slog.String("instance_id", instanceID.String()),
		slog.String("community_id", maskedCommunityID))
	logger := logging.ContextLogger(ctx, h.log)

	result, err := h.service.Delete(ctx, instanceID, communityID)
	if err != nil {
		status = "error"
		h.handleCommunityError(logger, w, "DELETE /communities/{communityId}", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)

	logger.Info("community deleted",
		slog.String("community_id", maskedCommunityID))
}

func (h *CommunitiesHandler) resolveInstance(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, uuid.UUID, *instances.Status, bool) {
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

func (h *CommunitiesHandler) handleInstanceServiceError(ctx context.Context, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, instances.ErrInstanceNotFound):
		respondError(w, http.StatusNotFound, "instance not found")
	case errors.Is(err, instances.ErrUnauthorized):
		respondError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, instances.ErrInstanceInactive):
		respondError(w, http.StatusForbidden, "instance subscription inactive")
	default:
		logger := logging.ContextLogger(ctx, h.log)
		logger.Error("instance service failure",
			slog.String("error", err.Error()))
		captureHandlerError("communities", "instance_status", "", err)
		respondError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *CommunitiesHandler) handleCommunityError(logger *slog.Logger, w http.ResponseWriter, operation string, instanceID uuid.UUID, err error) {
	switch {
	case errors.Is(err, communities.ErrInvalidPagination):
		respondError(w, http.StatusBadRequest, "invalid pagination parameters")
	case errors.Is(err, communities.ErrInvalidCommunityID):
		respondError(w, http.StatusBadRequest, "invalid community id")
	case errors.Is(err, communities.ErrInvalidCommunityName):
		respondError(w, http.StatusBadRequest, "invalid community name")
	case errors.Is(err, communities.ErrInvalidGroupList):
		respondError(w, http.StatusBadRequest, "invalid groups list")
	case errors.Is(err, communities.ErrInvalidPhoneList):
		respondError(w, http.StatusBadRequest, "invalid phone list")
	case errors.Is(err, communities.ErrInvalidCommunityDescription):
		respondError(w, http.StatusBadRequest, "invalid community description")
	case errors.Is(err, communities.ErrClientNotConnected):
		respondError(w, http.StatusServiceUnavailable, "whatsapp client not connected")
	case errors.Is(err, instances.ErrInstanceNotFound):
		respondError(w, http.StatusNotFound, "instance not found")
	default:
		logger.Error("community operation failed",
			slog.String("error", err.Error()))
		captureHandlerError("communities", operation, instanceID.String(), err)
		respondError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *CommunitiesHandler) notImplemented(operation string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.logNotImplemented(r.Context(), operation)
		respondError(w, http.StatusNotImplemented, "not implemented")
	}
}

func (h *CommunitiesHandler) logNotImplemented(ctx context.Context, operation string) {
	logger := logging.ContextLogger(ctx, h.log)
	logger.Warn("endpoint not implemented",
		slog.String("component", "communities_http_handler"),
		slog.String("operation", operation))
}
