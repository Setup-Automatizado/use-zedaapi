-- +goose Up
CREATE TABLE IF NOT EXISTS worker_sessions (
    worker_id TEXT PRIMARY KEY,
    hostname TEXT NOT NULL,
    app_env TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    capacity INTEGER NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS worker_sessions_app_env_last_seen_idx
    ON worker_sessions (app_env, last_seen DESC);

ALTER TABLE instances
    ADD COLUMN IF NOT EXISTS desired_worker_id TEXT;

-- +goose Down
ALTER TABLE instances
    DROP COLUMN IF EXISTS desired_worker_id;

DROP TABLE IF EXISTS worker_sessions;
