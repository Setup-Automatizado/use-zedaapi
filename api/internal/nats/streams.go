package nats

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

// Stream configuration constants.
const (
	SubjectMessagesAll = "messages.>"
	SubjectEventsAll   = "events.>"
	SubjectMediaTasks  = "media.tasks.>"
	SubjectMediaDone   = "media.done.>"
	SubjectDLQAll      = "dlq.>"
)

// MessageQueueStreamConfig returns the JetStream config for the MESSAGE_QUEUE stream.
func MessageQueueStreamConfig(name string) jetstream.StreamConfig {
	return jetstream.StreamConfig{
		Name:              name,
		Subjects:          []string{SubjectMessagesAll},
		Retention:         jetstream.WorkQueuePolicy,
		MaxAge:            72 * time.Hour,
		MaxBytes:          10 * 1024 * 1024 * 1024, // 10GB
		Storage:           jetstream.FileStorage,
		Discard:           jetstream.DiscardOld,
		Duplicates:        2 * time.Minute,
		MaxMsgSize:        8 * 1024 * 1024, // 8MB
		NoAck:             false,
		MaxMsgsPerSubject: -1,
	}
}

// WhatsAppEventsStreamConfig returns the JetStream config for the WHATSAPP_EVENTS stream.
func WhatsAppEventsStreamConfig(name string) jetstream.StreamConfig {
	return jetstream.StreamConfig{
		Name:              name,
		Subjects:          []string{SubjectEventsAll},
		Retention:         jetstream.LimitsPolicy,
		MaxAge:            168 * time.Hour,         // 7 days
		MaxBytes:          50 * 1024 * 1024 * 1024, // 50GB
		Storage:           jetstream.FileStorage,
		Discard:           jetstream.DiscardOld,
		Duplicates:        1 * time.Hour,
		MaxMsgSize:        2 * 1024 * 1024, // 2MB
		NoAck:             false,
		MaxMsgsPerSubject: -1,
	}
}

// MediaProcessingStreamConfig returns the JetStream config for the MEDIA_PROCESSING stream.
func MediaProcessingStreamConfig(name string) jetstream.StreamConfig {
	return jetstream.StreamConfig{
		Name:              name,
		Subjects:          []string{SubjectMediaTasks, SubjectMediaDone},
		Retention:         jetstream.LimitsPolicy,
		MaxAge:            168 * time.Hour,        // 7 days
		MaxBytes:          5 * 1024 * 1024 * 1024, // 5GB
		Storage:           jetstream.FileStorage,
		Discard:           jetstream.DiscardOld,
		Duplicates:        2 * time.Minute,
		MaxMsgSize:        1024 * 1024, // 1MB
		NoAck:             false,
		MaxMsgsPerSubject: -1,
	}
}

// DLQStreamConfig returns the JetStream config for the DLQ stream.
func DLQStreamConfig(name string) jetstream.StreamConfig {
	return jetstream.StreamConfig{
		Name:              name,
		Subjects:          []string{SubjectDLQAll},
		Retention:         jetstream.LimitsPolicy,
		MaxAge:            720 * time.Hour,        // 30 days
		MaxBytes:          5 * 1024 * 1024 * 1024, // 5GB
		Storage:           jetstream.FileStorage,
		Discard:           jetstream.DiscardOld,
		Duplicates:        2 * time.Minute,
		MaxMsgSize:        2 * 1024 * 1024, // 2MB
		NoAck:             false,
		MaxMsgsPerSubject: -1,
	}
}

// MessageConsumerConfig returns a consumer config for per-instance message processing.
// MaxAckPending=1 guarantees FIFO ordering per instance.
func MessageConsumerConfig(instanceID string) jetstream.ConsumerConfig {
	return jetstream.ConsumerConfig{
		Durable:       fmt.Sprintf("msg-%s", instanceID),
		FilterSubject: fmt.Sprintf("messages.%s", instanceID),
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
		MaxDeliver:    5,
		MaxAckPending: 1, // CRITICAL: FIFO guarantee
		BackOff:       []time.Duration{1 * time.Second, 5 * time.Second, 30 * time.Second, 2 * time.Minute, 5 * time.Minute},
		DeliverPolicy: jetstream.DeliverAllPolicy,
	}
}

// EventConsumerConfig returns a consumer config for per-instance event dispatch.
// BackOff replaces AckWait per delivery attempt: BackOff[0] is the ack timeout for
// the 1st delivery, BackOff[1] for the 2nd, etc. Values must be > 0 to give the
// worker enough time to process (HTTP transport + transform). Setting BackOff[0]=0
// causes immediate redelivery because the worker can never ACK within 0 seconds.
func EventConsumerConfig(instanceID string) jetstream.ConsumerConfig {
	return jetstream.ConsumerConfig{
		Durable:       fmt.Sprintf("evt-%s", instanceID),
		FilterSubject: fmt.Sprintf("events.%s.>", instanceID),
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       60 * time.Second,
		MaxDeliver:    6,
		MaxAckPending: 10,
		BackOff:       []time.Duration{30 * time.Second, 30 * time.Second, 1 * time.Minute, 2 * time.Minute, 5 * time.Minute, 15 * time.Minute},
		DeliverPolicy: jetstream.DeliverAllPolicy,
	}
}

// MediaConsumerConfig returns a consumer config for per-instance media processing.
func MediaConsumerConfig(instanceID string) jetstream.ConsumerConfig {
	return jetstream.ConsumerConfig{
		Durable:       fmt.Sprintf("media-%s", instanceID),
		FilterSubject: fmt.Sprintf("media.tasks.%s", instanceID),
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       120 * time.Second,
		MaxDeliver:    5,
		MaxAckPending: 3,
		BackOff:       []time.Duration{5 * time.Second, 30 * time.Second, 2 * time.Minute, 10 * time.Minute, 30 * time.Minute},
		DeliverPolicy: jetstream.DeliverAllPolicy,
	}
}

// MediaCompletionConsumerConfig returns a consumer config for media.done events.
func MediaCompletionConsumerConfig() jetstream.ConsumerConfig {
	return jetstream.ConsumerConfig{
		Durable:       "media-completion",
		FilterSubject: "media.done.>",
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
		MaxDeliver:    3,
		MaxAckPending: 10,
		DeliverPolicy: jetstream.DeliverAllPolicy,
	}
}

// EnsureAllStreams creates or updates all required JetStream streams.
func EnsureAllStreams(ctx context.Context, js jetstream.JetStream, cfg Config, log *slog.Logger) error {
	streams := []jetstream.StreamConfig{
		MessageQueueStreamConfig(cfg.StreamMessageQueue),
		WhatsAppEventsStreamConfig(cfg.StreamWhatsAppEvents),
		MediaProcessingStreamConfig(cfg.StreamMediaProcessing),
		DLQStreamConfig(cfg.StreamDLQ),
	}

	for _, streamCfg := range streams {
		stream, err := js.CreateOrUpdateStream(ctx, streamCfg)
		if err != nil {
			return fmt.Errorf("ensure stream %s: %w", streamCfg.Name, err)
		}
		info, err := stream.Info(ctx)
		if err != nil {
			log.Warn("failed to get stream info after create",
				slog.String("stream", streamCfg.Name),
				slog.String("error", err.Error()))
			continue
		}
		log.Info("stream ensured",
			slog.String("stream", streamCfg.Name),
			slog.Uint64("messages", info.State.Msgs),
			slog.Uint64("bytes", info.State.Bytes),
		)
	}

	return nil
}
