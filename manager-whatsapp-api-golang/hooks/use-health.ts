/**
 * Health Check Hook with Polling
 *
 * Fetches service health and readiness status with automatic polling.
 * Used for monitoring service dependencies and overall system health.
 *
 * @example
 * ```tsx
 * const { health, readiness, isHealthy, isLoading, error, refresh } = useHealth({
 *   interval: 30000,
 *   enabled: true
 * });
 * ```
 */

"use client";

import { usePolling } from "./use-polling";
import type { HealthResponse, ReadinessResponse } from "@/types";

/**
 * Health check hook options
 */
export interface UseHealthOptions {
	/**
	 * Enable or disable polling
	 * @default true
	 */
	enabled?: boolean;

	/**
	 * Polling interval in milliseconds
	 * @default 30000 (30 seconds)
	 */
	interval?: number;

	/**
	 * Fetch readiness status in addition to health
	 * @default true
	 */
	includeReadiness?: boolean;
}

/**
 * Health check hook result
 */
export interface UseHealthResult {
	/** Basic health status */
	health: HealthResponse | undefined;

	/** Detailed readiness status with component checks */
	readiness: ReadinessResponse | undefined;

	/** Overall system health (true if all components are healthy) */
	isHealthy: boolean;

	/** System is degraded but operational */
	isDegraded: boolean;

	/** Loading state */
	isLoading: boolean;

	/** Error object if request failed */
	error: Error | undefined;

	/** Revalidation state */
	isValidating: boolean;

	/** Manual refresh function */
	refresh: () => Promise<void>;
}

/**
 * Hook to fetch service health status with polling
 *
 * Features:
 * - Polls health and readiness endpoints
 * - Configurable polling interval
 * - Component-level health checks
 * - Error handling and retry logic
 *
 * @param options - Health check configuration options
 * @returns Health and readiness status with component details
 */
export function useHealth(options: UseHealthOptions = {}): UseHealthResult {
	const {
		enabled = true,
		interval = 30000,
		includeReadiness = true,
	} = options;

	// Fetch basic health status
	const {
		data: health,
		error: healthError,
		isLoading: healthLoading,
		isValidating: healthValidating,
		mutate: mutateHealth,
	} = usePolling<HealthResponse>("/api/health", {
		interval,
		enabled,
		dedupingInterval: 5000,
	});

	// Fetch detailed readiness status
	const {
		data: readiness,
		error: readinessError,
		isLoading: readinessLoading,
		isValidating: readinessValidating,
		mutate: mutateReadiness,
	} = usePolling<ReadinessResponse>(
		includeReadiness ? "/api/health/ready" : null,
		{
			interval,
			enabled: enabled && includeReadiness,
			dedupingInterval: 5000,
		},
	);

	// Determine overall health status
	const isHealthy = !!(
		health?.status === "ok" &&
		(!includeReadiness ||
			(readiness?.ready &&
				readiness.checks.database.status === "healthy" &&
				readiness.checks.redis.status === "healthy"))
	);

	// Determine if system is degraded
	const isDegraded = !!(
		health?.status === "ok" &&
		readiness?.ready &&
		(readiness.checks.database.status === "degraded" ||
			readiness.checks.redis.status === "degraded")
	);

	return {
		health,
		readiness,
		isHealthy,
		isDegraded,
		isLoading: healthLoading || (includeReadiness && readinessLoading),
		error: healthError || readinessError,
		isValidating:
			healthValidating || (includeReadiness && readinessValidating),
		refresh: async () => {
			await Promise.all(
				[mutateHealth(), includeReadiness && mutateReadiness()].filter(
					Boolean,
				),
			);
		},
	};
}
