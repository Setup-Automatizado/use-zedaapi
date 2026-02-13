/**
 * System Tab Component
 *
 * System health metrics including locks, circuit breakers, and workers.
 *
 * @module components/metrics/tabs/system-tab
 */

"use client";

import { AlertCircle, Phone, XCircle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
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
import type { CircuitBreakerState, HealthLevel, SystemMetrics, WorkerMetrics } from "@/types/metrics";
import { MetricChart } from "../metric-chart";
import { ProgressBar } from "../metric-gauge";
import { DurationCell, MetricTable, NumberCell } from "../metric-table";
import { StatusIndicator, StatusIndicatorWithLabel } from "../status-indicator";

export interface SystemTabProps {
	metrics?: SystemMetrics;
	workers?: WorkerMetrics;
	isLoading?: boolean;
}

export function SystemTab({ metrics, workers, isLoading = false }: SystemTabProps) {
	const { getDisplayName, getInstanceInfo } = useInstanceNames();

	// Circuit breaker instance data
	const circuitBreakerData = Object.entries(
		metrics?.circuitBreakerByInstance ?? {},
	).map(([instanceId, state]) => {
		const info = getInstanceInfo(instanceId);
		return {
			instanceId,
			displayName: info?.name || getDisplayName(instanceId),
			phone: info?.phone,
			avatarUrl: info?.avatarUrl,
			state: state as CircuitBreakerState,
		};
	});

	// Worker data - worker_type in metrics is actually instance_id
	const workerData = Object.entries(workers?.active ?? {}).map(
		([instanceId, count]) => {
			const info = getInstanceInfo(instanceId);
			return {
				instanceId,
				name: info?.name || getDisplayName(instanceId),
				phone: info?.phone || "",
				avatarUrl: info?.avatarUrl || "",
				active: count,
				errors: workers?.errors[instanceId] ?? 0,
				avgDuration: workers?.avgTaskDurationMs[instanceId] ?? 0,
			};
		},
	);

	// Health check data
	const healthCheckData = Object.entries(
		metrics?.healthChecks ?? {},
	).map(([component, data]) => ({
		component,
		healthy: data.healthy,
		unhealthy: data.unhealthy,
		degraded: data.degraded,
		status:
			data.unhealthy > 0
				? "critical"
				: data.degraded > 0
					? "warning"
					: ("healthy" as HealthLevel),
	}));

	return (
		<div className="space-y-6">
			{/* System Alerts */}
			{!isLoading && metrics && (
				<SystemAlerts
					splitBrain={metrics.splitBrainDetected}
					orphaned={metrics.orphanedInstances}
					circuitState={metrics.circuitBreakerState}
				/>
			)}

			{/* Circuit Breaker Status */}
			<Card>
				<CardHeader className="pb-2">
					<div className="flex items-center justify-between">
						<CardTitle className="text-base font-medium">
							Circuit Breaker Status
						</CardTitle>
						{!isLoading && metrics && (
							<CircuitBreakerBadge state={metrics.circuitBreakerState} />
						)}
					</div>
				</CardHeader>
				<CardContent>
					{isLoading ? (
						<Skeleton className="h-24 w-full" />
					) : circuitBreakerData.length > 0 ? (
						<div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
							{circuitBreakerData.map(({ instanceId, displayName, phone, avatarUrl, state }) => (
								<div
									key={instanceId}
									className="flex items-center gap-3 rounded-lg border p-3"
									title={instanceId}
								>
									<Avatar className="h-9 w-9 shrink-0">
										{avatarUrl ? (
											<AvatarImage src={avatarUrl} alt={displayName} />
										) : null}
										<AvatarFallback className="text-xs bg-muted">
											{displayName.slice(0, 2).toUpperCase()}
										</AvatarFallback>
									</Avatar>
									<div className="flex flex-col min-w-0 flex-1">
										<span className="text-sm font-medium truncate">
											{displayName}
										</span>
										{phone ? (
											<span className="text-xs text-muted-foreground flex items-center gap-1">
												<Phone className="h-3 w-3" />
												{formatPhoneNumber(phone)}
											</span>
										) : (
											<span className="text-xs text-muted-foreground font-mono truncate">
												{instanceId.slice(0, 8)}...
											</span>
										)}
									</div>
									<CircuitBreakerBadge state={state} size="sm" />
								</div>
							))}
						</div>
					) : (
						<p className="text-center text-sm text-muted-foreground py-4">
							No instance circuit breaker data available
						</p>
					)}
				</CardContent>
			</Card>

			{/* Lock Metrics */}
			<div className="grid gap-4 md:grid-cols-2">
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-base font-medium">
							Lock Acquisitions
						</CardTitle>
					</CardHeader>
					<CardContent className="space-y-4">
						{isLoading ? (
							<Skeleton className="h-20 w-full" />
						) : (
							<>
								<div className="grid grid-cols-2 gap-4">
									<div>
										<div className="flex items-center gap-2">
											<StatusIndicator status="healthy" size="sm" />
											<span className="text-sm text-muted-foreground">
												Successful
											</span>
										</div>
										<p className="text-2xl font-bold tabular-nums text-emerald-600 dark:text-emerald-400">
											{formatNumber(
												metrics?.lockAcquisitions.success ?? 0,
											)}
										</p>
									</div>
									<div>
										<div className="flex items-center gap-2">
											<StatusIndicator
												status={
													metrics && metrics.lockAcquisitions.failure > 0
														? "critical"
														: "healthy"
												}
												size="sm"
											/>
											<span className="text-sm text-muted-foreground">
												Failed
											</span>
										</div>
										<p
											className={cn(
												"text-2xl font-bold tabular-nums",
												metrics && metrics.lockAcquisitions.failure > 0
													? "text-red-600 dark:text-red-400"
													: "",
											)}
										>
											{formatNumber(
												metrics?.lockAcquisitions.failure ?? 0,
											)}
										</p>
									</div>
								</div>

								<ProgressBar
									value={metrics?.lockAcquisitions.success ?? 0}
									max={Math.max(
										(metrics?.lockAcquisitions.success ?? 0) +
											(metrics?.lockAcquisitions.failure ?? 0),
										1,
									)}
									label="Success Rate"
									status={
										metrics &&
										metrics.lockAcquisitions.failure /
											(metrics.lockAcquisitions.success +
												metrics.lockAcquisitions.failure) >
											0.1
											? "critical"
											: "healthy"
									}
								/>
							</>
						)}
					</CardContent>
				</Card>

				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-base font-medium">
							Lock Operations
						</CardTitle>
					</CardHeader>
					<CardContent>
						{isLoading ? (
							<Skeleton className="h-20 w-full" />
						) : (
							<div className="grid grid-cols-2 gap-4">
								<div>
									<p className="text-sm text-muted-foreground">
										Reacquisitions
									</p>
									<p className="text-2xl font-bold tabular-nums">
										{formatNumber(
											metrics?.lockAcquisitions.reacquisitions ?? 0,
										)}
									</p>
								</div>
								<div>
									<p className="text-sm text-muted-foreground">Fallbacks</p>
									<p
										className={cn(
											"text-2xl font-bold tabular-nums",
											metrics && metrics.lockAcquisitions.fallbacks > 0
												? "text-amber-600 dark:text-amber-400"
												: "",
										)}
									>
										{formatNumber(
											metrics?.lockAcquisitions.fallbacks ?? 0,
										)}
									</p>
								</div>
							</div>
						)}
					</CardContent>
				</Card>
			</div>

			{/* Reconciliation Stats */}
			<Card>
				<CardHeader className="pb-2">
					<CardTitle className="text-base font-medium">
						Reconciliation
					</CardTitle>
				</CardHeader>
				<CardContent>
					{isLoading ? (
						<Skeleton className="h-16 w-full" />
					) : (
						<div className="grid gap-4 md:grid-cols-5">
							<div>
								<StatusIndicatorWithLabel status="healthy" label="Success" />
								<p className="text-xl font-bold tabular-nums">
									{formatNumber(metrics?.reconciliation.success ?? 0)}
								</p>
							</div>
							<div>
								<StatusIndicatorWithLabel
									status={
										metrics && metrics.reconciliation.failure > 0
											? "critical"
											: "healthy"
									}
									label="Failure"
								/>
								<p
									className={cn(
										"text-xl font-bold tabular-nums",
										metrics && metrics.reconciliation.failure > 0
											? "text-red-600 dark:text-red-400"
											: "",
									)}
								>
									{formatNumber(metrics?.reconciliation.failure ?? 0)}
								</p>
							</div>
							<div>
								<p className="text-sm text-muted-foreground">Skipped</p>
								<p className="text-xl font-bold tabular-nums">
									{formatNumber(metrics?.reconciliation.skipped ?? 0)}
								</p>
							</div>
							<div>
								<p className="text-sm text-muted-foreground">Errors</p>
								<p
									className={cn(
										"text-xl font-bold tabular-nums",
										metrics && metrics.reconciliation.error > 0
											? "text-amber-600 dark:text-amber-400"
											: "",
									)}
								>
									{formatNumber(metrics?.reconciliation.error ?? 0)}
								</p>
							</div>
							<div>
								<p className="text-sm text-muted-foreground">Avg Duration</p>
								<p className="text-xl font-bold tabular-nums">
									{formatDuration(
										metrics?.reconciliation.avgDurationMs ?? 0,
									)}
								</p>
							</div>
						</div>
					)}
				</CardContent>
			</Card>

			{/* Workers */}
			<MetricChart
				type="bar"
				title="Active Workers"
				data={workerData}
				xKey="name"
				yKeys={[
					{
						key: "active",
						color: TAILWIND_CHART_COLORS.primary,
						label: "Active",
					},
					{
						key: "errors",
						color: TAILWIND_CHART_COLORS.error,
						label: "Errors",
					},
				]}
				height={250}
				isLoading={isLoading}
			/>

			{/* Worker Details Table */}
			<MetricTable
				title="Worker Details"
				data={workerData}
				columns={[
					{
						key: "name",
						header: "Instance",
						format: (_, row) => {
							const data = row as typeof workerData[0];
							return (
								<div className="flex items-center gap-2">
									<Avatar className="h-7 w-7">
										{data.avatarUrl ? (
											<AvatarImage src={data.avatarUrl} alt={data.name} />
										) : null}
										<AvatarFallback className="text-xs bg-muted">
											{data.name.slice(0, 2).toUpperCase()}
										</AvatarFallback>
									</Avatar>
									<div className="flex flex-col min-w-0">
										<span className="font-medium truncate">{data.name}</span>
										{data.phone && (
											<span className="text-xs text-muted-foreground flex items-center gap-1">
												<Phone className="h-3 w-3" />
												{formatPhoneNumber(data.phone)}
											</span>
										)}
									</div>
								</div>
							);
						},
					},
					{
						key: "active",
						header: "Active",
						align: "right",
						format: (v) => <NumberCell value={Number(v)} />,
					},
					{
						key: "errors",
						header: "Errors",
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
						key: "avgDuration",
						header: "Avg Task Duration",
						align: "right",
						format: (v) => <DurationCell ms={Number(v)} />,
					},
				]}
				isLoading={isLoading}
				emptyMessage="No worker data available"
			/>

			{/* Health Checks Table */}
			<MetricTable
				title="Health Check Results"
				data={healthCheckData}
				columns={[
					{
						key: "component",
						header: "Component",
						format: (v) => (
							<span className="capitalize">{String(v)}</span>
						),
					},
					{
						key: "status",
						header: "Status",
						format: (_, row) => (
							<div className="flex items-center gap-2">
								<StatusIndicator
									status={(row as { status: HealthLevel }).status}
									size="sm"
								/>
								<span className="capitalize">
									{(row as { status: HealthLevel }).status}
								</span>
							</div>
						),
					},
					{
						key: "healthy",
						header: "Healthy",
						align: "right",
						format: (v) => <NumberCell value={Number(v)} />,
					},
					{
						key: "degraded",
						header: "Degraded",
						align: "right",
						format: (v) => (
							<span
								className={cn(
									"tabular-nums",
									Number(v) > 0 && "text-amber-600 dark:text-amber-400",
								)}
							>
								{Number(v).toLocaleString()}
							</span>
						),
					},
					{
						key: "unhealthy",
						header: "Unhealthy",
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
				emptyMessage="No health check data available"
			/>
		</div>
	);
}

/**
 * System Alerts
 */
function SystemAlerts({
	splitBrain,
	orphaned,
	circuitState,
}: {
	splitBrain: number;
	orphaned: number;
	circuitState: CircuitBreakerState;
}) {
	const alerts: Array<{
		type: "error" | "warning";
		title: string;
		description: string;
	}> = [];

	if (splitBrain > 0) {
		alerts.push({
			type: "error",
			title: "Split-Brain Detected",
			description: `${splitBrain} split-brain event(s) detected. This may indicate distributed lock issues.`,
		});
	}

	if (circuitState === "open") {
		alerts.push({
			type: "error",
			title: "Circuit Breaker Open",
			description:
				"The circuit breaker is OPEN. The system is in protection mode.",
		});
	} else if (circuitState === "half-open") {
		alerts.push({
			type: "warning",
			title: "Circuit Breaker Half-Open",
			description:
				"The circuit breaker is testing recovery. Monitor closely.",
		});
	}

	if (orphaned > 0) {
		alerts.push({
			type: "warning",
			title: "Orphaned Instances",
			description: `${orphaned} instance(s) are orphaned in the database but not active.`,
		});
	}

	if (alerts.length === 0) return null;

	return (
		<div className="space-y-3">
			{alerts.map((alert, index) => (
				<Alert
					key={index}
					variant={alert.type === "error" ? "destructive" : "default"}
				>
					{alert.type === "error" ? (
						<XCircle className="h-4 w-4" />
					) : (
						<AlertCircle className="h-4 w-4" />
					)}
					<AlertTitle>{alert.title}</AlertTitle>
					<AlertDescription>{alert.description}</AlertDescription>
				</Alert>
			))}
		</div>
	);
}

/**
 * Circuit Breaker Badge
 */
function CircuitBreakerBadge({
	state,
	size = "default",
}: {
	state: CircuitBreakerState;
	size?: "default" | "sm";
}) {
	const variants: Record<CircuitBreakerState, { variant: "default" | "secondary" | "destructive" | "outline"; label: string }> = {
		closed: { variant: "default", label: "Closed" },
		"half-open": { variant: "secondary", label: "Half-Open" },
		open: { variant: "destructive", label: "Open" },
		unknown: { variant: "outline", label: "Unknown" },
	};

	const { variant, label } = variants[state] || variants.unknown;

	return (
		<Badge variant={variant} className={cn(size === "sm" && "text-xs px-1.5 py-0")}>
			{label}
		</Badge>
	);
}
