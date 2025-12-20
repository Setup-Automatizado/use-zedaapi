/**
 * Base Email Template
 *
 * Modern, responsive email template following the project design system.
 * Uses inline CSS for maximum email client compatibility.
 *
 * Theme: base=radix&style=maia&baseColor=neutral&theme=lime
 *
 * @module lib/email/templates/base
 */

import { appConfig, brandColors } from "../config";

export interface BaseTemplateProps {
	/** Email title (used in preview) */
	title: string;
	/** Preview text shown in email client */
	previewText?: string;
	/** Main content HTML */
	content: string;
	/** Optional footer content */
	footerContent?: string;
	/** Current year for copyright */
	year?: number;
}

/**
 * Generates the base HTML email template
 * Fully responsive with dark mode support
 */
export function baseTemplate({
	title,
	previewText,
	content,
	footerContent,
	year = new Date().getFullYear(),
}: BaseTemplateProps): string {
	const { name, url, supportEmail, logoUrl } = appConfig;
	const c = brandColors;

	return `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="x-apple-disable-message-reformatting">
  <meta name="color-scheme" content="light dark">
  <meta name="supported-color-schemes" content="light dark">
  <title>${title}</title>
  <!--[if mso]>
  <noscript>
    <xml>
      <o:OfficeDocumentSettings>
        <o:PixelsPerInch>96</o:PixelsPerInch>
      </o:OfficeDocumentSettings>
    </xml>
  </noscript>
  <![endif]-->
  <style>
    :root {
      color-scheme: light dark;
      supported-color-schemes: light dark;
    }

    @media (prefers-color-scheme: dark) {
      .email-body {
        background-color: ${c.backgroundDark} !important;
      }
      .email-container {
        background-color: #262626 !important;
      }
      .text-primary {
        color: ${c.foregroundDark} !important;
      }
      .text-secondary {
        color: #a3a3a3 !important;
      }
      .border-color {
        border-color: ${c.borderDark} !important;
      }
      .footer-text {
        color: #737373 !important;
      }
    }

    @media only screen and (max-width: 600px) {
      .email-container {
        width: 100% !important;
        padding: 16px !important;
      }
      .content-padding {
        padding: 24px 16px !important;
      }
      .button {
        width: 100% !important;
        display: block !important;
      }
    }
  </style>
</head>
<body class="email-body" style="margin: 0; padding: 0; background-color: ${c.background}; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
  ${
		previewText
			? `
  <!-- Preview Text -->
  <div style="display: none; max-height: 0; overflow: hidden; mso-hide: all;">
    ${previewText}
    ${"&nbsp;".repeat(150)}
  </div>
  `
			: ""
	}

  <!-- Email Wrapper -->
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="background-color: ${c.background};">
    <tr>
      <td align="center" style="padding: 40px 20px;">
        <!-- Email Container -->
        <table role="presentation" class="email-container" width="600" cellspacing="0" cellpadding="0" border="0" style="background-color: ${c.background}; border-radius: 16px; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1); overflow: hidden;">

          <!-- Header -->
          <tr>
            <td style="padding: 32px 40px; text-align: center; border-bottom: 1px solid ${c.border};" class="border-color">
              <a href="${url}" style="text-decoration: none; display: inline-flex; align-items: center; gap: 12px;">
                <img src="${logoUrl}" alt="${name}" width="48" height="48" style="display: block; border-radius: 12px;">
                <span style="font-size: 24px; font-weight: 700; color: ${c.foreground}; letter-spacing: -0.5px;" class="text-primary">${name}</span>
              </a>
            </td>
          </tr>

          <!-- Content -->
          <tr>
            <td class="content-padding" style="padding: 40px;">
              ${content}
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="padding: 24px 40px; background-color: #fafafa; border-top: 1px solid ${c.border};" class="border-color">
              ${
								footerContent ||
								`
              <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
                <tr>
                  <td style="text-align: center;">
                    <p class="footer-text" style="margin: 0 0 8px; font-size: 13px; color: ${c.muted};">
                      This email was sent by <a href="${url}" style="color: ${c.primary}; text-decoration: none; font-weight: 500;">${name}</a>
                    </p>
                    <p class="footer-text" style="margin: 0 0 8px; font-size: 13px; color: ${c.muted};">
                      If you did not request this email, you can safely ignore it.
                    </p>
                    <p class="footer-text" style="margin: 0; font-size: 13px; color: ${c.muted};">
                      Need help? <a href="mailto:${supportEmail}" style="color: ${c.primary}; text-decoration: none; font-weight: 500;">Contact us</a>
                    </p>
                  </td>
                </tr>
              </table>
              `
							}
            </td>
          </tr>

          <!-- Copyright -->
          <tr>
            <td style="padding: 16px 40px; text-align: center;">
              <p class="footer-text" style="margin: 0; font-size: 12px; color: ${c.mutedLight};">
                &copy; ${year} ${name}. All rights reserved.
              </p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>
  `.trim();
}

/**
 * Creates a primary action button
 */
export function primaryButton(text: string, href: string): string {
	const c = brandColors;
	return `
    <table role="presentation" cellspacing="0" cellpadding="0" border="0" style="margin: 24px 0;">
      <tr>
        <td>
          <a href="${href}" class="button" style="display: inline-block; padding: 14px 32px; background-color: ${c.primary}; color: #ffffff; font-size: 15px; font-weight: 600; text-decoration: none; border-radius: 10px; text-align: center; transition: background-color 0.2s;">
            ${text}
          </a>
        </td>
      </tr>
    </table>
  `;
}

/**
 * Creates a secondary/outline button
 */
export function secondaryButton(text: string, href: string): string {
	const c = brandColors;
	return `
    <table role="presentation" cellspacing="0" cellpadding="0" border="0" style="margin: 24px 0;">
      <tr>
        <td>
          <a href="${href}" class="button" style="display: inline-block; padding: 12px 28px; background-color: transparent; color: ${c.primary}; font-size: 15px; font-weight: 600; text-decoration: none; border-radius: 10px; border: 2px solid ${c.primary}; text-align: center;">
            ${text}
          </a>
        </td>
      </tr>
    </table>
  `;
}

/**
 * Creates an info box for codes or important text
 */
export function infoBox(
	content: string,
	variant: "code" | "info" | "warning" | "success" = "info",
): string {
	const c = brandColors;
	const colors = {
		code: { bg: "#f5f5f5", border: c.border, text: c.foreground },
		info: { bg: "#f0f9ff", border: "#bae6fd", text: "#0369a1" },
		warning: { bg: "#fffbeb", border: "#fde68a", text: "#b45309" },
		success: { bg: "#f0fdf4", border: "#bbf7d0", text: "#15803d" },
	};
	const style = colors[variant];

	return `
    <div style="padding: 20px; background-color: ${style.bg}; border: 1px solid ${style.border}; border-radius: 12px; margin: 16px 0; ${variant === "code" ? "text-align: center;" : ""}">
      ${
				variant === "code"
					? `
        <span style="font-family: 'SF Mono', Monaco, 'Courier New', monospace; font-size: 28px; font-weight: 700; letter-spacing: 4px; color: ${style.text};">
          ${content}
        </span>
      `
					: `
        <p style="margin: 0; font-size: 14px; color: ${style.text}; line-height: 1.6;">
          ${content}
        </p>
      `
			}
    </div>
  `;
}

/**
 * Creates a divider line
 */
export function divider(): string {
	return `<hr style="border: none; border-top: 1px solid ${brandColors.border}; margin: 24px 0;">`;
}

/**
 * Creates a text paragraph
 */
export function paragraph(
	text: string,
	options?: { muted?: boolean; small?: boolean; center?: boolean },
): string {
	const c = brandColors;
	const style = [
		"margin: 0 0 16px",
		`font-size: ${options?.small ? "13px" : "15px"}`,
		`color: ${options?.muted ? c.muted : c.foreground}`,
		"line-height: 1.6",
		options?.center ? "text-align: center" : "",
	]
		.filter(Boolean)
		.join("; ");

	return `<p class="text-${options?.muted ? "secondary" : "primary"}" style="${style}">${text}</p>`;
}

/**
 * Creates a heading
 */
export function heading(text: string, level: 1 | 2 | 3 = 1): string {
	const c = brandColors;
	const sizes = { 1: "28px", 2: "22px", 3: "18px" };
	const margins = { 1: "0 0 24px", 2: "0 0 16px", 3: "0 0 12px" };

	return `<h${level} class="text-primary" style="margin: ${margins[level]}; font-size: ${sizes[level]}; font-weight: 700; color: ${c.foreground}; line-height: 1.3; letter-spacing: -0.5px;">${text}</h${level}>`;
}
