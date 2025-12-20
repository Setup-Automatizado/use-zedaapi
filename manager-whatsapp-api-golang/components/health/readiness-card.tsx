"use client";

import { Activity, AlertCircle, CheckCircle2 } from "lucide-react";
import * as React from "react";
import { Badge } from "@/components/ui/badge";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import type { ReadinessResponse } from "@/types/health";
import { getCriticalStatus } from "@/types/health";
import { DependencyStatus } from "./dependency-status";

export interface ReadinessCardProps {
	readiness?: ReadinessResponse;
	isLoading?: boolean;
	error?: Error;
}

export function ReadinessCard({
	readiness,
	isLoading,
	error,
}: ReadinessCardProps) {
	const isReady = readiness?.ready ?? false;
	const showError = !isLoading && error;

	const criticalStatus = readiness?.checks
		? getCriticalStatus(readiness.checks)
		: "unhealthy";

	return (
		<Card>
			<CardHeader>
				<div className="flex items-center justify-between">
					<CardTitle>Service Readiness</CardTitle>
					{readiness && (
						<Badge variant={isReady ? "default" : "destructive"}>
							{isReady ? (
								<>
									<CheckCircle2 data-icon="inline-start" className="size-3" />
									Ready
								</>
							) : (
								<>
									<AlertCircle data-icon="inline-start" className="size-3" />
									Not Ready
								</>
							)}
						</Badge>
					)}
				</div>
				{readiness?.observed_at && (
					<CardDescription>
						Observed at:{" "}
						{new Date(readiness.observed_at).toLocaleString("en-US", {
							dateStyle: "short",
							timeStyle: "medium",
						})}
					</CardDescription>
				)}
			</CardHeader>

			<CardContent>
				{isLoading ? (
					<LoadingSkeleton />
				) : showError ? (
					<ErrorState error={error} />
				) : readiness ? (
					<DependenciesList
						readiness={readiness}
						criticalStatus={criticalStatus}
					/>
				) : (
					<EmptyState />
				)}
			</CardContent>
		</Card>
	);
}

function DependenciesList({
	readiness,
	criticalStatus,
}: {
	readiness: ReadinessResponse;
	criticalStatus: "healthy" | "degraded" | "unhealthy";
}) {
	const checkEntries = Object.entries(readiness.checks).filter(
		([, status]) => status !== undefined,
	);

	if (checkEntries.length === 0) {
		return <EmptyState />;
	}

	return (
		<div className="space-y-4">
			{/* Overall Status Summary */}
			<div
				className={cn(
					"rounded-lg p-4 border",
					criticalStatus === "healthy" &&
						"bg-green-50/50 dark:bg-green-950/20 border-green-200 dark:border-green-800",
					criticalStatus === "degraded" &&
						"bg-yellow-50/50 dark:bg-yellow-950/20 border-yellow-200 dark:border-yellow-800",
					criticalStatus === "unhealthy" &&
						"bg-red-50/50 dark:bg-red-950/20 border-red-200 dark:border-red-800",
				)}
			>
				<div className="flex items-center gap-2">
					<Activity
						className={cn(
							"size-4",
							criticalStatus === "healthy" &&
								"text-green-600 dark:text-green-400",
							criticalStatus === "degraded" &&
								"text-yellow-600 dark:text-yellow-400",
							criticalStatus === "unhealthy" &&
								"text-red-600 dark:text-red-400",
						)}
					/>
					<p className="text-sm font-medium">
						{criticalStatus === "healthy" && "All dependencies healthy"}
						{criticalStatus === "degraded" &&
							"Some dependencies with degraded performance"}
						{criticalStatus === "unhealthy" && "Some dependencies unavailable"}
					</p>
				</div>
			</div>

			{/* Dependencies List */}
			<div className="border rounded-lg overflow-hidden">
				{checkEntries.map(([name, status]) => (
					<DependencyStatus key={name} name={name} status={status} />
				))}
			</div>

			{/* Statistics */}
			<div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
				<StatCard
					label="Components"
					value={checkEntries.length}
					variant="neutral"
				/>
				<StatCard
					label="Healthy"
					value={checkEntries.filter(([, s]) => s.status === "healthy").length}
					variant="success"
				/>
				<StatCard
					label="Issues"
					value={
						checkEntries.filter(
							([, s]) => s.status === "degraded" || s.status === "unhealthy",
						).length
					}
					variant="error"
					className="col-span-2 sm:col-span-1"
				/>
			</div>
		</div>
	);
}

function StatCard({
	label,
	value,
	variant,
	className,
}: {
	label: string;
	value: number;
	variant: "neutral" | "success" | "error";
	className?: string;
}) {
	return (
		<div
			className={cn(
				"rounded-lg border p-3",
				variant === "neutral" && "bg-muted/30 border-border",
				variant === "success" &&
					"bg-green-50/50 dark:bg-green-950/20 border-green-200 dark:border-green-800",
				variant === "error" &&
					"bg-red-50/50 dark:bg-red-950/20 border-red-200 dark:border-red-800",
				className,
			)}
		>
			<p className="text-xs text-muted-foreground mb-1">{label}</p>
			<p
				className={cn(
					"text-2xl font-bold",
					variant === "success" && "text-green-600 dark:text-green-400",
					variant === "error" && "text-red-600 dark:text-red-400",
				)}
			>
				{value}
			</p>
		</div>
	);
}

function LoadingSkeleton() {
	return (
		<div className="space-y-4">
			<Skeleton className="h-16 w-full" />
			<div className="space-y-2">
				{[1, 2, 3].map((i) => (
					<div key={i} className="flex items-center gap-3 py-3">
						<Skeleton className="size-8 rounded-lg shrink-0" />
						<div className="flex-1 space-y-2">
							<Skeleton className="h-4 w-24" />
							<Skeleton className="h-3 w-32" />
						</div>
					</div>
				))}
			</div>
		</div>
	);
}

function ErrorState({ error }: { error: Error }) {
	return (
		<div className="flex flex-col items-center justify-center py-8 text-center">
			<div className="size-12 rounded-full bg-red-500/10 dark:bg-red-500/20 flex items-center justify-center mb-3">
				<AlertCircle className="size-6 text-red-600 dark:text-red-400" />
			</div>
			<p className="text-sm font-medium text-red-600 dark:text-red-400 mb-1">
				Error checking readiness
			</p>
			<p className="text-xs text-muted-foreground">{error.message}</p>
		</div>
	);
}

function EmptyState() {
	return (
		<div className="flex flex-col items-center justify-center py-8 text-center">
			<Activity className="size-8 text-muted-foreground/50 mb-3" />
			<p className="text-sm text-muted-foreground">
				No readiness information available
			</p>
		</div>
	);
}
