/**
 * Webhook Configuration Server Actions
 *
 * Server Actions for webhook URL configuration and notification settings.
 * Includes validation, API integration, and cache revalidation.
 *
 * @module actions/webhooks
 */

"use server";

import { revalidatePath } from "next/cache";
import { z } from "zod";
import {
	updateAllWebhooks as apiUpdateAllWebhooks,
	updateNotifySentByMe as apiUpdateNotifySentByMe,
	updateWebhook as apiUpdateWebhook,
} from "@/lib/api/webhooks";
import type {
	ActionResult,
	WebhookSettings,
	WebhookType,
	WebhookUpdateResponse,
} from "@/types";
import { error, success, validationError } from "@/types";

/**
 * Webhook URL validation schema
 */
const webhookUrlSchema = z
	.string()
	.url("Invalid webhook URL")
	.or(z.literal(""))
	.optional();

/**
 * Webhook type validation schema
 */
const webhookTypeSchema = z.enum([
	"delivery",
	"received",
	"received-delivery",
	"message-status",
	"connected",
	"disconnected",
	"chat-presence",
]);

/**
 * All webhooks update schema
 */
const allWebhooksSchema = z.object({
	value: webhookUrlSchema,
	notifySentByMe: z.boolean().optional(),
});

/**
 * Updates a single webhook URL
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param type - Webhook type to update
 * @param url - New webhook URL (empty string to disable)
 * @returns Action result with updated webhook configuration
 */
export async function updateWebhook(
	instanceId: string,
	instanceToken: string,
	type: WebhookType,
	url: string,
): Promise<ActionResult<WebhookUpdateResponse>> {
	try {
		// Validate inputs
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const typeValidation = webhookTypeSchema.safeParse(type);
		if (!typeValidation.success) {
			return error("Invalid webhook type");
		}

		const urlValidation = webhookUrlSchema.safeParse(url);
		if (!urlValidation.success) {
			return validationError({ url: ["Invalid webhook URL"] });
		}

		// Call API
		const result = await apiUpdateWebhook(
			instanceId,
			instanceToken,
			typeValidation.data,
			url || "",
		);

		// Revalidate instance details
		revalidatePath(`/instances/${instanceId}`);

		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to update webhook";
		return error(message);
	}
}

/**
 * Updates all webhook URLs to the same value
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param settings - Webhook configuration (URL and optional notify setting)
 * @returns Action result with updated webhook configuration
 */
export async function updateAllWebhooks(
	instanceId: string,
	instanceToken: string,
	settings: { value: string; notifySentByMe?: boolean },
): Promise<ActionResult<WebhookUpdateResponse>> {
	try {
		// Validate inputs
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const validation = allWebhooksSchema.safeParse(settings);
		if (!validation.success) {
			const errors: Record<string, string[]> = {};
			validation.error.issues.forEach((issue) => {
				const path = issue.path[0]?.toString() || "form";
				if (!errors[path]) {
					errors[path] = [];
				}
				errors[path].push(issue.message);
			});
			return validationError(errors);
		}

		// Call API
		const result = await apiUpdateAllWebhooks(instanceId, instanceToken, {
			value: validation.data.value || "",
			notifySentByMe: validation.data.notifySentByMe,
		});

		// Revalidate instance details
		revalidatePath(`/instances/${instanceId}`);

		return success(result);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to update webhooks";
		return error(message);
	}
}

/**
 * Updates the notify-sent-by-me setting
 * Controls whether webhooks are triggered for messages sent by the instance itself
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param enabled - Enable/disable notifications for own messages
 * @returns Action result with update confirmation
 */
export async function updateNotifySentByMe(
	instanceId: string,
	instanceToken: string,
	enabled: boolean,
): Promise<ActionResult<void>> {
	try {
		// Validate inputs
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		if (typeof enabled !== "boolean") {
			return error("Invalid enabled value");
		}

		// Call API
		await apiUpdateNotifySentByMe(instanceId, instanceToken, enabled);

		// Revalidate instance details
		revalidatePath(`/instances/${instanceId}`);

		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error
				? err.message
				: "Failed to update notification setting";
		return error(message);
	}
}

/**
 * Updates multiple webhook settings at once
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param settings - Complete webhook configuration
 * @returns Action result with update confirmation
 */
export async function updateWebhookSettings(
	instanceId: string,
	instanceToken: string,
	settings: Partial<WebhookSettings>,
): Promise<ActionResult<void>> {
	try {
		// Validate inputs
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		// Update each webhook individually
		const updates: Promise<WebhookUpdateResponse>[] = [];

		if (settings.deliveryCallbackUrl !== undefined) {
			updates.push(
				apiUpdateWebhook(
					instanceId,
					instanceToken,
					"delivery",
					settings.deliveryCallbackUrl,
				),
			);
		}

		if (settings.receivedCallbackUrl !== undefined) {
			updates.push(
				apiUpdateWebhook(
					instanceId,
					instanceToken,
					"received",
					settings.receivedCallbackUrl,
				),
			);
		}

		if (settings.receivedAndDeliveryCallbackUrl !== undefined) {
			updates.push(
				apiUpdateWebhook(
					instanceId,
					instanceToken,
					"received-delivery",
					settings.receivedAndDeliveryCallbackUrl,
				),
			);
		}

		if (settings.messageStatusCallbackUrl !== undefined) {
			updates.push(
				apiUpdateWebhook(
					instanceId,
					instanceToken,
					"message-status",
					settings.messageStatusCallbackUrl,
				),
			);
		}

		if (settings.connectedCallbackUrl !== undefined) {
			updates.push(
				apiUpdateWebhook(
					instanceId,
					instanceToken,
					"connected",
					settings.connectedCallbackUrl,
				),
			);
		}

		if (settings.disconnectedCallbackUrl !== undefined) {
			updates.push(
				apiUpdateWebhook(
					instanceId,
					instanceToken,
					"disconnected",
					settings.disconnectedCallbackUrl,
				),
			);
		}

		if (settings.presenceChatCallbackUrl !== undefined) {
			updates.push(
				apiUpdateWebhook(
					instanceId,
					instanceToken,
					"chat-presence",
					settings.presenceChatCallbackUrl,
				),
			);
		}

		// Wait for all updates to complete
		await Promise.all(updates);

		// Update notify setting if provided
		if (settings.notifySentByMe !== undefined) {
			await apiUpdateNotifySentByMe(
				instanceId,
				instanceToken,
				settings.notifySentByMe,
			);
		}

		// Revalidate instance details
		revalidatePath(`/instances/${instanceId}`);

		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to update webhook settings";
		return error(message);
	}
}
