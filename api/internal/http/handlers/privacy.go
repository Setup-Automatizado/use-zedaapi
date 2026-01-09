package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	wameow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"

	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/messages/queue"
)

// PrivacyHandler handles HTTP requests for privacy settings operations
type PrivacyHandler struct {
	coordinator     *queue.Coordinator
	instanceService InstanceStatusProvider
	log             *slog.Logger
}

// NewPrivacyHandler creates a new privacy handler
func NewPrivacyHandler(
	coordinator *queue.Coordinator,
	instanceService InstanceStatusProvider,
	log *slog.Logger,
) *PrivacyHandler {
	return &PrivacyHandler{
		coordinator:     coordinator,
		instanceService: instanceService,
		log:             log,
	}
}

// RegisterRoutes registers privacy routes within an existing route group
func (h *PrivacyHandler) RegisterRoutes(r chi.Router) {
	// Privacy settings endpoints
	r.Get("/privacy-settings", h.getPrivacySettings)
	r.Put("/privacy-settings/group-add", h.updatePrivacyGroupAdd)
	r.Put("/privacy-settings/last-seen", h.updatePrivacyLastSeen)
	r.Put("/privacy-settings/status", h.updatePrivacyStatus)
	r.Put("/privacy-settings/profile-photo", h.updatePrivacyProfilePhoto)
	r.Put("/privacy-settings/read-receipts", h.updatePrivacyReadReceipts)
	r.Put("/privacy-settings/online", h.updatePrivacyOnline)
	r.Put("/privacy-settings/call-add", h.updatePrivacyCallAdd)
}

// PrivacySettingsResponse represents the response for GET /privacy-settings
type PrivacySettingsResponse struct {
	GroupAdd     string `json:"groupAdd"`
	LastSeen     string `json:"lastSeen"`
	Status       string `json:"status"`
	Profile      string `json:"profile"`
	ReadReceipts string `json:"readReceipts"`
	Online       string `json:"online"`
	CallAdd      string `json:"callAdd"`
}

// UpdatePrivacyRequest represents the request for PUT /privacy-settings/*
type UpdatePrivacyRequest struct {
	Value string `json:"value"`
}

// UpdatePrivacyResponse represents the response for PUT /privacy-settings/*
type UpdatePrivacyResponse struct {
	Success  bool                    `json:"success"`
	Setting  string                  `json:"setting"`
	Value    string                  `json:"value"`
	Settings PrivacySettingsResponse `json:"settings"`
}

// resolveInstance validates instance ID and tokens, returns the instance context
func (h *PrivacyHandler) resolveInstance(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, uuid.UUID, *instances.Status, bool) {
	instanceIDStr := chi.URLParam(r, "instanceId")
	instanceID, err := uuid.Parse(instanceIDStr)
	if err != nil {
		h.log.WarnContext(ctx, "invalid instance_id",
			slog.String("instance_id", instanceIDStr),
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid instance ID format")
		return ctx, uuid.UUID{}, nil, false
	}

	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))

	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	status, err := h.instanceService.GetStatus(ctx, instanceID, clientToken, instanceToken)
	if err != nil {
		h.handleInstanceServiceError(ctx, w, err)
		return ctx, uuid.UUID{}, nil, false
	}

	return ctx, instanceID, status, true
}

// handleInstanceServiceError handles errors from instance service
func (h *PrivacyHandler) handleInstanceServiceError(ctx context.Context, w http.ResponseWriter, err error) {
	switch {
	case err.Error() == "instance not found":
		respondError(w, http.StatusNotFound, "Instance not found")
	case err.Error() == "unauthorized":
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
	default:
		logging.ContextLogger(ctx, h.log).Error("instance service error",
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Internal error")
	}
}

// getClient retrieves the whatsmeow client for an instance
func (h *PrivacyHandler) getClient(ctx context.Context, w http.ResponseWriter, instanceID uuid.UUID) (*wameow.Client, bool) {
	clientRegistry, ok := h.coordinator.GetClient(instanceID)
	if !ok {
		h.log.ErrorContext(ctx, "client registry not available")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not available")
		return nil, false
	}

	client, ok := clientRegistry.GetClient(instanceID.String())
	if !ok || client == nil {
		h.log.WarnContext(ctx, "whatsapp client not connected")
		respondError(w, http.StatusServiceUnavailable, "WhatsApp client not connected")
		return nil, false
	}

	return client, true
}

// getPrivacySettings handles GET /instances/{instanceId}/token/{token}/privacy-settings
func (h *PrivacyHandler) getPrivacySettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	client, ok := h.getClient(ctx, w, instanceID)
	if !ok {
		return
	}

	// Get privacy settings from whatsmeow
	settings := client.GetPrivacySettings(ctx)

	h.log.InfoContext(ctx, "retrieved privacy settings")

	response := PrivacySettingsResponse{
		GroupAdd:     string(settings.GroupAdd),
		LastSeen:     string(settings.LastSeen),
		Status:       string(settings.Status),
		Profile:      string(settings.Profile),
		ReadReceipts: string(settings.ReadReceipts),
		Online:       string(settings.Online),
		CallAdd:      string(settings.CallAdd),
	}

	respondJSON(w, http.StatusOK, response)
}

// updatePrivacySetting is a generic function to update a privacy setting
func (h *PrivacyHandler) updatePrivacySetting(w http.ResponseWriter, r *http.Request, settingType types.PrivacySettingType, settingName string) {
	ctx := r.Context()

	ctx, instanceID, _, ok := h.resolveInstance(ctx, w, r)
	if !ok {
		return
	}

	// Parse request body
	var req UpdatePrivacyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.WarnContext(ctx, "invalid request body",
			slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Value == "" {
		h.log.WarnContext(ctx, "missing value in request")
		respondError(w, http.StatusBadRequest, "Value is required")
		return
	}

	// Validate privacy setting value
	privacyValue := types.PrivacySetting(req.Value)
	if !isValidPrivacyValue(settingType, privacyValue) {
		h.log.WarnContext(ctx, "invalid privacy value",
			slog.String("setting", settingName),
			slog.String("value", req.Value))
		respondError(w, http.StatusBadRequest, "Invalid privacy value for "+settingName)
		return
	}

	client, ok := h.getClient(ctx, w, instanceID)
	if !ok {
		return
	}

	// Update privacy setting
	newSettings, err := client.SetPrivacySetting(ctx, settingType, privacyValue)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to update privacy setting",
			slog.String("setting", settingName),
			slog.String("value", req.Value),
			slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "Failed to update privacy setting: "+err.Error())
		return
	}

	h.log.InfoContext(ctx, "privacy setting updated",
		slog.String("setting", settingName),
		slog.String("value", req.Value))

	response := UpdatePrivacyResponse{
		Success: true,
		Setting: settingName,
		Value:   req.Value,
		Settings: PrivacySettingsResponse{
			GroupAdd:     string(newSettings.GroupAdd),
			LastSeen:     string(newSettings.LastSeen),
			Status:       string(newSettings.Status),
			Profile:      string(newSettings.Profile),
			ReadReceipts: string(newSettings.ReadReceipts),
			Online:       string(newSettings.Online),
			CallAdd:      string(newSettings.CallAdd),
		},
	}

	respondJSON(w, http.StatusOK, response)
}

// isValidPrivacyValue validates if a privacy value is valid for a given setting type
func isValidPrivacyValue(settingType types.PrivacySettingType, value types.PrivacySetting) bool {
	switch settingType {
	case types.PrivacySettingTypeGroupAdd, types.PrivacySettingTypeLastSeen, types.PrivacySettingTypeStatus, types.PrivacySettingTypeProfile:
		// Valid values: all, contacts, contact_blacklist, none
		return value == types.PrivacySettingAll ||
			value == types.PrivacySettingContacts ||
			value == types.PrivacySettingContactBlacklist ||
			value == types.PrivacySettingNone
	case types.PrivacySettingTypeReadReceipts:
		// Valid values: all, none
		return value == types.PrivacySettingAll || value == types.PrivacySettingNone
	case types.PrivacySettingTypeOnline:
		// Valid values: all, match_last_seen
		return value == types.PrivacySettingAll || value == types.PrivacySettingMatchLastSeen
	case types.PrivacySettingTypeCallAdd:
		// Valid values: all, known
		return value == types.PrivacySettingAll || value == types.PrivacySettingKnown
	default:
		return false
	}
}

// updatePrivacyGroupAdd handles PUT /instances/{instanceId}/token/{token}/privacy-settings/group-add
func (h *PrivacyHandler) updatePrivacyGroupAdd(w http.ResponseWriter, r *http.Request) {
	h.updatePrivacySetting(w, r, types.PrivacySettingTypeGroupAdd, "group-add")
}

// updatePrivacyLastSeen handles PUT /instances/{instanceId}/token/{token}/privacy-settings/last-seen
func (h *PrivacyHandler) updatePrivacyLastSeen(w http.ResponseWriter, r *http.Request) {
	h.updatePrivacySetting(w, r, types.PrivacySettingTypeLastSeen, "last-seen")
}

// updatePrivacyStatus handles PUT /instances/{instanceId}/token/{token}/privacy-settings/status
func (h *PrivacyHandler) updatePrivacyStatus(w http.ResponseWriter, r *http.Request) {
	h.updatePrivacySetting(w, r, types.PrivacySettingTypeStatus, "status")
}

// updatePrivacyProfilePhoto handles PUT /instances/{instanceId}/token/{token}/privacy-settings/profile-photo
func (h *PrivacyHandler) updatePrivacyProfilePhoto(w http.ResponseWriter, r *http.Request) {
	h.updatePrivacySetting(w, r, types.PrivacySettingTypeProfile, "profile-photo")
}

// updatePrivacyReadReceipts handles PUT /instances/{instanceId}/token/{token}/privacy-settings/read-receipts
func (h *PrivacyHandler) updatePrivacyReadReceipts(w http.ResponseWriter, r *http.Request) {
	h.updatePrivacySetting(w, r, types.PrivacySettingTypeReadReceipts, "read-receipts")
}

// updatePrivacyOnline handles PUT /instances/{instanceId}/token/{token}/privacy-settings/online
func (h *PrivacyHandler) updatePrivacyOnline(w http.ResponseWriter, r *http.Request) {
	h.updatePrivacySetting(w, r, types.PrivacySettingTypeOnline, "online")
}

// updatePrivacyCallAdd handles PUT /instances/{instanceId}/token/{token}/privacy-settings/call-add
func (h *PrivacyHandler) updatePrivacyCallAdd(w http.ResponseWriter, r *http.Request) {
	h.updatePrivacySetting(w, r, types.PrivacySettingTypeCallAdd, "call-add")
}
