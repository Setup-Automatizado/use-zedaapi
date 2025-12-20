/**
 * Login Alert Email Template
 *
 * Sent when a new login is detected from a new device or location.
 * Provides security information and action links.
 *
 * @module lib/email/templates/login-alert
 */

import { appConfig, brandColors } from "../config";
import {
	baseTemplate,
	divider,
	heading,
	infoBox,
	paragraph,
	primaryButton,
	secondaryButton,
} from "./base";

export interface LoginAlertData {
	/** User's name */
	userName: string;
	/** User's email */
	userEmail: string;
	/** Device/browser information */
	device: string;
	/** IP address */
	ipAddress: string;
	/** Location (city, country) */
	location?: string;
	/** Login timestamp */
	loginTime: Date;
	/** Link to security settings */
	securitySettingsUrl?: string;
	/** Link to report unauthorized access */
	reportUrl?: string;
}

/**
 * Generates a login alert email
 */
export function loginAlertTemplate(data: LoginAlertData): string {
	const {
		userName,
		userEmail,
		device,
		ipAddress,
		location,
		loginTime,
		securitySettingsUrl = `${appConfig.url}/settings/security`,
		reportUrl = `${appConfig.url}/settings/security?action=report`,
	} = data;

	const c = brandColors;
	const formattedTime = loginTime.toLocaleString("en-US", {
		dateStyle: "full",
		timeStyle: "short",
	});

	const content = `
    ${heading("New login detected")}

    ${paragraph(`Hello, <strong>${userName || "User"}</strong>!`)}

    ${paragraph("We detected a new login to your account. If this was you, you can ignore this email. Otherwise, we recommend that you change your password immediately.")}

    ${infoBox(
			`
      <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">Email:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${userEmail}</span>
          </td>
        </tr>
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">Device:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${device}</span>
          </td>
        </tr>
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">IP:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${ipAddress}</span>
          </td>
        </tr>
        ${
					location
						? `
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">Location:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${location}</span>
          </td>
        </tr>
        `
						: ""
				}
        <tr>
          <td style="padding: 8px 0;">
            <strong style="color: ${c.foreground};">Date and time:</strong>
            <span style="color: ${c.muted}; margin-left: 8px;">${formattedTime}</span>
          </td>
        </tr>
      </table>
    `,
			"info",
		)}

    ${divider()}

    ${paragraph("If you recognize this login, no action is required.", { muted: true })}

    ${paragraph("<strong>Was this not you?</strong> We recommend that you:", { muted: false })}

    <ul style="margin: 16px 0; padding-left: 24px; color: ${c.foreground};">
      <li style="margin-bottom: 8px; line-height: 1.6;">Change your password immediately</li>
      <li style="margin-bottom: 8px; line-height: 1.6;">Enable two-factor authentication</li>
      <li style="margin-bottom: 8px; line-height: 1.6;">Review connected devices</li>
    </ul>

    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
      <tr>
        <td align="center">
          ${primaryButton("Check Security Settings", securitySettingsUrl)}
        </td>
      </tr>
      <tr>
        <td align="center">
          ${secondaryButton("This was not me - Report access", reportUrl)}
        </td>
      </tr>
    </table>
  `;

	return baseTemplate({
		title: "New login detected - WhatsApp Manager",
		previewText: `New login detected on your account from ${device}`,
		content,
	});
}
