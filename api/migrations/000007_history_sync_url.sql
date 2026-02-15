-- +goose Up
ALTER TABLE webhook_configs ADD COLUMN IF NOT EXISTS history_sync_url TEXT;

-- +goose Down
ALTER TABLE webhook_configs DROP COLUMN IF EXISTS history_sync_url;
