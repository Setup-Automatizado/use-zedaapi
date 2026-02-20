import type { Job } from "bullmq";
import type { AffiliatePayoutJobData } from "@/lib/queue/types";
import { createLogger } from "@/lib/queue/logger";

const log = createLogger("processor:affiliate-payout");

export async function processAffiliatePayoutJob(
	job: Job<AffiliatePayoutJobData>,
): Promise<void> {
	const { affiliateId, amount, method } = job.data;

	log.info("Processing affiliate payout", {
		jobId: job.id,
		affiliateId,
		amount,
		method,
	});

	const done = log.timer("Affiliate payout", { affiliateId, amount });
	const { db } = await import("@/lib/db");

	const affiliate = await db.affiliate.findUnique({
		where: { id: affiliateId },
		include: { user: true },
	});

	if (!affiliate) {
		log.error("Affiliate not found", { affiliateId });
		throw new Error(`Affiliate ${affiliateId} not found`);
	}

	if (affiliate.status !== "active") {
		log.warn("Affiliate is not active, skipping payout", {
			affiliateId,
			status: affiliate.status,
		});
		return;
	}

	// Create payout record
	const payout = await db.payout.create({
		data: {
			affiliateId,
			amount,
			method,
			status: "processing",
		},
	});

	try {
		if (method === "stripe_transfer") {
			await processStripeTransfer(affiliate, payout, amount, db);
		} else {
			await processManualPayout(payout, db);
		}

		done();
	} catch (error) {
		await db.payout.update({
			where: { id: payout.id },
			data: { status: "failed" },
		});

		log.error("Affiliate payout failed", {
			affiliateId,
			payoutId: payout.id,
			error: error instanceof Error ? error.message : "Unknown error",
		});

		throw error;
	}
}

async function processStripeTransfer(
	affiliate: {
		id: string;
		stripeConnectAccountId: string | null;
		totalPaid: unknown;
	},
	payout: { id: string },
	amount: number,
	db: Awaited<ReturnType<typeof getDb>>,
): Promise<void> {
	if (!affiliate.stripeConnectAccountId) {
		log.error("Affiliate has no Stripe Connect account", {
			affiliateId: affiliate.id,
		});
		throw new Error("Missing Stripe Connect account for transfer");
	}

	// Import Stripe
	const Stripe = (await import("stripe")).default;
	const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!);

	const transfer = await stripe.transfers.create({
		amount: Math.round(amount * 100), // Convert to cents
		currency: "brl",
		destination: affiliate.stripeConnectAccountId,
		description: `Affiliate payout - ${affiliate.id}`,
	});

	await db.payout.update({
		where: { id: payout.id },
		data: {
			status: "completed",
			stripeTransferId: transfer.id,
			processedAt: new Date(),
		},
	});

	// Update affiliate totals
	await db.affiliate.update({
		where: { id: affiliate.id },
		data: {
			totalPaid: { increment: amount },
		},
	});

	log.info("Stripe transfer completed", {
		affiliateId: affiliate.id,
		payoutId: payout.id,
		transferId: transfer.id,
		amount,
	});
}

async function processManualPayout(
	payout: { id: string },
	db: Awaited<ReturnType<typeof getDb>>,
): Promise<void> {
	// Manual payouts are just marked as pending_manual for admin to process
	await db.payout.update({
		where: { id: payout.id },
		data: { status: "pending_manual" },
	});

	log.info("Manual payout marked for admin processing", {
		payoutId: payout.id,
	});
}

async function getDb() {
	const { db } = await import("@/lib/db");
	return db;
}
