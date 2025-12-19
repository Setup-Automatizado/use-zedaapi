/**
 * Barrel export for all Zod validation schemas.
 * This file aggregates schemas from instance, webhook, and settings modules.
 *
 * Usage:
 * import { CreateInstanceSchema, WebhookConfigSchema } from '@/schemas';
 */

// Instance schemas
export {
	CreateInstanceSchema,
	UpdateInstanceSchema,
	InstanceInfoSchema,
	type CreateInstanceInput,
	type UpdateInstanceInput,
	type InstanceInfo,
} from "./instance";

// Webhook schemas
export {
	WebhookUrlSchema,
	WebhookConfigSchema,
	SingleWebhookUpdateSchema,
	WebhookTestSchema,
	WebhookEventSchema,
	BulkWebhookUpdateSchema,
	type WebhookConfig,
	type WebhookConfigInput,
	type SingleWebhookUpdate,
	type WebhookTest,
	type WebhookEvent,
	type BulkWebhookUpdate,
} from "./webhook";

// Settings schemas
export {
	InstanceSettingsSchema,
	UpdateSettingsSchema,
	ProfileSettingsSchema,
	PrivacySettingsSchema,
	MessageSettingsSchema,
	AllSettingsSchema,
	type InstanceSettings,
	type InstanceSettingsInput,
	type UpdateSettings,
	type ProfileSettings,
	type PrivacySettings,
	type MessageSettings,
	type AllSettings,
} from "./settings";
