/**
 * Event Metrics Tab Component
 *
 * Event pipeline metrics and monitoring.
 *
 * @module components/metrics/tabs/event-metrics-tab
 */

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
	formatDuration,
	formatNumber,
	TAILWIND_CHART_COLORS,
} from "@/lib/metrics/constants";
import { cn } from "@/lib/utils";
import type { EventMetrics, HealthLevel } from "@/types/metrics";
import { HorizontalBarChart, MetricChart } from "../metric-chart";
import { ProgressBar } from "../metric-gauge";
import { MetricTable, NumberCell } from "../metric-table";
import { StatusIndicator } from "../status-indicator";

export interface EventMetricsTabProps {
	metrics?: EventMetrics;
	isLoading?: boolean;
}

export function EventMetricsTab({
	metrics,
	isLoading = false,
}: EventMetricsTabProps) {
	// Event type data for charts
	const eventTypeData = Object.entries(metrics?.byType ?? {}).map(
		([type, data]) => ({
			name: type,
			captured: data.captured,
			processed: data.processed,
			delivered: data.delivered,
			failed: data.failed,
		}),
	);

	// Instance backlog data
	const instanceBacklogData = Object.entries(metrics?.byInstance ?? {})
		.map(([instanceId, data]) => ({
			name: instanceId.slice(0, 8) + "...",
			value: data.backlog,
		}))
		.sort((a, b) => b.value - a.value)
		.slice(0, 10);

	// Instance table data
	const instanceTableData = Object.entries(metrics?.byInstance ?? {})
		.map(([instanceId, data]) => ({
			instanceId,
			captured: data.captured,
			processed: data.processed,
			delivered: data.delivered,
			failed: data.failed,
			backlog: data.backlog,
		}))
		.sort((a, b) => b.backlog - a.backlog);

	return (
		<div className="space-y-6">
			{/* Event Pipeline Flow */}
			<EventPipelineFlow
				captured={metrics?.captured ?? 0}
				buffered={metrics?.buffered ?? 0}
				inserted={metrics?.inserted ?? 0}
				processed={metrics?.processed ?? 0}
				delivered={metrics?.delivered ?? 0}
				failed={metrics?.failed ?? 0}
				retries={metrics?.retries ?? 0}
				isLoading={isLoading}
			/>

			{/* Performance Metrics */}
			<div className="grid gap-4 md:grid-cols-3">
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-base font-medium">
							Processing Latency
						</CardTitle>
					</CardHeader>
					<CardContent>
						{isLoading ? (
							<Skeleton className="h-8 w-24" />
						) : (
							<p className="text-2xl font-bold tabular-nums">
								{formatDuration(metrics?.avgProcessingMs ?? 0)}
							</p>
						)}
						<p className="text-xs text-muted-foreground">Average time</p>
					</CardContent>
				</Card>

				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-base font-medium">
							Delivery Latency
						</CardTitle>
					</CardHeader>
					<CardContent>
						{isLoading ? (
							<Skeleton className="h-8 w-24" />
						) : (
							<p className="text-2xl font-bold tabular-nums">
								{formatDuration(metrics?.avgDeliveryMs ?? 0)}
							</p>
						)}
						<p className="text-xs text-muted-foreground">End-to-end</p>
					</CardContent>
				</Card>

				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-base font-medium">DLQ Size</CardTitle>
					</CardHeader>
					<CardContent>
						{isLoading ? (
							<Skeleton className="h-8 w-24" />
						) : (
							<p
								className={cn(
									"text-2xl font-bold tabular-nums",
									metrics && metrics.dlqSize > 0
										? "text-amber-600 dark:text-amber-400"
										: "",
								)}
							>
								{formatNumber(metrics?.dlqSize ?? 0)}
							</p>
						)}
						<p className="text-xs text-muted-foreground">Dead letter queue</p>
					</CardContent>
				</Card>
			</div>

			{/* Charts Grid */}
			<div className="grid gap-4 md:grid-cols-2">
				{/* Events by Type */}
				<MetricChart
					type="bar"
					title="Events by Type"
					data={eventTypeData}
					xKey="name"
					yKeys={[
						{
							key: "captured",
							color: TAILWIND_CHART_COLORS.primary,
							label: "Captured",
						},
						{
							key: "delivered",
							color: TAILWIND_CHART_COLORS.success,
							label: "Delivered",
						},
						{
							key: "failed",
							color: TAILWIND_CHART_COLORS.error,
							label: "Failed",
						},
					]}
					height={300}
					isLoading={isLoading}
				/>

				{/* Backlog by Instance */}
				<HorizontalBarChart
					title="Outbox Backlog by Instance"
					data={instanceBacklogData}
					maxItems={10}
					color={TAILWIND_CHART_COLORS.warning}
					isLoading={isLoading}
				/>
			</div>

			{/* Instance Details Table */}
			<MetricTable
				title="Events by Instance"
				data={instanceTableData}
				columns={[
					{
						key: "instanceId",
						header: "Instance",
						format: (v) => (
							<span className="font-mono text-xs">
								{String(v).slice(0, 12)}...
							</span>
						),
					},
					{
						key: "captured",
						header: "Captured",
						align: "right",
						format: (v) => <NumberCell value={Number(v)} />,
					},
					{
						key: "processed",
						header: "Processed",
						align: "right",
						format: (v) => <NumberCell value={Number(v)} />,
					},
					{
						key: "delivered",
						header: "Delivered",
						align: "right",
						format: (v) => <NumberCell value={Number(v)} />,
					},
					{
						key: "failed",
						header: "Failed",
						align: "right",
						format: (v) => (
							<span
								className={cn(
									"tabular-nums",
									Number(v) > 0 && "text-red-600 dark:text-red-400",
								)}
							>
								{Number(v).toLocaleString()}
							</span>
						),
					},
					{
						key: "backlog",
						header: "Backlog",
						align: "right",
						format: (v) => (
							<span
								className={cn(
									"tabular-nums",
									Number(v) > 50 && "text-amber-600 dark:text-amber-400",
								)}
							>
								{Number(v).toLocaleString()}
							</span>
						),
					},
				]}
				isLoading={isLoading}
				emptyMessage="No instance data available"
			/>
		</div>
	);
}

/**
 * Event Pipeline Flow Diagram
 */
function EventPipelineFlow({
	captured,
	buffered,
	inserted,
	processed,
	delivered,
	failed,
	retries,
	isLoading,
}: {
	captured: number;
	buffered: number;
	inserted: number;
	processed: number;
	delivered: number;
	failed: number;
	retries: number;
	isLoading?: boolean;
}) {
	if (isLoading) {
		return (
			<Card>
				<CardHeader className="pb-2">
					<Skeleton className="h-5 w-40" />
				</CardHeader>
				<CardContent>
					<div className="flex items-center justify-between gap-4">
						{Array.from({ length: 5 }).map((_, i) => (
							<div key={i} className="flex-1">
								<Skeleton className="h-16 w-full" />
							</div>
						))}
					</div>
				</CardContent>
			</Card>
		);
	}

	const stages = [
		{ label: "Captured", value: captured, status: "healthy" as HealthLevel },
		{ label: "Buffered", value: buffered, status: "healthy" as HealthLevel },
		{ label: "Inserted", value: inserted, status: "healthy" as HealthLevel },
		{ label: "Processed", value: processed, status: "healthy" as HealthLevel },
		{
			label: "Delivered",
			value: delivered,
			status: "healthy" as HealthLevel,
		},
	];

	return (
		<Card>
			<CardHeader className="pb-2">
				<CardTitle className="text-base font-medium">
					Event Pipeline Flow
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div className="flex items-center gap-2 overflow-x-auto pb-2">
					{stages.map((stage, index) => (
						<div key={stage.label} className="flex items-center">
							<div className="flex flex-col items-center min-w-[80px]">
								<div className="flex items-center gap-1 mb-1">
									<StatusIndicator status={stage.status} size="sm" />
									<span className="text-xs text-muted-foreground">
										{stage.label}
									</span>
								</div>
								<span className="text-lg font-bold tabular-nums">
									{formatNumber(stage.value)}
								</span>
							</div>
							{index < stages.length - 1 && (
								<div className="mx-2 text-muted-foreground">→</div>
							)}
						</div>
					))}

					{/* Failed Branch */}
					{failed > 0 && (
						<>
							<div className="mx-2 text-muted-foreground">|</div>
							<div className="flex flex-col items-center min-w-[80px]">
								<div className="flex items-center gap-1 mb-1">
									<StatusIndicator status="critical" size="sm" />
									<span className="text-xs text-muted-foreground">Failed</span>
								</div>
								<span className="text-lg font-bold tabular-nums text-red-600 dark:text-red-400">
									{formatNumber(failed)}
								</span>
							</div>
						</>
					)}

					{/* Retries */}
					{retries > 0 && (
						<>
							<div className="mx-2 text-muted-foreground">|</div>
							<div className="flex flex-col items-center min-w-[80px]">
								<div className="flex items-center gap-1 mb-1">
									<StatusIndicator status="warning" size="sm" />
									<span className="text-xs text-muted-foreground">Retries</span>
								</div>
								<span className="text-lg font-bold tabular-nums text-amber-600 dark:text-amber-400">
									{formatNumber(retries)}
								</span>
							</div>
						</>
					)}
				</div>

				{/* Progress bars */}
				<div className="mt-4 space-y-2">
					<ProgressBar
						value={delivered}
						max={Math.max(captured, 1)}
						label="Delivery Rate"
						status={
							captured > 0
								? delivered / captured >= 0.95
									? "healthy"
									: delivered / captured >= 0.8
										? "warning"
										: "critical"
								: "healthy"
						}
					/>
				</div>
			</CardContent>
		</Card>
	);
}
