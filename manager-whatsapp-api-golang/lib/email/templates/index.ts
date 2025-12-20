/**
 * Email Templates Index
 *
 * Exports all email templates for easy importing throughout the application.
 *
 * @module lib/email/templates
 */

// Base template and helpers
export {
	type BaseTemplateProps,
	baseTemplate,
	divider,
	heading,
	infoBox,
	paragraph,
	primaryButton,
	secondaryButton,
} from "./base";

// Login alert template
export {
	type LoginAlertData,
	loginAlertTemplate,
} from "./login-alert";

// Password reset templates
export {
	type PasswordChangedData,
	type PasswordResetData,
	passwordChangedTemplate,
	passwordResetTemplate,
} from "./password-reset";

// Two-factor authentication templates
export {
	type TwoFactorCodeData,
	type TwoFactorDisabledData,
	type TwoFactorEnabledData,
	twoFactorCodeTemplate,
	twoFactorDisabledTemplate,
	twoFactorEnabledTemplate,
} from "./two-factor";

// User invite templates
export {
	type InviteAcceptedData,
	type InviteExpiredData,
	inviteAcceptedTemplate,
	inviteExpiredTemplate,
	type UserInviteData,
	userInviteTemplate,
} from "./user-invite";
