package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
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
	r.Delete("/instances/{instanceId}", h.deleteInstance)
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
	ID    string `json:"id"`    // Instance ID
	Token string `json:"token"` // Instance Token
	Due   int64  `json:"due"`   // Timestamp in milliseconds

	// Extended Fields (our API)
	InstanceID         uuid.UUID                  `json:"instanceId"`
	Name               string                     `json:"name"`
	SessionName        string                     `json:"sessionName"`
	InstanceToken      string                     `json:"instanceToken"`
	SubscriptionActive bool                       `json:"subscriptionActive"`
	CallRejectAuto     bool                       `json:"callRejectAuto"`
	CallRejectMessage  *string                    `json:"callRejectMessage,omitempty"`
	AutoReadMessage    bool                       `json:"autoReadMessage"`
	Middleware         string                     `json:"middleware"`
	Webhooks           *instances.WebhookSettings `json:"webhooks,omitempty"`
	CreatedAt          string                     `json:"createdAt"`
}

// listInstancesResponse represents the list response
type listInstancesResponse struct {
	Total     int                `json:"total"`
	TotalPage int                `json:"totalPage"`
	PageSize  int                `json:"pageSize"`
	Page      int                `json:"page"`
	Content   []instanceListItem `json:"content"`
}

// instanceListItem represents a single instance in the list
type instanceListItem struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	Due   int64  `json:"due"`

	// FUNNELCHAT Extended Fields
	Name              string `json:"name"`
	Created           string `json:"created"`
	PhoneConnected    bool   `json:"phoneConnected"`
	WhatsappConnected bool   `json:"whatsappConnected"`
	Middleware        string `json:"middleware"`

	// Webhook URLs (flattened from WebhookSettings)
	DeliveryCallbackUrl            *string `json:"deliveryCallbackUrl,omitempty"`
	ReceivedCallbackUrl            *string `json:"receivedCallbackUrl,omitempty"`
	ReceivedAndDeliveryCallbackUrl *string `json:"receivedAndDeliveryCallbackUrl,omitempty"`
	DisconnectedCallbackUrl        *string `json:"disconnectedCallbackUrl,omitempty"`
	ConnectedCallbackUrl           *string `json:"connectedCallbackUrl,omitempty"`
	MessageStatusCallbackUrl       *string `json:"messageStatusCallbackUrl,omitempty"`
	PresenceChatCallbackUrl        *string `json:"presenceChatCallbackUrl,omitempty"`

	// Additional Fields (our API)
	SessionName        string  `json:"sessionName,omitempty"`
	CallRejectAuto     bool    `json:"callRejectAuto"`
	CallRejectMessage  *string `json:"callRejectMessage,omitempty"`
	AutoReadMessage    bool    `json:"autoReadMessage"`
	SubscriptionActive bool    `json:"subscriptionActive"`
	NotifySentByMe     bool    `json:"notifySentByMe"`
	InstanceID         string  `json:"instanceId"`
	InstanceToken      string  `json:"instanceToken"`
}

func (h *PartnerHandler) createInstance(w http.ResponseWriter, r *http.Request) {
	var req partnerCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	ctx := r.Context()
	inst, err := h.service.CreatePartnerInstance(ctx, instances.PartnerCreateParams{
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
		logger := logging.ContextLogger(ctx, h.log)
		logger.Error("partner create instance", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to create instance")
		return
	}

	respondJSON(w, http.StatusCreated, partnerCreateResponse{

		ID:    inst.ID.String(),
		Token: inst.InstanceToken,
		Due:   time.Now().UnixMilli(),

		// Extended Fields
		InstanceID:         inst.ID,
		Name:               inst.Name,
		SessionName:        inst.SessionName,
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
	ctx := logging.WithAttrs(r.Context(), slog.String("instance_id", instanceID.String()))
	if err := h.service.SubscribeInstance(ctx, instanceID, instanceToken); err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "subscribed"})
}

func (h *PartnerHandler) cancelInstance(w http.ResponseWriter, r *http.Request) {
	instanceID, instanceToken, ok := h.parseInstancePath(w, r)
	if !ok {
		return
	}
	ctx := logging.WithAttrs(r.Context(), slog.String("instance_id", instanceID.String()))
	if err := h.service.CancelInstance(ctx, instanceID, instanceToken); err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "canceled"})
}

func (h *PartnerHandler) deleteInstance(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "instanceId")
	instanceID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid instance id")
		return
	}

	ctx := logging.WithAttrs(r.Context(), slog.String("instance_id", instanceID.String()))
	if err := h.service.DeleteInstance(ctx, instanceID); err != nil {
		h.handleServiceError(ctx, w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
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
	ctx := r.Context()
	result, err := h.service.ListInstances(ctx, filter)
	if err != nil {
		logger := logging.ContextLogger(ctx, h.log)
		logger.Error("partner list instances", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to list instances")
		return
	}

	totalPage := int(result.Total) / result.PageSize
	if int(result.Total)%result.PageSize > 0 {
		totalPage++
	}

	content := make([]instanceListItem, len(result.Data))
	for i, inst := range result.Data {
		item := instanceListItem{

			ID:                 inst.ID.String(),
			Token:              inst.InstanceToken,
			Due:                time.Now().UnixMilli(), // Current timestamp as default
			Name:               inst.Name,
			Created:            inst.CreatedAt.Format(time.RFC3339),
			PhoneConnected:     inst.PhoneConnected,
			WhatsappConnected:  inst.WhatsappConnected,
			Middleware:         inst.Middleware,
			SessionName:        inst.SessionName,
			CallRejectAuto:     inst.CallRejectAuto,
			CallRejectMessage:  inst.CallRejectMessage,
			AutoReadMessage:    inst.AutoReadMessage,
			SubscriptionActive: inst.SubscriptionActive,
			InstanceID:         inst.ID.String(),
			InstanceToken:      inst.InstanceToken,
		}

		// Flatten webhook settings
		if inst.Webhooks != nil {
			item.DeliveryCallbackUrl = inst.Webhooks.DeliveryURL
			item.ReceivedCallbackUrl = inst.Webhooks.ReceivedURL
			item.ReceivedAndDeliveryCallbackUrl = inst.Webhooks.ReceivedDeliveryURL
			item.DisconnectedCallbackUrl = inst.Webhooks.DisconnectedURL
			item.ConnectedCallbackUrl = inst.Webhooks.ConnectedURL
			item.MessageStatusCallbackUrl = inst.Webhooks.MessageStatusURL
			item.PresenceChatCallbackUrl = inst.Webhooks.ChatPresenceURL
			item.NotifySentByMe = inst.Webhooks.NotifySentByMe
		}

		content[i] = item
	}

	response := listInstancesResponse{
		Total:     int(result.Total),
		TotalPage: totalPage,
		PageSize:  result.PageSize,
		Page:      result.Page,
		Content:   content,
	}

	respondJSON(w, http.StatusOK, response)
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

func (h *PartnerHandler) handleServiceError(ctx context.Context, w http.ResponseWriter, err error) {
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
	logger := logging.ContextLogger(ctx, h.log)
	logger.Error("partner service error", slog.String("error", err.Error()))
	respondError(w, http.StatusInternalServerError, "internal error")
}
