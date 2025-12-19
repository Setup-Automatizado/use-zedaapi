/**
 * Instance Settings Server Actions
 *
 * Server Actions for instance configuration settings including
 * call rejection, auto-read, and profile updates.
 *
 * @module actions/settings
 */

"use server";

import { revalidatePath } from "next/cache";
import { z } from "zod";
import { api } from "@/lib/api/client";
import { success, error, validationError } from "@/types";
import type { ActionResult, InstanceSettings } from "@/types";

/**
 * Instance settings validation schema
 */
const instanceSettingsSchema = z.object({
	callRejectAuto: z.boolean(),
	callRejectMessage: z.string().max(200, "Message too long").optional(),
	autoReadMessage: z.boolean(),
	notifySentByMe: z.boolean(),
});

/**
 * Profile name validation schema
 */
const profileNameSchema = z.object({
	name: z.string().min(1, "Name is required").max(100, "Name too long"),
});

/**
 * Profile description validation schema
 */
const profileDescriptionSchema = z.object({
	description: z.string().max(500, "Description too long"),
});

/**
 * Updates instance settings
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param settings - Instance configuration settings
 * @returns Action result with update confirmation
 */
export async function updateInstanceSettings(
	instanceId: string,
	instanceToken: string,
	settings: InstanceSettings,
): Promise<ActionResult<void>> {
	try {
		// Validate inputs
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const validation = instanceSettingsSchema.safeParse(settings);
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

		const {
			callRejectAuto,
			callRejectMessage,
			autoReadMessage,
			notifySentByMe,
		} = validation.data;

		// Update call rejection settings
		await api.put<{ value: boolean }>(
			"/update-call-reject-auto",
			{ value: callRejectAuto },
			{ instanceId, instanceToken },
		);

		if (callRejectMessage !== undefined) {
			await api.put<{ value: boolean }>(
				"/update-call-reject-message",
				{ value: callRejectMessage },
				{ instanceId, instanceToken },
			);
		}

		// Update auto-read messages setting
		await api.put<{ value: boolean }>(
			"/update-auto-read-message",
			{ value: autoReadMessage },
			{ instanceId, instanceToken },
		);

		// Update notify-sent-by-me setting
		await api.put<{ value: boolean }>(
			"/update-notify-sent-by-me",
			{ notifySentByMe },
			{ instanceId, instanceToken },
		);

		// Revalidate instance details
		revalidatePath(`/instances/${instanceId}`);

		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error ? err.message : "Failed to update settings";
		return error(message);
	}
}

/**
 * Updates call rejection auto setting
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param enabled - Enable/disable automatic call rejection
 * @returns Action result with update confirmation
 */
export async function updateCallRejectAuto(
	instanceId: string,
	instanceToken: string,
	enabled: boolean,
): Promise<ActionResult<void>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		await api.put<{ value: boolean }>(
			"/update-call-reject-auto",
			{ value: enabled },
			{ instanceId, instanceToken },
		);

		revalidatePath(`/instances/${instanceId}`);
		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error
				? err.message
				: "Failed to update call rejection";
		return error(message);
	}
}

/**
 * Updates call rejection message
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param message - Custom rejection message
 * @returns Action result with update confirmation
 */
export async function updateCallRejectMessage(
	instanceId: string,
	instanceToken: string,
	message: string,
): Promise<ActionResult<void>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const validation = z.string().max(200).safeParse(message);
		if (!validation.success) {
			return validationError({ message: ["Message too long"] });
		}

		await api.put<{ value: boolean }>(
			"/update-call-reject-message",
			{ value: message },
			{ instanceId, instanceToken },
		);

		revalidatePath(`/instances/${instanceId}`);
		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error
				? err.message
				: "Failed to update rejection message";
		return error(message);
	}
}

/**
 * Updates auto-read messages setting
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param enabled - Enable/disable auto-read messages
 * @returns Action result with update confirmation
 */
export async function updateAutoReadMessage(
	instanceId: string,
	instanceToken: string,
	enabled: boolean,
): Promise<ActionResult<void>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		await api.put<{ value: boolean }>(
			"/update-auto-read-message",
			{ value: enabled },
			{ instanceId, instanceToken },
		);

		revalidatePath(`/instances/${instanceId}`);
		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error
				? err.message
				: "Failed to update auto-read setting";
		return error(message);
	}
}

/**
 * Updates profile name
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param name - New profile name
 * @returns Action result with update confirmation
 */
export async function updateProfileName(
	instanceId: string,
	instanceToken: string,
	name: string,
): Promise<ActionResult<void>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const validation = profileNameSchema.safeParse({ name });
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

		await api.post<{ value: boolean }>(
			"/update-name",
			{ name },
			{ instanceId, instanceToken },
		);

		revalidatePath(`/instances/${instanceId}`);
		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error
				? err.message
				: "Failed to update profile name";
		return error(message);
	}
}

/**
 * Updates profile description
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param description - New profile description
 * @returns Action result with update confirmation
 */
export async function updateProfileDescription(
	instanceId: string,
	instanceToken: string,
	description: string,
): Promise<ActionResult<void>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		const validation = profileDescriptionSchema.safeParse({ description });
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

		await api.post<{ value: boolean }>(
			"/profile-description",
			{ description },
			{ instanceId, instanceToken },
		);

		revalidatePath(`/instances/${instanceId}`);
		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error
				? err.message
				: "Failed to update profile description";
		return error(message);
	}
}

/**
 * Updates profile picture
 *
 * @param instanceId - Instance identifier
 * @param instanceToken - Instance authentication token
 * @param imageData - Base64-encoded image data
 * @returns Action result with update confirmation
 */
export async function updateProfilePicture(
	instanceId: string,
	instanceToken: string,
	imageData: string,
): Promise<ActionResult<void>> {
	try {
		if (!instanceId || !instanceToken) {
			return error("Instance ID and token are required");
		}

		if (!imageData) {
			return error("Image data is required");
		}

		await api.post<{ value: boolean }>(
			"/profile-picture",
			{ image: imageData },
			{ instanceId, instanceToken },
		);

		revalidatePath(`/instances/${instanceId}`);
		return success(undefined);
	} catch (err) {
		const message =
			err instanceof Error
				? err.message
				: "Failed to update profile picture";
		return error(message);
	}
}
