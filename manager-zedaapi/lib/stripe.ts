import Stripe from "stripe";

if (!process.env.STRIPE_SECRET_KEY) {
	throw new Error("STRIPE_SECRET_KEY is not set");
}

export const stripe = new Stripe(process.env.STRIPE_SECRET_KEY, {
	apiVersion: "2025-08-27.basil",
	typescript: true,
});

export async function constructWebhookEvent(
	body: string,
	signature: string,
): Promise<Stripe.Event> {
	const secret = process.env.STRIPE_WEBHOOK_SECRET;
	if (!secret) {
		throw new Error("STRIPE_WEBHOOK_SECRET is not set");
	}
	return stripe.webhooks.constructEvent(body, signature, secret);
}
