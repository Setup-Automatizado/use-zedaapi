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

// Re-export types for convenience
export type {
	CreateInstanceRequest,
	CreateInstanceResponse,
	DeviceInfo,
	HealthResponse,
	Instance,
	InstanceListResponse,
	InstanceStatus,
	PhonePairingResponse,
	QRCodeResponse,
	ReadinessResponse,
	WebhookType,
	WebhookUpdateResponse,
} from "@/types";
// Core client and error handling
export { api, apiClient } from "./client";
export { ApiError, isApiError } from "./errors";

// Health and monitoring
export { getHealth, getMetrics, getReadiness } from "./health";
// Instance management
export {
	activateSubscription,
	cancelSubscription,
	createInstance,
	deleteInstance,
	disconnectInstance,
	getDeviceInfo,
	getInstanceStatus,
	getPhonePairingCode,
	getQRCode,
	getQRCodeImage,
	listInstances,
	restartInstance,
} from "./instances";
// Webhook configuration
export {
	updateAllWebhooks,
	updateNotifySentByMe,
	updateWebhook,
} from "./webhooks";
