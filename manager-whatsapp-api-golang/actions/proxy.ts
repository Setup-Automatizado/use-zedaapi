/**
 * Proxy Configuration Server Actions
 *
 * Server Actions for proxy management including configuration,
 * testing, swapping, and removal.
 *
 * @module actions/proxy
 */

"use server";

import { revalidatePath } from "next/cache";
import {
	getProxy as apiGetProxy,
	getProxyHealth as apiGetProxyHealth,
	removeProxy as apiRemoveProxy,
	swapProxy as apiSwapProxy,
	testProxy as apiTestProxy,
	updateProxy as apiUpdateProxy,
} from "@/lib/api/proxy";
import {
	ProxyConfigSchema,
	ProxySwapSchema,
	ProxyTestSchema,
} from "@/schemas/proxy";
import type { ActionResult } from "@/types";
import { error, success, validationError } from "@/types";
import type {
	ProxyHealthResponse,
	ProxyResponse,
	ProxyTestResponse,
} from "@/types/proxy";

/**
 * Updates the proxy configuration for an instance
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param data - Proxy configuration form values
 * @returns Action result with updated proxy config
 */
export async function updateProxyConfig(
	instanceId: string,
	instanceToken: string,
	data: {
		proxyUrl: string;
		noWebsocket: boolean;
		onlyLogin: boolean;
		noMedia: boolean;
	},
): Promise<ActionResult<ProxyResponse>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const validation = ProxyConfigSchema.safeParse(data);
		if (!validation.success) {
			const errors: Record<string, string[]> = {};
			validation.error.issues.forEach((issue) => {
				const path = issue.path[0]?.toString() || "form";
				if (!errors[path]) errors[path] = [];
				errors[path].push(issue.message);
			});
			return validationError(errors);
		}

		const result = await apiUpdateProxy(instanceId, instanceToken, {
			proxyUrl: validation.data.proxyUrl,
			noWebsocket: validation.data.noWebsocket,
			onlyLogin: validation.data.onlyLogin,
			noMedia: validation.data.noMedia,
		});

		revalidatePath(`/instances/${instanceId}`);
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to update proxy";
		return error(message);
	}
}

/**
 * Removes the proxy configuration from an instance
 */
export async function removeProxyConfig(
	instanceId: string,
	instanceToken: string,
): Promise<ActionResult<void>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		await apiRemoveProxy(instanceId, instanceToken);
		revalidatePath(`/instances/${instanceId}`);
		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to remove proxy";
		return error(message);
	}
}

/**
 * Tests proxy connectivity without applying it
 */
export async function testProxyConnection(
	instanceId: string,
	instanceToken: string,
	proxyUrl: string,
): Promise<ActionResult<ProxyTestResponse>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const validation = ProxyTestSchema.safeParse({ proxyUrl });
		if (!validation.success) {
			const errors: Record<string, string[]> = {};
			validation.error.issues.forEach((issue) => {
				const path = issue.path[0]?.toString() || "form";
				if (!errors[path]) errors[path] = [];
				errors[path].push(issue.message);
			});
			return validationError(errors);
		}

		const result = await apiTestProxy(instanceId, instanceToken, {
			proxyUrl: validation.data.proxyUrl,
		});
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to test proxy";
		return error(message);
	}
}

/**
 * Hot-swaps proxy on an active WhatsApp connection
 */
export async function swapProxyConnection(
	instanceId: string,
	instanceToken: string,
	proxyUrl: string,
): Promise<ActionResult<ProxyResponse>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const validation = ProxySwapSchema.safeParse({ proxyUrl });
		if (!validation.success) {
			const errors: Record<string, string[]> = {};
			validation.error.issues.forEach((issue) => {
				const path = issue.path[0]?.toString() || "form";
				if (!errors[path]) errors[path] = [];
				errors[path].push(issue.message);
			});
			return validationError(errors);
		}

		const result = await apiSwapProxy(instanceId, instanceToken, {
			proxyUrl: validation.data.proxyUrl,
		});

		revalidatePath(`/instances/${instanceId}`);
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to swap proxy";
		return error(message);
	}
}

/**
 * Retrieves proxy configuration for an instance
 */
export async function fetchProxyConfig(
	instanceId: string,
	instanceToken: string,
): Promise<ActionResult<ProxyResponse>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const result = await apiGetProxy(instanceId, instanceToken);
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to fetch proxy config";
		return error(message);
	}
}

/**
 * Retrieves proxy health status and logs
 */
export async function fetchProxyHealth(
	instanceId: string,
	instanceToken: string,
): Promise<ActionResult<ProxyHealthResponse>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const result = await apiGetProxyHealth(instanceId, instanceToken);
		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to fetch proxy health";
		return error(message);
	}
}
