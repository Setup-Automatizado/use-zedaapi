/**
 * Email Sender Service
 *
 * Central service for sending transactional emails using Nodemailer.
 * Integrates with all email templates and provides logging.
 *
 * @module lib/email/sender
 */

import type { SendMailOptions } from 'nodemailer';
import { emailConfig, getTransporter, appConfig } from './config';
import {
  loginAlertTemplate,
  passwordResetTemplate,
  passwordChangedTemplate,
  twoFactorCodeTemplate,
  twoFactorEnabledTemplate,
  twoFactorDisabledTemplate,
  userInviteTemplate,
  inviteAcceptedTemplate,
  inviteExpiredTemplate,
  type LoginAlertData,
  type PasswordResetData,
  type PasswordChangedData,
  type TwoFactorCodeData,
  type TwoFactorEnabledData,
  type TwoFactorDisabledData,
  type UserInviteData,
  type InviteAcceptedData,
  type InviteExpiredData,
} from './templates';

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
    const errorMessage = error instanceof Error ? error.message : 'Unknown error';
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
  data: LoginAlertData
): Promise<EmailResult> {
  const html = loginAlertTemplate(data);
  return sendEmail({
    to,
    subject: `Novo acesso detectado - ${appConfig.name}`,
    html,
  });
}

/**
 * Send password reset email
 */
export async function sendPasswordReset(
  to: string,
  data: PasswordResetData
): Promise<EmailResult> {
  const html = passwordResetTemplate(data);
  return sendEmail({
    to,
    subject: `Redefinir sua senha - ${appConfig.name}`,
    html,
  });
}

/**
 * Send password changed confirmation email
 */
export async function sendPasswordChanged(
  to: string,
  data: PasswordChangedData
): Promise<EmailResult> {
  const html = passwordChangedTemplate(data);
  return sendEmail({
    to,
    subject: `Senha alterada com sucesso - ${appConfig.name}`,
    html,
  });
}

/**
 * Send 2FA verification code email
 */
export async function sendTwoFactorCode(
  to: string,
  data: TwoFactorCodeData
): Promise<EmailResult> {
  const html = twoFactorCodeTemplate(data);
  return sendEmail({
    to,
    subject: `Seu codigo de verificacao: ${data.verificationCode} - ${appConfig.name}`,
    html,
  });
}

/**
 * Send 2FA enabled confirmation email
 */
export async function sendTwoFactorEnabled(
  to: string,
  data: TwoFactorEnabledData
): Promise<EmailResult> {
  const html = twoFactorEnabledTemplate(data);
  return sendEmail({
    to,
    subject: `Autenticacao de dois fatores ativada - ${appConfig.name}`,
    html,
  });
}

/**
 * Send 2FA disabled confirmation email
 */
export async function sendTwoFactorDisabled(
  to: string,
  data: TwoFactorDisabledData
): Promise<EmailResult> {
  const html = twoFactorDisabledTemplate(data);
  return sendEmail({
    to,
    subject: `Autenticacao de dois fatores desativada - ${appConfig.name}`,
    html,
  });
}

/**
 * Send user invitation email
 */
export async function sendUserInvite(
  to: string,
  data: UserInviteData
): Promise<EmailResult> {
  const html = userInviteTemplate(data);
  const teamName = data.organizationName || appConfig.name;
  return sendEmail({
    to,
    subject: `${data.inviterName} convidou voce para ${teamName}`,
    html,
  });
}

/**
 * Send invite accepted notification email (to inviter)
 */
export async function sendInviteAccepted(
  to: string,
  data: InviteAcceptedData
): Promise<EmailResult> {
  const html = inviteAcceptedTemplate(data);
  return sendEmail({
    to,
    subject: `${data.newUserName} aceitou seu convite - ${appConfig.name}`,
    html,
  });
}

/**
 * Send invite expired notification email
 */
export async function sendInviteExpired(
  to: string,
  data: InviteExpiredData
): Promise<EmailResult> {
  const html = inviteExpiredTemplate(data);
  const teamName = data.organizationName || appConfig.name;
  return sendEmail({
    to,
    subject: `Seu convite para ${teamName} expirou`,
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
