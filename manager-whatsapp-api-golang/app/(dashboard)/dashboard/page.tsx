/**
 * Dashboard Page
 *
 * Main dashboard displaying instance statistics, recent instances,
 * quick actions, and health status.
 */

"use client";

import { AlertCircle } from "lucide-react";
import { useMemo } from "react";
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
	const {
		instances,
		deviceMap,
		isLoading: instancesLoading,
		error: instancesError,
	} = useInstancesWithDevice({
		page: 1,
		pageSize: 10,
	});

	const {
		isHealthy,
		isDegraded,
		isLoading: healthLoading,
	} = useHealth({
		interval: 30000,
		enabled: true,
	});

	const stats = useMemo(() => {
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
		<div className="space-y-8">
			{/* Header */}
			<div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
				<PageHeader
					title="Dashboard"
					description="Welcome to the WhatsApp API Manager"
				/>
				<div className="flex items-center gap-2 shrink-0">
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
				<div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
					{Array.from({ length: 4 }).map((_, i) => (
						<Skeleton key={i} className="h-28 rounded-xl" />
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

			{/* Recent Instances */}
			{instancesLoading ? (
				<Skeleton className="h-[400px] rounded-xl" />
			) : (
				<RecentInstances instances={instances || []} deviceMap={deviceMap} />
			)}
		</div>
	);
}
