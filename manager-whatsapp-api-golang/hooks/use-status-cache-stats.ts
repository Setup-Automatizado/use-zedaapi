/**
 * Status Cache Statistics Hook
 *
 * Fetches status cache statistics for a WhatsApp instance.
 * Uses SWR for caching and automatic revalidation.
 *
 * @example
 * ```tsx
 * const { stats, isLoading, error, mutate } = useStatusCacheStats(instanceId, instanceToken);
 * ```
 */

"use client";

import useSWR from "swr";
import { getStatusCacheStats } from "@/lib/api/status-cache";
import type { StatusCacheStats } from "@/types/status-cache";

/**
 * Status cache stats hook result
 */
export interface UseStatusCacheStatsResult {
	/** Status cache statistics data */
	stats: StatusCacheStats | undefined;

	/** Loading state */
	isLoading: boolean;

	/** Error object if request failed */
	error: Error | undefined;

	/** Revalidation state */
	isValidating: boolean;

	/** Manual revalidation function */
	mutate: () => Promise<void>;
}

/**
 * Hook to fetch status cache statistics for an instance
 *
 * Features:
 * - Automatic caching and revalidation (30s interval)
 * - Error handling
 * - Focus revalidation
 * - Network error recovery
 *
 * @param instanceId - Instance ID (UUID)
 * @param instanceToken - Instance authentication token
 * @returns Status cache statistics with loading and error states
 */
export function useStatusCacheStats(
	instanceId: string | null | undefined,
	instanceToken: string | null | undefined,
): UseStatusCacheStatsResult {
	const key =
		instanceId && instanceToken
			? `/instances/${instanceId}/status-cache/stats`
			: null;

	const { data, error, isLoading, isValidating, mutate } =
		useSWR<StatusCacheStats>(
			key,
			() => getStatusCacheStats(instanceId!, instanceToken!),
			{
				refreshInterval: 30000, // Refresh every 30 seconds
				revalidateOnFocus: true,
				revalidateOnReconnect: true,
				dedupingInterval: 3000,
				shouldRetryOnError: true,
				errorRetryCount: 3,
				errorRetryInterval: 2000,
			},
		);

	return {
		stats: data,
		isLoading,
		error,
		isValidating,
		mutate: async () => {
			await mutate();
		},
	};
}
