/**
 * Health Monitoring Page
 *
 * Displays API health status, readiness checks, and dependency status
 * with automatic polling for real-time monitoring.
 */

"use client";

import * as React from "react";
import { useHealth } from "@/hooks/use-health";
import { PageHeader } from "@/components/shared/page-header";
import {
	HealthStatusCard,
	ReadinessCard,
	AutoRefreshIndicator,
} from "@/components/health";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AlertCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";

const POLLING_INTERVAL = 30000; // 30 seconds

export default function HealthPage() {
	const [lastRefresh, setLastRefresh] = React.useState(new Date());

	// Fetch health data with polling
	const {
		health,
		readiness,
		isDegraded,
		isLoading,
		error,
		isValidating,
		refresh,
	} = useHealth({
		interval: POLLING_INTERVAL,
		enabled: true,
		includeReadiness: true,
	});

	const handleRefresh = async () => {
		await refresh();
		setLastRefresh(new Date());
	};

	// Update lastRefresh when data changes
	React.useEffect(() => {
		if (!isValidating && (health || readiness)) {
			setLastRefresh(new Date());
		}
	}, [isValidating, health, readiness]);

	// Show error state
	if (error && !health && !readiness) {
		return (
			<div className="space-y-6">
				<PageHeader
					title="API Status"
					description="Monitor system health and availability"
				/>
				<Alert variant="destructive">
					<AlertCircle className="h-4 w-4" />
					<AlertTitle>Error loading health data</AlertTitle>
					<AlertDescription>
						{error.message ||
							"Failed to connect to the API. Check your connection."}
					</AlertDescription>
				</Alert>
				<Button onClick={handleRefresh} variant="outline">
					<RefreshCw className="h-4 w-4" />
					Retry
				</Button>
			</div>
		);
	}

	return (
		<div className="space-y-6">
			{/* Page Header with Auto-Refresh Indicator */}
			<div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
				<PageHeader
					title="API Status"
					description="Monitor system health and availability"
				/>
				<AutoRefreshIndicator
					interval={POLLING_INTERVAL}
					lastRefresh={lastRefresh}
					onManualRefresh={handleRefresh}
					isRefreshing={isValidating}
				/>
			</div>

			{/* Overall Status Alert */}
			{!isLoading && readiness && (
				<>
					{!readiness.ready && (
						<Alert variant="destructive">
							<AlertCircle className="h-4 w-4" />
							<AlertTitle>System Unavailable</AlertTitle>
							<AlertDescription>
								One or more critical components have issues.
								Check the details below.
							</AlertDescription>
						</Alert>
					)}
					{isDegraded && (
						<Alert variant="default">
							<AlertCircle className="h-4 w-4" />
							<AlertTitle>System Degraded</AlertTitle>
							<AlertDescription>
								The system is operational, but some components
								have reduced performance.
							</AlertDescription>
						</Alert>
					)}
				</>
			)}

			{/* Health Cards Grid */}
			<div className="grid gap-6 md:grid-cols-2">
				{/* Basic Health Status */}
				<HealthStatusCard
					health={health}
					isLoading={isLoading}
					error={error}
				/>

				{/* Detailed Readiness Status */}
				<ReadinessCard
					readiness={readiness}
					isLoading={isLoading}
					error={error}
				/>
			</div>
		</div>
	);
}
