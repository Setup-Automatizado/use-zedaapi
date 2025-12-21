/**
 * Event Metrics Tab Component
 *
 * Event pipeline metrics and monitoring.
 *
 * @module components/metrics/tabs/event-metrics-tab
 */

"use client";

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
import type { EventMetrics, HealthLevel } from "@/types/metrics";
import { HorizontalBarChart, MetricChart } from "../metric-chart";
import { ProgressBar } from "../metric-gauge";
import { MetricTable, NumberCell } from "../metric-table";
import { StatusIndicator } from "../status-indicator";

// Event type friendly names and colors (24 event types from backend)
const EVENT_TYPE_CONFIG: Record<string, { label: string; color: string }> = {
	// Core messaging events
	message: { label: "Messages", color: "#3b82f6" }, // blue-500
	receipt: { label: "Receipts", color: "#10b981" }, // emerald-500
	undecryptable: { label: "Undecryptable", color: "#f87171" }, // red-400
	history_sync: { label: "History Sync", color: "#a78bfa" }, // violet-400

	// Connection events
	connected: { label: "Connected", color: "#22c55e" }, // green-500
	disconnected: { label: "Disconnected", color: "#ef4444" }, // red-500

	// Presence events
	presence: { label: "Presence", color: "#8b5cf6" }, // violet-500
	chat_presence: { label: "Chat Presence", color: "#a855f7" }, // purple-500

	// Profile events
	picture: { label: "Pictures", color: "#f59e0b" }, // amber-500
	push_name: { label: "Push Name", color: "#fb923c" }, // orange-400
	business_name: { label: "Business Name", color: "#06b6d4" }, // cyan-500
	user_about: { label: "User About", color: "#ec4899" }, // pink-500

	// Group events
	group_info: { label: "Group Info", color: "#6366f1" }, // indigo-500
	group_joined: { label: "Group Joined", color: "#818cf8" }, // indigo-400

	// Call events
	call_offer: { label: "Call Offer", color: "#14b8a6" }, // teal-500
	call_offer_notice: { label: "Call Notice", color: "#2dd4bf" }, // teal-400
	call_relay_latency: { label: "Call Latency", color: "#5eead4" }, // teal-300
	call_transport: { label: "Call Transport", color: "#99f6e4" }, // teal-200
	call_terminate: { label: "Call End", color: "#f97316" }, // orange-500
	call_reject: { label: "Call Reject", color: "#fb7185" }, // rose-400

	// Newsletter events
	newsletter_join: { label: "Newsletter Join", color: "#0ea5e9" }, // sky-500
	newsletter_leave: { label: "Newsletter Leave", color: "#38bdf8" }, // sky-400
	newsletter_mute_change: { label: "Newsletter Mute", color: "#7dd3fc" }, // sky-300
	newsletter_live_update: { label: "Newsletter Update", color: "#0284c7" }, // sky-600
};

export interface EventMetricsTabProps {
	metrics?: EventMetrics;
	isLoading?: boolean;
}

export function EventMetricsTab({
	metrics,
	isLoading = false,
}: EventMetricsTabProps) {
	// Get instance names for friendly display
	const { getInstanceInfo } = useInstanceNames();

	// Helper to get friendly event type name
	const getEventTypeName = (type: string) => {
		return EVENT_TYPE_CONFIG[type]?.label || type.replace(/_/g, " ").replace(/\b\w/g, l => l.toUpperCase());
	};

	// Event type data for charts with friendly names
	const eventTypeData = Object.entries(metrics?.byType ?? {}).map(
		([type, data]) => ({
			name: getEventTypeName(type),
			originalType: type,
			captured: data.captured,
			processed: data.processed,
			delivered: data.delivered,
			failed: data.failed,
		}),
	);

	// Instance backlog data with friendly names
	const instanceBacklogData = Object.entries(metrics?.byInstance ?? {})
		.map(([instanceId, data]) => {
			const info = getInstanceInfo(instanceId);
			return {
				name: info?.name || instanceId.slice(0, 8) + "...",
				fullId: instanceId,
				value: data.backlog,
			};
		})
		.sort((a, b) => b.value - a.value)
		.slice(0, 10);

	// Instance table data with full info
	const instanceTableData = Object.entries(metrics?.byInstance ?? {})
		.map(([instanceId, data]) => {
			const info = getInstanceInfo(instanceId);
			return {
				instanceId,
				name: info?.name || null,
				phone: info?.phone || null,
				avatarUrl: info?.avatarUrl || null,
				captured: data.captured,
				processed: data.processed,
				delivered: data.delivered,
				failed: data.failed,
				backlog: data.backlog,
			};
		})
		.sort((a, b) => b.captured - a.captured);

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

			{/* Instance Details Card */}
			<Card>
				<CardHeader className="pb-2">
					<CardTitle className="text-base font-medium">Events by Instance</CardTitle>
				</CardHeader>
				<CardContent>
					{isLoading ? (
						<div className="space-y-3">
							{Array.from({ length: 3 }).map((_, i) => (
								<Skeleton key={i} className="h-16 w-full" />
							))}
						</div>
					) : instanceTableData.length === 0 ? (
						<p className="text-center text-sm text-muted-foreground py-8">
							No instance data available
						</p>
					) : (
						<div className="space-y-3">
							{instanceTableData.map((instance) => (
								<div
									key={instance.instanceId}
									className="flex items-center gap-4 rounded-lg border p-4 transition-colors hover:bg-muted/50"
								>
									{/* Avatar */}
									<Avatar className="h-10 w-10 shrink-0">
										{instance.avatarUrl && (
											<AvatarImage src={instance.avatarUrl} alt={instance.name || ""} />
										)}
										<AvatarFallback className="bg-primary/10 text-primary text-sm font-medium">
											{instance.name?.slice(0, 2).toUpperCase() || instance.instanceId.slice(0, 2).toUpperCase()}
										</AvatarFallback>
									</Avatar>

									{/* Name & Phone */}
									<div className="min-w-0 flex-1">
										<p className="font-medium truncate">
											{instance.name || instance.instanceId.slice(0, 12) + "..."}
										</p>
										{instance.phone && (
											<p className="text-xs text-muted-foreground font-mono">
												{formatPhoneNumber(instance.phone)}
											</p>
										)}
									</div>

									{/* Stats Grid */}
									<div className="grid grid-cols-5 gap-4 text-center shrink-0">
										<div>
											<p className="text-xs text-muted-foreground">Captured</p>
											<p className="font-semibold tabular-nums text-blue-600 dark:text-blue-400">
												{instance.captured.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Processed</p>
											<p className="font-semibold tabular-nums">
												{instance.processed.toLocaleString()}
											</p>
										</div>
										<div>
											<p className="text-xs text-muted-foreground">Delivered</p>
											<p className="font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">
												{instance.delivered.toLocaleString()}
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
											<p className="text-xs text-muted-foreground">Backlog</p>
											<p className={cn(
												"font-semibold tabular-nums",
												instance.backlog > 50 ? "text-amber-600 dark:text-amber-400" : "text-muted-foreground"
											)}>
												{instance.backlog.toLocaleString()}
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
