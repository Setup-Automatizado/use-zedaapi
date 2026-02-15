package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	natsclient "go.mau.fi/whatsmeow/api/internal/nats"
)

// NATSPublisher publishes messages to the NATS MESSAGE_QUEUE stream.
type NATSPublisher struct {
	client *natsclient.Client
	cfg    NATSConfig
	log    *slog.Logger
}

// NewNATSPublisher creates a new publisher for the message queue stream.
func NewNATSPublisher(client *natsclient.Client, cfg NATSConfig, log *slog.Logger) *NATSPublisher {
	return &NATSPublisher{
		client: client,
		cfg:    cfg,
		log:    log.With(slog.String("component", "nats_msg_publisher")),
	}
}

// Publish publishes a SendMessageArgs to messages.{instance_id} with dedup via MsgID.
func (p *NATSPublisher) Publish(ctx context.Context, args SendMessageArgs) error {
	// Serialize the full args as payload
	payload, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("marshal message args: %w", err)
	}

	// Build envelope
	envelope := NATSMessageEnvelope{
		ZaapID:      args.ZaapID,
		InstanceID:  args.InstanceID,
		ScheduledAt: args.ScheduledFor,
		EnqueuedAt:  args.EnqueuedAt,
		Attempt:     0,
		MaxAttempts: p.cfg.MaxAttempts,
		Payload:     payload,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}

	// Build NATS message with headers for routing and filtering
	subject := fmt.Sprintf("messages.%s", args.InstanceID.String())
	msg := &natsgo.Msg{
		Subject: subject,
		Data:    data,
		Header:  natsgo.Header{},
	}

	msg.Header.Set(HeaderInstanceID, args.InstanceID.String())
	msg.Header.Set(HeaderZaapID, args.ZaapID)
	msg.Header.Set(HeaderMessageType, string(args.MessageType))
	msg.Header.Set(HeaderProcessingKey, args.Phone)
	msg.Header.Set(HeaderEnqueuedAt, args.EnqueuedAt.Format(time.RFC3339Nano))
	msg.Header.Set(HeaderMaxAttempts, strconv.Itoa(p.cfg.MaxAttempts))

	if !args.ScheduledFor.IsZero() {
		msg.Header.Set(HeaderScheduledAt, args.ScheduledFor.Format(time.RFC3339Nano))
	}

	// Publish with MsgID for deduplication
	_, err = p.client.PublishMsg(ctx, msg, jetstream.WithMsgID(args.ZaapID))
	if err != nil {
		return fmt.Errorf("nats publish message %s: %w", args.ZaapID, err)
	}

	p.log.Debug("message published to NATS",
		slog.String("instance_id", args.InstanceID.String()),
		slog.String("zaap_id", args.ZaapID),
		slog.String("subject", subject),
		slog.String("message_type", string(args.MessageType)),
	)

	return nil
}
