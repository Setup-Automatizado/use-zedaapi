/**
 * Message Queue Tab Component
 *
 * Message queue metrics and monitoring.
 *
 * @module components/metrics/tabs/message-queue-tab
 */

"use client";

import { Phone } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useInstanceNames } from "@/hooks/use-instance-names";
import {
	formatDuration,
	formatNumber,
	TAILWIND_CHART_COLORS,
} from "@/lib/metrics/constants";
import { formatPhoneNumber } from "@/lib/phone";
import { cn } from "@/lib/utils";
import type { MessageQueueMetrics } from "@/types/metrics";
import { HorizontalBarChart, MetricChart } from "../metric-chart";
import { MetricGaugeGroup, ProgressBar } from "../metric-gauge";
import { StatusIndicator } from "../status-indicator";

export interface MessageQueueTabProps {
	metrics?: MessageQueueMetrics;
	isLoading?: boolean;
}

export function MessageQueueTab({
	metrics,
	isLoading = false,
}: MessageQueueTabProps) {
	const { getDisplayName, getInstanceInfo } = useInstanceNames();

	// Check if queue system is active (has any data)
	const hasQueueData = metrics && (
		metrics.totalSize > 0 ||
		metrics.enqueued > 0 ||
		metrics.processed > 0 ||
		metrics.activeWorkers > 0 ||
		Object.keys(metrics.byInstance).length > 0 ||
		Object.keys(metrics.byType).length > 0
	);

	// Message type data
	const messageTypeData = Object.entries(metrics?.byType ?? {})
		.map(([type, data]) => ({
			name: type,
			enqueued: data.enqueued,
			processed: data.processed,
			failed: data.failed,
		}))
		.sort((a, b) => b.enqueued - a.enqueued);

	// Instance queue data with friendly names
	const instanceQueueData = Object.entries(
		metrics?.byInstance ?? {},
	)
		.map(([instanceId, data]) => {
			const info = getInstanceInfo(instanceId);
			return {
				name: info?.name || getDisplayName(instanceId),
				value: data.pending,
			};
		})
		.sort((a, b) => b.value - a.value)
		.slice(0, 10);

	// Instance data with full info
	const instanceData = Object.entries(
		metrics?.byInstance ?? {},
	)
		.map(([instanceId, data]) => {
			const info = getInstanceInfo(instanceId);
			return {
				instanceId,
				name: info?.name || getDisplayName(instanceId),
				phone: info?.phone || null,
				avatarUrl: info?.avatarUrl || null,
				size: data.size,
				pending: data.pending,
				processing: data.processing,
				sent: data.sent,
				failed: data.failed,
				workers: data.workers,
			};
		})
		.sort((a, b) => b.pending - a.pending);

	// Show info message when no queue data is available
	if (!isLoading && !hasQueueData) {
		return (
			<Card>
				<CardContent className="py-12">
					<div className="text-center space-y-4">
						<div className="mx-auto w-14 h-14 rounded-full bg-amber-500/10 flex items-center justify-center">
							<svg
								xmlns="http://www.w3.org/2000/svg"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								strokeWidth="2"
								strokeLinecap="round"
								strokeLinejoin="round"
								className="w-7 h-7 text-amber-600 dark:text-amber-400"
							>
								<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
								<line x1="12" y1="9" x2="12" y2="13" />
								<line x1="12" y1="17" x2="12.01" y2="17" />
							</svg>
						</div>
						<div>
							<h3 className="text-lg font-semibold">Message Queue Metrics Not Available</h3>
							<p className="text-sm text-muted-foreground mt-2 max-w-lg mx-auto">
								The message queue system is not reporting metrics. This feature requires the backend
								to expose <code className="text-xs bg-muted px-1.5 py-0.5 rounded">message_queue_*</code> Prometheus metrics.
							</p>
						</div>
						<div className="pt-2 space-y-2">
							<p className="text-xs text-muted-foreground font-medium">Expected metrics:</p>
							<div className="flex flex-wrap justify-center gap-2">
								{[
									"message_queue_size",
									"message_queue_enqueued_total",
									"message_queue_processed_total",
									"message_queue_workers_active",
								].map((metric) => (
									<span
										key={metric}
										className="text-xs font-mono bg-muted px-2 py-1 rounded"
									>
										{metric}
									</span>
								))}
							</div>
						</div>
					</div>
				</CardContent>
			</Card>
		);
	}

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

			{/* Instance Details */}
			<Card>
				<CardHeader className="pb-2">
					<CardTitle className="text-base font-medium">
						Queue by Instance
					</CardTitle>
				</CardHeader>
				<CardContent>
					{isLoading ? (
						<div className="space-y-3">
							{Array.from({ length: 3 }).map((_, i) => (
								<Skeleton key={i} className="h-16 w-full" />
							))}
						</div>
					) : instanceData.length === 0 ? (
						<p className="text-center text-sm text-muted-foreground py-8">
							No instance queue data available
						</p>
					) : (
						<div className="space-y-3">
							{instanceData.map((instance) => (
								<div
									key={instance.instanceId}
									className="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-4 rounded-lg border p-4 transition-colors hover:bg-muted/50"
								>
									{/* Instance Info */}
									<div className="flex items-center gap-3 min-w-0 sm:w-48 shrink-0">
										{/* Avatar */}
										<Avatar className="h-10 w-10 shrink-0">
											{instance.avatarUrl && (
												<AvatarImage src={instance.avatarUrl} alt={instance.name} />
											)}
											<AvatarFallback className="bg-primary/10 text-primary text-sm font-medium">
												{instance.name.slice(0, 2).toUpperCase()}
											</AvatarFallback>
										</Avatar>

										{/* Name & Phone */}
										<div className="min-w-0 flex-1">
											<p className="font-medium truncate">{instance.name}</p>
											{instance.phone && (
												<p className="text-xs text-muted-foreground flex items-center gap-1 truncate">
													<Phone className="h-3 w-3 shrink-0" />
													{formatPhoneNumber(instance.phone)}
												</p>
											)}
										</div>
									</div>

									{/* Stats Grid */}
									<div className="grid grid-cols-5 gap-2 sm:gap-4 text-center flex-1">
										<div>
											<p className="text-xs text-muted-foreground">Workers</p>
											<p className="font-semibold tabular-nums">
												{instance.workers.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Pending</p>
											<p className={cn(
												"font-semibold tabular-nums",
												instance.pending > 100 ? "text-amber-600 dark:text-amber-400" : ""
											)}>
												{instance.pending.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Processing</p>
											<p className="font-semibold tabular-nums text-cyan-600 dark:text-cyan-400">
												{instance.processing.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Sent</p>
											<p className="font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">
												{instance.sent.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Failed</p>
											<p className={cn(
												"font-semibold tabular-nums",
												instance.failed > 0 ? "text-red-600 dark:text-red-400" : "text-muted-foreground"
											)}>
												{instance.failed.toLocaleString()}
											</p>
										</div>
									</div>
								</div>
							))}
						</div>
					)}
				</CardContent>
			</Card>
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
