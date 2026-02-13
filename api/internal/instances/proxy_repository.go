package instances

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetProxyConfig retrieves the proxy configuration for an instance.
func (r *Repository) GetProxyConfig(ctx context.Context, id uuid.UUID) (*ProxyConfig, error) {
	query := `SELECT proxy_url, proxy_enabled, proxy_no_websocket, proxy_only_login, proxy_no_media,
	          proxy_health_status, proxy_last_health_check, proxy_health_failures
	          FROM instances WHERE id=$1`
	row := r.pool.QueryRow(ctx, query, id)
	var cfg ProxyConfig
	if err := row.Scan(
		&cfg.ProxyURL,
		&cfg.Enabled,
		&cfg.NoWebsocket,
		&cfg.OnlyLogin,
		&cfg.NoMedia,
		&cfg.HealthStatus,
		&cfg.LastHealthCheck,
		&cfg.HealthFailures,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrInstanceNotFound
		}
		return nil, fmt.Errorf("get proxy config: %w", err)
	}
	return &cfg, nil
}

// UpdateProxyConfig sets or updates the proxy configuration for an instance.
func (r *Repository) UpdateProxyConfig(ctx context.Context, id uuid.UUID, cfg ProxyConfig) error {
	query := `UPDATE instances SET
		proxy_url = $2,
		proxy_enabled = $3,
		proxy_no_websocket = $4,
		proxy_only_login = $5,
		proxy_no_media = $6,
		proxy_health_status = 'unknown',
		proxy_health_failures = 0,
		updated_at = NOW()
	WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id,
		cfg.ProxyURL,
		cfg.Enabled,
		cfg.NoWebsocket,
		cfg.OnlyLogin,
		cfg.NoMedia,
	)
	if err != nil {
		return fmt.Errorf("update proxy config: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrInstanceNotFound
	}
	return nil
}

// ClearProxyConfig removes proxy configuration from an instance.
func (r *Repository) ClearProxyConfig(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE instances SET
		proxy_url = NULL,
		proxy_enabled = FALSE,
		proxy_no_websocket = FALSE,
		proxy_only_login = FALSE,
		proxy_no_media = FALSE,
		proxy_health_status = 'unknown',
		proxy_last_health_check = NULL,
		proxy_health_failures = 0,
		updated_at = NOW()
	WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("clear proxy config: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrInstanceNotFound
	}
	return nil
}

// UpdateProxyHealthStatus updates the health tracking fields for an instance's proxy.
func (r *Repository) UpdateProxyHealthStatus(ctx context.Context, id uuid.UUID, status string, failures int) error {
	query := `UPDATE instances SET
		proxy_health_status = $2,
		proxy_health_failures = $3,
		proxy_last_health_check = NOW(),
		updated_at = NOW()
	WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id, status, failures)
	if err != nil {
		return fmt.Errorf("update proxy health status: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrInstanceNotFound
	}
	return nil
}

// ListInstancesWithProxy returns all instances that have proxy enabled.
func (r *Repository) ListInstancesWithProxy(ctx context.Context) ([]ProxyInstance, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, proxy_url, proxy_health_status, proxy_health_failures
		FROM instances
		WHERE proxy_enabled = TRUE AND proxy_url IS NOT NULL AND proxy_url <> ''
	`)
	if err != nil {
		return nil, fmt.Errorf("list instances with proxy: %w", err)
	}
	defer rows.Close()

	var result []ProxyInstance
	for rows.Next() {
		var p ProxyInstance
		if err := rows.Scan(&p.InstanceID, &p.ProxyURL, &p.HealthStatus, &p.HealthFailures); err != nil {
			return nil, fmt.Errorf("scan proxy instance: %w", err)
		}
		result = append(result, p)
	}
	return result, nil
}

// InsertProxyHealthLog records a health check result for auditing.
func (r *Repository) InsertProxyHealthLog(ctx context.Context, log ProxyHealthLog) error {
	query := `INSERT INTO proxy_health_log (instance_id, proxy_url, status, latency_ms, error_message, checked_at)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, query,
		log.InstanceID,
		log.ProxyURL,
		log.Status,
		log.LatencyMs,
		log.ErrorMessage,
		log.CheckedAt,
	)
	if err != nil {
		return fmt.Errorf("insert proxy health log: %w", err)
	}
	return nil
}

// GetProxyHealthLogs retrieves recent health logs for an instance.
func (r *Repository) GetProxyHealthLogs(ctx context.Context, id uuid.UUID, limit int) ([]ProxyHealthLog, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, instance_id, proxy_url, status, latency_ms, error_message, checked_at
		FROM proxy_health_log
		WHERE instance_id = $1
		ORDER BY checked_at DESC
		LIMIT $2
	`, id, limit)
	if err != nil {
		return nil, fmt.Errorf("get proxy health logs: %w", err)
	}
	defer rows.Close()

	var logs []ProxyHealthLog
	for rows.Next() {
		var l ProxyHealthLog
		if err := rows.Scan(&l.ID, &l.InstanceID, &l.ProxyURL, &l.Status, &l.LatencyMs, &l.ErrorMessage, &l.CheckedAt); err != nil {
			return nil, fmt.Errorf("scan proxy health log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// CleanupProxyHealthLogs removes health logs older than the retention period.
func (r *Repository) CleanupProxyHealthLogs(ctx context.Context, retention time.Duration) (int64, error) {
	cutoff := time.Now().Add(-retention)
	res, err := r.pool.Exec(ctx, `DELETE FROM proxy_health_log WHERE checked_at < $1`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("cleanup proxy health logs: %w", err)
	}
	return res.RowsAffected(), nil
}
