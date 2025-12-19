"use client";

import * as React from "react";
import { AlertCircle, CheckCircle2, RefreshCw } from "lucide-react";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import type { HealthResponse } from "@/types/health";

export interface HealthStatusCardProps {
	health?: HealthResponse;
	isLoading?: boolean;
	error?: Error;
	lastChecked?: Date;
	onRetry?: () => void;
}

export function HealthStatusCard({
	health,
	isLoading,
	error,
	lastChecked,
	onRetry,
}: HealthStatusCardProps) {
	const isOnline = health?.status === "ok";
	const showError = !isLoading && error;

	return (
		<Card>
			<CardHeader>
				<div className="flex items-center justify-between">
					<CardTitle>API Status</CardTitle>
					{showError && onRetry && (
						<Button
							variant="ghost"
							size="icon-sm"
							onClick={onRetry}
							aria-label="Retry"
						>
							<RefreshCw
								data-icon="inline-end"
								className="size-4"
							/>
						</Button>
					)}
				</div>
				{lastChecked && !isLoading && (
					<CardDescription>
						Last checked: {formatTimestamp(lastChecked)}
					</CardDescription>
				)}
			</CardHeader>

			<CardContent>
				{isLoading ? (
					<LoadingSkeleton />
				) : showError ? (
					<ErrorState error={error} />
				) : (
					<StatusDisplay health={health} isOnline={isOnline} />
				)}
			</CardContent>
		</Card>
	);
}

function StatusDisplay({
	health,
	isOnline,
}: {
	health?: HealthResponse;
	isOnline: boolean;
}) {
	return (
		<div className="flex items-center gap-4">
			<div className="relative">
				<div
					className={cn(
						"size-12 rounded-full flex items-center justify-center transition-colors",
						isOnline
							? "bg-green-500/10 text-green-600 dark:bg-green-500/20 dark:text-green-400"
							: "bg-red-500/10 text-red-600 dark:bg-red-500/20 dark:text-red-400",
					)}
				>
					{isOnline ? (
						<CheckCircle2 className="size-6" />
					) : (
						<AlertCircle className="size-6" />
					)}
				</div>
				{isOnline && (
					<span className="absolute top-0 right-0 flex size-3">
						<span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75" />
						<span className="relative inline-flex rounded-full size-3 bg-green-500" />
					</span>
				)}
			</div>

			<div className="flex-1 space-y-1">
				<div className="flex items-center gap-2">
					<p className="text-lg font-semibold">
						{isOnline ? "API Online" : "API Offline"}
					</p>
					<Badge variant={isOnline ? "default" : "destructive"}>
						{health?.status || "error"}
					</Badge>
				</div>
				{health?.service && (
					<p className="text-sm text-muted-foreground">
						{health.service}
					</p>
				)}
				{health?.timestamp && (
					<p className="text-xs text-muted-foreground">
						Timestamp:{" "}
						{new Date(health.timestamp).toLocaleString("en-US")}
					</p>
				)}
			</div>
		</div>
	);
}

function LoadingSkeleton() {
	return (
		<div className="flex items-center gap-4">
			<Skeleton className="size-12 rounded-full" />
			<div className="flex-1 space-y-2">
				<Skeleton className="h-6 w-32" />
				<Skeleton className="h-4 w-48" />
			</div>
		</div>
	);
}

function ErrorState({ error }: { error: Error }) {
	return (
		<div className="flex items-center gap-4">
			<div className="size-12 rounded-full bg-red-500/10 dark:bg-red-500/20 flex items-center justify-center">
				<AlertCircle className="size-6 text-red-600 dark:text-red-400" />
			</div>
			<div className="flex-1 space-y-1">
				<p className="text-lg font-semibold text-red-600 dark:text-red-400">
					Error checking status
				</p>
				<p className="text-sm text-muted-foreground">{error.message}</p>
			</div>
		</div>
	);
}

function formatTimestamp(date: Date): string {
	const now = new Date();
	const diffMs = now.getTime() - date.getTime();
	const diffSecs = Math.floor(diffMs / 1000);
	const diffMins = Math.floor(diffSecs / 60);

	if (diffSecs < 60) {
		return `${diffSecs}s ago`;
	}
	if (diffMins < 60) {
		return `${diffMins}m ago`;
	}
	return date.toLocaleTimeString("en-US", {
		hour: "2-digit",
		minute: "2-digit",
	});
}
