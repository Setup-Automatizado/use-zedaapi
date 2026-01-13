import { z } from "zod";

/**
 * Schema for validating a single webhook URL.
 * Accepts valid URLs or empty strings (empty string means clear webhook).
 */
export const WebhookUrlSchema = z
	.string()
	.url("URL de webhook invalida")
	.or(z.literal(""));

/**
 * Schema for complete webhook configuration.
 * Includes all 7 webhook types supported by the WhatsApp API.
 * Empty strings are preserved (not transformed to undefined) to allow clearing webhooks.
 */
export const WebhookConfigSchema = z.object({
	deliveryCallbackUrl: z
		.string()
		.url("URL de entrega invalida")
		.or(z.literal(""))
		.default(""),

	receivedCallbackUrl: z
		.string()
		.url("URL de recebimento invalida")
		.or(z.literal(""))
		.default(""),

	receivedAndDeliveryCallbackUrl: z
		.string()
		.url("URL de recebimento e entrega invalida")
		.or(z.literal(""))
		.default(""),

	messageStatusCallbackUrl: z
		.string()
		.url("URL de status de mensagem invalida")
		.or(z.literal(""))
		.default(""),

	connectedCallbackUrl: z
		.string()
		.url("URL de conexao invalida")
		.or(z.literal(""))
		.default(""),

	disconnectedCallbackUrl: z
		.string()
		.url("URL de desconexao invalida")
		.or(z.literal(""))
		.default(""),

	presenceChatCallbackUrl: z
		.string()
		.url("URL de presenca invalida")
		.or(z.literal(""))
		.default(""),

	notifySentByMe: z.boolean().default(false),
});

/**
 * Schema for updating a single webhook URL.
 * Used when updating one webhook at a time.
 * Empty string means clear the webhook.
 */
export const SingleWebhookUpdateSchema = z.object({
	webhookType: z.enum([
		"delivery",
		"received",
		"receivedAndDelivery",
		"messageStatus",
		"connected",
		"disconnected",
		"presenceChat",
	]),

	callbackUrl: z.string().url("URL de webhook invalida").or(z.literal("")),
});

/**
 * Schema for webhook test payload.
 * Used when sending test webhook events.
 */
export const WebhookTestSchema = z.object({
	webhookType: z.enum([
		"delivery",
		"received",
		"receivedAndDelivery",
		"messageStatus",
		"connected",
		"disconnected",
		"presenceChat",
	]),

	instanceId: z.string().uuid("ID de instancia invalido"),

	testPayload: z.record(z.string(), z.unknown()).optional(),
});

/**
 * Schema for webhook event history entry.
 * Used for displaying webhook delivery logs.
 */
export const WebhookEventSchema = z.object({
	id: z.string().uuid(),
	instanceId: z.string().uuid(),
	webhookType: z.string(),
	callbackUrl: z.string().url(),
	payload: z.record(z.string(), z.unknown()),
	statusCode: z.number().int().min(100).max(599),
	attempts: z.number().int().min(0),
	success: z.boolean(),
	error: z.string().optional(),
	createdAt: z.string().datetime(),
	deliveredAt: z.string().datetime().optional(),
});

/**
 * Schema for bulk webhook update.
 * Allows updating multiple webhook URLs at once.
 * Empty strings are preserved to allow clearing webhooks.
 */
export const BulkWebhookUpdateSchema = z
	.object({
		deliveryCallbackUrl: z
			.string()
			.url("URL de entrega invalida")
			.or(z.literal("")),

		receivedCallbackUrl: z
			.string()
			.url("URL de recebimento invalida")
			.or(z.literal("")),

		receivedAndDeliveryCallbackUrl: z
			.string()
			.url("URL de recebimento e entrega invalida")
			.or(z.literal("")),

		messageStatusCallbackUrl: z
			.string()
			.url("URL de status de mensagem invalida")
			.or(z.literal("")),

		connectedCallbackUrl: z
			.string()
			.url("URL de conexao invalida")
			.or(z.literal("")),

		disconnectedCallbackUrl: z
			.string()
			.url("URL de desconexao invalida")
			.or(z.literal("")),

		presenceChatCallbackUrl: z
			.string()
			.url("URL de presenca no chat invalida")
			.or(z.literal("")),

		notifySentByMe: z.boolean().default(false),
	})
	.partial();

// Inferred TypeScript types
export type WebhookConfig = z.infer<typeof WebhookConfigSchema>;
export type WebhookConfigInput = z.input<typeof WebhookConfigSchema>;
export type SingleWebhookUpdate = z.infer<typeof SingleWebhookUpdateSchema>;
export type WebhookTest = z.infer<typeof WebhookTestSchema>;
export type WebhookEvent = z.infer<typeof WebhookEventSchema>;
export type BulkWebhookUpdate = z.infer<typeof BulkWebhookUpdateSchema>;
