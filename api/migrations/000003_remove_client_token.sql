-- +goose Up
-- +goose StatementBegin

-- Migration: Remove per-instance client_token in favor of global CLIENT_AUTH_TOKEN
--
-- Context:
-- Previously, each instance had its own unique client_token generated at creation.
-- This migration removes that column as we now use a single global token from env
-- (CLIENT_AUTH_TOKEN), similar to the PARTNER_AUTH_TOKEN approach.
--
-- Breaking Change: YES
-- - The `client_token` field will no longer be returned in API responses
-- - Authentication now requires the global CLIENT_AUTH_TOKEN header value
-- - Existing client applications must update to use the new global token
--
-- Impact:
-- - Removes `client_token` column from instances table
-- - Removes unique index `idx_instances_client_token`
-- - instance_token remains intact and continues to work as before
-- - No data loss: only removes per-instance token column

-- Remove unique index first (prevents FK constraint issues)
DROP INDEX IF EXISTS idx_instances_client_token;

-- Remove column containing per-instance tokens
-- SAFE: Data is no longer needed as global token is now used
ALTER TABLE instances DROP COLUMN IF EXISTS client_token;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Rollback: Restore client_token column with placeholder values
--
-- WARNING: This rollback will add the column back with a default placeholder value.
-- Original per-instance tokens are permanently lost after migration runs.
-- If you need to rollback, you'll need to regenerate tokens for existing instances.

-- Restore column with default placeholder (old tokens are lost)
ALTER TABLE instances ADD COLUMN client_token TEXT NOT NULL DEFAULT 'deprecated-rollback-placeholder';

-- Restore unique index
-- NOTE: This will fail if duplicate placeholder values exist
-- In that case, manually update tokens before creating the index
CREATE UNIQUE INDEX idx_instances_client_token ON instances(client_token);

-- +goose StatementEnd
