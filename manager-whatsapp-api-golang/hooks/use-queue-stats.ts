/**
 * Queue Statistics Hook
 *
 * Fetches queue statistics for a WhatsApp instance.
 * Uses SWR for caching and automatic revalidation.
 *
 * @example
 * ```tsx
 * const { stats, isLoading, error, mutate } = useQueueStats(instanceId, instanceToken);
 * ```
 */

"use client";

import useSWR from "swr";
import { getQueueStats } from "@/lib/api/queue";
import type { QueueStats } from "@/types/queue";

/**
 * Queue stats hook result
 */
export interface UseQueueStatsResult {
	/** Queue statistics data */
	stats: QueueStats | undefined;

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
 * Hook to fetch queue statistics for an instance
 *
 * Features:
 * - Automatic caching and revalidation (30s interval)
 * - Error handling
 * - Focus revalidation
 * - Network error recovery
 *
 * @param instanceId - Instance ID (UUID)
 * @param instanceToken - Instance authentication token
 * @returns Queue statistics with loading and error states
 */
export function useQueueStats(
	instanceId: string | null | undefined,
	instanceToken: string | null | undefined,
): UseQueueStatsResult {
	const key =
		instanceId && instanceToken
			? `/instances/${instanceId}/queue/stats`
			: null;

	const { data, error, isLoading, isValidating, mutate } = useSWR<QueueStats>(
		key,
		() => getQueueStats(instanceId!, instanceToken!),
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
