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
  emailConfig,
  appConfig,
  brandColors,
  getTransporter,
  verifyEmailConnection,
} from './config';

// Templates
export * from './templates';

// Sender service
export {
  emailService,
  sendLoginAlert,
  sendPasswordReset,
  sendPasswordChanged,
  sendTwoFactorCode,
  sendTwoFactorEnabled,
  sendTwoFactorDisabled,
  sendUserInvite,
  sendInviteAccepted,
  sendInviteExpired,
  type EmailResult,
} from './sender';
