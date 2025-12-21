/**
 * Metrics Transformer
 *
 * Transforms parsed Prometheus metrics into dashboard-friendly structures.
 * Groups and aggregates metrics by category for efficient display.
 *
 * @module lib/metrics/transformer
 */

import type {
	CircuitBreakerState,
	DashboardMetrics,
	EventMetrics,
	HTTPMetrics,
	MediaMetrics,
	MessageQueueMetrics,
	MetricFamily,
	ParsedMetrics,
	SystemMetrics,
	TransportMetrics,
	WorkerMetrics,
} from "@/types/metrics";
import { extractHistograms } from "./parser";

/**
 * Metric name prefixes to try when looking up families
 * The backend may expose metrics with or without the whatsmeow_api_ prefix
 */
const METRIC_PREFIXES = ["whatsmeow_api_", ""];

/**
 * Get metric family by name, trying common prefixes and suffix variations
 * This handles the whatsmeow_api_ prefix used by the backend and
 * the parser's behavior of grouping samples under base names (without _total suffix)
 */
function getFamily(
	families: ParsedMetrics["families"],
	baseName: string,
): MetricFamily | undefined {
	// Try with and without _total suffix
	// Prefer the version without _total first as that's where the parser puts samples
	const namesToTry: string[] = [];
	if (baseName.endsWith("_total")) {
		namesToTry.push(baseName.slice(0, -6)); // without _total first
		namesToTry.push(baseName); // then with _total
	} else {
		namesToTry.push(baseName);
	}

	for (const prefix of METRIC_PREFIXES) {
		for (const name of namesToTry) {
			const fullName = prefix + name;
			const family = families[fullName];
			// Only return if family exists and has samples
			if (family && family.samples.length > 0) {
				return family;
			}
		}
	}

	// Fallback: return any existing family even without samples
	for (const prefix of METRIC_PREFIXES) {
		for (const name of namesToTry) {
			const fullName = prefix + name;
			if (families[fullName]) {
				return families[fullName];
			}
		}
	}

	return undefined;
}

/**
 * Transform options
 */
export interface TransformOptions {
	/** Filter metrics by instance ID */
	instanceId?: string | null;
}

/**
 * Transform parsed Prometheus metrics to dashboard format
 *
 * @param parsed - Parsed Prometheus metrics
 * @param options - Transform options
 * @returns Dashboard-ready metrics structure
 */
export function transformToDashboard(
	parsed: ParsedMetrics,
	options: TransformOptions = {},
): DashboardMetrics {
	const { instanceId } = options;
	const { families } = parsed;

	// Collect all unique instance IDs
	const instances = collectInstanceIds(families);

	return {
		timestamp: parsed.timestamp,
		http: transformHTTPMetrics(families, instanceId),
		events: transformEventMetrics(families, instanceId),
		messageQueue: transformMessageQueueMetrics(families, instanceId),
		media: transformMediaMetrics(families, instanceId),
		system: transformSystemMetrics(families, instanceId),
		workers: transformWorkerMetrics(families),
		transport: transformTransportMetrics(families, instanceId),
		instances,
	};
}

/**
 * Collect all unique instance IDs from metrics
 */
function collectInstanceIds(
	families: ParsedMetrics["families"],
): string[] {
	const instanceIds = new Set<string>();

	for (const family of Object.values(families)) {
		for (const sample of family.samples) {
			if (sample.labels.instance_id) {
				instanceIds.add(sample.labels.instance_id);
			}
		}
	}

	return Array.from(instanceIds).sort();
}

/**
 * Transform HTTP-related metrics
 */
function transformHTTPMetrics(
	families: ParsedMetrics["families"],
	_instanceId?: string | null,
): HTTPMetrics {
	const metrics: HTTPMetrics = {
		totalRequests: 0,
		requestsPerSecond: 0,
		errorRate: 0,
		avgLatencyMs: 0,
		p50LatencyMs: 0,
		p95LatencyMs: 0,
		p99LatencyMs: 0,
		byStatus: {},
		byPath: {},
		byMethod: {},
	};

	// http_requests_total
	const requestsFamily = getFamily(families, "http_requests_total");
	if (requestsFamily) {
		let totalRequests = 0;
		let errorRequests = 0;

		for (const sample of requestsFamily.samples) {
			if (sample.name.endsWith("_total") || sample.name === "http_requests_total") {
				const value = sample.value;
				totalRequests += value;

				// Count by status
				const status = sample.labels.status || "unknown";
				metrics.byStatus[status] = (metrics.byStatus[status] || 0) + value;

				// Count errors (4xx and 5xx)
				if (status.startsWith("4") || status.startsWith("5")) {
					errorRequests += value;
				}

				// Count by path
				const path = sample.labels.path || "/";
				if (!metrics.byPath[path]) {
					metrics.byPath[path] = { count: 0, avgLatencyMs: 0, errorCount: 0 };
				}
				metrics.byPath[path].count += value;
				if (status.startsWith("4") || status.startsWith("5")) {
					metrics.byPath[path].errorCount += value;
				}

				// Count by method
				const method = sample.labels.method || "GET";
				metrics.byMethod[method] = (metrics.byMethod[method] || 0) + value;
			}
		}

		metrics.totalRequests = totalRequests;
		metrics.errorRate = totalRequests > 0 ? (errorRequests / totalRequests) * 100 : 0;
	}

	// http_request_duration_seconds (histogram)
	const durationFamily = getFamily(families, "http_request_duration_seconds");
	if (durationFamily) {
		const histograms = extractHistograms(durationFamily);

		if (histograms.length > 0) {
			// Aggregate percentiles across all labels
			let totalSum = 0;
			let totalCount = 0;
			let p50Sum = 0;
			let p95Sum = 0;
			let p99Sum = 0;
			let validHistograms = 0;

			for (const h of histograms) {
				totalSum += h.sum;
				totalCount += h.count;
				if (h.p50 !== undefined) {
					p50Sum += h.p50;
					validHistograms++;
				}
				if (h.p95 !== undefined) {
					p95Sum += h.p95;
				}
				if (h.p99 !== undefined) {
					p99Sum += h.p99;
				}
			}

			if (totalCount > 0) {
				metrics.avgLatencyMs = (totalSum / totalCount) * 1000;
			}
			if (validHistograms > 0) {
				metrics.p50LatencyMs = (p50Sum / validHistograms) * 1000;
				metrics.p95LatencyMs = (p95Sum / validHistograms) * 1000;
				metrics.p99LatencyMs = (p99Sum / validHistograms) * 1000;
			}
		}
	}

	return metrics;
}

/**
 * Transform event system metrics
 */
function transformEventMetrics(
	families: ParsedMetrics["families"],
	instanceId?: string | null,
): EventMetrics {
	const metrics: EventMetrics = {
		captured: 0,
		buffered: 0,
		inserted: 0,
		processed: 0,
		delivered: 0,
		failed: 0,
		retries: 0,
		avgProcessingMs: 0,
		avgDeliveryMs: 0,
		outboxBacklog: 0,
		dlqSize: 0,
		byType: {},
		byInstance: {},
	};

	// Helper to filter by instance if specified
	const shouldInclude = (labels: Record<string, string>) => {
		if (!instanceId) return true;
		return labels.instance_id === instanceId;
	};

	// events_captured_total
	const capturedFamily = getFamily(families, "events_captured_total");
	if (capturedFamily) {
		for (const sample of capturedFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.captured += sample.value;

			// By type
			const eventType = sample.labels.event_type || "unknown";
			if (!metrics.byType[eventType]) {
				metrics.byType[eventType] = { captured: 0, processed: 0, delivered: 0, failed: 0 };
			}
			metrics.byType[eventType].captured += sample.value;

			// By instance
			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = { captured: 0, processed: 0, delivered: 0, failed: 0, backlog: 0 };
				}
				metrics.byInstance[instId].captured += sample.value;
			}
		}
	}

	// events_buffered (gauge)
	const bufferedFamily = getFamily(families, "events_buffered");
	if (bufferedFamily) {
		for (const sample of bufferedFamily.samples) {
			metrics.buffered += sample.value;
		}
	}

	// events_inserted_total
	const insertedFamily = getFamily(families, "events_inserted_total");
	if (insertedFamily) {
		for (const sample of insertedFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.inserted += sample.value;
		}
	}

	// events_processed_total
	const processedFamily = getFamily(families, "events_processed_total");
	if (processedFamily) {
		for (const sample of processedFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.processed += sample.value;

			const eventType = sample.labels.event_type || "unknown";
			if (!metrics.byType[eventType]) {
				metrics.byType[eventType] = { captured: 0, processed: 0, delivered: 0, failed: 0 };
			}
			metrics.byType[eventType].processed += sample.value;

			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = { captured: 0, processed: 0, delivered: 0, failed: 0, backlog: 0 };
				}
				metrics.byInstance[instId].processed += sample.value;
			}
		}
	}

	// events_delivered_total
	const deliveredFamily = getFamily(families, "events_delivered_total");
	if (deliveredFamily) {
		for (const sample of deliveredFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.delivered += sample.value;

			const eventType = sample.labels.event_type || "unknown";
			if (!metrics.byType[eventType]) {
				metrics.byType[eventType] = { captured: 0, processed: 0, delivered: 0, failed: 0 };
			}
			metrics.byType[eventType].delivered += sample.value;

			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = { captured: 0, processed: 0, delivered: 0, failed: 0, backlog: 0 };
				}
				metrics.byInstance[instId].delivered += sample.value;
			}
		}
	}

	// events_failed_total
	const failedFamily = getFamily(families, "events_failed_total");
	if (failedFamily) {
		for (const sample of failedFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.failed += sample.value;

			const eventType = sample.labels.event_type || "unknown";
			if (!metrics.byType[eventType]) {
				metrics.byType[eventType] = { captured: 0, processed: 0, delivered: 0, failed: 0 };
			}
			metrics.byType[eventType].failed += sample.value;

			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = { captured: 0, processed: 0, delivered: 0, failed: 0, backlog: 0 };
				}
				metrics.byInstance[instId].failed += sample.value;
			}
		}
	}

	// event_retries_total
	const retriesFamily = getFamily(families, "event_retries_total");
	if (retriesFamily) {
		for (const sample of retriesFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.retries += sample.value;
		}
	}

	// event_outbox_backlog (gauge)
	const backlogFamily = getFamily(families, "event_outbox_backlog");
	if (backlogFamily) {
		for (const sample of backlogFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.outboxBacklog += sample.value;

			const instId = sample.labels.instance_id;
			if (instId && metrics.byInstance[instId]) {
				metrics.byInstance[instId].backlog = sample.value;
			}
		}
	}

	// dlq_backlog (gauge)
	const dlqFamily = getFamily(families, "dlq_backlog");
	if (dlqFamily) {
		for (const sample of dlqFamily.samples) {
			metrics.dlqSize += sample.value;
		}
	}

	// event_processing_duration_seconds (histogram)
	const processingDurationFamily = getFamily(families, "event_processing_duration_seconds");
	if (processingDurationFamily) {
		const histograms = extractHistograms(processingDurationFamily);
		let totalSum = 0;
		let totalCount = 0;

		for (const h of histograms) {
			if (instanceId && h.labels.instance_id !== instanceId) continue;
			totalSum += h.sum;
			totalCount += h.count;
		}

		if (totalCount > 0) {
			metrics.avgProcessingMs = (totalSum / totalCount) * 1000;
		}
	}

	// event_delivery_duration_seconds (histogram)
	const deliveryDurationFamily = getFamily(families, "event_delivery_duration_seconds");
	if (deliveryDurationFamily) {
		const histograms = extractHistograms(deliveryDurationFamily);
		let totalSum = 0;
		let totalCount = 0;

		for (const h of histograms) {
			if (instanceId && h.labels.instance_id !== instanceId) continue;
			totalSum += h.sum;
			totalCount += h.count;
		}

		if (totalCount > 0) {
			metrics.avgDeliveryMs = (totalSum / totalCount) * 1000;
		}
	}

	return metrics;
}

/**
 * Transform message queue metrics
 */
function transformMessageQueueMetrics(
	families: ParsedMetrics["families"],
	instanceId?: string | null,
): MessageQueueMetrics {
	const metrics: MessageQueueMetrics = {
		totalSize: 0,
		pending: 0,
		processing: 0,
		sent: 0,
		failed: 0,
		enqueued: 0,
		processed: 0,
		retries: 0,
		errors: 0,
		dlqSize: 0,
		activeWorkers: 0,
		avgProcessingMs: 0,
		byInstance: {},
		byType: {},
	};

	const shouldInclude = (labels: Record<string, string>) => {
		if (!instanceId) return true;
		return labels.instance_id === instanceId;
	};

	// message_queue_size (gauge)
	const sizeFamily = getFamily(families, "message_queue_size");
	if (sizeFamily) {
		for (const sample of sizeFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;

			const status = sample.labels.status || "unknown";
			const instId = sample.labels.instance_id;

			metrics.totalSize += sample.value;

			switch (status) {
				case "pending":
					metrics.pending += sample.value;
					break;
				case "processing":
					metrics.processing += sample.value;
					break;
				case "sent":
					metrics.sent += sample.value;
					break;
				case "failed":
					metrics.failed += sample.value;
					break;
			}

			if (instId) {
				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = { size: 0, workers: 0, pending: 0, processing: 0, sent: 0, failed: 0 };
				}
				metrics.byInstance[instId].size += sample.value;
				if (status === "pending") metrics.byInstance[instId].pending += sample.value;
				if (status === "processing") metrics.byInstance[instId].processing += sample.value;
				if (status === "sent") metrics.byInstance[instId].sent += sample.value;
				if (status === "failed") metrics.byInstance[instId].failed += sample.value;
			}
		}
	}

	// message_queue_enqueued_total
	const enqueuedFamily = getFamily(families, "message_queue_enqueued_total");
	if (enqueuedFamily) {
		for (const sample of enqueuedFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.enqueued += sample.value;

			const msgType = sample.labels.message_type || "unknown";
			if (!metrics.byType[msgType]) {
				metrics.byType[msgType] = { enqueued: 0, processed: 0, failed: 0 };
			}
			metrics.byType[msgType].enqueued += sample.value;
		}
	}

	// message_queue_processed_total
	const processedFamily = getFamily(families, "message_queue_processed_total");
	if (processedFamily) {
		for (const sample of processedFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.processed += sample.value;

			const msgType = sample.labels.message_type || "unknown";
			if (!metrics.byType[msgType]) {
				metrics.byType[msgType] = { enqueued: 0, processed: 0, failed: 0 };
			}
			const status = sample.labels.status;
			if (status === "failed") {
				metrics.byType[msgType].failed += sample.value;
			} else {
				metrics.byType[msgType].processed += sample.value;
			}
		}
	}

	// message_queue_retries_total
	const retriesFamily = getFamily(families, "message_queue_retries_total");
	if (retriesFamily) {
		for (const sample of retriesFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.retries += sample.value;
		}
	}

	// message_queue_errors_total
	const errorsFamily = getFamily(families, "message_queue_errors_total");
	if (errorsFamily) {
		for (const sample of errorsFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.errors += sample.value;
		}
	}

	// message_queue_dlq_size (gauge)
	const dlqFamily = getFamily(families, "message_queue_dlq_size");
	if (dlqFamily) {
		for (const sample of dlqFamily.samples) {
			metrics.dlqSize += sample.value;
		}
	}

	// message_queue_workers_active (gauge)
	const workersFamily = getFamily(families, "message_queue_workers_active");
	if (workersFamily) {
		for (const sample of workersFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.activeWorkers += sample.value;

			const instId = sample.labels.instance_id;
			if (instId && metrics.byInstance[instId]) {
				metrics.byInstance[instId].workers = sample.value;
			}
		}
	}

	// message_queue_processing_duration_seconds (histogram)
	const durationFamily = getFamily(families, "message_queue_processing_duration_seconds");
	if (durationFamily) {
		const histograms = extractHistograms(durationFamily);
		let totalSum = 0;
		let totalCount = 0;

		for (const h of histograms) {
			if (instanceId && h.labels.instance_id !== instanceId) continue;
			totalSum += h.sum;
			totalCount += h.count;
		}

		if (totalCount > 0) {
			metrics.avgProcessingMs = (totalSum / totalCount) * 1000;
		}
	}

	return metrics;
}

/**
 * Transform media processing metrics
 */
function transformMediaMetrics(
	families: ParsedMetrics["families"],
	instanceId?: string | null,
): MediaMetrics {
	const metrics: MediaMetrics = {
		downloads: { total: 0, success: 0, failed: 0 },
		uploads: { total: 0, success: 0, failed: 0 },
		avgDownloadMs: 0,
		avgUploadMs: 0,
		totalDownloadBytes: 0,
		totalUploadBytes: 0,
		localStorageBytes: 0,
		localStorageFiles: 0,
		backlog: 0,
		cleanupRuns: 0,
		cleanupDeletedBytes: 0,
		byType: {},
		byInstance: {},
	};

	const shouldInclude = (labels: Record<string, string>) => {
		if (!instanceId) return true;
		return labels.instance_id === instanceId;
	};

	// media_downloads_total
	const downloadsFamily = getFamily(families, "media_downloads_total");
	if (downloadsFamily) {
		for (const sample of downloadsFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;

			metrics.downloads.total += sample.value;
			const status = sample.labels.status;
			if (status === "success") {
				metrics.downloads.success += sample.value;
			} else if (status === "failure") {
				metrics.downloads.failed += sample.value;
			}

			const mediaType = sample.labels.media_type || "unknown";
			if (!metrics.byType[mediaType]) {
				metrics.byType[mediaType] = { downloads: 0, uploads: 0, avgSizeBytes: 0, failures: 0 };
			}
			metrics.byType[mediaType].downloads += sample.value;
			if (status === "failure") {
				metrics.byType[mediaType].failures += sample.value;
			}

			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = { downloads: 0, uploads: 0, failures: 0 };
				}
				metrics.byInstance[instId].downloads += sample.value;
				if (status === "failure") {
					metrics.byInstance[instId].failures += sample.value;
				}
			}
		}
	}

	// media_uploads_total
	const uploadsFamily = getFamily(families, "media_uploads_total");
	if (uploadsFamily) {
		for (const sample of uploadsFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;

			metrics.uploads.total += sample.value;
			const status = sample.labels.status;
			if (status === "success") {
				metrics.uploads.success += sample.value;
			} else if (status === "failure") {
				metrics.uploads.failed += sample.value;
			}

			const mediaType = sample.labels.media_type || "unknown";
			if (!metrics.byType[mediaType]) {
				metrics.byType[mediaType] = { downloads: 0, uploads: 0, avgSizeBytes: 0, failures: 0 };
			}
			metrics.byType[mediaType].uploads += sample.value;

			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = { downloads: 0, uploads: 0, failures: 0 };
				}
				metrics.byInstance[instId].uploads += sample.value;
			}
		}
	}

	// media_backlog (gauge)
	const backlogFamily = getFamily(families, "media_backlog");
	if (backlogFamily) {
		for (const sample of backlogFamily.samples) {
			metrics.backlog += sample.value;
		}
	}

	// media_local_storage_bytes (gauge)
	const storageBytesFamily = getFamily(families, "media_local_storage_bytes");
	if (storageBytesFamily) {
		for (const sample of storageBytesFamily.samples) {
			metrics.localStorageBytes += sample.value;
		}
	}

	// media_local_storage_files (gauge)
	const storageFilesFamily = getFamily(families, "media_local_storage_files");
	if (storageFilesFamily) {
		for (const sample of storageFilesFamily.samples) {
			metrics.localStorageFiles += sample.value;
		}
	}

	// media_cleanup_runs_total
	const cleanupFamily = getFamily(families, "media_cleanup_runs_total");
	if (cleanupFamily) {
		for (const sample of cleanupFamily.samples) {
			metrics.cleanupRuns += sample.value;
		}
	}

	// media_cleanup_deleted_bytes_total
	const cleanupBytesFamily = getFamily(families, "media_cleanup_deleted_bytes_total");
	if (cleanupBytesFamily) {
		for (const sample of cleanupBytesFamily.samples) {
			metrics.cleanupDeletedBytes += sample.value;
		}
	}

	// media_download_duration_seconds (histogram)
	const downloadDurationFamily = getFamily(families, "media_download_duration_seconds");
	if (downloadDurationFamily) {
		const histograms = extractHistograms(downloadDurationFamily);
		let totalSum = 0;
		let totalCount = 0;

		for (const h of histograms) {
			if (instanceId && h.labels.instance_id !== instanceId) continue;
			totalSum += h.sum;
			totalCount += h.count;
		}

		if (totalCount > 0) {
			metrics.avgDownloadMs = (totalSum / totalCount) * 1000;
		}
	}

	// media_upload_duration_seconds (histogram)
	const uploadDurationFamily = getFamily(families, "media_upload_duration_seconds");
	if (uploadDurationFamily) {
		const histograms = extractHistograms(uploadDurationFamily);
		let totalSum = 0;
		let totalCount = 0;

		for (const h of histograms) {
			if (instanceId && h.labels.instance_id !== instanceId) continue;
			totalSum += h.sum;
			totalCount += h.count;
		}

		if (totalCount > 0) {
			metrics.avgUploadMs = (totalSum / totalCount) * 1000;
		}
	}

	return metrics;
}

/**
 * Transform system health metrics
 */
function transformSystemMetrics(
	families: ParsedMetrics["families"],
	_instanceId?: string | null,
): SystemMetrics {
	const metrics: SystemMetrics = {
		circuitBreakerState: "unknown",
		circuitBreakerByInstance: {},
		lockAcquisitions: { success: 0, failure: 0, reacquisitions: 0, fallbacks: 0 },
		splitBrainDetected: 0,
		healthChecks: {},
		orphanedInstances: 0,
		reconciliation: { success: 0, failure: 0, skipped: 0, error: 0, avgDurationMs: 0 },
	};

	// circuit_breaker_state (gauge)
	const circuitFamily = getFamily(families, "circuit_breaker_state");
	if (circuitFamily) {
		for (const sample of circuitFamily.samples) {
			metrics.circuitBreakerState = mapCircuitState(sample.value);
		}
	}

	// circuit_breaker_state_per_instance (gauge)
	const circuitPerInstanceFamily = getFamily(families, "circuit_breaker_state_per_instance");
	if (circuitPerInstanceFamily) {
		for (const sample of circuitPerInstanceFamily.samples) {
			const instId = sample.labels.instance_id;
			if (instId) {
				metrics.circuitBreakerByInstance[instId] = mapCircuitState(sample.value);
			}
		}
	}

	// lock_acquisitions_total
	const lockAcqFamily = getFamily(families, "lock_acquisitions_total");
	if (lockAcqFamily) {
		for (const sample of lockAcqFamily.samples) {
			const status = sample.labels.status;
			if (status === "success") {
				metrics.lockAcquisitions.success += sample.value;
			} else if (status === "failure") {
				metrics.lockAcquisitions.failure += sample.value;
			}
		}
	}

	// lock_reacquisition_attempts_total
	const reacqFamily = getFamily(families, "lock_reacquisition_attempts_total");
	if (reacqFamily) {
		for (const sample of reacqFamily.samples) {
			metrics.lockAcquisitions.reacquisitions += sample.value;
		}
	}

	// lock_reacquisition_fallbacks_total
	const fallbacksFamily = getFamily(families, "lock_reacquisition_fallbacks_total");
	if (fallbacksFamily) {
		for (const sample of fallbacksFamily.samples) {
			metrics.lockAcquisitions.fallbacks += sample.value;
		}
	}

	// split_brain_detected_total
	const splitBrainFamily = getFamily(families, "split_brain_detected_total");
	if (splitBrainFamily) {
		for (const sample of splitBrainFamily.samples) {
			metrics.splitBrainDetected += sample.value;
		}
	}

	// health_checks_total
	const healthFamily = getFamily(families, "health_checks_total");
	if (healthFamily) {
		for (const sample of healthFamily.samples) {
			const component = sample.labels.component || "unknown";
			const status = sample.labels.status || "unknown";

			if (!metrics.healthChecks[component]) {
				metrics.healthChecks[component] = { healthy: 0, unhealthy: 0, degraded: 0 };
			}

			if (status === "healthy") {
				metrics.healthChecks[component].healthy += sample.value;
			} else if (status === "unhealthy") {
				metrics.healthChecks[component].unhealthy += sample.value;
			} else if (status === "degraded") {
				metrics.healthChecks[component].degraded += sample.value;
			}
		}
	}

	// orphaned_instances (gauge)
	const orphanedFamily = getFamily(families, "orphaned_instances");
	if (orphanedFamily) {
		for (const sample of orphanedFamily.samples) {
			metrics.orphanedInstances += sample.value;
		}
	}

	// reconciliation_attempts_total
	const reconFamily = getFamily(families, "reconciliation_attempts_total");
	if (reconFamily) {
		for (const sample of reconFamily.samples) {
			const result = sample.labels.result;
			switch (result) {
				case "success":
					metrics.reconciliation.success += sample.value;
					break;
				case "failure":
					metrics.reconciliation.failure += sample.value;
					break;
				case "skipped":
					metrics.reconciliation.skipped += sample.value;
					break;
				case "error":
					metrics.reconciliation.error += sample.value;
					break;
			}
		}
	}

	// reconciliation_duration_seconds (histogram)
	const reconDurationFamily = getFamily(families, "reconciliation_duration_seconds");
	if (reconDurationFamily) {
		const histograms = extractHistograms(reconDurationFamily);
		let totalSum = 0;
		let totalCount = 0;

		for (const h of histograms) {
			totalSum += h.sum;
			totalCount += h.count;
		}

		if (totalCount > 0) {
			metrics.reconciliation.avgDurationMs = (totalSum / totalCount) * 1000;
		}
	}

	return metrics;
}

/**
 * Transform worker metrics
 */
function transformWorkerMetrics(
	families: ParsedMetrics["families"],
): WorkerMetrics {
	const metrics: WorkerMetrics = {
		active: {},
		errors: {},
		avgTaskDurationMs: {},
		totalActive: 0,
		totalErrors: 0,
	};

	// workers_active (gauge)
	const activeFamily = getFamily(families, "workers_active");
	if (activeFamily) {
		for (const sample of activeFamily.samples) {
			const workerType = sample.labels.worker_type || "unknown";
			metrics.active[workerType] = (metrics.active[workerType] || 0) + sample.value;
			metrics.totalActive += sample.value;
		}
	}

	// worker_errors_total
	const errorsFamily = getFamily(families, "worker_errors_total");
	if (errorsFamily) {
		for (const sample of errorsFamily.samples) {
			const workerType = sample.labels.worker_type || "unknown";
			metrics.errors[workerType] = (metrics.errors[workerType] || 0) + sample.value;
			metrics.totalErrors += sample.value;
		}
	}

	// worker_task_duration_seconds (histogram)
	const durationFamily = getFamily(families, "worker_task_duration_seconds");
	if (durationFamily) {
		const histograms = extractHistograms(durationFamily);

		for (const h of histograms) {
			const workerType = h.labels.worker_type || "unknown";
			if (h.count > 0) {
				const avgMs = (h.sum / h.count) * 1000;
				// Average if multiple task types
				if (metrics.avgTaskDurationMs[workerType]) {
					metrics.avgTaskDurationMs[workerType] = (metrics.avgTaskDurationMs[workerType] + avgMs) / 2;
				} else {
					metrics.avgTaskDurationMs[workerType] = avgMs;
				}
			}
		}
	}

	return metrics;
}

/**
 * Transform transport/webhook delivery metrics
 */
function transformTransportMetrics(
	families: ParsedMetrics["families"],
	instanceId?: string | null,
): TransportMetrics {
	const metrics: TransportMetrics = {
		totalDeliveries: 0,
		successfulDeliveries: 0,
		failedDeliveries: 0,
		totalRetries: 0,
		successRate: 0,
		avgDurationMs: 0,
		p50DurationMs: 0,
		p95DurationMs: 0,
		p99DurationMs: 0,
		byInstance: {},
		byErrorType: {},
	};

	const shouldInclude = (labels: Record<string, string>) => {
		if (!instanceId) return true;
		return labels.instance_id === instanceId;
	};

	// transport_deliveries_total
	const deliveriesFamily = getFamily(families, "transport_deliveries_total");
	if (deliveriesFamily) {
		for (const sample of deliveriesFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;

			const status = sample.labels.status || "unknown";
			const instId = sample.labels.instance_id;

			metrics.totalDeliveries += sample.value;

			if (status === "success") {
				metrics.successfulDeliveries += sample.value;
			} else if (status === "failed") {
				metrics.failedDeliveries += sample.value;
			}

			if (instId) {
				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = {
						deliveries: 0,
						success: 0,
						failed: 0,
						retries: 0,
						avgDurationMs: 0,
					};
				}
				metrics.byInstance[instId].deliveries += sample.value;
				if (status === "success") {
					metrics.byInstance[instId].success += sample.value;
				} else if (status === "failed") {
					metrics.byInstance[instId].failed += sample.value;
				}
			}
		}
	}

	// transport_errors_total
	const errorsFamily = getFamily(families, "transport_errors_total");
	if (errorsFamily) {
		for (const sample of errorsFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;

			const errorType = sample.labels.error_type || "unknown";
			metrics.byErrorType[errorType] = (metrics.byErrorType[errorType] || 0) + sample.value;
		}
	}

	// transport_retries_total
	const retriesFamily = getFamily(families, "transport_retries_total");
	if (retriesFamily) {
		for (const sample of retriesFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;

			metrics.totalRetries += sample.value;

			const instId = sample.labels.instance_id;
			if (instId && metrics.byInstance[instId]) {
				metrics.byInstance[instId].retries += sample.value;
			}
		}
	}

	// transport_duration_seconds (histogram)
	const durationFamily = getFamily(families, "transport_duration_seconds");
	if (durationFamily) {
		const histograms = extractHistograms(durationFamily);
		let totalSum = 0;
		let totalCount = 0;
		let p50Sum = 0;
		let p95Sum = 0;
		let p99Sum = 0;
		let validHistograms = 0;

		for (const h of histograms) {
			if (instanceId && h.labels.instance_id !== instanceId) continue;

			totalSum += h.sum;
			totalCount += h.count;

			if (h.p50 !== undefined) {
				p50Sum += h.p50;
				validHistograms++;
			}
			if (h.p95 !== undefined) {
				p95Sum += h.p95;
			}
			if (h.p99 !== undefined) {
				p99Sum += h.p99;
			}

			// Update per-instance avg duration
			const instId = h.labels.instance_id;
			if (instId && metrics.byInstance[instId] && h.count > 0) {
				metrics.byInstance[instId].avgDurationMs = (h.sum / h.count) * 1000;
			}
		}

		if (totalCount > 0) {
			metrics.avgDurationMs = (totalSum / totalCount) * 1000;
		}
		if (validHistograms > 0) {
			metrics.p50DurationMs = (p50Sum / validHistograms) * 1000;
			metrics.p95DurationMs = (p95Sum / validHistograms) * 1000;
			metrics.p99DurationMs = (p99Sum / validHistograms) * 1000;
		}
	}

	// Calculate success rate
	if (metrics.totalDeliveries > 0) {
		metrics.successRate = (metrics.successfulDeliveries / metrics.totalDeliveries) * 100;
	}

	return metrics;
}

/**
 * Map numeric circuit breaker state to string
 */
function mapCircuitState(value: number): CircuitBreakerState {
	switch (value) {
		case 0:
			return "closed";
		case 1:
			return "open";
		case 2:
			return "half-open";
		default:
			return "unknown";
	}
}
