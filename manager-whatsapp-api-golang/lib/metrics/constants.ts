/**
 * Metrics Constants
 *
 * Thresholds, colors, and configuration for metrics display.
 *
 * @module lib/metrics/constants
 */

import type {
	HealthLevel,
	MetricThreshold,
	RefreshInterval,
} from "@/types/metrics";

/**
 * Default refresh interval in milliseconds
 */
export const DEFAULT_REFRESH_INTERVAL: RefreshInterval = 15000;

/**
 * Available refresh interval options
 */
export const REFRESH_INTERVAL_OPTIONS: {
	value: RefreshInterval;
	label: string;
}[] = [
	{ value: 5000, label: "5 seconds" },
	{ value: 15000, label: "15 seconds" },
	{ value: 30000, label: "30 seconds" },
	{ value: 60000, label: "1 minute" },
	{ value: 0, label: "Off" },
];

/**
 * Metric thresholds for determining health status
 */
export const METRIC_THRESHOLDS: Record<string, MetricThreshold> = {
	// Error rates (percentage)
	errorRate: { warning: 1, critical: 5, unit: "%" },

	// Latency (milliseconds)
	p95LatencyMs: { warning: 500, critical: 2000, unit: "ms" },
	p99LatencyMs: { warning: 1000, critical: 5000, unit: "ms" },
	avgLatencyMs: { warning: 200, critical: 1000, unit: "ms" },

	// Queue sizes
	queueBacklog: { warning: 100, critical: 1000, unit: "" },
	dlqSize: { warning: 10, critical: 100, unit: "" },
	outboxBacklog: { warning: 50, critical: 500, unit: "" },

	// Media
	mediaBacklog: { warning: 50, critical: 200, unit: "" },

	// Workers
	activeWorkers: { warning: 1, critical: 0, unit: "", inverse: true },

	// System
	splitBrainEvents: { warning: 1, critical: 5, unit: "" },
	orphanedInstances: { warning: 1, critical: 5, unit: "" },
	lockFailureRate: { warning: 5, critical: 20, unit: "%" },
};

/**
 * Get health level based on value and threshold
 */
export function getHealthLevel(
	value: number,
	threshold: MetricThreshold,
): HealthLevel {
	if (threshold.inverse) {
		// Lower is worse (e.g., active workers)
		if (value <= threshold.critical) return "critical";
		if (value <= threshold.warning) return "warning";
		return "healthy";
	}

	// Higher is worse (default)
	if (value >= threshold.critical) return "critical";
	if (value >= threshold.warning) return "warning";
	return "healthy";
}

/**
 * Color classes for health levels
 */
export const HEALTH_COLORS: Record<
	HealthLevel,
	{ text: string; bg: string; border: string }
> = {
	healthy: {
		text: "text-emerald-600 dark:text-emerald-400",
		bg: "bg-emerald-500/10",
		border: "border-emerald-500/20",
	},
	warning: {
		text: "text-amber-600 dark:text-amber-400",
		bg: "bg-amber-500/10",
		border: "border-amber-500/20",
	},
	critical: {
		text: "text-red-600 dark:text-red-400",
		bg: "bg-red-500/10",
		border: "border-red-500/20",
	},
};

/**
 * Status dot colors for health levels
 */
export const STATUS_DOT_COLORS: Record<HealthLevel, string> = {
	healthy: "bg-emerald-500",
	warning: "bg-amber-500",
	critical: "bg-red-500",
};

/**
 * Circuit breaker state colors
 */
export const CIRCUIT_BREAKER_COLORS: Record<
	string,
	{ text: string; bg: string }
> = {
	closed: {
		text: "text-emerald-600 dark:text-emerald-400",
		bg: "bg-emerald-500/10",
	},
	open: {
		text: "text-red-600 dark:text-red-400",
		bg: "bg-red-500/10",
	},
	"half-open": {
		text: "text-amber-600 dark:text-amber-400",
		bg: "bg-amber-500/10",
	},
	unknown: {
		text: "text-gray-600 dark:text-gray-400",
		bg: "bg-gray-500/10",
	},
};

/**
 * Chart color palette (CSS variables for theme support)
 */
export const CHART_COLORS = {
	primary: "hsl(var(--chart-1))",
	secondary: "hsl(var(--chart-2))",
	tertiary: "hsl(var(--chart-3))",
	quaternary: "hsl(var(--chart-4))",
	quinary: "hsl(var(--chart-5))",
	success: "hsl(142, 76%, 36%)",
	warning: "hsl(38, 92%, 50%)",
	error: "hsl(0, 84%, 60%)",
	muted: "hsl(var(--muted-foreground))",
};

/**
 * Tailwind chart colors for direct use
 */
export const TAILWIND_CHART_COLORS = {
	primary: "#3b82f6", // blue-500
	secondary: "#8b5cf6", // violet-500
	tertiary: "#06b6d4", // cyan-500
	quaternary: "#f59e0b", // amber-500
	quinary: "#ec4899", // pink-500
	success: "#10b981", // emerald-500
	warning: "#f59e0b", // amber-500
	error: "#ef4444", // red-500
	muted: "#6b7280", // gray-500
};

/**
 * Format a number with appropriate suffix (K, M, B)
 */
export function formatNumber(value: number, decimals = 1): string {
	if (!Number.isFinite(value)) return "N/A";

	if (Math.abs(value) >= 1_000_000_000) {
		return `${(value / 1_000_000_000).toFixed(decimals)}B`;
	}
	if (Math.abs(value) >= 1_000_000) {
		return `${(value / 1_000_000).toFixed(decimals)}M`;
	}
	if (Math.abs(value) >= 1_000) {
		return `${(value / 1_000).toFixed(decimals)}K`;
	}

	return value.toFixed(decimals);
}

/**
 * Format bytes to human-readable string
 */
export function formatBytes(bytes: number, decimals = 2): string {
	if (!Number.isFinite(bytes) || bytes === 0) return "0 B";

	const k = 1024;
	const dm = decimals < 0 ? 0 : decimals;
	const sizes = ["B", "KB", "MB", "GB", "TB", "PB"];

	const i = Math.floor(Math.log(bytes) / Math.log(k));

	return `${Number.parseFloat((bytes / k ** i).toFixed(dm))} ${sizes[i]}`;
}

/**
 * Format duration in milliseconds to human-readable string
 */
export function formatDuration(ms: number): string {
	if (!Number.isFinite(ms)) return "N/A";

	if (ms < 1) {
		return `${(ms * 1000).toFixed(0)}μs`;
	}
	if (ms < 1000) {
		return `${ms.toFixed(1)}ms`;
	}
	if (ms < 60000) {
		return `${(ms / 1000).toFixed(2)}s`;
	}
	if (ms < 3600000) {
		return `${(ms / 60000).toFixed(1)}m`;
	}

	return `${(ms / 3600000).toFixed(1)}h`;
}

/**
 * Format percentage
 */
export function formatPercentage(value: number, decimals = 1): string {
	if (!Number.isFinite(value)) return "N/A";
	return `${value.toFixed(decimals)}%`;
}

/**
 * Format rate per second
 */
export function formatRate(value: number, decimals = 1): string {
	if (!Number.isFinite(value)) return "N/A";
	return `${formatNumber(value, decimals)}/s`;
}

/**
 * Get relative time string (e.g., "5 seconds ago")
 */
export function getRelativeTime(date: Date): string {
	const now = new Date();
	const diffMs = now.getTime() - date.getTime();
	const diffSec = Math.floor(diffMs / 1000);

	if (diffSec < 5) return "just now";
	if (diffSec < 60) return `${diffSec} seconds ago`;
	if (diffSec < 120) return "1 minute ago";
	if (diffSec < 3600) return `${Math.floor(diffSec / 60)} minutes ago`;
	if (diffSec < 7200) return "1 hour ago";
	if (diffSec < 86400) return `${Math.floor(diffSec / 3600)} hours ago`;

	return `${Math.floor(diffSec / 86400)} days ago`;
}

/**
 * Metric category labels
 */
export const METRIC_CATEGORIES = {
	http: "HTTP",
	events: "Events",
	messageQueue: "Message Queue",
	media: "Media",
	system: "System",
	workers: "Workers",
} as const;

/**
 * Tab configuration for metrics dashboard
 */
export const METRICS_TABS = [
	{ id: "overview", label: "Overview", icon: "LayoutDashboard" },
	{ id: "http", label: "HTTP", icon: "Globe" },
	{ id: "events", label: "Events", icon: "Zap" },
	{ id: "queue", label: "Queue", icon: "ListOrdered" },
	{ id: "media", label: "Media", icon: "Image" },
	{ id: "system", label: "System", icon: "Server" },
] as const;

export type MetricsTabId = (typeof METRICS_TABS)[number]["id"];

/**
 * HTTP Status Code Labels
 * Maps status codes and status code groups to friendly labels
 */
export const HTTP_STATUS_LABELS: Record<string, { label: string; description: string; color: string }> = {
	// Status code groups (used in charts)
	"1xx": { label: "Info", description: "Informational responses", color: TAILWIND_CHART_COLORS.muted },
	"2xx": { label: "Success", description: "Successful requests", color: TAILWIND_CHART_COLORS.success },
	"3xx": { label: "Redirect", description: "Redirection responses", color: TAILWIND_CHART_COLORS.tertiary },
	"4xx": { label: "Client Error", description: "Client-side errors", color: TAILWIND_CHART_COLORS.warning },
	"5xx": { label: "Server Error", description: "Server-side errors", color: TAILWIND_CHART_COLORS.error },

	// Individual status codes
	"100": { label: "Continue", description: "Continue with the request", color: TAILWIND_CHART_COLORS.muted },
	"101": { label: "Switching", description: "Switching protocols", color: TAILWIND_CHART_COLORS.muted },
	"200": { label: "Success", description: "Request succeeded", color: TAILWIND_CHART_COLORS.success },
	"201": { label: "Created", description: "Resource created", color: TAILWIND_CHART_COLORS.success },
	"202": { label: "Accepted", description: "Request accepted", color: TAILWIND_CHART_COLORS.success },
	"204": { label: "No Content", description: "No content to return", color: TAILWIND_CHART_COLORS.success },
	"301": { label: "Moved", description: "Resource moved permanently", color: TAILWIND_CHART_COLORS.tertiary },
	"302": { label: "Found", description: "Resource found (redirect)", color: TAILWIND_CHART_COLORS.tertiary },
	"304": { label: "Not Modified", description: "Resource not modified", color: TAILWIND_CHART_COLORS.tertiary },
	"400": { label: "Bad Request", description: "Invalid request syntax", color: TAILWIND_CHART_COLORS.warning },
	"401": { label: "Unauthorized", description: "Authentication required", color: TAILWIND_CHART_COLORS.warning },
	"403": { label: "Forbidden", description: "Access denied", color: TAILWIND_CHART_COLORS.warning },
	"404": { label: "Not Found", description: "Resource not found", color: TAILWIND_CHART_COLORS.warning },
	"405": { label: "Method Not Allowed", description: "HTTP method not allowed", color: TAILWIND_CHART_COLORS.warning },
	"409": { label: "Conflict", description: "Request conflicts with current state", color: TAILWIND_CHART_COLORS.warning },
	"422": { label: "Unprocessable", description: "Validation failed", color: TAILWIND_CHART_COLORS.warning },
	"429": { label: "Rate Limited", description: "Too many requests", color: TAILWIND_CHART_COLORS.warning },
	"500": { label: "Server Error", description: "Internal server error", color: TAILWIND_CHART_COLORS.error },
	"501": { label: "Not Implemented", description: "Feature not implemented", color: TAILWIND_CHART_COLORS.error },
	"502": { label: "Bad Gateway", description: "Invalid response from upstream", color: TAILWIND_CHART_COLORS.error },
	"503": { label: "Unavailable", description: "Service unavailable", color: TAILWIND_CHART_COLORS.error },
	"504": { label: "Gateway Timeout", description: "Upstream timeout", color: TAILWIND_CHART_COLORS.error },
};

/**
 * Get friendly label for HTTP status code
 */
export function getHttpStatusLabel(status: string | number): string {
	const statusStr = String(status);

	// Check for exact match first
	const exactMatch = HTTP_STATUS_LABELS[statusStr];
	if (exactMatch) {
		return exactMatch.label;
	}

	// Derive group from numeric status
	if (/^\d{3}$/.test(statusStr)) {
		const group = `${statusStr[0]}xx`;
		const groupMatch = HTTP_STATUS_LABELS[group];
		if (groupMatch) {
			return `${statusStr} ${groupMatch.label}`;
		}
	}

	// Check for status group directly (e.g., "2xx")
	if (statusStr.endsWith("xx")) {
		const groupMatch = HTTP_STATUS_LABELS[statusStr];
		if (groupMatch) {
			return groupMatch.label;
		}
	}

	return statusStr;
}

/**
 * Get color for HTTP status code
 */
export function getHttpStatusColor(status: string | number): string {
	const statusStr = String(status);

	// Check for exact match first
	const exactMatch = HTTP_STATUS_LABELS[statusStr];
	if (exactMatch) {
		return exactMatch.color;
	}

	// Derive group from numeric status
	if (/^\d{3}$/.test(statusStr)) {
		const group = `${statusStr[0]}xx`;
		const groupMatch = HTTP_STATUS_LABELS[group];
		if (groupMatch) {
			return groupMatch.color;
		}
	}

	// Check for status group directly
	if (statusStr.endsWith("xx")) {
		const groupMatch = HTTP_STATUS_LABELS[statusStr];
		if (groupMatch) {
			return groupMatch.color;
		}
	}

	return TAILWIND_CHART_COLORS.muted;
}
