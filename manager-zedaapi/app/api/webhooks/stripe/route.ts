import { NextRequest, NextResponse } from "next/server";
import { constructWebhookEvent } from "@/lib/stripe";
import { db } from "@/lib/db";
import { handleStripePayment } from "@/server/services/billing-service";
import { syncSubscription } from "@/server/services/stripe-service";
import { createLogger } from "@/lib/logger";
import type Stripe from "stripe";

const log = createLogger("webhook:stripe");

export const dynamic = "force-dynamic";

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
		log.error("Signature verification failed", {
			error: err instanceof Error ? err.message : String(err),
		});
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
				log.info("Checkout session expired", { sessionId: session.id });
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
				log.info("Account updated", { accountId: account.id });
				break;
			}

			case "transfer.created":
			case "transfer.updated":
			case "transfer.reversed": {
				log.info("Transfer event", { eventType: event.type });
				break;
			}

			default:
				log.info("Unhandled event", { eventType: event.type });
		}

		// Mark event as processed
		await db.stripeWebhookEvent.update({
			where: { stripeEventId: event.id },
			data: { status: "processed", processedAt: new Date() },
		});

		return NextResponse.json({ received: true });
	} catch (err) {
		log.error("Handler error", {
			error: err instanceof Error ? err.message : String(err),
		});

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
		log.error("Missing metadata in checkout session");
		return;
	}

	const stripeSubscriptionId =
		typeof session.subscription === "string"
			? session.subscription
			: session.subscription?.id;

	if (!stripeSubscriptionId) {
		log.error("No subscription in checkout session");
		return;
	}

	// Check if subscription already exists
	const existingSub = await db.subscription.findUnique({
		where: { stripeSubscriptionId },
	});

	// Fetch real subscription data from Stripe for accurate period and status
	const { getStripe } = await import("@/lib/stripe");
	const stripe = getStripe();
	const stripeSub = await stripe.subscriptions.retrieve(stripeSubscriptionId);
	const firstItem = stripeSub.items.data[0];
	const periodStart = firstItem
		? new Date(firstItem.current_period_start * 1000)
		: new Date();
	const periodEnd = firstItem
		? new Date(firstItem.current_period_end * 1000)
		: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000);

	const stripeStatus = stripeSub.status;
	const localStatus =
		stripeStatus === "trialing"
			? "trialing"
			: stripeStatus === "active"
				? "active"
				: "incomplete";
	const trialEnd = stripeSub.trial_end
		? new Date(stripeSub.trial_end * 1000)
		: null;

	if (existingSub) {
		// Update existing
		await db.subscription.update({
			where: { id: existingSub.id },
			data: {
				status: localStatus,
				stripeStatus,
				planId,
				currentPeriodStart: periodStart,
				currentPeriodEnd: periodEnd,
				trialEnd,
			},
		});
	} else {
		// Create new subscription
		await db.subscription.create({
			data: {
				userId,
				planId,
				stripeSubscriptionId,
				stripeStatus,
				paymentMethod: "stripe",
				status: localStatus,
				currentPeriodStart: periodStart,
				currentPeriodEnd: periodEnd,
				trialEnd,
			},
		});
	}

	// Send appropriate email notification
	const user = await db.user.findUnique({
		where: { id: userId },
		select: { email: true, name: true },
	});
	const plan = await db.plan.findUnique({
		where: { id: planId },
		select: { name: true, price: true },
	});

	if (user?.email) {
		const price = new Intl.NumberFormat("pt-BR", {
			style: "currency",
			currency: "BRL",
		}).format(Number(plan?.price || 0));

		const emailTemplate =
			localStatus === "trialing"
				? "trial-started"
				: "subscription-upgraded";

		await sendEmailNotification(user.email, emailTemplate, {
			userName: user.name || "Usuário",
			userId,
			planName: plan?.name || "Plano",
			price,
			...(trialEnd && {
				trialEndDate: trialEnd.toLocaleDateString("pt-BR"),
			}),
			dashboardUrl: `${process.env.NEXT_PUBLIC_APP_URL}/painel`,
		});
	}
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

	const localStatus =
		subscription.status === "trialing"
			? "trialing"
			: subscription.status === "active"
				? "active"
				: "incomplete";
	const trialEnd = subscription.trial_end
		? new Date(subscription.trial_end * 1000)
		: null;

	const existing = await db.subscription.findUnique({
		where: { stripeSubscriptionId: subscription.id },
	});

	if (existing) {
		await db.subscription.update({
			where: { id: existing.id },
			data: {
				stripeStatus: subscription.status,
				status: localStatus,
				currentPeriodStart: periodStart,
				currentPeriodEnd: periodEnd,
				trialEnd,
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
			status: localStatus,
			currentPeriodStart: periodStart,
			currentPeriodEnd: periodEnd,
			trialEnd,
		},
	});

	// Send appropriate email
	const user = await db.user.findUnique({
		where: { id: userId },
		select: { email: true, name: true },
	});
	const plan = await db.plan.findUnique({
		where: { id: planId },
		select: { name: true, price: true, maxInstances: true },
	});

	if (user?.email) {
		const price = new Intl.NumberFormat("pt-BR", {
			style: "currency",
			currency: "BRL",
		}).format(Number(plan?.price || 0));

		const emailTemplate =
			localStatus === "trialing"
				? "trial-started"
				: "subscription-upgraded";

		await sendEmailNotification(user.email, emailTemplate, {
			userName: user.name || "Usuário",
			userId,
			planName: plan?.name || "Plano",
			price,
			maxInstances: plan?.maxInstances || 1,
			nextBillingDate: periodEnd.toLocaleDateString("pt-BR"),
			...(trialEnd && {
				trialEndDate: trialEnd.toLocaleDateString("pt-BR"),
			}),
			dashboardUrl: `${process.env.NEXT_PUBLIC_APP_URL}/painel`,
		});
	}
}

async function handleSubscriptionDeleted(subscription: Stripe.Subscription) {
	const localSub = await db.subscription.findUnique({
		where: { stripeSubscriptionId: subscription.id },
		select: {
			id: true,
			userId: true,
			instances: { select: { id: true } },
			user: { select: { email: true, name: true } },
			plan: { select: { name: true } },
		},
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

	// Send cancellation email
	if (localSub.user?.email) {
		await sendEmailNotification(
			localSub.user.email,
			"subscription-change",
			{
				userName: localSub.user.name || "Usuário",
				userId: localSub.userId,
				action: "cancel",
				oldPlan: localSub.plan?.name || "Plano",
				newPlan: "",
				effectiveDate: new Date().toLocaleDateString("pt-BR"),
				newPrice: "R$ 0,00",
				dashboardUrl: `${process.env.NEXT_PUBLIC_APP_URL}/assinaturas`,
			},
		);
	}
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

	// Queue NFS-e emission
	const localInvoice = await db.invoice.findUnique({
		where: { stripeInvoiceId: invoice.id },
		include: { user: { select: { email: true, name: true } } },
	});

	if (localInvoice) {
		const { enqueueNfseIssuance } = await import("@/lib/queue/producers");
		await enqueueNfseIssuance({
			invoiceId: localInvoice.id,
			action: "emit",
		});

		// Send invoice paid email
		if (localInvoice.user?.email) {
			const amount = new Intl.NumberFormat("pt-BR", {
				style: "currency",
				currency: "BRL",
			}).format(Number(localInvoice.amount));

			await sendEmailNotification(localInvoice.user.email, "invoice", {
				userName: localInvoice.user.name || "Usuário",
				userId: localInvoice.userId,
				invoiceId: localInvoice.id,
				amount,
				paidAt: new Date().toLocaleDateString("pt-BR"),
				paymentMethod: "stripe",
				pdfUrl: invoice.hosted_invoice_url || "",
				dashboardUrl: `${process.env.NEXT_PUBLIC_APP_URL}/faturamento`,
			});
		}
	}
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

	// Send payment failed notification
	if (invoice.id) {
		const localInvoice = await db.invoice.findUnique({
			where: { stripeInvoiceId: invoice.id },
			include: { user: { select: { email: true, name: true } } },
		});

		if (localInvoice?.user?.email) {
			const amount = new Intl.NumberFormat("pt-BR", {
				style: "currency",
				currency: "BRL",
			}).format(Number(localInvoice.amount));

			await sendEmailNotification(
				localInvoice.user.email,
				"payment-failed",
				{
					userName: localInvoice.user.name || "Usuário",
					userId: localInvoice.userId,
					invoiceId: localInvoice.id,
					amount,
					dueDate:
						localInvoice.dueDate?.toLocaleDateString("pt-BR") || "",
					retryUrl: `${process.env.NEXT_PUBLIC_APP_URL}/faturamento`,
				},
			);
		}
	}
}

async function handleChargeRefunded(charge: Stripe.Charge) {
	log.info("Charge refunded", {
		chargeId: charge.id,
		amountRefunded: charge.amount_refunded,
	});

	// Send refund notification
	if (charge.amount_refunded && charge.invoice) {
		const stripeInvoiceId =
			typeof charge.invoice === "string"
				? charge.invoice
				: charge.invoice.id;

		const localInvoice = await db.invoice.findUnique({
			where: { stripeInvoiceId },
			include: {
				user: { select: { email: true, name: true, id: true } },
			},
		});

		if (localInvoice?.user?.email) {
			const amount = new Intl.NumberFormat("pt-BR", {
				style: "currency",
				currency: "BRL",
			}).format(charge.amount_refunded / 100);

			await sendEmailNotification(
				localInvoice.user.email,
				"charge-refunded",
				{
					userName: localInvoice.user.name || "Usuário",
					userId: localInvoice.user.id,
					amount,
					chargeId: charge.id,
					dashboardUrl: `${process.env.NEXT_PUBLIC_APP_URL}/faturamento`,
				},
			);
		}
	}
}
