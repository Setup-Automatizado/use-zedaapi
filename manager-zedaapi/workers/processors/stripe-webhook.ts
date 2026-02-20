import type { Job } from "bullmq";
import type { StripeWebhookJobData } from "@/lib/queue/types";
import { createLogger } from "@/lib/queue/logger";

const log = createLogger("processor:stripe-webhook");

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
	});

	if (!sub) {
		log.warn("Subscription not found in DB", {
			stripeSubscriptionId: stripeSubId,
		});
		return;
	}

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
}

async function handleSubscriptionDeleted(
	payload: Record<string, unknown>,
): Promise<void> {
	const { db } = await import("@/lib/db");
	const stripeSubId = payload.id as string;

	await db.subscription.updateMany({
		where: { stripeSubscriptionId: stripeSubId },
		data: {
			stripeStatus: "canceled",
			status: "canceled",
			canceledAt: new Date(),
		},
	});

	log.info("Subscription canceled", { stripeSubscriptionId: stripeSubId });
}

async function handleInvoicePaid(
	payload: Record<string, unknown>,
): Promise<void> {
	const { db } = await import("@/lib/db");
	const stripeInvoiceId = payload.id as string;

	const invoice = await db.invoice.findUnique({
		where: { stripeInvoiceId },
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
}

async function handleInvoicePaymentFailed(
	payload: Record<string, unknown>,
): Promise<void> {
	const { db } = await import("@/lib/db");
	const stripeInvoiceId = payload.id as string;

	await db.invoice.updateMany({
		where: { stripeInvoiceId },
		data: { status: "payment_failed" },
	});

	log.warn("Invoice payment failed", { stripeInvoiceId });
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

	await db.invoice.upsert({
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
}

async function handleCheckoutCompleted(
	payload: Record<string, unknown>,
): Promise<void> {
	log.info("Checkout session completed", {
		sessionId: payload.id,
		customerId: payload.customer,
	});
}

async function handleChargeRefunded(
	payload: Record<string, unknown>,
): Promise<void> {
	log.info("Charge refunded", {
		chargeId: payload.id,
		amount: payload.amount_refunded,
	});
}
