-- +goose Up

-- proxy_providers: External proxy sources (Webshare, BrightData, custom)
CREATE TABLE proxy_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    provider_type VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    priority INT NOT NULL DEFAULT 100,
    api_key TEXT,
    api_endpoint TEXT,
    max_proxies INT NOT NULL DEFAULT 0,
    max_instances_per_proxy INT NOT NULL DEFAULT 1,
    country_codes TEXT[] NOT NULL DEFAULT '{}',
    rate_limit_rpm INT NOT NULL DEFAULT 60,
    last_sync_at TIMESTAMPTZ,
    sync_error TEXT,
    proxy_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- proxy_pool: All cached proxies from all providers
CREATE TABLE proxy_pool (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL REFERENCES proxy_providers(id) ON DELETE CASCADE,
    external_id TEXT,
    proxy_url TEXT NOT NULL,
    country_code VARCHAR(5),
    city TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'available',
    health_status VARCHAR(20) NOT NULL DEFAULT 'unknown',
    health_failures INT NOT NULL DEFAULT 0,
    last_health_check TIMESTAMPTZ,
    assigned_count INT NOT NULL DEFAULT 0,
    max_assignments INT NOT NULL DEFAULT 1,
    valid BOOLEAN NOT NULL DEFAULT TRUE,
    last_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT proxy_pool_status_check CHECK (status IN ('available', 'assigned', 'unhealthy', 'retired')),
    CONSTRAINT proxy_pool_health_check CHECK (health_status IN ('healthy', 'unhealthy', 'unknown'))
);

CREATE UNIQUE INDEX idx_proxy_pool_url ON proxy_pool(proxy_url);
CREATE INDEX idx_proxy_pool_available ON proxy_pool(status, assigned_count) WHERE status = 'available';
CREATE INDEX idx_proxy_pool_provider ON proxy_pool(provider_id);

-- proxy_assignments: Maps pool proxies to instances
CREATE TABLE proxy_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pool_proxy_id UUID NOT NULL REFERENCES proxy_pool(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    group_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    released_at TIMESTAMPTZ,
    assigned_by VARCHAR(50) NOT NULL DEFAULT 'auto',
    release_reason TEXT,
    CONSTRAINT proxy_assignments_status_check CHECK (status IN ('active', 'pending_swap', 'inactive'))
);

CREATE INDEX idx_proxy_assignments_instance ON proxy_assignments(instance_id) WHERE status = 'active';
CREATE INDEX idx_proxy_assignments_pool_proxy ON proxy_assignments(pool_proxy_id) WHERE status = 'active';

-- proxy_groups: For N:1 proxy sharing
CREATE TABLE proxy_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    provider_id UUID REFERENCES proxy_providers(id) ON DELETE SET NULL,
    pool_proxy_id UUID REFERENCES proxy_pool(id) ON DELETE SET NULL,
    max_instances INT NOT NULL DEFAULT 1,
    country_code VARCHAR(5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS proxy_groups;
DROP TABLE IF EXISTS proxy_assignments;
DROP TABLE IF EXISTS proxy_pool;
DROP TABLE IF EXISTS proxy_providers;
