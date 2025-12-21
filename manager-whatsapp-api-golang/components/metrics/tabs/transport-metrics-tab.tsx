/**
 * Transport Metrics Tab Component
 *
 * Webhook/HTTP transport delivery metrics and monitoring.
 *
 * @module components/metrics/tabs/transport-metrics-tab
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
	formatPercentage,
	TAILWIND_CHART_COLORS,
} from "@/lib/metrics/constants";
import { formatPhoneNumber } from "@/lib/phone";
import { cn } from "@/lib/utils";
import type { TransportMetrics } from "@/types/metrics";
import { HorizontalBarChart, MetricChart } from "../metric-chart";
import { MetricGaugeGroup, ProgressBar } from "../metric-gauge";
import { DurationCell, MetricTable, NumberCell } from "../metric-table";
import { StatusIndicator } from "../status-indicator";

export interface TransportMetricsTabProps {
	metrics?: TransportMetrics;
	isLoading?: boolean;
}

export function TransportMetricsTab({
	metrics,
	isLoading = false,
}: TransportMetricsTabProps) {
	const { getDisplayName, getInstanceInfo } = useInstanceNames();

	// Check if transport system is active
	const hasTransportData = metrics && (
		metrics.totalDeliveries > 0 ||
		metrics.totalRetries > 0 ||
		Object.keys(metrics.byInstance).length > 0 ||
		Object.keys(metrics.byErrorType).length > 0
	);

	// Instance delivery data with friendly names
	const instanceDeliveryData = Object.entries(metrics?.byInstance ?? {})
		.map(([instanceId, data]) => {
			const info = getInstanceInfo(instanceId);
			return {
				instanceId,
				name: info?.name || getDisplayName(instanceId),
				phone: info?.phone || null,
				avatarUrl: info?.avatarUrl || null,
				deliveries: data.deliveries,
				success: data.success,
				failed: data.failed,
				retries: data.retries,
				avgDurationMs: data.avgDurationMs,
				successRate: data.deliveries > 0 ? (data.success / data.deliveries) * 100 : 100,
			};
		})
		.sort((a, b) => b.deliveries - a.deliveries);

	// Error type data for chart
	const errorTypeData = Object.entries(metrics?.byErrorType ?? {})
		.map(([errorType, count]) => ({
			name: formatErrorType(errorType),
			value: count,
		}))
		.sort((a, b) => b.value - a.value)
		.slice(0, 10);

	// Instance chart data
	const instanceChartData = instanceDeliveryData.slice(0, 10).map((inst) => ({
		name: inst.name.length > 12 ? inst.name.slice(0, 12) + "..." : inst.name,
		success: inst.success,
		failed: inst.failed,
		retries: inst.retries,
	}));

	// Show info message when no transport data is available
	if (!isLoading && !hasTransportData) {
		return (
			<Card>
				<CardContent className="py-12">
					<div className="text-center space-y-4">
						<div className="mx-auto w-14 h-14 rounded-full bg-blue-500/10 flex items-center justify-center">
							<svg
								xmlns="http://www.w3.org/2000/svg"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								strokeWidth="2"
								strokeLinecap="round"
								strokeLinejoin="round"
								className="w-7 h-7 text-blue-600 dark:text-blue-400"
							>
								<path d="M22 12h-4l-3 9L9 3l-3 9H2" />
							</svg>
						</div>
						<div>
							<h3 className="text-lg font-semibold">Transport Metrics Not Available</h3>
							<p className="text-sm text-muted-foreground mt-2 max-w-lg mx-auto">
								The transport/webhook delivery system is not reporting metrics yet.
								Metrics will appear once webhook deliveries start occurring.
							</p>
						</div>
						<div className="pt-2 space-y-2">
							<p className="text-xs text-muted-foreground font-medium">Expected metrics:</p>
							<div className="flex flex-wrap justify-center gap-2">
								{[
									"transport_deliveries_total",
									"transport_duration_seconds",
									"transport_errors_total",
									"transport_retries_total",
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
					title="Total Deliveries"
					value={formatNumber(metrics?.totalDeliveries ?? 0)}
					subtitle="Webhook calls"
					isLoading={isLoading}
				/>
				<SummaryCard
					title="Success Rate"
					value={formatPercentage(metrics?.successRate ?? 100)}
					subtitle="Delivery success"
					status={
						metrics && metrics.successRate < 95
							? metrics.successRate < 80
								? "critical"
								: "warning"
							: undefined
					}
					isLoading={isLoading}
				/>
				<SummaryCard
					title="Avg Latency"
					value={formatDuration(metrics?.avgDurationMs ?? 0)}
					subtitle="Response time"
					isLoading={isLoading}
				/>
				<SummaryCard
					title="Total Retries"
					value={formatNumber(metrics?.totalRetries ?? 0)}
					subtitle="Retry attempts"
					status={
						metrics && metrics.totalRetries > 0 ? "warning" : undefined
					}
					isLoading={isLoading}
				/>
			</div>

			{/* Delivery Status Breakdown */}
			<Card>
				<CardHeader className="pb-2">
					<CardTitle className="text-base font-medium">
						Delivery Status
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					{isLoading ? (
						<Skeleton className="h-8 w-full" />
					) : (
						<>
							<div className="grid grid-cols-3 gap-4 text-center">
								<div>
									<StatusIndicator status="healthy" size="sm" />
									<p className="text-2xl font-bold tabular-nums text-emerald-600 dark:text-emerald-400">
										{formatNumber(metrics?.successfulDeliveries ?? 0)}
									</p>
									<p className="text-xs text-muted-foreground">Successful</p>
								</div>
								<div>
									<StatusIndicator
										status={
											metrics && metrics.failedDeliveries > 0
												? "critical"
												: "healthy"
										}
										size="sm"
									/>
									<p
										className={cn(
											"text-2xl font-bold tabular-nums",
											metrics && metrics.failedDeliveries > 0
												? "text-red-600 dark:text-red-400"
												: "text-muted-foreground",
										)}
									>
										{formatNumber(metrics?.failedDeliveries ?? 0)}
									</p>
									<p className="text-xs text-muted-foreground">Failed</p>
								</div>
								<div>
									<StatusIndicator
										status={
											metrics && metrics.totalRetries > 0
												? "warning"
												: "healthy"
										}
										size="sm"
									/>
									<p
										className={cn(
											"text-2xl font-bold tabular-nums",
											metrics && metrics.totalRetries > 0
												? "text-amber-600 dark:text-amber-400"
												: "text-muted-foreground",
										)}
									>
										{formatNumber(metrics?.totalRetries ?? 0)}
									</p>
									<p className="text-xs text-muted-foreground">Retries</p>
								</div>
							</div>

							<ProgressBar
								value={metrics?.successfulDeliveries ?? 0}
								max={Math.max(metrics?.totalDeliveries ?? 0, 1)}
								label="Success Rate"
								status={
									metrics && metrics.totalDeliveries > 0
										? metrics.successRate >= 95
											? "healthy"
											: metrics.successRate >= 80
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
						Latency Distribution
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
				{/* Deliveries by Instance */}
				<MetricChart
					type="bar"
					title="Deliveries by Instance"
					data={instanceChartData}
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
						{
							key: "retries",
							color: TAILWIND_CHART_COLORS.warning,
							label: "Retries",
						},
					]}
					height={300}
					isLoading={isLoading}
				/>

				{/* Errors by Type */}
				<HorizontalBarChart
					title="Errors by Type"
					data={errorTypeData}
					maxItems={10}
					color={TAILWIND_CHART_COLORS.error}
					isLoading={isLoading}
				/>
			</div>

			{/* Health Gauges */}
			<MetricGaugeGroup
				title="Transport Health"
				gauges={[
					{
						value: metrics?.successRate ?? 100,
						label: "Success Rate",
						size: "sm",
						status:
							metrics && metrics.successRate < 95
								? metrics.successRate < 80
									? "critical"
									: "warning"
								: "healthy",
					},
					{
						value:
							metrics && metrics.totalDeliveries > 0
								? Math.min(
										(metrics.totalRetries / metrics.totalDeliveries) * 100,
										100,
									)
								: 0,
						label: "Retry Rate",
						size: "sm",
						thresholds: { warning: 5, critical: 20, unit: "%" },
					},
					{
						value:
							metrics && metrics.totalDeliveries > 0
								? Math.min(
										(metrics.failedDeliveries / metrics.totalDeliveries) * 100,
										100,
									)
								: 0,
						label: "Failure Rate",
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
						Transport by Instance
					</CardTitle>
				</CardHeader>
				<CardContent>
					{isLoading ? (
						<div className="space-y-3">
							{Array.from({ length: 3 }).map((_, i) => (
								<Skeleton key={i} className="h-16 w-full" />
							))}
						</div>
					) : instanceDeliveryData.length === 0 ? (
						<p className="text-center text-sm text-muted-foreground py-8">
							No instance transport data available
						</p>
					) : (
						<div className="space-y-3">
							{instanceDeliveryData.map((instance) => (
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
											<p className="text-xs text-muted-foreground">Total</p>
											<p className="font-semibold tabular-nums">
												{instance.deliveries.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Success</p>
											<p className="font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">
												{instance.success.toLocaleString()}
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
										<div>
											<p className="text-xs text-muted-foreground">Retries</p>
											<p className={cn(
												"font-semibold tabular-nums",
												instance.retries > 0 ? "text-amber-600 dark:text-amber-400" : "text-muted-foreground"
											)}>
												{instance.retries.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Latency</p>
											<p className="font-semibold tabular-nums">
												{formatDuration(instance.avgDurationMs)}
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
 * Format error type for display
 */
function formatErrorType(errorType: string): string {
	return errorType
		.replace(/_/g, " ")
		.replace(/\b\w/g, (l) => l.toUpperCase());
}
