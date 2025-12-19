/**
 * Password Reset Email Template
 *
 * Sent when a user requests to reset their password.
 * Contains a secure link with expiration time.
 *
 * @module lib/email/templates/password-reset
 */

import { appConfig, brandColors } from '../config';
import {
  baseTemplate,
  heading,
  paragraph,
  primaryButton,
  infoBox,
  divider,
} from './base';

export interface PasswordResetData {
  /** User's name */
  userName: string;
  /** User's email */
  userEmail: string;
  /** Password reset link */
  resetUrl: string;
  /** Token expiration time in minutes */
  expiresIn?: number;
}

/**
 * Generates a password reset email
 */
export function passwordResetTemplate(data: PasswordResetData): string {
  const {
    userName,
    userEmail,
    resetUrl,
    expiresIn = 60,
  } = data;

  const c = brandColors;

  const content = `
    ${heading('Reset your password')}

    ${paragraph(`Hello, <strong>${userName || 'User'}</strong>!`)}

    ${paragraph(`We received a request to reset the password for the account associated with <strong>${userEmail}</strong>.`)}

    ${paragraph('Click the button below to create a new password:')}

    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
      <tr>
        <td align="center">
          ${primaryButton('Reset Password', resetUrl)}
        </td>
      </tr>
    </table>

    ${infoBox(`
      <strong>Important:</strong> This link expires in <strong>${expiresIn} minutes</strong>.
      After that period, you will need to request a new reset link.
    `, 'warning')}

    ${divider()}

    ${paragraph('If you did not request a password reset, ignore this email. Your password will remain unchanged.', { muted: true })}

    ${paragraph('For security reasons, this link can only be used once.', { muted: true, small: true })}

    <div style="margin-top: 24px; padding: 16px; background-color: #f5f5f5; border-radius: 8px;">
      <p style="margin: 0 0 8px; font-size: 12px; color: ${c.muted}; font-weight: 600;">Cannot click the button? Copy and paste this link into your browser:</p>
      <p style="margin: 0; font-size: 12px; color: ${c.primary}; word-break: break-all;">${resetUrl}</p>
    </div>
  `;

  return baseTemplate({
    title: 'Reset password - WhatsApp Manager',
    previewText: `Password reset request for ${userEmail}`,
    content,
  });
}

/**
 * Password Changed Confirmation Email Template
 *
 * Sent after a password has been successfully changed.
 */
export interface PasswordChangedData {
  /** User's name */
  userName: string;
  /** User's email */
  userEmail: string;
  /** Change timestamp */
  changedAt: Date;
  /** Device/browser information */
  device?: string;
  /** IP address */
  ipAddress?: string;
  /** Link to security settings */
  securitySettingsUrl?: string;
}

/**
 * Generates a password changed confirmation email
 */
export function passwordChangedTemplate(data: PasswordChangedData): string {
  const {
    userName,
    userEmail,
    changedAt,
    device,
    ipAddress,
    securitySettingsUrl = `${appConfig.url}/settings/security`,
  } = data;

  const c = brandColors;
  const formattedTime = changedAt.toLocaleString('en-US', {
    dateStyle: 'full',
    timeStyle: 'short',
  });

  const content = `
    ${heading('Password changed successfully')}

    ${paragraph(`Hello, <strong>${userName || 'User'}</strong>!`)}

    ${paragraph(`The password for your account <strong>${userEmail}</strong> has been successfully changed.`)}

    ${infoBox(`
      <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
        <tr>
          <td style="padding: 4px 0;">
            <strong style="color: ${c.foreground};">Date and time:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${formattedTime}</span>
          </td>
        </tr>
        ${device ? `
        <tr>
          <td style="padding: 4px 0;">
            <strong style="color: ${c.foreground};">Device:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${device}</span>
          </td>
        </tr>
        ` : ''}
        ${ipAddress ? `
        <tr>
          <td style="padding: 4px 0;">
            <strong style="color: ${c.foreground};">IP:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${ipAddress}</span>
          </td>
        </tr>
        ` : ''}
      </table>
    `, 'success')}

    ${divider()}

    ${paragraph('<strong>Was this not you?</strong> If you did not make this change, your account may be compromised. We recommend that you:', { muted: false })}

    <ul style="margin: 16px 0; padding-left: 24px; color: ${c.foreground};">
      <li style="margin-bottom: 8px; line-height: 1.6;">Contact our support immediately</li>
      <li style="margin-bottom: 8px; line-height: 1.6;">Review recent activity on your account</li>
    </ul>

    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
      <tr>
        <td align="center">
          ${primaryButton('Check Security Settings', securitySettingsUrl)}
        </td>
      </tr>
    </table>
  `;

  return baseTemplate({
    title: 'Password changed - WhatsApp Manager',
    previewText: `Your password was successfully changed on ${formattedTime}`,
    content,
  });
}
