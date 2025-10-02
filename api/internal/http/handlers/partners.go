package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/instances"
)

type PartnerHandler struct {
	service *instances.Service
	log     *slog.Logger
}

func NewPartnerHandler(service *instances.Service, log *slog.Logger) *PartnerHandler {
	return &PartnerHandler{service: service, log: log}
}

func (h *PartnerHandler) Register(r chi.Router) {
	r.Post("/instances/integrator/on-demand", h.createInstance)
	r.Post("/instances/{instanceId}/token/{token}/integrator/on-demand/subscription", h.subscribeInstance)
	r.Post("/instances/{instanceId}/token/{token}/integrator/on-demand/cancel", h.cancelInstance)
	r.Get("/instances", h.listInstances)
}

type partnerCreateRequest struct {
	Name                        string  `json:"name"`
	SessionName                 string  `json:"sessionName"`
	DeliveryCallbackURL         *string `json:"deliveryCallbackUrl"`
	ReceivedCallbackURL         *string `json:"receivedCallbackUrl"`
	ReceivedDeliveryCallbackURL *string `json:"receivedAndDeliveryCallbackUrl"`
	DisconnectedCallbackURL     *string `json:"disconnectedCallbackUrl"`
	ConnectedCallbackURL        *string `json:"connectedCallbackUrl"`
	MessageStatusCallbackURL    *string `json:"messageStatusCallbackUrl"`
	PresenceCallbackURL         *string `json:"presenceChatCallbackUrl"`
	NotifySentByMe              bool    `json:"notifySentByMe"`
	CallRejectAuto              *bool   `json:"callRejectAuto"`
	CallRejectMessage           *string `json:"callRejectMessage"`
	AutoReadMessage             *bool   `json:"autoReadMessage"`
	IsDevice                    bool    `json:"isDevice"`
	BusinessDevice              bool    `json:"businessDevice"`
}

type partnerCreateResponse struct {
	InstanceID         uuid.UUID                  `json:"instanceId"`
	Name               string                     `json:"name"`
	SessionName        string                     `json:"sessionName"`
	ClientToken        string                     `json:"clientToken"`
	InstanceToken      string                     `json:"instanceToken"`
	SubscriptionActive bool                       `json:"subscriptionActive"`
	CallRejectAuto     bool                       `json:"callRejectAuto"`
	CallRejectMessage  *string                    `json:"callRejectMessage,omitempty"`
	AutoReadMessage    bool                       `json:"autoReadMessage"`
	Middleware         string                     `json:"middleware"`
	Webhooks           *instances.WebhookSettings `json:"webhooks,omitempty"`
	CreatedAt          string                     `json:"createdAt"`
}

func (h *PartnerHandler) createInstance(w http.ResponseWriter, r *http.Request) {
	var req partnerCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	inst, err := h.service.CreatePartnerInstance(r.Context(), instances.PartnerCreateParams{
		Name:                        req.Name,
		SessionName:                 req.SessionName,
		DeliveryCallbackURL:         req.DeliveryCallbackURL,
		ReceivedCallbackURL:         req.ReceivedCallbackURL,
		ReceivedDeliveryCallbackURL: req.ReceivedDeliveryCallbackURL,
		DisconnectedCallbackURL:     req.DisconnectedCallbackURL,
		ConnectedCallbackURL:        req.ConnectedCallbackURL,
		MessageStatusCallbackURL:    req.MessageStatusCallbackURL,
		ChatPresenceCallbackURL:     req.PresenceCallbackURL,
		NotifySentByMe:              req.NotifySentByMe,
		CallRejectAuto:              req.CallRejectAuto,
		CallRejectMessage:           req.CallRejectMessage,
		AutoReadMessage:             req.AutoReadMessage,
		IsDevice:                    req.IsDevice,
		BusinessDevice:              req.BusinessDevice,
	})
	if err != nil {
		h.log.Error("partner create instance", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to create instance")
		return
	}

	respondJSON(w, http.StatusCreated, partnerCreateResponse{
		InstanceID:         inst.ID,
		Name:               inst.Name,
		SessionName:        inst.SessionName,
		ClientToken:        inst.ClientToken,
		InstanceToken:      inst.InstanceToken,
		SubscriptionActive: inst.SubscriptionActive,
		CallRejectAuto:     inst.CallRejectAuto,
		CallRejectMessage:  inst.CallRejectMessage,
		AutoReadMessage:    inst.AutoReadMessage,
		Middleware:         inst.Middleware,
		Webhooks:           inst.Webhooks,
		CreatedAt:          inst.CreatedAt.Format(time.RFC3339),
	})
}

func (h *PartnerHandler) subscribeInstance(w http.ResponseWriter, r *http.Request) {
	instanceID, instanceToken, ok := h.parseInstancePath(w, r)
	if !ok {
		return
	}
	if err := h.service.SubscribeInstance(r.Context(), instanceID, instanceToken); err != nil {
		h.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "subscribed"})
}

func (h *PartnerHandler) cancelInstance(w http.ResponseWriter, r *http.Request) {
	instanceID, instanceToken, ok := h.parseInstancePath(w, r)
	if !ok {
		return
	}
	if err := h.service.CancelInstance(r.Context(), instanceID, instanceToken); err != nil {
		h.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "canceled"})
}

func (h *PartnerHandler) listInstances(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	filter := instances.ListFilter{
		Query:      r.URL.Query().Get("query"),
		Middleware: r.URL.Query().Get("middleware"),
		Page:       page,
		PageSize:   pageSize,
	}
	result, err := h.service.ListInstances(r.Context(), filter)
	if err != nil {
		h.log.Error("partner list instances", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to list instances")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (h *PartnerHandler) parseInstancePath(w http.ResponseWriter, r *http.Request) (uuid.UUID, string, bool) {
	idStr := chi.URLParam(r, "instanceId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid instance id")
		return uuid.UUID{}, "", false
	}
	token := chi.URLParam(r, "token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "missing instance token")
		return uuid.UUID{}, "", false
	}
	return id, token, true
}

func (h *PartnerHandler) handleServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, instances.ErrInstanceNotFound) {
		respondError(w, http.StatusNotFound, "instance not found")
		return
	}
	if errors.Is(err, instances.ErrInvalidWebhookURL) || errors.Is(err, instances.ErrMissingWebhookValue) {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, instances.ErrInstanceInactive) {
		respondError(w, http.StatusForbidden, "instance subscription inactive")
		return
	}
	h.log.Error("partner service error", slog.String("error", err.Error()))
	respondError(w, http.StatusInternalServerError, "internal error")
}
