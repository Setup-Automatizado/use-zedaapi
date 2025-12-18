package instances

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInstanceNotFound = errors.New("instance not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type StoreLink struct {
	ID       uuid.UUID
	StoreJID string
}

type ConnectionState struct {
	Connected       bool
	Status          string
	LastConnectedAt *time.Time
	WorkerID        *string
	DesiredWorkerID *string
}

func (r *Repository) Insert(ctx context.Context, inst *Instance) error {
	query := `INSERT INTO instances (
	    id, name, session_name, instance_token, store_jid,
	    is_device, business_device, subscription_active, canceled_at,
	    call_reject_auto, call_reject_message, auto_read_message
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	_, err := r.pool.Exec(ctx, query,
		inst.ID,
		inst.Name,
		inst.SessionName,
		inst.InstanceToken,
		inst.StoreJID,
		inst.IsDevice,
		inst.BusinessDevice,
		inst.SubscriptionActive,
		inst.CanceledAt,
		inst.CallRejectAuto,
		inst.CallRejectMessage,
		inst.AutoReadMessage,
	)
	if err != nil {
		return fmt.Errorf("insert instance: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Instance, error) {
	query := `SELECT id, name, session_name, instance_token, store_jid, is_device, business_device, subscription_active, canceled_at, call_reject_auto, call_reject_message, auto_read_message, created_at, updated_at, connected, connection_status, last_connected_at, worker_id, desired_worker_id
        FROM instances WHERE id=$1`
	row := r.pool.QueryRow(ctx, query, id)
	var inst Instance
	var storeJID *string
	var canceledAt *time.Time
	var callRejectMessage *string
	var connected bool
	var connectionStatus string
	var lastConnectedAt *time.Time
	var workerID *string
	var desiredWorkerID *string
	if err := row.Scan(
		&inst.ID,
		&inst.Name,
		&inst.SessionName,
		&inst.InstanceToken,
		&storeJID,
		&inst.IsDevice,
		&inst.BusinessDevice,
		&inst.SubscriptionActive,
		&canceledAt,
		&inst.CallRejectAuto,
		&callRejectMessage,
		&inst.AutoReadMessage,
		&inst.CreatedAt,
		&inst.UpdatedAt,
		&connected,
		&connectionStatus,
		&lastConnectedAt,
		&workerID,
		&desiredWorkerID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInstanceNotFound
		}
		return nil, fmt.Errorf("query instance: %w", err)
	}
	inst.StoreJID = storeJID
	inst.CanceledAt = canceledAt
	inst.CallRejectMessage = callRejectMessage
	inst.WhatsappConnected = connected
	inst.ConnectionStatus = connectionStatus
	inst.LastConnectedAt = lastConnectedAt
	inst.WorkerID = workerID
	inst.DesiredWorkerID = desiredWorkerID
	if inst.IsDevice {
		inst.Middleware = "mobile"
	} else {
		inst.Middleware = "web"
	}
	return &inst, nil
}

func (r *Repository) UpdateStoreJID(ctx context.Context, id uuid.UUID, jid *string) error {
	query := `UPDATE instances SET store_jid=$2 WHERE id=$1`
	res, err := r.pool.Exec(ctx, query, id, jid)
	if err != nil {
		return fmt.Errorf("update store jid: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrInstanceNotFound
	}
	return nil
}

func (r *Repository) VerifyToken(ctx context.Context, id uuid.UUID, token string) error {
	query := `SELECT 1 FROM instances WHERE id=$1 AND instance_token=$2`
	row := r.pool.QueryRow(ctx, query, id, token)
	var one int
	if err := row.Scan(&one); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInstanceNotFound
		}
		return fmt.Errorf("verify token: %w", err)
	}
	return nil
}

func (r *Repository) UpdateSubscription(ctx context.Context, id uuid.UUID, active bool) error {
	query := `UPDATE instances SET subscription_active=$2, canceled_at=$3, updated_at=NOW() WHERE id=$1`
	var canceledAt interface{}
	if !active {
		now := time.Now().UTC()
		canceledAt = now
	}
	res, err := r.pool.Exec(ctx, query, id, active, canceledAt)
	if err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrInstanceNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	// Cascades automáticos cuidam da remoção de:
	// - webhook_configs (ON DELETE CASCADE)
	// - message_sequences (ON DELETE CASCADE)
	query := `DELETE FROM instances WHERE id=$1`

	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete instance: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrInstanceNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]Instance, int64, error) {
	search := strings.TrimSpace(filter.Query)
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 15
	}
	offset := (filter.Page - 1) * filter.PageSize

	middleware := strings.ToLower(strings.TrimSpace(filter.Middleware))
	rows, err := r.pool.Query(ctx, `
        SELECT
            i.id, i.name, i.session_name, i.instance_token, i.store_jid,
            i.is_device, i.business_device, i.subscription_active, i.canceled_at,
            i.call_reject_auto, i.call_reject_message, i.auto_read_message,
            i.created_at, i.updated_at, i.connected, i.connection_status, i.last_connected_at, i.worker_id, i.desired_worker_id,
            wc.delivery_url, wc.received_url, wc.received_delivery_url, wc.message_status_url,
            wc.disconnected_url, wc.chat_presence_url, wc.connected_url, COALESCE(wc.notify_sent_by_me, FALSE)
        FROM instances i
        LEFT JOIN webhook_configs wc ON wc.instance_id = i.id
        WHERE ($1 = '' OR i.name ILIKE '%' || $1 || '%' OR i.session_name ILIKE '%' || $1 || '%')
          AND ($2 = '' OR ($2 = 'web' AND i.is_device = FALSE) OR ($2 = 'mobile' AND i.is_device = TRUE))
        ORDER BY i.created_at DESC
        LIMIT $3 OFFSET $4
    `, search, middleware, filter.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list instances: %w", err)
	}
	defer rows.Close()

	instances := make([]Instance, 0)
	for rows.Next() {
		var inst Instance
		var storeJID *string
		var canceledAt *time.Time
		var callRejectMessage *string
		var connected bool
		var connectionStatus string
		var lastConnectedAt *time.Time
		var workerID *string
		var desiredWorkerID *string
		var delivery, received, receivedDelivery, statusURL, disconnected, chatPresence, connectedURL *string
		var notify bool
		if err := rows.Scan(
			&inst.ID,
			&inst.Name,
			&inst.SessionName,
			&inst.InstanceToken,
			&storeJID,
			&inst.IsDevice,
			&inst.BusinessDevice,
			&inst.SubscriptionActive,
			&canceledAt,
			&inst.CallRejectAuto,
			&callRejectMessage,
			&inst.AutoReadMessage,
			&inst.CreatedAt,
			&inst.UpdatedAt,
			&connected,
			&connectionStatus,
			&lastConnectedAt,
			&workerID,
			&desiredWorkerID,
			&delivery,
			&received,
			&receivedDelivery,
			&statusURL,
			&disconnected,
			&chatPresence,
			&connectedURL,
			&notify,
		); err != nil {
			return nil, 0, fmt.Errorf("scan instance: %w", err)
		}
		inst.StoreJID = storeJID
		inst.CanceledAt = canceledAt
		inst.CallRejectMessage = callRejectMessage
		inst.WhatsappConnected = connected
		inst.ConnectionStatus = connectionStatus
		inst.LastConnectedAt = lastConnectedAt
		inst.WorkerID = workerID
		inst.DesiredWorkerID = desiredWorkerID
		inst.Webhooks = &WebhookSettings{
			DeliveryURL:         delivery,
			ReceivedURL:         received,
			ReceivedDeliveryURL: receivedDelivery,
			MessageStatusURL:    statusURL,
			DisconnectedURL:     disconnected,
			ChatPresenceURL:     chatPresence,
			ConnectedURL:        connectedURL,
			NotifySentByMe:      notify,
		}
		if inst.IsDevice {
			inst.Middleware = "mobile"
		} else {
			inst.Middleware = "web"
		}
		instances = append(instances, inst)
	}

	var total int64
	if err := r.pool.QueryRow(ctx, `
        SELECT COUNT(*) FROM instances
        WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR session_name ILIKE '%' || $1 || '%')
          AND ($2 = '' OR ($2 = 'web' AND is_device = FALSE) OR ($2 = 'mobile' AND is_device = TRUE))
    `, search, middleware).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count instances: %w", err)
	}

	return instances, total, nil
}

func (r *Repository) ListInstancesWithStoreJID(ctx context.Context) ([]StoreLink, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT id, store_jid
        FROM instances
        WHERE store_jid IS NOT NULL AND store_jid <> ''
    `)
	if err != nil {
		return nil, fmt.Errorf("list instances with store jid: %w", err)
	}
	defer rows.Close()

	var links []StoreLink
	for rows.Next() {
		var link StoreLink
		if err := rows.Scan(&link.ID, &link.StoreJID); err != nil {
			return nil, fmt.Errorf("scan store link: %w", err)
		}
		links = append(links, link)
	}

	return links, nil
}

func (r *Repository) UpdateConnectionStatus(ctx context.Context, id uuid.UUID, connected bool, status string, workerID *string, desiredWorkerID *string) error {
	var worker interface{}
	if workerID != nil && *workerID != "" {
		worker = *workerID
	}
	var desired interface{}
	if desiredWorkerID != nil && *desiredWorkerID != "" {
		desired = *desiredWorkerID
	}

	query := `
		UPDATE instances
		SET
			connected = $2,
			connection_status = $3,
			worker_id = $4,
			desired_worker_id = COALESCE($5, desired_worker_id),
			last_connected_at = CASE WHEN $2 = TRUE THEN NOW() ELSE last_connected_at END,
			updated_at = NOW()
		WHERE id = $1
	`

	res, err := r.pool.Exec(ctx, query, id, connected, status, worker, desired)
	if err != nil {
		return fmt.Errorf("update connection status: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrInstanceNotFound
	}
	return nil
}

func (r *Repository) GetConnectionState(ctx context.Context, id uuid.UUID) (*ConnectionState, error) {
	query := `
		SELECT connected, connection_status, last_connected_at, worker_id, desired_worker_id
		FROM instances
		WHERE id = $1
	`

	var state ConnectionState
	if err := r.pool.QueryRow(ctx, query, id).Scan(
		&state.Connected,
		&state.Status,
		&state.LastConnectedAt,
		&state.WorkerID,
		&state.DesiredWorkerID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInstanceNotFound
		}
		return nil, fmt.Errorf("get connection state: %w", err)
	}

	return &state, nil
}
