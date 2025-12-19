/**
 * Webhook management API functions
 *
 * Provides type-safe wrappers for webhook configuration operations including:
 * - Individual webhook URL updates
 * - Bulk webhook configuration
 * - Notification preferences
 *
 * @module lib/api/webhooks
 */

import "server-only";
import { api } from "./client";
import type {
	WebhookType,
	WebhookUpdateRequest,
	WebhookUpdateResponse,
	NotifySentByMeRequest,
	AllWebhooksUpdateRequest,
} from "@/types";

/**
 * Webhook endpoint mapping
 * Maps webhook types to their corresponding API endpoints
 */
const WEBHOOK_ENDPOINTS: Record<WebhookType, string> = {
	delivery: "/update-webhook-delivery",
	received: "/update-webhook-received",
	"received-delivery": "/update-webhook-received-delivery",
	"message-status": "/update-webhook-message-status",
	connected: "/update-webhook-connected",
	disconnected: "/update-webhook-disconnected",
	"chat-presence": "/update-webhook-chat-presence",
} as const;

/**
 * Updates a single webhook URL
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param type - Webhook type to update
 * @param url - New webhook URL (empty string to disable)
 * @returns Update confirmation with complete webhook settings
 */
export async function updateWebhook(
	instanceId: string,
	instanceToken: string,
	type: WebhookType,
	url: string,
): Promise<WebhookUpdateResponse> {
	const endpoint = WEBHOOK_ENDPOINTS[type];
	if (!endpoint) {
		throw new Error(`Invalid webhook type: ${type}`);
	}

	const body: WebhookUpdateRequest = { value: url };

	return api.post<WebhookUpdateResponse>(endpoint, body, {
		instanceId,
		instanceToken,
	});
}

/**
 * Updates all webhook URLs to the same value
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param settings - Webhook configuration (URL and optional notify setting)
 * @returns Update confirmation with complete webhook settings
 */
export async function updateAllWebhooks(
	instanceId: string,
	instanceToken: string,
	settings: AllWebhooksUpdateRequest,
): Promise<WebhookUpdateResponse> {
	return api.post<WebhookUpdateResponse>("/update-every-webhooks", settings, {
		instanceId,
		instanceToken,
	});
}

/**
 * Updates the notify-sent-by-me setting
 * Controls whether webhooks are triggered for messages sent by the instance itself
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param enabled - Enable/disable notifications for own messages
 * @returns Update confirmation
 */
export async function updateNotifySentByMe(
	instanceId: string,
	instanceToken: string,
	enabled: boolean,
): Promise<{ value: boolean }> {
	const body: NotifySentByMeRequest = { notifySentByMe: enabled };

	return api.post<{ value: boolean }>("/update-notify-sent-by-me", body, {
		instanceId,
		instanceToken,
	});
}
