package capture

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/events/transform"
	whatsmeowtransform "go.mau.fi/whatsmeow/api/internal/events/transform/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

type EventHandler struct {
	log         *slog.Logger
	metrics     *observability.Metrics
	router      *EventRouter
	instanceID  uuid.UUID
	transformer transform.SourceTransformer
	debugRaw    bool
	dumpDir     string

	mu      sync.RWMutex
	stopped bool
	stopCh  chan struct{}
}

func NewEventHandler(
	ctx context.Context,
	instanceID uuid.UUID,
	router *EventRouter,
	metrics *observability.Metrics,
	debugRaw bool,
	dumpDir string,
) *EventHandler {
	log := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "event_handler"),
		slog.String("instance_id", instanceID.String()),
	)

	sourceTransformer := whatsmeowtransform.NewTransformer(instanceID, debugRaw)

	return &EventHandler{
		log:         log,
		metrics:     metrics,
		router:      router,
		instanceID:  instanceID,
		transformer: sourceTransformer,
		debugRaw:    debugRaw,
		dumpDir:     dumpDir,
		stopCh:      make(chan struct{}),
	}
}

func (h *EventHandler) HandleEvent(ctx context.Context, rawEvent interface{}) error {
	h.mu.RLock()
	if h.stopped {
		h.mu.RUnlock()
		return fmt.Errorf("event handler stopped")
	}
	h.mu.RUnlock()

	h.logRawEvent(ctx, rawEvent, "incoming")
	if h.debugRaw {
		h.dumpRawEvent(rawEvent, "raw_event_captured")
	}

	internalEvent, err := h.transformer.Transform(ctx, rawEvent)
	if errors.Is(err, transform.ErrUnsupportedEvent) {
		h.log.DebugContext(ctx, "unsupported event type skipped",
			slog.String("type", fmt.Sprintf("%T", rawEvent)))
		h.dumpRawEvent(rawEvent, "unsupported_transform")
		return nil
	}
	if err != nil {
		h.metrics.EventsCaptured.WithLabelValues(
			h.instanceID.String(),
			"unsupported",
			string(types.SourceLibWhatsmeow),
		).Inc()
		h.dumpRawEvent(rawEvent, "transform_error")
		return fmt.Errorf("transform event: %w", err)
	}

	h.metrics.EventsCaptured.WithLabelValues(
		h.instanceID.String(),
		internalEvent.EventType,
		string(internalEvent.SourceLib),
	).Inc()

	h.log.DebugContext(ctx, "event captured",
		slog.String("event_id", internalEvent.EventID.String()),
		slog.String("event_type", internalEvent.EventType),
		slog.Bool("has_media", internalEvent.HasMedia),
	)

	h.logInternalEvent(ctx, internalEvent)
	if h.debugRaw {
		h.dumpInternalEvent(internalEvent, "internal_event_captured")
	}

	return h.router.RouteEvent(ctx, internalEvent)
}

func (h *EventHandler) logRawEvent(ctx context.Context, rawEvent interface{}, reason string) {
	if !h.debugRaw {
		return
	}

	payload := fmt.Sprintf("%+v", rawEvent)
	if len(payload) > 4096 {
		payload = payload[:4096] + "...<truncated>"
	}

	h.log.DebugContext(ctx, "raw whatsmeow event",
		slog.String("reason", reason),
		slog.String("type", fmt.Sprintf("%T", rawEvent)),
		slog.String("payload", payload))
}

func (h *EventHandler) logInternalEvent(ctx context.Context, event *types.InternalEvent) {
	if !h.debugRaw || event == nil {
		return
	}

	metadata := map[string]interface{}{
		"event_id":    event.EventID.String(),
		"event_type":  event.EventType,
		"instance_id": event.InstanceID.String(),
		"has_media":   event.HasMedia,
		"metadata":    event.Metadata,
	}

	raw, err := json.Marshal(metadata)
	if err != nil {
		h.log.DebugContext(ctx, "internal event metadata",
			slog.String("event_id", event.EventID.String()),
			slog.String("event_type", event.EventType))
		return
	}

	h.log.DebugContext(ctx, "internal event metadata", slog.String("details", string(raw)))
}

func (h *EventHandler) dumpRawEvent(rawEvent interface{}, reason string) {
	if h.dumpDir == "" {
		return
	}

	if err := os.MkdirAll(h.dumpDir, 0o755); err != nil {
		h.log.Warn("failed to create debug dump directory",
			slog.String("dir", h.dumpDir),
			slog.String("error", err.Error()))
		return
	}

	dump := fmt.Sprintf("timestamp=%s\ninstance_id=%s\nreason=%s\ntype=%T\npayload=%+v\n",
		time.Now().Format(time.RFC3339Nano),
		h.instanceID.String(),
		reason,
		rawEvent,
		rawEvent,
	)

	fileName := fmt.Sprintf("%s_%s_%d.log", reason, h.instanceID.String(), time.Now().UnixNano())
	filePath := filepath.Join(h.dumpDir, fileName)

	if err := os.WriteFile(filePath, []byte(dump), 0o644); err != nil {
		h.log.Warn("failed to write debug dump",
			slog.String("path", filePath),
			slog.String("error", err.Error()))
	}
}

func (h *EventHandler) dumpInternalEvent(event *types.InternalEvent, reason string) {
	if event == nil || h.dumpDir == "" {
		return
	}

	if err := os.MkdirAll(h.dumpDir, 0o755); err != nil {
		h.log.Warn("failed to create debug dump directory",
			slog.String("dir", h.dumpDir),
			slog.String("error", err.Error()))
		return
	}

	metadataCopy := make(map[string]string, len(event.Metadata))
	for k, v := range event.Metadata {
		metadataCopy[k] = v
	}

	dump := map[string]interface{}{
		"timestamp":         time.Now().Format(time.RFC3339Nano),
		"reason":            reason,
		"event_id":          event.EventID.String(),
		"event_type":        event.EventType,
		"instance_id":       event.InstanceID.String(),
		"source_lib":        string(event.SourceLib),
		"captured_at":       event.CapturedAt.Format(time.RFC3339Nano),
		"has_media":         event.HasMedia,
		"metadata":          metadataCopy,
		"media_key":         event.MediaKey,
		"direct_path":       event.DirectPath,
		"quoted_message_id": event.QuotedMessageID,
		"quoted_sender":     event.QuotedSender,
		"quoted_remote_jid": event.QuotedRemoteJID,
		"mentioned_jids":    event.MentionedJIDs,
		"is_forwarded":      event.IsForwarded,
		"ephemeral_expiry":  event.EphemeralExpiry,
		"transport_type":    event.TransportType,
	}

	if event.FileSHA256 != nil {
		dump["file_sha256"] = *event.FileSHA256
	}
	if event.FileEncSHA256 != nil {
		dump["file_enc_sha256"] = *event.FileEncSHA256
	}
	if event.MimeType != nil {
		dump["mime_type"] = *event.MimeType
	}
	if event.FileLength != nil {
		dump["file_length"] = *event.FileLength
	}
	if len(event.TransportConfig) > 0 {
		dump["transport_config"] = json.RawMessage(append([]byte(nil), event.TransportConfig...))
	}
	if len(event.MediaWaveform) > 0 {
		dump["media_waveform"] = append([]byte(nil), event.MediaWaveform...)
	}
	if event.RawPayload != nil {
		dump["raw_payload_type"] = fmt.Sprintf("%T", event.RawPayload)
		dump["raw_payload_repr"] = fmt.Sprintf("%+v", event.RawPayload)
	}

	data, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		data = []byte(fmt.Sprintf("%+v", dump))
	}

	fileName := fmt.Sprintf("internal_event_%s_%s_%d.json", event.EventType, event.EventID.String(), time.Now().UnixNano())
	filePath := filepath.Join(h.dumpDir, fileName)

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		h.log.Warn("failed to write internal event dump",
			slog.String("path", filePath),
			slog.String("error", err.Error()))
	}
}

func (h *EventHandler) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.stopped {
		h.stopped = true
		close(h.stopCh)
	}
}

func (h *EventHandler) IsStopped() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.stopped
}
