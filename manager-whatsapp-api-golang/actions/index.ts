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
// Proxy configuration actions
export {
	fetchProxyConfig,
	fetchProxyHealth,
	removeProxyConfig,
	swapProxyConnection,
	testProxyConnection,
	updateProxyConfig,
} from "./proxy";
// Webhook configuration actions
export {
	updateAllWebhooks,
	updateNotifySentByMe,
	updateWebhook,
	updateWebhookSettings,
} from "./webhooks";
// Pool management actions
export {
	assignInstancePoolProxy,
	assignInstanceToGroup,
	bulkAssignPoolProxies,
	createPoolGroup,
	createPoolProvider,
	deletePoolGroup,
	deletePoolProvider,
	fetchInstancePoolAssignment,
	fetchPoolGroups,
	fetchPoolProviders,
	fetchPoolProxies,
	fetchPoolStats,
	releaseInstancePoolProxy,
	syncPoolProvider,
	updatePoolProvider,
} from "./pool";
