/**
 * Media Metrics Tab Component
 *
 * Media processing metrics and storage monitoring.
 *
 * @module components/metrics/tabs/media-metrics-tab
 */

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
	formatBytes,
	formatDuration,
	formatNumber,
	TAILWIND_CHART_COLORS,
} from "@/lib/metrics/constants";
import { cn } from "@/lib/utils";
import type { MediaMetrics } from "@/types/metrics";
import { HorizontalBarChart, MetricChart } from "../metric-chart";
import { MetricGaugeGroup } from "../metric-gauge";
import { MetricTable, NumberCell } from "../metric-table";
import { StatusIndicator } from "../status-indicator";

export interface MediaMetricsTabProps {
	metrics?: MediaMetrics;
	isLoading?: boolean;
}

export function MediaMetricsTab({
	metrics,
	isLoading = false,
}: MediaMetricsTabProps) {
	// Operation success rates
	const downloadSuccessRate =
		metrics && metrics.downloads.total > 0
			? (metrics.downloads.success / metrics.downloads.total) * 100
			: 100;

	const uploadSuccessRate =
		metrics && metrics.uploads.total > 0
			? (metrics.uploads.success / metrics.uploads.total) * 100
			: 100;

	// Media type data
	const mediaTypeData = Object.entries(metrics?.byType ?? {})
		.map(([type, data]) => ({
			name: type,
			downloads: data.downloads,
			uploads: data.uploads,
			failures: data.failures,
		}))
		.sort((a, b) => b.downloads + b.uploads - (a.downloads + a.uploads));

	// Instance data
	const instanceData = Object.entries(metrics?.byInstance ?? {})
		.map(([instanceId, data]) => ({
			name: instanceId.slice(0, 8) + "...",
			value: data.downloads + data.uploads,
		}))
		.sort((a, b) => b.value - a.value)
		.slice(0, 10);

	// Instance table data
	const instanceTableData = Object.entries(metrics?.byInstance ?? {})
		.map(([instanceId, data]) => ({
			instanceId,
			downloads: data.downloads,
			uploads: data.uploads,
			failures: data.failures,
		}))
		.sort((a, b) => b.downloads + b.uploads - (a.downloads + a.uploads));

	return (
		<div className="space-y-6">
			{/* Summary Cards */}
			<div className="grid gap-4 md:grid-cols-4">
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-sm font-medium text-muted-foreground">
							Downloads
						</CardTitle>
					</CardHeader>
					<CardContent>
						{isLoading ? (
							<Skeleton className="h-8 w-20" />
						) : (
							<div className="flex items-center gap-2">
								<StatusIndicator
									status={
										metrics && metrics.downloads.failed > 0
											? "warning"
											: "healthy"
									}
									size="sm"
								/>
								<span className="text-2xl font-bold tabular-nums">
									{formatNumber(metrics?.downloads.total ?? 0)}
								</span>
							</div>
						)}
						<p className="text-xs text-muted-foreground">
							{formatNumber(metrics?.downloads.success ?? 0)} successful
						</p>
					</CardContent>
				</Card>

				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-sm font-medium text-muted-foreground">
							Uploads
						</CardTitle>
					</CardHeader>
					<CardContent>
						{isLoading ? (
							<Skeleton className="h-8 w-20" />
						) : (
							<div className="flex items-center gap-2">
								<StatusIndicator
									status={
										metrics && metrics.uploads.failed > 0
											? "warning"
											: "healthy"
									}
									size="sm"
								/>
								<span className="text-2xl font-bold tabular-nums">
									{formatNumber(metrics?.uploads.total ?? 0)}
								</span>
							</div>
						)}
						<p className="text-xs text-muted-foreground">
							{formatNumber(metrics?.uploads.success ?? 0)} successful
						</p>
					</CardContent>
				</Card>

				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-sm font-medium text-muted-foreground">
							Backlog
						</CardTitle>
					</CardHeader>
					<CardContent>
						{isLoading ? (
							<Skeleton className="h-8 w-20" />
						) : (
							<div className="flex items-center gap-2">
								<StatusIndicator
									status={
										metrics && metrics.backlog > 50
											? "warning"
											: "healthy"
									}
									size="sm"
								/>
								<span
									className={cn(
										"text-2xl font-bold tabular-nums",
										metrics && metrics.backlog > 50
											? "text-amber-600 dark:text-amber-400"
											: "",
									)}
								>
									{formatNumber(metrics?.backlog ?? 0)}
								</span>
							</div>
						)}
						<p className="text-xs text-muted-foreground">Pending items</p>
					</CardContent>
				</Card>

				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-sm font-medium text-muted-foreground">
							Local Storage
						</CardTitle>
					</CardHeader>
					<CardContent>
						{isLoading ? (
							<Skeleton className="h-8 w-20" />
						) : (
							<span className="text-2xl font-bold tabular-nums">
								{formatBytes(metrics?.localStorageBytes ?? 0)}
							</span>
						)}
						<p className="text-xs text-muted-foreground">
							{formatNumber(metrics?.localStorageFiles ?? 0)} files
						</p>
					</CardContent>
				</Card>
			</div>

			{/* Success Rate Gauges */}
			<MetricGaugeGroup
				title="Operation Success Rates"
				gauges={[
					{
						value: downloadSuccessRate,
						label: "Download Success",
						size: "md",
						status:
							downloadSuccessRate >= 95
								? "healthy"
								: downloadSuccessRate >= 80
									? "warning"
									: "critical",
					},
					{
						value: uploadSuccessRate,
						label: "Upload Success",
						size: "md",
						status:
							uploadSuccessRate >= 95
								? "healthy"
								: uploadSuccessRate >= 80
									? "warning"
									: "critical",
					},
				]}
				isLoading={isLoading}
			/>

			{/* Performance Metrics */}
			<div className="grid gap-4 md:grid-cols-2">
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-base font-medium">
							Download Latency
						</CardTitle>
					</CardHeader>
					<CardContent>
						{isLoading ? (
							<Skeleton className="h-12 w-24" />
						) : (
							<p className="text-3xl font-bold tabular-nums">
								{formatDuration(metrics?.avgDownloadMs ?? 0)}
							</p>
						)}
						<p className="text-xs text-muted-foreground">Average download time</p>
					</CardContent>
				</Card>

				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-base font-medium">
							Upload Latency
						</CardTitle>
					</CardHeader>
					<CardContent>
						{isLoading ? (
							<Skeleton className="h-12 w-24" />
						) : (
							<p className="text-3xl font-bold tabular-nums">
								{formatDuration(metrics?.avgUploadMs ?? 0)}
							</p>
						)}
						<p className="text-xs text-muted-foreground">Average upload time</p>
					</CardContent>
				</Card>
			</div>

			{/* Charts Grid */}
			<div className="grid gap-4 md:grid-cols-2">
				{/* Media by Type */}
				<MetricChart
					type="bar"
					title="Operations by Media Type"
					data={mediaTypeData}
					xKey="name"
					yKeys={[
						{
							key: "downloads",
							color: TAILWIND_CHART_COLORS.primary,
							label: "Downloads",
						},
						{
							key: "uploads",
							color: TAILWIND_CHART_COLORS.success,
							label: "Uploads",
						},
						{
							key: "failures",
							color: TAILWIND_CHART_COLORS.error,
							label: "Failures",
						},
					]}
					height={300}
					isLoading={isLoading}
				/>

				{/* Operations by Instance */}
				<HorizontalBarChart
					title="Operations by Instance"
					data={instanceData}
					maxItems={10}
					color={TAILWIND_CHART_COLORS.tertiary}
					isLoading={isLoading}
				/>
			</div>

			{/* Cleanup Stats */}
			<Card>
				<CardHeader className="pb-2">
					<CardTitle className="text-base font-medium">
						Cleanup Statistics
					</CardTitle>
				</CardHeader>
				<CardContent>
					{isLoading ? (
						<div className="grid gap-4 md:grid-cols-2">
							<Skeleton className="h-12 w-full" />
							<Skeleton className="h-12 w-full" />
						</div>
					) : (
						<div className="grid gap-4 md:grid-cols-2">
							<div>
								<p className="text-sm text-muted-foreground">Cleanup Runs</p>
								<p className="text-xl font-bold tabular-nums">
									{formatNumber(metrics?.cleanupRuns ?? 0)}
								</p>
							</div>
							<div>
								<p className="text-sm text-muted-foreground">Bytes Cleaned</p>
								<p className="text-xl font-bold tabular-nums">
									{formatBytes(metrics?.cleanupDeletedBytes ?? 0)}
								</p>
							</div>
						</div>
					)}
				</CardContent>
			</Card>

			{/* Instance Details Table */}
			<MetricTable
				title="Media by Instance"
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
						key: "downloads",
						header: "Downloads",
						align: "right",
						format: (v) => <NumberCell value={Number(v)} />,
					},
					{
						key: "uploads",
						header: "Uploads",
						align: "right",
						format: (v) => <NumberCell value={Number(v)} />,
					},
					{
						key: "failures",
						header: "Failures",
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
				emptyMessage="No instance media data available"
			/>
		</div>
	);
}
