import { z } from "zod";

/**
 * Schema for creating a new WhatsApp instance.
 * Validates instance name, session configuration, device type, and initial settings.
 */
export const CreateInstanceSchema = z
	.object({
		name: z
			.string()
			.min(2, "Nome deve ter pelo menos 2 caracteres")
			.max(100, "Nome deve ter no maximo 100 caracteres")
			.trim(),

		sessionName: z
			.string()
			.trim()
			.optional()
			.transform((val) => (val === "" ? undefined : val)),

		isDevice: z.boolean().default(false),

		businessDevice: z.boolean().default(false),

		// Webhook URLs - optional, but must be valid URLs if provided
		deliveryCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		receivedCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		receivedAndDeliveryCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		messageStatusCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		connectedCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		disconnectedCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		presenceChatCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		// Instance settings
		notifySentByMe: z.boolean().default(false),

		callRejectAuto: z.boolean().default(false),

		callRejectMessage: z
			.string()
			.max(500, "Mensagem deve ter no maximo 500 caracteres")
			.trim()
			.optional()
			.or(z.literal("")),

		autoReadMessage: z.boolean().default(false),
	})
	.transform((data) => {
		// Transform empty strings to undefined for optional URL fields
		const cleaned = { ...data };
		const urlFields = [
			"deliveryCallbackUrl",
			"receivedCallbackUrl",
			"receivedAndDeliveryCallbackUrl",
			"messageStatusCallbackUrl",
			"connectedCallbackUrl",
			"disconnectedCallbackUrl",
			"presenceChatCallbackUrl",
			"callRejectMessage",
		] as const;

		urlFields.forEach((field) => {
			if (cleaned[field] === "") {
				cleaned[field] = undefined;
			}
		});

		return cleaned;
	});

/**
 * Schema for updating an existing WhatsApp instance.
 * All fields are optional to allow partial updates.
 */
export const UpdateInstanceSchema = z
	.object({
		name: z
			.string()
			.min(2, "Nome deve ter pelo menos 2 caracteres")
			.max(100, "Nome deve ter no maximo 100 caracteres")
			.trim()
			.optional(),

		sessionName: z
			.string()
			.trim()
			.optional()
			.transform((val) => (val === "" ? undefined : val)),

		isDevice: z.boolean().optional(),

		businessDevice: z.boolean().optional(),

		// Webhook URLs
		deliveryCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		receivedCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		receivedAndDeliveryCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		messageStatusCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		connectedCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		disconnectedCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		presenceChatCallbackUrl: z
			.string()
			.url("URL de webhook invalida")
			.optional()
			.or(z.literal("")),

		// Instance settings
		notifySentByMe: z.boolean().optional(),

		callRejectAuto: z.boolean().optional(),

		callRejectMessage: z
			.string()
			.max(500, "Mensagem deve ter no maximo 500 caracteres")
			.trim()
			.optional()
			.or(z.literal("")),

		autoReadMessage: z.boolean().optional(),
	})
	.transform((data) => {
		// Transform empty strings to undefined for optional URL fields
		const cleaned = { ...data };
		const urlFields = [
			"deliveryCallbackUrl",
			"receivedCallbackUrl",
			"receivedAndDeliveryCallbackUrl",
			"messageStatusCallbackUrl",
			"connectedCallbackUrl",
			"disconnectedCallbackUrl",
			"presenceChatCallbackUrl",
			"callRejectMessage",
		] as const;

		urlFields.forEach((field) => {
			if (cleaned[field] === "") {
				cleaned[field] = undefined;
			}
		});

		return cleaned;
	});

/**
 * Schema for basic instance information display.
 * Used for read-only instance data.
 */
export const InstanceInfoSchema = z.object({
	id: z.string().uuid(),
	name: z.string(),
	sessionName: z.string().optional(),
	isDevice: z.boolean(),
	businessDevice: z.boolean(),
	status: z.enum(["connected", "disconnected", "connecting", "error"]),
	phoneNumber: z.string().optional(),
	createdAt: z.string().datetime(),
	updatedAt: z.string().datetime(),
});

// Inferred TypeScript types for use throughout the application
export type CreateInstanceInput = z.infer<typeof CreateInstanceSchema>;
export type UpdateInstanceInput = z.infer<typeof UpdateInstanceSchema>;
export type InstanceInfo = z.infer<typeof InstanceInfoSchema>;
