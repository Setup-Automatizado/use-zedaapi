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

	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/newsletters"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// NewslettersService defines the expected behaviour of the newsletters domain service.
type NewslettersService interface {
	List(ctx context.Context, instanceID uuid.UUID, params newsletters.ListParams) (newsletters.ListResult, error)
	Create(ctx context.Context, instanceID uuid.UUID, params newsletters.CreateParams) (newsletters.CreateResult, error)
	UpdatePicture(ctx context.Context, instanceID uuid.UUID, params newsletters.UpdatePictureParams) (newsletters.OperationResult, error)
	UpdateName(ctx context.Context, instanceID uuid.UUID, params newsletters.UpdateNameParams) (newsletters.OperationResult, error)
	UpdateDescription(ctx context.Context, instanceID uuid.UUID, params newsletters.UpdateDescriptionParams) (newsletters.OperationResult, error)
	Follow(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error)
	Unfollow(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error)
	Mute(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error)
	Unmute(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error)
	Delete(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error)
	Metadata(ctx context.Context, instanceID uuid.UUID, id string) (newsletters.MetadataResult, error)
	Search(ctx context.Context, instanceID uuid.UUID, params newsletters.SearchParams) (newsletters.SearchResult, error)
	UpdateSettings(ctx context.Context, instanceID uuid.UUID, params newsletters.SettingsParams) (newsletters.OperationResult, error)
	SendAdminInvite(ctx context.Context, instanceID uuid.UUID, params newsletters.AdminActionParams) (newsletters.OperationResult, error)
	AcceptAdminInvite(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error)
	RemoveAdmin(ctx context.Context, instanceID uuid.UUID, params newsletters.AdminActionParams) (newsletters.OperationResult, error)
	RevokeAdminInvite(ctx context.Context, instanceID uuid.UUID, params newsletters.AdminActionParams) (newsletters.OperationResult, error)
	TransferOwnership(ctx context.Context, instanceID uuid.UUID, params newsletters.TransferOwnershipParams) (newsletters.OperationResult, error)
}

// NewslettersHandler exposes newsletter-related endpoints mirroring the Z-API surface.
type NewslettersHandler struct {
	instanceService InstanceStatusProvider
	service         NewslettersService
	metrics         *observability.Metrics
	log             *slog.Logger
}

// NewNewslettersHandler wires the newsletters handler with its dependencies.
func NewNewslettersHandler(
	instanceService InstanceStatusProvider,
	service NewslettersService,
	metrics *observability.Metrics,
	log *slog.Logger,
) *NewslettersHandler {
	var handlerLogger *slog.Logger
	if log != nil {
		handlerLogger = log.With(slog.String("handler", "newsletters"))
	}
	return &NewslettersHandler{
		instanceService: instanceService,
		service:         service,
		metrics:         metrics,
		log:             handlerLogger,
	}
}

// RegisterRoutes binds all newsletter routes that must exist under the
// instance/token prefix to maintain compatibility with the Z-API contract.
func (h *NewslettersHandler) RegisterRoutes(r chi.Router) {
	r.Post("/create-newsletter", h.createNewsletter)
	r.Post("/update-newsletter-picture", h.updateNewsletterPicture)
	r.Post("/update-newsletter-name", h.updateNewsletterName)
	r.Post("/update-newsletter-description", h.updateNewsletterDescription)
	r.Put("/follow-newsletter", h.followNewsletter)
	r.Put("/unfollow-newsletter", h.unfollowNewsletter)
	r.Put("/mute-newsletter", h.muteNewsletter)
	r.Put("/unmute-newsletter", h.unmuteNewsletter)
	r.Delete("/delete-newsletter", h.deleteNewsletter)
	r.Get("/newsletter/metadata/{newsletterId}", h.getNewsletterMetadata)
	r.Get("/newsletter", h.listNewsletters)
	r.Post("/search-newsletter", h.searchNewsletters)
	r.Post("/newsletter/settings/{newsletterId}", h.updateNewsletterSettings)
	r.Post("/newsletter/accept-admin-invite/{newsletterId}", h.acceptNewsletterAdminInvite)
	r.Post("/newsletter/remove-admin/{newsletterId}", h.removeNewsletterAdmin)
	r.Post("/newsletter/revoke-admin-invite/{newsletterId}", h.revokeNewsletterAdminInvite)
	r.Post("/newsletter/transfer-ownership/{newsletterId}", h.transferNewsletterOwnership)
	r.Post("/send-newsletter-admin-invite", h.sendNewsletterAdminInvite)
}

func (h *NewslettersHandler) startMetrics(operation string) func(string) {
	start := time.Now()
	return func(status string) {
		if h.metrics == nil {
			return
		}
		if h.metrics.NewslettersLatency != nil {
			h.metrics.NewslettersLatency.WithLabelValues(operation).Observe(time.Since(start).Seconds())
		}
		if h.metrics.NewslettersRequests != nil {
			h.metrics.NewslettersRequests.WithLabelValues(operation, status).Inc()
		}
	}
}

func (h *NewslettersHandler) listNewsletters(w http.ResponseWriter, r *http.Request) {
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
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "GET /newsletter"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	page, pageSize, valid := h.parsePagination(r, logger)
	if !valid {
		status = "error"
		respondError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	result, err := h.service.List(ctx, instanceID, newsletters.ListParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		status = "error"
		switch {
		case errors.Is(err, newsletters.ErrInvalidPagination):
			respondError(w, http.StatusBadRequest, "invalid pagination parameters")
		case errors.Is(err, newsletters.ErrClientNotConnected):
			respondError(w, http.StatusServiceUnavailable, "whatsapp client not connected")
		case errors.Is(err, instances.ErrInstanceNotFound):
			respondError(w, http.StatusNotFound, "instance not found")
		default:
			logger.Error("failed to list newsletters",
				slog.String("error", err.Error()))
			captureHandlerError("newsletters", "GET /newsletter", instanceID.String(), err)
			respondError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	w.Header().Set("X-Total-Count", strconv.Itoa(result.Total))
	respondJSON(w, http.StatusOK, result.Items)

	logger.Info("newsletters listed successfully",
		slog.Int("returned_newsletters", len(result.Items)),
		slog.Int("total_newsletters", result.Total),
		slog.Int("page", page),
		slog.Int("page_size", pageSize))
}

func (h *NewslettersHandler) createNewsletter(w http.ResponseWriter, r *http.Request) {
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
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "POST /create-newsletter"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid create-newsletter payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Create(ctx, instanceID, newsletters.CreateParams{
		Name:        payload.Name,
		Description: payload.Description,
	})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "POST /create-newsletter", instanceID, err)
		return
	}

	respondJSON(w, http.StatusCreated, result)

	logger.Info("newsletter created via handler",
		slog.String("newsletter_id", result.ID),
		slog.String("name", payload.Name))
}

func (h *NewslettersHandler) updateNewsletterPicture(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("update_picture")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "POST /update-newsletter-picture"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		ID         string `json:"id"`
		PictureURL string `json:"pictureUrl"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid update-newsletter-picture payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.UpdatePicture(ctx, instanceID, newsletters.UpdatePictureParams{
		ID:      payload.ID,
		Picture: payload.PictureURL,
	})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "POST /update-newsletter-picture", instanceID, err)
		return
	}

	respondJSON(w, http.StatusCreated, result)
	logger.Info("newsletter picture updated",
		slog.String("newsletter_id", payload.ID))
}

func (h *NewslettersHandler) updateNewsletterName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("update_name")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "POST /update-newsletter-name"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid update-newsletter-name payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.UpdateName(ctx, instanceID, newsletters.UpdateNameParams{
		ID:   payload.ID,
		Name: payload.Name,
	})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "POST /update-newsletter-name", instanceID, err)
		return
	}

	respondJSON(w, http.StatusCreated, result)
	logger.Info("newsletter name updated",
		slog.String("newsletter_id", payload.ID))
}

func (h *NewslettersHandler) updateNewsletterDescription(w http.ResponseWriter, r *http.Request) {
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
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "POST /update-newsletter-description"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		ID          string `json:"id"`
		Description string `json:"description"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid update-newsletter-description payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.UpdateDescription(ctx, instanceID, newsletters.UpdateDescriptionParams{
		ID:          payload.ID,
		Description: payload.Description,
	})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "POST /update-newsletter-description", instanceID, err)
		return
	}

	respondJSON(w, http.StatusCreated, result)
	logger.Info("newsletter description updated",
		slog.String("newsletter_id", payload.ID))
}

func (h *NewslettersHandler) followNewsletter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("follow")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "PUT /follow-newsletter"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		ID string `json:"id"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid follow-newsletter payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Follow(ctx, instanceID, newsletters.IDParams{ID: payload.ID})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "PUT /follow-newsletter", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("newsletter followed",
		slog.String("newsletter_id", payload.ID))
}

func (h *NewslettersHandler) unfollowNewsletter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("unfollow")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "PUT /unfollow-newsletter"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		ID string `json:"id"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid unfollow-newsletter payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Unfollow(ctx, instanceID, newsletters.IDParams{ID: payload.ID})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "PUT /unfollow-newsletter", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("newsletter unfollowed",
		slog.String("newsletter_id", payload.ID))
}

func (h *NewslettersHandler) muteNewsletter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("mute")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "PUT /mute-newsletter"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		ID string `json:"id"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid mute-newsletter payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Mute(ctx, instanceID, newsletters.IDParams{ID: payload.ID})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "PUT /mute-newsletter", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("newsletter muted",
		slog.String("newsletter_id", payload.ID))
}

func (h *NewslettersHandler) unmuteNewsletter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("unmute")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "PUT /unmute-newsletter"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		ID string `json:"id"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid unmute-newsletter payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Unmute(ctx, instanceID, newsletters.IDParams{ID: payload.ID})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "PUT /unmute-newsletter", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("newsletter unmuted",
		slog.String("newsletter_id", payload.ID))
}

func (h *NewslettersHandler) deleteNewsletter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("delete")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "DELETE /delete-newsletter"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		ID string `json:"id"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid delete-newsletter payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Delete(ctx, instanceID, newsletters.IDParams{ID: payload.ID})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "DELETE /delete-newsletter", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("newsletter deleted",
		slog.String("newsletter_id", payload.ID))
}

func (h *NewslettersHandler) getNewsletterMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("metadata")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	newsletterID := chi.URLParam(r, "newsletterId")

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "GET /newsletter/metadata"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", newsletterID))
	logger := logging.ContextLogger(ctx, h.log)

	result, err := h.service.Metadata(ctx, instanceID, newsletterID)
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "GET /newsletter/metadata", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("newsletter metadata returned",
		slog.String("newsletter_id", newsletterID))
}

func (h *NewslettersHandler) searchNewsletters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("search")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "POST /search-newsletter"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		Limit   int    `json:"limit"`
		View    string `json:"view"`
		Filters struct {
			CountryCodes []string `json:"countryCodes"`
		} `json:"filters"`
		SearchText *string `json:"searchText"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid search-newsletter payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Search(ctx, instanceID, newsletters.SearchParams{
		Limit:        payload.Limit,
		View:         payload.View,
		CountryCodes: payload.Filters.CountryCodes,
		SearchText:   payload.SearchText,
	})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "POST /search-newsletter", instanceID, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
	logger.Info("newsletter search handled",
		slog.Int("results", len(result.Data)))
}

func (h *NewslettersHandler) updateNewsletterSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("settings")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	newsletterID := chi.URLParam(r, "newsletterId")

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "POST /newsletter/settings"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", newsletterID))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		ReactionCodes string `json:"reactionCodes"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid newsletter/settings payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.UpdateSettings(ctx, instanceID, newsletters.SettingsParams{
		ID:            newsletterID,
		ReactionCodes: payload.ReactionCodes,
	})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "POST /newsletter/settings", instanceID, err)
		return
	}

	respondJSON(w, http.StatusCreated, result)
	logger.Info("newsletter settings updated",
		slog.String("newsletter_id", newsletterID))
}

func (h *NewslettersHandler) acceptNewsletterAdminInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("accept_admin_invite")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	newsletterID := chi.URLParam(r, "newsletterId")

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "POST /newsletter/accept-admin-invite"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", newsletterID))
	logger := logging.ContextLogger(ctx, h.log)

	result, err := h.service.AcceptAdminInvite(ctx, instanceID, newsletters.IDParams{ID: newsletterID})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "POST /newsletter/accept-admin-invite", instanceID, err)
		return
	}

	respondJSON(w, http.StatusCreated, result)
	logger.Info("newsletter admin invite accepted",
		slog.String("newsletter_id", newsletterID))
}

func (h *NewslettersHandler) removeNewsletterAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("remove_admin")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	newsletterID := chi.URLParam(r, "newsletterId")

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "POST /newsletter/remove-admin"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", newsletterID))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		Phone string `json:"phone"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid remove-admin payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.RemoveAdmin(ctx, instanceID, newsletters.AdminActionParams{
		ID:    newsletterID,
		Phone: payload.Phone,
	})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "POST /newsletter/remove-admin", instanceID, err)
		return
	}

	respondJSON(w, http.StatusCreated, result)
	logger.Info("newsletter admin removed",
		slog.String("newsletter_id", newsletterID),
		slog.String("phone", payload.Phone))
}

func (h *NewslettersHandler) revokeNewsletterAdminInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("revoke_admin_invite")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	newsletterID := chi.URLParam(r, "newsletterId")

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "POST /newsletter/revoke-admin-invite"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", newsletterID))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		Phone string `json:"phone"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid revoke-admin-invite payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.RevokeAdminInvite(ctx, instanceID, newsletters.AdminActionParams{
		ID:    newsletterID,
		Phone: payload.Phone,
	})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "POST /newsletter/revoke-admin-invite", instanceID, err)
		return
	}

	respondJSON(w, http.StatusCreated, result)
	logger.Info("newsletter admin invite revoked",
		slog.String("newsletter_id", newsletterID),
		slog.String("phone", payload.Phone))
}

func (h *NewslettersHandler) transferNewsletterOwnership(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("transfer_ownership")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	newsletterID := chi.URLParam(r, "newsletterId")

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "POST /newsletter/transfer-ownership"),
		slog.String("instance_id", instanceID.String()),
		slog.String("newsletter_id", newsletterID))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		Phone     string `json:"phone"`
		QuitAdmin bool   `json:"quitAdmin"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid transfer-ownership payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.TransferOwnership(ctx, instanceID, newsletters.TransferOwnershipParams{
		ID:        newsletterID,
		Phone:     payload.Phone,
		QuitAdmin: payload.QuitAdmin,
	})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "POST /newsletter/transfer-ownership", instanceID, err)
		return
	}

	respondJSON(w, http.StatusCreated, result)
	logger.Info("newsletter ownership transferred",
		slog.String("newsletter_id", newsletterID),
		slog.String("phone", payload.Phone),
		slog.Bool("quit_admin", payload.QuitAdmin))
}

func (h *NewslettersHandler) sendNewsletterAdminInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "success"
	observe := h.startMetrics("send_admin_invite")
	defer func() { observe(status) }()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		status = "error"
		return
	}

	ctx = logging.WithAttrs(ctx,
		slog.String("component", "newsletters_http_handler"),
		slog.String("operation", "POST /send-newsletter-admin-invite"),
		slog.String("instance_id", instanceID.String()))
	logger := logging.ContextLogger(ctx, h.log)

	var payload struct {
		ID    string `json:"id"`
		Phone string `json:"phone"`
	}

	if err := decodeRequest(r, &payload); err != nil {
		status = "error"
		logger.Warn("invalid send-admin-invite payload",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.SendAdminInvite(ctx, instanceID, newsletters.AdminActionParams{
		ID:    payload.ID,
		Phone: payload.Phone,
	})
	if err != nil {
		status = "error"
		h.handleNewsletterError(logger, w, "POST /send-newsletter-admin-invite", instanceID, err)
		return
	}

	respondJSON(w, http.StatusCreated, result)
	logger.Info("newsletter admin invite sent",
		slog.String("newsletter_id", payload.ID),
		slog.String("phone", payload.Phone))
}

func (h *NewslettersHandler) parsePagination(r *http.Request, logger *slog.Logger) (int, int, bool) {
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

func (h *NewslettersHandler) resolveInstance(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, uuid.UUID, *instances.Status, bool) {
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

func (h *NewslettersHandler) handleNewsletterError(logger *slog.Logger, w http.ResponseWriter, operation string, instanceID uuid.UUID, err error) {
	switch {
	case errors.Is(err, newsletters.ErrInvalidPagination):
		respondError(w, http.StatusBadRequest, "invalid pagination parameters")
	case errors.Is(err, newsletters.ErrInvalidName):
		respondError(w, http.StatusBadRequest, "invalid name")
	case errors.Is(err, newsletters.ErrInvalidNewsletterID):
		respondError(w, http.StatusBadRequest, "invalid newsletter id")
	case errors.Is(err, newsletters.ErrInvalidPicture):
		respondError(w, http.StatusBadRequest, "invalid picture")
	case errors.Is(err, newsletters.ErrInvalidReactionCodes):
		respondError(w, http.StatusBadRequest, "invalid reaction codes")
	case errors.Is(err, newsletters.ErrInvalidPhone):
		respondError(w, http.StatusBadRequest, "invalid phone")
	case errors.Is(err, newsletters.ErrClientNotConnected):
		respondError(w, http.StatusServiceUnavailable, "whatsapp client not connected")
	case errors.Is(err, instances.ErrUnauthorized):
		respondError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, instances.ErrInstanceNotFound):
		respondError(w, http.StatusNotFound, "instance not found")
	default:
		logger.Error("newsletter operation failed",
			slog.String("error", err.Error()))
		captureHandlerError("newsletters", operation, instanceID.String(), err)
		respondError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *NewslettersHandler) handleInstanceServiceError(ctx context.Context, w http.ResponseWriter, err error) {
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
		captureHandlerError("newsletters", "instance_status", "", err)
		respondError(w, http.StatusInternalServerError, "internal error")
	}
}
