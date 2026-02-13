/**
 * Proxy Pool Types
 *
 * Type definitions for proxy pool management including providers,
 * pool proxies, assignments, groups, and statistics.
 *
 * @module types/pool
 */

/** Proxy provider record from proxy_providers table */
export interface PoolProvider {
	id: string;
	name: string;
	providerType: string;
	enabled: boolean;
	priority: number;
	apiEndpoint?: string;
	maxProxies: number;
	maxInstancesPerProxy: number;
	countryCodes: string[];
	rateLimitRpm: number;
	lastSyncAt?: string;
	syncError?: string;
	proxyCount: number;
	createdAt: string;
	updatedAt: string;
}

/** Pool proxy record from proxy_pool table */
export interface PoolProxy {
	id: string;
	providerId: string;
	externalId?: string;
	proxyUrl: string;
	countryCode?: string;
	city?: string;
	status: PoolProxyStatus;
	healthStatus: string;
	healthFailures: number;
	lastHealthCheck?: string;
	assignedCount: number;
	maxAssignments: number;
	valid: boolean;
	lastVerifiedAt?: string;
	createdAt: string;
	updatedAt: string;
}

/** Pool proxy status values */
export type PoolProxyStatus =
	| "available"
	| "assigned"
	| "unhealthy"
	| "retired";

/** Assignment record from proxy_assignments table */
export interface PoolAssignment {
	id: string;
	poolProxyId: string;
	instanceId: string;
	groupId?: string;
	status: string;
	assignedAt: string;
	releasedAt?: string;
	assignedBy: string;
	releaseReason?: string;
	proxyUrl?: string;
}

/** Proxy group record from proxy_groups table */
export interface PoolGroup {
	id: string;
	name: string;
	providerId?: string;
	poolProxyId?: string;
	maxInstances: number;
	countryCode?: string;
	createdAt: string;
	updatedAt: string;
}

/** Pool aggregate statistics */
export interface PoolStats {
	totalProxies: number;
	availableProxies: number;
	assignedProxies: number;
	unhealthyProxies: number;
	retiredProxies: number;
	totalAssignments: number;
	byProvider: ProviderStat[];
}

/** Per-provider statistics */
export interface ProviderStat {
	providerId: string;
	providerName: string;
	total: number;
	available: number;
	assigned: number;
	unhealthy: number;
}

/** Request to create a new provider */
export interface CreateProviderRequest {
	name: string;
	providerType: string;
	enabled: boolean;
	priority: number;
	apiKey: string;
	apiEndpoint?: string;
	maxProxies: number;
	maxInstancesPerProxy: number;
	countryCodes: string[];
	rateLimitRpm: number;
}

/** Request to update an existing provider */
export interface UpdateProviderRequest {
	name?: string;
	enabled?: boolean;
	priority?: number;
	apiKey?: string;
	apiEndpoint?: string;
	maxProxies?: number;
	maxInstancesPerProxy?: number;
	countryCodes?: string[];
	rateLimitRpm?: number;
}

/** Request to assign a pool proxy to an instance */
export interface AssignPoolProxyRequest {
	providerId?: string;
	countryCodes?: string[];
	noWebsocket?: boolean;
	onlyLogin?: boolean;
	noMedia?: boolean;
}

/** Request to assign an instance to a group */
export interface AssignGroupRequest {
	groupId: string;
}

/** Request to create a new group */
export interface CreateGroupRequest {
	name: string;
	providerId?: string;
	maxInstances: number;
	countryCode?: string;
}

/** Request for bulk-assigning pool proxies to multiple instances */
export interface BulkAssignRequest {
	instanceIds?: string[];
	providerId?: string;
	countryCodes?: string[];
}

/** Result of a bulk assignment operation */
export interface BulkAssignResult {
	total: number;
	assigned: number;
	skipped: number;
	failed: number;
	errors?: string[];
}

/** Paginated pool proxy list response */
export interface PoolProxyListResponse {
	data: PoolProxy[];
	total: number;
}

/** Helper to get pool proxy status badge color */
export function getPoolProxyStatusColor(
	status: PoolProxyStatus,
): "green" | "blue" | "red" | "gray" {
	switch (status) {
		case "available":
			return "green";
		case "assigned":
			return "blue";
		case "unhealthy":
			return "red";
		case "retired":
			return "gray";
		default:
			return "gray";
	}
}

/** Helper to get health status color */
export function getPoolHealthColor(status: string): "green" | "red" | "yellow" {
	switch (status) {
		case "healthy":
			return "green";
		case "unhealthy":
			return "red";
		default:
			return "yellow";
	}
}

/** Helper to mask proxy URL credentials for display.
 * Handles socks5:// and other non-standard schemes that new URL() can't parse. */
export function sanitizePoolProxyUrl(url: string): string {
	if (!url) return "";
	// Regex: scheme://user:pass@host → scheme://***:***@host
	const match = url.match(/^([a-z][a-z0-9+.-]*:\/\/)([^@]+)@(.+)$/i);
	if (match) {
		return `${match[1]}***:***@${match[3]}`;
	}
	// Fallback for URLs without credentials
	try {
		const parsed = new URL(url);
		if (parsed.username || parsed.password) {
			parsed.username = "***";
			parsed.password = "***";
		}
		return parsed.toString();
	} catch {
		return url;
	}
}
