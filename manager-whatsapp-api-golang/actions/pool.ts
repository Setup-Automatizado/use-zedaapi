/**
 * Proxy Pool Management Server Actions
 *
 * Server Actions for proxy pool providers, groups, assignments,
 * and pool statistics.
 *
 * @module actions/pool
 */

"use server";

import { revalidatePath } from "next/cache";
import {
	assignPoolProxy as apiAssignPoolProxy,
	assignToGroup as apiAssignToGroup,
	bulkAssignPoolProxies as apiBulkAssign,
	createGroup as apiCreateGroup,
	createProvider as apiCreateProvider,
	deleteGroup as apiDeleteGroup,
	deleteProvider as apiDeleteProvider,
	getPoolAssignment as apiGetPoolAssignment,
	getPoolStats as apiGetPoolStats,
	listGroups as apiListGroups,
	listPoolProxies as apiListPoolProxies,
	listProviders as apiListProviders,
	releasePoolProxy as apiReleasePoolProxy,
	triggerProviderSync as apiTriggerSync,
	updateProvider as apiUpdateProvider,
} from "@/lib/api/pool";
import {
	AssignGroupSchema,
	AssignPoolProxySchema,
	BulkAssignSchema,
	CreateGroupSchema,
	CreateProviderSchema,
	UpdateProviderSchema,
} from "@/schemas/pool";
import type { ActionResult } from "@/types";
import { error, success, validationError } from "@/types";
import type {
	BulkAssignResult,
	PoolAssignment,
	PoolGroup,
	PoolProvider,
	PoolProxyListResponse,
	PoolStats,
} from "@/types/pool";

// ---------------------------------------------------------------------------
// Provider actions (partner auth)
// ---------------------------------------------------------------------------

/**
 * Creates a new proxy provider
 */
export async function createPoolProvider(
	data: Record<string, unknown>,
): Promise<ActionResult<PoolProvider>> {
	try {
		const validation = CreateProviderSchema.safeParse(data);
		if (!validation.success) {
			const errors: Record<string, string[]> = {};
			validation.error.issues.forEach((issue) => {
				const path = issue.path[0]?.toString() || "form";
				if (!errors[path]) errors[path] = [];
				errors[path].push(issue.message);
			});
			return validationError(errors);
		}

		const result = await apiCreateProvider(validation.data);
		revalidatePath("/proxy-pool");
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to create provider";
		return error(message);
	}
}

/**
 * Updates an existing proxy provider
 */
export async function updatePoolProvider(
	id: string,
	data: Record<string, unknown>,
): Promise<ActionResult<PoolProvider>> {
	try {
		if (!id) return error("Provider ID is required");

		const validation = UpdateProviderSchema.safeParse(data);
		if (!validation.success) {
			const errors: Record<string, string[]> = {};
			validation.error.issues.forEach((issue) => {
				const path = issue.path[0]?.toString() || "form";
				if (!errors[path]) errors[path] = [];
				errors[path].push(issue.message);
			});
			return validationError(errors);
		}

		const result = await apiUpdateProvider(id, validation.data);
		revalidatePath("/proxy-pool");
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to update provider";
		return error(message);
	}
}

/**
 * Deletes a proxy provider
 */
export async function deletePoolProvider(
	id: string,
): Promise<ActionResult<void>> {
	try {
		if (!id) return error("Provider ID is required");
		await apiDeleteProvider(id);
		revalidatePath("/proxy-pool");
		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to delete provider";
		return error(message);
	}
}

/**
 * Triggers a sync for a proxy provider
 */
export async function syncPoolProvider(
	id: string,
): Promise<ActionResult<{ status: string }>> {
	try {
		if (!id) return error("Provider ID is required");
		const result = await apiTriggerSync(id);
		revalidatePath("/proxy-pool");
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to trigger sync";
		return error(message);
	}
}

// ---------------------------------------------------------------------------
// Pool stats and proxies (partner auth)
// ---------------------------------------------------------------------------

/**
 * Fetches aggregate pool statistics
 */
export async function fetchPoolStats(): Promise<ActionResult<PoolStats>> {
	try {
		const result = await apiGetPoolStats();
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to fetch pool stats";
		return error(message);
	}
}

/**
 * Fetches all proxy providers
 */
export async function fetchPoolProviders(): Promise<
	ActionResult<PoolProvider[]>
> {
	try {
		const result = await apiListProviders();
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to fetch providers";
		return error(message);
	}
}

/**
 * Fetches pool proxies with optional filters
 */
export async function fetchPoolProxies(params?: {
	providerId?: string;
	status?: string;
	limit?: number;
	offset?: number;
}): Promise<ActionResult<PoolProxyListResponse>> {
	try {
		const result = await apiListPoolProxies(params);
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to fetch pool proxies";
		return error(message);
	}
}

// ---------------------------------------------------------------------------
// Bulk assign (partner auth)
// ---------------------------------------------------------------------------

/**
 * Bulk-assigns pool proxies to all unassigned instances.
 * Uses Redis distributed lock to prevent concurrent operations.
 */
export async function bulkAssignPoolProxies(
	data: Record<string, unknown>,
): Promise<ActionResult<BulkAssignResult>> {
	try {
		const validation = BulkAssignSchema.safeParse(data);
		if (!validation.success) {
			const errors: Record<string, string[]> = {};
			validation.error.issues.forEach((issue) => {
				const path = issue.path[0]?.toString() || "form";
				if (!errors[path]) errors[path] = [];
				errors[path].push(issue.message);
			});
			return validationError(errors);
		}

		const result = await apiBulkAssign(validation.data);
		revalidatePath("/proxy-pool");
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error
				? err.message
				: "Failed to bulk assign proxies";
		return error(message);
	}
}

// ---------------------------------------------------------------------------
// Group actions (partner auth)
// ---------------------------------------------------------------------------

/**
 * Creates a new proxy group
 */
export async function createPoolGroup(
	data: Record<string, unknown>,
): Promise<ActionResult<PoolGroup>> {
	try {
		const validation = CreateGroupSchema.safeParse(data);
		if (!validation.success) {
			const errors: Record<string, string[]> = {};
			validation.error.issues.forEach((issue) => {
				const path = issue.path[0]?.toString() || "form";
				if (!errors[path]) errors[path] = [];
				errors[path].push(issue.message);
			});
			return validationError(errors);
		}

		const result = await apiCreateGroup(validation.data);
		revalidatePath("/proxy-pool");
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to create group";
		return error(message);
	}
}

/**
 * Deletes a proxy group
 */
export async function deletePoolGroup(id: string): Promise<ActionResult<void>> {
	try {
		if (!id) return error("Group ID is required");
		await apiDeleteGroup(id);
		revalidatePath("/proxy-pool");
		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to delete group";
		return error(message);
	}
}

/**
 * Fetches all proxy groups
 */
export async function fetchPoolGroups(): Promise<ActionResult<PoolGroup[]>> {
	try {
		const result = await apiListGroups();
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to fetch groups";
		return error(message);
	}
}

// ---------------------------------------------------------------------------
// Instance-level pool actions
// ---------------------------------------------------------------------------

/**
 * Assigns a pool proxy to an instance
 */
export async function assignInstancePoolProxy(
	instanceId: string,
	instanceToken: string,
	data: Record<string, unknown>,
): Promise<ActionResult<PoolAssignment>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const validation = AssignPoolProxySchema.safeParse(data);
		if (!validation.success) {
			const errors: Record<string, string[]> = {};
			validation.error.issues.forEach((issue) => {
				const path = issue.path[0]?.toString() || "form";
				if (!errors[path]) errors[path] = [];
				errors[path].push(issue.message);
			});
			return validationError(errors);
		}

		const result = await apiAssignPoolProxy(
			instanceId,
			instanceToken,
			validation.data,
		);
		revalidatePath(`/instances/${instanceId}`);
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to assign pool proxy";
		return error(message);
	}
}

/**
 * Releases the pool proxy assignment from an instance
 */
export async function releaseInstancePoolProxy(
	instanceId: string,
	instanceToken: string,
): Promise<ActionResult<void>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		await apiReleasePoolProxy(instanceId, instanceToken);
		revalidatePath(`/instances/${instanceId}`);
		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to release pool proxy";
		return error(message);
	}
}

/**
 * Fetches the current pool assignment for an instance
 */
export async function fetchInstancePoolAssignment(
	instanceId: string,
	instanceToken: string,
): Promise<ActionResult<PoolAssignment>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const result = await apiGetPoolAssignment(instanceId, instanceToken);
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error
				? err.message
				: "Failed to fetch pool assignment";
		return error(message);
	}
}

/**
 * Assigns an instance to a proxy group
 */
export async function assignInstanceToGroup(
	instanceId: string,
	instanceToken: string,
	groupId: string,
): Promise<ActionResult<PoolAssignment>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const validation = AssignGroupSchema.safeParse({ groupId });
		if (!validation.success) {
			const errors: Record<string, string[]> = {};
			validation.error.issues.forEach((issue) => {
				const path = issue.path[0]?.toString() || "form";
				if (!errors[path]) errors[path] = [];
				errors[path].push(issue.message);
			});
			return validationError(errors);
		}

		const result = await apiAssignToGroup(instanceId, instanceToken, {
			groupId: validation.data.groupId,
		});
		revalidatePath(`/instances/${instanceId}`);
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to assign to group";
		return error(message);
	}
}
