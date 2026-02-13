package proxy

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrProxyCapacityFull is returned when a proxy has reached its max_assignments limit.
var ErrProxyCapacityFull = errors.New("proxy capacity full")

// PoolRepository handles all database operations for proxy providers, pool, assignments, and groups.
type PoolRepository struct {
	pool *pgxpool.Pool
}

// NewPoolRepository creates a new PoolRepository.
func NewPoolRepository(pool *pgxpool.Pool) *PoolRepository {
	return &PoolRepository{pool: pool}
}

// ---------------------------------------------------------------------------
// Provider CRUD
// ---------------------------------------------------------------------------

// CreateProvider inserts a new proxy provider record.
func (r *PoolRepository) CreateProvider(ctx context.Context, req CreateProviderRequest) (*ProviderRecord, error) {
	query := `INSERT INTO proxy_providers
		(name, provider_type, enabled, priority, api_key, api_endpoint,
		 max_proxies, max_instances_per_proxy, country_codes, rate_limit_rpm)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, name, provider_type, enabled, priority, api_key, api_endpoint,
		          max_proxies, max_instances_per_proxy, country_codes, rate_limit_rpm,
		          last_sync_at, sync_error, proxy_count, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query,
		req.Name, req.ProviderType, req.Enabled, req.Priority,
		req.APIKey, req.APIEndpoint, req.MaxProxies, req.MaxInstancesPerProxy,
		req.CountryCodes, req.RateLimitRPM,
	)

	var rec ProviderRecord
	if err := row.Scan(
		&rec.ID, &rec.Name, &rec.ProviderType, &rec.Enabled, &rec.Priority,
		&rec.APIKey, &rec.APIEndpoint, &rec.MaxProxies, &rec.MaxInstancesPerProxy,
		&rec.CountryCodes, &rec.RateLimitRPM, &rec.LastSyncAt, &rec.SyncError,
		&rec.ProxyCount, &rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}
	return &rec, nil
}

// GetProvider retrieves a proxy provider by ID.
func (r *PoolRepository) GetProvider(ctx context.Context, id uuid.UUID) (*ProviderRecord, error) {
	query := `SELECT id, name, provider_type, enabled, priority, api_key, api_endpoint,
	          max_proxies, max_instances_per_proxy, country_codes, rate_limit_rpm,
	          last_sync_at, sync_error, proxy_count, created_at, updated_at
	          FROM proxy_providers WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)

	var rec ProviderRecord
	if err := row.Scan(
		&rec.ID, &rec.Name, &rec.ProviderType, &rec.Enabled, &rec.Priority,
		&rec.APIKey, &rec.APIEndpoint, &rec.MaxProxies, &rec.MaxInstancesPerProxy,
		&rec.CountryCodes, &rec.RateLimitRPM, &rec.LastSyncAt, &rec.SyncError,
		&rec.ProxyCount, &rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("provider not found: %w", err)
		}
		return nil, fmt.Errorf("get provider: %w", err)
	}
	return &rec, nil
}

// ListProviders returns all proxy providers ordered by priority ascending, then name ascending.
func (r *PoolRepository) ListProviders(ctx context.Context) ([]ProviderRecord, error) {
	query := `SELECT id, name, provider_type, enabled, priority, api_key, api_endpoint,
	          max_proxies, max_instances_per_proxy, country_codes, rate_limit_rpm,
	          last_sync_at, sync_error, proxy_count, created_at, updated_at
	          FROM proxy_providers ORDER BY priority ASC, name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list providers: %w", err)
	}
	defer rows.Close()

	var result []ProviderRecord
	for rows.Next() {
		var rec ProviderRecord
		if err := rows.Scan(
			&rec.ID, &rec.Name, &rec.ProviderType, &rec.Enabled, &rec.Priority,
			&rec.APIKey, &rec.APIEndpoint, &rec.MaxProxies, &rec.MaxInstancesPerProxy,
			&rec.CountryCodes, &rec.RateLimitRPM, &rec.LastSyncAt, &rec.SyncError,
			&rec.ProxyCount, &rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan provider: %w", err)
		}
		result = append(result, rec)
	}
	return result, nil
}

// UpdateProvider applies a partial update to a proxy provider. Only non-nil fields are updated.
func (r *PoolRepository) UpdateProvider(ctx context.Context, id uuid.UUID, req UpdateProviderRequest) (*ProviderRecord, error) {
	setClauses := []string{}
	args := []any{id}
	argIdx := 2

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Enabled != nil {
		setClauses = append(setClauses, fmt.Sprintf("enabled = $%d", argIdx))
		args = append(args, *req.Enabled)
		argIdx++
	}
	if req.Priority != nil {
		setClauses = append(setClauses, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, *req.Priority)
		argIdx++
	}
	if req.APIKey != nil {
		setClauses = append(setClauses, fmt.Sprintf("api_key = $%d", argIdx))
		args = append(args, *req.APIKey)
		argIdx++
	}
	if req.APIEndpoint != nil {
		setClauses = append(setClauses, fmt.Sprintf("api_endpoint = $%d", argIdx))
		args = append(args, *req.APIEndpoint)
		argIdx++
	}
	if req.MaxProxies != nil {
		setClauses = append(setClauses, fmt.Sprintf("max_proxies = $%d", argIdx))
		args = append(args, *req.MaxProxies)
		argIdx++
	}
	if req.MaxInstancesPerProxy != nil {
		setClauses = append(setClauses, fmt.Sprintf("max_instances_per_proxy = $%d", argIdx))
		args = append(args, *req.MaxInstancesPerProxy)
		argIdx++
	}
	if req.CountryCodes != nil {
		setClauses = append(setClauses, fmt.Sprintf("country_codes = $%d", argIdx))
		args = append(args, req.CountryCodes)
		argIdx++
	}
	if req.RateLimitRPM != nil {
		setClauses = append(setClauses, fmt.Sprintf("rate_limit_rpm = $%d", argIdx))
		args = append(args, *req.RateLimitRPM)
		argIdx++
	}

	if len(setClauses) == 0 {
		return r.GetProvider(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query := fmt.Sprintf(`UPDATE proxy_providers SET %s WHERE id = $1
		RETURNING id, name, provider_type, enabled, priority, api_key, api_endpoint,
		          max_proxies, max_instances_per_proxy, country_codes, rate_limit_rpm,
		          last_sync_at, sync_error, proxy_count, created_at, updated_at`,
		strings.Join(setClauses, ", "))

	row := r.pool.QueryRow(ctx, query, args...)

	var rec ProviderRecord
	if err := row.Scan(
		&rec.ID, &rec.Name, &rec.ProviderType, &rec.Enabled, &rec.Priority,
		&rec.APIKey, &rec.APIEndpoint, &rec.MaxProxies, &rec.MaxInstancesPerProxy,
		&rec.CountryCodes, &rec.RateLimitRPM, &rec.LastSyncAt, &rec.SyncError,
		&rec.ProxyCount, &rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("provider not found: %w", err)
		}
		return nil, fmt.Errorf("update provider: %w", err)
	}
	return &rec, nil
}

// DeleteProvider removes a proxy provider by ID. Cascades to pool and assignments.
func (r *PoolRepository) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM proxy_providers WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete provider: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("provider not found: %w", pgx.ErrNoRows)
	}
	return nil
}

// UpdateProviderSyncStatus updates last_sync_at, sync_error, and proxy_count after a sync.
func (r *PoolRepository) UpdateProviderSyncStatus(ctx context.Context, id uuid.UUID, syncErr *string, proxyCount int) error {
	query := `UPDATE proxy_providers SET
		last_sync_at = NOW(),
		sync_error = $2,
		proxy_count = $3,
		updated_at = NOW()
		WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id, syncErr, proxyCount)
	if err != nil {
		return fmt.Errorf("update provider sync status: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("provider not found: %w", pgx.ErrNoRows)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Pool operations
// ---------------------------------------------------------------------------

// UpsertPoolProxy inserts or updates a proxy in the pool. On conflict by proxy_url,
// it updates external_id, country_code, city, valid, and last_verified_at, but NEVER
// overwrites status or assigned_count for already-assigned proxies.
func (r *PoolRepository) UpsertPoolProxy(ctx context.Context, providerID uuid.UUID, entry ProxyEntry, maxAssignments int) (*PoolProxyRecord, error) {
	query := `INSERT INTO proxy_pool
		(provider_id, external_id, proxy_url, country_code, city, max_assignments, valid, last_verified_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (proxy_url) DO UPDATE SET
			external_id = EXCLUDED.external_id,
			country_code = EXCLUDED.country_code,
			city = EXCLUDED.city,
			valid = EXCLUDED.valid,
			last_verified_at = EXCLUDED.last_verified_at,
			max_assignments = EXCLUDED.max_assignments,
			updated_at = NOW()
		RETURNING id, provider_id, external_id, proxy_url, country_code, city, status,
		          health_status, health_failures, last_health_check, assigned_count,
		          max_assignments, valid, last_verified_at, created_at, updated_at`

	var lastVerified *time.Time
	if !entry.LastVerified.IsZero() {
		lastVerified = &entry.LastVerified
	}

	var countryCode *string
	if entry.CountryCode != "" {
		countryCode = &entry.CountryCode
	}

	var city *string
	if entry.City != "" {
		city = &entry.City
	}

	var externalID *string
	if entry.ExternalID != "" {
		externalID = &entry.ExternalID
	}

	row := r.pool.QueryRow(ctx, query,
		providerID, externalID, entry.ProxyURL, countryCode, city,
		maxAssignments, entry.Valid, lastVerified,
	)

	var rec PoolProxyRecord
	if err := row.Scan(
		&rec.ID, &rec.ProviderID, &rec.ExternalID, &rec.ProxyURL,
		&rec.CountryCode, &rec.City, &rec.Status, &rec.HealthStatus,
		&rec.HealthFailures, &rec.LastHealthCheck, &rec.AssignedCount,
		&rec.MaxAssignments, &rec.Valid, &rec.LastVerifiedAt,
		&rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("upsert pool proxy: %w", err)
	}
	return &rec, nil
}

// FindAvailableProxy selects a single available proxy using FOR UPDATE SKIP LOCKED
// for concurrent safety. Filters by country codes and provider ID if provided.
func (r *PoolRepository) FindAvailableProxy(ctx context.Context, countryCodes []string, providerID *uuid.UUID) (*PoolProxyRecord, error) {
	query := `SELECT id, provider_id, external_id, proxy_url, country_code, city, status,
	          health_status, health_failures, last_health_check, assigned_count,
	          max_assignments, valid, last_verified_at, created_at, updated_at
	          FROM proxy_pool
	          WHERE status = 'available' AND assigned_count < max_assignments AND valid = TRUE
	            AND ($1::text[] IS NULL OR country_code = ANY($1))
	            AND ($2::uuid IS NULL OR provider_id = $2)
	          ORDER BY assigned_count ASC, health_failures ASC, RANDOM()
	          LIMIT 1
	          FOR UPDATE SKIP LOCKED`

	var codes []string
	if len(countryCodes) > 0 {
		codes = countryCodes
	}

	row := r.pool.QueryRow(ctx, query, codes, providerID)

	var rec PoolProxyRecord
	if err := row.Scan(
		&rec.ID, &rec.ProviderID, &rec.ExternalID, &rec.ProxyURL,
		&rec.CountryCode, &rec.City, &rec.Status, &rec.HealthStatus,
		&rec.HealthFailures, &rec.LastHealthCheck, &rec.AssignedCount,
		&rec.MaxAssignments, &rec.Valid, &rec.LastVerifiedAt,
		&rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no available proxy found: %w", err)
		}
		return nil, fmt.Errorf("find available proxy: %w", err)
	}
	return &rec, nil
}

// IncrementAssignedCount atomically increments the assigned_count for a pool proxy
// ONLY if assigned_count < max_assignments. Updates status to 'assigned' when full.
// Returns ErrProxyCapacityFull if the proxy has already reached its limit (concurrent-safe).
func (r *PoolRepository) IncrementAssignedCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE proxy_pool SET
		assigned_count = assigned_count + 1,
		status = CASE WHEN assigned_count + 1 >= max_assignments THEN 'assigned' ELSE status END,
		updated_at = NOW()
		WHERE id = $1 AND assigned_count < max_assignments`
	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("increment assigned count: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrProxyCapacityFull
	}
	return nil
}

// DecrementAssignedCount decrements the assigned_count for a pool proxy (minimum 0)
// and updates status to 'available' if it was 'assigned' and now below max_assignments.
func (r *PoolRepository) DecrementAssignedCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE proxy_pool SET
		assigned_count = GREATEST(assigned_count - 1, 0),
		status = CASE
			WHEN status = 'assigned' AND GREATEST(assigned_count - 1, 0) < max_assignments THEN 'available'
			ELSE status
		END,
		updated_at = NOW()
		WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("decrement assigned count: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("pool proxy not found: %w", pgx.ErrNoRows)
	}
	return nil
}

// UpdatePoolProxyStatus updates the status of a pool proxy.
func (r *PoolRepository) UpdatePoolProxyStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE proxy_pool SET status = $2, updated_at = NOW() WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("update pool proxy status: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("pool proxy not found: %w", pgx.ErrNoRows)
	}
	return nil
}

// UpdatePoolProxyHealth updates the health_status and health_failures for a pool proxy.
func (r *PoolRepository) UpdatePoolProxyHealth(ctx context.Context, id uuid.UUID, healthStatus string, failures int) error {
	query := `UPDATE proxy_pool SET
		health_status = $2,
		health_failures = $3,
		last_health_check = NOW(),
		updated_at = NOW()
		WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id, healthStatus, failures)
	if err != nil {
		return fmt.Errorf("update pool proxy health: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("pool proxy not found: %w", pgx.ErrNoRows)
	}
	return nil
}

// RetirePoolProxies marks proxies as 'retired' for a provider where the external_id
// is NOT in the provided active list and the proxy is not currently assigned.
func (r *PoolRepository) RetirePoolProxies(ctx context.Context, providerID uuid.UUID, activeExternalIDs []string) (int64, error) {
	query := `UPDATE proxy_pool SET
		status = 'retired',
		updated_at = NOW()
		WHERE provider_id = $1
		  AND external_id IS NOT NULL
		  AND external_id != ALL($2::text[])
		  AND status != 'assigned'`
	res, err := r.pool.Exec(ctx, query, providerID, activeExternalIDs)
	if err != nil {
		return 0, fmt.Errorf("retire pool proxies: %w", err)
	}
	return res.RowsAffected(), nil
}

// ListPoolProxies returns a paginated list of pool proxies with optional filters.
// Returns the records and total count.
func (r *PoolRepository) ListPoolProxies(ctx context.Context, providerID *uuid.UUID, status *string, limit, offset int) ([]PoolProxyRecord, int, error) {
	countQuery := `SELECT COUNT(*) FROM proxy_pool WHERE
		($1::uuid IS NULL OR provider_id = $1)
		AND ($2::text IS NULL OR status = $2)`

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, providerID, status).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count pool proxies: %w", err)
	}

	query := `SELECT id, provider_id, external_id, proxy_url, country_code, city, status,
	          health_status, health_failures, last_health_check, assigned_count,
	          max_assignments, valid, last_verified_at, created_at, updated_at
	          FROM proxy_pool
	          WHERE ($1::uuid IS NULL OR provider_id = $1)
	            AND ($2::text IS NULL OR status = $2)
	          ORDER BY created_at DESC
	          LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, providerID, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list pool proxies: %w", err)
	}
	defer rows.Close()

	var result []PoolProxyRecord
	for rows.Next() {
		var rec PoolProxyRecord
		if err := rows.Scan(
			&rec.ID, &rec.ProviderID, &rec.ExternalID, &rec.ProxyURL,
			&rec.CountryCode, &rec.City, &rec.Status, &rec.HealthStatus,
			&rec.HealthFailures, &rec.LastHealthCheck, &rec.AssignedCount,
			&rec.MaxAssignments, &rec.Valid, &rec.LastVerifiedAt,
			&rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan pool proxy: %w", err)
		}
		result = append(result, rec)
	}
	return result, total, nil
}

// GetPoolProxyByURL retrieves a pool proxy by its proxy URL.
func (r *PoolRepository) GetPoolProxyByURL(ctx context.Context, proxyURL string) (*PoolProxyRecord, error) {
	query := `SELECT id, provider_id, external_id, proxy_url, country_code, city, status,
	          health_status, health_failures, last_health_check, assigned_count,
	          max_assignments, valid, last_verified_at, created_at, updated_at
	          FROM proxy_pool WHERE proxy_url = $1`

	row := r.pool.QueryRow(ctx, query, proxyURL)

	var rec PoolProxyRecord
	if err := row.Scan(
		&rec.ID, &rec.ProviderID, &rec.ExternalID, &rec.ProxyURL,
		&rec.CountryCode, &rec.City, &rec.Status, &rec.HealthStatus,
		&rec.HealthFailures, &rec.LastHealthCheck, &rec.AssignedCount,
		&rec.MaxAssignments, &rec.Valid, &rec.LastVerifiedAt,
		&rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("pool proxy not found: %w", err)
		}
		return nil, fmt.Errorf("get pool proxy by url: %w", err)
	}
	return &rec, nil
}

// GetPoolProxyByID retrieves a pool proxy by its UUID.
func (r *PoolRepository) GetPoolProxyByID(ctx context.Context, id uuid.UUID) (*PoolProxyRecord, error) {
	query := `SELECT id, provider_id, external_id, proxy_url, country_code, city, status,
	          health_status, health_failures, last_health_check, assigned_count,
	          max_assignments, valid, last_verified_at, created_at, updated_at
	          FROM proxy_pool WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)
	var rec PoolProxyRecord
	if err := row.Scan(
		&rec.ID, &rec.ProviderID, &rec.ExternalID, &rec.ProxyURL,
		&rec.CountryCode, &rec.City, &rec.Status, &rec.HealthStatus,
		&rec.HealthFailures, &rec.LastHealthCheck, &rec.AssignedCount,
		&rec.MaxAssignments, &rec.Valid, &rec.LastVerifiedAt,
		&rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("pool proxy not found: %w", err)
		}
		return nil, fmt.Errorf("get pool proxy by id: %w", err)
	}
	return &rec, nil
}

// ---------------------------------------------------------------------------
// Assignment operations
// ---------------------------------------------------------------------------

// CreateAssignment inserts a new proxy assignment record.
func (r *PoolRepository) CreateAssignment(ctx context.Context, poolProxyID, instanceID uuid.UUID, groupID *uuid.UUID, assignedBy string) (*AssignmentRecord, error) {
	query := `INSERT INTO proxy_assignments (pool_proxy_id, instance_id, group_id, assigned_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, pool_proxy_id, instance_id, group_id, status, assigned_at,
		          released_at, assigned_by, release_reason`

	row := r.pool.QueryRow(ctx, query, poolProxyID, instanceID, groupID, assignedBy)

	var rec AssignmentRecord
	if err := row.Scan(
		&rec.ID, &rec.PoolProxyID, &rec.InstanceID, &rec.GroupID,
		&rec.Status, &rec.AssignedAt, &rec.ReleasedAt,
		&rec.AssignedBy, &rec.ReleaseReason,
	); err != nil {
		return nil, fmt.Errorf("create assignment: %w", err)
	}
	return &rec, nil
}

// DeactivateAssignment marks the active assignment for an instance as inactive.
func (r *PoolRepository) DeactivateAssignment(ctx context.Context, instanceID uuid.UUID, reason string) error {
	query := `UPDATE proxy_assignments SET
		status = 'inactive',
		released_at = NOW(),
		release_reason = $2
		WHERE instance_id = $1 AND status = 'active'`
	res, err := r.pool.Exec(ctx, query, instanceID, reason)
	if err != nil {
		return fmt.Errorf("deactivate assignment: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("no active assignment found: %w", pgx.ErrNoRows)
	}
	return nil
}

// GetActiveAssignment retrieves the active assignment for an instance.
func (r *PoolRepository) GetActiveAssignment(ctx context.Context, instanceID uuid.UUID) (*AssignmentRecord, error) {
	query := `SELECT a.id, a.pool_proxy_id, a.instance_id, a.group_id, a.status, a.assigned_at,
	          a.released_at, a.assigned_by, a.release_reason, COALESCE(pp.proxy_url, '')
	          FROM proxy_assignments a
	          LEFT JOIN proxy_pool pp ON pp.id = a.pool_proxy_id
	          WHERE a.instance_id = $1 AND a.status = 'active'`

	row := r.pool.QueryRow(ctx, query, instanceID)

	var rec AssignmentRecord
	if err := row.Scan(
		&rec.ID, &rec.PoolProxyID, &rec.InstanceID, &rec.GroupID,
		&rec.Status, &rec.AssignedAt, &rec.ReleasedAt,
		&rec.AssignedBy, &rec.ReleaseReason, &rec.ProxyURL,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no active assignment found: %w", err)
		}
		return nil, fmt.Errorf("get active assignment: %w", err)
	}
	return &rec, nil
}

// ListActiveAssignments returns all assignments with status 'active'.
func (r *PoolRepository) ListActiveAssignments(ctx context.Context) ([]AssignmentRecord, error) {
	query := `SELECT id, pool_proxy_id, instance_id, group_id, status, assigned_at,
	          released_at, assigned_by, release_reason
	          FROM proxy_assignments WHERE status = 'active'
	          ORDER BY assigned_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list active assignments: %w", err)
	}
	defer rows.Close()

	var result []AssignmentRecord
	for rows.Next() {
		var rec AssignmentRecord
		if err := rows.Scan(
			&rec.ID, &rec.PoolProxyID, &rec.InstanceID, &rec.GroupID,
			&rec.Status, &rec.AssignedAt, &rec.ReleasedAt,
			&rec.AssignedBy, &rec.ReleaseReason,
		); err != nil {
			return nil, fmt.Errorf("scan active assignment: %w", err)
		}
		result = append(result, rec)
	}
	return result, nil
}

// GetAssignmentsByPoolProxy returns all assignments (any status) for a given pool proxy.
func (r *PoolRepository) GetAssignmentsByPoolProxy(ctx context.Context, poolProxyID uuid.UUID) ([]AssignmentRecord, error) {
	query := `SELECT id, pool_proxy_id, instance_id, group_id, status, assigned_at,
	          released_at, assigned_by, release_reason
	          FROM proxy_assignments WHERE pool_proxy_id = $1
	          ORDER BY assigned_at DESC`

	rows, err := r.pool.Query(ctx, query, poolProxyID)
	if err != nil {
		return nil, fmt.Errorf("get assignments by pool proxy: %w", err)
	}
	defer rows.Close()

	var result []AssignmentRecord
	for rows.Next() {
		var rec AssignmentRecord
		if err := rows.Scan(
			&rec.ID, &rec.PoolProxyID, &rec.InstanceID, &rec.GroupID,
			&rec.Status, &rec.AssignedAt, &rec.ReleasedAt,
			&rec.AssignedBy, &rec.ReleaseReason,
		); err != nil {
			return nil, fmt.Errorf("scan assignment by pool proxy: %w", err)
		}
		result = append(result, rec)
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Group operations
// ---------------------------------------------------------------------------

// CreateGroup inserts a new proxy group record.
func (r *PoolRepository) CreateGroup(ctx context.Context, name string, providerID *uuid.UUID, maxInstances int, countryCode *string) (*GroupRecord, error) {
	query := `INSERT INTO proxy_groups (name, provider_id, max_instances, country_code)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, provider_id, pool_proxy_id, max_instances, country_code, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query, name, providerID, maxInstances, countryCode)

	var rec GroupRecord
	if err := row.Scan(
		&rec.ID, &rec.Name, &rec.ProviderID, &rec.PoolProxyID,
		&rec.MaxInstances, &rec.CountryCode, &rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}
	return &rec, nil
}

// ListGroups returns all proxy groups.
func (r *PoolRepository) ListGroups(ctx context.Context) ([]GroupRecord, error) {
	query := `SELECT id, name, provider_id, pool_proxy_id, max_instances, country_code, created_at, updated_at
	          FROM proxy_groups ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	var result []GroupRecord
	for rows.Next() {
		var rec GroupRecord
		if err := rows.Scan(
			&rec.ID, &rec.Name, &rec.ProviderID, &rec.PoolProxyID,
			&rec.MaxInstances, &rec.CountryCode, &rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		result = append(result, rec)
	}
	return result, nil
}

// DeleteGroup removes a proxy group by ID.
func (r *PoolRepository) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM proxy_groups WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete group: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("group not found: %w", pgx.ErrNoRows)
	}
	return nil
}

// GetGroup retrieves a proxy group by ID.
func (r *PoolRepository) GetGroup(ctx context.Context, id uuid.UUID) (*GroupRecord, error) {
	query := `SELECT id, name, provider_id, pool_proxy_id, max_instances, country_code, created_at, updated_at
	          FROM proxy_groups WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)

	var rec GroupRecord
	if err := row.Scan(
		&rec.ID, &rec.Name, &rec.ProviderID, &rec.PoolProxyID,
		&rec.MaxInstances, &rec.CountryCode, &rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("group not found: %w", err)
		}
		return nil, fmt.Errorf("get group: %w", err)
	}
	return &rec, nil
}

// UpdateGroupProxy sets the pool_proxy_id for a group. Pass nil to clear it.
func (r *PoolRepository) UpdateGroupProxy(ctx context.Context, id uuid.UUID, poolProxyID *uuid.UUID) error {
	query := `UPDATE proxy_groups SET pool_proxy_id = $2, updated_at = NOW() WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id, poolProxyID)
	if err != nil {
		return fmt.Errorf("update group proxy: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("group not found: %w", pgx.ErrNoRows)
	}
	return nil
}

// ListUnassignedInstanceIDs returns IDs of active instances that do not have an active pool assignment.
// If instanceIDs is non-empty, it filters to only those specific instances.
func (r *PoolRepository) ListUnassignedInstanceIDs(ctx context.Context, instanceIDs []uuid.UUID) ([]uuid.UUID, error) {
	var query string
	var args []any

	if len(instanceIDs) > 0 {
		query = `SELECT i.id FROM instances i
			WHERE i.id = ANY($1::uuid[])
			  AND NOT EXISTS (
				SELECT 1 FROM proxy_assignments pa
				WHERE pa.instance_id = i.id AND pa.status = 'active'
			  )
			ORDER BY i.created_at ASC`
		args = append(args, instanceIDs)
	} else {
		query = `SELECT i.id FROM instances i
			WHERE i.subscription_active = TRUE
			  AND NOT EXISTS (
				SELECT 1 FROM proxy_assignments pa
				WHERE pa.instance_id = i.id AND pa.status = 'active'
			  )
			ORDER BY i.created_at ASC`
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list unassigned instance ids: %w", err)
	}
	defer rows.Close()

	var result []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan unassigned instance id: %w", err)
		}
		result = append(result, id)
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Statistics
// ---------------------------------------------------------------------------

// GetPoolStats returns aggregate statistics across the proxy pool and providers.
func (r *PoolRepository) GetPoolStats(ctx context.Context) (*PoolStats, error) {
	// Aggregate counts from proxy_pool
	summaryQuery := `SELECT
		COUNT(*) AS total,
		COUNT(*) FILTER (WHERE status = 'available') AS available,
		COUNT(*) FILTER (WHERE status = 'assigned') AS assigned,
		COUNT(*) FILTER (WHERE status = 'unhealthy') AS unhealthy,
		COUNT(*) FILTER (WHERE status = 'retired') AS retired
		FROM proxy_pool`

	var stats PoolStats
	if err := r.pool.QueryRow(ctx, summaryQuery).Scan(
		&stats.TotalProxies, &stats.AvailableProxies, &stats.AssignedProxies,
		&stats.UnhealthyProxies, &stats.RetiredProxies,
	); err != nil {
		return nil, fmt.Errorf("get pool stats summary: %w", err)
	}

	// Total active assignments
	assignQuery := `SELECT COUNT(*) FROM proxy_assignments WHERE status = 'active'`
	if err := r.pool.QueryRow(ctx, assignQuery).Scan(&stats.TotalAssignments); err != nil {
		return nil, fmt.Errorf("get pool stats assignments: %w", err)
	}

	// Per-provider breakdown
	providerQuery := `SELECT
		pp.id, pp.name,
		COUNT(pool.id) AS total,
		COUNT(pool.id) FILTER (WHERE pool.status = 'available') AS available,
		COUNT(pool.id) FILTER (WHERE pool.status = 'assigned') AS assigned,
		COUNT(pool.id) FILTER (WHERE pool.status = 'unhealthy') AS unhealthy
		FROM proxy_providers pp
		LEFT JOIN proxy_pool pool ON pool.provider_id = pp.id
		GROUP BY pp.id, pp.name
		ORDER BY pp.priority ASC, pp.name ASC`

	rows, err := r.pool.Query(ctx, providerQuery)
	if err != nil {
		return nil, fmt.Errorf("get pool stats by provider: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ps ProviderStat
		if err := rows.Scan(
			&ps.ProviderID, &ps.ProviderName,
			&ps.Total, &ps.Available, &ps.Assigned, &ps.Unhealthy,
		); err != nil {
			return nil, fmt.Errorf("scan provider stat: %w", err)
		}
		stats.ByProvider = append(stats.ByProvider, ps)
	}

	return &stats, nil
}
