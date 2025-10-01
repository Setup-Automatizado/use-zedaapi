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
)

// InstanceHandler exposes HTTP endpoints for managing instances.
type InstanceHandler struct {
	service *instances.Service
	log     *slog.Logger
}

func NewInstanceHandler(service *instances.Service, log *slog.Logger) *InstanceHandler {
	return &InstanceHandler{service: service, log: log}
}

// Register mounts routes under /instances.
func (h *InstanceHandler) Register(r chi.Router) {
	r.Route("/instances/{instanceId}/token/{token}", func(r chi.Router) {
		r.Get("/status", h.getStatus)
		r.Get("/qr-code", h.getQRCode)
		r.Get("/qr-code/image", h.getQRCodeImage)
		r.Get("/phone-code/{phone}", h.getPhoneCode)
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
	})
}

func (h *InstanceHandler) getStatus(w http.ResponseWriter, r *http.Request) {
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")
	status, err := h.service.GetStatus(r.Context(), instanceID, clientToken, instanceToken)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, status)
}

func (h *InstanceHandler) getQRCode(w http.ResponseWriter, r *http.Request) {
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	code, err := h.service.GetQRCode(r.Context(), instanceID, clientToken, instanceToken)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"code": code})
}

type qrImageResponse struct {
	Image string `json:"image"`
}

func (h *InstanceHandler) getQRCodeImage(w http.ResponseWriter, r *http.Request) {
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	image, err := h.service.GetQRCodeImage(r.Context(), instanceID, clientToken, instanceToken)
	if err != nil {
		h.handleServiceError(w, err)
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

type webhookUpdateResponse struct {
	Value    bool                       `json:"value"`
	Webhooks *instances.WebhookSettings `json:"webhooks,omitempty"`
}

func (h *InstanceHandler) getPhoneCode(w http.ResponseWriter, r *http.Request) {
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")
	phone := chi.URLParam(r, "phone")
	if phone == "" {
		respondError(w, http.StatusBadRequest, "missing phone number")
		return
	}

	code, err := h.service.GetPhoneCode(r.Context(), instanceID, clientToken, instanceToken, phone)
	if err != nil {
		h.handleServiceError(w, err)
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
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	var req notifySentByMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	settings, err := h.service.UpdateNotifySentByMe(r.Context(), instanceID, clientToken, instanceToken, req.NotifySentByMe)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, webhookUpdateResponse{Value: true, Webhooks: settings})
}

func (h *InstanceHandler) updateEveryWebhooks(w http.ResponseWriter, r *http.Request) {
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	var req webhookEveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	settings, err := h.service.UpdateEveryWebhooks(r.Context(), instanceID, clientToken, instanceToken, req.Value, req.NotifySentByMe)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, webhookUpdateResponse{Value: true, Webhooks: settings})
}

func (h *InstanceHandler) updateWebhookWithValue(w http.ResponseWriter, r *http.Request, updater func(context.Context, uuid.UUID, string, string, string) (*instances.WebhookSettings, error)) {
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	var req webhookValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	settings, err := updater(r.Context(), instanceID, clientToken, instanceToken, req.Value)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, webhookUpdateResponse{Value: true, Webhooks: settings})
}

func (h *InstanceHandler) restart(w http.ResponseWriter, r *http.Request) {
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	if err := h.service.Restart(r.Context(), instanceID, clientToken, instanceToken); err != nil {
		h.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "restarting"})
}

func (h *InstanceHandler) disconnect(w http.ResponseWriter, r *http.Request) {
	instanceID, ok := h.parseInstanceID(w, r)
	if !ok {
		return
	}
	instanceToken := chi.URLParam(r, "token")
	clientToken := r.Header.Get("Client-Token")

	if err := h.service.Disconnect(r.Context(), instanceID, clientToken, instanceToken); err != nil {
		h.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "disconnected"})
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

func (h *InstanceHandler) handleServiceError(w http.ResponseWriter, err error) {
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
	h.log.Error("service error", slog.String("error", err.Error()))
	respondError(w, http.StatusInternalServerError, "internal error")
}
