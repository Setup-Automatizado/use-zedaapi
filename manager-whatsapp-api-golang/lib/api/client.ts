/**
 * Core HTTP client for WhatsApp API
 *
 * Handles authentication, request/response processing, and error handling.
 * Supports both Client-Token and Partner-Token authentication schemes.
 *
 * @module lib/api/client
 */

import "server-only";
import { ApiError } from "./errors";

/**
 * Base API URL from environment or default to localhost
 */
const API_BASE_URL = process.env.WHATSAPP_API_URL || "http://localhost:8080";

/**
 * Client authentication token (required, >=16 chars)
 * Used for instance-specific operations
 */
const CLIENT_TOKEN = process.env.WHATSAPP_CLIENT_TOKEN || "";

/**
 * Partner authentication token
 * Used for partner-level operations (list instances, create, delete)
 */
const PARTNER_TOKEN = process.env.WHATSAPP_PARTNER_TOKEN || "";

/**
 * Extended fetch options with WhatsApp API specific fields
 */
interface FetchOptions extends Omit<RequestInit, "body"> {
	/** Instance ID for instance-scoped endpoints */
	instanceId?: string;
	/** Instance token for authentication */
	instanceToken?: string;
	/** Use Partner-Token instead of Client-Token */
	usePartnerToken?: boolean;
	/** Request body (will be JSON.stringified) */
	body?: unknown;
}

/**
 * Public endpoints that don't require authentication
 */
const PUBLIC_ENDPOINTS = ["/health", "/ready", "/metrics"] as const;

/**
 * Checks if an endpoint is public (no auth required)
 */
function isPublicEndpoint(endpoint: string): boolean {
	return PUBLIC_ENDPOINTS.some((pub) => endpoint.startsWith(pub));
}

/**
 * Generic API client function
 * Handles URL construction, authentication, serialization, and error handling
 *
 * @template T - Expected response type
 * @param endpoint - API endpoint path (e.g., '/status')
 * @param options - Request options including auth and body
 * @returns Typed response data
 * @throws {ApiError} When request fails or returns non-2xx status
 */
export async function apiClient<T>(
	endpoint: string,
	options: FetchOptions = {},
): Promise<T> {
	const {
		instanceId,
		instanceToken,
		usePartnerToken,
		body,
		...fetchOptions
	} = options;

	// Build URL with optional instance path parameters
	let url: string;
	if (instanceId && instanceToken) {
		url = `${API_BASE_URL}/instances/${instanceId}/token/${instanceToken}${endpoint}`;
	} else {
		url = `${API_BASE_URL}${endpoint}`;
	}

	// Build headers with proper typing
	const headers: Record<string, string> = {
		"Content-Type": "application/json",
	};

	// Merge existing headers if provided
	if (fetchOptions.headers) {
		const existingHeaders = fetchOptions.headers;
		if (existingHeaders instanceof Headers) {
			existingHeaders.forEach((value, key) => {
				headers[key] = value;
			});
		} else if (Array.isArray(existingHeaders)) {
			existingHeaders.forEach(([key, value]) => {
				headers[key] = value;
			});
		} else {
			Object.assign(headers, existingHeaders);
		}
	}

	// Add authentication tokens
	if (usePartnerToken) {
		if (!PARTNER_TOKEN) {
			throw new Error(
				"WHATSAPP_PARTNER_TOKEN environment variable is not set",
			);
		}
		headers["Authorization"] = `Bearer ${PARTNER_TOKEN}`;
	} else if (!isPublicEndpoint(endpoint)) {
		if (!CLIENT_TOKEN) {
			throw new Error(
				"WHATSAPP_CLIENT_TOKEN environment variable is not set",
			);
		}
		if (CLIENT_TOKEN.length < 16) {
			throw new Error(
				"WHATSAPP_CLIENT_TOKEN must be at least 16 characters",
			);
		}
		headers["Client-Token"] = CLIENT_TOKEN;
	}

	// Make request
	const response = await fetch(url, {
		...fetchOptions,
		headers,
		body: body ? JSON.stringify(body) : undefined,
	});

	// Handle error responses
	if (!response.ok) {
		let errorBody: unknown = null;
		try {
			errorBody = await response.json();
		} catch {
			// Response body is not JSON, use text or null
			const text = await response.text().catch(() => null);
			errorBody = text ? { error: text } : null;
		}
		throw new ApiError(response.status, response.statusText, errorBody);
	}

	// Handle empty responses (204 No Content, etc.)
	const text = await response.text();
	if (!text) {
		return {} as T;
	}

	// Parse JSON response
	try {
		return JSON.parse(text) as T;
	} catch {
		// Response is not JSON, return as-is (for plain text responses)
		return text as unknown as T;
	}
}

/**
 * Convenience methods for common HTTP verbs
 * Provides type-safe wrappers around apiClient
 */
export const api = {
	/**
	 * GET request
	 */
	get: <T>(endpoint: string, options?: FetchOptions) =>
		apiClient<T>(endpoint, { ...options, method: "GET" }),

	/**
	 * POST request with optional body
	 */
	post: <T>(endpoint: string, body?: unknown, options?: FetchOptions) =>
		apiClient<T>(endpoint, { ...options, method: "POST", body }),

	/**
	 * PUT request with optional body
	 */
	put: <T>(endpoint: string, body?: unknown, options?: FetchOptions) =>
		apiClient<T>(endpoint, { ...options, method: "PUT", body }),

	/**
	 * DELETE request
	 */
	delete: <T>(endpoint: string, options?: FetchOptions) =>
		apiClient<T>(endpoint, { ...options, method: "DELETE" }),

	/**
	 * PATCH request with optional body
	 */
	patch: <T>(endpoint: string, body?: unknown, options?: FetchOptions) =>
		apiClient<T>(endpoint, { ...options, method: "PATCH", body }),
} as const;
