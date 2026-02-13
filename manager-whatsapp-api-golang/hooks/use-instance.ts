/**
 * Single Instance Hook
 *
 * Fetches a single WhatsApp instance by ID.
 * Uses SWR for caching and automatic revalidation.
 *
 * @example
 * ```tsx
 * const { instance, isLoading, error, mutate } = useInstance(instanceId);
 * ```
 */

"use client";

import useSWR from "swr";
import type { Instance } from "@/types";

/**
 * Single instance hook result
 */
export interface UseInstanceResult {
	/** Instance data */
	instance: Instance | undefined;

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
 * Fetcher for single instance
 */
async function fetchInstance(url: string): Promise<Instance> {
	const response = await fetch(url, {
		method: "GET",
		headers: {
			"Content-Type": "application/json",
		},
		cache: "no-store",
	});

	if (!response.ok) {
		const error = new Error(
			`Failed to fetch instance: ${response.statusText}`,
		) as Error & { status?: number };
		error.status = response.status;
		throw error;
	}

	return response.json();
}

/**
 * Hook to fetch a single instance by ID
 *
 * Features:
 * - Automatic caching and revalidation
 * - Error handling
 * - Focus revalidation
 * - Network error recovery
 *
 * @param instanceId - Instance ID (UUID)
 * @returns Instance data with loading and error states
 */
export function useInstance(
	instanceId: string | null | undefined,
): UseInstanceResult {
	const endpoint = instanceId ? `/api/instances/${instanceId}` : null;

	const { data, error, isLoading, isValidating, mutate } = useSWR<Instance>(
		endpoint,
		fetchInstance,
		{
			revalidateOnFocus: true,
			revalidateOnReconnect: true,
			dedupingInterval: 3000,
			shouldRetryOnError: true,
			errorRetryCount: 3,
			errorRetryInterval: 2000,
		},
	);

	return {
		instance: data,
		isLoading,
		error,
		isValidating,
		mutate: async () => {
			await mutate();
		},
	};
}
