/**
 * Status Cache Metrics Tab Component
 *
 * Message status caching metrics for read, delivered, played, sent events.
 *
 * @module components/metrics/tabs/status-cache-tab
 */

"use client";

import { Database, Phone } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useInstanceNames } from "@/hooks/use-instance-names";
import {
	formatDuration,
	formatNumber,
	formatPercentage,
	TAILWIND_CHART_COLORS,
} from "@/lib/metrics/constants";
import { formatPhoneNumber } from "@/lib/phone";
import { cn } from "@/lib/utils";
import type { StatusCacheMetrics } from "@/types/metrics";
import { HorizontalBarChart, MetricChart } from "../metric-chart";
import { MetricGaugeGroup, ProgressBar } from "../metric-gauge";
import { StatusIndicator } from "../status-indicator";

export interface StatusCacheTabProps {
	metrics?: StatusCacheMetrics;
	isLoading?: boolean;
}

export function StatusCacheTab({
	metrics,
	isLoading = false,
}: StatusCacheTabProps) {
	const { getDisplayName, getInstanceInfo } = useInstanceNames();

	// Check if status cache system is active
	const hasStatusCacheData =
		metrics &&
		(metrics.totalOperations > 0 ||
			metrics.totalSize > 0 ||
			metrics.totalHits > 0 ||
			metrics.totalMisses > 0 ||
			Object.keys(metrics.byInstance).length > 0);

	// Instance cache data with friendly names
	const instanceCacheData = Object.entries(metrics?.byInstance ?? {})
		.map(([instanceId, data]) => {
			const info = getInstanceInfo(instanceId);
			return {
				instanceId,
				name: info?.name || getDisplayName(instanceId),
				phone: info?.phone || null,
				avatarUrl: info?.avatarUrl || null,
				size: data.size,
				operations: data.operations,
				hits: data.hits,
				misses: data.misses,
				hitRate: data.hitRate,
				suppressions: data.suppressions,
				flushed: data.flushed,
			};
		})
		.sort((a, b) => b.size - a.size);

	// Suppressions by status type data for chart
	const suppressionsByTypeData = Object.entries(metrics?.byStatusType ?? {})
		.map(([statusType, count]) => ({
			name: formatStatusType(statusType),
			value: count,
		}))
		.sort((a, b) => b.value - a.value);

	// Flushed by trigger data for chart
	const flushedByTriggerData = Object.entries(metrics?.byTrigger ?? {})
		.map(([trigger, count]) => ({
			name: formatTrigger(trigger),
			value: count,
		}))
		.sort((a, b) => b.value - a.value);

	// Operations by type data for chart
	const operationsByTypeData = Object.entries(metrics?.byOperation ?? {})
		.map(([operation, data]) => ({
			name: formatOperation(operation),
			success: data.success,
			failed: data.failed,
			total: data.count,
		}))
		.sort((a, b) => b.total - a.total);

	// Instance chart data for stacked bar
	const instanceChartData = instanceCacheData.slice(0, 10).map((inst) => ({
		name: inst.name.length > 12 ? inst.name.slice(0, 12) + "..." : inst.name,
		hits: inst.hits,
		misses: inst.misses,
		suppressions: inst.suppressions,
	}));

	// Show info message when no status cache data is available
	if (!isLoading && !hasStatusCacheData) {
		return (
			<Card>
				<CardContent className="py-12">
					<div className="text-center space-y-4">
						<div className="mx-auto w-14 h-14 rounded-full bg-purple-500/10 flex items-center justify-center">
							<Database className="w-7 h-7 text-purple-600 dark:text-purple-400" />
						</div>
						<div>
							<h3 className="text-lg font-semibold">
								Status Cache Metrics Not Available
							</h3>
							<p className="text-sm text-muted-foreground mt-2 max-w-lg mx-auto">
								The status cache system is not reporting metrics yet. Metrics
								will appear once message status events (read, delivered, played,
								sent) are cached.
							</p>
						</div>
						<div className="pt-2 space-y-2">
							<p className="text-xs text-muted-foreground font-medium">
								Expected metrics:
							</p>
							<div className="flex flex-wrap justify-center gap-2">
								{[
									"status_cache_operations_total",
									"status_cache_size",
									"status_cache_hits_total",
									"status_cache_misses_total",
									"status_cache_suppressions_total",
									"status_cache_flushed_total",
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
			{/* Summary Cards */}
			<div className="grid gap-4 md:grid-cols-4">
				<SummaryCard
					title="Cache Size"
					value={formatNumber(metrics?.totalSize ?? 0)}
					subtitle="Cached entries"
					isLoading={isLoading}
				/>
				<SummaryCard
					title="Hit Rate"
					value={formatPercentage(metrics?.hitRate ?? 0)}
					subtitle="Cache efficiency"
					status={
						metrics && metrics.hitRate < 80
							? metrics.hitRate < 50
								? "critical"
								: "warning"
							: undefined
					}
					isLoading={isLoading}
				/>
				<SummaryCard
					title="Suppressions"
					value={formatNumber(metrics?.totalSuppressions ?? 0)}
					subtitle="Webhooks suppressed"
					isLoading={isLoading}
				/>
				<SummaryCard
					title="Avg Latency"
					value={formatDuration(metrics?.avgDurationMs ?? 0)}
					subtitle="Operation time"
					isLoading={isLoading}
				/>
			</div>

			{/* Cache Hit/Miss Breakdown */}
			<Card>
				<CardHeader className="pb-2">
					<CardTitle className="text-base font-medium">
						Cache Performance
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
									<p className="text-2xl font-bold tabular-nums text-emerald-600 dark:text-emerald-400">
										{formatNumber(metrics?.totalHits ?? 0)}
									</p>
									<p className="text-xs text-muted-foreground">Cache Hits</p>
								</div>
								<div>
									<StatusIndicator
										status={
											metrics && metrics.totalMisses > metrics.totalHits
												? "warning"
												: "healthy"
										}
										size="sm"
									/>
									<p
										className={cn(
											"text-2xl font-bold tabular-nums",
											metrics && metrics.totalMisses > metrics.totalHits
												? "text-amber-600 dark:text-amber-400"
												: "text-muted-foreground",
										)}
									>
										{formatNumber(metrics?.totalMisses ?? 0)}
									</p>
									<p className="text-xs text-muted-foreground">Cache Misses</p>
								</div>
								<div>
									<StatusIndicator status="healthy" size="sm" />
									<p className="text-2xl font-bold tabular-nums text-purple-600 dark:text-purple-400">
										{formatNumber(metrics?.totalSuppressions ?? 0)}
									</p>
									<p className="text-xs text-muted-foreground">Suppressions</p>
								</div>
								<div>
									<StatusIndicator status="healthy" size="sm" />
									<p className="text-2xl font-bold tabular-nums text-blue-600 dark:text-blue-400">
										{formatNumber(metrics?.totalFlushed ?? 0)}
									</p>
									<p className="text-xs text-muted-foreground">Flushed</p>
								</div>
							</div>

							<ProgressBar
								value={metrics?.totalHits ?? 0}
								max={Math.max(
									(metrics?.totalHits ?? 0) + (metrics?.totalMisses ?? 0),
									1,
								)}
								label="Hit Rate"
								status={
									metrics && metrics.totalHits + metrics.totalMisses > 0
										? metrics.hitRate >= 80
											? "healthy"
											: metrics.hitRate >= 50
												? "warning"
												: "critical"
										: "healthy"
								}
							/>
						</>
					)}
				</CardContent>
			</Card>

			{/* Latency Percentiles */}
			<Card>
				<CardHeader className="pb-2">
					<CardTitle className="text-base font-medium">
						Operation Latency Distribution
					</CardTitle>
				</CardHeader>
				<CardContent>
					{isLoading ? (
						<Skeleton className="h-16 w-full" />
					) : (
						<div className="grid grid-cols-4 gap-4 text-center">
							<div>
								<p className="text-sm text-muted-foreground">Average</p>
								<p className="text-xl font-bold tabular-nums">
									{formatDuration(metrics?.avgDurationMs ?? 0)}
								</p>
							</div>
							<div>
								<p className="text-sm text-muted-foreground">P50</p>
								<p className="text-xl font-bold tabular-nums">
									{formatDuration(metrics?.p50DurationMs ?? 0)}
								</p>
							</div>
							<div>
								<p className="text-sm text-muted-foreground">P95</p>
								<p className="text-xl font-bold tabular-nums text-amber-600 dark:text-amber-400">
									{formatDuration(metrics?.p95DurationMs ?? 0)}
								</p>
							</div>
							<div>
								<p className="text-sm text-muted-foreground">P99</p>
								<p className="text-xl font-bold tabular-nums text-red-600 dark:text-red-400">
									{formatDuration(metrics?.p99DurationMs ?? 0)}
								</p>
							</div>
						</div>
					)}
				</CardContent>
			</Card>

			{/* Charts Grid */}
			<div className="grid gap-4 md:grid-cols-2">
				{/* Operations by Type */}
				<MetricChart
					type="bar"
					title="Operations by Type"
					data={operationsByTypeData}
					xKey="name"
					yKeys={[
						{
							key: "success",
							color: TAILWIND_CHART_COLORS.success,
							label: "Success",
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

				{/* Suppressions by Status Type */}
				<HorizontalBarChart
					title="Suppressions by Status Type"
					data={suppressionsByTypeData}
					maxItems={10}
					color={TAILWIND_CHART_COLORS.primary}
					isLoading={isLoading}
				/>
			</div>

			{/* Additional Charts */}
			<div className="grid gap-4 md:grid-cols-2">
				{/* Cache by Instance */}
				<MetricChart
					type="bar"
					title="Cache Activity by Instance"
					data={instanceChartData}
					xKey="name"
					yKeys={[
						{
							key: "hits",
							color: TAILWIND_CHART_COLORS.success,
							label: "Hits",
						},
						{
							key: "misses",
							color: TAILWIND_CHART_COLORS.warning,
							label: "Misses",
						},
						{
							key: "suppressions",
							color: TAILWIND_CHART_COLORS.primary,
							label: "Suppressions",
						},
					]}
					height={300}
					isLoading={isLoading}
				/>

				{/* Flushed by Trigger */}
				<HorizontalBarChart
					title="Flushed by Trigger"
					data={flushedByTriggerData}
					maxItems={10}
					color={TAILWIND_CHART_COLORS.tertiary}
					isLoading={isLoading}
				/>
			</div>

			{/* Health Gauges */}
			<MetricGaugeGroup
				title="Cache Health"
				gauges={[
					{
						value: metrics?.hitRate ?? 0,
						label: "Hit Rate",
						size: "sm",
						status:
							metrics && metrics.hitRate < 80
								? metrics.hitRate < 50
									? "critical"
									: "warning"
								: "healthy",
					},
					{
						value:
							metrics && metrics.totalOperations > 0
								? Math.min(
										(metrics.totalSuppressions / metrics.totalOperations) * 100,
										100,
									)
								: 0,
						label: "Suppression Rate",
						size: "sm",
						thresholds: { warning: 80, critical: 95, unit: "%" },
						status: "healthy", // High suppression is good
					},
					{
						value:
							metrics && metrics.totalOperations > 0
								? Math.min(
										(Object.values(metrics.byOperation).reduce(
											(acc, op) => acc + op.failed,
											0,
										) /
											metrics.totalOperations) *
											100,
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
			<Card>
				<CardHeader className="pb-2">
					<CardTitle className="text-base font-medium">
						Cache by Instance
					</CardTitle>
				</CardHeader>
				<CardContent>
					{isLoading ? (
						<div className="space-y-3">
							{Array.from({ length: 3 }).map((_, i) => (
								<Skeleton key={i} className="h-16 w-full" />
							))}
						</div>
					) : instanceCacheData.length === 0 ? (
						<p className="text-center text-sm text-muted-foreground py-8">
							No instance cache data available
						</p>
					) : (
						<div className="space-y-3">
							{instanceCacheData.map((instance) => (
								<div
									key={instance.instanceId}
									className="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-4 rounded-lg border p-4 transition-colors hover:bg-muted/50"
								>
									{/* Instance Info */}
									<div className="flex items-center gap-3 min-w-0 sm:w-48 shrink-0">
										{/* Avatar */}
										<Avatar className="h-10 w-10 shrink-0">
											{instance.avatarUrl && (
												<AvatarImage
													src={instance.avatarUrl}
													alt={instance.name}
												/>
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
									<div className="grid grid-cols-6 gap-2 sm:gap-4 text-center flex-1">
										<div>
											<p className="text-xs text-muted-foreground">Size</p>
											<p className="font-semibold tabular-nums">
												{instance.size.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Hits</p>
											<p className="font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">
												{instance.hits.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Misses</p>
											<p
												className={cn(
													"font-semibold tabular-nums",
													instance.misses > instance.hits
														? "text-amber-600 dark:text-amber-400"
														: "text-muted-foreground",
												)}
											>
												{instance.misses.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Hit Rate</p>
											<p
												className={cn(
													"font-semibold tabular-nums",
													instance.hitRate >= 80
														? "text-emerald-600 dark:text-emerald-400"
														: instance.hitRate >= 50
															? "text-amber-600 dark:text-amber-400"
															: "text-red-600 dark:text-red-400",
												)}
											>
												{formatPercentage(instance.hitRate)}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">
												Suppressed
											</p>
											<p className="font-semibold tabular-nums text-purple-600 dark:text-purple-400">
												{instance.suppressions.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Flushed</p>
											<p className="font-semibold tabular-nums text-blue-600 dark:text-blue-400">
												{instance.flushed.toLocaleString()}
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

/**
 * Format status type for display
 */
function formatStatusType(statusType: string): string {
	const mapping: Record<string, string> = {
		read: "Read",
		delivered: "Delivered",
		played: "Played",
		sent: "Sent",
	};
	return (
		mapping[statusType.toLowerCase()] ||
		statusType.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase())
	);
}

/**
 * Format trigger for display
 */
function formatTrigger(trigger: string): string {
	const mapping: Record<string, string> = {
		manual: "Manual Flush",
		ttl: "TTL Expiry",
		shutdown: "Shutdown",
		api: "API Request",
	};
	return (
		mapping[trigger.toLowerCase()] ||
		trigger.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase())
	);
}

/**
 * Format operation for display
 */
function formatOperation(operation: string): string {
	const mapping: Record<string, string> = {
		upsert: "Upsert",
		get: "Get",
		delete: "Delete",
		flush: "Flush",
		add_participant: "Add Participant",
	};
	return (
		mapping[operation.toLowerCase()] ||
		operation.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase())
	);
}
