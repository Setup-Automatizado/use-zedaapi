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
	HandlerMetrics,
	HTTPMetrics,
	MediaMetrics,
	MessageQueueMetrics,
	MetricFamily,
	ParsedMetrics,
	StatusCacheMetrics,
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
		queue: {
			size: 0,
			processing: 0,
		},
		events: transformEventMetrics(families, instanceId),
		messageQueue: transformMessageQueueMetrics(families, instanceId),
		media: transformMediaMetrics(families, instanceId),
		system: transformSystemMetrics(families, instanceId),
		workers: transformWorkerMetrics(families),
		transport: transformTransportMetrics(families, instanceId),
		statusCache: transformStatusCacheMetrics(families, instanceId),
		groups: transformHandlerMetrics(families, "groups"),
		communities: transformHandlerMetrics(families, "communities"),
		newsletters: transformHandlerMetrics(families, "newsletters"),
		instances,
	};
}

/**
 * Collect all unique instance IDs from metrics
 */
function collectInstanceIds(families: ParsedMetrics["families"]): string[] {
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
	_instanceId?: string | null, // eslint-disable-line @typescript-eslint/no-unused-vars
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
			if (
				sample.name.endsWith("_total") ||
				sample.name === "http_requests_total"
			) {
				const value = sample.value;
				totalRequests += value;

				// Count by status
				const status = sample.labels.status || "unknown";
				if (!metrics.byStatus) metrics.byStatus = {};
				metrics.byStatus[status] =
					(metrics.byStatus[status] || 0) + value;

				// Count errors (4xx and 5xx)
				if (status.startsWith("4") || status.startsWith("5")) {
					errorRequests += value;
				}

				// Count by path
				const path = sample.labels.path || "/";
				if (!metrics.byPath) metrics.byPath = {};
				if (!metrics.byPath[path]) {
					metrics.byPath[path] = {
						count: 0,
						avgLatencyMs: 0,
						errorCount: 0,
					};
				}
				metrics.byPath[path].count += value;
				if (status.startsWith("4") || status.startsWith("5")) {
					metrics.byPath[path].errorCount += value;
				}

				// Count by method
				const method = sample.labels.method || "GET";
				if (!metrics.byMethod) metrics.byMethod = {};
				metrics.byMethod[method] =
					(metrics.byMethod[method] || 0) + value;
			}
		}

		metrics.totalRequests = totalRequests;
		metrics.errorRate =
			totalRequests > 0 ? (errorRequests / totalRequests) * 100 : 0;
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
		dlqEvents: 0,
		dlqReprocessAttempts: 0,
		dlqReprocessSuccess: 0,
		sequenceGaps: 0,
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
			if (!metrics.byType) metrics.byType = {};
			if (!metrics.byType[eventType]) {
				metrics.byType[eventType] = {
					captured: 0,
					buffered: 0,
					inserted: 0,
					processed: 0,
					delivered: 0,
					failed: 0,
				};
			}
			metrics.byType[eventType].captured += sample.value;

			// By instance
			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance) metrics.byInstance = {};
				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = {
						captured: 0,
						buffered: 0,
						inserted: 0,
						processed: 0,
						delivered: 0,
						failed: 0,
						backlog: 0,
						sequenceGaps: 0,
					};
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
			if (!metrics.byType) metrics.byType = {};

			if (!metrics.byType[eventType]) {
				metrics.byType[eventType] = {
					captured: 0,
					buffered: 0,
					inserted: 0,
					processed: 0,
					delivered: 0,
					failed: 0,
				};
			}
			metrics.byType[eventType].processed += sample.value;

			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance) metrics.byInstance = {};

				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = {
						captured: 0,
						buffered: 0,
						inserted: 0,
						processed: 0,
						delivered: 0,
						failed: 0,
						backlog: 0,
						sequenceGaps: 0,
					};
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
			if (!metrics.byType) metrics.byType = {};

			if (!metrics.byType[eventType]) {
				metrics.byType[eventType] = {
					captured: 0,
					buffered: 0,
					inserted: 0,
					processed: 0,
					delivered: 0,
					failed: 0,
				};
			}
			metrics.byType[eventType].delivered += sample.value;

			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance) metrics.byInstance = {};

				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = {
						captured: 0,
						buffered: 0,
						inserted: 0,
						processed: 0,
						delivered: 0,
						failed: 0,
						backlog: 0,
						sequenceGaps: 0,
					};
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
			if (!metrics.byType) metrics.byType = {};

			if (!metrics.byType[eventType]) {
				metrics.byType[eventType] = {
					captured: 0,
					buffered: 0,
					inserted: 0,
					processed: 0,
					delivered: 0,
					failed: 0,
				};
			}
			metrics.byType[eventType].failed += sample.value;

			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance) metrics.byInstance = {};

				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = {
						captured: 0,
						buffered: 0,
						inserted: 0,
						processed: 0,
						delivered: 0,
						failed: 0,
						backlog: 0,
						sequenceGaps: 0,
					};
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
			if (!metrics.byInstance) metrics.byInstance = {};
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

	// dlq_events_total
	const dlqEventsFamily = getFamily(families, "dlq_events_total");
	if (dlqEventsFamily) {
		for (const sample of dlqEventsFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.dlqEvents += sample.value;
		}
	}

	// dlq_reprocess_attempts_total
	const dlqReprocessFamily = getFamily(
		families,
		"dlq_reprocess_attempts_total",
	);
	if (dlqReprocessFamily) {
		for (const sample of dlqReprocessFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.dlqReprocessAttempts += sample.value;
		}
	}

	// dlq_reprocess_success_total
	const dlqSuccessFamily = getFamily(families, "dlq_reprocess_success_total");
	if (dlqSuccessFamily) {
		for (const sample of dlqSuccessFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.dlqReprocessSuccess += sample.value;
		}
	}

	// event_sequence_gaps (gauge)
	const sequenceGapsFamily = getFamily(families, "event_sequence_gaps");
	if (sequenceGapsFamily) {
		for (const sample of sequenceGapsFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.sequenceGaps += sample.value;

			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance) metrics.byInstance = {};

				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = {
						captured: 0,
						buffered: 0,
						inserted: 0,
						processed: 0,
						delivered: 0,
						failed: 0,
						backlog: 0,
						sequenceGaps: 0,
					};
				}
				metrics.byInstance[instId].sequenceGaps = sample.value;
			}
		}
	}

	// event_processing_duration_seconds (histogram)
	const processingDurationFamily = getFamily(
		families,
		"event_processing_duration_seconds",
	);
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
	const deliveryDurationFamily = getFamily(
		families,
		"event_delivery_duration_seconds",
	);
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
				if (!metrics.byInstance) metrics.byInstance = {};

				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = {
						size: 0,
						workers: 0,
						pending: 0,
						processing: 0,
						sent: 0,
						failed: 0,
					};
				}
				metrics.byInstance[instId].size += sample.value;
				if (status === "pending")
					metrics.byInstance[instId].pending += sample.value;
				if (status === "processing")
					metrics.byInstance[instId].processing += sample.value;
				if (status === "sent")
					metrics.byInstance[instId].sent += sample.value;
				if (status === "failed")
					metrics.byInstance[instId].failed += sample.value;
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
			if (!metrics.byType) metrics.byType = {};

			if (!metrics.byType[msgType]) {
				metrics.byType[msgType] = {
					enqueued: 0,
					processed: 0,
					failed: 0,
				};
			}
			metrics.byType[msgType].enqueued += sample.value;
		}
	}

	// message_queue_processed_total
	const processedFamily = getFamily(
		families,
		"message_queue_processed_total",
	);
	if (processedFamily) {
		for (const sample of processedFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.processed += sample.value;

			const msgType = sample.labels.message_type || "unknown";
			if (!metrics.byType) metrics.byType = {};

			if (!metrics.byType[msgType]) {
				metrics.byType[msgType] = {
					enqueued: 0,
					processed: 0,
					failed: 0,
				};
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
			if (!metrics.byInstance) metrics.byInstance = {};
			if (instId && metrics.byInstance[instId]) {
				metrics.byInstance[instId].workers = sample.value;
			}
		}
	}

	// message_queue_processing_duration_seconds (histogram)
	const durationFamily = getFamily(
		families,
		"message_queue_processing_duration_seconds",
	);
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
		avgProcessingMs: 0,
		totalDownloadBytes: 0,
		totalUploadBytes: 0,
		localStorageBytes: 0,
		localStorageFiles: 0,
		s3StorageBytes: 0,
		s3StorageFiles: 0,
		backlog: 0,
		cleanupRuns: 0,
		filesDeleted: 0,
		cleanupDeletedBytes: 0,
		// New detailed metrics
		downloadErrors: 0,
		uploadAttempts: 0,
		uploadErrors: 0,
		uploadSizeBytes: 0,
		presignedUrlGenerated: 0,
		deleteAttempts: 0,
		failures: 0,
		fallbackAttempts: 0,
		fallbackSuccess: 0,
		fallbackFailure: 0,
		cleanupTotal: 0,
		cleanupErrors: 0,
		avgCleanupDurationMs: 0,
		serveRequests: 0,
		serveBytes: 0,
		byType: {},
		byInstance: {},
		byErrorType: {},
		byFallbackType: {},
		byStorageType: {},
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
			if (!metrics.byType) metrics.byType = {};

			if (!metrics.byType[mediaType]) {
				metrics.byType[mediaType] = {
					downloads: 0,
					uploads: 0,
					failures: 0,
				};
			}
			metrics.byType[mediaType].downloads += sample.value;
			if (status === "failure") {
				metrics.byType[mediaType].failures += sample.value;
			}

			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance) metrics.byInstance = {};

				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = {
						downloads: 0,
						uploads: 0,
						failures: 0,
					};
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
			if (!metrics.byType) metrics.byType = {};

			if (!metrics.byType[mediaType]) {
				metrics.byType[mediaType] = {
					downloads: 0,
					uploads: 0,
					failures: 0,
				};
			}
			metrics.byType[mediaType].uploads += sample.value;

			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance) metrics.byInstance = {};

				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = {
						downloads: 0,
						uploads: 0,
						failures: 0,
					};
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
	const cleanupBytesFamily = getFamily(
		families,
		"media_cleanup_deleted_bytes_total",
	);
	if (cleanupBytesFamily) {
		for (const sample of cleanupBytesFamily.samples) {
			metrics.cleanupDeletedBytes += sample.value;
		}
	}

	// media_download_duration_seconds (histogram)
	const downloadDurationFamily = getFamily(
		families,
		"media_download_duration_seconds",
	);
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
	const uploadDurationFamily = getFamily(
		families,
		"media_upload_duration_seconds",
	);
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

	// media_download_size_bytes (histogram)
	const downloadSizeFamily = getFamily(families, "media_download_size_bytes");
	if (downloadSizeFamily) {
		const histograms = extractHistograms(downloadSizeFamily);
		for (const h of histograms) {
			if (instanceId && h.labels.instance_id !== instanceId) continue;
			metrics.totalDownloadBytes += h.sum;
		}
	}

	// media_download_errors_total
	const downloadErrorsFamily = getFamily(
		families,
		"media_download_errors_total",
	);
	if (downloadErrorsFamily) {
		for (const sample of downloadErrorsFamily.samples) {
			metrics.downloadErrors += sample.value;
			const errorType = sample.labels.error_type || "unknown";
			if (!metrics.byErrorType) metrics.byErrorType = {};
			metrics.byErrorType[errorType] =
				(metrics.byErrorType[errorType] || 0) + sample.value;
		}
	}

	// media_upload_attempts_total
	const uploadAttemptsFamily = getFamily(
		families,
		"media_upload_attempts_total",
	);
	if (uploadAttemptsFamily) {
		for (const sample of uploadAttemptsFamily.samples) {
			metrics.uploadAttempts += sample.value;
		}
	}

	// media_upload_errors_total
	const uploadErrorsFamily = getFamily(families, "media_upload_errors_total");
	if (uploadErrorsFamily) {
		for (const sample of uploadErrorsFamily.samples) {
			metrics.uploadErrors += sample.value;
			const errorType = sample.labels.error_type || "unknown";
			if (!metrics.byErrorType) metrics.byErrorType = {};
			metrics.byErrorType[errorType] =
				(metrics.byErrorType[errorType] || 0) + sample.value;
		}
	}

	// media_upload_size_bytes_total
	const uploadSizeFamily = getFamily(
		families,
		"media_upload_size_bytes_total",
	);
	if (uploadSizeFamily) {
		for (const sample of uploadSizeFamily.samples) {
			metrics.uploadSizeBytes += sample.value;
			metrics.totalUploadBytes += sample.value;
		}
	}

	// media_presigned_url_generated_total
	const presignedFamily = getFamily(
		families,
		"media_presigned_url_generated_total",
	);
	if (presignedFamily) {
		for (const sample of presignedFamily.samples) {
			metrics.presignedUrlGenerated += sample.value;
		}
	}

	// media_delete_attempts_total
	const deleteAttemptsFamily = getFamily(
		families,
		"media_delete_attempts_total",
	);
	if (deleteAttemptsFamily) {
		for (const sample of deleteAttemptsFamily.samples) {
			metrics.deleteAttempts += sample.value;
		}
	}

	// media_failures_total
	const failuresFamily = getFamily(families, "media_failures_total");
	if (failuresFamily) {
		for (const sample of failuresFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.failures += sample.value;
		}
	}

	// media_fallback_attempts_total
	const fallbackAttemptsFamily = getFamily(
		families,
		"media_fallback_attempts_total",
	);
	if (fallbackAttemptsFamily) {
		for (const sample of fallbackAttemptsFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.fallbackAttempts += sample.value;
			const fallbackType = sample.labels.fallback_type || "unknown";
			if (!metrics.byFallbackType) metrics.byFallbackType = {};
			metrics.byFallbackType[fallbackType] =
				(metrics.byFallbackType[fallbackType] || 0) + sample.value;
		}
	}

	// media_fallback_success_total
	const fallbackSuccessFamily = getFamily(
		families,
		"media_fallback_success_total",
	);
	if (fallbackSuccessFamily) {
		for (const sample of fallbackSuccessFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.fallbackSuccess += sample.value;
			const storageType = sample.labels.storage_type || "unknown";
			if (!metrics.byStorageType) metrics.byStorageType = {};
			metrics.byStorageType[storageType] =
				(metrics.byStorageType[storageType] || 0) + sample.value;
		}
	}

	// media_fallback_failure_total
	const fallbackFailureFamily = getFamily(
		families,
		"media_fallback_failure_total",
	);
	if (fallbackFailureFamily) {
		for (const sample of fallbackFailureFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.fallbackFailure += sample.value;
		}
	}

	// media_cleanup_total
	const cleanupTotalFamily = getFamily(families, "media_cleanup_total");
	if (cleanupTotalFamily) {
		for (const sample of cleanupTotalFamily.samples) {
			metrics.cleanupTotal += sample.value;
		}
	}

	// media_cleanup_errors_total
	const cleanupErrorsFamily = getFamily(
		families,
		"media_cleanup_errors_total",
	);
	if (cleanupErrorsFamily) {
		for (const sample of cleanupErrorsFamily.samples) {
			metrics.cleanupErrors += sample.value;
		}
	}

	// media_cleanup_duration_seconds (histogram)
	const cleanupDurationFamily = getFamily(
		families,
		"media_cleanup_duration_seconds",
	);
	if (cleanupDurationFamily) {
		const histograms = extractHistograms(cleanupDurationFamily);
		let totalSum = 0;
		let totalCount = 0;

		for (const h of histograms) {
			totalSum += h.sum;
			totalCount += h.count;
		}

		if (totalCount > 0) {
			metrics.avgCleanupDurationMs = (totalSum / totalCount) * 1000;
		}
	}

	// media_serve_requests_total
	const serveRequestsFamily = getFamily(
		families,
		"media_serve_requests_total",
	);
	if (serveRequestsFamily) {
		for (const sample of serveRequestsFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.serveRequests += sample.value;
		}
	}

	// media_serve_bytes_total
	const serveBytesFamily = getFamily(families, "media_serve_bytes_total");
	if (serveBytesFamily) {
		for (const sample of serveBytesFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;
			metrics.serveBytes += sample.value;
		}
	}

	return metrics;
}

/**
 * Transform system health metrics
 */

function transformSystemMetrics(
	families: ParsedMetrics["families"],
	_instanceId?: string | null, // eslint-disable-line @typescript-eslint/no-unused-vars
): SystemMetrics {
	const metrics: SystemMetrics = {
		circuitBreakerState: "unknown",
		circuitBreakerByInstance: {},
		circuitBreakerTransitions: 0,
		lockAcquisitions: {
			success: 0,
			failure: 0,
			reacquisitions: 0,
			fallbacks: 0,
		},
		splitBrainDetected: 0,
		splitBrainInvalidLocks: 0,
		healthChecks: {},
		orphanedInstances: 0,
		reconciliation: {
			success: 0,
			failure: 0,
			skipped: 0,
			error: 0,
			avgDurationMs: 0,
		},
		queueDrain: {
			durationMs: 0,
			timeouts: 0,
			drainedMessages: 0,
		},
	};

	// circuit_breaker_state (gauge)
	const circuitFamily = getFamily(families, "circuit_breaker_state");
	if (circuitFamily) {
		for (const sample of circuitFamily.samples) {
			metrics.circuitBreakerState = mapCircuitState(sample.value);
		}
	}

	// circuit_breaker_state_per_instance (gauge)
	const circuitPerInstanceFamily = getFamily(
		families,
		"circuit_breaker_state_per_instance",
	);
	if (circuitPerInstanceFamily) {
		for (const sample of circuitPerInstanceFamily.samples) {
			const instId = sample.labels.instance_id;
			if (instId) {
				metrics.circuitBreakerByInstance[instId] = mapCircuitState(
					sample.value,
				);
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
	const reacqFamily = getFamily(
		families,
		"lock_reacquisition_attempts_total",
	);
	if (reacqFamily) {
		for (const sample of reacqFamily.samples) {
			metrics.lockAcquisitions.reacquisitions += sample.value;
		}
	}

	// lock_reacquisition_fallbacks_total
	const fallbacksFamily = getFamily(
		families,
		"lock_reacquisition_fallbacks_total",
	);
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
				metrics.healthChecks[component] = {
					healthy: 0,
					unhealthy: 0,
					degraded: 0,
				};
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
	const reconDurationFamily = getFamily(
		families,
		"reconciliation_duration_seconds",
	);
	if (reconDurationFamily) {
		const histograms = extractHistograms(reconDurationFamily);
		let totalSum = 0;
		let totalCount = 0;

		for (const h of histograms) {
			totalSum += h.sum;
			totalCount += h.count;
		}

		if (totalCount > 0) {
			metrics.reconciliation.avgDurationMs =
				(totalSum / totalCount) * 1000;
		}
	}

	// circuit_breaker_transitions_total
	const circuitTransitionsFamily = getFamily(
		families,
		"circuit_breaker_transitions_total",
	);
	if (circuitTransitionsFamily) {
		for (const sample of circuitTransitionsFamily.samples) {
			metrics.circuitBreakerTransitions += sample.value;
		}
	}

	// split_brain_invalid_locks_total
	const invalidLocksFamily = getFamily(
		families,
		"split_brain_invalid_locks_total",
	);
	if (invalidLocksFamily) {
		for (const sample of invalidLocksFamily.samples) {
			metrics.splitBrainInvalidLocks += sample.value;
		}
	}

	// queue_drain_duration_seconds (histogram)
	const queueDrainDurationFamily = getFamily(
		families,
		"queue_drain_duration_seconds",
	);
	if (queueDrainDurationFamily) {
		const histograms = extractHistograms(queueDrainDurationFamily);
		let totalSum = 0;
		let totalCount = 0;

		for (const h of histograms) {
			totalSum += h.sum;
			totalCount += h.count;
		}

		if (totalCount > 0) {
			metrics.queueDrain.durationMs = (totalSum / totalCount) * 1000;
		}
	}

	// queue_drain_timeouts_total
	const queueDrainTimeoutsFamily = getFamily(
		families,
		"queue_drain_timeouts_total",
	);
	if (queueDrainTimeoutsFamily) {
		for (const sample of queueDrainTimeoutsFamily.samples) {
			metrics.queueDrain.timeouts += sample.value;
		}
	}

	// queue_drained_messages_total
	const queueDrainedMessagesFamily = getFamily(
		families,
		"queue_drained_messages_total",
	);
	if (queueDrainedMessagesFamily) {
		for (const sample of queueDrainedMessagesFamily.samples) {
			metrics.queueDrain.drainedMessages += sample.value;
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
	};

	// workers_active (gauge)
	const activeFamily = getFamily(families, "workers_active");
	if (activeFamily) {
		for (const sample of activeFamily.samples) {
			const workerType = sample.labels.worker_type || "unknown";
			metrics.active[workerType] =
				(metrics.active[workerType] || 0) + sample.value;
			metrics.totalActive += sample.value;
		}
	}

	// worker_errors_total
	const errorsFamily = getFamily(families, "worker_errors_total");
	if (errorsFamily) {
		for (const sample of errorsFamily.samples) {
			const workerType = sample.labels.worker_type || "unknown";
			metrics.errors[workerType] =
				(metrics.errors[workerType] || 0) + sample.value;
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
					metrics.avgTaskDurationMs[workerType] =
						(metrics.avgTaskDurationMs[workerType] + avgMs) / 2;
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
		totalMessages: 0,
		totalDeliveries: 0,
		totalRetries: 0,
		sent: 0,
		received: 0,
		failed: 0,
		errors: 0,
		avgLatencyMs: 0,
		p50LatencyMs: 0,
		p95LatencyMs: 0,
		p99LatencyMs: 0,
		successRate: 0,
		avgDurationMs: 0,
		successfulDeliveries: 0,
		failedDeliveries: 0,
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
				if (!metrics.byInstance) metrics.byInstance = {};

				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = {
						sent: 0,
						received: 0,
						failed: 0,
						errors: 0,
						avgLatencyMs: 0,
						deliveries: 0,
						success: 0,
						retries: 0,
						errorRate: 0,
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
			metrics.byErrorType[errorType] =
				(metrics.byErrorType[errorType] || 0) + sample.value;
		}
	}

	// transport_retries_total
	const retriesFamily = getFamily(families, "transport_retries_total");
	if (retriesFamily) {
		for (const sample of retriesFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;

			metrics.totalRetries += sample.value;

			const instId = sample.labels.instance_id;
			if (!metrics.byInstance) metrics.byInstance = {};
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
				metrics.byInstance[instId].avgDurationMs =
					(h.sum / h.count) * 1000;
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
		metrics.successRate =
			(metrics.successfulDeliveries / metrics.totalDeliveries) * 100;
	}

	return metrics;
}

/**
 * Transform status cache metrics
 * Handles message status caching for read, delivered, played, sent events
 */
function transformStatusCacheMetrics(
	families: ParsedMetrics["families"],
	instanceId?: string | null,
): StatusCacheMetrics {
	const metrics: StatusCacheMetrics = {
		totalOperations: 0,
		totalSize: 0,
		totalHits: 0,
		totalMisses: 0,
		hitRate: 0,
		totalSuppressions: 0,
		totalFlushed: 0,
		avgDurationMs: 0,
		p50DurationMs: 0,
		p95DurationMs: 0,
		p99DurationMs: 0,
		byOperation: {},
		byInstance: {},
		byStatusType: {},
		byTrigger: {},
	};

	const shouldInclude = (labels: Record<string, string>) => {
		if (!instanceId) return true;
		return labels.instance_id === instanceId;
	};

	// status_cache_operations_total (labels: instance_id, operation, status)
	const operationsFamily = getFamily(
		families,
		"status_cache_operations_total",
	);
	if (operationsFamily) {
		for (const sample of operationsFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;

			const operation = sample.labels.operation || "unknown";
			const status = sample.labels.status || "unknown";
			const instId = sample.labels.instance_id;

			metrics.totalOperations += sample.value;

			// By operation
			if (!metrics.byOperation) metrics.byOperation = {};

			if (!metrics.byOperation[operation]) {
				metrics.byOperation[operation] = {
					count: 0,
					success: 0,
					failed: 0,
				};
			}
			metrics.byOperation[operation].count += sample.value;
			if (status === "success") {
				metrics.byOperation[operation].success += sample.value;
			} else if (status === "failed" || status === "error") {
				metrics.byOperation[operation].failed += sample.value;
			}

			// By instance
			if (instId) {
				if (!metrics.byInstance) metrics.byInstance = {};

				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = {
						size: 0,
						operations: 0,
						hits: 0,
						misses: 0,
						hitRate: 0,
						suppressions: 0,
						flushed: 0,
					};
				}
				metrics.byInstance[instId].operations += sample.value;
			}
		}
	}

	// status_cache_size (gauge, labels: instance_id)
	const sizeFamily = getFamily(families, "status_cache_size");
	if (sizeFamily) {
		for (const sample of sizeFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;

			metrics.totalSize += sample.value;

			const instId = sample.labels.instance_id;
			if (instId) {
				if (!metrics.byInstance) metrics.byInstance = {};

				if (!metrics.byInstance[instId]) {
					metrics.byInstance[instId] = {
						size: 0,
						operations: 0,
						hits: 0,
						misses: 0,
						hitRate: 0,
						suppressions: 0,
						flushed: 0,
					};
				}
				metrics.byInstance[instId].size = sample.value;
			}
		}
	}

	// status_cache_hits_total (labels: instance_id)
	const hitsFamily = getFamily(families, "status_cache_hits_total");
	if (hitsFamily) {
		for (const sample of hitsFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;

			metrics.totalHits += sample.value;

			const instId = sample.labels.instance_id;
			if (!metrics.byInstance) metrics.byInstance = {};
			if (instId && metrics.byInstance[instId]) {
				metrics.byInstance[instId].hits += sample.value;
			}
		}
	}

	// status_cache_misses_total (labels: instance_id)
	const missesFamily = getFamily(families, "status_cache_misses_total");
	if (missesFamily) {
		for (const sample of missesFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;

			metrics.totalMisses += sample.value;

			const instId = sample.labels.instance_id;
			if (!metrics.byInstance) metrics.byInstance = {};
			if (instId && metrics.byInstance[instId]) {
				metrics.byInstance[instId].misses += sample.value;
			}
		}
	}

	// Calculate hit rates
	const totalLookups = metrics.totalHits + metrics.totalMisses;
	if (totalLookups > 0) {
		metrics.hitRate = (metrics.totalHits / totalLookups) * 100;
	}

	// Calculate per-instance hit rates
	for (const instId of Object.keys(metrics.byInstance)) {
		const inst = metrics.byInstance[instId];
		const instLookups = inst.hits + inst.misses;
		if (instLookups > 0) {
			inst.hitRate = (inst.hits / instLookups) * 100;
		}
	}

	// status_cache_suppressions_total (labels: instance_id, status_type)
	const suppressionsFamily = getFamily(
		families,
		"status_cache_suppressions_total",
	);
	if (suppressionsFamily) {
		for (const sample of suppressionsFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;

			const statusType = sample.labels.status_type || "unknown";
			const instId = sample.labels.instance_id;

			metrics.totalSuppressions += sample.value;

			// By status type (read, delivered, played, sent)
			if (!metrics.byStatusType) metrics.byStatusType = {};
			metrics.byStatusType[statusType] =
				(metrics.byStatusType[statusType] || 0) + sample.value;

			// By instance
			if (!metrics.byInstance) metrics.byInstance = {};
			if (instId && metrics.byInstance[instId]) {
				metrics.byInstance[instId].suppressions += sample.value;
			}
		}
	}

	// status_cache_flushed_total (labels: instance_id, trigger)
	const flushedFamily = getFamily(families, "status_cache_flushed_total");
	if (flushedFamily) {
		for (const sample of flushedFamily.samples) {
			if (!shouldInclude(sample.labels)) continue;

			const trigger = sample.labels.trigger || "unknown";
			const instId = sample.labels.instance_id;

			metrics.totalFlushed += sample.value;

			// By trigger (manual, ttl, shutdown)
			if (!metrics.byTrigger) metrics.byTrigger = {};
			metrics.byTrigger[trigger] =
				(metrics.byTrigger[trigger] || 0) + sample.value;

			// By instance
			if (!metrics.byInstance) metrics.byInstance = {};
			if (instId && metrics.byInstance[instId]) {
				metrics.byInstance[instId].flushed += sample.value;
			}
		}
	}

	// status_cache_operation_duration_seconds (histogram, labels: instance_id, operation)
	const durationFamily = getFamily(
		families,
		"status_cache_operation_duration_seconds",
	);
	if (durationFamily) {
		const histograms = extractHistograms(durationFamily);
		let totalSum = 0;
		let totalCount = 0;
		let p50Sum = 0;
		let p95Sum = 0;
		let p99Sum = 0;
		let validHistograms = 0;

		// Track per-operation durations
		const operationDurations: Record<
			string,
			{ sum: number; count: number }
		> = {};

		for (const h of histograms) {
			if (instanceId && h.labels.instance_id !== instanceId) continue;

			const operation = h.labels.operation || "unknown";

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

			// Track per-operation
			if (!operationDurations[operation]) {
				operationDurations[operation] = { sum: 0, count: 0 };
			}
			operationDurations[operation].sum += h.sum;
			operationDurations[operation].count += h.count;
		}

		if (totalCount > 0) {
			metrics.avgDurationMs = (totalSum / totalCount) * 1000;
		}
		if (validHistograms > 0) {
			metrics.p50DurationMs = (p50Sum / validHistograms) * 1000;
			metrics.p95DurationMs = (p95Sum / validHistograms) * 1000;
			metrics.p99DurationMs = (p99Sum / validHistograms) * 1000;
		}

		// Note: Per-operation avgDurationMs is not part of the byOperation interface
		// It's calculated at the top level only
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

/**
 * Transform handler metrics for groups, communities, newsletters
 */
function transformHandlerMetrics(
	families: ParsedMetrics["families"],
	handlerType: "groups" | "communities" | "newsletters",
): HandlerMetrics {
	const metrics: HandlerMetrics = {
		totalRequests: 0,
		successRequests: 0,
		failedRequests: 0,
		avgLatencyMs: 0,
		byOperation: {},
	};

	// {handler}_requests_total (labels: operation, status)
	const requestsFamily = getFamily(families, `${handlerType}_requests_total`);
	if (requestsFamily) {
		for (const sample of requestsFamily.samples) {
			const operation = sample.labels.operation || "unknown";
			const status = sample.labels.status || "unknown";

			metrics.totalRequests += sample.value;

			if (status === "success") {
				metrics.successRequests += sample.value;
			} else if (status === "failure" || status === "error") {
				metrics.failedRequests += sample.value;
			}

			// By operation
			if (!metrics.byOperation) metrics.byOperation = {};

			if (!metrics.byOperation[operation]) {
				metrics.byOperation[operation] = {
					total: 0,
					success: 0,
					failed: 0,
					avgLatencyMs: 0,
				};
			}
			metrics.byOperation[operation].total += sample.value;
			if (status === "success") {
				metrics.byOperation[operation].success += sample.value;
			} else if (status === "failure" || status === "error") {
				metrics.byOperation[operation].failed += sample.value;
			}
		}
	}

	// {handler}_latency_seconds (histogram, labels: operation)
	const latencyFamily = getFamily(families, `${handlerType}_latency_seconds`);
	if (latencyFamily) {
		const histograms = extractHistograms(latencyFamily);
		let totalSum = 0;
		let totalCount = 0;

		// Track per-operation latencies
		const operationLatencies: Record<
			string,
			{ sum: number; count: number }
		> = {};

		for (const h of histograms) {
			const operation = h.labels.operation || "unknown";

			totalSum += h.sum;
			totalCount += h.count;

			// Track per-operation
			if (!operationLatencies[operation]) {
				operationLatencies[operation] = { sum: 0, count: 0 };
			}
			operationLatencies[operation].sum += h.sum;
			operationLatencies[operation].count += h.count;
		}

		if (totalCount > 0) {
			metrics.avgLatencyMs = (totalSum / totalCount) * 1000;
		}

		// Update per-operation avgLatencyMs
		for (const [operation, data] of Object.entries(operationLatencies)) {
			if (
				metrics.byOperation &&
				metrics.byOperation[operation] &&
				data.count > 0
			) {
				metrics.byOperation[operation].avgLatencyMs =
					(data.sum / data.count) * 1000;
			}
		}
	}

	return metrics;
}
