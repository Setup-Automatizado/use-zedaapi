/**
 * Dashboard Page
 *
 * Main dashboard displaying instance statistics, recent instances,
 * quick actions, and health status.
 */

"use client";

import { AlertCircle } from "lucide-react";
import * as React from "react";
import {
	HealthSummary,
	QuickActions,
	RecentInstances,
	StatsCards,
} from "@/components/dashboard";
import { PageHeader } from "@/components/shared/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import { useHealth } from "@/hooks/use-health";
import { useInstancesWithDevice } from "@/hooks/use-instances-with-device";

export default function DashboardPage() {
	// Fetch instances with device info for avatars
	const {
		instances,
		deviceMap,
		isLoading: instancesLoading,
		error: instancesError,
	} = useInstancesWithDevice({
		page: 1,
		pageSize: 10, // Recent instances with avatars
	});

	// Fetch health status
	const {
		isHealthy,
		isDegraded,
		isLoading: healthLoading,
	} = useHealth({
		interval: 30000, // Poll every 30 seconds
		enabled: true,
	});

	// Calculate statistics
	const stats = React.useMemo(() => {
		if (!instances) {
			return { total: 0, connected: 0, disconnected: 0, pending: 0 };
		}

		const total = instances.length;
		const connected = instances.filter(
			(instance) => instance.whatsappConnected && instance.phoneConnected,
		).length;
		const disconnected = instances.filter(
			(instance) => !instance.whatsappConnected || !instance.phoneConnected,
		).length;
		const pending = instances.filter(
			(instance) => !instance.phoneConnected,
		).length;

		return { total, connected, disconnected, pending };
	}, [instances]);

	// Show error state
	if (instancesError) {
		return (
			<div className="space-y-6">
				<PageHeader
					title="Dashboard"
					description="Welcome to the WhatsApp API Manager"
				/>
				<Alert variant="destructive">
					<AlertCircle className="h-4 w-4" />
					<AlertTitle>Error loading data</AlertTitle>
					<AlertDescription>
						{instancesError.message ||
							"Failed to load instances. Please try again."}
					</AlertDescription>
				</Alert>
			</div>
		);
	}

	return (
		<div className="space-y-6">
			{/* Header with Quick Actions */}
			<div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
				<PageHeader
					title="Dashboard"
					description="Welcome to the WhatsApp API Manager"
				/>
				<div className="flex items-center gap-2">
					<QuickActions />
					{!healthLoading && (
						<HealthSummary
							isHealthy={isHealthy}
							isDegraded={isDegraded}
							lastChecked={new Date()}
						/>
					)}
				</div>
			</div>

			{/* Stats Cards */}
			{instancesLoading ? (
				<div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
					{Array.from({ length: 4 }).map((_, i) => (
						<Skeleton key={i} className="h-32" />
					))}
				</div>
			) : (
				<StatsCards
					total={stats.total}
					connected={stats.connected}
					disconnected={stats.disconnected}
					pending={stats.pending}
				/>
			)}

			{/* Recent Instances - Full Width */}
			{instancesLoading ? (
				<Skeleton className="h-96" />
			) : (
				<RecentInstances instances={instances || []} deviceMap={deviceMap} />
			)}
		</div>
	);
}
