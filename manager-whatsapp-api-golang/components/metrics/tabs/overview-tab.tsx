/**
 * Overview Tab Component
 *
 * System overview with health summary and key metrics.
 *
 * @module components/metrics/tabs/overview-tab
 */

"use client";

import { AlertCircle, Phone, XCircle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useInstanceNames } from "@/hooks/use-instance-names";
import {
	formatBytes,
	formatNumber,
	METRIC_THRESHOLDS,
} from "@/lib/metrics/constants";
import type { DashboardMetrics, HealthLevel } from "@/types/metrics";
import { MetricGauge } from "../metric-gauge";
import { MetricsOverview } from "../metrics-overview";
import { StatusIndicator } from "../status-indicator";

export interface OverviewTabProps {
	metrics?: DashboardMetrics;
	isLoading?: boolean;
}

export function OverviewTab({ metrics, isLoading = false }: OverviewTabProps) {
	// Calculate overall system health
	const systemHealth = calculateSystemHealth(metrics);

	return (
		<div className="space-y-6">
			{/* System Health Alert */}
			{!isLoading && metrics && systemHealth !== "healthy" && (
				<SystemHealthAlert health={systemHealth} metrics={metrics} />
			)}

			{/* KPI Overview Grid */}
			<MetricsOverview metrics={metrics} isLoading={isLoading} />

			{/* Gauges Row */}
			<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
				<MetricGauge
					value={metrics?.http.errorRate ?? 0}
					label="Error Rate"
					thresholds={METRIC_THRESHOLDS.errorRate}
					isLoading={isLoading}
				/>
				<MetricGauge
					value={
						metrics
							? Math.min(
									(metrics.events.delivered /
										Math.max(metrics.events.captured, 1)) *
										100,
									100,
								)
							: 0
					}
					label="Event Delivery Rate"
					status={
						metrics && metrics.events.captured > 0
							? metrics.events.delivered / metrics.events.captured >= 0.95
								? "healthy"
								: metrics.events.delivered / metrics.events.captured >= 0.8
									? "warning"
									: "critical"
							: "healthy"
					}
					isLoading={isLoading}
				/>
				<MetricGauge
					value={
						metrics
							? Math.min(
									(metrics.media.uploads.success /
										Math.max(metrics.media.uploads.total, 1)) *
										100,
									100,
								)
							: 0
					}
					label="Media Upload Success"
					status={
						metrics && metrics.media.uploads.total > 0
							? metrics.media.uploads.success / metrics.media.uploads.total >=
								0.95
								? "healthy"
								: metrics.media.uploads.success /
											metrics.media.uploads.total >=
									  0.8
									? "warning"
									: "critical"
							: "healthy"
					}
					isLoading={isLoading}
				/>
				<MetricGauge
					value={
						metrics
							? Math.min(
									(metrics.system.lockAcquisitions.success /
										Math.max(
											metrics.system.lockAcquisitions.success +
												metrics.system.lockAcquisitions.failure,
											1,
										)) *
										100,
									100,
								)
							: 100
					}
					label="Lock Success Rate"
					status={
						metrics
							? metrics.system.lockAcquisitions.failure === 0
								? "healthy"
								: metrics.system.lockAcquisitions.failure < 5
									? "warning"
									: "critical"
							: "healthy"
					}
					isLoading={isLoading}
				/>
			</div>

			{/* Component Status Grid */}
			<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
				<ComponentStatusCard
					title="Events Pipeline"
					items={[
						{
							label: "Captured",
							value: formatNumber(metrics?.events.captured ?? 0),
						},
						{
							label: "Processed",
							value: formatNumber(metrics?.events.processed ?? 0),
						},
						{
							label: "Delivered",
							value: formatNumber(metrics?.events.delivered ?? 0),
						},
						{
							label: "Failed",
							value: formatNumber(metrics?.events.failed ?? 0),
							status:
								metrics && metrics.events.failed > 0 ? "warning" : "healthy",
						},
					]}
					isLoading={isLoading}
				/>

				<ComponentStatusCard
					title="Message Queue"
					items={[
						{
							label: "Pending",
							value: formatNumber(metrics?.messageQueue.pending ?? 0),
						},
						{
							label: "Processing",
							value: formatNumber(metrics?.messageQueue.processing ?? 0),
						},
						{
							label: "Sent",
							value: formatNumber(metrics?.messageQueue.sent ?? 0),
						},
						{
							label: "Failed",
							value: formatNumber(metrics?.messageQueue.failed ?? 0),
							status:
								metrics && metrics.messageQueue.failed > 0
									? "warning"
									: "healthy",
						},
					]}
					isLoading={isLoading}
				/>

				<ComponentStatusCard
					title="Media Processing"
					items={[
						{
							label: "Downloads",
							value: formatNumber(metrics?.media.downloads.total ?? 0),
						},
						{
							label: "Uploads",
							value: formatNumber(metrics?.media.uploads.total ?? 0),
						},
						{
							label: "Backlog",
							value: formatNumber(metrics?.media.backlog ?? 0),
							status:
								metrics && metrics.media.backlog > 50 ? "warning" : "healthy",
						},
						{
							label: "Storage",
							value: formatBytes(metrics?.media.localStorageBytes ?? 0),
						},
					]}
					isLoading={isLoading}
				/>
			</div>

			{/* Instance Summary */}
			{metrics && metrics.instances.length > 0 && (
				<InstanceSummaryCard
					instances={metrics.instances}
					circuitBreakerByInstance={metrics.system.circuitBreakerByInstance}
				/>
			)}
		</div>
	);
}

/**
 * Calculate overall system health
 */
function calculateSystemHealth(metrics?: DashboardMetrics): HealthLevel {
	if (!metrics) return "healthy";

	// Check critical conditions
	if (metrics.system.circuitBreakerState === "open") return "critical";
	if (metrics.http.errorRate > 5) return "critical";
	if (metrics.system.splitBrainDetected > 0) return "critical";

	// Check warning conditions
	if (metrics.system.circuitBreakerState === "half-open") return "warning";
	if (metrics.http.errorRate > 1) return "warning";
	if (metrics.messageQueue.failed > 0) return "warning";
	if (metrics.events.failed > 0) return "warning";

	return "healthy";
}

/**
 * System Health Alert
 */
function SystemHealthAlert({
	health,
	metrics,
}: {
	health: HealthLevel;
	metrics: DashboardMetrics;
}) {
	const issues: string[] = [];

	if (metrics.system.circuitBreakerState === "open") {
		issues.push("Circuit breaker is OPEN");
	}
	if (metrics.system.circuitBreakerState === "half-open") {
		issues.push("Circuit breaker is in half-open state");
	}
	if (metrics.http.errorRate > 1) {
		issues.push(`HTTP error rate is ${metrics.http.errorRate.toFixed(1)}%`);
	}
	if (metrics.system.splitBrainDetected > 0) {
		issues.push(
			`${metrics.system.splitBrainDetected} split-brain events detected`,
		);
	}
	if (metrics.events.failed > 0) {
		issues.push(`${metrics.events.failed} events failed`);
	}
	if (metrics.messageQueue.failed > 0) {
		issues.push(`${metrics.messageQueue.failed} messages failed`);
	}

	const variant = health === "critical" ? "destructive" : "default";
	const Icon = health === "critical" ? XCircle : AlertCircle;

	return (
		<Alert variant={variant}>
			<Icon className="h-4 w-4" />
			<AlertTitle>
				{health === "critical" ? "System Critical" : "System Degraded"}
			</AlertTitle>
			<AlertDescription>
				<ul className="mt-1 list-inside list-disc">
					{issues.map((issue, index) => (
						<li key={index}>{issue}</li>
					))}
				</ul>
			</AlertDescription>
		</Alert>
	);
}

/**
 * Component Status Card
 */
interface ComponentStatusItem {
	label: string;
	value: string;
	status?: HealthLevel;
}

function ComponentStatusCard({
	title,
	items,
	isLoading,
}: {
	title: string;
	items: ComponentStatusItem[];
	isLoading?: boolean;
}) {
	if (isLoading) {
		return (
			<Card>
				<CardHeader className="pb-2">
					<Skeleton className="h-5 w-32" />
				</CardHeader>
				<CardContent className="space-y-2">
					{Array.from({ length: 4 }).map((_, i) => (
						<Skeleton key={i} className="h-6 w-full" />
					))}
				</CardContent>
			</Card>
		);
	}

	return (
		<Card>
			<CardHeader className="pb-2">
				<CardTitle className="text-base font-medium">{title}</CardTitle>
			</CardHeader>
			<CardContent className="space-y-2">
				{items.map((item) => (
					<div key={item.label} className="flex items-center justify-between">
						<span className="text-sm text-muted-foreground">{item.label}</span>
						<div className="flex items-center gap-2">
							{item.status && item.status !== "healthy" && (
								<StatusIndicator status={item.status} size="sm" />
							)}
							<span className="font-medium tabular-nums">{item.value}</span>
						</div>
					</div>
				))}
			</CardContent>
		</Card>
	);
}

/**
 * Instance Summary Card with avatar, name, and phone
 */
function InstanceSummaryCard({
	instances,
	circuitBreakerByInstance,
}: {
	instances: string[];
	circuitBreakerByInstance: Record<string, string>;
}) {
	const { getDisplayName, getInstanceInfo, formatPhone } = useInstanceNames();

	return (
		<Card>
			<CardHeader className="pb-2">
				<CardTitle className="text-base font-medium">
					Active Instances ({instances.length})
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
					{instances.slice(0, 12).map((instanceId) => {
						const info = getInstanceInfo(instanceId);
						const displayName = info?.name || getDisplayName(instanceId);
						const circuitState = circuitBreakerByInstance[instanceId];
						const status: HealthLevel =
							circuitState === "open"
								? "critical"
								: circuitState === "half-open"
									? "warning"
									: "healthy";

						return (
							<div
								key={instanceId}
								className="flex items-center gap-3 rounded-lg border bg-card p-3 transition-colors hover:bg-muted/50"
								title={instanceId}
							>
								<Avatar className="h-10 w-10 shrink-0">
									{info?.avatarUrl ? (
										<AvatarImage src={info.avatarUrl} alt={displayName} />
									) : null}
									<AvatarFallback className="text-sm bg-muted">
										{displayName.slice(0, 2).toUpperCase()}
									</AvatarFallback>
								</Avatar>
								<div className="flex flex-col min-w-0 flex-1">
									<div className="flex items-center gap-2">
										<span className="text-sm font-medium truncate">
											{displayName}
										</span>
										<StatusIndicator status={status} size="sm" />
									</div>
									{info?.phone ? (
										<span className="text-xs text-muted-foreground flex items-center gap-1 truncate">
											<Phone className="h-3 w-3 shrink-0" />
											{formatPhone(info.phone)}
										</span>
									) : (
										<span className="text-xs text-muted-foreground font-mono truncate">
											{instanceId.slice(0, 8)}...
										</span>
									)}
								</div>
							</div>
						);
					})}
				</div>
				{instances.length > 12 && (
					<p className="mt-3 text-sm text-muted-foreground text-center">
						+{instances.length - 12} more instances
					</p>
				)}
			</CardContent>
		</Card>
	);
}
