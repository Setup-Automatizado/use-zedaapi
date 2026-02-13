package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/statuscache"
)

// StatusCacheHandler handles HTTP requests for status cache operations
type StatusCacheHandler struct {
	service statuscache.Service
	log     *slog.Logger
}

// NewStatusCacheHandler creates a new status cache handler
func NewStatusCacheHandler(service statuscache.Service, log *slog.Logger) *StatusCacheHandler {
	return &StatusCacheHandler{
		service: service,
		log:     log.With(slog.String("component", "status_cache_handler")),
	}
}

// RegisterRoutes registers status cache routes within an existing route group
func (h *StatusCacheHandler) RegisterRoutes(r chi.Router) {
	r.Get("/messages-status", h.getMessagesStatus)
	r.Post("/messages-status/flush", h.flushStatus)
	r.Delete("/messages-status/cache", h.clearCache)
	r.Get("/messages-status/stats", h.getStats)
}

// getMessagesStatus handles GET /messages-status
// Query params: messageId, groupId, phone, limit, offset, includeParticipants, format (aggregated|raw)
func (h *StatusCacheHandler) getMessagesStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))

	// Parse query parameters
	messageID := r.URL.Query().Get("messageId")
	groupID := r.URL.Query().Get("groupId")
	phone := r.URL.Query().Get("phone")
	format := r.URL.Query().Get("format")

	if format == "" {
		format = "aggregated"
	}

	// Parse pagination
	params := statuscache.DefaultQueryParams()
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if limit > 1000 {
				limit = 1000
			}
			params.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			params.Offset = offset
		}
	}
	if r.URL.Query().Get("includeParticipants") == "true" {
		params.IncludeParticipants = true
	}

	h.log.Debug("querying messages status",
		slog.String("instance_id", instanceID.String()),
		slog.String("message_id", messageID),
		slog.String("group_id", groupID),
		slog.String("phone", phone),
		slog.String("format", format),
	)

	// Determine query type based on parameters
	// If no filter provided, return ALL entries (paginated)
	switch {
	case messageID != "":
		h.handleMessageIDQuery(ctx, w, instanceID.String(), messageID, format, params.IncludeParticipants)
	case groupID != "":
		h.handleGroupQuery(ctx, w, instanceID.String(), groupID, format, params)
	case phone != "":
		h.handlePhoneQuery(ctx, w, instanceID.String(), phone, format, params)
	default:
		// No filter = return ALL entries with pagination
		h.handleAllQuery(ctx, w, instanceID.String(), format, params)
	}
}

func (h *StatusCacheHandler) handleMessageIDQuery(ctx context.Context, w http.ResponseWriter, instanceID, messageID, format string, includeParticipants bool) {
	if format == "raw" {
		result, err := h.service.GetRawStatus(ctx, instanceID, messageID)
		if err != nil {
			h.log.Error("failed to get raw status", slog.String("error", err.Error()))
			respondError(w, http.StatusInternalServerError, "failed to query status cache")
			return
		}
		if result == nil || len(result.Data) == 0 {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"data": []interface{}{},
				"meta": map[string]interface{}{"total": 0, "limit": 0, "offset": 0},
			})
			return
		}
		respondJSON(w, http.StatusOK, result)
		return
	}

	// Aggregated format
	status, err := h.service.GetStatus(ctx, instanceID, messageID, includeParticipants)
	if err != nil {
		h.log.Error("failed to get status", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to query status cache")
		return
	}
	if status == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"data": []interface{}{},
			"meta": map[string]interface{}{"total": 0, "limit": 0, "offset": 0},
		})
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": []*statuscache.AggregatedStatus{status},
		"meta": map[string]interface{}{"total": 1, "limit": 1, "offset": 0},
	})
}

func (h *StatusCacheHandler) handleGroupQuery(ctx context.Context, w http.ResponseWriter, instanceID, groupID, format string, params statuscache.QueryParams) {
	if format == "raw" {
		result, err := h.service.QueryRawByGroup(ctx, instanceID, groupID, params)
		if err != nil {
			h.log.Error("failed to query raw by group", slog.String("error", err.Error()))
			respondError(w, http.StatusInternalServerError, "failed to query status cache")
			return
		}
		respondJSON(w, http.StatusOK, result)
		return
	}

	// Aggregated format
	result, err := h.service.QueryByGroup(ctx, instanceID, groupID, params)
	if err != nil {
		h.log.Error("failed to query by group", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to query status cache")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (h *StatusCacheHandler) handlePhoneQuery(ctx context.Context, w http.ResponseWriter, instanceID, phone, format string, params statuscache.QueryParams) {
	if format == "raw" {
		result, err := h.service.QueryRawByPhone(ctx, instanceID, phone, params)
		if err != nil {
			h.log.Error("failed to query raw by phone", slog.String("error", err.Error()))
			respondError(w, http.StatusInternalServerError, "failed to query status cache")
			return
		}
		respondJSON(w, http.StatusOK, result)
		return
	}

	// Aggregated format
	result, err := h.service.QueryByPhone(ctx, instanceID, phone, params)
	if err != nil {
		h.log.Error("failed to query by phone", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to query status cache")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

// handleAllQuery handles queries without filters - returns ALL cached entries with pagination
// Use format=raw to get data in EXACT webhook payload format (RawStatusPayload)
func (h *StatusCacheHandler) handleAllQuery(ctx context.Context, w http.ResponseWriter, instanceID, format string, params statuscache.QueryParams) {
	h.log.Debug("querying all status cache entries",
		slog.String("instance_id", instanceID),
		slog.String("format", format),
		slog.Int("limit", params.Limit),
		slog.Int("offset", params.Offset),
	)

	if format == "raw" {
		// Raw format = EXACT webhook payload structure (RawStatusPayload)
		result, err := h.service.QueryRawAll(ctx, instanceID, params)
		if err != nil {
			h.log.Error("failed to query raw all", slog.String("error", err.Error()))
			respondError(w, http.StatusInternalServerError, "failed to query status cache")
			return
		}
		respondJSON(w, http.StatusOK, result)
		return
	}

	// Aggregated format (default)
	result, err := h.service.QueryAll(ctx, instanceID, params)
	if err != nil {
		h.log.Error("failed to query all", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to query status cache")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

// flushStatus handles POST /messages-status/flush
func (h *StatusCacheHandler) flushStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))

	var req statuscache.FlushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.log.Debug("flushing status cache",
		slog.String("instance_id", instanceID.String()),
		slog.String("message_id", req.MessageID),
		slog.Bool("all", req.All),
	)

	var result *statuscache.FlushResult
	var err error

	if req.MessageID != "" {
		result, err = h.service.FlushMessage(ctx, instanceID.String(), req.MessageID)
	} else if req.All {
		result, err = h.service.FlushAll(ctx, instanceID.String())
	} else {
		respondError(w, http.StatusBadRequest, "specify messageId or set all=true")
		return
	}

	if err != nil {
		h.log.Error("failed to flush status cache", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to flush status cache")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// clearCache handles DELETE /messages-status/cache
func (h *StatusCacheHandler) clearCache(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))

	messageID := r.URL.Query().Get("messageId")
	clearAll := r.URL.Query().Get("all") == "true"

	h.log.Debug("clearing status cache",
		slog.String("instance_id", instanceID.String()),
		slog.String("message_id", messageID),
		slog.Bool("all", clearAll),
	)

	if messageID != "" {
		if err := h.service.ClearMessage(ctx, instanceID.String(), messageID); err != nil {
			h.log.Error("failed to clear message from cache", slog.String("error", err.Error()))
			respondError(w, http.StatusInternalServerError, "failed to clear cache")
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"cleared":   1,
			"messageId": messageID,
		})
		return
	}

	if clearAll {
		count, err := h.service.ClearInstance(ctx, instanceID.String())
		if err != nil {
			h.log.Error("failed to clear instance cache", slog.String("error", err.Error()))
			respondError(w, http.StatusInternalServerError, "failed to clear cache")
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"cleared": count,
		})
		return
	}

	respondError(w, http.StatusBadRequest, "specify messageId or set all=true")
}

// getStats handles GET /messages-status/stats
func (h *StatusCacheHandler) getStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))

	stats, err := h.service.GetStats(ctx, instanceID.String())
	if err != nil {
		h.log.Error("failed to get stats", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

func (h *StatusCacheHandler) parseInstanceID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "instanceId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid instance id")
		return uuid.UUID{}, false
	}
	return id, true
}
