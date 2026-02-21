import nodemailer from "nodemailer";
import type { Transporter } from "nodemailer";

// =============================================================================
// Email Transport (Nodemailer Singleton)
// =============================================================================

let _transporter: Transporter | null = null;

function getTransporter(): Transporter {
	if (_transporter) return _transporter;

	const host = process.env.SMTP_HOST;
	const port = Number(process.env.SMTP_PORT) || 587;
	const user = process.env.SMTP_USER;
	const pass = process.env.SMTP_PASSWORD;

	if (!host || !user || !pass) {
		throw new Error(
			"SMTP not configured: set SMTP_HOST, SMTP_USER, SMTP_PASSWORD",
		);
	}

	_transporter = nodemailer.createTransport({
		host,
		port,
		secure: port === 465,
		auth: { user, pass },
	});

	return _transporter;
}

const defaultFrom =
	process.env.SMTP_FROM || "Zé da API Manager <noreply@zedaapi.com>";

/**
 * Send a plain HTML email.
 */
export async function sendEmail(
	to: string,
	subject: string,
	html: string,
): Promise<void> {
	const transporter = getTransporter();

	await transporter.sendMail({
		from: defaultFrom,
		to,
		subject,
		html,
	});
}

/**
 * Render a React Email template and send it.
 * Uses dynamic import to avoid bundling all templates at startup.
 */
export async function sendTemplateEmail(
	to: string,
	templateSlug: string,
	data: Record<string, unknown>,
): Promise<void> {
	const { render } = await import("@react-email/components");

	const templates: Record<
		string,
		() => Promise<{ default: React.FC<Record<string, unknown>> }>
	> = {
		// Existing templates
		welcome: () => import("@/emails/templates/welcome"),
		invoice: () => import("@/emails/templates/invoice"),
		"subscription-change": () =>
			import("@/emails/templates/subscription-change"),
		"nfse-issued": () => import("@/emails/templates/nfse-issued"),
		"payment-failed": () => import("@/emails/templates/payment-failed"),
		"waitlist-approved": () =>
			import("@/emails/templates/waitlist-approved"),

		// Auth & Security
		"magic-link": () => import("@/emails/templates/magic-link"),
		"two-factor-enabled": () =>
			import("@/emails/templates/two-factor-enabled"),
		"two-factor-code": () => import("@/emails/templates/two-factor-code"),
		"password-changed": () => import("@/emails/templates/password-changed"),
		"email-changed": () => import("@/emails/templates/email-changed"),
		"login-new-device": () => import("@/emails/templates/login-new-device"),
		"account-deactivated": () =>
			import("@/emails/templates/account-deactivated"),

		// Subscription Lifecycle
		"subscription-renewed": () =>
			import("@/emails/templates/subscription-renewed"),
		"subscription-upgraded": () =>
			import("@/emails/templates/subscription-upgraded"),
		"subscription-downgraded": () =>
			import("@/emails/templates/subscription-downgraded"),
		"subscription-resumed": () =>
			import("@/emails/templates/subscription-resumed"),
		"trial-started": () => import("@/emails/templates/trial-started"),
		"trial-ending": () => import("@/emails/templates/trial-ending"),

		// Payment & Billing
		"charge-refunded": () => import("@/emails/templates/charge-refunded"),
		"pix-charge-created": () =>
			import("@/emails/templates/pix-charge-created"),
		"boleto-charge-created": () =>
			import("@/emails/templates/boleto-charge-created"),
		"pix-expired": () => import("@/emails/templates/pix-expired"),
		"invoice-created": () => import("@/emails/templates/invoice-created"),
		"upcoming-renewal": () => import("@/emails/templates/upcoming-renewal"),
		"payment-method-updated": () =>
			import("@/emails/templates/payment-method-updated"),

		// Instance Management
		"instance-created": () => import("@/emails/templates/instance-created"),
		"instance-connected": () =>
			import("@/emails/templates/instance-connected"),
		"instance-deleted": () => import("@/emails/templates/instance-deleted"),

		// Affiliate System
		"affiliate-registered": () =>
			import("@/emails/templates/affiliate-registered"),
		"commission-earned": () =>
			import("@/emails/templates/commission-earned"),
		"referral-converted": () =>
			import("@/emails/templates/referral-converted"),
		"payout-failed": () => import("@/emails/templates/payout-failed"),

		// Organization
		"member-invited": () => import("@/emails/templates/member-invited"),
		"member-joined": () => import("@/emails/templates/member-joined"),
		"member-removed": () => import("@/emails/templates/member-removed"),

		// API Keys
		"api-key-created": () => import("@/emails/templates/api-key-created"),
		"api-key-revoked": () => import("@/emails/templates/api-key-revoked"),

		// Admin
		"admin-new-user": () => import("@/emails/templates/admin-new-user"),
		"admin-new-subscription": () =>
			import("@/emails/templates/admin-new-subscription"),
		"admin-nfse-error": () => import("@/emails/templates/admin-nfse-error"),

		// System
		"data-deletion-requested": () =>
			import("@/emails/templates/data-deletion-requested"),
		"contact-form-received": () =>
			import("@/emails/templates/contact-form-received"),
		"maintenance-scheduled": () =>
			import("@/emails/templates/maintenance-scheduled"),
	};

	const loader = templates[templateSlug];
	if (!loader) {
		throw new Error(`Email template not found: ${templateSlug}`);
	}

	const mod = await loader();
	const Component = mod.default;
	const element = Component(data) as React.ReactElement;
	const html = await render(element);

	// Extract subject from template data or use default
	const subject =
		(data.subject as string) || `Zé da API Manager — ${templateSlug}`;

	await sendEmail(to, subject, html);
}
