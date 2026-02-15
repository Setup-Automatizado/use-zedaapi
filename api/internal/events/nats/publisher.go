package nats

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
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// EventPublisher publishes events to the NATS WHATSAPP_EVENTS stream.
type EventPublisher struct {
	client  *natsclient.Client
	log     *slog.Logger
	metrics *observability.Metrics
}

// NewEventPublisher creates a new event publisher.
func NewEventPublisher(client *natsclient.Client, log *slog.Logger, metrics *observability.Metrics) *EventPublisher {
	return &EventPublisher{
		client:  client,
		log:     log.With(slog.String("component", "nats_event_publisher")),
		metrics: metrics,
	}
}

// Publish publishes a single event envelope to events.{instance_id}.{event_type}.
func (p *EventPublisher) Publish(ctx context.Context, envelope NATSEventEnvelope) error {
	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal event envelope: %w", err)
	}

	subject := fmt.Sprintf("events.%s.%s", envelope.InstanceID.String(), envelope.EventType)
	msg := &natsgo.Msg{
		Subject: subject,
		Data:    data,
		Header:  natsgo.Header{},
	}

	msg.Header.Set(HeaderInstanceID, envelope.InstanceID.String())
	msg.Header.Set(HeaderEventID, envelope.EventID.String())
	msg.Header.Set(HeaderEventType, envelope.EventType)
	msg.Header.Set(HeaderSourceLib, envelope.SourceLib)
	msg.Header.Set(HeaderCapturedAt, envelope.CapturedAt.Format(time.RFC3339Nano))
	msg.Header.Set(HeaderHasMedia, strconv.FormatBool(envelope.HasMedia))

	// Use EventID as MsgID for deduplication
	msgID := fmt.Sprintf("evt-%s", envelope.EventID.String())
	_, err = p.client.PublishMsg(ctx, msg, jetstream.WithMsgID(msgID))
	if err != nil {
		if p.metrics != nil {
			p.metrics.EventsInserted.WithLabelValues(
				envelope.InstanceID.String(),
				envelope.EventType,
				"nats_publish_failed",
			).Inc()
		}
		return fmt.Errorf("nats publish event %s: %w", envelope.EventID, err)
	}

	p.log.Debug("event published to NATS",
		slog.String("instance_id", envelope.InstanceID.String()),
		slog.String("event_id", envelope.EventID.String()),
		slog.String("event_type", envelope.EventType),
		slog.String("subject", subject),
	)

	return nil
}

// PublishBatch publishes multiple event envelopes.
func (p *EventPublisher) PublishBatch(ctx context.Context, envelopes []NATSEventEnvelope) error {
	for i, env := range envelopes {
		if err := p.Publish(ctx, env); err != nil {
			return fmt.Errorf("publish batch event %d/%d: %w", i+1, len(envelopes), err)
		}
	}
	return nil
}
