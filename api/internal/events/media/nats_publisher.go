package media

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	natsclient "go.mau.fi/whatsmeow/api/internal/nats"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// NATSMediaPublisher publishes media tasks and results to NATS streams.
type NATSMediaPublisher struct {
	client  *natsclient.Client
	log     *slog.Logger
	metrics *observability.Metrics
}

// NewNATSMediaPublisher creates a new media publisher.
func NewNATSMediaPublisher(client *natsclient.Client, log *slog.Logger, metrics *observability.Metrics) *NATSMediaPublisher {
	return &NATSMediaPublisher{
		client:  client,
		log:     log.With(slog.String("component", "nats_media_publisher")),
		metrics: metrics,
	}
}

// PublishTask publishes a media task to media.tasks.{instance_id}.
func (p *NATSMediaPublisher) PublishTask(ctx context.Context, task MediaTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal media task: %w", err)
	}

	subject := fmt.Sprintf("media.tasks.%s", task.InstanceID.String())
	msg := &natsgo.Msg{
		Subject: subject,
		Data:    data,
		Header:  natsgo.Header{},
	}

	msg.Header.Set(NATSHeaderInstanceID, task.InstanceID.String())
	msg.Header.Set(NATSHeaderEventID, task.EventID.String())
	msg.Header.Set(NATSHeaderMediaType, task.MediaType)
	msg.Header.Set(NATSHeaderMediaKey, task.MediaKey)

	msgID := fmt.Sprintf("media-%s-%s", task.EventID.String(), task.MediaKey)
	_, err = p.client.PublishMsg(ctx, msg, jetstream.WithMsgID(msgID))
	if err != nil {
		return fmt.Errorf("nats publish media task %s: %w", task.EventID, err)
	}

	p.log.Debug("media task published to NATS",
		slog.String("instance_id", task.InstanceID.String()),
		slog.String("event_id", task.EventID.String()),
		slog.String("media_type", task.MediaType),
		slog.String("subject", subject))

	return nil
}

// PublishResult publishes a media processing result to media.done.{instance_id}.{event_id}.
func (p *NATSMediaPublisher) PublishResult(ctx context.Context, result MediaResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal media result: %w", err)
	}

	subject := fmt.Sprintf("media.done.%s.%s", result.InstanceID.String(), result.EventID.String())
	msg := &natsgo.Msg{
		Subject: subject,
		Data:    data,
		Header:  natsgo.Header{},
	}

	msg.Header.Set(NATSHeaderInstanceID, result.InstanceID.String())
	msg.Header.Set(NATSHeaderEventID, result.EventID.String())

	status := "ok"
	if !result.Success {
		status = "fail"
	}
	msgID := fmt.Sprintf("media-done-%s-%s", result.EventID.String(), status)
	_, err = p.client.PublishMsg(ctx, msg, jetstream.WithMsgID(msgID))
	if err != nil {
		return fmt.Errorf("nats publish media result %s: %w", result.EventID, err)
	}

	p.log.Debug("media result published to NATS",
		slog.String("instance_id", result.InstanceID.String()),
		slog.String("event_id", result.EventID.String()),
		slog.Bool("success", result.Success),
		slog.String("subject", subject))

	return nil
}
