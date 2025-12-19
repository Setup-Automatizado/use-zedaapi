"use client";

import * as React from "react";
import { Clock, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
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
	const [formattedTime, setFormattedTime] = React.useState<string>("");
	const [mounted, setMounted] = React.useState(false);

	React.useEffect(() => {
		setMounted(true);
	}, []);

	React.useEffect(() => {
		const calculateRemaining = () => {
			const now = Date.now();
			const elapsed = now - lastRefresh.getTime();
			const remaining = Math.max(0, interval - elapsed);
			setTimeRemaining(remaining);
			setFormattedTime(
				lastRefresh.toLocaleTimeString("pt-BR", {
					hour: "2-digit",
					minute: "2-digit",
					second: "2-digit",
				}),
			);
		};

		calculateRemaining();
		const intervalId = setInterval(calculateRemaining, 1000);

		return () => clearInterval(intervalId);
	}, [interval, lastRefresh]);

	const secondsRemaining = Math.ceil(timeRemaining / 1000);
	const progressPercentage = ((interval - timeRemaining) / interval) * 100;

	return (
		<div
			className={cn(
				"flex items-center gap-3 rounded-lg border bg-card p-3 text-sm",
				className,
			)}
		>
			{/* Auto-refresh info */}
			<div className="flex items-center gap-2 flex-1 min-w-0">
				<Clock className="size-4 text-muted-foreground shrink-0" />
				<div className="flex-1 min-w-0">
					<p className="text-xs text-muted-foreground truncate">
						Atualizacao automatica a cada {interval / 1000}s
					</p>
					<div className="flex items-center gap-2 mt-1">
						{/* Progress bar */}
						<div className="flex-1 h-1.5 bg-muted rounded-full overflow-hidden">
							<div
								className="h-full bg-primary transition-all duration-1000 ease-linear rounded-full"
								style={{ width: `${progressPercentage}%` }}
							/>
						</div>
						{/* Countdown */}
						<span className="text-xs font-mono text-muted-foreground shrink-0 min-w-[3ch]">
							{secondsRemaining}s
						</span>
					</div>
				</div>
			</div>

			{/* Manual refresh button */}
			<Button
				variant="outline"
				size="sm"
				onClick={onManualRefresh}
				disabled={isRefreshing}
				className="shrink-0"
			>
				<RefreshCw
					data-icon="inline-start"
					className={cn("size-4", isRefreshing && "animate-spin")}
				/>
				Atualizar
			</Button>

			{/* Last refresh time */}
			<div className="hidden sm:block text-xs text-muted-foreground shrink-0">
				Ultima atualizacao:{" "}
				<span className="font-mono">
					{mounted ? formattedTime : "--:--:--"}
				</span>
			</div>
		</div>
	);
}

/**
 * Compact version for mobile/constrained spaces
 */
export function AutoRefreshIndicatorCompact({
	interval,
	lastRefresh,
	onManualRefresh,
	isRefreshing = false,
	className,
}: AutoRefreshIndicatorProps) {
	const [timeRemaining, setTimeRemaining] = React.useState(interval);
	const [formattedTime, setFormattedTime] = React.useState<string>("");
	const [mounted, setMounted] = React.useState(false);

	React.useEffect(() => {
		setMounted(true);
	}, []);

	React.useEffect(() => {
		const calculateRemaining = () => {
			const now = Date.now();
			const elapsed = now - lastRefresh.getTime();
			const remaining = Math.max(0, interval - elapsed);
			setTimeRemaining(remaining);
			setFormattedTime(
				lastRefresh.toLocaleTimeString("pt-BR", {
					hour: "2-digit",
					minute: "2-digit",
				}),
			);
		};

		calculateRemaining();
		const intervalId = setInterval(calculateRemaining, 1000);

		return () => clearInterval(intervalId);
	}, [interval, lastRefresh]);

	const secondsRemaining = Math.ceil(timeRemaining / 1000);
	const progressPercentage = ((interval - timeRemaining) / interval) * 100;

	return (
		<div className={cn("flex items-center gap-2", className)}>
			{/* Progress indicator */}
			<div className="flex-1 min-w-0">
				<div className="flex items-center justify-between text-xs text-muted-foreground mb-1">
					<span className="truncate">
						Auto-refresh em {mounted ? secondsRemaining : "--"}s
					</span>
					<span className="font-mono shrink-0">
						{mounted ? formattedTime : "--:--"}
					</span>
				</div>
				<div className="h-1 bg-muted rounded-full overflow-hidden">
					<div
						className="h-full bg-primary transition-all duration-1000 ease-linear rounded-full"
						style={{ width: `${progressPercentage}%` }}
					/>
				</div>
			</div>

			{/* Manual refresh button */}
			<Button
				variant="ghost"
				size="icon-sm"
				onClick={onManualRefresh}
				disabled={isRefreshing}
				aria-label="Atualizar agora"
			>
				<RefreshCw
					className={cn("size-4", isRefreshing && "animate-spin")}
				/>
			</Button>
		</div>
	);
}
