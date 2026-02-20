"use server";

import { db } from "@/lib/db";
import { sendTemplateEmail, sendEmail } from "@/lib/email";

// =============================================================================
// Notification Service
// =============================================================================

type NotificationType =
	| "welcome"
	| "invoice"
	| "subscription_change"
	| "nfse_issued"
	| "payment_failed"
	| "waitlist_approved"
	| "security"
	| "general";

type NotificationChannel = "email" | "whatsapp" | "in_app";

/**
 * Check user notification preferences before sending.
 * Returns true if the user has the given channel enabled.
 */
async function checkPreferences(
	userId: string,
	channel: NotificationChannel,
	type: NotificationType,
): Promise<boolean> {
	const prefs = await db.notificationPreference.findUnique({
		where: { userId },
	});

	if (!prefs) return true; // Default to enabled

	if (channel === "email" && !prefs.emailEnabled) return false;
	if (channel === "whatsapp" && !prefs.whatsappEnabled) return false;

	// Security and billing alerts bypass marketing preference
	if (
		type !== "security" &&
		type !== "payment_failed" &&
		type !== "invoice"
	) {
		if (!prefs.marketingEnabled && type === "general") return false;
	}

	// Billing alerts check
	if (
		(type === "invoice" || type === "payment_failed") &&
		!prefs.billingAlerts
	) {
		return false;
	}

	// Security alerts always sent
	if (type === "security" && !prefs.securityAlerts) return false;

	return true;
}

/**
 * Core notification dispatcher. Records notification and sends via the specified channel.
 */
export async function sendNotification(
	userId: string,
	type: NotificationType,
	channel: NotificationChannel,
	data: {
		subject?: string;
		body: string;
		templateSlug?: string;
		templateData?: Record<string, unknown>;
		metadata?: Record<string, unknown>;
	},
): Promise<void> {
	const allowed = await checkPreferences(userId, channel, type);
	if (!allowed) return;

	// Record notification in DB
	const notification = await db.notification.create({
		data: {
			userId,
			type,
			channel,
			subject: data.subject,
			body: data.body,
			status: "pending",
			metadata: (data.metadata ?? undefined) as object | undefined,
		},
	});

	try {
		if (channel === "email") {
			const user = await db.user.findUnique({
				where: { id: userId },
				select: { email: true, name: true },
			});

			if (!user) return;

			if (data.templateSlug) {
				await sendTemplateEmail(user.email, data.templateSlug, {
					...(data.templateData || {}),
					userName: user.name,
					subject: data.subject,
				});
			} else {
				await sendEmail(
					user.email,
					data.subject || "Zé da API Manager",
					data.body,
				);
			}
		}

		// Mark as sent
		await db.notification.update({
			where: { id: notification.id },
			data: { status: "sent", sentAt: new Date() },
		});
	} catch (error) {
		// Mark as failed
		await db.notification.update({
			where: { id: notification.id },
			data: {
				status: "failed",
				error: error instanceof Error ? error.message : "Unknown error",
			},
		});
	}
}

/**
 * Send welcome email to a newly registered user.
 */
export async function sendWelcomeEmail(userId: string): Promise<void> {
	const user = await db.user.findUnique({
		where: { id: userId },
		select: { name: true },
	});

	await sendNotification(userId, "welcome", "email", {
		subject: "Bem-vindo ao Zé da API Manager!",
		body: `Olá ${user?.name}, sua conta foi criada com sucesso.`,
		templateSlug: "welcome",
		templateData: {
			dashboardUrl: `${process.env.NEXT_PUBLIC_APP_URL}/dashboard`,
		},
	});
}

/**
 * Send invoice email after successful payment.
 */
export async function sendInvoiceEmail(
	userId: string,
	invoiceId: string,
): Promise<void> {
	const invoice = await db.invoice.findUnique({
		where: { id: invoiceId },
		select: {
			id: true,
			amount: true,
			paidAt: true,
			paymentMethod: true,
			pdfUrl: true,
			dueDate: true,
		},
	});

	if (!invoice) return;

	const amount = new Intl.NumberFormat("pt-BR", {
		style: "currency",
		currency: "BRL",
	}).format(Number(invoice.amount));

	await sendNotification(userId, "invoice", "email", {
		subject: `Fatura Zé da API — ${amount}`,
		body: `Sua fatura de ${amount} foi registrada.`,
		templateSlug: "invoice",
		templateData: {
			invoiceId: invoice.id,
			amount,
			paidAt: invoice.paidAt?.toISOString(),
			paymentMethod: invoice.paymentMethod,
			pdfUrl: invoice.pdfUrl,
			dueDate: invoice.dueDate?.toISOString(),
		},
		metadata: { invoiceId },
	});
}

/**
 * Send notification when subscription changes (upgrade, downgrade, cancel).
 */
export async function sendSubscriptionChangeEmail(
	userId: string,
	change: {
		oldPlan: string;
		newPlan: string;
		effectiveDate: string;
		newPrice: string;
		action: "upgrade" | "downgrade" | "cancel" | "resume";
	},
): Promise<void> {
	const actionLabels: Record<string, string> = {
		upgrade: "Upgrade de plano",
		downgrade: "Downgrade de plano",
		cancel: "Cancelamento de assinatura",
		resume: "Reativação de assinatura",
	};

	await sendNotification(userId, "subscription_change", "email", {
		subject: `Zé da API — ${actionLabels[change.action]}`,
		body: `Seu plano foi alterado de ${change.oldPlan} para ${change.newPlan}.`,
		templateSlug: "subscription-change",
		templateData: change,
		metadata: change,
	});
}

/**
 * Send notification when NFS-e is issued for an invoice.
 */
export async function sendNfseIssuedEmail(
	userId: string,
	invoiceId: string,
): Promise<void> {
	const invoice = await db.invoice.findUnique({
		where: { id: invoiceId },
		select: {
			nfseNumber: true,
			amount: true,
			nfsePdfUrl: true,
		},
	});

	if (!invoice || !invoice.nfseNumber) return;

	const amount = new Intl.NumberFormat("pt-BR", {
		style: "currency",
		currency: "BRL",
	}).format(Number(invoice.amount));

	await sendNotification(userId, "nfse_issued", "email", {
		subject: `NFS-e emitida — ${invoice.nfseNumber}`,
		body: `Sua NFS-e ${invoice.nfseNumber} foi emitida no valor de ${amount}.`,
		templateSlug: "nfse-issued",
		templateData: {
			nfseNumber: invoice.nfseNumber,
			amount,
			pdfUrl: invoice.nfsePdfUrl,
		},
		metadata: { invoiceId, nfseNumber: invoice.nfseNumber },
	});
}

/**
 * Send notification when a payment fails.
 */
export async function sendPaymentFailedEmail(
	userId: string,
	invoiceId: string,
): Promise<void> {
	const invoice = await db.invoice.findUnique({
		where: { id: invoiceId },
		select: {
			id: true,
			amount: true,
			dueDate: true,
		},
	});

	if (!invoice) return;

	const amount = new Intl.NumberFormat("pt-BR", {
		style: "currency",
		currency: "BRL",
	}).format(Number(invoice.amount));

	await sendNotification(userId, "payment_failed", "email", {
		subject: "Falha no pagamento — Zé da API",
		body: `O pagamento de ${amount} não foi processado.`,
		templateSlug: "payment-failed",
		templateData: {
			invoiceId: invoice.id,
			amount,
			dueDate: invoice.dueDate?.toISOString(),
			retryUrl: `${process.env.NEXT_PUBLIC_APP_URL}/billing`,
		},
		metadata: { invoiceId },
	});
}

/**
 * Send notification when a waitlist entry is approved.
 * This does not require userId — uses email directly.
 */
export async function sendWaitlistApprovedEmail(
	email: string,
	name: string,
): Promise<void> {
	const { sendTemplateEmail: send } = await import("@/lib/email");

	try {
		await send(email, "waitlist-approved", {
			userName: name,
			subject: "Sua conta foi aprovada! — Zé da API Manager",
			signUpUrl: `${process.env.NEXT_PUBLIC_APP_URL}/sign-up`,
		});
	} catch (error) {
		console.error(
			"[notification] Failed to send waitlist approved email:",
			error,
		);
	}
}
