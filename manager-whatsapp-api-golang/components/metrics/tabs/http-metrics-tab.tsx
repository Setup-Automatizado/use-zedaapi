/**
 * HTTP Metrics Tab Component
 *
 * HTTP request metrics, latency, and error rates with friendly status labels.
 *
 * @module components/metrics/tabs/http-metrics-tab
 */

"use client";

import {
	getHttpStatusColor,
	getHttpStatusLabel,
	HTTP_STATUS_LABELS,
	TAILWIND_CHART_COLORS,
} from "@/lib/metrics/constants";
import type { HTTPMetrics } from "@/types/metrics";
import { HorizontalBarChart, MetricChart } from "../metric-chart";
import { MetricGaugeGroup } from "../metric-gauge";
import { DurationCell, MetricTable, NumberCell, PercentageCell } from "../metric-table";

export interface HTTPMetricsTabProps {
	metrics?: HTTPMetrics;
	isLoading?: boolean;
}

export function HTTPMetricsTab({
	metrics,
	isLoading = false,
}: HTTPMetricsTabProps) {
	// Prepare status code data for pie chart with friendly labels
	// Group by status code category (2xx, 4xx, 5xx) for cleaner visualization
	const statusGroups: Record<string, { count: number; codes: string[] }> = {};

	for (const [status, count] of Object.entries(metrics?.byStatus ?? {})) {
		// Derive group from status code (e.g., "200" -> "2xx", "404" -> "4xx")
		const group = status.length === 3 ? `${status[0]}xx` : status;
		if (!statusGroups[group]) {
			statusGroups[group] = { count: 0, codes: [] };
		}
		statusGroups[group].count += count;
		statusGroups[group].codes.push(status);
	}

	const statusCodeData = Object.entries(statusGroups).map(([group, data]) => {
		const info = HTTP_STATUS_LABELS[group];
		return {
			name: info ? info.label : group,
			fullLabel: info ? `${group} - ${info.description}` : group,
			value: data.count,
			color: info?.color || TAILWIND_CHART_COLORS.muted,
			codes: data.codes,
		};
	});

	// Prepare path data for horizontal bar
	const pathData = Object.entries(metrics?.byPath ?? {})
		.map(([path, data]) => ({
			name: path,
			value: data.count,
		}))
		.sort((a, b) => b.value - a.value)
		.slice(0, 10);

	// HTTP method colors for better visual distinction
	const METHOD_COLORS: Record<string, string> = {
		GET: "#10b981",      // emerald-500 - safe, read-only
		POST: "#3b82f6",     // blue-500 - create
		PUT: "#f59e0b",      // amber-500 - update
		PATCH: "#8b5cf6",    // violet-500 - partial update
		DELETE: "#ef4444",   // red-500 - destructive
		OPTIONS: "#6b7280",  // gray-500 - preflight
		HEAD: "#06b6d4",     // cyan-500 - metadata
	};

	// Prepare method data with colors
	const methodData = Object.entries(metrics?.byMethod ?? {}).map(
		([method, count]) => ({
			name: method,
			value: count,
			color: METHOD_COLORS[method.toUpperCase()] || TAILWIND_CHART_COLORS.muted,
		}),
	);

	// Prepare path table data
	const pathTableData = Object.entries(metrics?.byPath ?? {})
		.map(([path, data]) => ({
			path,
			count: data.count,
			avgLatencyMs: data.avgLatencyMs,
			errorCount: data.errorCount,
			errorRate:
				data.count > 0 ? (data.errorCount / data.count) * 100 : 0,
		}))
		.sort((a, b) => b.count - a.count);

	return (
		<div className="space-y-6">
			{/* Latency Gauges */}
			<MetricGaugeGroup
				title="Response Latency"
				gauges={[
					{
						value: metrics?.avgLatencyMs ?? 0,
						max: 2000,
						label: "Average",
						unit: "ms",
						size: "sm",
						thresholds: { warning: 200, critical: 1000, unit: "ms" },
					},
					{
						value: metrics?.p50LatencyMs ?? 0,
						max: 2000,
						label: "P50",
						unit: "ms",
						size: "sm",
						thresholds: { warning: 100, critical: 500, unit: "ms" },
					},
					{
						value: metrics?.p95LatencyMs ?? 0,
						max: 2000,
						label: "P95",
						unit: "ms",
						size: "sm",
						thresholds: { warning: 500, critical: 2000, unit: "ms" },
					},
					{
						value: metrics?.p99LatencyMs ?? 0,
						max: 5000,
						label: "P99",
						unit: "ms",
						size: "sm",
						thresholds: { warning: 1000, critical: 5000, unit: "ms" },
					},
				]}
				isLoading={isLoading}
			/>

			{/* Charts Grid */}
			<div className="grid gap-4 md:grid-cols-2">
				{/* Status Code Distribution with friendly labels */}
				<MetricChart
					type="pie"
					title="Response Status"
					data={statusCodeData}
					xKey="name"
					yKeys={[
						{ key: "value", color: TAILWIND_CHART_COLORS.primary, label: "Requests" },
					]}
					height={250}
					isLoading={isLoading}
					colors={statusCodeData.map((d) => d.color)}
				/>

				{/* Method Distribution */}
				<MetricChart
					type="bar"
					title="Request Methods"
					data={methodData}
					xKey="name"
					yKeys={[
						{ key: "value", color: TAILWIND_CHART_COLORS.primary, label: "Requests" },
					]}
					height={250}
					showLegend={false}
					isLoading={isLoading}
					colors={methodData.map((d) => d.color)}
				/>
			</div>

			{/* Status Code Legend */}
			{statusCodeData.length > 0 && (
				<StatusCodeLegend data={statusCodeData} />
			)}

			{/* Top Endpoints */}
			<HorizontalBarChart
				title="Top Endpoints"
				data={pathData}
				maxItems={10}
				color={TAILWIND_CHART_COLORS.tertiary}
				isLoading={isLoading}
			/>

			{/* Endpoint Details Table */}
			<MetricTable
				title="Endpoint Performance"
				data={pathTableData}
				columns={[
					{
						key: "path",
						header: "Path",
						format: (v) => (
							<span className="font-mono text-xs">{String(v)}</span>
						),
					},
					{
						key: "count",
						header: "Requests",
						align: "right",
						format: (v) => <NumberCell value={Number(v)} />,
					},
					{
						key: "avgLatencyMs",
						header: "Avg Latency",
						align: "right",
						format: (v) => <DurationCell ms={Number(v)} />,
					},
					{
						key: "errorCount",
						header: "Errors",
						align: "right",
						format: (v) => <NumberCell value={Number(v)} />,
					},
					{
						key: "errorRate",
						header: "Error Rate",
						align: "right",
						format: (v) => <PercentageCell value={Number(v)} />,
					},
				]}
				isLoading={isLoading}
				emptyMessage="No endpoint data available"
			/>
		</div>
	);
}

/**
 * Status Code Legend Component
 * Shows friendly descriptions for each status code category
 */
function StatusCodeLegend({
	data,
}: {
	data: Array<{ name: string; fullLabel: string; value: number; color: string; codes?: string[] }>;
}) {
	// Only show if we have data
	if (data.length === 0) return null;

	// Sort by value descending
	const sorted = [...data].sort((a, b) => b.value - a.value);
	const total = sorted.reduce((sum, d) => sum + d.value, 0);

	return (
		<div className="flex flex-wrap gap-4 px-2">
			{sorted.map((item) => {
				const percentage = total > 0 ? ((item.value / total) * 100).toFixed(1) : "0";
				const codesInfo = item.codes?.length ? ` (${item.codes.sort().join(", ")})` : "";
				return (
					<div
						key={item.name}
						className="flex items-center gap-2"
						title={`${item.fullLabel}${codesInfo}`}
					>
						<span
							className="h-3 w-3 rounded-full"
							style={{ backgroundColor: item.color }}
						/>
						<span className="text-sm font-medium">{item.name}</span>
						<span className="text-sm text-muted-foreground">
							{item.value.toLocaleString()} ({percentage}%)
						</span>
					</div>
				);
			})}
		</div>
	);
}

/**
 * HTTP Status Badge
 * Shows a colored badge with the friendly status label
 */
export function HttpStatusBadge({
	status,
	count,
}: {
	status: string | number;
	count?: number;
}) {
	const label = getHttpStatusLabel(status);
	const color = getHttpStatusColor(status);

	return (
		<span
			className="inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-xs font-medium"
			style={{
				backgroundColor: `${color}20`,
				color: color,
			}}
		>
			<span
				className="h-2 w-2 rounded-full"
				style={{ backgroundColor: color }}
			/>
			{label}
			{count !== undefined && (
				<span className="opacity-75">({count.toLocaleString()})</span>
			)}
		</span>
	);
}
