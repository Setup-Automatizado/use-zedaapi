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
  createInstance,
  deleteInstance,
  restartInstance,
  disconnectInstance,
  activateSubscription,
  cancelSubscription,
} from './instances';

// Webhook configuration actions
export {
  updateWebhook,
  updateAllWebhooks,
  updateNotifySentByMe,
  updateWebhookSettings,
} from './webhooks';

// Instance settings actions
export {
  updateInstanceSettings,
  updateCallRejectAuto,
  updateCallRejectMessage,
  updateAutoReadMessage,
  updateProfileName,
  updateProfileDescription,
  updateProfilePicture,
} from './settings';
