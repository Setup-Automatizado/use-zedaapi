/**
 * Barrel export for all Zod validation schemas.
 * This file aggregates schemas from instance, webhook, and settings modules.
 *
 * Usage:
 * import { CreateInstanceSchema, WebhookConfigSchema } from '@/schemas';
 */

// Instance schemas
export {
	type CreateInstanceInput,
	CreateInstanceSchema,
	type InstanceInfo,
	InstanceInfoSchema,
	type UpdateInstanceInput,
	UpdateInstanceSchema,
} from "./instance";
// Settings schemas
export {
	type AllSettings,
	AllSettingsSchema,
	type InstanceSettings,
	type InstanceSettingsInput,
	InstanceSettingsSchema,
	type MessageSettings,
	MessageSettingsSchema,
	type PrivacySettings,
	PrivacySettingsSchema,
	type ProfileSettings,
	ProfileSettingsSchema,
	type UpdateSettings,
	UpdateSettingsSchema,
} from "./settings";
// Webhook schemas
export {
	type BulkWebhookUpdate,
	BulkWebhookUpdateSchema,
	type SingleWebhookUpdate,
	SingleWebhookUpdateSchema,
	type WebhookConfig,
	type WebhookConfigInput,
	WebhookConfigSchema,
	type WebhookEvent,
	WebhookEventSchema,
	type WebhookTest,
	WebhookTestSchema,
	WebhookUrlSchema,
} from "./webhook";
