/**
 * Server Actions Barrel Export
 *
 * Centralized export for all Server Actions used in the WhatsApp Instance Manager.
 * Import from this barrel file for consistent access across the application.
 *
 * @example
 * ```typescript
 * import { createInstance, updateWebhook } from '@/actions';
 * ```
 */

// Instance management actions
export {
	activateSubscription,
	cancelSubscription,
	createInstance,
	deleteInstance,
	disconnectInstance,
	restartInstance,
} from "./instances";
// Instance settings actions
export {
	updateAutoReadMessage,
	updateCallRejectAuto,
	updateCallRejectMessage,
	updateInstanceSettings,
	updateProfileDescription,
	updateProfileName,
	updateProfilePicture,
} from "./settings";
// Webhook configuration actions
export {
	updateAllWebhooks,
	updateNotifySentByMe,
	updateWebhook,
	updateWebhookSettings,
} from "./webhooks";
