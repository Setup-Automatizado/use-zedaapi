/**
 * Instance Names Hook
 *
 * Maps instance UUIDs to friendly display names with phone and avatar.
 * Supports large instance counts (5000+) with pagination and caching.
 *
 * @module hooks/use-instance-names
 */

"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import useSWR from "swr";
import { formatPhoneNumber } from "@/lib/phone";

interface Instance {
	id: string;
	instanceId?: string;
	name?: string;
	sessionName?: string;
	token?: string;
	instanceToken?: string;
	phoneConnected?: boolean;
	whatsappConnected?: boolean;
}

interface DeviceInfo {
	phone: string;
	imgUrl?: string;
	name: string;
	device?: {
		sessionName: string;
		device_model: string;
		wa_version: string;
		platform: string;
		os_version: string;
		device_manufacturer: string;
	};
	isBusiness?: boolean;
}

interface DeviceBatchResponse {
	devices: Record<string, DeviceInfo | null>;
}

export interface InstanceInfo {
	id: string;
	name: string;
	sessionName?: string;
	phone?: string;
	formattedPhone?: string;
	avatarUrl?: string;
	connected: boolean;
	isBusiness?: boolean;
}

const fetcher = (url: string) => fetch(url).then((res) => res.json());

/**
 * Fetch all instances with pagination support for large counts
 */
async function fetchAllInstances(): Promise<Instance[]> {
	const PAGE_SIZE = 100;
	const allInstances: Instance[] = [];
	let page = 1;
	let hasMore = true;

	while (hasMore) {
		const response = await fetcher(`/api/instances?page=${page}&pageSize=${PAGE_SIZE}`);
		const instances = response.content || [];
		allInstances.push(...instances);

		// Check if there are more pages
		hasMore = instances.length === PAGE_SIZE && allInstances.length < response.total;
		page++;

		// Safety limit to prevent infinite loops
		if (page > 100) break;
	}

	return allInstances;
}

const deviceFetcher = async (instances: Instance[]): Promise<DeviceBatchResponse> => {
	const connectedInstances = instances
		.filter((i) => i.whatsappConnected && i.phoneConnected)
		.map((i) => ({
			instanceId: i.id || i.instanceId || "",
			instanceToken: i.token || i.instanceToken || "",
			connected: true,
		}));

	if (connectedInstances.length === 0) {
		return { devices: {} };
	}

	// Batch requests to avoid overwhelming the API
	const BATCH_SIZE = 50;
	const allDevices: Record<string, DeviceInfo | null> = {};

	for (let i = 0; i < connectedInstances.length; i += BATCH_SIZE) {
		const batch = connectedInstances.slice(i, i + BATCH_SIZE);

		try {
			const response = await fetch("/api/instances/device", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ instances: batch }),
			});

			if (response.ok) {
				const data = await response.json();
				Object.assign(allDevices, data.devices || {});
			}
		} catch {
			// Continue with next batch on error
		}
	}

	return { devices: allDevices };
};

/**
 * Hook to get instance name mapping with phone and avatar
 * Supports 5000+ instances with automatic pagination
 */
export function useInstanceNames() {
	const [allInstances, setAllInstances] = useState<Instance[]>([]);
	const [isLoadingInstances, setIsLoadingInstances] = useState(true);
	const [instancesError, setInstancesError] = useState<Error | null>(null);

	// Fetch all instances with pagination on mount
	useEffect(() => {
		let mounted = true;

		async function loadInstances() {
			try {
				setIsLoadingInstances(true);
				const instances = await fetchAllInstances();
				if (mounted) {
					setAllInstances(instances);
					setInstancesError(null);
				}
			} catch (error) {
				if (mounted) {
					setInstancesError(error instanceof Error ? error : new Error("Failed to load instances"));
				}
			} finally {
				if (mounted) {
					setIsLoadingInstances(false);
				}
			}
		}

		loadInstances();

		return () => {
			mounted = false;
		};
	}, []);

	// Fetch device info for connected instances
	const { data: deviceData } = useSWR<DeviceBatchResponse>(
		allInstances.length > 0 ? ["devices", allInstances] : null,
		([, instances]: [string, Instance[]]) => deviceFetcher(instances),
		{
			revalidateOnFocus: false,
			dedupingInterval: 120000, // Cache for 2 minutes
		}
	);

	const instanceMap = useMemo(() => {
		const map = new Map<string, InstanceInfo>();
		for (const instance of allInstances) {
			const id = instance.id || instance.instanceId || "";
			const device = deviceData?.devices?.[id];
			const phone = device?.phone;

			map.set(id, {
				id,
				name: instance.sessionName || instance.name || truncateId(id),
				sessionName: instance.sessionName,
				phone,
				formattedPhone: phone ? formatPhoneNumber(phone) : undefined,
				avatarUrl: device?.imgUrl,
				connected: Boolean(instance.whatsappConnected && instance.phoneConnected),
				isBusiness: device?.isBusiness,
			});
		}
		return map;
	}, [allInstances, deviceData]);

	/**
	 * Get friendly display name for an instance
	 */
	const getDisplayName = useCallback((instanceId: string): string => {
		const instance = instanceMap.get(instanceId);
		if (instance) {
			if (instance.name && instance.name !== "string") {
				return instance.name;
			}
		}
		return truncateId(instanceId);
	}, [instanceMap]);

	/**
	 * Get short display name (max 15 chars)
	 */
	const getShortName = useCallback((instanceId: string): string => {
		const name = getDisplayName(instanceId);
		if (name.length <= 15) return name;
		return `${name.slice(0, 12)}...`;
	}, [getDisplayName]);

	/**
	 * Get phone number for an instance
	 */
	const getPhone = useCallback((instanceId: string): string | undefined => {
		return instanceMap.get(instanceId)?.phone;
	}, [instanceMap]);

	/**
	 * Get formatted phone number for an instance
	 */
	const getFormattedPhone = useCallback((instanceId: string): string | undefined => {
		return instanceMap.get(instanceId)?.formattedPhone;
	}, [instanceMap]);

	/**
	 * Get avatar URL for an instance
	 */
	const getAvatarUrl = useCallback((instanceId: string): string | undefined => {
		return instanceMap.get(instanceId)?.avatarUrl;
	}, [instanceMap]);

	/**
	 * Get full instance info
	 */
	const getInstanceInfo = useCallback((instanceId: string): InstanceInfo | undefined => {
		return instanceMap.get(instanceId);
	}, [instanceMap]);

	/**
	 * Search instances by name or phone
	 */
	const searchInstances = useCallback((query: string): InstanceInfo[] => {
		if (!query.trim()) return [];

		const lowerQuery = query.toLowerCase().trim();
		const results: InstanceInfo[] = [];

		for (const info of instanceMap.values()) {
			// Search by name
			if (info.name.toLowerCase().includes(lowerQuery)) {
				results.push(info);
				continue;
			}
			// Search by phone (raw or formatted)
			if (info.phone?.includes(lowerQuery) || info.formattedPhone?.toLowerCase().includes(lowerQuery)) {
				results.push(info);
				continue;
			}
			// Search by ID
			if (info.id.toLowerCase().includes(lowerQuery)) {
				results.push(info);
			}
		}

		return results;
	}, [instanceMap]);

	/**
	 * Get all instances with display info
	 */
	const instancesWithInfo = useMemo(() => {
		return Array.from(instanceMap.values());
	}, [instanceMap]);

	return {
		instanceMap,
		getDisplayName,
		getShortName,
		getPhone,
		getFormattedPhone,
		getAvatarUrl,
		getInstanceInfo,
		searchInstances,
		instancesWithInfo,
		isLoading: isLoadingInstances,
		error: instancesError,
		totalInstances: allInstances.length,
	};
}

/**
 * Truncate UUID for display
 */
function truncateId(id: string): string {
	if (id.length <= 12) return id;
	return `${id.slice(0, 8)}...${id.slice(-4)}`;
}

/**
 * Format instance ID with optional name
 */
export function formatInstanceLabel(
	instanceId: string,
	name?: string | null
): string {
	if (name && name !== "string" && name.trim()) {
		return name;
	}
	return truncateId(instanceId);
}
