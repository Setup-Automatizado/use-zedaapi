export interface PrometheusMetric {
	name: string;
	type: "counter" | "gauge" | "histogram" | "summary";
	help: string;
	metrics: Array<{
		labels: Record<string, string>;
		value: number;
	}>;
}

export interface InstanceMetrics {
	messages_sent: number;
	messages_received: number;
	messages_failed: number;
	avg_latency_ms: number;
	transport_errors: number;
}

/**
 * Refresh interval in milliseconds
 */
export type RefreshInterval = 0 | 5000 | 10000 | 15000 | 30000 | 60000;

/**
 * Metric types
 */
export type MetricType =
	| "counter"
	| "gauge"
	| "histogram"
	| "summary"
	| "unknown";

/**
 * Metric sample
 */
export interface MetricSample {
	name: string;
	labels: Record<string, string>;
	value: number;
	timestamp?: number;
}

/**
 * Histogram bucket
 */
export interface HistogramBucket {
	le: string;
	count: number;
}

/**
 * Histogram metric
 */
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

/**
 * Metric family
 */
export interface MetricFamily {
	name: string;
	type: MetricType;
	help: string;
	samples: MetricSample[];
}

/**
 * Parsed Prometheus metrics
 */
export interface ParsedMetrics {
	families: Record<string, MetricFamily>;
	timestamp: string;
	parseErrors: string[];
}

/**
 * Metrics API response
 */
export interface MetricsResponse {
	success: boolean;
	data?: ParsedMetrics | DashboardMetrics;
	error?: string;
	timestamp: string;
}

/**
 * Health status level
 */
export type HealthLevel = "healthy" | "warning" | "critical";

/**
 * Metric threshold configuration
 */
export interface MetricThreshold {
	warning: number;
	critical: number;
	unit?: string;
	inverse?: boolean;
}

/**
 * Trend data for metrics
 */
export interface TrendData {
	direction: "up" | "down" | "stable";
	percentage: number;
	value: number;
	isPositive: boolean;
}

/**
 * Circuit breaker state
 */
export type CircuitBreakerState = "open" | "half-open" | "closed" | "unknown";

/**
 * HTTP metrics
 */
export interface HTTPMetrics {
	totalRequests: number;
	requestsPerSecond: number;
	errorRate: number;
	avgLatencyMs: number;
	p50LatencyMs: number;
	p95LatencyMs: number;
	p99LatencyMs: number;
	byStatus?: Record<string, number>;
	byMethod?: Record<string, number>;
	byPath?: Record<
		string,
		{
			count: number;
			avgLatencyMs: number;
			errorCount: number;
		}
	>;
}

/**
 * Media metrics
 */
export interface MediaMetrics {
	downloads: {
		total: number;
		success: number;
		failed: number;
	};
	uploads: {
		total: number;
		success: number;
		failed: number;
	};
	backlog: number;
	totalDownloadBytes: number;
	totalUploadBytes: number;
	localStorageBytes: number;
	localStorageFiles: number;
	s3StorageBytes: number;
	s3StorageFiles: number;
	avgProcessingMs: number;
	avgDownloadMs: number;
	avgUploadMs: number;
	cleanupRuns: number;
	filesDeleted: number;
	cleanupDeletedBytes: number;
	// New detailed metrics
	downloadErrors: number;
	uploadAttempts: number;
	uploadErrors: number;
	uploadSizeBytes: number;
	presignedUrlGenerated: number;
	deleteAttempts: number;
	failures: number;
	fallbackAttempts: number;
	fallbackSuccess: number;
	fallbackFailure: number;
	cleanupTotal: number;
	cleanupErrors: number;
	avgCleanupDurationMs: number;
	serveRequests: number;
	serveBytes: number;
	byType?: Record<
		string,
		{
			downloads: number;
			uploads: number;
			failures: number;
		}
	>;
	byInstance?: Record<
		string,
		{
			downloads: number;
			uploads: number;
			failures: number;
		}
	>;
	byErrorType?: Record<string, number>;
	byFallbackType?: Record<string, number>;
	byStorageType?: Record<string, number>;
}

/**
 * System metrics
 */
export interface SystemMetrics {
	circuitBreakerByInstance: Record<string, string>;
	circuitBreakerState: CircuitBreakerState;
	circuitBreakerTransitions: number;
	healthChecks: Record<
		string,
		{
			healthy: number;
			unhealthy: number;
			degraded: number;
		}
	>;
	lockAcquisitions: {
		success: number;
		failure: number;
		reacquisitions: number;
		fallbacks: number;
	};
	splitBrainInvalidLocks: number;
	reconciliation: {
		success: number;
		failure: number;
		skipped: number;
		error: number;
		avgDurationMs: number;
	};
	queueDrain: {
		durationMs: number;
		timeouts: number;
		drainedMessages: number;
	};
	splitBrainDetected: number;
	orphanedInstances: number;
}

/**
 * Worker metrics
 */
export interface WorkerMetrics {
	totalActive: number;
	active: Record<string, number>;
	errors: Record<string, number>;
	avgTaskDurationMs: Record<string, number>;
}

/**
 * Transport metrics
 */
export interface TransportMetrics {
	totalMessages: number;
	totalDeliveries: number;
	totalRetries: number;
	sent: number;
	received: number;
	failed: number;
	errors: number;
	avgLatencyMs: number;
	p50LatencyMs: number;
	p95LatencyMs: number;
	p99LatencyMs: number;
	successRate: number;
	avgDurationMs: number;
	successfulDeliveries: number;
	failedDeliveries: number;
	p50DurationMs: number;
	p95DurationMs: number;
	p99DurationMs: number;
	byType?: Record<
		string,
		{
			sent: number;
			received: number;
			failed: number;
		}
	>;
	byInstance: Record<
		string,
		{
			sent: number;
			received: number;
			failed: number;
			errors: number;
			avgLatencyMs: number;
			deliveries: number;
			success: number;
			retries: number;
			errorRate: number;
			avgDurationMs: number;
		}
	>;
	byErrorType: Record<string, number>;
}

/**
 * Status Cache metrics
 */
export interface StatusCacheMetrics {
	totalOperations: number;
	totalSize: number;
	totalHits: number;
	totalMisses: number;
	totalSuppressions: number;
	totalFlushed: number;
	hitRate: number;
	avgDurationMs: number;
	p50DurationMs: number;
	p95DurationMs: number;
	p99DurationMs: number;
	byInstance: Record<
		string,
		{
			size: number;
			operations: number;
			hits: number;
			misses: number;
			hitRate: number;
			suppressions: number;
			flushed: number;
		}
	>;
	byStatusType?: Record<string, number>;
	byTrigger?: Record<string, number>;
	byOperation?: Record<
		string,
		{
			success: number;
			failed: number;
			count: number;
		}
	>;
}

/**
 * Message Queue metrics
 */
export interface MessageQueueMetrics {
	totalSize: number;
	pending: number;
	processing: number;
	sent: number;
	enqueued: number;
	processed: number;
	failed: number;
	retries: number;
	errors: number;
	activeWorkers: number;
	avgProcessingMs: number;
	dlqSize: number;
	byType: Record<
		string,
		{
			enqueued: number;
			processed: number;
			failed: number;
		}
	>;
	byInstance: Record<
		string,
		{
			size: number;
			pending: number;
			processing: number;
			sent: number;
			failed: number;
			workers: number;
		}
	>;
}

/**
 * Event metrics
 */
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
	// DLQ detailed metrics
	dlqEvents: number;
	dlqReprocessAttempts: number;
	dlqReprocessSuccess: number;
	// Sequence gaps
	sequenceGaps: number;
	byType?: Record<
		string,
		{
			captured: number;
			buffered: number;
			inserted: number;
			processed: number;
			delivered: number;
			failed: number;
		}
	>;
	byInstance?: Record<
		string,
		{
			captured: number;
			buffered: number;
			inserted: number;
			processed: number;
			delivered: number;
			failed: number;
			backlog: number;
			sequenceGaps: number;
		}
	>;
}

/**
 * Handler metrics (for groups, communities, newsletters)
 */
export interface HandlerMetrics {
	totalRequests: number;
	successRequests: number;
	failedRequests: number;
	avgLatencyMs: number;
	byOperation?: Record<
		string,
		{
			total: number;
			success: number;
			failed: number;
			avgLatencyMs: number;
		}
	>;
}

/**
 * Use Metrics Options
 */
export interface UseMetricsOptions {
	enabled?: boolean;
	interval?: number;
	instanceId?: string | null;
}

/**
 * Use Metrics Result
 */
export interface UseMetricsResult {
	metrics: DashboardMetrics | undefined;
	isLoading: boolean;
	isError: boolean;
	isValidating: boolean;
	error: Error | undefined;
	refresh: () => Promise<void>;
	lastUpdated: Date | undefined;
}

/**
 * Dashboard metrics
 */
export interface DashboardMetrics {
	timestamp: string;
	http: HTTPMetrics;
	queue: {
		size: number;
		processing: number;
	};
	messageQueue: MessageQueueMetrics;
	events: EventMetrics;
	workers: WorkerMetrics;
	media: MediaMetrics;
	transport: TransportMetrics;
	statusCache: StatusCacheMetrics;
	// Handler metrics for specialized endpoints
	groups: HandlerMetrics;
	communities: HandlerMetrics;
	newsletters: HandlerMetrics;
	system: SystemMetrics & {
		circuitBreakerState: CircuitBreakerState;
		lockAcquisitions: {
			success: number;
			failure: number;
		};
		circuitBreakerByInstance?: Record<string, string>;
		splitBrainDetected: number;
	};
	instances?: string[];
}
