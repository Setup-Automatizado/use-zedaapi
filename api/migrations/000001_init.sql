CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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

CREATE TABLE IF NOT EXISTS webhook_dlq (
    id UUID PRIMARY KEY,
    instance_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    failure_reason TEXT,
    failed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS instance_events_log (
    id BIGSERIAL PRIMARY KEY,
    instance_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

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
