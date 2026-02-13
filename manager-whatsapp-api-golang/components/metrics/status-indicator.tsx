/**
 * Status Indicator Component
 *
 * Displays a colored dot indicating health status.
 *
 * @module components/metrics/status-indicator
 */

import { cn } from "@/lib/utils";
import type { HealthLevel } from "@/types/metrics";

export interface StatusIndicatorProps {
	/** Health status level */
	status: HealthLevel;
	/** Size variant */
	size?: "sm" | "md" | "lg";
	/** Show pulse animation */
	pulse?: boolean;
	/** Additional CSS classes */
	className?: string;
}

const sizeClasses = {
	sm: "h-2 w-2",
	md: "h-3 w-3",
	lg: "h-4 w-4",
};

const statusClasses: Record<HealthLevel, string> = {
	healthy: "bg-emerald-500",
	warning: "bg-amber-500",
	critical: "bg-red-500",
};

const pulseClasses: Record<HealthLevel, string> = {
	healthy: "animate-pulse bg-emerald-400",
	warning: "animate-pulse bg-amber-400",
	critical: "animate-pulse bg-red-400",
};

export function StatusIndicator({
	status,
	size = "md",
	pulse = false,
	className,
}: StatusIndicatorProps) {
	return (
		<span className={cn("relative inline-flex", className)}>
			{pulse && (
				<span
					className={cn(
						"absolute inline-flex h-full w-full rounded-full opacity-75",
						pulseClasses[status],
					)}
				/>
			)}
			<span
				className={cn(
					"relative inline-flex rounded-full",
					sizeClasses[size],
					statusClasses[status],
				)}
			/>
		</span>
	);
}

/**
 * Status Indicator with Label
 */
export interface StatusIndicatorWithLabelProps extends StatusIndicatorProps {
	/** Label text */
	label: string;
}

export function StatusIndicatorWithLabel({
	status,
	size = "md",
	pulse = false,
	label,
	className,
}: StatusIndicatorWithLabelProps) {
	const labelClasses: Record<HealthLevel, string> = {
		healthy: "text-emerald-600 dark:text-emerald-400",
		warning: "text-amber-600 dark:text-amber-400",
		critical: "text-red-600 dark:text-red-400",
	};

	return (
		<span className={cn("inline-flex items-center gap-2", className)}>
			<StatusIndicator status={status} size={size} pulse={pulse} />
			<span className={cn("text-sm font-medium", labelClasses[status])}>
				{label}
			</span>
		</span>
	);
}
