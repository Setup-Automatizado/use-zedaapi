"use server";

import { getStripe } from "@/lib/stripe";
import { db } from "@/lib/db";
import type Stripe from "stripe";
import { createLogger } from "@/lib/logger";

const log = createLogger("service:stripe");

/**
 * Get or create a Stripe customer for the given user.
 * Links stripe customer ID to user record.
 */
export async function createOrGetCustomer(userId: string): Promise<string> {
	const stripe = getStripe();
	const user = await db.user.findUnique({
		where: { id: userId },
		select: {
			stripeCustomerId: true,
			email: true,
			name: true,
			cpfCnpj: true,
		},
	});

	if (!user) {
		throw new Error("User not found");
	}

	if (user.stripeCustomerId) {
		return user.stripeCustomerId;
	}

	const customer = await stripe.customers.create({
		email: user.email,
		name: user.name || undefined,
		metadata: {
			userId,
			cpfCnpj: user.cpfCnpj || "",
		},
	});

	await db.user.update({
		where: { id: userId },
		data: { stripeCustomerId: customer.id },
	});

	return customer.id;
}

/**
 * Create a Stripe Checkout Session for a subscription.
 */
export async function createCheckoutSession(
	userId: string,
	planId: string,
	paymentMethod: "stripe" | "pix" | "boleto" = "stripe",
): Promise<Stripe.Checkout.Session> {
	const stripe = getStripe();
	const plan = await db.plan.findUnique({
		where: { id: planId },
		select: { stripePriceId: true, name: true, slug: true },
	});

	if (!plan?.stripePriceId) {
		throw new Error("Plan not found or missing Stripe price ID");
	}

	const customerId = await createOrGetCustomer(userId);

	const appUrl = process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000";

	const paymentMethodTypes: Stripe.Checkout.SessionCreateParams.PaymentMethodType[] =
		paymentMethod === "pix"
			? ["card"]
			: paymentMethod === "boleto"
				? ["card", "boleto"]
				: ["card"];

	const session = await stripe.checkout.sessions.create({
		customer: customerId,
		mode: "subscription",
		payment_method_types: paymentMethodTypes,
		line_items: [
			{
				price: plan.stripePriceId,
				quantity: 1,
			},
		],
		success_url: `${appUrl}/assinaturas?success=true&session_id={CHECKOUT_SESSION_ID}`,
		cancel_url: `${appUrl}/assinaturas/planos?canceled=true`,
		metadata: {
			userId,
			planId,
			planSlug: plan.slug,
		},
		subscription_data: {
			metadata: {
				userId,
				planId,
				planSlug: plan.slug,
			},
		},
		allow_promotion_codes: true,
		billing_address_collection: "required",
		locale: "pt-BR",
	});

	return session;
}

/**
 * Create a Stripe billing portal session for the customer.
 */
export async function createBillingPortalSession(
	customerId: string,
): Promise<string> {
	const stripe = getStripe();
	const appUrl = process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000";

	const session = await stripe.billingPortal.sessions.create({
		customer: customerId,
		return_url: `${appUrl}/faturamento`,
	});

	return session.url;
}

/**
 * Sync a Stripe subscription state to the local database.
 */
export async function syncSubscription(
	stripeSubscriptionId: string,
): Promise<void> {
	const stripe = getStripe();
	const stripeSub = await stripe.subscriptions.retrieve(stripeSubscriptionId);

	const subscription = await db.subscription.findUnique({
		where: { stripeSubscriptionId },
	});

	if (!subscription) {
		log.warn("No local subscription found", { stripeSubscriptionId });
		return;
	}

	const statusMap: Record<string, string> = {
		active: "active",
		past_due: "past_due",
		canceled: "canceled",
		unpaid: "unpaid",
		trialing: "trialing",
		paused: "paused",
		incomplete: "incomplete",
		incomplete_expired: "incomplete_expired",
	};

	// In Stripe v18, period dates are on the subscription item level
	const firstItem = stripeSub.items.data[0];
	const periodStart = firstItem
		? new Date(firstItem.current_period_start * 1000)
		: undefined;
	const periodEnd = firstItem
		? new Date(firstItem.current_period_end * 1000)
		: undefined;

	await db.subscription.update({
		where: { id: subscription.id },
		data: {
			stripeStatus: stripeSub.status,
			status: statusMap[stripeSub.status] ?? "active",
			...(periodStart && { currentPeriodStart: periodStart }),
			...(periodEnd && { currentPeriodEnd: periodEnd }),
			cancelAtPeriodEnd: stripeSub.cancel_at_period_end,
			canceledAt: stripeSub.canceled_at
				? new Date(stripeSub.canceled_at * 1000)
				: null,
		},
	});
}

/**
 * Cancel a Stripe subscription (at period end by default).
 */
export async function cancelStripeSubscription(
	stripeSubscriptionId: string,
	immediately = false,
): Promise<Stripe.Subscription> {
	const stripe = getStripe();
	if (immediately) {
		return stripe.subscriptions.cancel(stripeSubscriptionId);
	}

	return stripe.subscriptions.update(stripeSubscriptionId, {
		cancel_at_period_end: true,
	});
}

/**
 * Resume a canceled (at period end) subscription.
 */
export async function resumeStripeSubscription(
	stripeSubscriptionId: string,
): Promise<Stripe.Subscription> {
	const stripe = getStripe();
	return stripe.subscriptions.update(stripeSubscriptionId, {
		cancel_at_period_end: false,
	});
}

/**
 * Update a Stripe subscription to a new price (plan change).
 */
export async function updateStripeSubscription(
	stripeSubscriptionId: string,
	newPriceId: string,
): Promise<Stripe.Subscription> {
	const stripe = getStripe();
	const subscription =
		await stripe.subscriptions.retrieve(stripeSubscriptionId);

	const firstItem = subscription.items.data[0];
	if (!firstItem) {
		throw new Error("Subscription has no items");
	}

	return stripe.subscriptions.update(stripeSubscriptionId, {
		items: [
			{
				id: firstItem.id,
				price: newPriceId,
			},
		],
		proration_behavior: "create_prorations",
	});
}

/**
 * Retrieve a Stripe checkout session with expanded subscription.
 */
export async function getCheckoutSession(
	sessionId: string,
): Promise<Stripe.Checkout.Session> {
	const stripe = getStripe();
	return stripe.checkout.sessions.retrieve(sessionId, {
		expand: ["subscription", "customer"],
	});
}
