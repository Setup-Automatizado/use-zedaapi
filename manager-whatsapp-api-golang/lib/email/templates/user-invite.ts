/**
 * User Invite Email Template
 *
 * Sent when an administrator invites a new user to join the platform.
 * Contains a secure invitation link with expiration time.
 *
 * @module lib/email/templates/user-invite
 */

import { appConfig, brandColors } from '../config';
import {
  baseTemplate,
  heading,
  paragraph,
  primaryButton,
  secondaryButton,
  infoBox,
  divider,
} from './base';

export interface UserInviteData {
  /** Invitee's email */
  inviteeEmail: string;
  /** Inviter's name */
  inviterName: string;
  /** Inviter's email */
  inviterEmail?: string;
  /** Organization/team name */
  organizationName?: string;
  /** Role being assigned */
  role?: string;
  /** Invitation link */
  inviteUrl: string;
  /** Token expiration time in hours */
  expiresIn?: number;
  /** Custom welcome message */
  welcomeMessage?: string;
}

/**
 * Generates a user invitation email
 */
export function userInviteTemplate(data: UserInviteData): string {
  const {
    inviteeEmail,
    inviterName,
    inviterEmail,
    organizationName,
    role,
    inviteUrl,
    expiresIn = 48,
    welcomeMessage,
  } = data;

  const c = brandColors;
  const teamName = organizationName || appConfig.name;

  const content = `
    ${heading('You have been invited!')}

    ${paragraph(`<strong>${inviterName}</strong>${inviterEmail ? ` (${inviterEmail})` : ''} has invited you to join <strong>${teamName}</strong>.`)}

    ${welcomeMessage ? `
    ${infoBox(`
      <div style="font-style: italic; color: ${c.foreground};">
        "${welcomeMessage}"
      </div>
      <div style="margin-top: 8px; text-align: right; color: ${c.muted}; font-size: 13px;">
        - ${inviterName}
      </div>
    `, 'info')}
    ` : ''}

    ${infoBox(`
      <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">Access email:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${inviteeEmail}</span>
          </td>
        </tr>
        ${role ? `
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">Role:</strong>
            <span style="color: ${c.primary}; margin-left: 8px; font-weight: 600;">${role}</span>
          </td>
        </tr>
        ` : ''}
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">Team:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${teamName}</span>
          </td>
        </tr>
      </table>
    `, 'info')}

    ${paragraph('To accept the invitation and create your account, click the button below:')}

    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
      <tr>
        <td align="center">
          ${primaryButton('Accept Invitation', inviteUrl)}
        </td>
      </tr>
    </table>

    ${infoBox(`
      <strong>Important:</strong> This invitation expires in <strong>${expiresIn} hours</strong>.
      After that period, you will need to request a new invitation.
    `, 'warning')}

    ${divider()}

    ${heading('What you will be able to do', 2)}

    <ul style="margin: 16px 0; padding-left: 24px; color: ${c.foreground};">
      <li style="margin-bottom: 12px; line-height: 1.6;">
        <strong>Manage WhatsApp instances</strong><br>
        <span style="color: ${c.muted}; font-size: 14px;">Create, configure and monitor your connections</span>
      </li>
      <li style="margin-bottom: 12px; line-height: 1.6;">
        <strong>Configure webhooks</strong><br>
        <span style="color: ${c.muted}; font-size: 14px;">Receive real-time events</span>
      </li>
      <li style="margin-bottom: 12px; line-height: 1.6;">
        <strong>Monitor API health</strong><br>
        <span style="color: ${c.muted}; font-size: 14px;">Track metrics and service status</span>
      </li>
    </ul>

    ${divider()}

    ${paragraph('If you were not expecting this invitation, you can safely ignore it.', { muted: true })}

    <div style="margin-top: 24px; padding: 16px; background-color: #f5f5f5; border-radius: 8px;">
      <p style="margin: 0 0 8px; font-size: 12px; color: ${c.muted}; font-weight: 600;">Cannot click the button? Copy and paste this link into your browser:</p>
      <p style="margin: 0; font-size: 12px; color: ${c.primary}; word-break: break-all;">${inviteUrl}</p>
    </div>
  `;

  return baseTemplate({
    title: `Invitation to ${teamName} - WhatsApp Manager`,
    previewText: `${inviterName} has invited you to join ${teamName}`,
    content,
  });
}

export interface InviteAcceptedData {
  /** Inviter's name */
  inviterName: string;
  /** Inviter's email */
  inviterEmail: string;
  /** New user's name */
  newUserName: string;
  /** New user's email */
  newUserEmail: string;
  /** Organization/team name */
  organizationName?: string;
  /** Role assigned */
  role?: string;
  /** Accepted timestamp */
  acceptedAt: Date;
  /** Link to team management */
  teamUrl?: string;
}

/**
 * Generates an invite accepted notification email (sent to inviter)
 */
export function inviteAcceptedTemplate(data: InviteAcceptedData): string {
  const {
    inviterName,
    newUserName,
    newUserEmail,
    organizationName,
    role,
    acceptedAt,
    teamUrl = `${appConfig.url}/settings/team`,
  } = data;

  const c = brandColors;
  const teamName = organizationName || appConfig.name;
  const formattedTime = acceptedAt.toLocaleString('en-US', {
    dateStyle: 'full',
    timeStyle: 'short',
  });

  const content = `
    ${heading('Invitation accepted!')}

    ${paragraph(`Hello, <strong>${inviterName}</strong>!`)}

    ${paragraph(`Good news! <strong>${newUserName}</strong> has accepted your invitation and is now part of <strong>${teamName}</strong>.`)}

    ${infoBox(`
      <div style="text-align: center;">
        <div style="display: inline-block; width: 48px; height: 48px; background-color: ${c.success}; border-radius: 50%; margin-bottom: 12px;">
          <span style="color: #ffffff; font-size: 24px; line-height: 48px;">&#10003;</span>
        </div>
        <p style="margin: 0; color: ${c.success}; font-weight: 600;">New member added successfully!</p>
      </div>
    `, 'success')}

    ${infoBox(`
      <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">Name:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${newUserName}</span>
          </td>
        </tr>
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">Email:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${newUserEmail}</span>
          </td>
        </tr>
        ${role ? `
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">Role:</strong>
            <span style="color: ${c.primary}; margin-left: 8px; font-weight: 600;">${role}</span>
          </td>
        </tr>
        ` : ''}
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">Joined on:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${formattedTime}</span>
          </td>
        </tr>
      </table>
    `, 'info')}

    ${divider()}

    ${paragraph('You can manage your team members at any time in the settings.', { muted: true })}

    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
      <tr>
        <td align="center">
          ${primaryButton('Manage Team', teamUrl)}
        </td>
      </tr>
    </table>
  `;

  return baseTemplate({
    title: 'Invitation accepted - WhatsApp Manager',
    previewText: `${newUserName} accepted your invitation and is now part of ${teamName}`,
    content,
  });
}

export interface InviteExpiredData {
  /** Invitee's email */
  inviteeEmail: string;
  /** Inviter's name */
  inviterName: string;
  /** Organization/team name */
  organizationName?: string;
  /** Link to resend invite */
  resendUrl?: string;
}

/**
 * Generates an expired invite notification email
 */
export function inviteExpiredTemplate(data: InviteExpiredData): string {
  const {
    inviteeEmail,
    inviterName,
    organizationName,
    resendUrl = `${appConfig.url}/settings/team/invite`,
  } = data;

  const c = brandColors;
  const teamName = organizationName || appConfig.name;

  const content = `
    ${heading('Invitation expired')}

    ${paragraph(`Hello!`)}

    ${paragraph(`The invitation sent by <strong>${inviterName}</strong> to join <strong>${teamName}</strong> has expired.`)}

    ${infoBox(`
      <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">Invited email:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${inviteeEmail}</span>
          </td>
        </tr>
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">Team:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${teamName}</span>
          </td>
        </tr>
      </table>
    `, 'warning')}

    ${divider()}

    ${paragraph('If you still want to join, request a new invitation from the administrator.', { muted: true })}

    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
      <tr>
        <td align="center">
          ${secondaryButton('Request New Invitation', resendUrl)}
        </td>
      </tr>
    </table>
  `;

  return baseTemplate({
    title: 'Invitation expired - WhatsApp Manager',
    previewText: `Your invitation to join ${teamName} has expired`,
    content,
  });
}
