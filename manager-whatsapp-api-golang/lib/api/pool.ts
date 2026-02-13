/**
 * Proxy Pool API functions
 *
 * Provides type-safe wrappers for proxy pool management operations.
 * Partner routes use Partner-Token auth, instance routes use instanceId/instanceToken.
 *
 * @module lib/api/pool
 */

import "server-only";
import type {
	AssignGroupRequest,
	AssignPoolProxyRequest,
	BulkAssignRequest,
	BulkAssignResult,
	CreateGroupRequest,
	CreateProviderRequest,
	PoolAssignment,
	PoolGroup,
	PoolProvider,
	PoolProxyListResponse,
	PoolStats,
	UpdateProviderRequest,
} from "@/types/pool";
import { api } from "./client";

// ---------------------------------------------------------------------------
// Partner-auth routes (admin)
// ---------------------------------------------------------------------------

export async function createProvider(
	data: CreateProviderRequest,
): Promise<PoolProvider> {
	return api.post<PoolProvider>("/proxy-providers", data, {
		usePartnerToken: true,
	});
}

export async function listProviders(): Promise<PoolProvider[]> {
	return api.get<PoolProvider[]>("/proxy-providers", {
		usePartnerToken: true,
	});
}

export async function getProvider(id: string): Promise<PoolProvider> {
	return api.get<PoolProvider>(`/proxy-providers/${id}`, {
		usePartnerToken: true,
	});
}

export async function updateProvider(
	id: string,
	data: UpdateProviderRequest,
): Promise<PoolProvider> {
	return api.put<PoolProvider>(`/proxy-providers/${id}`, data, {
		usePartnerToken: true,
	});
}

export async function deleteProvider(id: string): Promise<void> {
	return api.delete<void>(`/proxy-providers/${id}`, {
		usePartnerToken: true,
	});
}

export async function triggerProviderSync(
	id: string,
): Promise<{ status: string }> {
	return api.post<{ status: string }>(
		`/proxy-providers/${id}/sync`,
		undefined,
		{ usePartnerToken: true },
	);
}

export async function getPoolStats(): Promise<PoolStats> {
	return api.get<PoolStats>("/proxy-pool/stats", {
		usePartnerToken: true,
	});
}

export async function listPoolProxies(params?: {
	providerId?: string;
	status?: string;
	limit?: number;
	offset?: number;
}): Promise<PoolProxyListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.providerId) searchParams.set("provider_id", params.providerId);
	if (params?.status) searchParams.set("status", params.status);
	if (params?.limit) searchParams.set("limit", params.limit.toString());
	if (params?.offset) searchParams.set("offset", params.offset.toString());
	const query = searchParams.toString();
	return api.get<PoolProxyListResponse>(
		`/proxy-pool${query ? `?${query}` : ""}`,
		{ usePartnerToken: true },
	);
}

export async function bulkAssignPoolProxies(
	data: BulkAssignRequest,
): Promise<BulkAssignResult> {
	return api.post<BulkAssignResult>("/proxy-pool/bulk-assign", data, {
		usePartnerToken: true,
	});
}

export async function createGroup(
	data: CreateGroupRequest,
): Promise<PoolGroup> {
	return api.post<PoolGroup>("/proxy-groups", data, {
		usePartnerToken: true,
	});
}

export async function listGroups(): Promise<PoolGroup[]> {
	return api.get<PoolGroup[]>("/proxy-groups", {
		usePartnerToken: true,
	});
}

export async function deleteGroup(id: string): Promise<void> {
	return api.delete<void>(`/proxy-groups/${id}`, {
		usePartnerToken: true,
	});
}

// ---------------------------------------------------------------------------
// Instance-auth routes
// ---------------------------------------------------------------------------

export async function assignPoolProxy(
	instanceId: string,
	instanceToken: string,
	data: AssignPoolProxyRequest,
): Promise<PoolAssignment> {
	return api.post<PoolAssignment>("/proxy/pool/assign", data, {
		instanceId,
		instanceToken,
	});
}

export async function releasePoolProxy(
	instanceId: string,
	instanceToken: string,
): Promise<{ status: string }> {
	return api.delete<{ status: string }>("/proxy/pool/release", {
		instanceId,
		instanceToken,
	});
}

export async function getPoolAssignment(
	instanceId: string,
	instanceToken: string,
): Promise<PoolAssignment> {
	return api.get<PoolAssignment>("/proxy/pool/assignment", {
		instanceId,
		instanceToken,
	});
}

export async function assignToGroup(
	instanceId: string,
	instanceToken: string,
	data: AssignGroupRequest,
): Promise<PoolAssignment> {
	return api.post<PoolAssignment>("/proxy/pool/assign-group", data, {
		instanceId,
		instanceToken,
	});
}
