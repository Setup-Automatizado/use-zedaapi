/**
 * Prometheus Metrics Type Definitions
 *
 * Types for parsing Prometheus exposition format and transforming
 * into dashboard-friendly structures.
 *
 * @module types/metrics
 */

// ============================================================================
// Prometheus Raw Types
// ============================================================================

/** Metric type from Prometheus exposition format */
export type MetricType =
	| "counter"
	| "gauge"
	| "histogram"
	| "summary"
	| "unknown";

/** Single metric sample with labels */
export interface MetricSample {
	name: string;
	labels: Record<string, string>;
	value: number;
	timestamp?: number;
}

/** Histogram bucket data */
export interface HistogramBucket {
	/** Less than or equal boundary */
	le: string;
	/** Cumulative count */
	count: number;
}

/** Histogram metric with computed percentiles */
export interface HistogramMetric {
	name: string;
	labels: Record<string, string>;
	buckets: HistogramBucket[];
	sum: number;
	count: number;
	p50?: number;
	p90?: number;
	p95?: number;
	p99?: number;
}

/** Parsed metric family from Prometheus */
export interface MetricFamily {
	name: string;
	help: string;
	type: MetricType;
	samples: MetricSample[];
}

/** Raw parsed metrics response */
export interface ParsedMetrics {
	families: Record<string, MetricFamily>;
	timestamp: string;
	parseErrors: string[];
}

// ============================================================================
// Dashboard Metrics Types
// ============================================================================

/** Dashboard-ready metrics structure */
export interface DashboardMetrics {
	timestamp: string;
	http: HTTPMetrics;
	events: EventMetrics;
	messageQueue: MessageQueueMetrics;
	media: MediaMetrics;
	system: SystemMetrics;
	workers: WorkerMetrics;
	transport: TransportMetrics;
	statusCache: StatusCacheMetrics;
	instances: string[];
}

/** HTTP request metrics */
export interface HTTPMetrics {
	totalRequests: number;
	requestsPerSecond: number;
	errorRate: number;
	avgLatencyMs: number;
	p50LatencyMs: number;
	p95LatencyMs: number;
	p99LatencyMs: number;
	byStatus: Record<string, number>;
	byPath: Record<string, PathMetrics>;
	byMethod: Record<string, number>;
}

export interface PathMetrics {
	count: number;
	avgLatencyMs: number;
	errorCount: number;
}

/** Event system metrics */
export interface EventMetrics {
	captured: number;
	buffered: number;
	inserted: number;
	processed: number;
	delivered: number;
	failed: number;
	retries: number;
	avgProcessingMs: number;
	avgDeliveryMs: number;
	outboxBacklog: number;
	dlqSize: number;
	byType: Record<string, EventTypeMetrics>;
	byInstance: Record<string, EventInstanceMetrics>;
}

export interface EventTypeMetrics {
	captured: number;
	processed: number;
	delivered: number;
	failed: number;
}

export interface EventInstanceMetrics {
	captured: number;
	processed: number;
	delivered: number;
	failed: number;
	backlog: number;
}

/** Message queue metrics */
export interface MessageQueueMetrics {
	totalSize: number;
	pending: number;
	processing: number;
	sent: number;
	failed: number;
	enqueued: number;
	processed: number;
	retries: number;
	errors: number;
	dlqSize: number;
	activeWorkers: number;
	avgProcessingMs: number;
	byInstance: Record<string, MessageQueueInstanceMetrics>;
	byType: Record<string, MessageTypeMetrics>;
}

export interface MessageQueueInstanceMetrics {
	size: number;
	workers: number;
	pending: number;
	processing: number;
	sent: number;
	failed: number;
}

export interface MessageTypeMetrics {
	enqueued: number;
	processed: number;
	failed: number;
}

/** Media processing metrics */
export interface MediaMetrics {
	downloads: MediaOperationMetrics;
	uploads: MediaOperationMetrics;
	avgDownloadMs: number;
	avgUploadMs: number;
	totalDownloadBytes: number;
	totalUploadBytes: number;
	localStorageBytes: number;
	localStorageFiles: number;
	backlog: number;
	cleanupRuns: number;
	cleanupDeletedBytes: number;
	byType: Record<string, MediaTypeMetrics>;
	byInstance: Record<string, MediaInstanceMetrics>;
}

export interface MediaOperationMetrics {
	total: number;
	success: number;
	failed: number;
}

export interface MediaTypeMetrics {
	downloads: number;
	uploads: number;
	avgSizeBytes: number;
	failures: number;
}

export interface MediaInstanceMetrics {
	downloads: number;
	uploads: number;
	failures: number;
}

/** System health metrics */
export interface SystemMetrics {
	circuitBreakerState: CircuitBreakerState;
	circuitBreakerByInstance: Record<string, CircuitBreakerState>;
	lockAcquisitions: LockMetrics;
	splitBrainDetected: number;
	healthChecks: Record<string, HealthCheckMetrics>;
	orphanedInstances: number;
	reconciliation: ReconciliationMetrics;
}

export interface LockMetrics {
	success: number;
	failure: number;
	reacquisitions: number;
	fallbacks: number;
}

export interface HealthCheckMetrics {
	healthy: number;
	unhealthy: number;
	degraded: number;
}

export interface ReconciliationMetrics {
	success: number;
	failure: number;
	skipped: number;
	error: number;
	avgDurationMs: number;
}

export type CircuitBreakerState = "closed" | "open" | "half-open" | "unknown";

/** Worker metrics */
export interface WorkerMetrics {
	active: Record<string, number>;
	errors: Record<string, number>;
	avgTaskDurationMs: Record<string, number>;
	totalActive: number;
	totalErrors: number;
}

/** Transport/Webhook delivery metrics */
export interface TransportMetrics {
	totalDeliveries: number;
	successfulDeliveries: number;
	failedDeliveries: number;
	totalRetries: number;
	successRate: number;
	avgDurationMs: number;
	p50DurationMs: number;
	p95DurationMs: number;
	p99DurationMs: number;
	byInstance: Record<string, TransportInstanceMetrics>;
	byErrorType: Record<string, number>;
}

export interface TransportInstanceMetrics {
	deliveries: number;
	success: number;
	failed: number;
	retries: number;
	avgDurationMs: number;
}

/** Status cache metrics for message status events (read, delivered, played, sent) */
export interface StatusCacheMetrics {
	/** Total cache operations (upsert, get, delete, flush) */
	totalOperations: number;
	/** Current total entries in cache across all instances */
	totalSize: number;
	/** Total cache hits */
	totalHits: number;
	/** Total cache misses */
	totalMisses: number;
	/** Hit rate percentage (0-100) */
	hitRate: number;
	/** Total webhook suppressions */
	totalSuppressions: number;
	/** Total flushed entries */
	totalFlushed: number;
	/** Average operation duration in ms */
	avgDurationMs: number;
	/** P50 operation duration in ms */
	p50DurationMs: number;
	/** P95 operation duration in ms */
	p95DurationMs: number;
	/** P99 operation duration in ms */
	p99DurationMs: number;
	/** Operations breakdown by type */
	byOperation: Record<string, StatusCacheOperationMetrics>;
	/** Metrics per instance */
	byInstance: Record<string, StatusCacheInstanceMetrics>;
	/** Suppressions by status type (read, delivered, played, sent) */
	byStatusType: Record<string, number>;
	/** Flushed entries by trigger (manual, ttl, shutdown) */
	byTrigger: Record<string, number>;
}

export interface StatusCacheOperationMetrics {
	/** Total count of this operation */
	count: number;
	/** Successful operations */
	success: number;
	/** Failed operations */
	failed: number;
	/** Average duration in ms */
	avgDurationMs: number;
}

export interface StatusCacheInstanceMetrics {
	/** Current cache size for this instance */
	size: number;
	/** Total operations */
	operations: number;
	/** Cache hits */
	hits: number;
	/** Cache misses */
	misses: number;
	/** Hit rate percentage */
	hitRate: number;
	/** Total suppressions */
	suppressions: number;
	/** Total flushed */
	flushed: number;
}

// ============================================================================
// UI/UX Types
// ============================================================================

/** Metric threshold configuration */
export interface MetricThreshold {
	warning: number;
	critical: number;
	unit: string;
	inverse?: boolean; // true if lower is worse (e.g., success rate)
}

/** Health status based on thresholds */
export type HealthLevel = "healthy" | "warning" | "critical";

/** Time range for historical data */
export type TimeRange = "5m" | "15m" | "1h" | "6h" | "24h";

/** Refresh interval options in milliseconds */
export type RefreshInterval = 5000 | 15000 | 30000 | 60000 | 0;

/** Refresh interval labels */
export const REFRESH_INTERVALS: Record<RefreshInterval, string> = {
	5000: "5 seconds",
	15000: "15 seconds",
	30000: "30 seconds",
	60000: "1 minute",
	0: "Off",
};

/** KPI card data structure */
export interface KPICardData {
	title: string;
	value: number | string;
	unit?: string;
	status: HealthLevel;
	trend?: TrendData;
	subtitle?: string;
}

export interface TrendData {
	direction: "up" | "down" | "stable";
	value: number;
	isPositive: boolean;
}

/** Chart data point */
export interface ChartDataPoint {
	timestamp: string;
	[key: string]: string | number;
}

/** Metrics API response */
export interface MetricsResponse {
	success: boolean;
	data?: DashboardMetrics;
	error?: string;
	timestamp: string;
}

/** Metrics hook options */
export interface UseMetricsOptions {
	enabled?: boolean;
	interval?: RefreshInterval;
	instanceId?: string | null;
}

/** Metrics hook result */
export interface UseMetricsResult {
	metrics: DashboardMetrics | undefined;
	isLoading: boolean;
	isValidating: boolean;
	error: Error | undefined;
	lastUpdated: Date | undefined;
	refresh: () => Promise<void>;
}
