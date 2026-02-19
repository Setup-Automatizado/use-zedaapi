package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
	natsclient "go.mau.fi/whatsmeow/api/internal/nats"
)

// DLQHandler handles Dead Letter Queue management endpoints.
type DLQHandler struct {
	dlqRepo         persistence.DLQRepository
	outboxRepo      persistence.OutboxRepository
	instanceService InstanceStatusProvider
	natsClient      *natsclient.Client
	natsEnabled     bool
	log             *slog.Logger
}

// NewDLQHandler creates a new DLQ handler.
func NewDLQHandler(
	dlqRepo persistence.DLQRepository,
	outboxRepo persistence.OutboxRepository,
	instanceService InstanceStatusProvider,
	natsClient *natsclient.Client,
	natsEnabled bool,
	log *slog.Logger,
) *DLQHandler {
	return &DLQHandler{
		dlqRepo:         dlqRepo,
		outboxRepo:      outboxRepo,
		instanceService: instanceService,
		natsClient:      natsClient,
		natsEnabled:     natsEnabled,
		log:             log.With(slog.String("component", "dlq_handler")),
	}
}

// RegisterRoutes registers DLQ endpoints under the instance route group.
func (h *DLQHandler) RegisterRoutes(r chi.Router) {
	r.Get("/dlq/stats", h.getStats)
	r.Get("/dlq/events", h.listEvents)
	r.Get("/dlq/events/{eventId}", h.getEvent)
	r.Post("/dlq/events/{eventId}/retry", h.retryEvent)
	r.Post("/dlq/events/{eventId}/discard", h.discardEvent)
	r.Post("/dlq/retry-all", h.retryAllPending)
	r.Delete("/dlq/purge", h.purgeResolved)
}

// --- Response types ---

type dlqStatsResponse struct {
	TotalEvents     int                 `json:"total_events"`
	ByStatus        map[string]int      `json:"by_status"`
	ByEventType     map[string]int      `json:"by_event_type"`
	ByFailureReason map[string]int      `json:"by_failure_reason"`
	NATSStream      *natsDLQStreamStats `json:"nats_stream,omitempty"`
}

type natsDLQStreamStats struct {
	Messages uint64 `json:"messages"`
	Bytes    uint64 `json:"bytes"`
	Subjects uint64 `json:"subjects"`
}

type dlqEventResponse struct {
	EventID           string          `json:"event_id"`
	InstanceID        string          `json:"instance_id"`
	EventType         string          `json:"event_type"`
	SourceLib         string          `json:"source_lib"`
	FailureReason     string          `json:"failure_reason"`
	LastError         string          `json:"last_error"`
	TotalAttempts     int             `json:"total_attempts"`
	ReprocessStatus   string          `json:"reprocess_status"`
	ReprocessAttempts int             `json:"reprocess_attempts"`
	MovedToDLQAt      time.Time       `json:"moved_to_dlq_at"`
	FirstAttemptAt    time.Time       `json:"first_attempt_at"`
	LastAttemptAt     time.Time       `json:"last_attempt_at"`
	CreatedAt         time.Time       `json:"created_at"`
	OriginalPayload   json.RawMessage `json:"original_payload,omitempty"`
	OriginalMetadata  json.RawMessage `json:"original_metadata,omitempty"`
}

type dlqListResponse struct {
	Events   []dlqEventResponse `json:"events"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

type dlqRetryResponse struct {
	Success bool   `json:"success"`
	EventID string `json:"event_id"`
	Message string `json:"message"`
}

type dlqRetryAllResponse struct {
	Success      bool `json:"success"`
	RetriedCount int  `json:"retried_count"`
	FailedCount  int  `json:"failed_count"`
}

type dlqPurgeResponse struct {
	Success      bool  `json:"success"`
	DeletedCount int64 `json:"deleted_count"`
}

// validDLQStatuses is the set of valid DLQ reprocess status values for filter validation.
var validDLQStatuses = map[persistence.DLQReprocessStatus]bool{
	persistence.DLQReprocessPending:    true,
	persistence.DLQReprocessProcessing: true,
	persistence.DLQReprocessSuccess:    true,
	persistence.DLQReprocessFailed:     true,
	persistence.DLQReprocessDiscarded:  true,
}

// --- Handlers ---

func (h *DLQHandler) getStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	stats, err := h.dlqRepo.GetStatsByInstance(ctx, instanceID)
	if err != nil {
		h.log.Error("failed to get DLQ stats",
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to get DLQ stats")
		return
	}

	resp := dlqStatsResponse{
		TotalEvents:     stats.TotalEvents,
		ByStatus:        stats.ByStatus,
		ByEventType:     stats.ByEventType,
		ByFailureReason: stats.ByFailureReason,
	}

	// Include NATS DLQ stream stats if available
	if h.natsEnabled && h.natsClient != nil {
		streamInfo, err := h.natsClient.StreamInfo(ctx, "DLQ")
		if err == nil {
			resp.NATSStream = &natsDLQStreamStats{
				Messages: streamInfo.Messages,
				Bytes:    streamInfo.Bytes,
				Subjects: streamInfo.Subjects,
			}
		}
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *DLQHandler) listEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	page := parseIntParam(r, "page", 1)
	pageSize := parseIntParam(r, "pageSize", 20)
	if pageSize > 100 {
		pageSize = 100
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if page < 1 {
		page = 1
	}

	filter := persistence.DLQFilter{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := persistence.DLQReprocessStatus(statusStr)
		if !validDLQStatuses[status] {
			respondError(w, http.StatusBadRequest, "invalid status filter value")
			return
		}
		filter.Status = &status
	}
	if eventType := r.URL.Query().Get("eventType"); eventType != "" {
		filter.EventType = &eventType
	}

	events, total, err := h.dlqRepo.GetByInstanceIDFiltered(ctx, instanceID, filter)
	if err != nil {
		h.log.Error("failed to list DLQ events",
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to list DLQ events")
		return
	}

	resp := dlqListResponse{
		Events:   make([]dlqEventResponse, 0, len(events)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, e := range events {
		resp.Events = append(resp.Events, toDLQEventResponse(e, false))
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *DLQHandler) getEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	eventID, err := uuid.Parse(chi.URLParam(r, "eventId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	event, err := h.dlqRepo.GetEventByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, persistence.ErrDLQEventNotFound) {
			respondError(w, http.StatusNotFound, "DLQ event not found")
			return
		}
		h.log.Error("failed to get DLQ event",
			slog.String("event_id", eventID.String()),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to get DLQ event")
		return
	}

	// Validate event belongs to requesting instance
	if event.InstanceID != instanceID {
		respondError(w, http.StatusNotFound, "DLQ event not found")
		return
	}

	respondJSON(w, http.StatusOK, toDLQEventResponse(event, true))
}

func (h *DLQHandler) retryEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	eventID, err := uuid.Parse(chi.URLParam(r, "eventId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	event, err := h.dlqRepo.GetEventByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, persistence.ErrDLQEventNotFound) {
			respondError(w, http.StatusNotFound, "DLQ event not found")
			return
		}
		h.log.Error("failed to get DLQ event for retry",
			slog.String("event_id", eventID.String()),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to get DLQ event")
		return
	}

	if event.InstanceID != instanceID {
		respondError(w, http.StatusNotFound, "DLQ event not found")
		return
	}

	if event.ReprocessStatus != persistence.DLQReprocessPending && event.ReprocessStatus != persistence.DLQReprocessFailed {
		respondError(w, http.StatusConflict, "event is not in a retryable state")
		return
	}

	// Re-enqueue via outbox (works for both PG and NATS modes)
	outboxEvent := &persistence.OutboxEvent{
		InstanceID:     event.InstanceID,
		EventID:        uuid.New(), // New event ID to avoid dedup conflicts
		EventType:      event.EventType,
		SourceLib:      event.SourceLib,
		Payload:        event.OriginalPayload,
		Metadata:       event.OriginalMetadata,
		Status:         persistence.EventStatusPending,
		Attempts:       0,
		MaxAttempts:    6,
		HasMedia:       false,
		MediaProcessed: true,
		TransportType:  persistence.TransportWebhook,
	}

	if err := h.outboxRepo.InsertEvent(ctx, outboxEvent); err != nil {
		h.log.Error("failed to re-enqueue DLQ event",
			slog.String("event_id", eventID.String()),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to retry event")
		return
	}

	// Mark as processing (without incrementing reprocess_attempts)
	if err := h.dlqRepo.MarkReprocessing(ctx, eventID); err != nil {
		h.log.Warn("failed to update DLQ reprocess status",
			slog.String("event_id", eventID.String()),
			slog.String("error", err.Error()))
	}

	h.log.Info("DLQ event retried",
		slog.String("instance_id", instanceID.String()),
		slog.String("event_id", eventID.String()))

	respondJSON(w, http.StatusOK, dlqRetryResponse{
		Success: true,
		EventID: eventID.String(),
		Message: "event re-enqueued for delivery",
	})
}

func (h *DLQHandler) discardEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	eventID, err := uuid.Parse(chi.URLParam(r, "eventId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	// Verify event exists and belongs to instance
	event, err := h.dlqRepo.GetEventByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, persistence.ErrDLQEventNotFound) {
			respondError(w, http.StatusNotFound, "DLQ event not found")
			return
		}
		h.log.Error("failed to get DLQ event for discard",
			slog.String("event_id", eventID.String()),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to get DLQ event")
		return
	}

	if event.InstanceID != instanceID {
		respondError(w, http.StatusNotFound, "DLQ event not found")
		return
	}

	if err := h.dlqRepo.MarkDiscarded(ctx, eventID); err != nil {
		h.log.Error("failed to discard DLQ event",
			slog.String("event_id", eventID.String()),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to discard event")
		return
	}

	h.log.Info("DLQ event discarded",
		slog.String("instance_id", instanceID.String()),
		slog.String("event_id", eventID.String()))

	respondJSON(w, http.StatusOK, dlqRetryResponse{
		Success: true,
		EventID: eventID.String(),
		Message: "event discarded",
	})
}

func (h *DLQHandler) retryAllPending(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	limit := parseIntParam(r, "limit", 100)
	if limit > 500 {
		limit = 500
	}

	// Get retryable events (pending + failed) for this instance
	events, _, err := h.dlqRepo.GetRetryableByInstance(ctx, instanceID, limit)
	if err != nil {
		h.log.Error("failed to get retryable DLQ events for retry-all",
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to get retryable events")
		return
	}

	retriedCount := 0
	failedCount := 0

	for _, event := range events {
		outboxEvent := &persistence.OutboxEvent{
			InstanceID:     event.InstanceID,
			EventID:        uuid.New(),
			EventType:      event.EventType,
			SourceLib:      event.SourceLib,
			Payload:        event.OriginalPayload,
			Metadata:       event.OriginalMetadata,
			Status:         persistence.EventStatusPending,
			Attempts:       0,
			MaxAttempts:    6,
			HasMedia:       false,
			MediaProcessed: true,
			TransportType:  persistence.TransportWebhook,
		}

		if err := h.outboxRepo.InsertEvent(ctx, outboxEvent); err != nil {
			h.log.Warn("failed to re-enqueue DLQ event during retry-all",
				slog.String("event_id", event.EventID.String()),
				slog.String("error", err.Error()))
			failedCount++
			continue
		}

		if err := h.dlqRepo.MarkReprocessing(ctx, event.EventID); err != nil {
			h.log.Warn("failed to update reprocess status during retry-all",
				slog.String("event_id", event.EventID.String()),
				slog.String("error", err.Error()))
		}
		retriedCount++
	}

	h.log.Info("DLQ retry-all completed",
		slog.String("instance_id", instanceID.String()),
		slog.Int("retried", retriedCount),
		slog.Int("failed", failedCount))

	respondJSON(w, http.StatusOK, dlqRetryAllResponse{
		Success:      true,
		RetriedCount: retriedCount,
		FailedCount:  failedCount,
	})
}

func (h *DLQHandler) purgeResolved(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	olderThanDays := parseIntParam(r, "olderThanDays", 7)
	if olderThanDays < 1 {
		olderThanDays = 1
	}

	olderThan := time.Now().AddDate(0, 0, -olderThanDays)

	deleted, err := h.dlqRepo.DeleteResolvedByInstance(ctx, instanceID, olderThan)
	if err != nil {
		h.log.Error("failed to purge resolved DLQ events",
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to purge events")
		return
	}

	h.log.Info("DLQ purge completed",
		slog.String("instance_id", instanceID.String()),
		slog.Int64("deleted", deleted),
		slog.Int("older_than_days", olderThanDays))

	respondJSON(w, http.StatusOK, dlqPurgeResponse{
		Success:      true,
		DeletedCount: deleted,
	})
}

// --- Helpers ---

// resolveInstance validates instance ID, token, and Client-Token header.
// Returns the validated instance UUID or writes an error response.
func (h *DLQHandler) resolveInstance(ctx context.Context, w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "instanceId")
	instanceID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid instance id")
		return uuid.UUID{}, false
	}

	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))

	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	_, err = h.instanceService.GetStatus(ctx, instanceID, clientToken, instanceToken)
	if err != nil {
		if errors.Is(err, instances.ErrInstanceNotFound) {
			respondError(w, http.StatusNotFound, "instance not found")
			return uuid.UUID{}, false
		}
		if errors.Is(err, instances.ErrUnauthorized) {
			respondError(w, http.StatusUnauthorized, "invalid credentials")
			return uuid.UUID{}, false
		}
		logging.ContextLogger(ctx, h.log).Error("instance service error",
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "internal error")
		return uuid.UUID{}, false
	}

	return instanceID, true
}

func parseIntParam(r *http.Request, name string, defaultVal int) int {
	valStr := r.URL.Query().Get(name)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func toDLQEventResponse(e *persistence.DLQEvent, includePayload bool) dlqEventResponse {
	resp := dlqEventResponse{
		EventID:           e.EventID.String(),
		InstanceID:        e.InstanceID.String(),
		EventType:         e.EventType,
		SourceLib:         e.SourceLib,
		FailureReason:     e.FailureReason,
		LastError:         e.LastError,
		TotalAttempts:     e.TotalAttempts,
		ReprocessStatus:   string(e.ReprocessStatus),
		ReprocessAttempts: e.ReprocessAttempts,
		MovedToDLQAt:      e.MovedToDLQAt,
		FirstAttemptAt:    e.FirstAttemptAt,
		LastAttemptAt:     e.LastAttemptAt,
		CreatedAt:         e.CreatedAt,
	}
	if includePayload {
		resp.OriginalPayload = e.OriginalPayload
		resp.OriginalMetadata = e.OriginalMetadata
	}
	return resp
}
