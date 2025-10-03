package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"go.mau.fi/whatsmeow/api/internal/events/media"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// MediaHandler handles serving local media files
type MediaHandler struct {
	localStorage *media.LocalMediaStorage
	metrics      *observability.Metrics
	logger       *slog.Logger
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(
	localStorage *media.LocalMediaStorage,
	metrics *observability.Metrics,
	logger *slog.Logger,
) *MediaHandler {
	return &MediaHandler{
		localStorage: localStorage,
		metrics:      metrics,
		logger:       logger.With(slog.String("handler", "media")),
	}
}

// ServeMedia handles GET /v1/media/{instance_id}/{path}?expires={timestamp}&signature={hmac}
func (h *MediaHandler) ServeMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.ContextLogger(ctx, h.logger)

	// Extract instance_id and path from URL
	instanceID := chi.URLParam(r, "instance_id")
	pathSuffix := chi.URLParam(r, "*") // Everything after instance_id/

	if instanceID == "" || pathSuffix == "" {
		logger.Warn("missing instance_id or path in request",
			slog.String("instance_id", instanceID),
			slog.String("path", pathSuffix))
		http.Error(w, "Invalid path", http.StatusBadRequest)
		h.metrics.MediaServeRequests.WithLabelValues(instanceID, "error", "invalid_path").Inc()
		return
	}

	// Reconstruct full relative path
	relativePath := fmt.Sprintf("%s/%s", instanceID, pathSuffix)

	// Extract query parameters
	expiresStr := r.URL.Query().Get("expires")
	signature := r.URL.Query().Get("signature")

	if expiresStr == "" || signature == "" {
		logger.Warn("missing expires or signature parameters",
			slog.String("path", relativePath))
		http.Error(w, "Missing expires or signature", http.StatusBadRequest)
		h.metrics.MediaServeRequests.WithLabelValues(instanceID, "error", "missing_params").Inc()
		return
	}

	logger.Debug("serving media request",
		slog.String("path", relativePath),
		slog.String("expires", expiresStr))

	// Serve media using CopyToWriter for efficiency
	contentType, fileSize, err := h.localStorage.CopyToWriter(ctx, relativePath, expiresStr, signature, w)
	if err != nil {
		h.handleMediaError(w, r, instanceID, relativePath, err)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
	w.Header().Set("Cache-Control", "public, max-age=86400") // 24 hours
	w.Header().Set("X-Content-Type-Options", "nosniff")

	logger.Info("media served successfully",
		slog.String("path", relativePath),
		slog.Int64("size", fileSize),
		slog.String("content_type", contentType))

	h.metrics.MediaServeRequests.WithLabelValues(instanceID, "success", "ok").Inc()
	h.metrics.MediaServeBytes.WithLabelValues(instanceID).Add(float64(fileSize))
}

// handleMediaError handles errors when serving media
func (h *MediaHandler) handleMediaError(w http.ResponseWriter, r *http.Request, instanceID, path string, err error) {
	logger := logging.ContextLogger(r.Context(), h.logger)

	errMsg := err.Error()

	// Determine HTTP status based on error
	var status int
	var errorType string

	switch {
	case strings.Contains(errMsg, "validation failed"):
		if strings.Contains(errMsg, "expired") {
			status = http.StatusGone
			errorType = "url_expired"
		} else if strings.Contains(errMsg, "invalid signature") {
			status = http.StatusForbidden
			errorType = "invalid_signature"
		} else {
			status = http.StatusBadRequest
			errorType = "validation_error"
		}

	case strings.Contains(errMsg, "file not found"):
		status = http.StatusNotFound
		errorType = "not_found"

	case strings.Contains(errMsg, "path traversal"):
		status = http.StatusBadRequest
		errorType = "path_traversal"

	case strings.Contains(errMsg, "path is a directory"):
		status = http.StatusBadRequest
		errorType = "invalid_path"

	default:
		status = http.StatusInternalServerError
		errorType = "internal_error"
	}

	logger.Warn("failed to serve media",
		slog.String("path", path),
		slog.String("error", errMsg),
		slog.String("error_type", errorType),
		slog.Int("status", status))

	http.Error(w, errMsg, status)
	h.metrics.MediaServeRequests.WithLabelValues(instanceID, "error", errorType).Inc()
}

// GetStats returns statistics about local media storage
func (h *MediaHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.ContextLogger(ctx, h.logger)

	stats, err := h.localStorage.GetStats(ctx)
	if err != nil {
		logger.Error("failed to get storage stats",
			slog.String("error", err.Error()))
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Manually construct JSON to avoid importing encoding/json
	fmt.Fprintf(w, `{"total_files":%d,"total_bytes":%d,"base_path":"%s","url_expiry":"%s","public_base_url":"%s"}`,
		stats["total_files"],
		stats["total_bytes"],
		stats["base_path"],
		stats["url_expiry"],
		stats["public_base_url"])

	logger.Debug("storage stats returned")
}
