/**
 * Health Check Types
 *
 * Type definitions for service health monitoring and readiness checks.
 * Used for Kubernetes probes and service mesh integration.
 */

/**
 * Component health status levels
 */
export type HealthStatus = 'healthy' | 'degraded' | 'unhealthy';

/**
 * Circuit breaker states for fault tolerance
 */
export type CircuitState = 'closed' | 'open' | 'half-open';

/**
 * Basic health check response
 * Lightweight endpoint for liveness probes
 */
export interface HealthResponse {
  /** Overall service status */
  readonly status: 'ok';

  /** Service identifier */
  service: string;

  /** Response timestamp (ISO 8601) */
  timestamp?: string;
}

/**
 * Detailed readiness check response
 * Comprehensive dependency status for readiness probes
 */
export interface ReadinessResponse {
  /** Overall readiness status */
  ready: boolean;

  /** Observation timestamp (ISO 8601) */
  observed_at: string;

  /** Component-level health checks */
  checks: HealthChecks;
}

/**
 * Individual component health checks
 */
export interface HealthChecks {
  /** PostgreSQL database connectivity */
  database: ComponentStatus;

  /** Redis cache/lock manager connectivity */
  redis: ComponentStatus;

  /** S3/MinIO storage connectivity (if configured) */
  storage?: ComponentStatus;

  /** WhatsApp service connectivity */
  whatsapp?: ComponentStatus;
}

/**
 * Component status details
 * Includes health level, timing, and circuit breaker state
 */
export interface ComponentStatus {
  /** Component health level */
  status: HealthStatus;

  /** Error message if unhealthy or degraded */
  error?: string;

  /** Health check duration in milliseconds */
  duration_ms: number;

  /** Circuit breaker state for fault tolerance */
  circuit_state?: CircuitState;

  /** Additional component-specific metadata */
  metadata?: Record<string, unknown>;
}

/**
 * Extended health check with version information
 */
export interface DetailedHealthResponse extends HealthResponse {
  /** Application version */
  version: string;

  /** Build commit hash */
  commit?: string;

  /** Build timestamp */
  build_time?: string;

  /** Go runtime version */
  go_version?: string;

  /** Service uptime in seconds */
  uptime_seconds?: number;
}

/**
 * Health check options for configurable checks
 */
export interface HealthCheckOptions {
  /** Include version information */
  includeVersion?: boolean;

  /** Include detailed component checks */
  includeComponents?: boolean;

  /** Timeout for health checks in milliseconds */
  timeout?: number;
}

/**
 * Type guard to check if service is healthy
 */
export function isHealthy(response: ReadinessResponse): boolean {
  return response.ready &&
         response.checks.database.status === 'healthy' &&
         response.checks.redis.status === 'healthy';
}

/**
 * Type guard to check if service is degraded
 */
export function isDegraded(response: ReadinessResponse): boolean {
  return response.ready &&
         (response.checks.database.status === 'degraded' ||
          response.checks.redis.status === 'degraded');
}

/**
 * Type guard to check if component is operational
 */
export function isComponentOperational(status: ComponentStatus): boolean {
  return status.status === 'healthy' || status.status === 'degraded';
}

/**
 * Get the most critical component status
 */
export function getCriticalStatus(checks: HealthChecks): HealthStatus {
  const statuses = Object.values(checks)
    .map(check => check.status)
    .filter((status): status is HealthStatus => status !== undefined);

  if (statuses.includes('unhealthy')) return 'unhealthy';
  if (statuses.includes('degraded')) return 'degraded';
  return 'healthy';
}

/**
 * Calculate average health check duration
 */
export function getAverageHealthCheckDuration(checks: HealthChecks): number {
  const durations = Object.values(checks)
    .map(check => check.duration_ms)
    .filter((duration): duration is number => duration !== undefined);

  if (durations.length === 0) return 0;
  return durations.reduce((sum, duration) => sum + duration, 0) / durations.length;
}
