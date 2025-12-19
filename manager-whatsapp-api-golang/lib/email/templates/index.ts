/**
 * Email Templates Index
 *
 * Exports all email templates for easy importing throughout the application.
 *
 * @module lib/email/templates
 */

// Base template and helpers
export {
  baseTemplate,
  primaryButton,
  secondaryButton,
  infoBox,
  divider,
  paragraph,
  heading,
  type BaseTemplateProps,
} from './base';

// Login alert template
export {
  loginAlertTemplate,
  type LoginAlertData,
} from './login-alert';

// Password reset templates
export {
  passwordResetTemplate,
  passwordChangedTemplate,
  type PasswordResetData,
  type PasswordChangedData,
} from './password-reset';

// Two-factor authentication templates
export {
  twoFactorCodeTemplate,
  twoFactorEnabledTemplate,
  twoFactorDisabledTemplate,
  type TwoFactorCodeData,
  type TwoFactorEnabledData,
  type TwoFactorDisabledData,
} from './two-factor';

// User invite templates
export {
  userInviteTemplate,
  inviteAcceptedTemplate,
  inviteExpiredTemplate,
  type UserInviteData,
  type InviteAcceptedData,
  type InviteExpiredData,
} from './user-invite';
