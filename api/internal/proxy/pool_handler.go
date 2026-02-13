package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"go.mau.fi/whatsmeow/api/internal/logging"
)

// PoolHandler handles HTTP requests for proxy pool management.
type PoolHandler struct {
	service *PoolService
	log     *slog.Logger
}

// NewPoolHandler creates a new PoolHandler.
func NewPoolHandler(service *PoolService, log *slog.Logger) *PoolHandler {
	return &PoolHandler{
		service: service,
		log:     log.With(slog.String("component", "pool_handler")),
	}
}

// RegisterPartnerRoutes registers partner-auth routes for pool management.
func (h *PoolHandler) RegisterPartnerRoutes(r chi.Router) {
	r.Route("/proxy-providers", func(r chi.Router) {
		r.Post("/", h.createProvider)
		r.Get("/", h.listProviders)
		r.Get("/{id}", h.getProvider)
		r.Put("/{id}", h.updateProvider)
		r.Delete("/{id}", h.deleteProvider)
		r.Post("/{id}/sync", h.triggerSync)
	})
	r.Get("/proxy-pool/stats", h.getPoolStats)
	r.Get("/proxy-pool", h.listPoolProxies)
	r.Post("/proxy-pool/bulk-assign", h.bulkAssign)
	r.Route("/proxy-groups", func(r chi.Router) {
		r.Post("/", h.createGroup)
		r.Get("/", h.listGroups)
		r.Delete("/{id}", h.deleteGroup)
	})
}

// RegisterInstanceRoutes registers instance-level routes for pool proxy management.
func (h *PoolHandler) RegisterInstanceRoutes(r chi.Router) {
	r.Post("/proxy/pool/assign", h.assignPoolProxy)
	r.Delete("/proxy/pool/release", h.releasePoolProxy)
	r.Get("/proxy/pool/assignment", h.getPoolAssignment)
	r.Post("/proxy/pool/assign-group", h.assignToGroup)
}

// ---------------------------------------------------------------------------
// Partner routes
// ---------------------------------------------------------------------------

func (h *PoolHandler) createProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	result, err := h.service.CreateProvider(ctx, req)
	if err != nil {
		h.handleError(ctx, w, err)
		return
	}
	h.respondJSON(w, http.StatusCreated, result)
}

func (h *PoolHandler) listProviders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	providers, err := h.service.ListProviders(ctx)
	if err != nil {
		h.handleError(ctx, w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, providers)
}

func (h *PoolHandler) getProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := h.parseUUID(w, r, "id")
	if !ok {
		return
	}
	provider, err := h.service.GetProvider(ctx, id)
	if err != nil {
		h.handleError(ctx, w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, provider)
}

func (h *PoolHandler) updateProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := h.parseUUID(w, r, "id")
	if !ok {
		return
	}
	var req UpdateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	result, err := h.service.UpdateProvider(ctx, id, req)
	if err != nil {
		h.handleError(ctx, w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, result)
}

func (h *PoolHandler) deleteProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := h.parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteProvider(ctx, id); err != nil {
		h.handleError(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PoolHandler) triggerSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := h.parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.service.TriggerSync(ctx, id); err != nil {
		h.handleError(ctx, w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "sync_started"})
}

func (h *PoolHandler) getPoolStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats, err := h.service.GetPoolStats(ctx)
	if err != nil {
		h.handleError(ctx, w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, stats)
}

func (h *PoolHandler) listPoolProxies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var providerID *uuid.UUID
	if pidStr := r.URL.Query().Get("provider_id"); pidStr != "" {
		parsed, err := uuid.Parse(pidStr)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid provider_id")
			return
		}
		providerID = &parsed
	}

	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	proxies, total, err := h.service.ListPoolProxies(ctx, providerID, status, limit, offset)
	if err != nil {
		h.handleError(ctx, w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data":  proxies,
		"total": total,
	})
}

func (h *PoolHandler) createGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body struct {
		Name         string     `json:"name"`
		ProviderID   *uuid.UUID `json:"providerId,omitempty"`
		MaxInstances int        `json:"maxInstances"`
		CountryCode  *string    `json:"countryCode,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	result, err := h.service.CreateGroup(ctx, body.Name, body.ProviderID, body.MaxInstances, body.CountryCode)
	if err != nil {
		h.handleError(ctx, w, err)
		return
	}
	h.respondJSON(w, http.StatusCreated, result)
}

func (h *PoolHandler) listGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	groups, err := h.service.ListGroups(ctx)
	if err != nil {
		h.handleError(ctx, w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, groups)
}

func (h *PoolHandler) deleteGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := h.parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteGroup(ctx, id); err != nil {
		h.handleError(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PoolHandler) bulkAssign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req BulkAssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	result, err := h.service.BulkAssignPoolProxies(ctx, req)
	if err != nil {
		h.handleError(ctx, w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, result)
}

// ---------------------------------------------------------------------------
// Instance routes
// ---------------------------------------------------------------------------

func (h *PoolHandler) assignPoolProxy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, err := uuid.Parse(chi.URLParam(r, "instanceId"))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid instance id")
		return
	}
	var req AssignPoolProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	result, err := h.service.AssignPoolProxy(ctx, instanceID, req)
	if err != nil {
		h.handleError(ctx, w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, result)
}

func (h *PoolHandler) releasePoolProxy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, err := uuid.Parse(chi.URLParam(r, "instanceId"))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid instance id")
		return
	}
	if err := h.service.ReleasePoolProxy(ctx, instanceID); err != nil {
		h.handleError(ctx, w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "released"})
}

func (h *PoolHandler) getPoolAssignment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, err := uuid.Parse(chi.URLParam(r, "instanceId"))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid instance id")
		return
	}
	assignment, err := h.service.GetPoolAssignment(ctx, instanceID)
	if err != nil {
		h.handleError(ctx, w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, assignment)
}

func (h *PoolHandler) assignToGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, err := uuid.Parse(chi.URLParam(r, "instanceId"))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid instance id")
		return
	}
	var req AssignGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	result, err := h.service.AssignToGroup(ctx, instanceID, req.GroupID)
	if err != nil {
		h.handleError(ctx, w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, result)
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

func (h *PoolHandler) respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *PoolHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

func (h *PoolHandler) handleError(ctx context.Context, w http.ResponseWriter, err error) {
	logger := logging.ContextLogger(ctx, h.log)
	logger.Error("pool handler error", slog.String("error", err.Error()))

	// Differentiate error types
	if errors.Is(err, pgx.ErrNoRows) {
		h.respondError(w, http.StatusNotFound, "resource not found")
		return
	}

	// Capture unexpected errors in Sentry
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("component", "pool_handler")
		sentry.CaptureException(err)
	})

	h.respondError(w, http.StatusInternalServerError, "internal error")
}

func (h *PoolHandler) parseUUID(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid "+param)
		return uuid.UUID{}, false
	}
	return id, true
}
