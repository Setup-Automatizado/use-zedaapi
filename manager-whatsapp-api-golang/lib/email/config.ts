/**
 * Email Configuration
 *
 * SMTP configuration for transactional emails using Nodemailer.
 * Supports multiple providers: Gmail, SendGrid, Mailgun, AWS SES, etc.
 *
 * @module lib/email/config
 */

import nodemailer from 'nodemailer';
import type { Transporter } from 'nodemailer';

/**
 * Email configuration from environment variables
 */
export const emailConfig = {
  host: process.env.SMTP_HOST || 'smtp.gmail.com',
  port: parseInt(process.env.SMTP_PORT || '587', 10),
  secure: process.env.SMTP_SECURE === 'true',
  auth: {
    user: process.env.SMTP_USER || '',
    pass: process.env.SMTP_PASSWORD || '',
  },
  from: {
    name: process.env.EMAIL_FROM_NAME || 'WhatsApp Manager',
    address: process.env.EMAIL_FROM_ADDRESS || 'noreply@whatsapp-manager.com',
  },
} as const;

/**
 * Application URLs for email links
 */
export const appConfig = {
  name: 'WhatsApp Manager',
  url: process.env.BETTER_AUTH_URL || 'http://localhost:3000',
  supportEmail: process.env.SUPPORT_EMAIL || 'suporte@whatsapp-manager.com',
  logoUrl: `${process.env.BETTER_AUTH_URL || 'http://localhost:3000'}/android-chrome-192x192.png`,
} as const;

/**
 * Brand colors following the project theme
 * base=radix&style=maia&baseColor=neutral&theme=lime
 */
export const brandColors = {
  primary: '#84cc16', // lime-500 (oklch(0.65 0.18 132))
  primaryDark: '#65a30d', // lime-600
  primaryLight: '#bef264', // lime-300
  background: '#ffffff',
  backgroundDark: '#171717', // neutral-900
  foreground: '#171717', // neutral-900
  foregroundDark: '#fafafa', // neutral-50
  muted: '#737373', // neutral-500
  mutedLight: '#d4d4d4', // neutral-300
  border: '#e5e5e5', // neutral-200
  borderDark: '#404040', // neutral-700
  success: '#22c55e', // green-500
  warning: '#f59e0b', // amber-500
  error: '#ef4444', // red-500
} as const;

/**
 * Create and configure the nodemailer transporter
 */
let transporter: Transporter | null = null;

export function getTransporter(): Transporter {
  if (!transporter) {
    if (!emailConfig.auth.user || !emailConfig.auth.pass) {
      console.warn('SMTP credentials not configured. Email sending will be disabled.');
      // Return a mock transporter for development
      transporter = nodemailer.createTransport({
        jsonTransport: true,
      });
    } else {
      transporter = nodemailer.createTransport({
        host: emailConfig.host,
        port: emailConfig.port,
        secure: emailConfig.secure,
        auth: emailConfig.auth,
      });
    }
  }
  return transporter;
}

/**
 * Verify SMTP connection
 */
export async function verifyEmailConnection(): Promise<boolean> {
  try {
    const transport = getTransporter();
    await transport.verify();
    console.log('SMTP connection verified successfully');
    return true;
  } catch (error) {
    console.error('SMTP connection failed:', error);
    return false;
  }
}
