package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles database operations for the message queue
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new queue repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Enqueue adds a new message to the queue
// The message will be processed after scheduledAt time
// Use time.Now() for immediate processing
//
// FIFO by Recipient: Extracts phone number from payload and sets processing_key
// This ensures messages to the same recipient are processed in order
func (r *Repository) Enqueue(ctx context.Context, instanceID uuid.UUID, payload interface{}, scheduledAt time.Time, maxAttempts int) (int64, error) {
	// Serialize payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal payload: %w", err)
	}

	// Check payload size (PostgreSQL JSONB limit is ~1GB, but we'll enforce 10MB for safety)
	const maxPayloadSize = 10 * 1024 * 1024 // 10MB
	if len(payloadJSON) > maxPayloadSize {
		return 0, ErrPayloadTooLarge
	}

	// Extract phone number from payload for FIFO by recipient
	// This enables strict FIFO ordering per phone number across replicas
	var processingKey string
	if args, ok := payload.(*SendMessageArgs); ok {
		processingKey = args.Phone
	} else {
		// Try to unmarshal to extract phone (fallback for interface{} payload)
		var args SendMessageArgs
		if err := json.Unmarshal(payloadJSON, &args); err == nil {
			processingKey = args.Phone
		}
		// If extraction fails, processing_key will be empty string (NULL in DB)
		// This is acceptable as messages without phone will still be processed
	}

	query := `
		INSERT INTO message_queue (
			instance_id,
			payload,
			scheduled_at,
			max_attempts,
			processing_key
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int64
	err = r.pool.QueryRow(ctx, query, instanceID, payloadJSON, scheduledAt, maxAttempts, processingKey).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert message: %w", err)
	}

	return id, nil
}

// Dequeue retrieves the next pending message for an instance
// This uses FOR UPDATE SKIP LOCKED to ensure:
// 1. Only one worker processes a message at a time
// 2. Multiple replicas can safely poll the same queue
// 3. FIFO ordering is maintained via ORDER BY processing_key, id
//
// FIFO by Recipient: Messages to the same phone are processed in order
// Messages to different phones can be processed in parallel
func (r *Repository) Dequeue(ctx context.Context, instanceID uuid.UUID) (*QueueMessage, error) {
	query := `
		UPDATE message_queue
		SET
			status = 'processing',
			last_attempt_at = NOW(),
			attempts = attempts + 1
		WHERE id = (
			SELECT id
			FROM message_queue
			WHERE instance_id = $1
			  AND status = 'pending'
			  AND scheduled_at <= NOW()
			ORDER BY processing_key ASC NULLS LAST, id ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING
			id, instance_id, payload, status, scheduled_at, created_at,
			attempts, max_attempts, last_error, last_attempt_at, processed_at, processing_key
	`

	var msg QueueMessage
	err := r.pool.QueryRow(ctx, query, instanceID).Scan(
		&msg.ID,
		&msg.InstanceID,
		&msg.Payload,
		&msg.Status,
		&msg.ScheduledAt,
		&msg.CreatedAt,
		&msg.Attempts,
		&msg.MaxAttempts,
		&msg.LastError,
		&msg.LastAttemptAt,
		&msg.ProcessedAt,
		&msg.ProcessingKey,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil // No messages available
		}
		return nil, fmt.Errorf("dequeue message: %w", err)
	}

	return &msg, nil
}

// MarkCompleted marks a message as successfully processed
func (r *Repository) MarkCompleted(ctx context.Context, id int64) error {
	query := `
		UPDATE message_queue
		SET
			status = 'completed',
			processed_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark completed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("message %d not found", id)
	}

	return nil
}

// MarkFailed handles a failed message with retry logic
// If attempts < maxAttempts, reschedules with backoff
// Otherwise, moves to DLQ
func (r *Repository) MarkFailed(ctx context.Context, id int64, errorMsg string, retryDelay time.Duration) (bool, error) {
	// First, get current message state
	query := `
		SELECT attempts, max_attempts
		FROM message_queue
		WHERE id = $1
	`

	var attempts, maxAttempts int
	err := r.pool.QueryRow(ctx, query, id).Scan(&attempts, &maxAttempts)
	if err != nil {
		return false, fmt.Errorf("get message state: %w", err)
	}

	// Check if we should retry or move to DLQ
	if attempts >= maxAttempts {
		// Move to DLQ
		return false, r.MoveToDLQ(ctx, id, errorMsg)
	}

	// Reschedule with backoff
	updateQuery := `
		UPDATE message_queue
		SET
			status = 'pending',
			last_error = $1,
			scheduled_at = NOW() + $2::interval
		WHERE id = $3
	`

	_, err = r.pool.Exec(ctx, updateQuery, errorMsg, fmt.Sprintf("%d seconds", int(retryDelay.Seconds())), id)
	if err != nil {
		return false, fmt.Errorf("reschedule message: %w", err)
	}

	return true, nil // Will retry
}

// MoveToDLQ moves a permanently failed message to the Dead Letter Queue
func (r *Repository) MoveToDLQ(ctx context.Context, id int64, errorMsg string) error {
	// Use a transaction to ensure atomic move
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get message details
	selectQuery := `
		SELECT instance_id, payload, attempts, created_at
		FROM message_queue
		WHERE id = $1
	`

	var instanceID uuid.UUID
	var payload json.RawMessage
	var attempts int
	var createdAt time.Time

	err = tx.QueryRow(ctx, selectQuery, id).Scan(&instanceID, &payload, &attempts, &createdAt)
	if err != nil {
		return fmt.Errorf("get message for DLQ: %w", err)
	}

	// Insert into DLQ
	insertQuery := `
		INSERT INTO message_dlq (
			original_id,
			instance_id,
			payload,
			error,
			attempts,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.Exec(ctx, insertQuery, id, instanceID, payload, errorMsg, attempts, createdAt)
	if err != nil {
		return fmt.Errorf("insert into DLQ: %w", err)
	}

	// Mark original message as failed
	updateQuery := `
		UPDATE message_queue
		SET
			status = 'failed',
			last_error = $1,
			processed_at = NOW()
		WHERE id = $2
	`

	_, err = tx.Exec(ctx, updateQuery, errorMsg, id)
	if err != nil {
		return fmt.Errorf("mark message as failed: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// RescheduleOnDisconnect reschedules a message when WhatsApp is disconnected
// This preserves FIFO order by keeping the original ID
func (r *Repository) RescheduleOnDisconnect(ctx context.Context, id int64, retryDelay time.Duration) error {
	query := `
		UPDATE message_queue
		SET
			status = 'pending',
			last_error = 'WhatsApp not connected',
			scheduled_at = NOW() + $1::interval,
			attempts = attempts - 1
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, fmt.Sprintf("%d seconds", int(retryDelay.Seconds())), id)
	if err != nil {
		return fmt.Errorf("reschedule on disconnect: %w", err)
	}

	return nil
}

// CountActiveMessages returns the total number of messages that are pending or currently processing.
func (r *Repository) CountActiveMessages(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM message_queue
		WHERE status IN ('pending', 'processing')
	`

	var count int
	if err := r.pool.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count active messages: %w", err)
	}

	return count, nil
}

// GetStats returns statistics for a specific instance queue
func (r *Repository) GetStats(ctx context.Context, instanceID uuid.UUID) (*QueueStats, error) {
	stats := &QueueStats{
		InstanceID: instanceID,
	}

	// Count by status
	countQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0) as pending,
			COALESCE(SUM(CASE WHEN status = 'processing' THEN 1 ELSE 0 END), 0) as processing,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END), 0) as completed,
			COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0) as failed
		FROM message_queue
		WHERE instance_id = $1
	`

	err := r.pool.QueryRow(ctx, countQuery, instanceID).Scan(
		&stats.PendingCount,
		&stats.ProcessingCount,
		&stats.CompletedCount,
		&stats.FailedCount,
	)
	if err != nil {
		return nil, fmt.Errorf("get queue counts: %w", err)
	}

	// Count DLQ
	dlqQuery := `
		SELECT COUNT(*) FROM message_dlq WHERE instance_id = $1
	`
	err = r.pool.QueryRow(ctx, dlqQuery, instanceID).Scan(&stats.DLQCount)
	if err != nil {
		return nil, fmt.Errorf("get DLQ count: %w", err)
	}

	// Get oldest pending message age
	ageQuery := `
		SELECT EXTRACT(EPOCH FROM (NOW() - created_at))::bigint as age_seconds
		FROM message_queue
		WHERE instance_id = $1 AND status = 'pending'
		ORDER BY created_at ASC
		LIMIT 1
	`

	var ageSeconds *int64
	err = r.pool.QueryRow(ctx, ageQuery, instanceID).Scan(&ageSeconds)
	if err != nil && err.Error() != "no rows in result set" {
		return nil, fmt.Errorf("get oldest pending age: %w", err)
	}

	if ageSeconds != nil {
		age := time.Duration(*ageSeconds) * time.Second
		stats.OldestPendingAge = &age
	}

	// Get average processing time
	avgQuery := `
		SELECT AVG(EXTRACT(EPOCH FROM (processed_at - last_attempt_at)))::bigint
		FROM message_queue
		WHERE instance_id = $1
		  AND status = 'completed'
		  AND processed_at IS NOT NULL
		  AND last_attempt_at IS NOT NULL
	`

	var avgSeconds *int64
	err = r.pool.QueryRow(ctx, avgQuery, instanceID).Scan(&avgSeconds)
	if err != nil && err.Error() != "no rows in result set" {
		return nil, fmt.Errorf("get avg processing time: %w", err)
	}

	if avgSeconds != nil {
		avg := time.Duration(*avgSeconds) * time.Second
		stats.AvgProcessingTime = &avg
	}

	return stats, nil
}

// CleanupOldMessages removes old completed and failed messages
func (r *Repository) CleanupOldMessages(ctx context.Context, completedRetention, failedRetention time.Duration) (int, error) {
	query := `
		DELETE FROM message_queue
		WHERE
			(status = 'completed' AND processed_at < NOW() - $1::interval)
			OR
			(status = 'failed' AND processed_at < NOW() - $2::interval)
	`

	result, err := r.pool.Exec(
		ctx,
		query,
		fmt.Sprintf("%d seconds", int(completedRetention.Seconds())),
		fmt.Sprintf("%d seconds", int(failedRetention.Seconds())),
	)
	if err != nil {
		return 0, fmt.Errorf("cleanup old messages: %w", err)
	}

	return int(result.RowsAffected()), nil
}

// ResetStuckMessages resets messages that have been processing for too long
// This handles cases where a worker crashes mid-processing
func (r *Repository) ResetStuckMessages(ctx context.Context, timeout time.Duration) (int, error) {
	query := `
		UPDATE message_queue
		SET
			status = 'pending',
			scheduled_at = NOW(),
			last_error = 'Processing timeout - worker may have crashed'
		WHERE status = 'processing'
		  AND last_attempt_at < NOW() - $1::interval
	`

	result, err := r.pool.Exec(ctx, query, fmt.Sprintf("%d seconds", int(timeout.Seconds())))
	if err != nil {
		return 0, fmt.Errorf("reset stuck messages: %w", err)
	}

	return int(result.RowsAffected()), nil
}

// ListMessages retrieves messages for an instance with pagination
// Returns messages ordered by ID (FIFO order) with limit and offset
func (r *Repository) ListMessages(ctx context.Context, instanceID uuid.UUID, limit, offset int) ([]*QueueMessage, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM message_queue
		WHERE instance_id = $1
		  AND status IN ('pending', 'processing')
	`

	var total int
	err := r.pool.QueryRow(ctx, countQuery, instanceID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count messages: %w", err)
	}

	// Get messages with pagination
	query := `
		SELECT
			id, instance_id, payload, status, scheduled_at, created_at,
			attempts, max_attempts, last_error, last_attempt_at, processed_at, processing_key
		FROM message_queue
		WHERE instance_id = $1
		  AND status IN ('pending', 'processing')
		ORDER BY processing_key ASC NULLS LAST, id ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, instanceID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	messages := make([]*QueueMessage, 0)
	for rows.Next() {
		var msg QueueMessage
		err := rows.Scan(
			&msg.ID,
			&msg.InstanceID,
			&msg.Payload,
			&msg.Status,
			&msg.ScheduledAt,
			&msg.CreatedAt,
			&msg.Attempts,
			&msg.MaxAttempts,
			&msg.LastError,
			&msg.LastAttemptAt,
			&msg.ProcessedAt,
			&msg.ProcessingKey,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate messages: %w", err)
	}

	return messages, total, nil
}

// DeleteByInstance removes all messages for an instance
// This is used to clear the queue
func (r *Repository) DeleteByInstance(ctx context.Context, instanceID uuid.UUID) error {
	query := `
		DELETE FROM message_queue
		WHERE instance_id = $1
		  AND status IN ('pending', 'processing')
	`

	_, err := r.pool.Exec(ctx, query, instanceID)
	if err != nil {
		return fmt.Errorf("delete messages by instance: %w", err)
	}

	return nil
}

// GetByZaapID retrieves a message by its ZaapID
// This requires unmarshaling the payload to extract ZaapID
func (r *Repository) GetByZaapID(ctx context.Context, zaapID string) (*QueueMessage, error) {
	query := `
		SELECT
			id, instance_id, payload, status, scheduled_at, created_at,
			attempts, max_attempts, last_error, last_attempt_at, processed_at, processing_key
		FROM message_queue
		WHERE payload->>'zaap_id' = $1
		LIMIT 1
	`

	var msg QueueMessage
	err := r.pool.QueryRow(ctx, query, zaapID).Scan(
		&msg.ID,
		&msg.InstanceID,
		&msg.Payload,
		&msg.Status,
		&msg.ScheduledAt,
		&msg.CreatedAt,
		&msg.Attempts,
		&msg.MaxAttempts,
		&msg.LastError,
		&msg.LastAttemptAt,
		&msg.ProcessedAt,
		&msg.ProcessingKey,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, ErrQueueNotFound
		}
		return nil, fmt.Errorf("get message by zaap_id: %w", err)
	}

	return &msg, nil
}

// DeleteByID removes a specific message by its ID
func (r *Repository) DeleteByID(ctx context.Context, id int64) error {
	query := `
		DELETE FROM message_queue
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete message by id: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrQueueNotFound
	}

	return nil
}

// ListenForNotifications establishes a PostgreSQL LISTEN connection
// Returns a channel that receives instance_id notifications
// The caller must call Close() when done to release resources
func (r *Repository) ListenForNotifications(ctx context.Context) (<-chan uuid.UUID, error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}

	_, err = conn.Exec(ctx, "LISTEN message_queue_channel")
	if err != nil {
		conn.Release()
		return nil, fmt.Errorf("listen on channel: %w", err)
	}

	notificationChan := make(chan uuid.UUID, 100) // Buffer to prevent blocking

	// Start goroutine to receive notifications
	go func() {
		defer conn.Release()
		defer close(notificationChan)

		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				// Context cancelled or connection closed
				return
			}

			// Parse instance_id from payload
			if instanceID, err := uuid.Parse(notification.Payload); err == nil {
				select {
				case notificationChan <- instanceID:
				case <-ctx.Done():
					return
				default:
					// Buffer full, skip this notification
					// Worker will poll as fallback
				}
			}
		}
	}()

	return notificationChan, nil
}
