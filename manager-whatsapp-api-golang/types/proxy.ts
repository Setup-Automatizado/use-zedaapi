/**
 * Proxy Configuration Types
 *
 * Type definitions for per-instance proxy management, including
 * configuration, health checking, and test results.
 *
 * @module types/proxy
 */

/**
 * Proxy configuration for a WhatsApp instance
 */
export interface ProxyConfig {
	/** Proxy URL (http, https, or socks5) */
	proxyUrl?: string;
	/** Whether the proxy is enabled */
	proxyEnabled: boolean;
	/** Disable WebSocket connections through proxy */
	noWebsocket: boolean;
	/** Only use proxy for login/registration */
	onlyLogin: boolean;
	/** Disable media download/upload through proxy */
	noMedia: boolean;
	/** Current health status: healthy, unhealthy, or unknown */
	healthStatus: string;
	/** Last health check timestamp */
	lastHealthCheck?: string;
	/** Number of consecutive health check failures */
	healthFailures: number;
}

/**
 * Request payload for updating proxy configuration
 */
export interface ProxyUpdateRequest {
	/** Proxy URL (http://, https://, or socks5://) */
	proxyUrl: string;
	/** Disable WebSocket connections through proxy */
	noWebsocket?: boolean;
	/** Only use proxy for login/registration */
	onlyLogin?: boolean;
	/** Disable media download/upload through proxy */
	noMedia?: boolean;
}

/**
 * Request payload for testing a proxy
 */
export interface ProxyTestRequest {
	/** Proxy URL to test */
	proxyUrl: string;
}

/**
 * Response from proxy test endpoint
 */
export interface ProxyTestResponse {
	/** Whether the proxy is reachable */
	reachable: boolean;
	/** Connection latency in milliseconds */
	latencyMs?: number;
	/** Error message if not reachable */
	error?: string;
}

/**
 * Request payload for swapping proxy on active connection
 */
export interface ProxySwapRequest {
	/** New proxy URL to swap to */
	proxyUrl: string;
}

/**
 * Single proxy health check log entry
 */
export interface ProxyHealthLog {
	/** Instance identifier */
	instanceId: string;
	/** Proxy URL that was checked */
	proxyUrl: string;
	/** Check result: healthy or unhealthy */
	status: string;
	/** Check latency in milliseconds */
	latencyMs?: number;
	/** Error message if unhealthy */
	errorMessage?: string;
	/** When the check was performed */
	checkedAt: string;
}

/**
 * Response from proxy health endpoint
 */
export interface ProxyHealthResponse {
	/** Current proxy configuration */
	proxy?: ProxyConfig;
	/** Recent health check logs */
	logs: ProxyHealthLog[];
}

/**
 * API response wrapper for proxy operations
 */
export interface ProxyResponse {
	value: boolean;
	proxy?: ProxyConfig;
}

/**
 * Helper to determine proxy health status color
 */
export function getProxyHealthColor(
	status: string,
): "green" | "red" | "yellow" {
	switch (status) {
		case "healthy":
			return "green";
		case "unhealthy":
			return "red";
		default:
			return "yellow";
	}
}

/**
 * Helper to format proxy URL for display (mask credentials).
 * Handles socks5:// and other non-standard schemes that new URL() can't parse.
 */
export function sanitizeProxyUrlForDisplay(url: string): string {
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
