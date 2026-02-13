/**
 * Instances List Hook
 *
 * Fetches paginated list of WhatsApp instances with filtering and sorting.
 * Uses SWR for caching and automatic revalidation.
 *
 * @example
 * ```tsx
 * const { instances, pagination, isLoading, error, mutate } = useInstances({
 *   page: 1,
 *   pageSize: 20,
 *   query: 'production',
 *   status: 'connected'
 * });
 * ```
 */

"use client";

import useSWR from "swr";
import type { InstanceListResponse } from "@/types";

/**
 * Instance list filter options
 */
export interface UseInstancesParams {
	/**
	 * Current page number (1-indexed)
	 * @default 1
	 */
	page?: number;

	/**
	 * Items per page
	 * @default 10
	 */
	pageSize?: number;

	/**
	 * Search query for instance name
	 */
	query?: string;

	/**
	 * Filter by connection status
	 */
	status?: "connected" | "disconnected" | "all";

	/**
	 * Filter by subscription status
	 */
	subscription?: "active" | "expired" | "all";

	/**
	 * Sort field
	 */
	sortBy?: "name" | "created" | "due";

	/**
	 * Sort direction
	 */
	sortOrder?: "asc" | "desc";
}

/**
 * Instance list hook result
 */
export interface UseInstancesResult {
	/** Array of instances for current page */
	instances: InstanceListResponse["content"] | undefined;

	/** Pagination metadata */
	pagination: Omit<InstanceListResponse, "content"> | undefined;

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
 * Build query string from parameters
 */
function buildQueryString(params: UseInstancesParams): string {
	const searchParams = new URLSearchParams();

	if (params.page) searchParams.set("page", params.page.toString());
	if (params.pageSize)
		searchParams.set("pageSize", params.pageSize.toString());
	if (params.query) searchParams.set("query", params.query);
	if (params.status && params.status !== "all")
		searchParams.set("status", params.status);
	if (params.subscription && params.subscription !== "all")
		searchParams.set("subscription", params.subscription);
	if (params.sortBy) searchParams.set("sortBy", params.sortBy);
	if (params.sortOrder) searchParams.set("sortOrder", params.sortOrder);

	const queryString = searchParams.toString();
	return queryString ? `?${queryString}` : "";
}

/**
 * Fetcher for instances list
 */
async function fetchInstances(url: string): Promise<InstanceListResponse> {
	const response = await fetch(url, {
		method: "GET",
		headers: {
			"Content-Type": "application/json",
		},
		cache: "no-store",
	});

	const data = await response.json();

	if (!response.ok) {
		// If the response includes a valid shape (content array), use it
		if (data && Array.isArray(data.content)) {
			return data as InstanceListResponse;
		}
		const error = new Error(
			data?.error || `Failed to fetch instances: ${response.statusText}`,
		) as Error & { status?: number };
		error.status = response.status;
		throw error;
	}

	// Ensure content is always an array
	return {
		...data,
		content: data.content ?? [],
	} as InstanceListResponse;
}

/**
 * Hook to fetch paginated list of instances
 *
 * Features:
 * - Pagination support
 * - Search and filtering
 * - Sorting
 * - Automatic caching and revalidation
 * - Error handling
 *
 * @param params - Filter and pagination parameters
 * @returns Instance list with pagination metadata
 */
export function useInstances(
	params: UseInstancesParams = {},
): UseInstancesResult {
	const queryString = buildQueryString(params);
	const endpoint = `/api/instances${queryString}`;

	const { data, error, isLoading, isValidating, mutate } =
		useSWR<InstanceListResponse>(endpoint, fetchInstances, {
			revalidateOnFocus: true,
			revalidateOnReconnect: true,
			dedupingInterval: 5000,
			keepPreviousData: true,
		});

	return {
		instances: data?.content ?? undefined,
		pagination: data
			? {
					total: data.total ?? 0,
					totalPage: data.totalPage ?? 0,
					pageSize: data.pageSize ?? 10,
					page: data.page ?? 1,
				}
			: undefined,
		isLoading,
		error,
		isValidating,
		mutate: async () => {
			await mutate();
		},
	};
}
