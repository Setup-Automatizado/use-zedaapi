/**
 * Metrics Hook with Polling
 *
 * Fetches dashboard metrics with configurable polling interval.
 * Uses SWR for caching and revalidation.
 *
 * @example
 * ```tsx
 * const { metrics, isLoading, error, refresh } = useMetrics({
 *   interval: 15000,
 *   instanceId: 'abc-123'
 * });
 * ```
 *
 * @module hooks/use-metrics
 */

"use client";

import * as React from "react";
import { DEFAULT_REFRESH_INTERVAL } from "@/lib/metrics/constants";
import type {
	DashboardMetrics,
	MetricsResponse,
	UseMetricsOptions,
	UseMetricsResult,
} from "@/types/metrics";
import { usePolling } from "./use-polling";

/**
 * Hook to fetch dashboard metrics with polling
 *
 * Features:
 * - Configurable polling interval (5s, 15s, 30s, 60s, off)
 * - Optional instance filtering
 * - Automatic refresh on focus
 * - Error retry with backoff
 * - Loading and validation states
 *
 * @param options - Metrics fetch configuration
 * @returns Metrics data, loading states, and refresh function
 */
export function useMetrics(options: UseMetricsOptions = {}): UseMetricsResult {
	const {
		enabled = true,
		interval = DEFAULT_REFRESH_INTERVAL,
		instanceId = null,
	} = options;

	const [lastUpdated, setLastUpdated] = React.useState<Date | undefined>(undefined);

	// Build API URL with optional instance filter
	const apiUrl = React.useMemo(() => {
		const params = new URLSearchParams();
		if (instanceId) {
			params.set("instance_id", instanceId);
		}
		const query = params.toString();
		return `/api/metrics${query ? `?${query}` : ""}`;
	}, [instanceId]);

	// Custom fetcher that extracts data from response
	const fetcher = React.useCallback(async (url: string): Promise<DashboardMetrics | undefined> => {
		const response = await fetch(url, {
			method: "GET",
			headers: {
				"Content-Type": "application/json",
			},
			cache: "no-store",
		});

		if (!response.ok) {
			throw new Error(`HTTP ${response.status}: ${response.statusText}`);
		}

		const json: MetricsResponse = await response.json();

		if (!json.success) {
			throw new Error(json.error || "Failed to fetch metrics");
		}

		return json.data;
	}, []);

	// Use polling hook
	const {
		data: metrics,
		error,
		isLoading,
		isValidating,
		mutate,
	} = usePolling<DashboardMetrics | undefined>(apiUrl, {
		interval: interval || 0,
		enabled,
		fetcher,
		dedupingInterval: 2000,
	});

	// Update lastUpdated when data changes
	React.useEffect(() => {
		if (!isValidating && metrics) {
			setLastUpdated(new Date());
		}
	}, [isValidating, metrics]);

	// Manual refresh function
	const refresh = React.useCallback(async () => {
		await mutate();
	}, [mutate]);

	return {
		metrics,
		isLoading,
		isValidating,
		error,
		lastUpdated,
		refresh,
	};
}

/**
 * Hook to get list of available instances from metrics
 */
export function useMetricsInstances(): {
	instances: string[];
	isLoading: boolean;
	error: Error | undefined;
} {
	const { metrics, isLoading, error } = useMetrics({
		enabled: true,
		interval: 60000, // Less frequent for instance list
	});

	return {
		instances: metrics?.instances ?? [],
		isLoading,
		error,
	};
}
