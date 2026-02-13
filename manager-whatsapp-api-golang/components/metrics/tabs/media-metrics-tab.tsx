/**
 * Media Metrics Tab Component
 *
 * Media processing metrics and storage monitoring.
 *
 * @module components/metrics/tabs/media-metrics-tab
 */

"use client";

import { Phone } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useInstanceNames } from "@/hooks/use-instance-names";
import {
	formatBytes,
	formatDuration,
	formatNumber,
	TAILWIND_CHART_COLORS,
} from "@/lib/metrics/constants";
import { formatPhoneNumber } from "@/lib/phone";
import { cn } from "@/lib/utils";
import type { MediaMetrics } from "@/types/metrics";
import { HorizontalBarChart, MetricChart } from "../metric-chart";
import { MetricGaugeGroup } from "../metric-gauge";
import { StatusIndicator } from "../status-indicator";

export interface MediaMetricsTabProps {
	metrics?: MediaMetrics;
	isLoading?: boolean;
}

export function MediaMetricsTab({
	metrics,
	isLoading = false,
}: MediaMetricsTabProps) {
	const { getDisplayName, getInstanceInfo } = useInstanceNames();

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

	// Instance chart data with friendly names
	const instanceChartData = Object.entries(metrics?.byInstance ?? {})
		.map(([instanceId, data]) => {
			const info = getInstanceInfo(instanceId);
			return {
				name: info?.name || getDisplayName(instanceId),
				value: data.downloads + data.uploads,
			};
		})
		.sort((a, b) => b.value - a.value)
		.slice(0, 10);

	// Instance data with full info
	const instanceData = Object.entries(metrics?.byInstance ?? {})
		.map(([instanceId, data]) => {
			const info = getInstanceInfo(instanceId);
			return {
				instanceId,
				name: info?.name || getDisplayName(instanceId),
				phone: info?.phone || null,
				avatarUrl: info?.avatarUrl || null,
				downloads: data.downloads,
				uploads: data.uploads,
				failures: data.failures,
			};
		})
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
					data={instanceChartData}
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

			{/* Instance Details */}
			<Card>
				<CardHeader className="pb-2">
					<CardTitle className="text-base font-medium">
						Media by Instance
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
							No instance media data available
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
									<div className="grid grid-cols-3 gap-4 sm:gap-6 text-center flex-1">
										<div>
											<p className="text-xs text-muted-foreground">Downloads</p>
											<p className="font-semibold tabular-nums text-blue-600 dark:text-blue-400">
												{instance.downloads.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Uploads</p>
											<p className="font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">
												{instance.uploads.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Failures</p>
											<p className={cn(
												"font-semibold tabular-nums",
												instance.failures > 0 ? "text-red-600 dark:text-red-400" : "text-muted-foreground"
											)}>
												{instance.failures.toLocaleString()}
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
