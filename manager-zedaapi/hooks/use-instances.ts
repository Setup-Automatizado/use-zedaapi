"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

// =============================================================================
// Instance Hooks (React Query)
// =============================================================================

interface Instance {
	id: string;
	name: string;
	status: string;
	phone: string | null;
	whatsappConnected: boolean;
	profilePicUrl: string | null;
	profileName: string | null;
	zedaapiInstanceId: string;
	lastSyncAt: string | null;
	createdAt: string;
	subscription: {
		plan: { name: string };
	} | null;
}

interface InstancesResponse {
	items: Instance[];
	total: number;
	page: number;
	pageSize: number;
	totalPages: number;
	hasMore: boolean;
}

interface InstanceStatus {
	connected: boolean;
	connectionStatus: string;
	smartphoneConnected: boolean;
	error: string | null;
}

async function fetchApi<T>(url: string, init?: RequestInit): Promise<T> {
	const res = await fetch(url, init);
	if (!res.ok) {
		const body = await res
			.json()
			.catch(() => ({ error: "Request failed" }));
		throw new Error(body.error || `HTTP ${res.status}`);
	}
	return res.json();
}

/**
 * List instances for the current user.
 */
export function useInstances(page = 1, pageSize = 10) {
	return useQuery<InstancesResponse>({
		queryKey: ["instances", page, pageSize],
		queryFn: () =>
			fetchApi(`/api/instances?page=${page}&pageSize=${pageSize}`),
	});
}

/**
 * Get a single instance by ID.
 */
export function useInstance(id: string | undefined) {
	return useQuery<Instance>({
		queryKey: ["instance", id],
		queryFn: () => fetchApi(`/api/instances/${id}`),
		enabled: !!id,
	});
}

/**
 * Poll instance status (refetches every 10s when enabled).
 */
export function useInstanceStatus(id: string | undefined, enabled = true) {
	return useQuery<InstanceStatus>({
		queryKey: ["instance-status", id],
		queryFn: () => fetchApi(`/api/instances/${id}/status`),
		enabled: !!id && enabled,
		refetchInterval: enabled ? 10_000 : false,
	});
}

/**
 * Create a new instance.
 */
export function useCreateInstance() {
	const queryClient = useQueryClient();

	return useMutation<
		Instance,
		Error,
		{ name: string; subscriptionId: string }
	>({
		mutationFn: (data) =>
			fetchApi("/api/instances", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify(data),
			}),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["instances"] });
		},
	});
}

/**
 * Delete (deprovision) an instance.
 */
export function useDeleteInstance() {
	const queryClient = useQueryClient();

	return useMutation<void, Error, string>({
		mutationFn: (id) =>
			fetchApi(`/api/instances/${id}`, { method: "DELETE" }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["instances"] });
		},
	});
}
