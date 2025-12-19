/**
 * WhatsApp API client library
 *
 * Central export point for all API functions, types, and utilities.
 * Import from this module to access all API functionality.
 *
 * @example
 * ```typescript
 * import { listInstances, createInstance, ApiError } from '@/lib/api';
 *
 * try {
 *   const instances = await listInstances();
 *   console.log(instances);
 * } catch (error) {
 *   if (error instanceof ApiError) {
 *     console.error(`API error: ${error.message}`);
 *   }
 * }
 * ```
 *
 * @module lib/api
 */

import "server-only";

// Core client and error handling
export { api, apiClient } from "./client";
export { ApiError, isApiError } from "./errors";

// Instance management
export {
	listInstances,
	createInstance,
	deleteInstance,
	getInstanceStatus,
	getQRCode,
	getQRCodeImage,
	getPhonePairingCode,
	getDeviceInfo,
	restartInstance,
	disconnectInstance,
	activateSubscription,
	cancelSubscription,
} from "./instances";

// Health and monitoring
export { getHealth, getReadiness, getMetrics } from "./health";

// Webhook configuration
export {
	updateWebhook,
	updateAllWebhooks,
	updateNotifySentByMe,
} from "./webhooks";

// Re-export types for convenience
export type {
	Instance,
	InstanceListResponse,
	InstanceStatus,
	DeviceInfo,
	CreateInstanceRequest,
	CreateInstanceResponse,
	QRCodeResponse,
	PhonePairingResponse,
	HealthResponse,
	ReadinessResponse,
	WebhookType,
	WebhookUpdateResponse,
} from "@/types";
