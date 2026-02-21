import type { Job } from "bullmq";
import type { StripeWebhookJobData } from "@/lib/queue/types";
import { createLogger } from "@/lib/queue/logger";

const log = createLogger("processor:stripe-webhook");

async function sendEmailNotification(
	to: string,
	template: string,
	data: Record<string, unknown>,
): Promise<void> {
	try {
		const { enqueueEmailSending } = await import("@/lib/queue/producers");
		await enqueueEmailSending({ to, template, data });
	} catch (error) {
		log.error("Failed to enqueue email notification", {
			template,
			to,
			error: error instanceof Error ? error.message : "Unknown error",
		});
	}
}

export async function processStripeWebhookJob(
	job: Job<StripeWebhookJobData>,
): Promise<void> {
	const { eventId, type, payload } = job.data;

	log.info("Processing Stripe webhook", {
		jobId: job.id,
		eventId,
		type,
	});

	const done = log.timer("Stripe webhook processing", { eventId, type });

	const { db } = await import("@/lib/db");

	// Idempotency check
	const existing = await db.stripeWebhookEvent.findUnique({
		where: { stripeEventId: eventId },
	});

	if (existing?.status === "processed") {
		log.info("Stripe event already processed, skipping", { eventId });
		return;
	}

	// Mark as processing
	await db.stripeWebhookEvent.upsert({
		where: { stripeEventId: eventId },
		create: {
			stripeEventId: eventId,
			type,
			status: "processing",
			payload: payload as object,
		},
		update: { status: "processing" },
	});

	try {
		switch (type) {
			// Subscription lifecycle
			case "customer.subscription.created":
			case "customer.subscription.updated":
				await handleSubscriptionChange(payload);
				break;

			case "customer.subscription.deleted":
				await handleSubscriptionDeleted(payload);
				break;

			// Invoice events
			case "invoice.paid":
				await handleInvoicePaid(payload);
				break;

			case "invoice.payment_failed":
				await handleInvoicePaymentFailed(payload);
				break;

			case "invoice.created":
				await handleInvoiceCreated(payload);
				break;

			// Checkout
			case "checkout.session.completed":
				await handleCheckoutCompleted(payload);
				break;

			// Refunds
			case "charge.refunded":
				await handleChargeRefunded(payload);
				break;

			default:
				log.debug("Unhandled Stripe event type", { type, eventId });
		}

		await db.stripeWebhookEvent.update({
			where: { stripeEventId: eventId },
			data: { status: "processed", processedAt: new Date() },
		});

		done();
	} catch (error) {
		const errorMessage =
			error instanceof Error ? error.message : "Unknown error";

		await db.stripeWebhookEvent.update({
			where: { stripeEventId: eventId },
			data: { status: "failed", error: errorMessage },
		});

		log.error("Stripe webhook processing failed", {
			eventId,
			type,
			error: errorMessage,
		});

		throw error;
	}
}

async function handleSubscriptionChange(
	payload: Record<string, unknown>,
): Promise<void> {
	const { db } = await import("@/lib/db");
	const subscription = payload as Record<string, unknown>;
	const stripeSubId = subscription.id as string;
	const status = subscription.status as string;

	const sub = await db.subscription.findUnique({
		where: { stripeSubscriptionId: stripeSubId },
		include: {
			user: { select: { email: true, name: true } },
			plan: { select: { name: true, price: true } },
		},
	});

	if (!sub) {
		log.warn("Subscription not found in DB", {
			stripeSubscriptionId: stripeSubId,
		});
		return;
	}

	const previousStatus = sub.stripeStatus;
	const periodEnd = subscription.current_period_end as number;
	const periodStart = subscription.current_period_start as number;
	const cancelAtEnd = subscription.cancel_at_period_end as boolean;

	await db.subscription.update({
		where: { stripeSubscriptionId: stripeSubId },
		data: {
			stripeStatus: status,
			status:
				status === "active" || status === "trialing"
					? "active"
					: "inactive",
			currentPeriodStart: new Date(periodStart * 1000),
			currentPeriodEnd: new Date(periodEnd * 1000),
			cancelAtPeriodEnd: cancelAtEnd ?? false,
		},
	});

	log.info("Subscription updated", {
		stripeSubscriptionId: stripeSubId,
		status,
	});

	// Send email notification based on status transition
	if (sub.user?.email) {
		const emailData = {
			userName: sub.user.name || "Usuário",
			userId: sub.userId,
			planName: sub.plan?.name || "Plano",
			price: new Intl.NumberFormat("pt-BR", {
				style: "currency",
				currency: "BRL",
			}).format(Number(sub.plan?.price || 0)),
			nextBillingDate: new Date(periodEnd * 1000).toLocaleDateString(
				"pt-BR",
			),
			dashboardUrl: `${process.env.NEXT_PUBLIC_APP_URL}/assinaturas`,
		};

		if (status === "active" && previousStatus !== "active") {
			await sendEmailNotification(
				sub.user.email,
				"subscription-renewed",
				emailData,
			);
		} else if (cancelAtEnd && !sub.cancelAtPeriodEnd) {
			await sendEmailNotification(sub.user.email, "subscription-change", {
				...emailData,
				action: "cancel",
				oldPlan: sub.plan?.name || "",
				newPlan: "",
				effectiveDate: new Date(periodEnd * 1000).toLocaleDateString(
					"pt-BR",
				),
				newPrice: "R$ 0,00",
			});
		}
	}
}

async function handleSubscriptionDeleted(
	payload: Record<string, unknown>,
): Promise<void> {
	const { db } = await import("@/lib/db");
	const stripeSubId = payload.id as string;

	const sub = await db.subscription.findUnique({
		where: { stripeSubscriptionId: stripeSubId },
		include: {
			user: { select: { email: true, name: true } },
			plan: { select: { name: true } },
		},
	});

	await db.subscription.updateMany({
		where: { stripeSubscriptionId: stripeSubId },
		data: {
			stripeStatus: "canceled",
			status: "canceled",
			canceledAt: new Date(),
		},
	});

	log.info("Subscription canceled", { stripeSubscriptionId: stripeSubId });

	if (sub?.user?.email) {
		await sendEmailNotification(sub.user.email, "subscription-change", {
			userName: sub.user.name || "Usuário",
			userId: sub.userId,
			action: "cancel",
			oldPlan: sub.plan?.name || "Plano",
			newPlan: "",
			effectiveDate: new Date().toLocaleDateString("pt-BR"),
			newPrice: "R$ 0,00",
			dashboardUrl: `${process.env.NEXT_PUBLIC_APP_URL}/assinaturas`,
		});
	}
}

async function handleInvoicePaid(
	payload: Record<string, unknown>,
): Promise<void> {
	const { db } = await import("@/lib/db");
	const stripeInvoiceId = payload.id as string;

	const invoice = await db.invoice.findUnique({
		where: { stripeInvoiceId },
		include: {
			user: { select: { email: true, name: true } },
		},
	});

	if (!invoice) {
		log.debug("Invoice not found for paid event", { stripeInvoiceId });
		return;
	}

	await db.invoice.update({
		where: { stripeInvoiceId },
		data: {
			status: "paid",
			paidAt: new Date(),
		},
	});

	// Trigger NFS-e issuance
	const { enqueueNfseIssuance } = await import("@/lib/queue/producers");
	await enqueueNfseIssuance({
		invoiceId: invoice.id,
		action: "emit",
	});

	log.info("Invoice marked as paid, NFS-e enqueued", {
		invoiceId: invoice.id,
		stripeInvoiceId,
	});

	// Send invoice paid email
	if (invoice.user?.email) {
		const amount = new Intl.NumberFormat("pt-BR", {
			style: "currency",
			currency: "BRL",
		}).format(Number(invoice.amount));

		await sendEmailNotification(invoice.user.email, "invoice", {
			userName: invoice.user.name || "Usuário",
			userId: invoice.userId,
			invoiceId: invoice.id,
			amount,
			paidAt: new Date().toLocaleDateString("pt-BR"),
			paymentMethod: invoice.paymentMethod || "stripe",
			pdfUrl: (payload.hosted_invoice_url as string) || "",
			dashboardUrl: `${process.env.NEXT_PUBLIC_APP_URL}/faturamento`,
		});
	}
}

async function handleInvoicePaymentFailed(
	payload: Record<string, unknown>,
): Promise<void> {
	const { db } = await import("@/lib/db");
	const stripeInvoiceId = payload.id as string;

	const invoice = await db.invoice.findUnique({
		where: { stripeInvoiceId },
		include: {
			user: { select: { email: true, name: true } },
		},
	});

	await db.invoice.updateMany({
		where: { stripeInvoiceId },
		data: { status: "payment_failed" },
	});

	log.warn("Invoice payment failed", { stripeInvoiceId });

	// Send payment failed email
	if (invoice?.user?.email) {
		const amount = new Intl.NumberFormat("pt-BR", {
			style: "currency",
			currency: "BRL",
		}).format(Number(invoice.amount));

		await sendEmailNotification(invoice.user.email, "payment-failed", {
			userName: invoice.user.name || "Usuário",
			userId: invoice.userId,
			invoiceId: invoice.id,
			amount,
			dueDate: invoice.dueDate?.toLocaleDateString("pt-BR") || "",
			retryUrl: `${process.env.NEXT_PUBLIC_APP_URL}/faturamento`,
		});
	}
}

async function handleInvoiceCreated(
	payload: Record<string, unknown>,
): Promise<void> {
	const { db } = await import("@/lib/db");
	const stripeInvoiceId = payload.id as string;
	const subscriptionId = payload.subscription as string | undefined;
	const amountDue = payload.amount_due as number;
	const currency = (payload.currency as string)?.toUpperCase() ?? "BRL";
	const customerId = payload.customer as string;

	if (!subscriptionId) return;

	const sub = await db.subscription.findUnique({
		where: { stripeSubscriptionId: subscriptionId },
		select: { id: true, userId: true },
	});

	if (!sub) {
		log.debug("Subscription not found for invoice", {
			stripeInvoiceId,
			subscriptionId,
		});
		return;
	}

	const invoice = await db.invoice.upsert({
		where: { stripeInvoiceId },
		create: {
			userId: sub.userId,
			subscriptionId: sub.id,
			stripeInvoiceId,
			amount: amountDue / 100,
			currency,
			status: "draft",
		},
		update: {
			amount: amountDue / 100,
			currency,
		},
	});

	log.info("Invoice created/updated", { stripeInvoiceId, customerId });

	// Send invoice created email
	const user = await db.user.findUnique({
		where: { id: sub.userId },
		select: { email: true, name: true },
	});

	if (user?.email) {
		const amount = new Intl.NumberFormat("pt-BR", {
			style: "currency",
			currency: "BRL",
		}).format(amountDue / 100);

		await sendEmailNotification(user.email, "invoice-created", {
			userName: user.name || "Usuário",
			userId: sub.userId,
			invoiceId: invoice.id,
			amount,
			currency,
			dashboardUrl: `${process.env.NEXT_PUBLIC_APP_URL}/faturamento`,
		});
	}
}

async function handleCheckoutCompleted(
	payload: Record<string, unknown>,
): Promise<void> {
	const metadata = payload.metadata as Record<string, string> | undefined;
	const userId = metadata?.userId;

	log.info("Checkout session completed", {
		sessionId: payload.id,
		customerId: payload.customer,
		userId,
	});

	if (userId) {
		const { db } = await import("@/lib/db");
		const user = await db.user.findUnique({
			where: { id: userId },
			select: { email: true, name: true },
		});

		if (user?.email) {
			await sendEmailNotification(user.email, "subscription-upgraded", {
				userName: user.name || "Usuário",
				userId,
				dashboardUrl: `${process.env.NEXT_PUBLIC_APP_URL}/painel`,
			});
		}
	}
}

async function handleChargeRefunded(
	payload: Record<string, unknown>,
): Promise<void> {
	const amountRefunded = payload.amount_refunded as number;
	const customerId = payload.customer as string | undefined;

	log.info("Charge refunded", {
		chargeId: payload.id,
		amount: amountRefunded,
	});

	if (customerId) {
		const { db } = await import("@/lib/db");

		// Find user by Stripe customer ID via subscription
		const sub = await db.subscription.findFirst({
			where: {
				stripeSubscriptionId: { not: null },
			},
			include: {
				user: { select: { email: true, name: true, id: true } },
			},
			orderBy: { createdAt: "desc" },
		});

		if (sub?.user?.email && amountRefunded) {
			const amount = new Intl.NumberFormat("pt-BR", {
				style: "currency",
				currency: "BRL",
			}).format(amountRefunded / 100);

			await sendEmailNotification(sub.user.email, "charge-refunded", {
				userName: sub.user.name || "Usuário",
				userId: sub.user.id,
				amount,
				chargeId: payload.id as string,
				dashboardUrl: `${process.env.NEXT_PUBLIC_APP_URL}/faturamento`,
			});
		}
	}
}
