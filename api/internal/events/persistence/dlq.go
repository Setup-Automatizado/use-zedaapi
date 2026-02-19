package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDLQEventNotFound = errors.New("dlq event not found")
)

// RawDLQEntry represents a DLQ event inserted from raw data (e.g. from NATS DLQ handler).
type RawDLQEntry struct {
	InstanceID    uuid.UUID
	EventID       uuid.UUID
	EventType     string
	SourceLib     string
	Payload       json.RawMessage
	Metadata      json.RawMessage
	FailureReason string // Category of failure (e.g. "max_retries_exceeded", "permanent_4xx")
	LastError     string // Full error message
	Attempts      int
}

// DLQInstanceStats holds aggregated DLQ statistics for an instance.
type DLQInstanceStats struct {
	TotalEvents     int
	ByStatus        map[string]int
	ByEventType     map[string]int
	ByFailureReason map[string]int
}

// DLQFilter holds filter parameters for DLQ event queries.
type DLQFilter struct {
	Status    *DLQReprocessStatus
	EventType *string
	Limit     int
	Offset    int
}

type DLQReprocessStatus string

const (
	DLQReprocessPending    DLQReprocessStatus = "pending"
	DLQReprocessProcessing DLQReprocessStatus = "processing"
	DLQReprocessSuccess    DLQReprocessStatus = "success"
	DLQReprocessFailed     DLQReprocessStatus = "failed"
	DLQReprocessDiscarded  DLQReprocessStatus = "discarded"
)

type DLQEvent struct {
	ID                     int64
	InstanceID             uuid.UUID
	EventID                uuid.UUID
	EventType              string
	SourceLib              string
	OriginalPayload        json.RawMessage
	OriginalMetadata       json.RawMessage
	OriginalSequenceNumber int64
	FailureReason          string
	LastError              string
	TotalAttempts          int
	AttemptHistory         json.RawMessage
	TransportType          TransportType
	TransportConfig        json.RawMessage
	LastTransportResponse  json.RawMessage
	FirstAttemptAt         time.Time
	LastAttemptAt          time.Time
	MovedToDLQAt           time.Time
	ReprocessStatus        DLQReprocessStatus
	ReprocessedAt          *time.Time
	ReprocessResult        *string
	ReprocessAttempts      int
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type DLQRepository interface {
	InsertFromOutbox(ctx context.Context, event *OutboxEvent, failureReason string, attemptHistory json.RawMessage) error

	InsertRaw(ctx context.Context, entry *RawDLQEntry) error

	GetPendingReprocessEvents(ctx context.Context, limit int) ([]*DLQEvent, error)

	UpdateReprocessStatus(ctx context.Context, eventID uuid.UUID, status DLQReprocessStatus, result *string) error

	GetEventByID(ctx context.Context, eventID uuid.UUID) (*DLQEvent, error)

	GetByInstanceID(ctx context.Context, instanceID uuid.UUID, limit, offset int) ([]*DLQEvent, error)

	CountByInstanceID(ctx context.Context, instanceID uuid.UUID) (int, error)

	GetStatsByInstance(ctx context.Context, instanceID uuid.UUID) (*DLQInstanceStats, error)

	GetByInstanceIDFiltered(ctx context.Context, instanceID uuid.UUID, filter DLQFilter) ([]*DLQEvent, int, error)

	GetFailureStats(ctx context.Context, since time.Time) (map[string]int, error)

	MarkDiscarded(ctx context.Context, eventID uuid.UUID) error

	MarkReprocessing(ctx context.Context, eventID uuid.UUID) error

	GetRetryableByInstance(ctx context.Context, instanceID uuid.UUID, limit int) ([]*DLQEvent, int, error)

	DeleteOldEvents(ctx context.Context, olderThan time.Time) (int64, error)

	DeleteResolvedByInstance(ctx context.Context, instanceID uuid.UUID, olderThan time.Time) (int64, error)
}

type dlqRepository struct {
	pool *pgxpool.Pool
}

func NewDLQRepository(pool *pgxpool.Pool) DLQRepository {
	return &dlqRepository{pool: pool}
}

func (r *dlqRepository) InsertFromOutbox(ctx context.Context, event *OutboxEvent, failureReason string, attemptHistory json.RawMessage) error {
	// Use ON CONFLICT to handle race conditions where multiple workers
	// may try to insert the same event into DLQ simultaneously.
	// This makes the operation idempotent - subsequent inserts update the existing record.
	query := `
		INSERT INTO event_dlq (
			instance_id, event_id, event_type, source_lib,
			original_payload, original_metadata, original_sequence_number,
			failure_reason, last_error, total_attempts, attempt_history,
			transport_type, transport_config, last_transport_response,
			first_attempt_at, last_attempt_at, moved_to_dlq_at,
			reprocess_status, reprocess_attempts
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), 'pending', 0
		)
		ON CONFLICT (event_id) DO UPDATE SET
			failure_reason = EXCLUDED.failure_reason,
			last_error = EXCLUDED.last_error,
			total_attempts = EXCLUDED.total_attempts,
			attempt_history = EXCLUDED.attempt_history,
			last_transport_response = EXCLUDED.last_transport_response,
			last_attempt_at = EXCLUDED.last_attempt_at,
			updated_at = NOW()`

	firstAttemptAt := event.CreatedAt
	lastError := ""
	if event.LastError != nil {
		lastError = *event.LastError
	}

	_, err := r.pool.Exec(
		ctx, query,
		event.InstanceID, event.EventID, event.EventType, event.SourceLib,
		event.Payload, event.Metadata, event.SequenceNumber,
		failureReason, lastError, event.Attempts, attemptHistory,
		event.TransportType, event.TransportConfig, event.TransportResponse,
		firstAttemptAt, event.UpdatedAt,
	)

	return err
}

func (r *dlqRepository) GetPendingReprocessEvents(ctx context.Context, limit int) ([]*DLQEvent, error) {
	query := `
		SELECT
			id, instance_id, event_id, event_type, source_lib,
			original_payload, original_metadata, original_sequence_number,
			failure_reason, last_error, total_attempts, attempt_history,
			transport_type, transport_config, last_transport_response,
			first_attempt_at, last_attempt_at, moved_to_dlq_at,
			reprocess_status, reprocessed_at, reprocess_result, reprocess_attempts,
			created_at, updated_at
		FROM event_dlq
		WHERE reprocess_status = 'pending'
		ORDER BY moved_to_dlq_at ASC
		LIMIT $1`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*DLQEvent
	for rows.Next() {
		event := &DLQEvent{}
		err := rows.Scan(
			&event.ID, &event.InstanceID, &event.EventID, &event.EventType, &event.SourceLib,
			&event.OriginalPayload, &event.OriginalMetadata, &event.OriginalSequenceNumber,
			&event.FailureReason, &event.LastError, &event.TotalAttempts, &event.AttemptHistory,
			&event.TransportType, &event.TransportConfig, &event.LastTransportResponse,
			&event.FirstAttemptAt, &event.LastAttemptAt, &event.MovedToDLQAt,
			&event.ReprocessStatus, &event.ReprocessedAt, &event.ReprocessResult, &event.ReprocessAttempts,
			&event.CreatedAt, &event.UpdatedAt,
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

func (r *dlqRepository) UpdateReprocessStatus(ctx context.Context, eventID uuid.UUID, status DLQReprocessStatus, result *string) error {
	query := `
		UPDATE event_dlq
		SET reprocess_status = $2,
		    reprocessed_at = NOW(),
		    reprocess_result = $3,
		    reprocess_attempts = reprocess_attempts + 1
		WHERE event_id = $1`

	commandTag, err := r.pool.Exec(ctx, query, eventID, status, result)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrDLQEventNotFound
	}

	return nil
}

func (r *dlqRepository) GetEventByID(ctx context.Context, eventID uuid.UUID) (*DLQEvent, error) {
	query := `
		SELECT
			id, instance_id, event_id, event_type, source_lib,
			original_payload, original_metadata, original_sequence_number,
			failure_reason, last_error, total_attempts, attempt_history,
			transport_type, transport_config, last_transport_response,
			first_attempt_at, last_attempt_at, moved_to_dlq_at,
			reprocess_status, reprocessed_at, reprocess_result, reprocess_attempts,
			created_at, updated_at
		FROM event_dlq
		WHERE event_id = $1`

	event := &DLQEvent{}
	err := r.pool.QueryRow(ctx, query, eventID).Scan(
		&event.ID, &event.InstanceID, &event.EventID, &event.EventType, &event.SourceLib,
		&event.OriginalPayload, &event.OriginalMetadata, &event.OriginalSequenceNumber,
		&event.FailureReason, &event.LastError, &event.TotalAttempts, &event.AttemptHistory,
		&event.TransportType, &event.TransportConfig, &event.LastTransportResponse,
		&event.FirstAttemptAt, &event.LastAttemptAt, &event.MovedToDLQAt,
		&event.ReprocessStatus, &event.ReprocessedAt, &event.ReprocessResult, &event.ReprocessAttempts,
		&event.CreatedAt, &event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDLQEventNotFound
		}
		return nil, err
	}

	return event, nil
}

func (r *dlqRepository) GetByInstanceID(ctx context.Context, instanceID uuid.UUID, limit, offset int) ([]*DLQEvent, error) {
	query := `
		SELECT
			id, instance_id, event_id, event_type, source_lib,
			original_payload, original_metadata, original_sequence_number,
			failure_reason, last_error, total_attempts, attempt_history,
			transport_type, transport_config, last_transport_response,
			first_attempt_at, last_attempt_at, moved_to_dlq_at,
			reprocess_status, reprocessed_at, reprocess_result, reprocess_attempts,
			created_at, updated_at
		FROM event_dlq
		WHERE instance_id = $1
		ORDER BY moved_to_dlq_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, instanceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*DLQEvent
	for rows.Next() {
		event := &DLQEvent{}
		err := rows.Scan(
			&event.ID, &event.InstanceID, &event.EventID, &event.EventType, &event.SourceLib,
			&event.OriginalPayload, &event.OriginalMetadata, &event.OriginalSequenceNumber,
			&event.FailureReason, &event.LastError, &event.TotalAttempts, &event.AttemptHistory,
			&event.TransportType, &event.TransportConfig, &event.LastTransportResponse,
			&event.FirstAttemptAt, &event.LastAttemptAt, &event.MovedToDLQAt,
			&event.ReprocessStatus, &event.ReprocessedAt, &event.ReprocessResult, &event.ReprocessAttempts,
			&event.CreatedAt, &event.UpdatedAt,
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

func (r *dlqRepository) CountByInstanceID(ctx context.Context, instanceID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM event_dlq
		WHERE instance_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, instanceID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *dlqRepository) GetFailureStats(ctx context.Context, since time.Time) (map[string]int, error) {
	query := `
		SELECT event_type, COUNT(*)
		FROM event_dlq
		WHERE moved_to_dlq_at >= $1
		GROUP BY event_type
		ORDER BY COUNT(*) DESC`

	rows, err := r.pool.Query(ctx, query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var eventType string
		var count int
		if err := rows.Scan(&eventType, &count); err != nil {
			return nil, err
		}
		stats[eventType] = count
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *dlqRepository) MarkDiscarded(ctx context.Context, eventID uuid.UUID) error {
	query := `
		UPDATE event_dlq
		SET reprocess_status = 'discarded',
		    updated_at = NOW()
		WHERE event_id = $1`

	commandTag, err := r.pool.Exec(ctx, query, eventID)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrDLQEventNotFound
	}

	return nil
}

func (r *dlqRepository) DeleteOldEvents(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `
		DELETE FROM event_dlq
		WHERE moved_to_dlq_at < $1
		  AND reprocess_status IN ('success', 'discarded')`

	commandTag, err := r.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}

	return commandTag.RowsAffected(), nil
}

func (r *dlqRepository) InsertRaw(ctx context.Context, entry *RawDLQEntry) error {
	query := `
		INSERT INTO event_dlq (
			instance_id, event_id, event_type, source_lib,
			original_payload, original_metadata,
			failure_reason, last_error, total_attempts,
			first_attempt_at, last_attempt_at, moved_to_dlq_at,
			reprocess_status, reprocess_attempts
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW(), NOW(), 'pending', 0
		)
		ON CONFLICT (event_id) DO UPDATE SET
			failure_reason = EXCLUDED.failure_reason,
			last_error = EXCLUDED.last_error,
			total_attempts = EXCLUDED.total_attempts,
			last_attempt_at = NOW(),
			updated_at = NOW()`

	payload := entry.Payload
	if payload == nil {
		payload = json.RawMessage("null")
	}
	metadata := entry.Metadata
	if metadata == nil {
		metadata = json.RawMessage("null")
	}

	_, err := r.pool.Exec(
		ctx, query,
		entry.InstanceID, entry.EventID, entry.EventType, entry.SourceLib,
		payload, metadata,
		entry.FailureReason, entry.LastError, entry.Attempts,
	)
	return err
}

func (r *dlqRepository) GetStatsByInstance(ctx context.Context, instanceID uuid.UUID) (*DLQInstanceStats, error) {
	stats := &DLQInstanceStats{
		ByStatus:        make(map[string]int),
		ByEventType:     make(map[string]int),
		ByFailureReason: make(map[string]int),
	}

	// Single query using GROUPING SETS for all 3 breakdowns in one round trip.
	// dimension: 'status', 'event_type', or 'failure_reason'
	// key: the grouped value
	// cnt: count for that group
	query := `
		SELECT dimension, key, cnt FROM (
			SELECT 'status' AS dimension, reprocess_status AS key, COUNT(*) AS cnt
			FROM event_dlq WHERE instance_id = $1
			GROUP BY reprocess_status
			UNION ALL
			SELECT 'event_type', event_type, COUNT(*)
			FROM event_dlq WHERE instance_id = $1
			GROUP BY event_type
			UNION ALL
			SELECT 'failure_reason', failure_reason, COUNT(*)
			FROM event_dlq WHERE instance_id = $1
			GROUP BY failure_reason
		) sub
		ORDER BY dimension, cnt DESC`

	rows, err := r.pool.Query(ctx, query, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		var dimension, key string
		var cnt int
		if err := rows.Scan(&dimension, &key, &cnt); err != nil {
			return nil, err
		}
		switch dimension {
		case "status":
			stats.ByStatus[key] = cnt
			total += cnt
		case "event_type":
			stats.ByEventType[key] = cnt
		case "failure_reason":
			stats.ByFailureReason[key] = cnt
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	stats.TotalEvents = total

	return stats, nil
}

func (r *dlqRepository) MarkReprocessing(ctx context.Context, eventID uuid.UUID) error {
	query := `
		UPDATE event_dlq
		SET reprocess_status = 'processing',
		    updated_at = NOW()
		WHERE event_id = $1`

	commandTag, err := r.pool.Exec(ctx, query, eventID)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrDLQEventNotFound
	}

	return nil
}

func (r *dlqRepository) GetRetryableByInstance(ctx context.Context, instanceID uuid.UUID, limit int) ([]*DLQEvent, int, error) {
	// Count total retryable
	countQuery := `
		SELECT COUNT(*)
		FROM event_dlq
		WHERE instance_id = $1
		  AND reprocess_status IN ('pending', 'failed')`

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, instanceID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch up to limit
	dataQuery := `
		SELECT
			id, instance_id, event_id, event_type, source_lib,
			original_payload, original_metadata, original_sequence_number,
			failure_reason, last_error, total_attempts, attempt_history,
			transport_type, transport_config, last_transport_response,
			first_attempt_at, last_attempt_at, moved_to_dlq_at,
			reprocess_status, reprocessed_at, reprocess_result, reprocess_attempts,
			created_at, updated_at
		FROM event_dlq
		WHERE instance_id = $1
		  AND reprocess_status IN ('pending', 'failed')
		ORDER BY moved_to_dlq_at ASC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, dataQuery, instanceID, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []*DLQEvent
	for rows.Next() {
		event := &DLQEvent{}
		err := rows.Scan(
			&event.ID, &event.InstanceID, &event.EventID, &event.EventType, &event.SourceLib,
			&event.OriginalPayload, &event.OriginalMetadata, &event.OriginalSequenceNumber,
			&event.FailureReason, &event.LastError, &event.TotalAttempts, &event.AttemptHistory,
			&event.TransportType, &event.TransportConfig, &event.LastTransportResponse,
			&event.FirstAttemptAt, &event.LastAttemptAt, &event.MovedToDLQAt,
			&event.ReprocessStatus, &event.ReprocessedAt, &event.ReprocessResult, &event.ReprocessAttempts,
			&event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *dlqRepository) GetByInstanceIDFiltered(ctx context.Context, instanceID uuid.UUID, filter DLQFilter) ([]*DLQEvent, int, error) {
	// Build WHERE clause dynamically
	args := []interface{}{instanceID}
	where := "WHERE instance_id = $1"
	argIdx := 2

	if filter.Status != nil {
		where += fmt.Sprintf(" AND reprocess_status = $%d", argIdx)
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.EventType != nil {
		where += fmt.Sprintf(" AND event_type = $%d", argIdx)
		args = append(args, *filter.EventType)
		argIdx++
	}

	// Count total matching
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM event_dlq %s", where)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch page
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	dataQuery := fmt.Sprintf(`
		SELECT
			id, instance_id, event_id, event_type, source_lib,
			original_payload, original_metadata, original_sequence_number,
			failure_reason, last_error, total_attempts, attempt_history,
			transport_type, transport_config, last_transport_response,
			first_attempt_at, last_attempt_at, moved_to_dlq_at,
			reprocess_status, reprocessed_at, reprocess_result, reprocess_attempts,
			created_at, updated_at
		FROM event_dlq
		%s
		ORDER BY moved_to_dlq_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []*DLQEvent
	for rows.Next() {
		event := &DLQEvent{}
		err := rows.Scan(
			&event.ID, &event.InstanceID, &event.EventID, &event.EventType, &event.SourceLib,
			&event.OriginalPayload, &event.OriginalMetadata, &event.OriginalSequenceNumber,
			&event.FailureReason, &event.LastError, &event.TotalAttempts, &event.AttemptHistory,
			&event.TransportType, &event.TransportConfig, &event.LastTransportResponse,
			&event.FirstAttemptAt, &event.LastAttemptAt, &event.MovedToDLQAt,
			&event.ReprocessStatus, &event.ReprocessedAt, &event.ReprocessResult, &event.ReprocessAttempts,
			&event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *dlqRepository) DeleteResolvedByInstance(ctx context.Context, instanceID uuid.UUID, olderThan time.Time) (int64, error) {
	query := `
		DELETE FROM event_dlq
		WHERE instance_id = $1
		  AND moved_to_dlq_at < $2
		  AND reprocess_status IN ('success', 'discarded')`

	commandTag, err := r.pool.Exec(ctx, query, instanceID, olderThan)
	if err != nil {
		return 0, err
	}

	return commandTag.RowsAffected(), nil
}
