package instances

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type WebhookConfig struct {
	InstanceID          uuid.UUID
	DeliveryURL         *string
	ReceivedURL         *string
	ReceivedDeliveryURL *string
	MessageStatusURL    *string
	DisconnectedURL     *string
	ChatPresenceURL     *string
	ConnectedURL        *string
	HistorySyncURL      *string
	NotifySentByMe      bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (r *Repository) UpsertWebhookConfig(ctx context.Context, cfg WebhookConfig) error {
	query := `
        INSERT INTO webhook_configs (
            instance_id,
            delivery_url,
            received_url,
            received_delivery_url,
            message_status_url,
            disconnected_url,
            chat_presence_url,
            connected_url,
            history_sync_url,
            notify_sent_by_me
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        ON CONFLICT (instance_id) DO UPDATE SET
            delivery_url = EXCLUDED.delivery_url,
            received_url = EXCLUDED.received_url,
            received_delivery_url = EXCLUDED.received_delivery_url,
            message_status_url = EXCLUDED.message_status_url,
            disconnected_url = EXCLUDED.disconnected_url,
            chat_presence_url = EXCLUDED.chat_presence_url,
            connected_url = EXCLUDED.connected_url,
            history_sync_url = EXCLUDED.history_sync_url,
            notify_sent_by_me = EXCLUDED.notify_sent_by_me,
            updated_at = NOW()
    `
	_, err := r.pool.Exec(ctx, query,
		cfg.InstanceID,
		cfg.DeliveryURL,
		cfg.ReceivedURL,
		cfg.ReceivedDeliveryURL,
		cfg.MessageStatusURL,
		cfg.DisconnectedURL,
		cfg.ChatPresenceURL,
		cfg.ConnectedURL,
		cfg.HistorySyncURL,
		cfg.NotifySentByMe,
	)
	if err != nil {
		return fmt.Errorf("upsert webhook config: %w", err)
	}
	return nil
}

func (r *Repository) GetWebhookConfig(ctx context.Context, id uuid.UUID) (*WebhookConfig, error) {
	query := `SELECT instance_id, delivery_url, received_url, received_delivery_url,
	          message_status_url, disconnected_url, chat_presence_url, connected_url,
	          history_sync_url, notify_sent_by_me, created_at, updated_at
	          FROM webhook_configs WHERE instance_id=$1`
	row := r.pool.QueryRow(ctx, query, id)
	var cfg WebhookConfig
	if err := row.Scan(
		&cfg.InstanceID,
		&cfg.DeliveryURL,
		&cfg.ReceivedURL,
		&cfg.ReceivedDeliveryURL,
		&cfg.MessageStatusURL,
		&cfg.DisconnectedURL,
		&cfg.ChatPresenceURL,
		&cfg.ConnectedURL,
		&cfg.HistorySyncURL,
		&cfg.NotifySentByMe,
		&cfg.CreatedAt,
		&cfg.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			cfg.InstanceID = id
			return &cfg, nil
		}
		return nil, fmt.Errorf("get webhook config: %w", err)
	}
	return &cfg, nil
}
