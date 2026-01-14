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
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/whatsmeow"
	whatsstore "go.mau.fi/whatsmeow/store"
	watypes "go.mau.fi/whatsmeow/types"
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
	repo            *Repository
	registry        *whatsmeow.ClientRegistry
	clientAuthToken string
	log             *slog.Logger
}

func NewService(repo *Repository, registry *whatsmeow.ClientRegistry, clientAuthToken string, log *slog.Logger) *Service {
	return &Service{
		repo:            repo,
		registry:        registry,
		clientAuthToken: clientAuthToken,
		log:             log,
	}
}

func (s *Service) Create(ctx context.Context, params CreateParams) (*Instance, error) {
	inst := Instance{
		ID:              uuid.New(),
		Name:            params.Name,
		SessionName:     params.Name,
		InstanceToken:   uuid.NewString(),
		IsDevice:        false,
		BusinessDevice:  false,
		CallRejectAuto:  false,
		AutoReadMessage: false,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	inst.Middleware = "web"

	ctx = logging.WithAttrs(ctx, slog.String("instance_id", inst.ID.String()))
	logger := logging.ContextLogger(ctx, s.log)

	if err := s.repo.Insert(ctx, &inst); err != nil {
		return nil, err
	}

	go func(instance Instance, baseLogger *slog.Logger) {
		bgCtx := logging.WithLogger(context.Background(), baseLogger)
		ctx, cancel := context.WithTimeout(bgCtx, 30*time.Second)
		defer cancel()
		if _, _, err := s.registry.EnsureClient(ctx, toInstanceInfo(instance)); err != nil {
			baseLogger.Error("ensure client",
				slog.String("instance_id", instance.ID.String()),
				slog.String("error", err.Error()))
		}
	}(inst, logger)

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
	connectionStatus := inst.ConnectionStatus
	if snapshot.ConnectionStatus != "" {
		connectionStatus = snapshot.ConnectionStatus
	} else if connectionStatus == "" {
		if snapshot.Connected {
			connectionStatus = "connected"
		} else {
			connectionStatus = "disconnected"
		}
	}

	lastConnected := inst.LastConnectedAt
	if snapshot.LastConnected != nil {
		lastConnected = snapshot.LastConnected
	}

	storeJID := snapshot.StoreJID
	if storeJID == nil {
		storeJID = inst.StoreJID
	}

	worker := snapshot.WorkerID
	if worker == "" && inst.WorkerID != nil {
		worker = *inst.WorkerID
	}

	smartphoneConnected := false
	if storeJID != nil && *storeJID != "" {
		smartphoneConnected = snapshot.Connected
	}

	statusError := deriveStatusError(snapshot.Connected, connectionStatus, storeJID)

	return &Status{
		Connected:           snapshot.Connected,
		ConnectionStatus:    connectionStatus,
		StoreJID:            storeJID,
		LastConnected:       lastConnected,
		InstanceID:          inst.ID,
		AutoReconnect:       snapshot.AutoReconnect,
		WorkerAssigned:      worker,
		SubscriptionActive:  inst.SubscriptionActive,
		Error:               statusError,
		SmartphoneConnected: smartphoneConnected,
	}, nil
}

func deriveStatusError(connected bool, status string, storeJID *string) string {
	if connected {
		return ""
	}

	normalized := strings.ToLower(status)
	switch {
	case strings.HasPrefix(normalized, "logged_out"):
		return "You need to restore the session."
	case storeJID != nil && *storeJID != "":
		return "You are already connected."
	default:
		return "You are not connected."
	}
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

	if err := s.registry.Disconnect(ctx, toInstanceInfo(*inst)); err != nil {
		return err
	}

	if err := s.repo.UpdateStoreJID(ctx, id, nil); err != nil {
		logger := logging.ContextLogger(ctx, s.log)
		logger.Error("failed to clear store jid after disconnect",
			slog.String("instance_id", id.String()),
			slog.String("error", err.Error()))
	}

	return nil
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

func (s *Service) GetDevice(ctx context.Context, id uuid.UUID, clientToken, instanceToken string) (*DeviceResponse, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return nil, ErrUnauthorized
	}

	device := &DeviceResponse{
		Phone: "",
		Name:  inst.Name,
		Device: DeviceMetadata{
			SessionName: inst.SessionName,
			WAVersion:   whatsstore.GetWAVersion().String(),
		},
		SessionID:  0,
		IsBusiness: inst.BusinessDevice,
	}

	if ua := whatsstore.BaseClientPayload.GetUserAgent(); ua != nil {
		device.Device.MCC = ua.GetMcc()
		device.Device.MNC = ua.GetMnc()
		device.Device.OSVersion = ua.GetOsVersion()
		device.Device.DeviceManufacturer = ua.GetManufacturer()
		device.Device.DeviceModel = ua.GetDevice()
		device.Device.OSBuildNumber = ua.GetOsBuildNumber()
	}

	if inst.StoreJID != nil {
		if parsed, phone := parsePhoneFromJID(*inst.StoreJID); phone != "" {
			device.Phone = phone
			device.SessionID = int(parsed.Device)
		}
	}

	if client, ok := s.registry.GetClient(inst.ID); ok && client != nil {
		s.hydrateDeviceFromClient(ctx, device, client, inst)
	}

	if strings.TrimSpace(device.Device.DeviceModel) == "" {
		device.Device.DeviceModel = inst.SessionName
	}
	if strings.TrimSpace(device.Device.Platform) == "" {
		device.Device.Platform = deriveOriginalDevice(inst, "")
	}
	device.OriginalDevice = deriveOriginalDevice(inst, device.Device.Platform)

	return device, nil
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

	logger := logging.ContextLogger(ctx, s.log)
	cleaned := make([]uuid.UUID, 0)
	for _, link := range links {
		instanceLogger := logger.With(
			slog.String("instance_id", link.ID.String()),
			slog.String("store_jid", link.StoreJID),
		)
		loopCtx := logging.WithAttrs(ctx, slog.String("instance_id", link.ID.String()))
		exists, checkErr := s.registry.HasStoreDevice(loopCtx, link.StoreJID)
		if checkErr != nil {
			instanceLogger.Error("check store device", slog.String("error", checkErr.Error()))
		}
		if checkErr != nil || !exists {
			s.registry.RemoveClient(link.ID, "reconcile_missing_store")
			if err := s.repo.UpdateStoreJID(loopCtx, link.ID, nil); err != nil {
				instanceLogger.Error("clear store jid", slog.String("error", err.Error()))
				continue
			}
			instanceLogger.Info("reconciled missing store")
			cleaned = append(cleaned, link.ID)
		}
	}
	return cleaned, nil
}

func (s *Service) tokensMatch(inst *Instance, clientToken, instanceToken string) bool {
	if inst == nil {
		return false
	}
	// Validate against global client token (from env) and per-instance token
	return s.clientAuthToken == clientToken && inst.InstanceToken == instanceToken
}

// GetByID retrieves an instance by its ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Instance, error) {
	return s.repo.GetByID(ctx, id)
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

func (s *Service) UpdateCallRejectAuto(ctx context.Context, id uuid.UUID, clientToken, instanceToken string, value bool) error {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return ErrUnauthorized
	}
	return s.repo.UpdateCallRejectAuto(ctx, id, value)
}

func (s *Service) UpdateCallRejectMessage(ctx context.Context, id uuid.UUID, clientToken, instanceToken string, value *string) error {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return ErrUnauthorized
	}
	return s.repo.UpdateCallRejectMessage(ctx, id, value)
}

func (s *Service) UpdateAutoReadMessage(ctx context.Context, id uuid.UUID, clientToken, instanceToken string, value bool) error {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return ErrUnauthorized
	}
	return s.repo.UpdateAutoReadMessage(ctx, id, value)
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

	ctx = logging.WithAttrs(ctx, slog.String("instance_id", inst.ID.String()))
	logger := logging.ContextLogger(ctx, s.log)

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

	go func(instance Instance, baseLogger *slog.Logger) {
		bgCtx := logging.WithLogger(context.Background(), baseLogger)
		ctx, cancel := context.WithTimeout(bgCtx, 30*time.Second)
		defer cancel()
		if _, _, err := s.registry.EnsureClient(ctx, toInstanceInfo(instance)); err != nil {
			baseLogger.Error("ensure client",
				slog.String("instance_id", instance.ID.String()),
				slog.String("error", err.Error()))
		}
	}(inst, logger)

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

func (s *Service) DeleteInstance(ctx context.Context, id uuid.UUID) error {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	deleteCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s.registry.ResetClient(id, "instance_deleted")

	if err := s.repo.Delete(deleteCtx, id); err != nil {
		return fmt.Errorf("delete instance from database: %w", err)
	}

	logger := logging.ContextLogger(ctx, s.log)
	logger.Info("instance permanently deleted",
		slog.String("instance_id", id.String()),
		slog.String("name", inst.Name))

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
		// Populate connection status merging live registry snapshot with persisted state
		snapshot := s.registry.Status(toInstanceInfo(instances[i]))
		if snapshot.StoreJID != nil {
			instances[i].StoreJID = snapshot.StoreJID
		}
		instances[i].WhatsappConnected = snapshot.Connected
		instances[i].PhoneConnected = snapshot.Connected && snapshot.StoreJID != nil && *snapshot.StoreJID != ""

		if snapshot.ConnectionStatus != "" {
			instances[i].ConnectionStatus = snapshot.ConnectionStatus
		} else if instances[i].ConnectionStatus == "" {
			instances[i].ConnectionStatus = "disconnected"
		}

		if snapshot.LastConnected != nil {
			instances[i].LastConnectedAt = snapshot.LastConnected
		}

		if snapshot.WorkerID != "" {
			instances[i].WorkerID = strPtr(snapshot.WorkerID)
		} else if instances[i].WorkerID == nil {
			instances[i].WorkerID = nil
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
	// Allow empty value to clear the webhook
	if trimmed == "" {
		return "", nil
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
	if v == "" {
		return nil
	}
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

func (s *Service) hydrateDeviceFromClient(ctx context.Context, device *DeviceResponse, client *whatsmeowpkg.Client, inst *Instance) {
	if client == nil || device == nil {
		return
	}
	if client.Store != nil {
		if client.Store.ID != nil {
			if phone := phoneFromJID(*client.Store.ID); phone != "" {
				device.Phone = phone
			}
			device.SessionID = int(client.Store.ID.Device)
		}
		if push := strings.TrimSpace(client.Store.PushName); push != "" {
			device.Name = push
		}
		if platform := strings.TrimSpace(client.Store.Platform); platform != "" {
			device.Device.Platform = strings.ToLower(platform)
			if strings.TrimSpace(device.Device.DeviceModel) == "" {
				device.Device.DeviceModel = platform
			}
		}
		if client.Store.BusinessName != "" {
			device.IsBusiness = true
		}
	}
	if img := s.selfProfilePictureURL(ctx, client, inst); img != "" {
		device.ImgURL = &img
	}
}

func (s *Service) selfProfilePictureURL(ctx context.Context, client *whatsmeowpkg.Client, inst *Instance) string {
	if client == nil || client.Store == nil {
		return ""
	}
	var jid watypes.JID
	if client.Store.ID != nil {
		jid = client.Store.ID.ToNonAD()
	} else if inst != nil && inst.StoreJID != nil {
		parsed, err := watypes.ParseJID(*inst.StoreJID)
		if err == nil {
			jid = parsed.ToNonAD()
		}
	}
	if jid == watypes.EmptyJID {
		return ""
	}
	info, err := client.GetProfilePictureInfo(ctx, jid, nil)
	if err != nil {
		if errors.Is(err, whatsmeowpkg.ErrProfilePictureNotSet) || errors.Is(err, whatsmeowpkg.ErrProfilePictureUnauthorized) {
			return ""
		}
		logger := logging.ContextLogger(ctx, s.log)
		logger.Debug("failed to fetch profile picture",
			slog.String("instance_id", inst.ID.String()),
			slog.String("error", err.Error()))
		return ""
	}
	return strings.TrimSpace(info.URL)
}

func parsePhoneFromJID(raw string) (watypes.JID, string) {
	jid, err := watypes.ParseJID(raw)
	if err != nil {
		return watypes.JID{}, ""
	}
	return jid, phoneFromJID(jid)
}

func phoneFromJID(jid watypes.JID) string {
	return strings.TrimSpace(jid.User)
}

func deriveOriginalDevice(inst *Instance, candidate string) string {
	if trimmed := strings.TrimSpace(candidate); trimmed != "" {
		return strings.ToLower(trimmed)
	}
	if inst != nil {
		if inst.IsDevice {
			return "smba"
		}
		return "smbi"
	}
	return ""
}
