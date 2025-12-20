/**
 * Email Sender Service
 *
 * Central service for sending transactional emails using Nodemailer.
 * Integrates with all email templates and provides logging.
 *
 * @module lib/email/sender
 */

import type { SendMailOptions } from "nodemailer";
import { appConfig, emailConfig, getTransporter } from "./config";
import {
	type InviteAcceptedData,
	type InviteExpiredData,
	inviteAcceptedTemplate,
	inviteExpiredTemplate,
	type LoginAlertData,
	loginAlertTemplate,
	type PasswordChangedData,
	type PasswordResetData,
	passwordChangedTemplate,
	passwordResetTemplate,
	type TwoFactorCodeData,
	type TwoFactorDisabledData,
	type TwoFactorEnabledData,
	twoFactorCodeTemplate,
	twoFactorDisabledTemplate,
	twoFactorEnabledTemplate,
	type UserInviteData,
	userInviteTemplate,
} from "./templates";

/**
 * Email send result
 */
export interface EmailResult {
	success: boolean;
	messageId?: string;
	error?: string;
}

/**
 * Base send email function
 */
async function sendEmail(options: SendMailOptions): Promise<EmailResult> {
	try {
		const transporter = getTransporter();
		const from = `${emailConfig.from.name} <${emailConfig.from.address}>`;

		const result = await transporter.sendMail({
			from,
			...options,
		});

		console.log(`Email sent successfully: ${result.messageId}`);
		return {
			success: true,
			messageId: result.messageId,
		};
	} catch (error) {
		const errorMessage =
			error instanceof Error ? error.message : "Unknown error";
		console.error(`Failed to send email: ${errorMessage}`);
		return {
			success: false,
			error: errorMessage,
		};
	}
}

/**
 * Send login alert email
 */
export async function sendLoginAlert(
	to: string,
	data: LoginAlertData,
): Promise<EmailResult> {
	const html = loginAlertTemplate(data);
	return sendEmail({
		to,
		subject: `New login detected - ${appConfig.name}`,
		html,
	});
}

/**
 * Send password reset email
 */
export async function sendPasswordReset(
	to: string,
	data: PasswordResetData,
): Promise<EmailResult> {
	const html = passwordResetTemplate(data);
	return sendEmail({
		to,
		subject: `Reset your password - ${appConfig.name}`,
		html,
	});
}

/**
 * Send password changed confirmation email
 */
export async function sendPasswordChanged(
	to: string,
	data: PasswordChangedData,
): Promise<EmailResult> {
	const html = passwordChangedTemplate(data);
	return sendEmail({
		to,
		subject: `Password changed successfully - ${appConfig.name}`,
		html,
	});
}

/**
 * Send 2FA verification code email
 */
export async function sendTwoFactorCode(
	to: string,
	data: TwoFactorCodeData,
): Promise<EmailResult> {
	const html = twoFactorCodeTemplate(data);
	return sendEmail({
		to,
		subject: `Your verification code: ${data.verificationCode} - ${appConfig.name}`,
		html,
	});
}

/**
 * Send 2FA enabled confirmation email
 */
export async function sendTwoFactorEnabled(
	to: string,
	data: TwoFactorEnabledData,
): Promise<EmailResult> {
	const html = twoFactorEnabledTemplate(data);
	return sendEmail({
		to,
		subject: `Two-factor authentication enabled - ${appConfig.name}`,
		html,
	});
}

/**
 * Send 2FA disabled confirmation email
 */
export async function sendTwoFactorDisabled(
	to: string,
	data: TwoFactorDisabledData,
): Promise<EmailResult> {
	const html = twoFactorDisabledTemplate(data);
	return sendEmail({
		to,
		subject: `Two-factor authentication disabled - ${appConfig.name}`,
		html,
	});
}

/**
 * Send user invitation email
 */
export async function sendUserInvite(
	to: string,
	data: UserInviteData,
): Promise<EmailResult> {
	const html = userInviteTemplate(data);
	const teamName = data.organizationName || appConfig.name;
	return sendEmail({
		to,
		subject: `${data.inviterName} invited you to ${teamName}`,
		html,
	});
}

/**
 * Send invite accepted notification email (to inviter)
 */
export async function sendInviteAccepted(
	to: string,
	data: InviteAcceptedData,
): Promise<EmailResult> {
	const html = inviteAcceptedTemplate(data);
	return sendEmail({
		to,
		subject: `${data.newUserName} accepted your invitation - ${appConfig.name}`,
		html,
	});
}

/**
 * Send invite expired notification email
 */
export async function sendInviteExpired(
	to: string,
	data: InviteExpiredData,
): Promise<EmailResult> {
	const html = inviteExpiredTemplate(data);
	const teamName = data.organizationName || appConfig.name;
	return sendEmail({
		to,
		subject: `Your invitation to ${teamName} has expired`,
		html,
	});
}

/**
 * Email service object with all send methods
 */
export const emailService = {
	sendLoginAlert,
	sendPasswordReset,
	sendPasswordChanged,
	sendTwoFactorCode,
	sendTwoFactorEnabled,
	sendTwoFactorDisabled,
	sendUserInvite,
	sendInviteAccepted,
	sendInviteExpired,
} as const;

export default emailService;
