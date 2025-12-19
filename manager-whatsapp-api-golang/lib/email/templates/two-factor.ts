/**
 * Two-Factor Authentication Email Template
 *
 * Sent when user requests 2FA verification code via email.
 * Also handles 2FA enable/disable confirmations.
 *
 * @module lib/email/templates/two-factor
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

export interface TwoFactorCodeData {
  /** User's name */
  userName: string;
  /** User's email */
  userEmail: string;
  /** 6-digit verification code */
  verificationCode: string;
  /** Code expiration time in minutes */
  expiresIn?: number;
  /** Device/browser information */
  device?: string;
  /** IP address */
  ipAddress?: string;
  /** Verification link (optional alternative to code) */
  verifyUrl?: string;
}

/**
 * Generates a 2FA verification code email
 */
export function twoFactorCodeTemplate(data: TwoFactorCodeData): string {
  const {
    userName,
    verificationCode,
    expiresIn = 10,
    device,
    ipAddress,
    verifyUrl,
  } = data;

  const c = brandColors;

  const content = `
    ${heading('Verification code')}

    ${paragraph(`Hello, <strong>${userName || 'User'}</strong>!`)}

    ${paragraph('Use the code below to complete your two-factor authentication:')}

    ${infoBox(verificationCode, 'code')}

    ${infoBox(`
      <strong>Attention:</strong> This code expires in <strong>${expiresIn} minutes</strong>.
      Do not share this code with anyone.
    `, 'warning')}

    ${verifyUrl ? `
    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
      <tr>
        <td align="center">
          ${primaryButton('Verify Code Automatically', verifyUrl)}
        </td>
      </tr>
    </table>
    ` : ''}

    ${(device || ipAddress) ? `
    ${divider()}

    ${paragraph('Request details:', { muted: false })}

    ${infoBox(`
      <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
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
    `, 'info')}
    ` : ''}

    ${divider()}

    ${paragraph('If you did not request this code, someone may be trying to access your account. We recommend that you change your password immediately.', { muted: true })}
  `;

  return baseTemplate({
    title: 'Verification code - WhatsApp Manager',
    previewText: `Your verification code is: ${verificationCode}`,
    content,
  });
}

export interface TwoFactorEnabledData {
  /** User's name */
  userName: string;
  /** User's email */
  userEmail: string;
  /** Enable timestamp */
  enabledAt: Date;
  /** Recovery codes (if applicable) */
  recoveryCodes?: string[];
  /** Link to security settings */
  securitySettingsUrl?: string;
}

/**
 * Generates a 2FA enabled confirmation email
 */
export function twoFactorEnabledTemplate(data: TwoFactorEnabledData): string {
  const {
    userName,
    userEmail,
    enabledAt,
    recoveryCodes,
    securitySettingsUrl = `${appConfig.url}/settings/security`,
  } = data;

  const c = brandColors;
  const formattedTime = enabledAt.toLocaleString('en-US', {
    dateStyle: 'full',
    timeStyle: 'short',
  });

  const content = `
    ${heading('Two-factor authentication enabled')}

    ${paragraph(`Hello, <strong>${userName || 'User'}</strong>!`)}

    ${paragraph(`Two-factor authentication has been <strong>successfully enabled</strong> for your account <strong>${userEmail}</strong>.`)}

    ${infoBox(`
      <div style="text-align: center;">
        <div style="display: inline-block; width: 48px; height: 48px; background-color: ${c.success}; border-radius: 50%; margin-bottom: 12px;">
          <span style="color: #ffffff; font-size: 24px; line-height: 48px;">&#10003;</span>
        </div>
        <p style="margin: 0; color: ${c.success}; font-weight: 600;">Your account is now more secure!</p>
      </div>
    `, 'success')}

    ${paragraph(`<strong>Enabled on:</strong> ${formattedTime}`, { muted: true })}

    ${recoveryCodes && recoveryCodes.length > 0 ? `
    ${divider()}

    ${heading('Recovery codes', 2)}

    ${paragraph('Keep these codes in a safe place. They can be used to access your account if you lose access to your authentication device.')}

    ${infoBox(`
      <div style="font-family: 'SF Mono', Monaco, 'Courier New', monospace; font-size: 14px; line-height: 2;">
        ${recoveryCodes.map(code => `<div style="color: ${c.foreground};">${code}</div>`).join('')}
      </div>
    `, 'code')}

    ${infoBox(`
      <strong>Important:</strong> Each code can only be used once.
      Keep them in a safe place and do not share with anyone.
    `, 'warning')}
    ` : ''}

    ${divider()}

    ${paragraph('From now on, an additional verification code will be required when logging in.', { muted: true })}

    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
      <tr>
        <td align="center">
          ${primaryButton('Manage Security Settings', securitySettingsUrl)}
        </td>
      </tr>
    </table>
  `;

  return baseTemplate({
    title: '2FA enabled - WhatsApp Manager',
    previewText: 'Two-factor authentication successfully enabled on your account',
    content,
  });
}

export interface TwoFactorDisabledData {
  /** User's name */
  userName: string;
  /** User's email */
  userEmail: string;
  /** Disable timestamp */
  disabledAt: Date;
  /** Device/browser information */
  device?: string;
  /** IP address */
  ipAddress?: string;
  /** Link to re-enable 2FA */
  enableUrl?: string;
}

/**
 * Generates a 2FA disabled confirmation email
 */
export function twoFactorDisabledTemplate(data: TwoFactorDisabledData): string {
  const {
    userName,
    userEmail,
    disabledAt,
    device,
    ipAddress,
    enableUrl = `${appConfig.url}/settings/security`,
  } = data;

  const c = brandColors;
  const formattedTime = disabledAt.toLocaleString('en-US', {
    dateStyle: 'full',
    timeStyle: 'short',
  });

  const content = `
    ${heading('Two-factor authentication disabled')}

    ${paragraph(`Hello, <strong>${userName || 'User'}</strong>!`)}

    ${paragraph(`Two-factor authentication has been <strong>disabled</strong> for your account <strong>${userEmail}</strong>.`)}

    ${infoBox(`
      <div style="text-align: center;">
        <div style="display: inline-block; width: 48px; height: 48px; background-color: ${c.warning}; border-radius: 50%; margin-bottom: 12px;">
          <span style="color: #ffffff; font-size: 24px; line-height: 48px;">!</span>
        </div>
        <p style="margin: 0; color: ${c.warning}; font-weight: 600;">Your account is less protected</p>
      </div>
    `, 'warning')}

    ${infoBox(`
      <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
        <tr>
          <td style="padding: 4px 0;">
            <strong style="color: ${c.foreground};">Disabled on:</strong>
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
    `, 'info')}

    ${divider()}

    ${paragraph('<strong>Was this not you?</strong> If you did not disable two-factor authentication:', { muted: false })}

    <ul style="margin: 16px 0; padding-left: 24px; color: ${c.foreground};">
      <li style="margin-bottom: 8px; line-height: 1.6;">Change your password immediately</li>
      <li style="margin-bottom: 8px; line-height: 1.6;">Re-enable two-factor authentication</li>
      <li style="margin-bottom: 8px; line-height: 1.6;">Contact our support</li>
    </ul>

    ${paragraph('We strongly recommend keeping two-factor authentication enabled to protect your account.', { muted: true })}

    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
      <tr>
        <td align="center">
          ${primaryButton('Re-enable 2FA', enableUrl)}
        </td>
      </tr>
    </table>
  `;

  return baseTemplate({
    title: '2FA disabled - WhatsApp Manager',
    previewText: 'Attention: Two-factor authentication was disabled on your account',
    content,
  });
}
