-- +goose Up

-- Per-instance proxy configuration columns
ALTER TABLE instances
    ADD COLUMN IF NOT EXISTS proxy_url TEXT,
    ADD COLUMN IF NOT EXISTS proxy_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS proxy_no_websocket BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS proxy_only_login BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS proxy_no_media BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS proxy_health_status VARCHAR(20) NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS proxy_last_health_check TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS proxy_health_failures INT NOT NULL DEFAULT 0;

ALTER TABLE instances
    ADD CONSTRAINT instances_proxy_health_status_check
    CHECK (proxy_health_status IN ('healthy', 'unhealthy', 'unknown'));

CREATE INDEX IF NOT EXISTS idx_instances_proxy_enabled
    ON instances (proxy_enabled) WHERE proxy_enabled = TRUE;

CREATE INDEX IF NOT EXISTS idx_instances_proxy_health
    ON instances (proxy_health_status) WHERE proxy_enabled = TRUE;

-- Proxy health check log for audit trail and debugging
CREATE TABLE IF NOT EXISTS proxy_health_log (
    id BIGSERIAL PRIMARY KEY,
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    proxy_url TEXT NOT NULL,
    status VARCHAR(20) NOT NULL,
    latency_ms INT,
    error_message TEXT,
    checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT proxy_health_log_status_check
        CHECK (status IN ('healthy', 'unhealthy', 'timeout', 'unreachable'))
);

CREATE INDEX IF NOT EXISTS idx_proxy_health_log_instance
    ON proxy_health_log (instance_id, checked_at DESC);

-- Index for efficient cleanup queries on health logs
CREATE INDEX IF NOT EXISTS idx_proxy_health_log_cleanup
    ON proxy_health_log (checked_at);

-- +goose Down
DROP INDEX IF EXISTS idx_proxy_health_log_cleanup;
DROP INDEX IF EXISTS idx_proxy_health_log_instance;
DROP TABLE IF EXISTS proxy_health_log;

DROP INDEX IF EXISTS idx_instances_proxy_health;
DROP INDEX IF EXISTS idx_instances_proxy_enabled;

ALTER TABLE instances
    DROP CONSTRAINT IF EXISTS instances_proxy_health_status_check;

ALTER TABLE instances
    DROP COLUMN IF EXISTS proxy_health_failures,
    DROP COLUMN IF EXISTS proxy_last_health_check,
    DROP COLUMN IF EXISTS proxy_health_status,
    DROP COLUMN IF EXISTS proxy_no_media,
    DROP COLUMN IF EXISTS proxy_only_login,
    DROP COLUMN IF EXISTS proxy_no_websocket,
    DROP COLUMN IF EXISTS proxy_enabled,
    DROP COLUMN IF EXISTS proxy_url;
