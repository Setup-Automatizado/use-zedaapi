/**
 * Metrics Overview Component
 *
 * Grid of KPI cards showing key metrics at a glance.
 *
 * @module components/metrics/metrics-overview
 */

import {
	Activity,
	AlertTriangle,
	Clock,
	Image,
	ListOrdered,
	Server,
	Users,
	Zap,
} from "lucide-react";
import {
	formatDuration,
	formatNumber,
	formatPercentage,
	getHealthLevel,
	METRIC_THRESHOLDS,
} from "@/lib/metrics/constants";
import { cn } from "@/lib/utils";
import type { DashboardMetrics, HealthLevel } from "@/types/metrics";
import { MetricKPICard } from "./metric-kpi-card";

export interface MetricsOverviewProps {
	/** Dashboard metrics data */
	metrics?: DashboardMetrics;
	/** Loading state */
	isLoading?: boolean;
	/** Additional CSS classes */
	className?: string;
}

export function MetricsOverview({
	metrics,
	isLoading = false,
	className,
}: MetricsOverviewProps) {
	// Calculate derived metrics
	const requestsPerSec = metrics?.http.requestsPerSecond ?? 0;
	const eventsPerSec = metrics ? metrics.events.delivered / 60 : 0; // Approximate rate
	const queueSize = metrics?.messageQueue.pending ?? 0;
	const errorRate = metrics?.http.errorRate ?? 0;
	const p95Latency = metrics?.http.p95LatencyMs ?? 0;
	const activeWorkers = metrics?.workers.totalActive ?? 0;
	const mediaBacklog = metrics?.media.backlog ?? 0;
	const circuitState = metrics?.system.circuitBreakerState ?? "unknown";

	// Determine health levels
	const errorStatus = getHealthLevel(errorRate, METRIC_THRESHOLDS.errorRate);
	const latencyStatus = getHealthLevel(p95Latency, METRIC_THRESHOLDS.p95LatencyMs);
	const queueStatus = getHealthLevel(queueSize, METRIC_THRESHOLDS.queueBacklog);
	const mediaStatus = getHealthLevel(mediaBacklog, METRIC_THRESHOLDS.mediaBacklog);
	const workerStatus = getHealthLevel(activeWorkers, METRIC_THRESHOLDS.activeWorkers);
	const circuitStatus = getCircuitBreakerHealthLevel(circuitState);

	return (
		<div className={cn("grid grid-cols-2 gap-4 lg:grid-cols-4", className)}>
			{/* Requests/sec */}
			<MetricKPICard
				title="Requests/sec"
				value={formatNumber(requestsPerSec, 1)}
				icon={Activity}
				status="healthy"
				subtitle="HTTP requests"
				isLoading={isLoading}
			/>

			{/* Events/sec */}
			<MetricKPICard
				title="Events/sec"
				value={formatNumber(eventsPerSec, 1)}
				icon={Zap}
				status="healthy"
				subtitle="Delivered events"
				isLoading={isLoading}
			/>

			{/* Queue Size */}
			<MetricKPICard
				title="Queue Size"
				value={formatNumber(queueSize, 0)}
				icon={ListOrdered}
				status={queueStatus}
				subtitle="Pending messages"
				isLoading={isLoading}
			/>

			{/* Error Rate */}
			<MetricKPICard
				title="Error Rate"
				value={formatPercentage(errorRate, 2)}
				icon={AlertTriangle}
				status={errorStatus}
				subtitle="HTTP errors"
				isLoading={isLoading}
			/>

			{/* P95 Latency */}
			<MetricKPICard
				title="P95 Latency"
				value={formatDuration(p95Latency)}
				icon={Clock}
				status={latencyStatus}
				subtitle="Request latency"
				isLoading={isLoading}
			/>

			{/* Active Workers */}
			<MetricKPICard
				title="Workers"
				value={activeWorkers}
				icon={Users}
				status={workerStatus}
				subtitle="Active workers"
				isLoading={isLoading}
			/>

			{/* Media Backlog */}
			<MetricKPICard
				title="Media Backlog"
				value={formatNumber(mediaBacklog, 0)}
				icon={Image}
				status={mediaStatus}
				subtitle="Pending uploads"
				isLoading={isLoading}
			/>

			{/* Circuit Breaker */}
			<MetricKPICard
				title="Circuit Breaker"
				value={formatCircuitState(circuitState)}
				icon={Server}
				status={circuitStatus}
				subtitle="System protection"
				isLoading={isLoading}
			/>
		</div>
	);
}

/**
 * Get health level for circuit breaker state
 */
function getCircuitBreakerHealthLevel(state: string): HealthLevel {
	switch (state) {
		case "closed":
			return "healthy";
		case "half-open":
			return "warning";
		case "open":
			return "critical";
		default:
			return "warning";
	}
}

/**
 * Format circuit breaker state for display
 */
function formatCircuitState(state: string): string {
	switch (state) {
		case "closed":
			return "Closed";
		case "half-open":
			return "Half-Open";
		case "open":
			return "Open";
		default:
			return "Unknown";
	}
}
