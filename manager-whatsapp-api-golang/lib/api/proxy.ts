/**
 * Proxy management API functions
 *
 * Provides type-safe wrappers for proxy configuration operations including:
 * - Get/update/remove proxy configuration
 * - Test proxy connectivity
 * - Hot-swap proxy on active connections
 * - Retrieve proxy health logs
 *
 * @module lib/api/proxy
 */

import "server-only";
import type {
	ProxyConfig,
	ProxyHealthResponse,
	ProxyResponse,
	ProxySwapRequest,
	ProxyTestRequest,
	ProxyTestResponse,
	ProxyUpdateRequest,
} from "@/types/proxy";
import { api } from "./client";

/**
 * Gets the current proxy configuration for an instance
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns Current proxy configuration
 */
export async function getProxy(
	instanceId: string,
	instanceToken: string,
): Promise<ProxyResponse> {
	return api.get<ProxyResponse>("/proxy", {
		instanceId,
		instanceToken,
	});
}

/**
 * Updates the proxy configuration for an instance
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param data - Proxy configuration to apply
 * @returns Updated proxy configuration
 */
export async function updateProxy(
	instanceId: string,
	instanceToken: string,
	data: ProxyUpdateRequest,
): Promise<ProxyResponse> {
	return api.put<ProxyResponse>("/update-proxy", data, {
		instanceId,
		instanceToken,
	});
}

/**
 * Removes the proxy configuration from an instance
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns Confirmation
 */
export async function removeProxy(
	instanceId: string,
	instanceToken: string,
): Promise<{ value: boolean }> {
	return api.delete<{ value: boolean }>("/proxy", {
		instanceId,
		instanceToken,
	});
}

/**
 * Tests proxy connectivity without applying it
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param data - Proxy URL to test
 * @returns Test result with reachability and latency
 */
export async function testProxy(
	instanceId: string,
	instanceToken: string,
	data: ProxyTestRequest,
): Promise<ProxyTestResponse> {
	return api.post<ProxyTestResponse>("/proxy/test", data, {
		instanceId,
		instanceToken,
	});
}

/**
 * Hot-swaps the proxy on an active WhatsApp connection
 * The client is disconnected and reconnected through the new proxy.
 * The WhatsApp session is preserved.
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param data - New proxy URL
 * @returns Updated proxy configuration
 */
export async function swapProxy(
	instanceId: string,
	instanceToken: string,
	data: ProxySwapRequest,
): Promise<ProxyResponse> {
	return api.post<ProxyResponse>("/proxy/swap", data, {
		instanceId,
		instanceToken,
	});
}

/**
 * Retrieves proxy health check logs and current status
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns Health status and recent check logs
 */
export async function getProxyHealth(
	instanceId: string,
	instanceToken: string,
): Promise<ProxyHealthResponse> {
	return api.get<ProxyHealthResponse>("/proxy/health", {
		instanceId,
		instanceToken,
	});
}
