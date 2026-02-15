package media

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/events/encoding"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	natsclient "go.mau.fi/whatsmeow/api/internal/nats"
	"go.mau.fi/whatsmeow/api/internal/observability"
	whatsmeowevents "go.mau.fi/whatsmeow/types/events"
)

// ClientProvider provides WhatsApp clients for media downloading.
type ClientProvider interface {
	GetClient(instanceID uuid.UUID) (*whatsmeow.Client, bool)
}

// NATSMediaWorker processes media tasks from a NATS consumer.
// It reuses the existing MediaProcessor (download + upload pipeline).
type NATSMediaWorker struct {
	instanceID     uuid.UUID
	natsClient     *natsclient.Client
	clientProvider ClientProvider
	processor      *MediaProcessor
	publisher      *NATSMediaPublisher
	cfg            *config.Config
	mediaCfg       NATSMediaConfig
	metrics        *observability.Metrics
	natsMetrics    *natsclient.NATSMetrics
	log            *slog.Logger

	// Consumer management
	consumer jetstream.Consumer
	consCtx  jetstream.ConsumeContext
	cancel   context.CancelFunc
}

// NATSMediaWorkerConfig holds dependencies for creating a NATSMediaWorker.
type NATSMediaWorkerConfig struct {
	InstanceID     uuid.UUID
	NATSClient     *natsclient.Client
	ClientProvider ClientProvider
	Processor      *MediaProcessor
	Publisher      *NATSMediaPublisher
	Config         *config.Config
	MediaConfig    NATSMediaConfig
	Metrics        *observability.Metrics
	NATSMetrics    *natsclient.NATSMetrics
	Logger         *slog.Logger
}

// NewNATSMediaWorker creates a new NATS-based media worker.
func NewNATSMediaWorker(cfg NATSMediaWorkerConfig) *NATSMediaWorker {
	return &NATSMediaWorker{
		instanceID:     cfg.InstanceID,
		natsClient:     cfg.NATSClient,
		clientProvider: cfg.ClientProvider,
		processor:      cfg.Processor,
		publisher:      cfg.Publisher,
		cfg:            cfg.Config,
		mediaCfg:       cfg.MediaConfig,
		metrics:        cfg.Metrics,
		natsMetrics:    cfg.NATSMetrics,
		log:            cfg.Logger.With(slog.String("instance_id", cfg.InstanceID.String()), slog.String("component", "nats_media_worker")),
	}
}

// Start creates the consumer and begins processing media tasks.
func (w *NATSMediaWorker) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	consumerCfg := natsclient.MediaConsumerConfig(w.instanceID.String())
	consumer, err := w.natsClient.EnsureConsumer(ctx, "MEDIA_PROCESSING", consumerCfg)
	if err != nil {
		cancel()
		return fmt.Errorf("ensure media consumer for %s: %w", w.instanceID, err)
	}
	w.consumer = consumer

	consCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		w.handleMessage(ctx, msg)
	})
	if err != nil {
		cancel()
		return fmt.Errorf("start media consume for %s: %w", w.instanceID, err)
	}
	w.consCtx = consCtx

	w.log.Info("NATS media worker started",
		slog.String("consumer", consumerCfg.Durable),
		slog.String("filter", consumerCfg.FilterSubject))

	return nil
}

// Stop gracefully stops the worker.
func (w *NATSMediaWorker) Stop(ctx context.Context) error {
	if w.consCtx != nil {
		w.consCtx.Stop()
	}
	if w.cancel != nil {
		w.cancel()
	}
	w.log.Info("NATS media worker stopped")
	return nil
}

// handleMessage processes a single media task from NATS.
func (w *NATSMediaWorker) handleMessage(ctx context.Context, msg jetstream.Msg) {
	start := time.Now()

	var task MediaTask
	if err := json.Unmarshal(msg.Data(), &task); err != nil {
		w.log.Error("failed to unmarshal media task",
			slog.String("error", err.Error()))
		if termErr := msg.Term(); termErr != nil {
			w.log.Error("failed to term malformed media task", slog.String("error", termErr.Error()))
		}
		return
	}

	logFields := []any{
		slog.String("event_id", task.EventID.String()),
		slog.String("media_type", task.MediaType),
		slog.String("media_key", task.MediaKey),
	}

	// DEBUG: Log full MediaTask details
	w.log.Info("media task received - debug",
		slog.String("event_id", task.EventID.String()),
		slog.String("instance_id", task.InstanceID.String()),
		slog.String("media_type", task.MediaType),
		slog.String("media_key", task.MediaKey),
		slog.String("direct_path", task.DirectPath),
		slog.String("mime_type", task.MimeType),
		slog.Int64("file_length", task.FileLength),
		slog.Int("payload_len", len(task.Payload)),
		slog.Time("published_at", task.PublishedAt),
	)

	// Get WhatsApp client for downloading
	client, ok := w.clientProvider.GetClient(w.instanceID)
	if !ok || client == nil {
		w.log.Warn("whatsapp client not available for media download, NAK with delay", logFields...)
		if err := msg.NakWithDelay(30 * time.Second); err != nil {
			w.log.Error("failed to nak media task", slog.String("error", err.Error()))
		}
		return
	}

	if !client.IsConnected() {
		w.log.Warn("whatsapp client not connected, NAK with delay", logFields...)
		if err := msg.NakWithDelay(30 * time.Second); err != nil {
			w.log.Error("failed to nak media task", slog.String("error", err.Error()))
		}
		return
	}

	// Decode the internal event to reconstruct the proto message
	internalEvent, err := encoding.DecodeInternalEvent(task.Payload)
	if err != nil {
		w.log.Error("failed to decode internal event for media",
			append(logFields, slog.String("error", err.Error()))...)
		if termErr := msg.Term(); termErr != nil {
			w.log.Error("failed to term decode-failed media task", slog.String("error", termErr.Error()))
		}
		return
	}

	// Reconstruct the proto message from internal event
	protoMsg, err := w.reconstructProtoMessage(internalEvent, task.MediaType)
	if err != nil {
		w.log.Error("failed to reconstruct proto message",
			append(logFields, slog.String("error", err.Error()))...)
		if termErr := msg.Term(); termErr != nil {
			w.log.Error("failed to term reconstruct-failed media task", slog.String("error", termErr.Error()))
		}
		return
	}

	// DEBUG: Log reconstructed proto message fields
	if downloadable, ok := protoMsg.(whatsmeow.DownloadableMessage); ok {
		var protoFileLength int64 = -1
		if sized, ok2 := protoMsg.(interface{ GetFileLength() uint64 }); ok2 {
			protoFileLength = int64(sized.GetFileLength())
		}
		var protoURL string
		if urlable, ok2 := protoMsg.(interface{ GetURL() string }); ok2 {
			protoURL = urlable.GetURL()
		}
		w.log.Info("reconstructed proto message - debug",
			slog.String("event_id", task.EventID.String()),
			slog.String("proto_type", fmt.Sprintf("%T", protoMsg)),
			slog.String("direct_path", downloadable.GetDirectPath()),
			slog.Int("media_key_len", len(downloadable.GetMediaKey())),
			slog.Int("file_sha256_len", len(downloadable.GetFileSHA256())),
			slog.Int("file_enc_sha256_len", len(downloadable.GetFileEncSHA256())),
			slog.Int64("proto_file_length", protoFileLength),
			slog.Bool("has_url", protoURL != ""),
			slog.String("url_preview", truncateForLog(protoURL, 80)),
		)
	} else {
		w.log.Warn("proto message is not DownloadableMessage",
			slog.String("event_id", task.EventID.String()),
			slog.String("proto_type", fmt.Sprintf("%T", protoMsg)),
		)
	}

	// Process the media using existing processor
	result, err := w.processor.ProcessWithRetry(ctx, client, task.InstanceID, task.EventID, protoMsg, task.MediaKey)
	duration := time.Since(start)

	if err != nil {
		w.log.Error("media processing failed",
			append(logFields,
				slog.String("error", err.Error()),
				slog.Duration("duration", duration))...)

		if w.metrics != nil {
			w.metrics.MediaFailures.WithLabelValues(w.instanceID.String(), task.MediaType, "nats_process").Inc()
		}

		// Check delivery count for DLQ
		meta, metaErr := msg.Metadata()
		if metaErr == nil && int(meta.NumDelivered) >= w.mediaCfg.MaxAttempts {
			// Publish failure result
			if w.publisher != nil {
				if pubErr := w.publisher.PublishResult(ctx, MediaResult{
					InstanceID:  task.InstanceID,
					EventID:     task.EventID,
					Success:     false,
					Error:       err.Error(),
					ProcessedAt: time.Now(),
				}); pubErr != nil {
					w.log.Error("failed to publish media failure result",
						slog.String("event_id", task.EventID.String()),
						slog.String("error", pubErr.Error()))
				}
			}
			if termErr := msg.Term(); termErr != nil {
				w.log.Error("failed to term media task after max attempts", slog.String("error", termErr.Error()))
			}
			return
		}

		// NAK for retry
		if nakErr := msg.Nak(); nakErr != nil {
			w.log.Error("failed to nak media task", slog.String("error", nakErr.Error()))
		}
		return
	}

	// Success - publish result and acknowledge
	if w.publisher != nil {
		if pubErr := w.publisher.PublishResult(ctx, MediaResult{
			InstanceID:  task.InstanceID,
			EventID:     task.EventID,
			Success:     true,
			MediaURL:    result.S3URL,
			S3Key:       result.S3Key,
			ContentType: result.ContentType,
			FileSize:    result.FileSize,
			ProcessedAt: time.Now(),
		}); pubErr != nil {
			w.log.Error("failed to publish media success result",
				slog.String("event_id", task.EventID.String()),
				slog.String("error", pubErr.Error()))
		}
	}

	if err := msg.Ack(); err != nil {
		w.log.Error("failed to ack media task", slog.String("error", err.Error()))
		return
	}

	if w.metrics != nil {
		w.metrics.MediaDownloadsTotal.WithLabelValues(w.instanceID.String(), task.MediaType, "success").Inc()
		w.metrics.MediaDownloadDuration.WithLabelValues(w.instanceID.String(), task.MediaType).Observe(duration.Seconds())
	}

	w.log.Info("media processed successfully",
		append(logFields,
			slog.String("s3_key", result.S3Key),
			slog.Int64("file_size", result.FileSize),
			slog.Duration("duration", duration))...)
}

// reconstructProtoMessage extracts the media proto message from a decoded InternalEvent.
func (w *NATSMediaWorker) reconstructProtoMessage(event *types.InternalEvent, mediaType string) (proto.Message, error) {
	msgEvent, ok := event.RawPayload.(*whatsmeowevents.Message)
	if !ok {
		return nil, fmt.Errorf("event raw payload is not a message: %T", event.RawPayload)
	}

	if msgEvent.Message == nil {
		return nil, fmt.Errorf("message event missing proto payload")
	}

	switch mediaType {
	case "image":
		if img := msgEvent.Message.GetImageMessage(); img != nil {
			return img, nil
		}
	case "video":
		if video := msgEvent.Message.GetVideoMessage(); video != nil {
			return video, nil
		}
	case "audio", "voice":
		if audio := msgEvent.Message.GetAudioMessage(); audio != nil {
			return audio, nil
		}
	case "document":
		if doc := msgEvent.Message.GetDocumentMessage(); doc != nil {
			return doc, nil
		}
	case "sticker":
		if sticker := msgEvent.Message.GetStickerMessage(); sticker != nil {
			return sticker, nil
		}
	}

	return nil, fmt.Errorf("unsupported media type: %s", mediaType)
}

// truncateForLog truncates a string for safe logging.
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
