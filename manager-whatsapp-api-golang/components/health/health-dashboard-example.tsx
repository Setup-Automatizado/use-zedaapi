"use client";

import * as React from "react";
import {
	AutoRefreshIndicator,
	HealthStatusCard,
	MetricsDisplay,
	ReadinessCard,
} from "@/components/health";
import type { HealthResponse, ReadinessResponse } from "@/types/health";

/**
 * Example Health Dashboard Component
 *
 * This is a complete example showing how to use all health monitoring
 * components together with SWR for data fetching and auto-refresh.
 *
 * Usage:
 * ```tsx
 * import { HealthDashboard } from "@/components/health/health-dashboard-example"
 *
 * export default function HealthPage() {
 *   return <HealthDashboard />
 * }
 * ```
 */

// Auto-refresh interval (30 seconds)
const REFRESH_INTERVAL = 30000;

export function HealthDashboard() {
	const [lastRefresh, setLastRefresh] = React.useState(new Date());
	const [isManualRefreshing, setIsManualRefreshing] = React.useState(false);

	// Fetch health status
	const {
		data: health,
		error: healthError,
		isLoading: healthLoading,
		mutate: refreshHealth,
	} = useFetchHealth(REFRESH_INTERVAL);

	// Fetch readiness status
	const {
		data: readiness,
		error: readinessError,
		isLoading: readinessLoading,
		mutate: refreshReadiness,
	} = useFetchReadiness(REFRESH_INTERVAL);

	// Manual refresh handler
	const handleManualRefresh = React.useCallback(async () => {
		setIsManualRefreshing(true);
		await Promise.all([refreshHealth(), refreshReadiness()]);
		setLastRefresh(new Date());
		setIsManualRefreshing(false);
	}, [refreshHealth, refreshReadiness]);

	// Update last refresh timestamp when data changes
	React.useEffect(() => {
		if (health || readiness) {
			setLastRefresh(new Date());
		}
	}, [health, readiness]);

	return (
		<div className="space-y-6">
			{/* Auto-refresh indicator */}
			<AutoRefreshIndicator
				interval={REFRESH_INTERVAL}
				lastRefresh={lastRefresh}
				onManualRefresh={handleManualRefresh}
				isRefreshing={isManualRefreshing}
			/>

			{/* Health and Readiness Cards */}
			<div className="grid gap-6 md:grid-cols-2">
				<HealthStatusCard
					health={health}
					isLoading={healthLoading}
					error={healthError}
					lastChecked={lastRefresh}
					onRetry={refreshHealth}
				/>

				<ReadinessCard
					readiness={readiness}
					isLoading={readinessLoading}
					error={readinessError}
				/>
			</div>

			{/* Metrics Display (placeholder for future) */}
			<MetricsDisplay />
		</div>
	);
}

/**
 * Custom hook for fetching health status
 */
function useFetchHealth(refreshInterval: number) {
	const [data, setData] = React.useState<HealthResponse | undefined>();
	const [error, setError] = React.useState<Error | undefined>();
	const [isLoading, setIsLoading] = React.useState(true);

	const fetchHealth = React.useCallback(async () => {
		try {
			const response = await fetch("/api/health");
			if (!response.ok) {
				throw new Error(`HTTP ${response.status}: ${response.statusText}`);
			}
			const data = await response.json();
			setData(data);
			setError(undefined);
			return data;
		} catch (err) {
			const error = err instanceof Error ? err : new Error("Erro desconhecido");
			setError(error);
			throw error;
		} finally {
			setIsLoading(false);
		}
	}, []);

	// Initial fetch
	React.useEffect(() => {
		fetchHealth();
	}, [fetchHealth]);

	// Auto-refresh
	React.useEffect(() => {
		const interval = setInterval(() => {
			fetchHealth();
		}, refreshInterval);

		return () => clearInterval(interval);
	}, [fetchHealth, refreshInterval]);

	return {
		data,
		error,
		isLoading,
		mutate: fetchHealth,
	};
}

/**
 * Custom hook for fetching readiness status
 */
function useFetchReadiness(refreshInterval: number) {
	const [data, setData] = React.useState<ReadinessResponse | undefined>();
	const [error, setError] = React.useState<Error | undefined>();
	const [isLoading, setIsLoading] = React.useState(true);

	const fetchReadiness = React.useCallback(async () => {
		try {
			const response = await fetch("/api/ready");
			if (!response.ok) {
				throw new Error(`HTTP ${response.status}: ${response.statusText}`);
			}
			const data = await response.json();
			setData(data);
			setError(undefined);
			return data;
		} catch (err) {
			const error = err instanceof Error ? err : new Error("Erro desconhecido");
			setError(error);
			throw error;
		} finally {
			setIsLoading(false);
		}
	}, []);

	// Initial fetch
	React.useEffect(() => {
		fetchReadiness();
	}, [fetchReadiness]);

	// Auto-refresh
	React.useEffect(() => {
		const interval = setInterval(() => {
			fetchReadiness();
		}, refreshInterval);

		return () => clearInterval(interval);
	}, [fetchReadiness, refreshInterval]);

	return {
		data,
		error,
		isLoading,
		mutate: fetchReadiness,
	};
}

/**
 * Alternative: Using SWR for data fetching
 *
 * If you prefer to use SWR (already installed in package.json):
 *
 * ```tsx
 * import useSWR from 'swr'
 *
 * const fetcher = (url: string) => fetch(url).then(r => r.json())
 *
 * export function HealthDashboard() {
 *   const { data: health, error: healthError, isLoading: healthLoading, mutate: refreshHealth } =
 *     useSWR<HealthResponse>('/api/health', fetcher, {
 *       refreshInterval: REFRESH_INTERVAL,
 *       revalidateOnFocus: true,
 *     })
 *
 *   const { data: readiness, error: readinessError, isLoading: readinessLoading, mutate: refreshReadiness } =
 *     useSWR<ReadinessResponse>('/api/ready', fetcher, {
 *       refreshInterval: REFRESH_INTERVAL,
 *       revalidateOnFocus: true,
 *     })
 *
 *   // ... rest of the component
 * }
 * ```
 */
