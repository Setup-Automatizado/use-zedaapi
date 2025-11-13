-- +goose Up
-- +goose StatementBegin

ALTER TABLE instances
    ADD COLUMN IF NOT EXISTS connected BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS connection_status TEXT NOT NULL DEFAULT 'disconnected',
    ADD COLUMN IF NOT EXISTS last_connected_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS worker_id TEXT;

CREATE INDEX IF NOT EXISTS idx_instances_connected ON instances(connected);
CREATE INDEX IF NOT EXISTS idx_instances_connection_status ON instances(connection_status);

UPDATE instances
SET connection_status = COALESCE(connection_status, 'disconnected')
WHERE connection_status IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_instances_connection_status;
DROP INDEX IF EXISTS idx_instances_connected;

ALTER TABLE instances
    DROP COLUMN IF EXISTS worker_id,
    DROP COLUMN IF EXISTS last_connected_at,
    DROP COLUMN IF EXISTS connection_status,
    DROP COLUMN IF EXISTS connected;

-- +goose StatementEnd
