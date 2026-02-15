package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	natsclient "go.mau.fi/whatsmeow/api/internal/nats"
)

// NATSDLQEntry represents a message that failed permanently.
type NATSDLQEntry struct {
	ZaapID     string          `json:"zaap_id"`
	InstanceID string          `json:"instance_id"`
	Envelope   json.RawMessage `json:"envelope"`
	Error      string          `json:"error"`
	Attempts   int             `json:"attempts"`
	FailedAt   time.Time       `json:"failed_at"`
}

// NATSDLQHandler publishes failed messages to the DLQ stream.
type NATSDLQHandler struct {
	client *natsclient.Client
	log    *slog.Logger
}

// NewNATSDLQHandler creates a new DLQ handler.
func NewNATSDLQHandler(client *natsclient.Client, log *slog.Logger) *NATSDLQHandler {
	return &NATSDLQHandler{
		client: client,
		log:    log.With(slog.String("component", "nats_msg_dlq")),
	}
}

// SendToDLQ publishes a failed message to dlq.messages.{instance_id}.
func (h *NATSDLQHandler) SendToDLQ(ctx context.Context, instanceID string, envelope json.RawMessage, zaapID string, attempts int, errorMsg string) error {
	entry := NATSDLQEntry{
		ZaapID:     zaapID,
		InstanceID: instanceID,
		Envelope:   envelope,
		Error:      errorMsg,
		Attempts:   attempts,
		FailedAt:   time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal DLQ entry: %w", err)
	}

	subject := fmt.Sprintf("dlq.messages.%s", instanceID)
	msg := &natsgo.Msg{
		Subject: subject,
		Data:    data,
		Header:  natsgo.Header{},
	}
	msg.Header.Set(HeaderInstanceID, instanceID)
	msg.Header.Set(HeaderZaapID, zaapID)

	dlqMsgID := fmt.Sprintf("dlq-msg-%s", zaapID)
	_, err = h.client.PublishMsg(ctx, msg, jetstream.WithMsgID(dlqMsgID))
	if err != nil {
		return fmt.Errorf("publish to DLQ %s: %w", subject, err)
	}

	h.log.Warn("message sent to DLQ",
		slog.String("instance_id", instanceID),
		slog.String("zaap_id", zaapID),
		slog.Int("attempts", attempts),
		slog.String("error", errorMsg),
	)

	return nil
}
