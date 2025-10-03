package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrEventNotFound    = errors.New("event not found")
	ErrSequenceConflict = errors.New("sequence number conflict")
)

// EventStatus represents the status of an event in the outbox
type EventStatus string

const (
	EventStatusPending    EventStatus = "pending"
	EventStatusProcessing EventStatus = "processing"
	EventStatusRetrying   EventStatus = "retrying"
	EventStatusDelivered  EventStatus = "delivered"
	EventStatusFailed     EventStatus = "failed"
)

// TransportType represents the delivery mechanism
type TransportType string

const (
	TransportWebhook  TransportType = "webhook"
	TransportRabbitMQ TransportType = "rabbitmq"
	TransportSQS      TransportType = "sqs"
	TransportNATS     TransportType = "nats"
	TransportKafka    TransportType = "kafka"
)

// OutboxEvent represents an event in the event_outbox table
type OutboxEvent struct {
	ID                int64
	InstanceID        uuid.UUID
	EventID           uuid.UUID
	EventType         string
	SourceLib         string
	Payload           json.RawMessage
	Metadata          json.RawMessage
	SequenceNumber    int64
	CreatedAt         time.Time
	Status            EventStatus
	Attempts          int
	MaxAttempts       int
	NextAttemptAt     *time.Time
	HasMedia          bool
	MediaProcessed    bool
	MediaURL          *string
	MediaError        *string
	DeliveredAt       *time.Time
	TransportType     TransportType
	TransportConfig   json.RawMessage
	TransportResponse json.RawMessage
	LastError         *string
	UpdatedAt         time.Time
}

// OutboxRepository defines operations for event_outbox table
type OutboxRepository interface {
	// InsertEvent inserts a new event with automatic sequence generation
	InsertEvent(ctx context.Context, event *OutboxEvent) error

	// PollPendingEvents retrieves pending events for an instance, ordered by sequence
	PollPendingEvents(ctx context.Context, instanceID uuid.UUID, limit int) ([]*OutboxEvent, error)

	// UpdateEventStatus updates event status and attempts
	UpdateEventStatus(ctx context.Context, eventID uuid.UUID, status EventStatus, attempts int, nextAttempt *time.Time, lastError *string) error

	// UpdateMediaInfo updates media processing information
	UpdateMediaInfo(ctx context.Context, eventID uuid.UUID, mediaURL, mediaError *string, processed bool) error

	// MarkDelivered marks event as successfully delivered
	MarkDelivered(ctx context.Context, eventID uuid.UUID, transportResponse json.RawMessage) error

	// GetEventByID retrieves a single event by ID
	GetEventByID(ctx context.Context, eventID uuid.UUID) (*OutboxEvent, error)

	// CountPendingByInstance counts pending events for an instance
	CountPendingByInstance(ctx context.Context, instanceID uuid.UUID) (int, error)

	// DeleteDeliveredEvents removes delivered events older than retention period
	DeleteDeliveredEvents(ctx context.Context, olderThan time.Time) (int64, error)

	// GetOldestSequence gets the oldest pending sequence number for an instance
	GetOldestSequence(ctx context.Context, instanceID uuid.UUID) (int64, error)
}

// outboxRepository implements OutboxRepository using pgx
type outboxRepository struct {
	pool *pgxpool.Pool
}

// NewOutboxRepository creates a new OutboxRepository
func NewOutboxRepository(pool *pgxpool.Pool) OutboxRepository {
	return &outboxRepository{pool: pool}
}

func (r *outboxRepository) InsertEvent(ctx context.Context, event *OutboxEvent) error {
	// Generate sequence number atomically
	var sequenceNumber int64
	err := r.pool.QueryRow(ctx, "SELECT get_next_event_sequence($1)", event.InstanceID).Scan(&sequenceNumber)
	if err != nil {
		return err
	}

	event.SequenceNumber = sequenceNumber

	query := `
		INSERT INTO event_outbox (
			instance_id, event_id, event_type, source_lib,
			payload, metadata, sequence_number,
			status, attempts, max_attempts, next_attempt_at,
			has_media, media_processed, media_url, media_error,
			transport_type, transport_config
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		) RETURNING id, created_at, updated_at`

	err = r.pool.QueryRow(
		ctx, query,
		event.InstanceID, event.EventID, event.EventType, event.SourceLib,
		event.Payload, event.Metadata, event.SequenceNumber,
		event.Status, event.Attempts, event.MaxAttempts, event.NextAttemptAt,
		event.HasMedia, event.MediaProcessed, event.MediaURL, event.MediaError,
		event.TransportType, event.TransportConfig,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *outboxRepository) PollPendingEvents(ctx context.Context, instanceID uuid.UUID, limit int) ([]*OutboxEvent, error) {
	query := `
		SELECT
			id, instance_id, event_id, event_type, source_lib,
			payload, metadata, sequence_number, created_at,
			status, attempts, max_attempts, next_attempt_at,
			has_media, media_processed, media_url, media_error,
			delivered_at, transport_type, transport_config,
			transport_response, last_error, updated_at
		FROM event_outbox
		WHERE instance_id = $1
		  AND status IN ('pending', 'retrying')
		  AND (next_attempt_at IS NULL OR next_attempt_at <= NOW())
		  AND attempts < max_attempts
		  AND (has_media = FALSE OR media_processed = TRUE)
		ORDER BY sequence_number ASC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, instanceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*OutboxEvent
	for rows.Next() {
		event := &OutboxEvent{}
		err := rows.Scan(
			&event.ID, &event.InstanceID, &event.EventID, &event.EventType, &event.SourceLib,
			&event.Payload, &event.Metadata, &event.SequenceNumber, &event.CreatedAt,
			&event.Status, &event.Attempts, &event.MaxAttempts, &event.NextAttemptAt,
			&event.HasMedia, &event.MediaProcessed, &event.MediaURL, &event.MediaError,
			&event.DeliveredAt, &event.TransportType, &event.TransportConfig,
			&event.TransportResponse, &event.LastError, &event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *outboxRepository) UpdateEventStatus(ctx context.Context, eventID uuid.UUID, status EventStatus, attempts int, nextAttempt *time.Time, lastError *string) error {
	query := `
		UPDATE event_outbox
		SET status = $2, attempts = $3, next_attempt_at = $4, last_error = $5
		WHERE event_id = $1`

	result, err := r.pool.Exec(ctx, query, eventID, status, attempts, nextAttempt, lastError)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrEventNotFound
	}

	return nil
}

func (r *outboxRepository) UpdateMediaInfo(ctx context.Context, eventID uuid.UUID, mediaURL, mediaError *string, processed bool) error {
	query := `
		UPDATE event_outbox
		SET media_url = $2, media_error = $3, media_processed = $4
		WHERE event_id = $1`

	result, err := r.pool.Exec(ctx, query, eventID, mediaURL, mediaError, processed)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrEventNotFound
	}

	return nil
}

func (r *outboxRepository) MarkDelivered(ctx context.Context, eventID uuid.UUID, transportResponse json.RawMessage) error {
	query := `
		UPDATE event_outbox
		SET status = 'delivered',
		    delivered_at = NOW(),
		    transport_response = $2
		WHERE event_id = $1`

	result, err := r.pool.Exec(ctx, query, eventID, transportResponse)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrEventNotFound
	}

	return nil
}

func (r *outboxRepository) GetEventByID(ctx context.Context, eventID uuid.UUID) (*OutboxEvent, error) {
	query := `
		SELECT
			id, instance_id, event_id, event_type, source_lib,
			payload, metadata, sequence_number, created_at,
			status, attempts, max_attempts, next_attempt_at,
			has_media, media_processed, media_url, media_error,
			delivered_at, transport_type, transport_config,
			transport_response, last_error, updated_at
		FROM event_outbox
		WHERE event_id = $1`

	event := &OutboxEvent{}
	err := r.pool.QueryRow(ctx, query, eventID).Scan(
		&event.ID, &event.InstanceID, &event.EventID, &event.EventType, &event.SourceLib,
		&event.Payload, &event.Metadata, &event.SequenceNumber, &event.CreatedAt,
		&event.Status, &event.Attempts, &event.MaxAttempts, &event.NextAttemptAt,
		&event.HasMedia, &event.MediaProcessed, &event.MediaURL, &event.MediaError,
		&event.DeliveredAt, &event.TransportType, &event.TransportConfig,
		&event.TransportResponse, &event.LastError, &event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}

	return event, nil
}

func (r *outboxRepository) CountPendingByInstance(ctx context.Context, instanceID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM event_outbox
		WHERE instance_id = $1
		  AND status IN ('pending', 'retrying')
		  AND attempts < max_attempts`

	var count int
	err := r.pool.QueryRow(ctx, query, instanceID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *outboxRepository) DeleteDeliveredEvents(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `
		DELETE FROM event_outbox
		WHERE status = 'delivered'
		  AND delivered_at < $1`

	result, err := r.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (r *outboxRepository) GetOldestSequence(ctx context.Context, instanceID uuid.UUID) (int64, error) {
	query := `
		SELECT MIN(sequence_number)
		FROM event_outbox
		WHERE instance_id = $1
		  AND status IN ('pending', 'retrying')`

	var sequence *int64
	err := r.pool.QueryRow(ctx, query, instanceID).Scan(&sequence)
	if err != nil {
		return 0, err
	}

	if sequence == nil {
		return 0, nil
	}

	return *sequence, nil
}
