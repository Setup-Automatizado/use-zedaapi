/**
 * Metric KPI Card Component
 *
 * Displays a single key performance indicator with icon and status.
 *
 * @module components/metrics/metric-kpi-card
 */

import type { LucideIcon } from "lucide-react";
import { ArrowDown, ArrowUp, Minus } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { HEALTH_COLORS } from "@/lib/metrics/constants";
import { cn } from "@/lib/utils";
import type { HealthLevel, TrendData } from "@/types/metrics";
import { StatusIndicator } from "./status-indicator";

export interface MetricKPICardProps {
	/** Card title */
	title: string;
	/** Main value to display */
	value: number | string;
	/** Unit suffix (e.g., "ms", "%", "/s") */
	unit?: string;
	/** Icon component */
	icon: LucideIcon;
	/** Health status */
	status?: HealthLevel;
	/** Trend data */
	trend?: TrendData;
	/** Subtitle text */
	subtitle?: string;
	/** Loading state */
	isLoading?: boolean;
	/** Additional CSS classes */
	className?: string;
}

export function MetricKPICard({
	title,
	value,
	unit,
	icon: Icon,
	status = "healthy",
	trend,
	subtitle,
	isLoading = false,
	className,
}: MetricKPICardProps) {
	const colors = HEALTH_COLORS[status];

	if (isLoading) {
		return (
			<Card className={cn("relative overflow-hidden", className)}>
				<CardContent className="p-5">
					<div className="flex items-center justify-between">
						<div className="space-y-2">
							<Skeleton className="h-4 w-24" />
							<Skeleton className="h-8 w-16" />
							{subtitle && <Skeleton className="h-3 w-20" />}
						</div>
						<Skeleton className="h-12 w-12 rounded-full" />
					</div>
				</CardContent>
			</Card>
		);
	}

	return (
		<Card className={cn("relative overflow-hidden", className)}>
			<CardContent className="p-5">
				<div className="flex items-center justify-between">
					<div className="space-y-1">
						<div className="flex items-center gap-2">
							<p className="text-sm text-muted-foreground">{title}</p>
							<StatusIndicator status={status} size="sm" />
						</div>
						<div className="flex items-baseline gap-1">
							<p className="text-3xl font-bold tracking-tight tabular-nums">
								{value}
							</p>
							{unit && (
								<span className="text-lg text-muted-foreground">{unit}</span>
							)}
						</div>
						{(subtitle || trend) && (
							<div className="flex items-center gap-2 text-xs text-muted-foreground">
								{trend && <TrendIndicator trend={trend} />}
								{subtitle && <span>{subtitle}</span>}
							</div>
						)}
					</div>
					<div
						className={cn(
							"flex h-12 w-12 items-center justify-center rounded-full",
							colors.bg,
						)}
					>
						<Icon className={cn("h-6 w-6", colors.text)} />
					</div>
				</div>
			</CardContent>
		</Card>
	);
}

/**
 * Trend Indicator
 */
function TrendIndicator({ trend }: { trend: TrendData }) {
	const { direction, value, isPositive } = trend;

	const Icon = direction === "up" ? ArrowUp : direction === "down" ? ArrowDown : Minus;

	const colorClass =
		direction === "stable"
			? "text-muted-foreground"
			: isPositive
				? "text-emerald-600 dark:text-emerald-400"
				: "text-red-600 dark:text-red-400";

	return (
		<span className={cn("inline-flex items-center gap-0.5", colorClass)}>
			<Icon className="h-3 w-3" />
			<span>{value.toFixed(1)}%</span>
		</span>
	);
}

/**
 * Compact KPI Card variant
 */
export interface MetricKPICardCompactProps {
	title: string;
	value: number | string;
	unit?: string;
	status?: HealthLevel;
	isLoading?: boolean;
	className?: string;
}

export function MetricKPICardCompact({
	title,
	value,
	unit,
	status = "healthy",
	isLoading = false,
	className,
}: MetricKPICardCompactProps) {
	const colors = HEALTH_COLORS[status];

	if (isLoading) {
		return (
			<div className={cn("space-y-1", className)}>
				<Skeleton className="h-3 w-16" />
				<Skeleton className="h-6 w-12" />
			</div>
		);
	}

	return (
		<div className={cn("space-y-1", className)}>
			<div className="flex items-center gap-1.5">
				<StatusIndicator status={status} size="sm" />
				<span className="text-xs text-muted-foreground">{title}</span>
			</div>
			<div className="flex items-baseline gap-0.5">
				<span className={cn("text-xl font-semibold tabular-nums", colors.text)}>
					{value}
				</span>
				{unit && (
					<span className="text-sm text-muted-foreground">{unit}</span>
				)}
			</div>
		</div>
	);
}
