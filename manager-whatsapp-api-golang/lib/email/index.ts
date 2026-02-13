/**
 * Email Module
 *
 * Central export for all email functionality including configuration,
 * templates, and sender service.
 *
 * @module lib/email
 */

// Configuration
export {
	appConfig,
	brandColors,
	emailConfig,
	getTransporter,
	verifyEmailConnection,
} from "./config";
// Sender service
export {
	type EmailResult,
	emailService,
	sendInviteAccepted,
	sendInviteExpired,
	sendLoginAlert,
	sendPasswordChanged,
	sendPasswordReset,
	sendTwoFactorCode,
	sendTwoFactorDisabled,
	sendTwoFactorEnabled,
	sendUserInvite,
} from "./sender";
// Templates
export * from "./templates";
