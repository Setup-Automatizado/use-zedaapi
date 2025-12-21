/**
 * Instance Names Hook
 *
 * Maps instance UUIDs to friendly display names with phone and avatar.
 * Fetches instance data and device info for connected instances.
 *
 * @module hooks/use-instance-names
 */

"use client";

import { useCallback, useMemo } from "react";
import useSWR from "swr";

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

interface InstancesResponse {
	content: Instance[];
	total: number;
}

interface DeviceBatchResponse {
	devices: Record<string, DeviceInfo | null>;
}

export interface InstanceInfo {
	id: string;
	name: string;
	sessionName?: string;
	phone?: string;
	avatarUrl?: string;
	connected: boolean;
	isBusiness?: boolean;
}

const fetcher = (url: string) => fetch(url).then((res) => res.json());

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

	const response = await fetch("/api/instances/device", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ instances: connectedInstances }),
	});

	if (!response.ok) {
		return { devices: {} };
	}

	return response.json();
};

/**
 * Hook to get instance name mapping with phone and avatar
 */
export function useInstanceNames() {
	const { data, error, isLoading } = useSWR<InstancesResponse>(
		"/api/instances?pageSize=100",
		fetcher,
		{
			revalidateOnFocus: false,
			dedupingInterval: 60000, // Cache for 1 minute
		}
	);

	// Fetch device info for connected instances
	const { data: deviceData } = useSWR<DeviceBatchResponse>(
		data?.content ? ["devices", data.content] : null,
		([, instances]: [string, Instance[]]) => deviceFetcher(instances),
		{
			revalidateOnFocus: false,
			dedupingInterval: 120000, // Cache for 2 minutes
		}
	);

	const instanceMap = useMemo(() => {
		const map = new Map<string, InstanceInfo>();
		if (data?.content) {
			for (const instance of data.content) {
				const id = instance.id || instance.instanceId || "";
				const device = deviceData?.devices?.[id];

				map.set(id, {
					id,
					name: instance.sessionName || instance.name || truncateId(id),
					sessionName: instance.sessionName,
					phone: device?.phone,
					avatarUrl: device?.imgUrl,
					connected: Boolean(instance.whatsappConnected && instance.phoneConnected),
					isBusiness: device?.isBusiness,
				});
			}
		}
		return map;
	}, [data, deviceData]);

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
	 * Format phone number for display (e.g., +55 11 99999-9999)
	 */
	const formatPhone = useCallback((phone: string | undefined): string => {
		if (!phone) return "";
		// Remove all non-digits
		const digits = phone.replace(/\D/g, "");
		if (digits.length < 10) return phone;

		// Format Brazilian numbers
		if (digits.startsWith("55") && digits.length >= 12) {
			const country = digits.slice(0, 2);
			const area = digits.slice(2, 4);
			const part1 = digits.slice(4, digits.length - 4);
			const part2 = digits.slice(-4);
			return `+${country} ${area} ${part1}-${part2}`;
		}

		// Generic international format
		if (digits.length > 10) {
			return `+${digits.slice(0, digits.length - 10)} ${digits.slice(-10, -4)}-${digits.slice(-4)}`;
		}

		return phone;
	}, []);

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
		getAvatarUrl,
		getInstanceInfo,
		formatPhone,
		instancesWithInfo,
		isLoading,
		error,
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
