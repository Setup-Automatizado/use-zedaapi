"use client";

import {
	Activity,
	AlertTriangle,
	Database,
	Globe,
	Image as ImageIcon,
	MessageSquare,
	Send,
	Settings,
	Zap,
} from "lucide-react";
import { useState } from "react";
import {
	EventMetricsTab,
	HTTPMetricsTab,
	InstanceFilter,
	MediaMetricsTab,
	MessageQueueTab,
	OverviewTab,
	RefreshControl,
	StatusCacheTab,
	SystemTab,
	TransportMetricsTab,
} from "@/components/metrics";
import { PageHeader } from "@/components/shared/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useMetrics, useMetricsInstances } from "@/hooks/use-metrics";
import type { RefreshInterval } from "@/types/metrics";

export default function MetricsPage() {
	const [refreshInterval, setRefreshInterval] =
		useState<RefreshInterval>(15000);
	const [selectedInstance, setSelectedInstance] = useState<string | null>(
		null,
	);

	const { metrics, error, isLoading, isValidating, refresh } = useMetrics({
		interval: refreshInterval,
		instanceId: selectedInstance ?? undefined,
	});

	const { instances } = useMetricsInstances();

	const handleRefresh = () => {
		refresh();
	};

	if (error) {
		return (
			<div className="space-y-6">
				<PageHeader
					title="Metrics"
					description="Real-time system metrics and performance monitoring"
				/>
				<Alert variant="destructive">
					<AlertTriangle className="h-4 w-4" />
					<AlertTitle>Failed to load metrics</AlertTitle>
					<AlertDescription>
						{error.message ||
							"Unable to fetch metrics from the API. Please check if the backend is running."}
					</AlertDescription>
				</Alert>
			</div>
		);
	}

	return (
		<div className="space-y-6">
			{/* Header */}
			<div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
				<PageHeader
					title="Metrics"
					description="Real-time system metrics and performance monitoring"
				/>
				<div className="flex flex-col gap-3 sm:flex-row sm:items-center shrink-0">
					<InstanceFilter
						instances={instances}
						selectedInstance={selectedInstance}
						onSelect={setSelectedInstance}
					/>
					<RefreshControl
						interval={refreshInterval}
						onIntervalChange={setRefreshInterval}
						onRefresh={handleRefresh}
						isRefreshing={isValidating}
						lastUpdated={metrics ? new Date() : undefined}
					/>
				</div>
			</div>

			{/* Loading State */}
			{isLoading && !metrics && <MetricsLoadingSkeleton />}

			{/* Content */}
			{metrics && (
				<Tabs defaultValue="overview" className="w-full">
					<TabsList className="h-auto flex flex-wrap gap-1 bg-muted/50 p-1 rounded-lg">
						<TabsTrigger
							value="overview"
							className="flex items-center gap-2 px-3 py-2 text-sm data-[state=active]:bg-background rounded-md"
						>
							<Activity className="h-4 w-4 shrink-0" />
							<span className="hidden sm:inline">Overview</span>
						</TabsTrigger>
						<TabsTrigger
							value="http"
							className="flex items-center gap-2 px-3 py-2 text-sm data-[state=active]:bg-background rounded-md"
						>
							<Globe className="h-4 w-4 shrink-0" />
							<span className="hidden sm:inline">HTTP</span>
						</TabsTrigger>
						<TabsTrigger
							value="events"
							className="flex items-center gap-2 px-3 py-2 text-sm data-[state=active]:bg-background rounded-md"
						>
							<Zap className="h-4 w-4 shrink-0" />
							<span className="hidden sm:inline">Events</span>
						</TabsTrigger>
						<TabsTrigger
							value="queue"
							className="flex items-center gap-2 px-3 py-2 text-sm data-[state=active]:bg-background rounded-md"
						>
							<MessageSquare className="h-4 w-4 shrink-0" />
							<span className="hidden sm:inline">Queue</span>
						</TabsTrigger>
						<TabsTrigger
							value="media"
							className="flex items-center gap-2 px-3 py-2 text-sm data-[state=active]:bg-background rounded-md"
						>
							<ImageIcon className="h-4 w-4 shrink-0" />
							<span className="hidden sm:inline">Media</span>
						</TabsTrigger>
						<TabsTrigger
							value="transport"
							className="flex items-center gap-2 px-3 py-2 text-sm data-[state=active]:bg-background rounded-md"
						>
							<Send className="h-4 w-4 shrink-0" />
							<span className="hidden sm:inline">Transport</span>
						</TabsTrigger>
						<TabsTrigger
							value="status-cache"
							className="flex items-center gap-2 px-3 py-2 text-sm data-[state=active]:bg-background rounded-md"
						>
							<Database className="h-4 w-4 shrink-0" />
							<span className="hidden sm:inline">
								Status Cache
							</span>
						</TabsTrigger>
						<TabsTrigger
							value="system"
							className="flex items-center gap-2 px-3 py-2 text-sm data-[state=active]:bg-background rounded-md"
						>
							<Settings className="h-4 w-4 shrink-0" />
							<span className="hidden sm:inline">System</span>
						</TabsTrigger>
					</TabsList>

					<TabsContent value="overview" className="mt-6">
						<OverviewTab metrics={metrics} />
					</TabsContent>

					<TabsContent value="http" className="mt-6">
						<HTTPMetricsTab metrics={metrics.http} />
					</TabsContent>

					<TabsContent value="events" className="mt-6">
						<EventMetricsTab metrics={metrics.events} />
					</TabsContent>

					<TabsContent value="queue" className="mt-6">
						<MessageQueueTab metrics={metrics.messageQueue} />
					</TabsContent>

					<TabsContent value="media" className="mt-6">
						<MediaMetricsTab metrics={metrics.media} />
					</TabsContent>

					<TabsContent value="transport" className="mt-6">
						<TransportMetricsTab metrics={metrics.transport} />
					</TabsContent>

					<TabsContent value="status-cache" className="mt-6">
						<StatusCacheTab metrics={metrics.statusCache} />
					</TabsContent>

					<TabsContent value="system" className="mt-6">
						<SystemTab
							metrics={metrics.system}
							workers={metrics.workers}
						/>
					</TabsContent>
				</Tabs>
			)}
		</div>
	);
}

function MetricsLoadingSkeleton() {
	return (
		<div className="flex flex-col gap-6">
			{/* Tabs skeleton */}
			<div className="flex gap-2">
				{Array.from({ length: 8 }).map((_, i) => (
					<Skeleton key={i} className="h-10 w-24" />
				))}
			</div>

			{/* KPI Cards skeleton */}
			<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
				{Array.from({ length: 8 }).map((_, i) => (
					<Skeleton key={i} className="h-32 w-full" />
				))}
			</div>

			{/* Charts skeleton */}
			<div className="grid gap-4 md:grid-cols-2">
				<Skeleton className="h-64 w-full" />
				<Skeleton className="h-64 w-full" />
			</div>

			{/* Table skeleton */}
			<Skeleton className="h-48 w-full" />
		</div>
	);
}
