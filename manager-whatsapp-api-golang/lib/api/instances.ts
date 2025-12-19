/**
 * Instance management API functions
 *
 * Provides type-safe wrappers for WhatsApp instance operations including:
 * - Partner-level operations (list, create, delete)
 * - Instance-level operations (status, QR code, pairing, control)
 *
 * @module lib/api/instances
 */

import "server-only";
import { api } from "./client";
import type {
	Instance,
	InstanceListResponse,
	RawInstanceListResponse,
	InstanceStatus,
	DeviceInfo,
	CreateInstanceRequest,
	CreateInstanceResponse,
	QRCodeResponse,
	PhonePairingResponse,
} from "@/types";
import { normalizeInstances } from "@/types";

// ============================================================================
// Partner-level endpoints (require Partner-Token)
// ============================================================================

/**
 * Lists all instances for the authenticated partner (raw API response)
 *
 * @param page - Page number (1-indexed)
 * @param pageSize - Number of items per page
 * @param query - Optional search query
 * @returns Paginated list of raw instances from API
 */
async function listInstancesRaw(
	page = 1,
	pageSize = 20,
	query?: string,
): Promise<RawInstanceListResponse> {
	const params = new URLSearchParams({
		page: String(page),
		pageSize: String(pageSize),
		...(query && { query }),
	});

	return api.get<RawInstanceListResponse>(`/instances?${params.toString()}`, {
		usePartnerToken: true,
	});
}

/**
 * Lists all instances for the authenticated partner
 * Returns normalized instances with UI-friendly aliases (id, token, created)
 *
 * @param page - Page number (1-indexed)
 * @param pageSize - Number of items per page
 * @param query - Optional search query
 * @returns Paginated list of normalized instances
 */
export async function listInstances(
	page = 1,
	pageSize = 20,
	query?: string,
): Promise<InstanceListResponse> {
	const response = await listInstancesRaw(page, pageSize, query);
	return {
		...response,
		content: normalizeInstances(response.content),
	};
}

/** Raw response from create instance API */
interface RawCreateInstanceResponse {
	readonly instanceId: string;
	readonly instanceToken: string;
	due?: string;
	name: string;
	sessionName: string;
}

/**
 * Creates a new WhatsApp instance
 *
 * @param data - Instance creation parameters
 * @returns Created instance details with credentials (normalized with aliases)
 */
export async function createInstance(
	data: CreateInstanceRequest,
): Promise<CreateInstanceResponse> {
	const response = await api.post<RawCreateInstanceResponse>(
		"/instances/integrator/on-demand",
		data,
		{ usePartnerToken: true },
	);
	// Add aliases for UI compatibility
	return {
		...response,
		id: response.instanceId,
		token: response.instanceToken,
	};
}

/**
 * Deletes an instance permanently
 *
 * @param instanceId - Instance ID to delete
 * @returns Status confirmation
 */
export async function deleteInstance(
	instanceId: string,
): Promise<{ status: string }> {
	return api.delete<{ status: string }>(`/instances/${instanceId}`, {
		usePartnerToken: true,
	});
}

/**
 * Gets a single instance by ID
 * Since there's no dedicated endpoint, we fetch from the list
 *
 * @param instanceId - Instance ID to fetch
 * @returns Normalized instance details or null if not found
 */
export async function getInstance(
	instanceId: string,
): Promise<Instance | null> {
	// Fetch all instances and find the one with matching ID
	// listInstances already returns normalized instances
	const response = await listInstances(1, 100);
	const instance = response.content.find(
		(inst) => inst.instanceId === instanceId,
	);
	return instance || null;
}

// ============================================================================
// Instance-level endpoints (require Client-Token + instance credentials)
// ============================================================================

/**
 * Gets current connection status and metadata for an instance
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns Connection status and device information
 */
export async function getInstanceStatus(
	instanceId: string,
	instanceToken: string,
): Promise<InstanceStatus> {
	return api.get<InstanceStatus>("/status", { instanceId, instanceToken });
}

/**
 * Gets QR code for WhatsApp pairing (plain text format)
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns QR code string
 */
export async function getQRCode(
	instanceId: string,
	instanceToken: string,
): Promise<string> {
	return api.get<string>("/qr-code", { instanceId, instanceToken });
}

/**
 * Gets QR code as base64-encoded image
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns QR code image data
 */
export async function getQRCodeImage(
	instanceId: string,
	instanceToken: string,
): Promise<QRCodeResponse> {
	return api.get<QRCodeResponse>("/qr-code/image", {
		instanceId,
		instanceToken,
	});
}

/**
 * Gets phone pairing code for linking device
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param phone - Phone number (E.164 format recommended)
 * @returns Pairing code
 */
export async function getPhonePairingCode(
	instanceId: string,
	instanceToken: string,
	phone: string,
): Promise<PhonePairingResponse> {
	return api.get<PhonePairingResponse>(`/phone-code/${phone}`, {
		instanceId,
		instanceToken,
	});
}

/**
 * Gets device information for connected instance
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns Device details
 */
export async function getDeviceInfo(
	instanceId: string,
	instanceToken: string,
): Promise<DeviceInfo> {
	return api.get<DeviceInfo>("/device", { instanceId, instanceToken });
}

/**
 * Restarts WhatsApp connection for instance
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns Success status
 */
export async function restartInstance(
	instanceId: string,
	instanceToken: string,
): Promise<{ value: boolean }> {
	return api.post<{ value: boolean }>("/restart", undefined, {
		instanceId,
		instanceToken,
	});
}

/**
 * Disconnects WhatsApp connection (logs out)
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns Success status
 */
export async function disconnectInstance(
	instanceId: string,
	instanceToken: string,
): Promise<{ value: boolean }> {
	return api.post<{ value: boolean }>("/disconnect", undefined, {
		instanceId,
		instanceToken,
	});
}

// ============================================================================
// Subscription management (hybrid - requires both tokens)
// ============================================================================

/**
 * Activates subscription for an instance
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns Subscription status
 */
export async function activateSubscription(
	instanceId: string,
	instanceToken: string,
): Promise<{ status: string }> {
	return api.post<{ status: string }>(
		"/integrator/on-demand/subscription",
		undefined,
		{
			instanceId,
			instanceToken,
			usePartnerToken: true,
		},
	);
}

/**
 * Cancels subscription for an instance
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns Cancellation status
 */
export async function cancelSubscription(
	instanceId: string,
	instanceToken: string,
): Promise<{ status: string }> {
	return api.post<{ status: string }>(
		"/integrator/on-demand/cancel",
		undefined,
		{
			instanceId,
			instanceToken,
			usePartnerToken: true,
		},
	);
}
