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
		welcome: () => import("@/emails/templates/welcome"),
		invoice: () => import("@/emails/templates/invoice"),
		"subscription-change": () =>
			import("@/emails/templates/subscription-change"),
		"nfse-issued": () => import("@/emails/templates/nfse-issued"),
		"payment-failed": () => import("@/emails/templates/payment-failed"),
		"waitlist-approved": () =>
			import("@/emails/templates/waitlist-approved"),
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
