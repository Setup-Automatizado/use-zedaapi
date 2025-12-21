/**
 * HTTP Metrics Tab Component
 *
 * HTTP request metrics, latency, and error rates.
 *
 * @module components/metrics/tabs/http-metrics-tab
 */

"use client";

import { TAILWIND_CHART_COLORS } from "@/lib/metrics/constants";
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
	// Prepare status code data for pie chart
	const statusCodeData = Object.entries(metrics?.byStatus ?? {}).map(
		([status, count]) => ({
			name: `${status}xx`,
			value: count,
		}),
	);

	// Prepare path data for horizontal bar
	const pathData = Object.entries(metrics?.byPath ?? {})
		.map(([path, data]) => ({
			name: path,
			value: data.count,
		}))
		.sort((a, b) => b.value - a.value)
		.slice(0, 10);

	// Prepare method data
	const methodData = Object.entries(metrics?.byMethod ?? {}).map(
		([method, count]) => ({
			name: method,
			value: count,
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
				{/* Status Code Distribution */}
				<MetricChart
					type="pie"
					title="Status Codes"
					data={statusCodeData}
					xKey="name"
					yKeys={[
						{ key: "value", color: TAILWIND_CHART_COLORS.primary, label: "Requests" },
					]}
					height={250}
					isLoading={isLoading}
				/>

				{/* Method Distribution */}
				<MetricChart
					type="bar"
					title="Request Methods"
					data={methodData}
					xKey="name"
					yKeys={[
						{ key: "value", color: TAILWIND_CHART_COLORS.secondary, label: "Requests" },
					]}
					height={250}
					showLegend={false}
					isLoading={isLoading}
				/>
			</div>

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
