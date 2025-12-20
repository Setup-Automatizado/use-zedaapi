/**
 * Instances with Device Info Hook
 *
 * Extends the base useInstances hook to include device information
 * (avatars, phone numbers) for connected instances.
 *
 * @example
 * ```tsx
 * const { instances, deviceMap, isLoading } = useInstancesWithDevice({
 *   page: 1,
 *   pageSize: 10,
 * });
 *
 * // Access device info
 * const avatar = deviceMap[instance.id]?.imgUrl;
 * ```
 */

"use client";

import useSWR from "swr";
import type { DeviceInfo, Instance } from "@/types";
import { type UseInstancesParams, useInstances } from "./use-instances";

export interface DeviceMap {
	[instanceId: string]: DeviceInfo | null;
}

export interface UseInstancesWithDeviceResult {
	/** Array of instances for current page */
	instances: Instance[] | undefined;

	/** Map of instanceId to DeviceInfo */
	deviceMap: DeviceMap;

	/** Pagination metadata */
	pagination:
		| {
				total: number;
				totalPage: number;
				pageSize: number;
				page: number;
		  }
		| undefined;

	/** Loading state for instances */
	isLoading: boolean;

	/** Loading state for device info */
	isLoadingDevices: boolean;

	/** Error object if request failed */
	error: Error | undefined;

	/** Manual revalidation function */
	mutate: () => Promise<void>;
}

interface BatchDeviceRequest {
	instances: Array<{
		instanceId: string;
		instanceToken: string;
		connected: boolean;
	}>;
}

interface BatchDeviceResponse {
	devices: DeviceMap;
}

/**
 * Fetcher for batch device info
 */
async function fetchDevicesBatch(
	instances: Instance[],
): Promise<BatchDeviceResponse> {
	if (!instances || instances.length === 0) {
		return { devices: {} };
	}

	const body: BatchDeviceRequest = {
		instances: instances.map((inst) => ({
			instanceId: inst.instanceId,
			instanceToken: inst.instanceToken,
			connected: inst.whatsappConnected && inst.phoneConnected,
		})),
	};

	const response = await fetch("/api/instances/device", {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
		},
		body: JSON.stringify(body),
	});

	if (!response.ok) {
		throw new Error("Failed to fetch device info");
	}

	return response.json();
}

/**
 * Filter instances by status
 */
function filterInstancesByStatus(
	instances: Instance[] | undefined,
	status: string | undefined,
): Instance[] | undefined {
	if (!instances || !status || status === "all") {
		return instances;
	}

	return instances.filter((instance) => {
		const isConnected = instance.whatsappConnected && instance.phoneConnected;

		if (status === "connected") {
			return isConnected;
		}
		if (status === "disconnected") {
			return !isConnected;
		}
		return true;
	});
}

/**
 * Hook to fetch instances with device information (avatars)
 *
 * This hook:
 * 1. Fetches the paginated list of instances
 * 2. For connected instances, fetches device info in batch
 * 3. Filters instances by status on the client side
 * 4. Returns both instances and a device map for avatar lookup
 */
export function useInstancesWithDevice(
	params: UseInstancesParams = {},
): UseInstancesWithDeviceResult {
	// Extract status for client-side filtering
	const { status, ...apiParams } = params;

	// Fetch base instances (without status filter - API doesn't support it)
	const {
		instances: rawInstances,
		pagination: rawPagination,
		isLoading: isLoadingInstances,
		error: instancesError,
		mutate: mutateInstances,
	} = useInstances(apiParams);

	// Filter instances by status on client side
	const filteredInstances = filterInstancesByStatus(rawInstances, status);

	// Adjust pagination for filtered results
	const pagination = rawPagination
		? {
				...rawPagination,
				total: filteredInstances?.length ?? 0,
				totalPage: 1, // Client-side filtering doesn't support pagination
			}
		: undefined;

	// Build a stable key for device fetching (use raw instances to get all device info)
	const instanceIds = rawInstances?.map((i) => i.instanceId).join(",") || "";

	// Fetch device info for all instances
	const {
		data: deviceData,
		isLoading: isLoadingDevices,
		mutate: mutateDevices,
	} = useSWR<BatchDeviceResponse>(
		rawInstances && rawInstances.length > 0
			? ["devices-batch", instanceIds]
			: null,
		() => fetchDevicesBatch(rawInstances || []),
		{
			revalidateOnFocus: false,
			revalidateOnReconnect: false,
			dedupingInterval: 30000, // Cache for 30 seconds
			keepPreviousData: true,
		},
	);

	return {
		instances: filteredInstances,
		deviceMap: deviceData?.devices || {},
		pagination,
		isLoading: isLoadingInstances,
		isLoadingDevices,
		error: instancesError,
		mutate: async () => {
			await Promise.all([mutateInstances(), mutateDevices()]);
		},
	};
}
