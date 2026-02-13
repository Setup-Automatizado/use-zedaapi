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
 * Webhook field to API type mapping
 * Maps WebhookSettings fields to their corresponding API webhook types
 */
const WEBHOOK_FIELD_MAPPING: Array<{
	field: keyof Omit<WebhookSettings, "notifySentByMe">;
	type: WebhookType;
}> = [
	{ field: "deliveryCallbackUrl", type: "delivery" },
	{ field: "receivedCallbackUrl", type: "received" },
	{ field: "receivedAndDeliveryCallbackUrl", type: "received-delivery" },
	{ field: "messageStatusCallbackUrl", type: "message-status" },
	{ field: "connectedCallbackUrl", type: "connected" },
	{ field: "disconnectedCallbackUrl", type: "disconnected" },
	{ field: "presenceChatCallbackUrl", type: "chat-presence" },
];

/**
 * Result of a single webhook update attempt
 */
interface WebhookUpdateResult {
	type: WebhookType;
	field: keyof Omit<WebhookSettings, "notifySentByMe">;
	success: boolean;
	error?: string;
}

/**
 * Updates multiple webhook settings at once
 * Uses Promise.allSettled to handle partial failures gracefully.
 * Empty strings are sent to clear webhooks.
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param settings - Complete webhook configuration
 * @returns Action result with update confirmation or detailed error
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

		// Create update promises for all webhook fields
		const updatePromises = WEBHOOK_FIELD_MAPPING.map(({ field, type }) =>
			apiUpdateWebhook(
				instanceId,
				instanceToken,
				type,
				settings[field] ?? "", // undefined -> "" to clear webhook
			).then(
				() => ({ type, field, success: true }) as WebhookUpdateResult,
				(err) =>
					({
						type,
						field,
						success: false,
						error: err instanceof Error ? err.message : "Unknown error",
					}) as WebhookUpdateResult,
			),
		);

		// Wait for all updates using allSettled to handle partial failures
		const results = await Promise.all(updatePromises);

		// Check for failures
		const failures = results.filter((r) => !r.success);
		const successes = results.filter((r) => r.success);

		// Update notify setting if provided (do this even if some webhooks failed)
		if (settings.notifySentByMe !== undefined) {
			try {
				await apiUpdateNotifySentByMe(
					instanceId,
					instanceToken,
					settings.notifySentByMe,
				);
			} catch (notifyErr) {
				failures.push({
					type: "delivery", // placeholder type
					field: "deliveryCallbackUrl", // placeholder field
					success: false,
					error: `notifySentByMe: ${notifyErr instanceof Error ? notifyErr.message : "Unknown error"}`,
				});
			}
		}

		// Revalidate instance details
		revalidatePath(`/instances/${instanceId}`);

		// If all failed, return error
		if (failures.length === results.length) {
			return error(
				`Failed to update all webhooks: ${failures.map((f) => f.type).join(", ")}`,
			);
		}

		// If some failed, return partial success error
		if (failures.length > 0) {
			const failedTypes = failures.map((f) => f.type).join(", ");
			const successTypes = successes.map((s) => s.type).join(", ");
			return error(
				`Partial update: succeeded (${successTypes}), failed (${failedTypes})`,
			);
		}

		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to update webhook settings";
		return error(message);
	}
}
