package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
)

type InstanceHandler struct {
	service            *instances.Service
	messageHandler     *MessageHandler
	groupsHandler      *GroupsHandler
	communitiesHandler *CommunitiesHandler
	newslettersHandler *NewslettersHandler
	log                *slog.Logger
}

func NewInstanceHandler(service *instances.Service, log *slog.Logger) *InstanceHandler {
	return &InstanceHandler{service: service, log: log}
}

// SetMessageHandler injects the MessageHandler for route registration
func (h *InstanceHandler) SetMessageHandler(messageHandler *MessageHandler) {
	h.messageHandler = messageHandler
}

// SetGroupsHandler injects the GroupsHandler for route registration
func (h *InstanceHandler) SetGroupsHandler(groupsHandler *GroupsHandler) {
	h.groupsHandler = groupsHandler
}

// SetCommunitiesHandler injects the CommunitiesHandler for route registration
func (h *InstanceHandler) SetCommunitiesHandler(communitiesHandler *CommunitiesHandler) {
	h.communitiesHandler = communitiesHandler
}

// SetNewslettersHandler injects the NewslettersHandler for route registration
func (h *InstanceHandler) SetNewslettersHandler(newslettersHandler *NewslettersHandler) {
	h.newslettersHandler = newslettersHandler
}

func (h *InstanceHandler) Register(r chi.Router) {
	r.Route("/instances/{instanceId}/token/{token}", func(r chi.Router) {
		// Instance management routes
		r.Get("/status", h.getStatus)
		r.Get("/qr-code", h.getQRCode)
		r.Get("/qr-code/image", h.getQRCodeImage)
		r.Get("/device", h.getDevice)
		r.Get("/phone-code/{phone}", h.getPhoneCode)
		r.Get("/restart", h.restart)
		r.Get("/disconnect", h.disconnect)
		r.Post("/restart", h.restart)
		r.Post("/disconnect", h.disconnect)
		r.Put("/update-webhook-delivery", h.updateWebhookDelivery)
		r.Put("/update-webhook-received", h.updateWebhookReceived)
		r.Put("/update-webhook-received-delivery", h.updateWebhookReceivedDelivery)
		r.Put("/update-webhook-message-status", h.updateWebhookMessageStatus)
		r.Put("/update-webhook-disconnected", h.updateWebhookDisconnected)
		r.Put("/update-webhook-connected", h.updateWebhookConnected)
		r.Put("/update-webhook-chat-presence", h.updateWebhookChatPresence)
		r.Put("/update-notify-sent-by-me", h.updateNotifySentByMe)
		r.Put("/update-every-webhooks", h.updateEveryWebhooks)
		r.Put("/update-call-reject-auto", h.updateCallRejectAuto)
		r.Put("/update-call-reject-message", h.updateCallRejectMessage)
		r.Put("/update-auto-read-message", h.updateAutoReadMessage)

		// Message routes (delegated to MessageHandler)
		if h.messageHandler != nil {
			h.messageHandler.RegisterRoutes(r)
		}

		// Group/community shared routes
		if h.groupsHandler != nil {
			h.groupsHandler.RegisterRoutes(r)
		}

		// Community specific routes
		if h.communitiesHandler != nil {
			h.communitiesHandler.RegisterRoutes(r)
		}

		// Newsletter routes
		if h.newslettersHandler != nil {
			h.newslettersHandler.RegisterRoutes(r)
		}
	})
}

func (h *InstanceHandler) getStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")
	status, err := h.service.GetStatus(ctx, instanceID, clientToken, instanceToken)
	if err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, status)
}

func (h *InstanceHandler) getQRCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	code, err := h.service.GetQRCode(ctx, instanceID, clientToken, instanceToken)
	if err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(code))
}

type qrImageResponse struct {
	Image string `json:"image"`
}

func (h *InstanceHandler) getDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	device, err := h.service.GetDevice(ctx, instanceID, clientToken, instanceToken)
	if err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, device)
}

func (h *InstanceHandler) getQRCodeImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	image, err := h.service.GetQRCodeImage(ctx, instanceID, clientToken, instanceToken)
	if err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, qrImageResponse{Image: image})
}

type phoneCodeResponse struct {
	Code string `json:"code"`
}

type webhookValueRequest struct {
	Value string `json:"value"`
}

type webhookEveryRequest struct {
	Value          string `json:"value"`
	NotifySentByMe *bool  `json:"notifySentByMe,omitempty"`
}

type notifySentByMeRequest struct {
	NotifySentByMe bool `json:"notifySentByMe"`
}

type boolValueRequest struct {
	Value bool `json:"value"`
}

type stringValueRequest struct {
	Value string `json:"value"`
}

type webhookUpdateResponse struct {
	Value    bool                       `json:"value"`
	Webhooks *instances.WebhookSettings `json:"webhooks,omitempty"`
}

type valueResponse struct {
	Value bool `json:"value"`
}

func (h *InstanceHandler) getPhoneCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")
	phone := chi.URLParam(r, "phone")
	if phone == "" {
		respondError(w, http.StatusBadRequest, "missing phone number")
		return
	}

	code, err := h.service.GetPhoneCode(ctx, instanceID, clientToken, instanceToken, phone)
	if err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, phoneCodeResponse{Code: code})
}

func (h *InstanceHandler) updateWebhookDelivery(w http.ResponseWriter, r *http.Request) {
	h.updateWebhookWithValue(w, r, h.service.UpdateWebhookDelivery)
}

func (h *InstanceHandler) updateWebhookReceived(w http.ResponseWriter, r *http.Request) {
	h.updateWebhookWithValue(w, r, h.service.UpdateWebhookReceived)
}

func (h *InstanceHandler) updateWebhookReceivedDelivery(w http.ResponseWriter, r *http.Request) {
	h.updateWebhookWithValue(w, r, h.service.UpdateWebhookReceivedDelivery)
}

func (h *InstanceHandler) updateWebhookMessageStatus(w http.ResponseWriter, r *http.Request) {
	h.updateWebhookWithValue(w, r, h.service.UpdateWebhookMessageStatus)
}

func (h *InstanceHandler) updateWebhookDisconnected(w http.ResponseWriter, r *http.Request) {
	h.updateWebhookWithValue(w, r, h.service.UpdateWebhookDisconnected)
}

func (h *InstanceHandler) updateWebhookConnected(w http.ResponseWriter, r *http.Request) {
	h.updateWebhookWithValue(w, r, h.service.UpdateWebhookConnected)
}

func (h *InstanceHandler) updateWebhookChatPresence(w http.ResponseWriter, r *http.Request) {
	h.updateWebhookWithValue(w, r, h.service.UpdateWebhookChatPresence)
}

func (h *InstanceHandler) updateNotifySentByMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	var req notifySentByMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	settings, err := h.service.UpdateNotifySentByMe(ctx, instanceID, clientToken, instanceToken, req.NotifySentByMe)
	if err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, webhookUpdateResponse{Value: true, Webhooks: settings})
}

func (h *InstanceHandler) updateEveryWebhooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	var req webhookEveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	settings, err := h.service.UpdateEveryWebhooks(ctx, instanceID, clientToken, instanceToken, req.Value, req.NotifySentByMe)
	if err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, webhookUpdateResponse{Value: true, Webhooks: settings})
}

func (h *InstanceHandler) updateWebhookWithValue(w http.ResponseWriter, r *http.Request, updater func(context.Context, uuid.UUID, string, string, string) (*instances.WebhookSettings, error)) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	var req webhookValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	settings, err := updater(ctx, instanceID, clientToken, instanceToken, req.Value)
	if err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, webhookUpdateResponse{Value: true, Webhooks: settings})
}

func (h *InstanceHandler) restart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	if err := h.service.Restart(ctx, instanceID, clientToken, instanceToken); err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, valueResponse{Value: true})
}

func (h *InstanceHandler) disconnect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	if err := h.service.Disconnect(ctx, instanceID, clientToken, instanceToken); err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, valueResponse{Value: true})
}

func (h *InstanceHandler) parseInstanceID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "instanceId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid instance id")
		return uuid.UUID{}, false
	}
	return id, true
}

func (h *InstanceHandler) handleServiceError(ctx context.Context, w http.ResponseWriter, err error) {
	if errors.Is(err, instances.ErrInstanceNotFound) {
		respondError(w, http.StatusNotFound, "instance not found")
		return
	}
	if errors.Is(err, instances.ErrUnauthorized) {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if errors.Is(err, instances.ErrInvalidWebhookURL) || errors.Is(err, instances.ErrMissingWebhookValue) || errors.Is(err, instances.ErrInvalidPhoneNumber) {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, instances.ErrInstanceInactive) {
		respondError(w, http.StatusForbidden, "instance subscription inactive")
		return
	}
	if errors.Is(err, instances.ErrInstanceAlreadyPaired) {
		respondError(w, http.StatusConflict, "instance already paired")
		return
	}
	logger := logging.ContextLogger(ctx, h.log)
	logger.Error("service error", slog.String("error", err.Error()))
	respondError(w, http.StatusInternalServerError, "internal error")
}

func (h *InstanceHandler) updateCallRejectAuto(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	var req boolValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	if err := h.service.UpdateCallRejectAuto(ctx, instanceID, clientToken, instanceToken, req.Value); err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, valueResponse{Value: true})
}

func (h *InstanceHandler) updateCallRejectMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	var req stringValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	var value *string
	if req.Value != "" {
		value = &req.Value
	}
	if err := h.service.UpdateCallRejectMessage(ctx, instanceID, clientToken, instanceToken, value); err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, valueResponse{Value: true})
}

func (h *InstanceHandler) updateAutoReadMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	ctx = logging.WithAttrs(ctx, slog.String("instance_id", instanceID.String()))
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	var req boolValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	if err := h.service.UpdateAutoReadMessage(ctx, instanceID, clientToken, instanceToken, req.Value); err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, valueResponse{Value: true})
}
