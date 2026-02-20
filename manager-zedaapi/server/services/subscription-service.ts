"use server";

import { db } from "@/lib/db";
import {
	createCheckoutSession,
	createOrGetCustomer,
	cancelStripeSubscription,
	updateStripeSubscription,
	createBillingPortalSession,
} from "./stripe-service";

/**
 * Create a new subscription for a user.
 * For Stripe: redirects to Stripe Checkout.
 * For PIX/Boleto: creates a pending subscription + Sicredi charge.
 */
export async function createSubscription(
	userId: string,
	planId: string,
	paymentMethod: "stripe" | "pix" | "boleto" = "stripe",
) {
	const plan = await db.plan.findUnique({
		where: { id: planId },
	});

	if (!plan) {
		throw new Error("Plano não encontrado");
	}

	if (!plan.active) {
		throw new Error("Plano não está disponível");
	}

	// Check for existing active subscription
	const existingSub = await db.subscription.findFirst({
		where: {
			userId,
			status: { in: ["active", "trialing"] },
		},
	});

	if (existingSub) {
		throw new Error("Você já possui uma assinatura ativa");
	}

	if (paymentMethod === "stripe") {
		const session = await createCheckoutSession(userId, planId, "stripe");
		return { type: "redirect" as const, url: session.url };
	}

	// PIX or Boleto: create a pending subscription locally
	const now = new Date();
	const periodEnd = new Date(now);
	if (plan.interval === "year") {
		periodEnd.setFullYear(periodEnd.getFullYear() + 1);
	} else {
		periodEnd.setMonth(periodEnd.getMonth() + 1);
	}

	const subscription = await db.subscription.create({
		data: {
			userId,
			planId,
			paymentMethod,
			status: "incomplete",
			currentPeriodStart: now,
			currentPeriodEnd: periodEnd,
		},
	});

	// Create the invoice
	const invoice = await db.invoice.create({
		data: {
			userId,
			subscriptionId: subscription.id,
			amount: plan.price,
			currency: plan.currency,
			status: "pending",
			dueDate:
				paymentMethod === "boleto" ? calculateDueDate(3) : undefined,
			paymentMethod,
		},
	});

	return {
		type: "pending" as const,
		subscriptionId: subscription.id,
		invoiceId: invoice.id,
	};
}

/**
 * Cancel a subscription.
 */
export async function cancelSubscription(
	subscriptionId: string,
	immediately = false,
) {
	const subscription = await db.subscription.findUnique({
		where: { id: subscriptionId },
	});

	if (!subscription) {
		throw new Error("Assinatura não encontrada");
	}

	// For Stripe subscriptions, cancel via Stripe API
	if (subscription.stripeSubscriptionId) {
		await cancelStripeSubscription(
			subscription.stripeSubscriptionId,
			immediately,
		);

		if (immediately) {
			await db.subscription.update({
				where: { id: subscriptionId },
				data: {
					status: "canceled",
					canceledAt: new Date(),
				},
			});
		} else {
			await db.subscription.update({
				where: { id: subscriptionId },
				data: {
					cancelAtPeriodEnd: true,
				},
			});
		}

		return;
	}

	// For PIX/Boleto subscriptions, cancel directly
	await db.subscription.update({
		where: { id: subscriptionId },
		data: {
			status: "canceled",
			canceledAt: new Date(),
		},
	});
}

/**
 * Change a subscription to a different plan.
 */
export async function changeSubscription(
	subscriptionId: string,
	newPlanId: string,
) {
	const subscription = await db.subscription.findUnique({
		where: { id: subscriptionId },
		include: { plan: true },
	});

	if (!subscription) {
		throw new Error("Assinatura não encontrada");
	}

	if (subscription.status !== "active") {
		throw new Error("Somente assinaturas ativas podem ser alteradas");
	}

	const newPlan = await db.plan.findUnique({
		where: { id: newPlanId },
	});

	if (!newPlan?.stripePriceId) {
		throw new Error("Novo plano não encontrado ou sem preço configurado");
	}

	// Stripe-managed subscription
	if (subscription.stripeSubscriptionId) {
		await updateStripeSubscription(
			subscription.stripeSubscriptionId,
			newPlan.stripePriceId,
		);

		await db.subscription.update({
			where: { id: subscriptionId },
			data: { planId: newPlanId },
		});

		return;
	}

	// PIX/Boleto: just update the plan (next billing cycle will use new price)
	await db.subscription.update({
		where: { id: subscriptionId },
		data: { planId: newPlanId },
	});
}

/**
 * Get the active subscription for a user.
 */
export async function getActiveSubscription(userId: string) {
	return db.subscription.findFirst({
		where: {
			userId,
			status: { in: ["active", "trialing", "past_due"] },
		},
		include: {
			plan: true,
			_count: {
				select: { instances: true },
			},
		},
		orderBy: { createdAt: "desc" },
	});
}

/**
 * Check if the subscription has room for more instances.
 */
export async function checkInstanceLimit(
	subscriptionId: string,
): Promise<{ allowed: boolean; current: number; max: number }> {
	const subscription = await db.subscription.findUnique({
		where: { id: subscriptionId },
		include: {
			plan: true,
			_count: {
				select: { instances: true },
			},
		},
	});

	if (!subscription) {
		return { allowed: false, current: 0, max: 0 };
	}

	const current = subscription._count.instances;
	const max = subscription.plan.maxInstances;

	return {
		allowed: current < max,
		current,
		max,
	};
}

/**
 * Get billing portal URL for a user.
 */
export async function getBillingPortalUrl(userId: string): Promise<string> {
	const customerId = await createOrGetCustomer(userId);
	return createBillingPortalSession(customerId);
}

function calculateDueDate(businessDays: number): Date {
	const date = new Date();
	let added = 0;
	while (added < businessDays) {
		date.setDate(date.getDate() + 1);
		const dayOfWeek = date.getDay();
		if (dayOfWeek !== 0 && dayOfWeek !== 6) {
			added++;
		}
	}
	return date;
}
