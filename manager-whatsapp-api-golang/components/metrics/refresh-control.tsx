/**
 * Refresh Control Component
 *
 * Controls for polling interval and manual refresh.
 *
 * @module components/metrics/refresh-control
 */

"use client";

import { RefreshCw } from "lucide-react";
import * as React from "react";
import { Button } from "@/components/ui/button";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { getRelativeTime, REFRESH_INTERVAL_OPTIONS } from "@/lib/metrics/constants";
import { cn } from "@/lib/utils";
import type { RefreshInterval } from "@/types/metrics";

export interface RefreshControlProps {
	/** Current polling interval */
	interval: RefreshInterval;
	/** Callback when interval changes */
	onIntervalChange: (interval: RefreshInterval) => void;
	/** Callback for manual refresh */
	onRefresh: () => void;
	/** Whether currently refreshing */
	isRefreshing: boolean;
	/** Last update timestamp */
	lastUpdated?: Date;
	/** Additional CSS classes */
	className?: string;
}

export function RefreshControl({
	interval,
	onIntervalChange,
	onRefresh,
	isRefreshing,
	lastUpdated,
	className,
}: RefreshControlProps) {
	const [relativeTime, setRelativeTime] = React.useState<string>("");

	// Update relative time every second
	React.useEffect(() => {
		if (!lastUpdated) return;

		const update = () => {
			setRelativeTime(getRelativeTime(lastUpdated));
		};

		update();
		const timer = setInterval(update, 1000);

		return () => clearInterval(timer);
	}, [lastUpdated]);

	return (
		<div className={cn("flex items-center gap-3", className)}>
			{/* Last Updated */}
			{lastUpdated && relativeTime && (
				<span className="text-xs text-muted-foreground hidden sm:inline">
					Updated {relativeTime}
				</span>
			)}

			{/* Interval Selector */}
			<Select
				value={String(interval)}
				onValueChange={(value) => onIntervalChange(Number(value) as RefreshInterval)}
			>
				<SelectTrigger className="w-[130px] h-9">
					<SelectValue placeholder="Refresh interval" />
				</SelectTrigger>
				<SelectContent>
					{REFRESH_INTERVAL_OPTIONS.map((option) => (
						<SelectItem key={option.value} value={String(option.value)}>
							{option.label}
						</SelectItem>
					))}
				</SelectContent>
			</Select>

			{/* Manual Refresh Button */}
			<Button
				variant="outline"
				size="sm"
				onClick={onRefresh}
				disabled={isRefreshing}
				className="h-9"
			>
				<RefreshCw
					className={cn("h-4 w-4", isRefreshing && "animate-spin")}
				/>
				<span className="sr-only">Refresh</span>
			</Button>
		</div>
	);
}

/**
 * Compact variant for inline use
 */
export function RefreshControlCompact({
	onRefresh,
	isRefreshing,
	lastUpdated,
	className,
}: Pick<RefreshControlProps, "onRefresh" | "isRefreshing" | "lastUpdated" | "className">) {
	const [relativeTime, setRelativeTime] = React.useState<string>("");

	React.useEffect(() => {
		if (!lastUpdated) return;

		const update = () => {
			setRelativeTime(getRelativeTime(lastUpdated));
		};

		update();
		const timer = setInterval(update, 1000);

		return () => clearInterval(timer);
	}, [lastUpdated]);

	return (
		<div className={cn("flex items-center gap-2 text-xs text-muted-foreground", className)}>
			{lastUpdated && relativeTime && (
				<span>Updated {relativeTime}</span>
			)}
			<button
				type="button"
				onClick={onRefresh}
				disabled={isRefreshing}
				className="inline-flex items-center hover:text-foreground transition-colors disabled:opacity-50"
			>
				<RefreshCw
					className={cn("h-3 w-3", isRefreshing && "animate-spin")}
				/>
			</button>
		</div>
	);
}
