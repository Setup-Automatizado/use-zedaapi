/**
 * Message Queue Tab Component
 *
 * Message queue metrics and monitoring.
 *
 * @module components/metrics/tabs/message-queue-tab
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
import type { MessageQueueMetrics } from "@/types/metrics";
import { HorizontalBarChart, MetricChart } from "../metric-chart";
import { MetricGaugeGroup, ProgressBar } from "../metric-gauge";
import { MetricTable, NumberCell } from "../metric-table";
import { StatusIndicator } from "../status-indicator";

export interface MessageQueueTabProps {
	metrics?: MessageQueueMetrics;
	isLoading?: boolean;
}

export function MessageQueueTab({
	metrics,
	isLoading = false,
}: MessageQueueTabProps) {
	// Message type data
	const messageTypeData = Object.entries(metrics?.byType ?? {})
		.map(([type, data]) => ({
			name: type,
			enqueued: data.enqueued,
			processed: data.processed,
			failed: data.failed,
		}))
		.sort((a, b) => b.enqueued - a.enqueued);

	// Instance queue data
	const instanceQueueData = Object.entries(
		metrics?.byInstance ?? {},
	)
		.map(([instanceId, data]) => ({
			name: instanceId.slice(0, 8) + "...",
			value: data.pending,
		}))
		.sort((a, b) => b.value - a.value)
		.slice(0, 10);

	// Instance table data
	const instanceTableData = Object.entries(
		metrics?.byInstance ?? {},
	)
		.map(([instanceId, data]) => ({
			instanceId,
			size: data.size,
			pending: data.pending,
			processing: data.processing,
			sent: data.sent,
			failed: data.failed,
			workers: data.workers,
		}))
		.sort((a, b) => b.pending - a.pending);

	return (
		<div className="space-y-6">
			{/* Queue Summary Cards */}
			<div className="grid gap-4 md:grid-cols-4">
				<SummaryCard
					title="Total Queue Size"
					value={formatNumber(metrics?.totalSize ?? 0)}
					subtitle="Messages in queue"
					isLoading={isLoading}
				/>
				<SummaryCard
					title="Active Workers"
					value={metrics?.activeWorkers ?? 0}
					subtitle="Processing messages"
					isLoading={isLoading}
				/>
				<SummaryCard
					title="Avg Processing Time"
					value={formatDuration(metrics?.avgProcessingMs ?? 0)}
					subtitle="Per message"
					isLoading={isLoading}
				/>
				<SummaryCard
					title="DLQ Size"
					value={formatNumber(metrics?.dlqSize ?? 0)}
					subtitle="Dead letter queue"
					status={
						metrics && metrics.dlqSize > 0 ? "warning" : undefined
					}
					isLoading={isLoading}
				/>
			</div>

			{/* Queue Status Breakdown */}
			<Card>
				<CardHeader className="pb-2">
					<CardTitle className="text-base font-medium">
						Queue Status Breakdown
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					{isLoading ? (
						<Skeleton className="h-8 w-full" />
					) : (
						<>
							<div className="grid grid-cols-4 gap-4 text-center">
								<div>
									<StatusIndicator status="healthy" size="sm" />
									<p className="text-2xl font-bold tabular-nums text-blue-600 dark:text-blue-400">
										{formatNumber(metrics?.pending ?? 0)}
									</p>
									<p className="text-xs text-muted-foreground">Pending</p>
								</div>
								<div>
									<StatusIndicator status="healthy" size="sm" />
									<p className="text-2xl font-bold tabular-nums text-cyan-600 dark:text-cyan-400">
										{formatNumber(metrics?.processing ?? 0)}
									</p>
									<p className="text-xs text-muted-foreground">Processing</p>
								</div>
								<div>
									<StatusIndicator status="healthy" size="sm" />
									<p className="text-2xl font-bold tabular-nums text-emerald-600 dark:text-emerald-400">
										{formatNumber(metrics?.sent ?? 0)}
									</p>
									<p className="text-xs text-muted-foreground">Sent</p>
								</div>
								<div>
									<StatusIndicator
										status={
											metrics && metrics.failed > 0
												? "critical"
												: "healthy"
										}
										size="sm"
									/>
									<p
										className={cn(
											"text-2xl font-bold tabular-nums",
											metrics && metrics.failed > 0
												? "text-red-600 dark:text-red-400"
												: "text-muted-foreground",
										)}
									>
										{formatNumber(metrics?.failed ?? 0)}
									</p>
									<p className="text-xs text-muted-foreground">Failed</p>
								</div>
							</div>

							{/* Progress bar showing throughput */}
							<ProgressBar
								value={metrics?.sent ?? 0}
								max={Math.max(metrics?.enqueued ?? 0, 1)}
								label="Throughput"
								status={
									metrics && metrics.enqueued > 0
										? metrics.sent / metrics.enqueued >= 0.95
											? "healthy"
											: metrics.sent / metrics.enqueued >= 0.8
												? "warning"
												: "critical"
										: "healthy"
								}
							/>
						</>
					)}
				</CardContent>
			</Card>

			{/* Charts Grid */}
			<div className="grid gap-4 md:grid-cols-2">
				{/* Messages by Type */}
				<MetricChart
					type="bar"
					title="Messages by Type"
					data={messageTypeData}
					xKey="name"
					yKeys={[
						{
							key: "enqueued",
							color: TAILWIND_CHART_COLORS.primary,
							label: "Enqueued",
						},
						{
							key: "processed",
							color: TAILWIND_CHART_COLORS.success,
							label: "Processed",
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

				{/* Pending by Instance */}
				<HorizontalBarChart
					title="Pending Messages by Instance"
					data={instanceQueueData}
					maxItems={10}
					color={TAILWIND_CHART_COLORS.warning}
					isLoading={isLoading}
				/>
			</div>

			{/* Retry and Error Stats */}
			<MetricGaugeGroup
				title="Queue Health"
				gauges={[
					{
						value:
							metrics && metrics.enqueued > 0
								? Math.min(
										(metrics.sent / metrics.enqueued) * 100,
										100,
									)
								: 100,
						label: "Success Rate",
						size: "sm",
						status:
							metrics && metrics.enqueued > 0
								? metrics.sent / metrics.enqueued >= 0.95
									? "healthy"
									: metrics.sent / metrics.enqueued >= 0.8
										? "warning"
										: "critical"
								: "healthy",
					},
					{
						value:
							metrics && metrics.processed > 0
								? Math.min(
										(metrics.retries / metrics.processed) * 100,
										100,
									)
								: 0,
						label: "Retry Rate",
						size: "sm",
						thresholds: { warning: 5, critical: 20, unit: "%" },
					},
					{
						value:
							metrics && metrics.processed > 0
								? Math.min(
										(metrics.errors / metrics.processed) * 100,
										100,
									)
								: 0,
						label: "Error Rate",
						size: "sm",
						thresholds: { warning: 1, critical: 5, unit: "%" },
					},
				]}
				isLoading={isLoading}
			/>

			{/* Instance Details Table */}
			<MetricTable
				title="Queue by Instance"
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
						key: "workers",
						header: "Workers",
						align: "right",
						format: (v) => <NumberCell value={Number(v)} />,
					},
					{
						key: "pending",
						header: "Pending",
						align: "right",
						format: (v) => (
							<span
								className={cn(
									"tabular-nums",
									Number(v) > 100 && "text-amber-600 dark:text-amber-400",
								)}
							>
								{Number(v).toLocaleString()}
							</span>
						),
					},
					{
						key: "processing",
						header: "Processing",
						align: "right",
						format: (v) => <NumberCell value={Number(v)} />,
					},
					{
						key: "sent",
						header: "Sent",
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
				]}
				isLoading={isLoading}
				emptyMessage="No instance queue data available"
			/>
		</div>
	);
}

/**
 * Summary Card
 */
function SummaryCard({
	title,
	value,
	subtitle,
	status,
	isLoading,
}: {
	title: string;
	value: string | number;
	subtitle: string;
	status?: "warning" | "critical";
	isLoading?: boolean;
}) {
	if (isLoading) {
		return (
			<Card>
				<CardHeader className="pb-2">
					<Skeleton className="h-4 w-24" />
				</CardHeader>
				<CardContent>
					<Skeleton className="h-8 w-16" />
					<Skeleton className="mt-1 h-3 w-20" />
				</CardContent>
			</Card>
		);
	}

	return (
		<Card>
			<CardHeader className="pb-2">
				<CardTitle className="text-sm font-medium text-muted-foreground">
					{title}
				</CardTitle>
			</CardHeader>
			<CardContent>
				<p
					className={cn(
						"text-2xl font-bold tabular-nums",
						status === "warning" && "text-amber-600 dark:text-amber-400",
						status === "critical" && "text-red-600 dark:text-red-400",
					)}
				>
					{value}
				</p>
				<p className="text-xs text-muted-foreground">{subtitle}</p>
			</CardContent>
		</Card>
	);
}
