/**
 * Metric Gauge Component
 *
 * Circular gauge for displaying percentage values.
 *
 * @module components/metrics/metric-gauge
 */

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { getHealthLevel, HEALTH_COLORS } from "@/lib/metrics/constants";
import { cn } from "@/lib/utils";
import type { HealthLevel, MetricThreshold } from "@/types/metrics";

export interface MetricGaugeProps {
	/** Current value (0-100 by default) */
	value: number;
	/** Maximum value */
	max?: number;
	/** Gauge label */
	label: string;
	/** Unit suffix */
	unit?: string;
	/** Gauge size */
	size?: "sm" | "md" | "lg";
	/** Health thresholds */
	thresholds?: MetricThreshold;
	/** Override health status */
	status?: HealthLevel;
	/** Show as card */
	showCard?: boolean;
	/** Loading state */
	isLoading?: boolean;
	/** Additional CSS classes */
	className?: string;
}

const sizeConfig = {
	sm: { size: 80, strokeWidth: 6, fontSize: "text-lg" },
	md: { size: 120, strokeWidth: 8, fontSize: "text-2xl" },
	lg: { size: 160, strokeWidth: 10, fontSize: "text-3xl" },
};

export function MetricGauge({
	value,
	max = 100,
	label,
	unit = "%",
	size = "md",
	thresholds,
	status: overrideStatus,
	showCard = true,
	isLoading = false,
	className,
}: MetricGaugeProps) {
	const config = sizeConfig[size];
	const normalizedValue = Math.min(Math.max(value, 0), max);
	const percentage = (normalizedValue / max) * 100;

	// Calculate status
	const status =
		overrideStatus ||
		(thresholds
			? getHealthLevel(value, thresholds)
			: percentage >= 90
				? "critical"
				: percentage >= 70
					? "warning"
					: "healthy");

	const colors = HEALTH_COLORS[status];

	// SVG calculations
	const radius = (config.size - config.strokeWidth) / 2;
	const circumference = radius * 2 * Math.PI;
	const offset = circumference - (percentage / 100) * circumference;

	const renderGaugeContent = () => {
		if (isLoading) {
			return (
				<div className="flex flex-col items-center gap-2">
					<Skeleton
						className="rounded-full"
						style={{ width: config.size, height: config.size }}
					/>
					<Skeleton className="h-4 w-20" />
				</div>
			);
		}

		return (
			<div className="flex flex-col items-center gap-2">
				<div className="relative" style={{ width: config.size, height: config.size }}>
					{/* Background circle */}
					<svg
						className="absolute transform -rotate-90"
						width={config.size}
						height={config.size}
					>
						<circle
							cx={config.size / 2}
							cy={config.size / 2}
							r={radius}
							fill="none"
							stroke="currentColor"
							strokeWidth={config.strokeWidth}
							className="text-muted"
						/>
					</svg>

					{/* Progress circle */}
					<svg
						className="absolute transform -rotate-90"
						width={config.size}
						height={config.size}
					>
						<circle
							cx={config.size / 2}
							cy={config.size / 2}
							r={radius}
							fill="none"
							stroke="currentColor"
							strokeWidth={config.strokeWidth}
							strokeDasharray={circumference}
							strokeDashoffset={offset}
							strokeLinecap="round"
							className={cn(
								"transition-all duration-500",
								status === "healthy" && "text-emerald-500",
								status === "warning" && "text-amber-500",
								status === "critical" && "text-red-500",
							)}
						/>
					</svg>

					{/* Center value */}
					<div className="absolute inset-0 flex items-center justify-center">
						<span className={cn("font-bold tabular-nums", config.fontSize, colors.text)}>
							{Math.round(normalizedValue)}
							{unit && <span className="text-sm opacity-70">{unit}</span>}
						</span>
					</div>
				</div>

				<span className="text-sm text-muted-foreground">{label}</span>
			</div>
		);
	};

	if (!showCard) {
		return (
			<div className={cn("inline-flex", className)}>
				{renderGaugeContent()}
			</div>
		);
	}

	return (
		<Card className={className}>
			<CardContent className="flex items-center justify-center py-6">
				{renderGaugeContent()}
			</CardContent>
		</Card>
	);
}

/**
 * Multiple gauges in a row
 */
export interface MetricGaugeGroupProps {
	gauges: Array<Omit<MetricGaugeProps, "showCard" | "className">>;
	title?: string;
	isLoading?: boolean;
	className?: string;
}

export function MetricGaugeGroup({
	gauges,
	title,
	isLoading = false,
	className,
}: MetricGaugeGroupProps) {
	return (
		<Card className={className}>
			{title && (
				<CardHeader className="pb-2">
					<CardTitle className="text-base font-medium">{title}</CardTitle>
				</CardHeader>
			)}
			<CardContent>
				<div className="flex flex-wrap items-center justify-around gap-4">
					{gauges.map((gauge, index) => (
						<MetricGauge
							key={index}
							{...gauge}
							showCard={false}
							isLoading={isLoading}
						/>
					))}
				</div>
			</CardContent>
		</Card>
	);
}

/**
 * Simple linear progress bar
 */
export interface ProgressBarProps {
	value: number;
	max?: number;
	label?: string;
	showValue?: boolean;
	status?: HealthLevel;
	className?: string;
}

export function ProgressBar({
	value,
	max = 100,
	label,
	showValue = true,
	status = "healthy",
	className,
}: ProgressBarProps) {
	const percentage = Math.min(Math.max((value / max) * 100, 0), 100);

	const bgColors: Record<HealthLevel, string> = {
		healthy: "bg-emerald-500",
		warning: "bg-amber-500",
		critical: "bg-red-500",
	};

	return (
		<div className={cn("space-y-1", className)}>
			{(label || showValue) && (
				<div className="flex items-center justify-between text-sm">
					{label && <span className="text-muted-foreground">{label}</span>}
					{showValue && (
						<span className="font-medium tabular-nums">
							{value.toLocaleString()} / {max.toLocaleString()}
						</span>
					)}
				</div>
			)}
			<div className="h-2 w-full overflow-hidden rounded-full bg-muted">
				<div
					className={cn(
						"h-full rounded-full transition-all duration-300",
						bgColors[status],
					)}
					style={{ width: `${percentage}%` }}
				/>
			</div>
		</div>
	);
}
