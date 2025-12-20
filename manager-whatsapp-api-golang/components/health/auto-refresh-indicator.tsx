"use client";

import * as React from "react";
import { RefreshCw } from "lucide-react";
import { cn } from "@/lib/utils";

export interface AutoRefreshIndicatorProps {
	interval: number;
	lastRefresh: Date;
	onManualRefresh: () => void;
	isRefreshing?: boolean;
	className?: string;
}

export function AutoRefreshIndicator({
	interval,
	lastRefresh,
	onManualRefresh,
	isRefreshing = false,
	className,
}: AutoRefreshIndicatorProps) {
	const [timeRemaining, setTimeRemaining] = React.useState(interval);

	React.useEffect(() => {
		const calculateRemaining = () => {
			const now = Date.now();
			const elapsed = now - lastRefresh.getTime();
			const remaining = Math.max(0, interval - elapsed);
			setTimeRemaining(remaining);
		};

		calculateRemaining();
		const intervalId = setInterval(calculateRemaining, 1000);

		return () => clearInterval(intervalId);
	}, [interval, lastRefresh]);

	const secondsRemaining = Math.ceil(timeRemaining / 1000);
	const progressPercentage = ((interval - timeRemaining) / interval) * 100;

	return (
		<button
			onClick={onManualRefresh}
			disabled={isRefreshing}
			className={cn(
				"group flex items-center gap-2 text-xs text-muted-foreground hover:text-foreground transition-colors disabled:opacity-50",
				className,
			)}
		>
			<RefreshCw
				className={cn(
					"size-3.5 transition-transform",
					isRefreshing && "animate-spin",
					!isRefreshing && "group-hover:rotate-45",
				)}
			/>
			<span className="tabular-nums">{secondsRemaining}s</span>
			<div className="w-12 h-1 bg-muted rounded-full overflow-hidden">
				<div
					className="h-full bg-primary/60 transition-all duration-1000 ease-linear rounded-full"
					style={{ width: `${progressPercentage}%` }}
				/>
			</div>
		</button>
	);
}

/**
 * Compact version - just icon with countdown
 */
export function AutoRefreshIndicatorCompact({
	interval,
	lastRefresh,
	onManualRefresh,
	isRefreshing = false,
	className,
}: AutoRefreshIndicatorProps) {
	const [timeRemaining, setTimeRemaining] = React.useState(interval);

	React.useEffect(() => {
		const calculateRemaining = () => {
			const now = Date.now();
			const elapsed = now - lastRefresh.getTime();
			const remaining = Math.max(0, interval - elapsed);
			setTimeRemaining(remaining);
		};

		calculateRemaining();
		const intervalId = setInterval(calculateRemaining, 1000);

		return () => clearInterval(intervalId);
	}, [interval, lastRefresh]);

	const secondsRemaining = Math.ceil(timeRemaining / 1000);

	return (
		<button
			onClick={onManualRefresh}
			disabled={isRefreshing}
			title="Click to refresh"
			className={cn(
				"group flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors disabled:opacity-50",
				className,
			)}
		>
			<RefreshCw
				className={cn(
					"size-3.5 transition-transform",
					isRefreshing && "animate-spin",
					!isRefreshing && "group-hover:rotate-45",
				)}
			/>
			<span className="tabular-nums">{secondsRemaining}s</span>
		</button>
	);
}
