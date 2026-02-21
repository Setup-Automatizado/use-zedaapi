import Stripe from "stripe";

let _stripe: Stripe | undefined;

/**
 * Lazy-initialized Stripe client.
 * Defers creation until first call so module evaluation
 * doesn't throw during Next.js build (Docker without env vars).
 */
export function getStripe(): Stripe {
	if (!_stripe) {
		if (!process.env.STRIPE_SECRET_KEY) {
			throw new Error("STRIPE_SECRET_KEY is not set");
		}
		_stripe = new Stripe(process.env.STRIPE_SECRET_KEY, {
			apiVersion: "2025-08-27.basil",
			typescript: true,
		});
	}
	return _stripe;
}

export async function constructWebhookEvent(
	body: string,
	signature: string,
): Promise<Stripe.Event> {
	const secret = process.env.STRIPE_WEBHOOK_SECRET;
	if (!secret) {
		throw new Error("STRIPE_WEBHOOK_SECRET is not set");
	}
	return getStripe().webhooks.constructEvent(body, signature, secret);
}
