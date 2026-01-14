//go:build ignore
// +build ignore

// NOTE: This file contains message queue service implementation with correct metrics.
// Once the supporting types (SendTextRequest, etc.) are implemented in types.go,
// remove the build ignore tags from both files.
//
// Metrics fixed in this file:
// - MessageQueueEnqueued (was MessagesQueued)
// - MessageQueueDuration (was MessageQueueLatency)

package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/types"

	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/messages/queue"
	"go.mau.fi/whatsmeow/api/internal/observability"
	"go.mau.fi/whatsmeow/api/internal/whatsmeow"
)

// Service handles message operations
type Service struct {
	instancesService *instances.Service
	queueRepo        queue.Repository
	registry         *whatsmeow.ClientRegistry
	validator        *Validator
	metrics          *observability.Metrics
	log              *slog.Logger
}

// NewService creates a new messages service
func NewService(
	instancesService *instances.Service,
	queueRepo queue.Repository,
	registry *whatsmeow.ClientRegistry,
	metrics *observability.Metrics,
	log *slog.Logger,
) *Service {
	return &Service{
		instancesService: instancesService,
		queueRepo:        queueRepo,
		registry:         registry,
		validator:        NewValidator(),
		metrics:          metrics,
		log:              log,
	}
}

// SendText sends a text message
func (s *Service) SendText(ctx context.Context, instanceID uuid.UUID, clientToken, instanceToken string, req *SendTextRequest) (*SendMessageResult, error) {
	start := time.Now()
	logger := logging.ContextLogger(ctx, s.log)

	logger = logger.With(
		slog.String("instance_id", instanceID.String()),
		slog.String("operation", "send_text"),
		slog.String("phone", req.GetPhone()),
	)
	ctx = logging.WithLogger(ctx, logger)

	logger.Info("processing send text request")

	// Validate request
	if err := s.validator.ValidateSendTextRequest(req); err != nil {
		logger.Warn("validation failed", slog.String("error", err.Error()))
		s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "text", "validation_failed").Inc()
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify instance exists and tokens match
	inst, err := s.instancesService.GetByID(ctx, instanceID)
	if err != nil {
		logger.Error("failed to get instance", slog.String("error", err.Error()))
		return nil, fmt.Errorf("instance not found: %w", err)
	}

	if !s.tokensMatch(inst, clientToken, instanceToken) {
		logger.Warn("unauthorized access attempt")
		return nil, fmt.Errorf("unauthorized")
	}

	// Check if instance is connected
	if !s.registry.IsConnected(instanceID) {
		logger.Warn("instance not connected")
		return nil, fmt.Errorf("instance not connected")
	}

	// Normalize phone to JID format
	jid, err := s.normalizePhoneToJID(req.GetPhone())
	if err != nil {
		logger.Error("failed to normalize phone", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Generate unique message ID (zaapId)
	zaapID := uuid.New().String()

	// Create queue message
	args := queue.SendMessageArgs{
		ZaapID:      zaapID,
		InstanceID:  instanceID,
		Phone:       jid.String(),
		MessageType: queue.MessageTypeText,
		TextContent: &queue.TextMessage{
			Message: req.GetMessage(),
		},
		DelayMessage: int64(req.GetDelay()),
		EnqueuedAt:   time.Now(),
	}

	// Serialize payload
	payload, err := json.Marshal(args)
	if err != nil {
		logger.Error("failed to serialize message", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Enqueue message
	queuedMsg := &QueuedMessage{
		ID:          uuid.New(),
		InstanceID:  instanceID,
		MessageID:   zaapID,
		Type:        string(queue.MessageTypeText),
		Phone:       jid.String(),
		Payload:     payload,
		Status:      "pending",
		Attempts:    0,
		MaxAttempts: 3,
		ScheduledAt: s.calculateScheduledAt(req.GetDelay()),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.queueRepo.Enqueue(ctx, queuedMsg); err != nil {
		logger.Error("failed to enqueue message", slog.String("error", err.Error()))
		s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "text", "enqueue_failed").Inc()
		return nil, fmt.Errorf("failed to enqueue message: %w", err)
	}

	// Update metrics
	duration := time.Since(start)
	s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "text", "success").Inc()
	s.metrics.MessageQueueDuration.WithLabelValues(instanceID.String(), "text").Observe(duration.Seconds())

	logger.Info("text message queued successfully",
		slog.String("zaap_id", zaapID),
		slog.Duration("duration", duration))

	return &SendMessageResult{
		ZaapID:    zaapID,
		MessageID: zaapID,
		ID:        zaapID,
		Status:    "queued",
		QueuedAt:  queuedMsg.CreatedAt,
	}, nil
}

// SendImage sends an image message
func (s *Service) SendImage(ctx context.Context, instanceID uuid.UUID, clientToken, instanceToken string, req *SendImageRequest) (*SendMessageResult, error) {
	start := time.Now()
	logger := logging.ContextLogger(ctx, s.log)

	logger = logger.With(
		slog.String("instance_id", instanceID.String()),
		slog.String("operation", "send_image"),
		slog.String("phone", req.GetPhone()),
	)
	ctx = logging.WithLogger(ctx, logger)

	logger.Info("processing send image request")

	// Validate request
	if err := s.validator.ValidateSendImageRequest(req); err != nil {
		logger.Warn("validation failed", slog.String("error", err.Error()))
		s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "image", "validation_failed").Inc()
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify instance exists and tokens match
	inst, err := s.instancesService.GetByID(ctx, instanceID)
	if err != nil {
		logger.Error("failed to get instance", slog.String("error", err.Error()))
		return nil, fmt.Errorf("instance not found: %w", err)
	}

	if !s.tokensMatch(inst, clientToken, instanceToken) {
		logger.Warn("unauthorized access attempt")
		return nil, fmt.Errorf("unauthorized")
	}

	// Check if instance is connected
	if !s.registry.IsConnected(instanceID) {
		logger.Warn("instance not connected")
		return nil, fmt.Errorf("instance not connected")
	}

	// Normalize phone to JID format
	jid, err := s.normalizePhoneToJID(req.GetPhone())
	if err != nil {
		logger.Error("failed to normalize phone", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Generate unique message ID (zaapId)
	zaapID := uuid.New().String()

	// Prepare caption
	var caption *string
	if req.GetCaption() != "" {
		c := req.GetCaption()
		caption = &c
	}

	// Create queue message
	args := queue.SendMessageArgs{
		ZaapID:      zaapID,
		InstanceID:  instanceID,
		Phone:       jid.String(),
		MessageType: queue.MessageTypeImage,
		ImageContent: &queue.MediaMessage{
			MediaURL: req.GetImage(),
			Caption:  caption,
		},
		DelayMessage: int64(req.GetDelay()),
		EnqueuedAt:   time.Now(),
	}

	// Serialize payload
	payload, err := json.Marshal(args)
	if err != nil {
		logger.Error("failed to serialize message", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Enqueue message
	queuedMsg := &QueuedMessage{
		ID:          uuid.New(),
		InstanceID:  instanceID,
		MessageID:   zaapID,
		Type:        string(queue.MessageTypeImage),
		Phone:       jid.String(),
		Payload:     payload,
		Status:      "pending",
		Attempts:    0,
		MaxAttempts: 3,
		ScheduledAt: s.calculateScheduledAt(req.GetDelay()),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.queueRepo.Enqueue(ctx, queuedMsg); err != nil {
		logger.Error("failed to enqueue message", slog.String("error", err.Error()))
		s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "image", "enqueue_failed").Inc()
		return nil, fmt.Errorf("failed to enqueue message: %w", err)
	}

	// Update metrics
	duration := time.Since(start)
	s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "image", "success").Inc()
	s.metrics.MessageQueueDuration.WithLabelValues(instanceID.String(), "image").Observe(duration.Seconds())

	logger.Info("image message queued successfully",
		slog.String("zaap_id", zaapID),
		slog.Duration("duration", duration))

	return &SendMessageResult{
		ZaapID:    zaapID,
		MessageID: zaapID,
		ID:        zaapID,
		Status:    "queued",
		QueuedAt:  queuedMsg.CreatedAt,
	}, nil
}

// SendAudio sends an audio message
func (s *Service) SendAudio(ctx context.Context, instanceID uuid.UUID, clientToken, instanceToken string, req *SendAudioRequest) (*SendMessageResult, error) {
	start := time.Now()
	logger := logging.ContextLogger(ctx, s.log)

	logger = logger.With(
		slog.String("instance_id", instanceID.String()),
		slog.String("operation", "send_audio"),
		slog.String("phone", req.GetPhone()),
	)
	ctx = logging.WithLogger(ctx, logger)

	logger.Info("processing send audio request")

	// Validate request
	if err := s.validator.ValidateSendAudioRequest(req); err != nil {
		logger.Warn("validation failed", slog.String("error", err.Error()))
		s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "audio", "validation_failed").Inc()
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify instance exists and tokens match
	inst, err := s.instancesService.GetByID(ctx, instanceID)
	if err != nil {
		logger.Error("failed to get instance", slog.String("error", err.Error()))
		return nil, fmt.Errorf("instance not found: %w", err)
	}

	if !s.tokensMatch(inst, clientToken, instanceToken) {
		logger.Warn("unauthorized access attempt")
		return nil, fmt.Errorf("unauthorized")
	}

	// Check if instance is connected
	if !s.registry.IsConnected(instanceID) {
		logger.Warn("instance not connected")
		return nil, fmt.Errorf("instance not connected")
	}

	// Normalize phone to JID format
	jid, err := s.normalizePhoneToJID(req.GetPhone())
	if err != nil {
		logger.Error("failed to normalize phone", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Generate unique message ID (zaapId)
	zaapID := uuid.New().String()

	// Create queue message
	args := queue.SendMessageArgs{
		ZaapID:      zaapID,
		InstanceID:  instanceID,
		Phone:       jid.String(),
		MessageType: queue.MessageTypeAudio,
		AudioContent: &queue.MediaMessage{
			MediaURL: req.GetAudio(),
		},
		DelayMessage: int64(req.GetDelay()),
		EnqueuedAt:   time.Now(),
	}

	// Serialize payload
	payload, err := json.Marshal(args)
	if err != nil {
		logger.Error("failed to serialize message", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Enqueue message
	queuedMsg := &QueuedMessage{
		ID:          uuid.New(),
		InstanceID:  instanceID,
		MessageID:   zaapID,
		Type:        string(queue.MessageTypeAudio),
		Phone:       jid.String(),
		Payload:     payload,
		Status:      "pending",
		Attempts:    0,
		MaxAttempts: 3,
		ScheduledAt: s.calculateScheduledAt(req.GetDelay()),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.queueRepo.Enqueue(ctx, queuedMsg); err != nil {
		logger.Error("failed to enqueue message", slog.String("error", err.Error()))
		s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "audio", "enqueue_failed").Inc()
		return nil, fmt.Errorf("failed to enqueue message: %w", err)
	}

	// Update metrics
	duration := time.Since(start)
	s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "audio", "success").Inc()
	s.metrics.MessageQueueDuration.WithLabelValues(instanceID.String(), "audio").Observe(duration.Seconds())

	logger.Info("audio message queued successfully",
		slog.String("zaap_id", zaapID),
		slog.Duration("duration", duration))

	return &SendMessageResult{
		ZaapID:    zaapID,
		MessageID: zaapID,
		ID:        zaapID,
		Status:    "queued",
		QueuedAt:  queuedMsg.CreatedAt,
	}, nil
}

// SendVideo sends a video message
func (s *Service) SendVideo(ctx context.Context, instanceID uuid.UUID, clientToken, instanceToken string, req *SendVideoRequest) (*SendMessageResult, error) {
	start := time.Now()
	logger := logging.ContextLogger(ctx, s.log)

	logger = logger.With(
		slog.String("instance_id", instanceID.String()),
		slog.String("operation", "send_video"),
		slog.String("phone", req.GetPhone()),
	)
	ctx = logging.WithLogger(ctx, logger)

	logger.Info("processing send video request")

	// Validate request
	if err := s.validator.ValidateSendVideoRequest(req); err != nil {
		logger.Warn("validation failed", slog.String("error", err.Error()))
		s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "video", "validation_failed").Inc()
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify instance exists and tokens match
	inst, err := s.instancesService.GetByID(ctx, instanceID)
	if err != nil {
		logger.Error("failed to get instance", slog.String("error", err.Error()))
		return nil, fmt.Errorf("instance not found: %w", err)
	}

	if !s.tokensMatch(inst, clientToken, instanceToken) {
		logger.Warn("unauthorized access attempt")
		return nil, fmt.Errorf("unauthorized")
	}

	// Check if instance is connected
	if !s.registry.IsConnected(instanceID) {
		logger.Warn("instance not connected")
		return nil, fmt.Errorf("instance not connected")
	}

	// Normalize phone to JID format
	jid, err := s.normalizePhoneToJID(req.GetPhone())
	if err != nil {
		logger.Error("failed to normalize phone", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Generate unique message ID (zaapId)
	zaapID := uuid.New().String()

	// Prepare caption
	var caption *string
	if req.GetCaption() != "" {
		c := req.GetCaption()
		caption = &c
	}

	// Create queue message
	args := queue.SendMessageArgs{
		ZaapID:      zaapID,
		InstanceID:  instanceID,
		Phone:       jid.String(),
		MessageType: queue.MessageTypeVideo,
		VideoContent: &queue.MediaMessage{
			MediaURL: req.GetVideo(),
			Caption:  caption,
		},
		DelayMessage: int64(req.GetDelay()),
		EnqueuedAt:   time.Now(),
	}

	// Serialize payload
	payload, err := json.Marshal(args)
	if err != nil {
		logger.Error("failed to serialize message", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Enqueue message
	queuedMsg := &QueuedMessage{
		ID:          uuid.New(),
		InstanceID:  instanceID,
		MessageID:   zaapID,
		Type:        string(queue.MessageTypeVideo),
		Phone:       jid.String(),
		Payload:     payload,
		Status:      "pending",
		Attempts:    0,
		MaxAttempts: 3,
		ScheduledAt: s.calculateScheduledAt(req.GetDelay()),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.queueRepo.Enqueue(ctx, queuedMsg); err != nil {
		logger.Error("failed to enqueue message", slog.String("error", err.Error()))
		s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "video", "enqueue_failed").Inc()
		return nil, fmt.Errorf("failed to enqueue message: %w", err)
	}

	// Update metrics
	duration := time.Since(start)
	s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "video", "success").Inc()
	s.metrics.MessageQueueDuration.WithLabelValues(instanceID.String(), "video").Observe(duration.Seconds())

	logger.Info("video message queued successfully",
		slog.String("zaap_id", zaapID),
		slog.Duration("duration", duration))

	return &SendMessageResult{
		ZaapID:    zaapID,
		MessageID: zaapID,
		ID:        zaapID,
		Status:    "queued",
		QueuedAt:  queuedMsg.CreatedAt,
	}, nil
}

// SendSticker sends a sticker message
func (s *Service) SendSticker(ctx context.Context, instanceID uuid.UUID, clientToken, instanceToken string, req *SendStickerRequest) (*SendMessageResult, error) {
	start := time.Now()
	logger := logging.ContextLogger(ctx, s.log)

	logger = logger.With(
		slog.String("instance_id", instanceID.String()),
		slog.String("operation", "send_sticker"),
		slog.String("phone", req.GetPhone()),
	)
	ctx = logging.WithLogger(ctx, logger)

	logger.Info("processing send sticker request")

	// Validate request
	if err := s.validator.ValidateSendStickerRequest(req); err != nil {
		logger.Warn("validation failed", slog.String("error", err.Error()))
		s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "sticker", "validation_failed").Inc()
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify instance exists and tokens match
	inst, err := s.instancesService.GetByID(ctx, instanceID)
	if err != nil {
		logger.Error("failed to get instance", slog.String("error", err.Error()))
		return nil, fmt.Errorf("instance not found: %w", err)
	}

	if !s.tokensMatch(inst, clientToken, instanceToken) {
		logger.Warn("unauthorized access attempt")
		return nil, fmt.Errorf("unauthorized")
	}

	// Check if instance is connected
	if !s.registry.IsConnected(instanceID) {
		logger.Warn("instance not connected")
		return nil, fmt.Errorf("instance not connected")
	}

	// Normalize phone to JID format
	jid, err := s.normalizePhoneToJID(req.GetPhone())
	if err != nil {
		logger.Error("failed to normalize phone", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Generate unique message ID (zaapId)
	zaapID := uuid.New().String()

	// Create queue message - stickers use image message type with special handling
	args := queue.SendMessageArgs{
		ZaapID:      zaapID,
		InstanceID:  instanceID,
		Phone:       jid.String(),
		MessageType: queue.MessageTypeImage, // Stickers are handled as images
		ImageContent: &queue.MediaMessage{
			MediaURL: req.GetSticker(),
		},
		DelayMessage: int64(req.GetDelay()),
		EnqueuedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"is_sticker": true,
		},
	}

	// Serialize payload
	payload, err := json.Marshal(args)
	if err != nil {
		logger.Error("failed to serialize message", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Enqueue message
	queuedMsg := &QueuedMessage{
		ID:          uuid.New(),
		InstanceID:  instanceID,
		MessageID:   zaapID,
		Type:        "sticker", // Use sticker type for tracking
		Phone:       jid.String(),
		Payload:     payload,
		Status:      "pending",
		Attempts:    0,
		MaxAttempts: 3,
		ScheduledAt: s.calculateScheduledAt(req.GetDelay()),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.queueRepo.Enqueue(ctx, queuedMsg); err != nil {
		logger.Error("failed to enqueue message", slog.String("error", err.Error()))
		s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "sticker", "enqueue_failed").Inc()
		return nil, fmt.Errorf("failed to enqueue message: %w", err)
	}

	// Update metrics
	duration := time.Since(start)
	s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "sticker", "success").Inc()
	s.metrics.MessageQueueDuration.WithLabelValues(instanceID.String(), "sticker").Observe(duration.Seconds())

	logger.Info("sticker message queued successfully",
		slog.String("zaap_id", zaapID),
		slog.Duration("duration", duration))

	return &SendMessageResult{
		ZaapID:    zaapID,
		MessageID: zaapID,
		ID:        zaapID,
		Status:    "queued",
		QueuedAt:  queuedMsg.CreatedAt,
	}, nil
}

// SendGif sends a GIF message
func (s *Service) SendGif(ctx context.Context, instanceID uuid.UUID, clientToken, instanceToken string, req *SendGifRequest) (*SendMessageResult, error) {
	start := time.Now()
	logger := logging.ContextLogger(ctx, s.log)

	logger = logger.With(
		slog.String("instance_id", instanceID.String()),
		slog.String("operation", "send_gif"),
		slog.String("phone", req.GetPhone()),
	)
	ctx = logging.WithLogger(ctx, logger)

	logger.Info("processing send gif request")

	// Validate request
	if err := s.validator.ValidateSendGifRequest(req); err != nil {
		logger.Warn("validation failed", slog.String("error", err.Error()))
		s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "gif", "validation_failed").Inc()
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify instance exists and tokens match
	inst, err := s.instancesService.GetByID(ctx, instanceID)
	if err != nil {
		logger.Error("failed to get instance", slog.String("error", err.Error()))
		return nil, fmt.Errorf("instance not found: %w", err)
	}

	if !s.tokensMatch(inst, clientToken, instanceToken) {
		logger.Warn("unauthorized access attempt")
		return nil, fmt.Errorf("unauthorized")
	}

	// Check if instance is connected
	if !s.registry.IsConnected(instanceID) {
		logger.Warn("instance not connected")
		return nil, fmt.Errorf("instance not connected")
	}

	// Normalize phone to JID format
	jid, err := s.normalizePhoneToJID(req.GetPhone())
	if err != nil {
		logger.Error("failed to normalize phone", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Generate unique message ID (zaapId)
	zaapID := uuid.New().String()

	// Prepare caption
	var caption *string
	if req.GetCaption() != "" {
		c := req.GetCaption()
		caption = &c
	}

	// Create queue message - GIFs are sent as videos with gifPlayback flag
	args := queue.SendMessageArgs{
		ZaapID:      zaapID,
		InstanceID:  instanceID,
		Phone:       jid.String(),
		MessageType: queue.MessageTypeVideo, // GIFs are handled as videos
		VideoContent: &queue.MediaMessage{
			MediaURL: req.GetGif(),
			Caption:  caption,
		},
		DelayMessage: int64(req.GetDelay()),
		EnqueuedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"is_gif": true,
		},
	}

	// Serialize payload
	payload, err := json.Marshal(args)
	if err != nil {
		logger.Error("failed to serialize message", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Enqueue message
	queuedMsg := &QueuedMessage{
		ID:          uuid.New(),
		InstanceID:  instanceID,
		MessageID:   zaapID,
		Type:        "gif", // Use gif type for tracking
		Phone:       jid.String(),
		Payload:     payload,
		Status:      "pending",
		Attempts:    0,
		MaxAttempts: 3,
		ScheduledAt: s.calculateScheduledAt(req.GetDelay()),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.queueRepo.Enqueue(ctx, queuedMsg); err != nil {
		logger.Error("failed to enqueue message", slog.String("error", err.Error()))
		s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "gif", "enqueue_failed").Inc()
		return nil, fmt.Errorf("failed to enqueue message: %w", err)
	}

	// Update metrics
	duration := time.Since(start)
	s.metrics.MessageQueueEnqueued.WithLabelValues(instanceID.String(), "gif", "success").Inc()
	s.metrics.MessageQueueDuration.WithLabelValues(instanceID.String(), "gif").Observe(duration.Seconds())

	logger.Info("gif message queued successfully",
		slog.String("zaap_id", zaapID),
		slog.Duration("duration", duration))

	return &SendMessageResult{
		ZaapID:    zaapID,
		MessageID: zaapID,
		ID:        zaapID,
		Status:    "queued",
		QueuedAt:  queuedMsg.CreatedAt,
	}, nil
}

// Helper methods

// tokensMatch checks if the provided tokens match the instance tokens
func (s *Service) tokensMatch(inst *instances.Instance, clientToken, instanceToken string) bool {
	return inst.ClientToken == clientToken && inst.InstanceToken == instanceToken
}

// normalizePhoneToJID converts a phone number to WhatsApp JID format
func (s *Service) normalizePhoneToJID(phone string) (types.JID, error) {
	// Normalize phone number
	normalized := NormalizePhone(phone)

	// Remove + prefix for JID creation
	if len(normalized) > 0 && normalized[0] == '+' {
		normalized = normalized[1:]
	}

	// Create JID - assume user JID by default (@s.whatsapp.net)
	// Group JIDs should already include @g.us in the phone parameter
	if len(normalized) == 0 {
		return types.JID{}, fmt.Errorf("invalid phone number")
	}

	// Check if it's a group JID
	if len(phone) > 4 && phone[len(phone)-4:] == "@g.us" {
		return types.ParseJID(phone)
	}

	// Check if it's a newsletter JID
	if len(phone) > 11 && phone[len(phone)-11:] == "@newsletter" {
		return types.ParseJID(phone)
	}

	// Check if it's a broadcast JID
	if len(phone) > 10 && phone[len(phone)-10:] == "@broadcast" {
		return types.ParseJID(phone)
	}

	// Default to user JID
	jid := types.JID{
		User:   normalized,
		Server: types.DefaultUserServer,
	}

	return jid, nil
}

// calculateScheduledAt calculates when a message should be processed
func (s *Service) calculateScheduledAt(delayMS int) time.Time {
	if delayMS <= 0 {
		return time.Now()
	}

	delay := time.Duration(delayMS) * time.Millisecond
	return time.Now().Add(delay)
}
