import { NextRequest, NextResponse } from "next/server";
import { constructWebhookEvent, stripe } from "@/lib/stripe";
import { db } from "@/lib/db";
import {
	handleStripePayment,
	handleSicrediPayment,
} from "@/server/services/billing-service";
import { syncSubscription } from "@/server/services/stripe-service";
import type Stripe from "stripe";

export const dynamic = "force-dynamic";

export async function POST(request: NextRequest) {
	const body = await request.text();
	const signature = request.headers.get("stripe-signature");

	if (!signature) {
		return NextResponse.json(
			{ error: "Missing stripe-signature header" },
			{ status: 400 },
		);
	}

	let event: Stripe.Event;

	try {
		event = await constructWebhookEvent(body, signature);
	} catch (err) {
		console.error("[stripe-webhook] Signature verification failed:", err);
		return NextResponse.json(
			{ error: "Webhook signature verification failed" },
			{ status: 400 },
		);
	}

	// Idempotency: check if we already processed this event
	const existing = await db.stripeWebhookEvent.findUnique({
		where: { stripeEventId: event.id },
	});

	if (existing?.status === "processed") {
		return NextResponse.json({
			received: true,
			status: "already_processed",
		});
	}

	// Record the event
	await db.stripeWebhookEvent.upsert({
		where: { stripeEventId: event.id },
		create: {
			stripeEventId: event.id,
			type: event.type,
			status: "processing",
			payload: event.data.object as object,
		},
		update: {
			status: "processing",
		},
	});

	try {
		switch (event.type) {
			// === Checkout Sessions ===
			case "checkout.session.completed": {
				const session = event.data.object as Stripe.Checkout.Session;
				await handleCheckoutCompleted(session);
				break;
			}

			case "checkout.session.expired": {
				const session = event.data.object as Stripe.Checkout.Session;
				console.log(
					`[stripe-webhook] Checkout session expired: ${session.id}`,
				);
				break;
			}

			case "checkout.session.async_payment_succeeded": {
				const session = event.data.object as Stripe.Checkout.Session;
				await handleCheckoutCompleted(session);
				break;
			}

			// === Subscriptions ===
			case "customer.subscription.created": {
				const sub = event.data.object as Stripe.Subscription;
				await handleSubscriptionCreated(sub);
				break;
			}

			case "customer.subscription.updated": {
				const sub = event.data.object as Stripe.Subscription;
				await syncSubscription(sub.id);
				break;
			}

			case "customer.subscription.deleted": {
				const sub = event.data.object as Stripe.Subscription;
				await handleSubscriptionDeleted(sub);
				break;
			}

			// === Invoices ===
			case "invoice.paid": {
				const invoice = event.data.object as Stripe.Invoice;
				await handleInvoicePaid(invoice);
				break;
			}

			case "invoice.payment_failed": {
				const invoice = event.data.object as Stripe.Invoice;
				await handleInvoicePaymentFailed(invoice);
				break;
			}

			// === Refunds ===
			case "charge.refunded": {
				const charge = event.data.object as Stripe.Charge;
				await handleChargeRefunded(charge);
				break;
			}

			// === Connect ===
			case "account.updated": {
				const account = event.data.object as Stripe.Account;
				console.log(`[stripe-webhook] Account updated: ${account.id}`);
				break;
			}

			case "transfer.created":
			case "transfer.updated":
			case "transfer.reversed": {
				console.log(`[stripe-webhook] Transfer event: ${event.type}`);
				break;
			}

			default:
				console.log(`[stripe-webhook] Unhandled event: ${event.type}`);
		}

		// Mark event as processed
		await db.stripeWebhookEvent.update({
			where: { stripeEventId: event.id },
			data: { status: "processed", processedAt: new Date() },
		});

		return NextResponse.json({ received: true });
	} catch (err) {
		console.error("[stripe-webhook] Handler error:", err);

		await db.stripeWebhookEvent.update({
			where: { stripeEventId: event.id },
			data: {
				status: "error",
				error: err instanceof Error ? err.message : "Unknown error",
			},
		});

		return NextResponse.json(
			{ error: "Webhook handler failed" },
			{ status: 500 },
		);
	}
}

async function handleCheckoutCompleted(session: Stripe.Checkout.Session) {
	const userId = session.metadata?.userId;
	const planId = session.metadata?.planId;

	if (!userId || !planId) {
		console.error("[stripe-webhook] Missing metadata in checkout session");
		return;
	}

	const stripeSubscriptionId =
		typeof session.subscription === "string"
			? session.subscription
			: session.subscription?.id;

	if (!stripeSubscriptionId) {
		console.error("[stripe-webhook] No subscription in checkout session");
		return;
	}

	// Check if subscription already exists
	const existingSub = await db.subscription.findUnique({
		where: { stripeSubscriptionId },
	});

	if (existingSub) {
		// Update existing
		await db.subscription.update({
			where: { id: existingSub.id },
			data: {
				status: "active",
				stripeStatus: "active",
				planId,
			},
		});
		return;
	}

	// Create new subscription
	const now = new Date();
	const periodEnd = new Date(now);
	periodEnd.setMonth(periodEnd.getMonth() + 1);

	await db.subscription.create({
		data: {
			userId,
			planId,
			stripeSubscriptionId,
			stripeStatus: "active",
			paymentMethod: "stripe",
			status: "active",
			currentPeriodStart: now,
			currentPeriodEnd: periodEnd,
		},
	});
}

async function handleSubscriptionCreated(subscription: Stripe.Subscription) {
	const userId = subscription.metadata?.userId;
	if (!userId) return;

	const planId = subscription.metadata?.planId;
	if (!planId) return;

	// In Stripe v18, period dates are on the subscription item level
	const firstItem = subscription.items.data[0];
	const periodStart = firstItem
		? new Date(firstItem.current_period_start * 1000)
		: new Date();
	const periodEnd = firstItem
		? new Date(firstItem.current_period_end * 1000)
		: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000);

	const existing = await db.subscription.findUnique({
		where: { stripeSubscriptionId: subscription.id },
	});

	if (existing) {
		await db.subscription.update({
			where: { id: existing.id },
			data: {
				stripeStatus: subscription.status,
				status:
					subscription.status === "active" ? "active" : "incomplete",
				currentPeriodStart: periodStart,
				currentPeriodEnd: periodEnd,
			},
		});
		return;
	}

	await db.subscription.create({
		data: {
			userId,
			planId,
			stripeSubscriptionId: subscription.id,
			stripeStatus: subscription.status,
			paymentMethod: "stripe",
			status: subscription.status === "active" ? "active" : "incomplete",
			currentPeriodStart: periodStart,
			currentPeriodEnd: periodEnd,
		},
	});
}

async function handleSubscriptionDeleted(subscription: Stripe.Subscription) {
	const localSub = await db.subscription.findUnique({
		where: { stripeSubscriptionId: subscription.id },
		select: { id: true, instances: { select: { id: true } } },
	});

	if (!localSub) return;

	await db.subscription.update({
		where: { id: localSub.id },
		data: {
			status: "canceled",
			stripeStatus: "canceled",
			canceledAt: new Date(),
		},
	});

	// TODO: Optionally disconnect/deactivate instances via ZedaAPI
	// for (const instance of localSub.instances) {
	//   await deactivateInstance(instance.id);
	// }
}

async function handleInvoicePaid(invoice: Stripe.Invoice) {
	// In Stripe v18, subscription is under parent.subscription_details
	const subDetails = invoice.parent?.subscription_details;
	if (!subDetails?.subscription) return;

	const subscriptionId =
		typeof subDetails.subscription === "string"
			? subDetails.subscription
			: subDetails.subscription.id;

	if (!invoice.id) return;

	await handleStripePayment(
		subscriptionId,
		invoice.id,
		invoice.amount_paid,
		invoice.currency,
		invoice.hosted_invoice_url || null,
	);

	// Sync subscription state
	await syncSubscription(subscriptionId);

	// TODO: Queue NFSe emission via BullMQ
	// await nfseQueue.add('emit-nfse', { invoiceId: invoice.id });
}

async function handleInvoicePaymentFailed(invoice: Stripe.Invoice) {
	// In Stripe v18, subscription is under parent.subscription_details
	const subDetails = invoice.parent?.subscription_details;
	if (!subDetails?.subscription) return;

	const subscriptionId =
		typeof subDetails.subscription === "string"
			? subDetails.subscription
			: subDetails.subscription.id;

	await db.subscription.updateMany({
		where: { stripeSubscriptionId: subscriptionId },
		data: {
			status: "past_due",
			stripeStatus: "past_due",
		},
	});

	// TODO: Send payment failed notification
}

async function handleChargeRefunded(charge: Stripe.Charge) {
	console.log(
		`[stripe-webhook] Charge refunded: ${charge.id}, amount: ${charge.amount_refunded}`,
	);

	// TODO: Handle refund logic (reverse commission, update invoice status)
}
