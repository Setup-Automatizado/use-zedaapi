/**
 * Instance Management Server Actions
 *
 * Server Actions for WhatsApp instance CRUD operations.
 * Includes form validation, API integration, and cache revalidation.
 *
 * @module actions/instances
 */

"use server";

import { revalidatePath } from "next/cache";
import { z } from "zod";
import {
	activateSubscription as apiActivateSubscription,
	cancelSubscription as apiCancelSubscription,
	createInstance as apiCreateInstance,
	deleteInstance as apiDeleteInstance,
	disconnectInstance as apiDisconnectInstance,
	restartInstance as apiRestartInstance,
} from "@/lib/api/instances";
import type { ActionResult, CreateInstanceResponse } from "@/types";
import { error, success, validationError } from "@/types";

/**
 * Validation schema for instance creation
 */
const createInstanceSchema = z.object({
	name: z
		.string()
		.min(1, "Instance name is required")
		.max(100, "Name too long"),
	sessionName: z.string().optional(),

	// Webhook URLs
	deliveryCallbackUrl: z
		.string()
		.url("Invalid URL")
		.optional()
		.or(z.literal("")),
	receivedCallbackUrl: z
		.string()
		.url("Invalid URL")
		.optional()
		.or(z.literal("")),
	receivedAndDeliveryCallbackUrl: z
		.string()
		.url("Invalid URL")
		.optional()
		.or(z.literal("")),
	messageStatusCallbackUrl: z
		.string()
		.url("Invalid URL")
		.optional()
		.or(z.literal("")),
	connectedCallbackUrl: z
		.string()
		.url("Invalid URL")
		.optional()
		.or(z.literal("")),
	disconnectedCallbackUrl: z
		.string()
		.url("Invalid URL")
		.optional()
		.or(z.literal("")),
	presenceChatCallbackUrl: z
		.string()
		.url("Invalid URL")
		.optional()
		.or(z.literal("")),

	// Settings
	notifySentByMe: z.boolean().optional().default(false),
	callRejectAuto: z.boolean().optional().default(false),
	callRejectMessage: z.string().max(200, "Message too long").optional(),
	autoReadMessage: z.boolean().optional().default(false),

	// Device configuration
	isDevice: z.boolean().optional().default(false),
	businessDevice: z.boolean().optional().default(false),
});

/**
 * Creates a new WhatsApp instance
 *
 * @param formData - Form data containing instance configuration
 * @returns Action result with created instance details
 */
export async function createInstance(
	formData: FormData,
): Promise<ActionResult<CreateInstanceResponse>> {
	try {
		// Extract and parse form data
		const rawData = {
			name: formData.get("name"),
			sessionName: formData.get("sessionName") || undefined,
			deliveryCallbackUrl: formData.get("deliveryCallbackUrl") || undefined,
			receivedCallbackUrl: formData.get("receivedCallbackUrl") || undefined,
			receivedAndDeliveryCallbackUrl:
				formData.get("receivedAndDeliveryCallbackUrl") || undefined,
			messageStatusCallbackUrl:
				formData.get("messageStatusCallbackUrl") || undefined,
			connectedCallbackUrl: formData.get("connectedCallbackUrl") || undefined,
			disconnectedCallbackUrl:
				formData.get("disconnectedCallbackUrl") || undefined,
			presenceChatCallbackUrl:
				formData.get("presenceChatCallbackUrl") || undefined,
			notifySentByMe: formData.get("notifySentByMe") === "true",
			callRejectAuto: formData.get("callRejectAuto") === "true",
			callRejectMessage: formData.get("callRejectMessage") || undefined,
			autoReadMessage: formData.get("autoReadMessage") === "true",
			isDevice: formData.get("isDevice") === "true",
			businessDevice: formData.get("businessDevice") === "true",
		};

		// Validate input
		const validation = createInstanceSchema.safeParse(rawData);

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
		const instance = await apiCreateInstance(validation.data);

		// Revalidate instances list
		revalidatePath("/instances");

		return success(instance);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to create instance";
		return error(message);
	}
}

/**
 * Deletes a WhatsApp instance permanently
 *
 * @param instanceId - Instance ID to delete
 * @returns Action result with deletion confirmation
 */
export async function deleteInstance(
	instanceId: string,
): Promise<ActionResult<void>> {
	try {
		// Validate input
		if (!instanceId || typeof instanceId !== "string") {
			return error("Invalid instance ID");
		}

		// Call API
		await apiDeleteInstance(instanceId);

		// Revalidate instances list
		revalidatePath("/instances");

		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to delete instance";
		return error(message);
	}
}

/**
 * Restarts WhatsApp connection for an instance
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns Action result with restart confirmation
 */
export async function restartInstance(
	instanceId: string,
	instanceToken: string,
): Promise<ActionResult<void>> {
	try {
		// Validate inputs
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		// Call API
		await apiRestartInstance(instanceId, instanceToken);

		// Revalidate instance details
		revalidatePath(`/instances/${instanceId}`);
		revalidatePath("/instances");

		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to restart instance";
		return error(message);
	}
}

/**
 * Disconnects WhatsApp connection (logs out)
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns Action result with disconnect confirmation
 */
export async function disconnectInstance(
	instanceId: string,
	instanceToken: string,
): Promise<ActionResult<void>> {
	try {
		// Validate inputs
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		// Call API
		await apiDisconnectInstance(instanceId, instanceToken);

		// Revalidate instance details
		revalidatePath(`/instances/${instanceId}`);
		revalidatePath("/instances");

		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to disconnect instance";
		return error(message);
	}
}

/**
 * Activates subscription for an instance
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns Action result with activation confirmation
 */
export async function activateSubscription(
	instanceId: string,
	instanceToken: string,
): Promise<ActionResult<void>> {
	try {
		// Validate inputs
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		// Call API
		await apiActivateSubscription(instanceId, instanceToken);

		// Revalidate instance details
		revalidatePath(`/instances/${instanceId}`);
		revalidatePath("/instances");

		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to activate subscription";
		return error(message);
	}
}

/**
 * Cancels subscription for an instance
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @returns Action result with cancellation confirmation
 */
export async function cancelSubscription(
	instanceId: string,
	instanceToken: string,
): Promise<ActionResult<void>> {
	try {
		// Validate inputs
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		// Call API
		await apiCancelSubscription(instanceId, instanceToken);

		// Revalidate instance details
		revalidatePath(`/instances/${instanceId}`);
		revalidatePath("/instances");

		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to cancel subscription";
		return error(message);
	}
}
