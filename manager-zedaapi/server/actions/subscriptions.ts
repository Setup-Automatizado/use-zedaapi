"use server";

import { requireAuth } from "@/lib/auth-server";
import { db } from "@/lib/db";
import {
	createSubscription as createSub,
	cancelSubscription as cancelSub,
	changeSubscription as changeSub,
	getActiveSubscription as getActiveSub,
	getBillingPortalUrl,
} from "@/server/services/subscription-service";
import { resumeStripeSubscription } from "@/server/services/stripe-service";
import type { ActionResult } from "@/types";

/**
 * Get all available plans.
 */
export async function getPlans() {
	try {
		return db.plan.findMany({
			where: { active: true },
			orderBy: { sortOrder: "asc" },
		});
	} catch {
		return [];
	}
}

/**
 * Get current user's active subscription with plan details.
 */
export async function getSubscription() {
	try {
		const session = await requireAuth();
		return getActiveSub(session.user.id);
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return null;
	}
}

/**
 * Start checkout for a plan.
 */
export async function checkout(
	planId: string,
	paymentMethod: "stripe" | "pix" | "boleto" = "stripe",
): Promise<ActionResult<{ url?: string }>> {
	try {
		const session = await requireAuth();
		const result = await createSub(session.user.id, planId, paymentMethod);
		const url = "url" in result ? (result.url ?? undefined) : undefined;
		return { success: true, data: { url } };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return {
			success: false,
			error:
				error instanceof Error
					? error.message
					: "Failed to create checkout",
		};
	}
}

/**
 * Cancel a subscription.
 */
export async function cancelSubscription(
	subscriptionId: string,
	immediately = false,
): Promise<ActionResult> {
	try {
		const session = await requireAuth();

		const subscription = await db.subscription.findFirst({
			where: {
				id: subscriptionId,
				userId: session.user.id,
			},
		});

		if (!subscription) {
			return { success: false, error: "Subscription not found" };
		}

		await cancelSub(subscriptionId, immediately);
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return {
			success: false,
			error:
				error instanceof Error
					? error.message
					: "Failed to cancel subscription",
		};
	}
}

/**
 * Change to a different plan.
 */
export async function changePlan(
	subscriptionId: string,
	newPlanId: string,
): Promise<ActionResult> {
	try {
		const session = await requireAuth();

		const subscription = await db.subscription.findFirst({
			where: {
				id: subscriptionId,
				userId: session.user.id,
			},
		});

		if (!subscription) {
			return { success: false, error: "Subscription not found" };
		}

		await changeSub(subscriptionId, newPlanId);
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return {
			success: false,
			error:
				error instanceof Error
					? error.message
					: "Failed to change plan",
		};
	}
}

/**
 * Resume a subscription that was scheduled for cancellation.
 */
export async function resumeSubscription(
	subscriptionId: string,
): Promise<ActionResult> {
	try {
		const session = await requireAuth();

		const subscription = await db.subscription.findFirst({
			where: {
				id: subscriptionId,
				userId: session.user.id,
			},
		});

		if (!subscription) {
			return { success: false, error: "Subscription not found" };
		}

		if (!subscription.cancelAtPeriodEnd) {
			return {
				success: false,
				error: "Subscription is not scheduled for cancellation",
			};
		}

		if (!subscription.stripeSubscriptionId) {
			return {
				success: false,
				error: "Only Stripe subscriptions can be resumed",
			};
		}

		await resumeStripeSubscription(subscription.stripeSubscriptionId);

		await db.subscription.update({
			where: { id: subscriptionId },
			data: { cancelAtPeriodEnd: false },
		});

		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return {
			success: false,
			error:
				error instanceof Error
					? error.message
					: "Failed to resume subscription",
		};
	}
}

/**
 * Get Stripe billing portal URL.
 */
export async function getBillingPortal(): Promise<
	ActionResult<{ url: string }>
> {
	try {
		const session = await requireAuth();
		const url = await getBillingPortalUrl(session.user.id);
		return { success: true, data: { url } };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return {
			success: false,
			error:
				error instanceof Error
					? error.message
					: "Failed to get billing portal",
		};
	}
}
