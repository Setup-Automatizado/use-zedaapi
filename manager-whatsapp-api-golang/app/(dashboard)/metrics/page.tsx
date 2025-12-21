"use client";

import {
  Activity,
  AlertTriangle,
  Globe,
  Image as ImageIcon,
  MessageSquare,
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
  SystemTab,
} from "@/components/metrics";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useMetrics, useMetricsInstances } from "@/hooks/use-metrics";
import type { RefreshInterval } from "@/types/metrics";

export default function MetricsPage() {
  const [refreshInterval, setRefreshInterval] = useState<RefreshInterval>(15000);
  const [selectedInstance, setSelectedInstance] = useState<string | null>(null);

  const {
    metrics,
    error,
    isLoading,
    isValidating,
    refresh,
  } = useMetrics({
    interval: refreshInterval,
    instanceId: selectedInstance ?? undefined,
  });

  const { instances } = useMetricsInstances();

  const handleRefresh = () => {
    refresh();
  };

  if (error) {
    return (
      <div className="flex flex-col gap-6 p-6">
        <div className="flex flex-col gap-2">
          <h1 className="text-2xl font-semibold tracking-tight">Metrics</h1>
          <p className="text-sm text-muted-foreground">
            Real-time system metrics and performance monitoring
          </p>
        </div>
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Failed to load metrics</AlertTitle>
          <AlertDescription>
            {error.message || "Unable to fetch metrics from the API. Please check if the backend is running."}
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6 p-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-semibold tracking-tight">Metrics</h1>
          <p className="text-sm text-muted-foreground">
            Real-time system metrics and performance monitoring
          </p>
        </div>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
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
          <TabsList className="grid w-full grid-cols-3 lg:grid-cols-6">
            <TabsTrigger value="overview" className="flex items-center gap-2">
              <Activity className="h-4 w-4" />
              <span className="hidden sm:inline">Overview</span>
            </TabsTrigger>
            <TabsTrigger value="http" className="flex items-center gap-2">
              <Globe className="h-4 w-4" />
              <span className="hidden sm:inline">HTTP</span>
            </TabsTrigger>
            <TabsTrigger value="events" className="flex items-center gap-2">
              <Zap className="h-4 w-4" />
              <span className="hidden sm:inline">Events</span>
            </TabsTrigger>
            <TabsTrigger value="queue" className="flex items-center gap-2">
              <MessageSquare className="h-4 w-4" />
              <span className="hidden sm:inline">Queue</span>
            </TabsTrigger>
            <TabsTrigger value="media" className="flex items-center gap-2">
              <ImageIcon className="h-4 w-4" />
              <span className="hidden sm:inline">Media</span>
            </TabsTrigger>
            <TabsTrigger value="system" className="flex items-center gap-2">
              <Settings className="h-4 w-4" />
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

          <TabsContent value="system" className="mt-6">
            <SystemTab metrics={metrics.system} workers={metrics.workers} />
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
        {Array.from({ length: 6 }).map((_, i) => (
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
