-- +goose Up
-- +goose StatementBegin

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Instances: WhatsApp client instances
CREATE TABLE IF NOT EXISTS instances (
    id UUID PRIMARY KEY,
    name TEXT,
    session_name TEXT,
    client_token TEXT NOT NULL,
    instance_token TEXT NOT NULL,
    store_jid TEXT,
    is_device BOOLEAN NOT NULL DEFAULT FALSE,
    business_device BOOLEAN NOT NULL DEFAULT FALSE,
    subscription_active BOOLEAN NOT NULL DEFAULT FALSE,
    call_reject_auto BOOLEAN NOT NULL DEFAULT FALSE,
    call_reject_message TEXT,
    auto_read_message BOOLEAN NOT NULL DEFAULT FALSE,
    canceled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_instances_client_token ON instances(client_token);
CREATE UNIQUE INDEX IF NOT EXISTS idx_instances_instance_token ON instances(instance_token);

-- Webhook Configs: Per-instance webhook configuration
CREATE TABLE IF NOT EXISTS webhook_configs (
    instance_id UUID PRIMARY KEY REFERENCES instances(id) ON DELETE CASCADE,
    delivery_url TEXT,
    received_url TEXT,
    received_delivery_url TEXT,
    message_status_url TEXT,
    disconnected_url TEXT,
    chat_presence_url TEXT,
    connected_url TEXT,
    notify_sent_by_me BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Legacy Webhook Outbox (kept for backward compatibility, superseded by event_outbox)
CREATE TABLE IF NOT EXISTS webhook_outbox (
    id UUID PRIMARY KEY,
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhook_outbox_status_next_attempt ON webhook_outbox(status, next_attempt_at);
CREATE INDEX IF NOT EXISTS idx_webhook_outbox_instance_created ON webhook_outbox(instance_id, created_at DESC);

-- Legacy Webhook DLQ (kept for backward compatibility, superseded by event_dlq)
CREATE TABLE IF NOT EXISTS webhook_dlq (
    id UUID PRIMARY KEY,
    instance_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    failure_reason TEXT,
    failed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Instance Events Log: Audit trail
CREATE TABLE IF NOT EXISTS instance_events_log (
    id BIGSERIAL PRIMARY KEY,
    instance_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Event Outbox: Primary queue for all WhatsApp events
-- Guarantees ordered processing per instance using sequence numbers
CREATE TABLE IF NOT EXISTS event_outbox (
    id BIGSERIAL PRIMARY KEY,
    instance_id UUID NOT NULL,
    event_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    source_lib VARCHAR(20) NOT NULL DEFAULT 'whatsmeow',

    -- Event data (Z-API compatible schema)
    payload JSONB NOT NULL,
    metadata JSONB,

    -- Ordering control (critical for maintaining event sequence)
    sequence_number BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Processing status
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 6,
    next_attempt_at TIMESTAMP,

    -- Media tracking
    has_media BOOLEAN NOT NULL DEFAULT FALSE,
    media_processed BOOLEAN NOT NULL DEFAULT FALSE,
    media_url TEXT,
    media_error TEXT,

    -- Delivery tracking
    delivered_at TIMESTAMP,
    transport_type VARCHAR(20) NOT NULL DEFAULT 'webhook',
    transport_config JSONB,
    transport_response JSONB,
    last_error TEXT,

    -- Audit
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT event_outbox_status_check CHECK (status IN ('pending', 'processing', 'retrying', 'delivered', 'failed')),
    CONSTRAINT event_outbox_sequence_unique UNIQUE (instance_id, sequence_number)
);

-- Indexes for event_outbox
CREATE INDEX IF NOT EXISTS idx_outbox_instance_pending
    ON event_outbox(instance_id, status, next_attempt_at, sequence_number)
    WHERE status IN ('pending', 'retrying');

CREATE INDEX IF NOT EXISTS idx_outbox_media_pending
    ON event_outbox(instance_id, id)
    WHERE has_media = TRUE AND media_processed = FALSE AND status IN ('pending', 'processing');

CREATE INDEX IF NOT EXISTS idx_outbox_instance_recent
    ON event_outbox(instance_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_outbox_event_type
    ON event_outbox(event_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_outbox_failed
    ON event_outbox(instance_id, created_at DESC)
    WHERE status = 'failed';

-- Dead Letter Queue: Events that failed after maximum retry attempts
-- Allows manual inspection and reprocessing
CREATE TABLE IF NOT EXISTS event_dlq (
    id BIGSERIAL PRIMARY KEY,
    instance_id UUID NOT NULL,
    event_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    source_lib VARCHAR(20) NOT NULL DEFAULT 'whatsmeow',

    -- Original event data (preserved for debugging)
    original_payload JSONB NOT NULL,
    original_metadata JSONB,
    original_sequence_number BIGINT NOT NULL,

    -- Failure tracking
    failure_reason TEXT NOT NULL,
    last_error TEXT NOT NULL,
    total_attempts INT NOT NULL,
    attempt_history JSONB NOT NULL DEFAULT '[]'::JSONB,

    -- Transport context (helps diagnose delivery issues)
    transport_type VARCHAR(20) NOT NULL,
    transport_config JSONB,
    last_transport_response JSONB,

    -- Timestamps
    first_attempt_at TIMESTAMP NOT NULL,
    last_attempt_at TIMESTAMP NOT NULL,
    moved_to_dlq_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Reprocessing control
    reprocess_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reprocessed_at TIMESTAMP,
    reprocess_result TEXT,
    reprocess_attempts INT NOT NULL DEFAULT 0,

    -- Audit
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT event_dlq_event_unique UNIQUE (event_id),
    CONSTRAINT event_dlq_reprocess_status_check CHECK (reprocess_status IN ('pending', 'processing', 'success', 'failed', 'discarded'))
);

-- Indexes for event_dlq
CREATE INDEX IF NOT EXISTS idx_dlq_instance
    ON event_dlq(instance_id, moved_to_dlq_at DESC);

CREATE INDEX IF NOT EXISTS idx_dlq_reprocess_pending
    ON event_dlq(reprocess_status, moved_to_dlq_at)
    WHERE reprocess_status = 'pending';

CREATE INDEX IF NOT EXISTS idx_dlq_event_type
    ON event_dlq(event_type, moved_to_dlq_at DESC);

CREATE INDEX IF NOT EXISTS idx_dlq_failure_reason
    ON event_dlq(failure_reason, moved_to_dlq_at DESC);

-- Media Metadata: Tracks WhatsApp media download and S3 upload status
-- Enables parallel media processing without blocking event delivery
CREATE TABLE IF NOT EXISTS media_metadata (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL,
    instance_id UUID NOT NULL,

    -- WhatsApp media reference (from message proto)
    media_key TEXT NOT NULL,
    file_sha256 TEXT,
    file_enc_sha256 TEXT,
    direct_path TEXT NOT NULL,
    media_type VARCHAR(20) NOT NULL,
    mime_type VARCHAR(100),
    file_length BIGINT,

    -- Download tracking
    download_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    download_attempts INT NOT NULL DEFAULT 0,
    download_started_at TIMESTAMP,
    downloaded_at TIMESTAMP,
    download_duration_ms INT,
    downloaded_size_bytes BIGINT,
    download_error TEXT,

    -- S3 upload tracking
    s3_bucket VARCHAR(100),
    s3_key TEXT,
    s3_url TEXT,
    s3_url_type VARCHAR(20) DEFAULT 'presigned',
    url_expires_at TIMESTAMP,
    upload_started_at TIMESTAMP,
    uploaded_at TIMESTAMP,
    upload_duration_ms INT,
    upload_error TEXT,

    -- Storage fallback support (from migration 000006)
    storage_type TEXT DEFAULT 's3' CHECK (storage_type IN ('s3', 'local', 'null')),
    fallback_attempted BOOLEAN DEFAULT false,
    fallback_error TEXT,

    -- Processing status
    processing_worker_id VARCHAR(100),
    processing_started_at TIMESTAMP,
    completed_at TIMESTAMP,

    -- Retry control
    next_retry_at TIMESTAMP,
    max_retries INT NOT NULL DEFAULT 3,

    -- Audit
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT media_event_unique UNIQUE (event_id),
    CONSTRAINT media_download_status_check CHECK (download_status IN ('pending', 'downloading', 'downloaded', 'failed', 'completed')),
    CONSTRAINT media_type_check CHECK (media_type IN ('image', 'video', 'audio', 'document', 'sticker', 'voice'))
);

-- Indexes for media_metadata
CREATE INDEX IF NOT EXISTS idx_media_pending
    ON media_metadata(download_status, next_retry_at, created_at)
    WHERE download_status IN ('pending', 'failed')
    AND download_attempts < max_retries;

CREATE INDEX IF NOT EXISTS idx_media_by_event
    ON media_metadata(event_id, download_status);

CREATE INDEX IF NOT EXISTS idx_media_instance
    ON media_metadata(instance_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_media_type_stats
    ON media_metadata(media_type, download_status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_media_failed
    ON media_metadata(instance_id, created_at DESC)
    WHERE download_status = 'failed' AND download_attempts >= max_retries;

-- Fallback indexes (from migration 000006)
CREATE INDEX IF NOT EXISTS idx_media_metadata_storage_type
    ON media_metadata(storage_type, created_at)
    WHERE storage_type = 'local';

CREATE INDEX IF NOT EXISTS idx_media_metadata_fallback
    ON media_metadata(fallback_attempted, storage_type)
    WHERE fallback_attempted = true;

-- Column comments
COMMENT ON COLUMN media_metadata.storage_type IS
  'Storage location: s3 (S3 bucket), local (local filesystem), null (no storage - media_url is NULL)';

COMMENT ON COLUMN media_metadata.fallback_attempted IS
  'Whether fallback to local storage was attempted after S3 failure';

COMMENT ON COLUMN media_metadata.fallback_error IS
  'Error message if all storage methods failed';

-- Instance Event Sequence: Atomic sequence generator per instance
-- Guarantees ordered event processing (critical requirement)
CREATE TABLE IF NOT EXISTS instance_event_sequence (
    instance_id UUID PRIMARY KEY,
    current_sequence BIGINT NOT NULL DEFAULT 0,
    last_event_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Audit: track sequence generation rate
    total_events BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for monitoring: find most active instances
CREATE INDEX IF NOT EXISTS idx_sequence_activity
    ON instance_event_sequence(last_event_at DESC NULLS LAST);

-- ============================================================================
-- FUNCTIONS AND TRIGGERS
-- ============================================================================

-- Function: Update updated_at timestamp (original)
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function: Update event_outbox updated_at
CREATE OR REPLACE FUNCTION update_event_outbox_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function: Update event_dlq updated_at
CREATE OR REPLACE FUNCTION update_event_dlq_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function: Update media_metadata updated_at
CREATE OR REPLACE FUNCTION update_media_metadata_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function: Get next sequence number atomically
CREATE OR REPLACE FUNCTION get_next_event_sequence(p_instance_id UUID)
RETURNS BIGINT AS $$
DECLARE
    v_sequence BIGINT;
BEGIN
    INSERT INTO instance_event_sequence (
        instance_id,
        current_sequence,
        last_event_at,
        total_events,
        updated_at,
        created_at
    )
    VALUES (
        p_instance_id,
        1,
        NOW(),
        1,
        NOW(),
        NOW()
    )
    ON CONFLICT (instance_id)
    DO UPDATE SET
        current_sequence = instance_event_sequence.current_sequence + 1,
        last_event_at = NOW(),
        total_events = instance_event_sequence.total_events + 1,
        updated_at = NOW()
    RETURNING current_sequence INTO v_sequence;

    RETURN v_sequence;
END;
$$ LANGUAGE plpgsql;

-- Function: Reset sequence (admin use only - use with caution)
CREATE OR REPLACE FUNCTION reset_event_sequence(p_instance_id UUID)
RETURNS VOID AS $$
BEGIN
    UPDATE instance_event_sequence
    SET
        current_sequence = 0,
        last_event_at = NULL,
        total_events = 0,
        updated_at = NOW()
    WHERE instance_id = p_instance_id;
END;
$$ LANGUAGE plpgsql;

-- Function: Check for sequence gaps (monitoring/alerting)
CREATE OR REPLACE FUNCTION check_sequence_gaps(p_instance_id UUID)
RETURNS TABLE(missing_sequence BIGINT) AS $$
BEGIN
    RETURN QUERY
    WITH expected_sequences AS (
        SELECT generate_series(
            1::BIGINT,
            (SELECT current_sequence FROM instance_event_sequence WHERE instance_id = p_instance_id)
        ) AS seq
    ),
    existing_sequences AS (
        SELECT sequence_number
        FROM event_outbox
        WHERE instance_id = p_instance_id
    )
    SELECT e.seq
    FROM expected_sequences e
    LEFT JOIN existing_sequences x ON e.seq = x.sequence_number
    WHERE x.sequence_number IS NULL
    ORDER BY e.seq;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Original triggers
CREATE TRIGGER set_timestamp_instances
BEFORE UPDATE ON instances
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_webhook_configs
BEFORE UPDATE ON webhook_configs
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_webhook_outbox
BEFORE UPDATE ON webhook_outbox
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

-- Event system triggers
CREATE TRIGGER trg_event_outbox_updated_at
    BEFORE UPDATE ON event_outbox
    FOR EACH ROW
    EXECUTE FUNCTION update_event_outbox_updated_at();

CREATE TRIGGER trg_event_dlq_updated_at
    BEFORE UPDATE ON event_dlq
    FOR EACH ROW
    EXECUTE FUNCTION update_event_dlq_updated_at();

CREATE TRIGGER trg_media_metadata_updated_at
    BEFORE UPDATE ON media_metadata
    FOR EACH ROW
    EXECUTE FUNCTION update_media_metadata_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop triggers
DROP TRIGGER IF EXISTS trg_media_metadata_updated_at ON media_metadata;
DROP TRIGGER IF EXISTS trg_event_dlq_updated_at ON event_dlq;
DROP TRIGGER IF EXISTS trg_event_outbox_updated_at ON event_outbox;
DROP TRIGGER IF EXISTS set_timestamp_webhook_outbox ON webhook_outbox;
DROP TRIGGER IF EXISTS set_timestamp_webhook_configs ON webhook_configs;
DROP TRIGGER IF EXISTS set_timestamp_instances ON instances;

-- Drop functions
DROP FUNCTION IF EXISTS check_sequence_gaps(UUID);
DROP FUNCTION IF EXISTS reset_event_sequence(UUID);
DROP FUNCTION IF EXISTS get_next_event_sequence(UUID);
DROP FUNCTION IF EXISTS update_media_metadata_updated_at();
DROP FUNCTION IF EXISTS update_event_dlq_updated_at();
DROP FUNCTION IF EXISTS update_event_outbox_updated_at();
DROP FUNCTION IF EXISTS trigger_set_timestamp();

-- Drop indexes (instance_event_sequence)
DROP INDEX IF EXISTS idx_sequence_activity;

-- Drop indexes (media_metadata)
DROP INDEX IF EXISTS idx_media_metadata_fallback;
DROP INDEX IF EXISTS idx_media_metadata_storage_type;
DROP INDEX IF EXISTS idx_media_failed;
DROP INDEX IF EXISTS idx_media_type_stats;
DROP INDEX IF EXISTS idx_media_instance;
DROP INDEX IF EXISTS idx_media_by_event;
DROP INDEX IF EXISTS idx_media_pending;

-- Drop indexes (event_dlq)
DROP INDEX IF EXISTS idx_dlq_failure_reason;
DROP INDEX IF EXISTS idx_dlq_event_type;
DROP INDEX IF EXISTS idx_dlq_reprocess_pending;
DROP INDEX IF EXISTS idx_dlq_instance;

-- Drop indexes (event_outbox)
DROP INDEX IF EXISTS idx_outbox_failed;
DROP INDEX IF EXISTS idx_outbox_event_type;
DROP INDEX IF EXISTS idx_outbox_instance_recent;
DROP INDEX IF EXISTS idx_outbox_media_pending;
DROP INDEX IF EXISTS idx_outbox_instance_pending;

-- Drop indexes (webhook_outbox)
DROP INDEX IF EXISTS idx_webhook_outbox_instance_created;
DROP INDEX IF EXISTS idx_webhook_outbox_status_next_attempt;

-- Drop indexes (instances)
DROP INDEX IF EXISTS idx_instances_instance_token;
DROP INDEX IF EXISTS idx_instances_client_token;

-- Drop tables
DROP TABLE IF EXISTS instance_event_sequence;
DROP TABLE IF EXISTS media_metadata;
DROP TABLE IF EXISTS event_dlq;
DROP TABLE IF EXISTS event_outbox;
DROP TABLE IF EXISTS instance_events_log;
DROP TABLE IF EXISTS webhook_dlq;
DROP TABLE IF EXISTS webhook_outbox;
DROP TABLE IF EXISTS webhook_configs;
DROP TABLE IF EXISTS instances;

-- Drop extensions
DROP EXTENSION IF EXISTS "uuid-ossp";

-- +goose StatementEnd
