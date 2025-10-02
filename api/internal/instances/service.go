package instances

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"

	whatsmeowpkg "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/whatsmeow"
)

var (
	ErrUnauthorized          = errors.New("unauthorized")
	ErrInvalidWebhookURL     = errors.New("webhook url must use https")
	ErrMissingWebhookValue   = errors.New("webhook value is required")
	ErrInvalidPhoneNumber    = errors.New("invalid phone number")
	ErrInstanceInactive      = errors.New("instance subscription inactive")
	ErrInstanceAlreadyPaired = errors.New("instance already paired")
)

type Service struct {
	repo     *Repository
	registry *whatsmeow.ClientRegistry
	log      *slog.Logger
}

func NewService(repo *Repository, registry *whatsmeow.ClientRegistry, log *slog.Logger) *Service {
	return &Service{repo: repo, registry: registry, log: log}
}

func (s *Service) Create(ctx context.Context, params CreateParams) (*Instance, error) {
	inst := Instance{
		ID:              uuid.New(),
		Name:            params.Name,
		SessionName:     params.Name,
		ClientToken:     uuid.NewString(),
		InstanceToken:   uuid.NewString(),
		IsDevice:        false,
		BusinessDevice:  false,
		CallRejectAuto:  false,
		AutoReadMessage: false,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	inst.Middleware = "web"

	if err := s.repo.Insert(ctx, &inst); err != nil {
		return nil, err
	}

	go func(instance Instance) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, _, err := s.registry.EnsureClient(ctx, toInstanceInfo(instance)); err != nil {
			s.log.Error("ensure client", slog.String("instanceId", instance.ID.String()), slog.String("error", err.Error()))
		}
	}(inst)

	return &inst, nil
}

func (s *Service) GetStatus(ctx context.Context, id uuid.UUID, clientToken, instanceToken string) (*Status, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return nil, ErrUnauthorized
	}
	snapshot := s.registry.Status(toInstanceInfo(*inst))
	return &Status{
		Connected:          snapshot.Connected,
		StoreJID:           snapshot.StoreJID,
		LastConnected:      snapshot.LastConnected,
		InstanceID:         inst.ID,
		AutoReconnect:      snapshot.AutoReconnect,
		WorkerAssigned:     snapshot.WorkerID,
		SubscriptionActive: inst.SubscriptionActive,
	}, nil
}

func (s *Service) Restart(ctx context.Context, id uuid.UUID, clientToken, instanceToken string) error {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return ErrUnauthorized
	}
	if err := ensureActive(inst); err != nil {
		return err
	}
	return s.registry.Restart(ctx, toInstanceInfo(*inst))
}

func (s *Service) Disconnect(ctx context.Context, id uuid.UUID, clientToken, instanceToken string) error {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return ErrUnauthorized
	}
	return s.registry.Disconnect(ctx, toInstanceInfo(*inst))
}

func (s *Service) GetQRCode(ctx context.Context, id uuid.UUID, clientToken, instanceToken string) (string, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return "", ErrUnauthorized
	}
	if err := ensureActive(inst); err != nil {
		return "", err
	}
	code, err := s.registry.GetQRCode(ctx, toInstanceInfo(*inst))
	if err != nil {
		if errors.Is(err, whatsmeow.ErrInstanceAlreadyPaired) {
			return "", ErrInstanceAlreadyPaired
		}
		return "", fmt.Errorf("get qr code: %w", err)
	}
	return code, nil
}

func (s *Service) GetQRCodeImage(ctx context.Context, id uuid.UUID, clientToken, instanceToken string) (string, error) {
	code, err := s.GetQRCode(ctx, id, clientToken, instanceToken)
	if err != nil {
		return "", err
	}
	png, err := qrcode.Encode(code, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("encode qr image: %w", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

func (s *Service) GetPhoneCode(ctx context.Context, id uuid.UUID, clientToken, instanceToken, phone string) (string, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return "", ErrUnauthorized
	}
	if err := ensureActive(inst); err != nil {
		return "", err
	}
	code, err := s.registry.PairPhone(ctx, toInstanceInfo(*inst), phone)
	if err != nil {
		if errors.Is(err, whatsmeowpkg.ErrPhoneNumberTooShort) || errors.Is(err, whatsmeowpkg.ErrPhoneNumberIsNotInternational) {
			return "", ErrInvalidPhoneNumber
		}
		if errors.Is(err, whatsmeow.ErrInstanceAlreadyPaired) {
			return "", ErrInstanceAlreadyPaired
		}
		return "", fmt.Errorf("pair phone: %w", err)
	}
	return code, nil
}

func (s *Service) ReconcileDetachedStores(ctx context.Context) ([]uuid.UUID, error) {
	links, err := s.repo.ListInstancesWithStoreJID(ctx)
	if err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return nil, nil
	}

	cleaned := make([]uuid.UUID, 0)
	for _, link := range links {
		exists, checkErr := s.registry.HasStoreDevice(ctx, link.StoreJID)
		if checkErr != nil {
			s.log.Error(
				"check store device",
				slog.String("instanceId", link.ID.String()),
				slog.String("storeJid", link.StoreJID),
				slog.String("error", checkErr.Error()),
			)
		}
		if checkErr != nil || !exists {
			s.registry.RemoveClient(link.ID, "reconcile_missing_store")
			if err := s.repo.UpdateStoreJID(ctx, link.ID, nil); err != nil {
				s.log.Error(
					"clear store jid",
					slog.String("instanceId", link.ID.String()),
					slog.String("error", err.Error()),
				)
				continue
			}
			s.log.Info(
				"reconciled missing store",
				slog.String("instanceId", link.ID.String()),
				slog.String("storeJid", link.StoreJID),
			)
			cleaned = append(cleaned, link.ID)
		}
	}
	return cleaned, nil
}

func (s *Service) tokensMatch(inst *Instance, clientToken, instanceToken string) bool {
	if inst == nil {
		return false
	}
	return inst.ClientToken == clientToken && inst.InstanceToken == instanceToken
}

func (s *Service) UpdateWebhookDelivery(ctx context.Context, id uuid.UUID, clientToken, instanceToken, value string) (*WebhookSettings, error) {
	normalized, err := normalizeWebhookValue(value)
	if err != nil {
		return nil, err
	}
	return s.updateWebhookConfig(ctx, id, clientToken, instanceToken, func(cfg *WebhookConfig) error {
		cfg.DeliveryURL = strPtr(normalized)
		return nil
	})
}

func (s *Service) UpdateWebhookReceived(ctx context.Context, id uuid.UUID, clientToken, instanceToken, value string) (*WebhookSettings, error) {
	normalized, err := normalizeWebhookValue(value)
	if err != nil {
		return nil, err
	}
	return s.updateWebhookConfig(ctx, id, clientToken, instanceToken, func(cfg *WebhookConfig) error {
		cfg.ReceivedURL = strPtr(normalized)
		return nil
	})
}

func (s *Service) UpdateWebhookReceivedDelivery(ctx context.Context, id uuid.UUID, clientToken, instanceToken, value string) (*WebhookSettings, error) {
	normalized, err := normalizeWebhookValue(value)
	if err != nil {
		return nil, err
	}
	return s.updateWebhookConfig(ctx, id, clientToken, instanceToken, func(cfg *WebhookConfig) error {
		cfg.ReceivedDeliveryURL = strPtr(normalized)
		return nil
	})
}

func (s *Service) UpdateWebhookMessageStatus(ctx context.Context, id uuid.UUID, clientToken, instanceToken, value string) (*WebhookSettings, error) {
	normalized, err := normalizeWebhookValue(value)
	if err != nil {
		return nil, err
	}
	return s.updateWebhookConfig(ctx, id, clientToken, instanceToken, func(cfg *WebhookConfig) error {
		cfg.MessageStatusURL = strPtr(normalized)
		return nil
	})
}

func (s *Service) UpdateWebhookDisconnected(ctx context.Context, id uuid.UUID, clientToken, instanceToken, value string) (*WebhookSettings, error) {
	normalized, err := normalizeWebhookValue(value)
	if err != nil {
		return nil, err
	}
	return s.updateWebhookConfig(ctx, id, clientToken, instanceToken, func(cfg *WebhookConfig) error {
		cfg.DisconnectedURL = strPtr(normalized)
		return nil
	})
}

func (s *Service) UpdateWebhookConnected(ctx context.Context, id uuid.UUID, clientToken, instanceToken, value string) (*WebhookSettings, error) {
	normalized, err := normalizeWebhookValue(value)
	if err != nil {
		return nil, err
	}
	return s.updateWebhookConfig(ctx, id, clientToken, instanceToken, func(cfg *WebhookConfig) error {
		cfg.ConnectedURL = strPtr(normalized)
		return nil
	})
}

func (s *Service) UpdateWebhookChatPresence(ctx context.Context, id uuid.UUID, clientToken, instanceToken, value string) (*WebhookSettings, error) {
	normalized, err := normalizeWebhookValue(value)
	if err != nil {
		return nil, err
	}
	return s.updateWebhookConfig(ctx, id, clientToken, instanceToken, func(cfg *WebhookConfig) error {
		cfg.ChatPresenceURL = strPtr(normalized)
		return nil
	})
}

func (s *Service) UpdateNotifySentByMe(ctx context.Context, id uuid.UUID, clientToken, instanceToken string, notify bool) (*WebhookSettings, error) {
	return s.updateWebhookConfig(ctx, id, clientToken, instanceToken, func(cfg *WebhookConfig) error {
		cfg.NotifySentByMe = notify
		return nil
	})
}

func (s *Service) UpdateEveryWebhooks(ctx context.Context, id uuid.UUID, clientToken, instanceToken, value string, notify *bool) (*WebhookSettings, error) {
	normalized, err := normalizeWebhookValue(value)
	if err != nil {
		return nil, err
	}
	return s.updateWebhookConfig(ctx, id, clientToken, instanceToken, func(cfg *WebhookConfig) error {
		cfg.DeliveryURL = strPtr(normalized)
		cfg.ReceivedURL = strPtr(normalized)
		cfg.ReceivedDeliveryURL = strPtr(normalized)
		cfg.MessageStatusURL = strPtr(normalized)
		cfg.DisconnectedURL = strPtr(normalized)
		cfg.ChatPresenceURL = strPtr(normalized)
		cfg.ConnectedURL = strPtr(normalized)
		if notify != nil {
			cfg.NotifySentByMe = *notify
		}
		return nil
	})
}

type PartnerCreateParams struct {
	Name                        string
	SessionName                 string
	DeliveryCallbackURL         *string
	ReceivedCallbackURL         *string
	ReceivedDeliveryCallbackURL *string
	DisconnectedCallbackURL     *string
	ConnectedCallbackURL        *string
	MessageStatusCallbackURL    *string
	ChatPresenceCallbackURL     *string
	NotifySentByMe              bool
	CallRejectAuto              *bool
	CallRejectMessage           *string
	AutoReadMessage             *bool
	IsDevice                    bool
	BusinessDevice              bool
}

func (s *Service) CreatePartnerInstance(ctx context.Context, params PartnerCreateParams) (*Instance, error) {
	if params.SessionName == "" {
		params.SessionName = params.Name
	}
	inst := Instance{
		ID:                 uuid.New(),
		Name:               params.Name,
		SessionName:        params.SessionName,
		ClientToken:        uuid.NewString(),
		InstanceToken:      uuid.NewString(),
		IsDevice:           params.IsDevice,
		BusinessDevice:     params.BusinessDevice,
		SubscriptionActive: false,
		CallRejectAuto:     derefBool(params.CallRejectAuto),
		CallRejectMessage:  params.CallRejectMessage,
		AutoReadMessage:    derefBool(params.AutoReadMessage),
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	if inst.CallRejectMessage != nil && strings.TrimSpace(*inst.CallRejectMessage) == "" {
		inst.CallRejectMessage = nil
	}

	delivery, err := s.normalizeWebhookPointer(params.DeliveryCallbackURL)
	if err != nil {
		return nil, err
	}
	received, err := s.normalizeWebhookPointer(params.ReceivedCallbackURL)
	if err != nil {
		return nil, err
	}
	receivedDelivery, err := s.normalizeWebhookPointer(params.ReceivedDeliveryCallbackURL)
	if err != nil {
		return nil, err
	}
	disconnected, err := s.normalizeWebhookPointer(params.DisconnectedCallbackURL)
	if err != nil {
		return nil, err
	}
	connected, err := s.normalizeWebhookPointer(params.ConnectedCallbackURL)
	if err != nil {
		return nil, err
	}
	messageStatus, err := s.normalizeWebhookPointer(params.MessageStatusCallbackURL)
	if err != nil {
		return nil, err
	}
	chatPresence, err := s.normalizeWebhookPointer(params.ChatPresenceCallbackURL)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Insert(ctx, &inst); err != nil {
		return nil, err
	}

	cfg := WebhookConfig{
		InstanceID:          inst.ID,
		DeliveryURL:         delivery,
		ReceivedURL:         received,
		ReceivedDeliveryURL: receivedDelivery,
		MessageStatusURL:    messageStatus,
		DisconnectedURL:     disconnected,
		ChatPresenceURL:     chatPresence,
		ConnectedURL:        connected,
		NotifySentByMe:      params.NotifySentByMe,
	}
	if err := s.repo.UpsertWebhookConfig(ctx, cfg); err != nil {
		return nil, err
	}
	inst.Webhooks = toWebhookSettings(&cfg)

	go func(instance Instance) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, _, err := s.registry.EnsureClient(ctx, toInstanceInfo(instance)); err != nil {
			s.log.Error("ensure client", slog.String("instanceId", instance.ID.String()), slog.String("error", err.Error()))
		}
	}(inst)

	if inst.IsDevice {
		inst.Middleware = "mobile"
	} else {
		inst.Middleware = "web"
	}

	return &inst, nil
}

func (s *Service) SubscribeInstance(ctx context.Context, id uuid.UUID, instanceToken string) error {
	if err := s.repo.VerifyToken(ctx, id, instanceToken); err != nil {
		return err
	}
	return s.repo.UpdateSubscription(ctx, id, true)
}

func (s *Service) CancelInstance(ctx context.Context, id uuid.UUID, instanceToken string) error {
	if err := s.repo.VerifyToken(ctx, id, instanceToken); err != nil {
		return err
	}
	if err := s.repo.UpdateSubscription(ctx, id, false); err != nil {
		return err
	}
	cancelCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if inst, err := s.repo.GetByID(ctx, id); err == nil {
		_ = s.registry.Disconnect(cancelCtx, toInstanceInfo(*inst))
	}
	return nil
}

func (s *Service) ListInstances(ctx context.Context, filter ListFilter) (ListResult, error) {
	instances, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return ListResult{}, err
	}
	for i := range instances {
		if instances[i].Webhooks == nil {
			instances[i].Webhooks = &WebhookSettings{NotifySentByMe: false}
		}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 15
	}
	return ListResult{
		Data:     instances,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) updateWebhookConfig(ctx context.Context, id uuid.UUID, clientToken, instanceToken string, mutate func(*WebhookConfig) error) (*WebhookSettings, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return nil, ErrUnauthorized
	}
	cfg, err := s.repo.GetWebhookConfig(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := mutate(cfg); err != nil {
		return nil, err
	}
	if err := s.repo.UpsertWebhookConfig(ctx, *cfg); err != nil {
		return nil, err
	}
	return toWebhookSettings(cfg), nil
}

func toWebhookSettings(cfg *WebhookConfig) *WebhookSettings {
	if cfg == nil {
		return nil
	}
	return &WebhookSettings{
		DeliveryURL:         cfg.DeliveryURL,
		ReceivedURL:         cfg.ReceivedURL,
		ReceivedDeliveryURL: cfg.ReceivedDeliveryURL,
		MessageStatusURL:    cfg.MessageStatusURL,
		DisconnectedURL:     cfg.DisconnectedURL,
		ChatPresenceURL:     cfg.ChatPresenceURL,
		ConnectedURL:        cfg.ConnectedURL,
		NotifySentByMe:      cfg.NotifySentByMe,
	}
}

func normalizeWebhookValue(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", ErrMissingWebhookValue
	}
	if !strings.HasPrefix(strings.ToLower(trimmed), "https://") && !strings.HasPrefix(strings.ToLower(trimmed), "http://") {
		return "", ErrInvalidWebhookURL
	}
	return trimmed, nil
}

func (s *Service) normalizeWebhookPointer(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	if !strings.HasPrefix(strings.ToLower(trimmed), "https://") && !strings.HasPrefix(strings.ToLower(trimmed), "http://") {
		return nil, ErrInvalidWebhookURL
	}
	copyVal := trimmed
	return &copyVal, nil
}

func strPtr(v string) *string {
	copyVal := v
	return &copyVal
}

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func toInstanceInfo(inst Instance) whatsmeow.InstanceInfo {
	return whatsmeow.InstanceInfo{
		ID:            inst.ID,
		Name:          inst.Name,
		SessionName:   inst.SessionName,
		ClientToken:   inst.ClientToken,
		InstanceToken: inst.InstanceToken,
		StoreJID:      inst.StoreJID,
	}
}

func ensureActive(inst *Instance) error {
	if inst == nil {
		return ErrInstanceNotFound
	}
	if !inst.SubscriptionActive {
		return ErrInstanceInactive
	}
	return nil
}
