"use server";

import { db } from "@/lib/db";
import { AppError, NotFoundError } from "@/lib/errors";

// =============================================================================
// Affiliate System Service
// =============================================================================

/**
 * Generate a unique affiliate referral code from user name + random suffix.
 */
function generateAffiliateCode(userName: string): string {
	const base = userName
		.toLowerCase()
		.replace(/[^a-z0-9]/g, "")
		.slice(0, 10);
	const suffix = Math.random().toString(36).slice(2, 8);
	return `${base}-${suffix}`;
}

/**
 * Register a user as an affiliate. Creates affiliate record with unique code.
 */
export async function registerAffiliate(userId: string) {
	const existing = await db.affiliate.findUnique({ where: { userId } });
	if (existing) {
		throw new AppError("Usuário já é afiliado", 409, "ALREADY_AFFILIATE");
	}

	const user = await db.user.findUnique({
		where: { id: userId },
		select: { name: true },
	});

	if (!user) {
		throw new NotFoundError("User");
	}

	const code = generateAffiliateCode(user.name);

	const affiliate = await db.affiliate.create({
		data: {
			userId,
			code,
			status: "active",
			commissionRate: 0.2, // 20% default
		},
	});

	return affiliate;
}

/**
 * Get affiliate details for a user.
 */
export async function getAffiliate(userId: string) {
	const affiliate = await db.affiliate.findUnique({
		where: { userId },
		include: {
			_count: {
				select: {
					referrals: true,
					commissions: true,
					payouts: true,
				},
			},
		},
	});

	if (!affiliate) {
		return null;
	}

	return affiliate;
}

/**
 * Track a referral: link referred user to the affiliate.
 */
export async function trackReferral(
	affiliateCode: string,
	referredUserId: string,
) {
	const affiliate = await db.affiliate.findUnique({
		where: { code: affiliateCode },
	});

	if (!affiliate) {
		throw new NotFoundError("Affiliate");
	}

	if (affiliate.status !== "active") {
		throw new AppError(
			"Afiliado não está ativo",
			400,
			"AFFILIATE_INACTIVE",
		);
	}

	// Don't allow self-referral
	if (affiliate.userId === referredUserId) {
		throw new AppError(
			"Auto-referência não permitida",
			400,
			"SELF_REFERRAL",
		);
	}

	// Check if user already has a referral
	const existingReferral = await db.referral.findUnique({
		where: { referredUserId },
	});
	if (existingReferral) {
		return existingReferral;
	}

	const referral = await db.referral.create({
		data: {
			affiliateId: affiliate.id,
			referredUserId,
			status: "pending",
		},
	});

	return referral;
}

/**
 * Calculate commission for a paid invoice.
 * Finds the referral chain and creates a commission record.
 */
export async function calculateCommission(invoiceId: string) {
	const invoice = await db.invoice.findUnique({
		where: { id: invoiceId },
		select: {
			id: true,
			amount: true,
			userId: true,
			status: true,
			subscriptionId: true,
		},
	});

	if (!invoice || invoice.status !== "paid") {
		return null;
	}

	// Find referral for this user
	const referral = await db.referral.findUnique({
		where: { referredUserId: invoice.userId },
		include: {
			affiliate: {
				select: { id: true, commissionRate: true, status: true },
			},
		},
	});

	if (!referral || referral.affiliate.status !== "active") {
		return null;
	}

	// Check for existing commission on this invoice
	const existing = await db.commission.findFirst({
		where: { invoiceId, affiliateId: referral.affiliateId },
	});
	if (existing) {
		return existing;
	}

	// Mark referral as converted if not yet
	if (referral.status === "pending") {
		await db.referral.update({
			where: { id: referral.id },
			data: {
				status: "converted",
				convertedAt: new Date(),
				subscriptionId: invoice.subscriptionId,
			},
		});
	}

	const commissionAmount =
		Number(invoice.amount) * Number(referral.affiliate.commissionRate);

	const commission = await db.commission.create({
		data: {
			affiliateId: referral.affiliateId,
			referralId: referral.id,
			invoiceId: invoice.id,
			amount: commissionAmount,
			status: "pending",
		},
	});

	return commission;
}

/**
 * Approve a pending commission for payout.
 */
export async function approveCommission(commissionId: string) {
	const commission = await db.commission.findUnique({
		where: { id: commissionId },
	});

	if (!commission) {
		throw new NotFoundError("Commission");
	}

	if (commission.status !== "pending") {
		throw new AppError(
			"Comissão já foi processada",
			400,
			"COMMISSION_ALREADY_PROCESSED",
		);
	}

	const updated = await db.commission.update({
		where: { id: commissionId },
		data: { status: "approved" },
	});

	return updated;
}

/**
 * Create a payout for the affiliate.
 */
export async function createPayout(
	affiliateId: string,
	amount: number,
	method: string,
) {
	const affiliate = await db.affiliate.findUnique({
		where: { id: affiliateId },
	});

	if (!affiliate) {
		throw new NotFoundError("Affiliate");
	}

	// Verify sufficient approved commissions
	const approvedCommissions = await db.commission.aggregate({
		where: { affiliateId, status: "approved" },
		_sum: { amount: true },
	});

	const availableAmount = Number(approvedCommissions._sum.amount || 0);
	if (amount > availableAmount) {
		throw new AppError(
			`Saldo insuficiente. Disponível: R$ ${availableAmount.toFixed(2)}`,
			400,
			"INSUFFICIENT_BALANCE",
		);
	}

	const payout = await db.$transaction(async (tx) => {
		// Create payout record
		const p = await tx.payout.create({
			data: {
				affiliateId,
				amount,
				method,
				status: "pending",
			},
		});

		// Mark commissions as paid up to payout amount
		let remaining = amount;
		const commissions = await tx.commission.findMany({
			where: { affiliateId, status: "approved" },
			orderBy: { createdAt: "asc" },
		});

		for (const c of commissions) {
			if (remaining <= 0) break;
			const commissionAmount = Number(c.amount);
			if (commissionAmount <= remaining) {
				await tx.commission.update({
					where: { id: c.id },
					data: { status: "paid", paidAt: new Date() },
				});
				remaining -= commissionAmount;
			}
		}

		// Update affiliate totals
		await tx.affiliate.update({
			where: { id: affiliateId },
			data: {
				totalPaid: { increment: amount },
			},
		});

		return p;
	});

	return payout;
}

/**
 * List commissions for an affiliate with pagination.
 */
export async function getCommissions(
	affiliateId: string,
	page = 1,
	pageSize = 20,
) {
	const skip = (page - 1) * pageSize;

	const [items, total] = await Promise.all([
		db.commission.findMany({
			where: { affiliateId },
			include: {
				referral: {
					include: {
						referredUser: {
							select: { name: true, email: true },
						},
					},
				},
				invoice: {
					select: { id: true, amount: true, paidAt: true },
				},
			},
			orderBy: { createdAt: "desc" },
			skip,
			take: pageSize,
		}),
		db.commission.count({ where: { affiliateId } }),
	]);

	return { items, total, page, pageSize };
}

/**
 * List payouts for an affiliate.
 */
export async function getPayouts(affiliateId: string, page = 1, pageSize = 20) {
	const skip = (page - 1) * pageSize;

	const [items, total] = await Promise.all([
		db.payout.findMany({
			where: { affiliateId },
			orderBy: { createdAt: "desc" },
			skip,
			take: pageSize,
		}),
		db.payout.count({ where: { affiliateId } }),
	]);

	return { items, total, page, pageSize };
}

/**
 * List referrals for an affiliate.
 */
export async function getReferrals(
	affiliateId: string,
	page = 1,
	pageSize = 20,
) {
	const skip = (page - 1) * pageSize;

	const [items, total] = await Promise.all([
		db.referral.findMany({
			where: { affiliateId },
			include: {
				referredUser: {
					select: { name: true, email: true, createdAt: true },
				},
				subscription: {
					select: {
						plan: { select: { name: true } },
						status: true,
					},
				},
			},
			orderBy: { createdAt: "desc" },
			skip,
			take: pageSize,
		}),
		db.referral.count({ where: { affiliateId } }),
	]);

	return { items, total, page, pageSize };
}

/**
 * Get aggregate stats for an affiliate.
 */
export async function getAffiliateStats(affiliateId: string) {
	const [affiliate, pendingSum, approvedSum, paidSum, referralCount] =
		await Promise.all([
			db.affiliate.findUnique({
				where: { id: affiliateId },
				select: {
					totalEarnings: true,
					totalPaid: true,
					commissionRate: true,
				},
			}),
			db.commission.aggregate({
				where: { affiliateId, status: "pending" },
				_sum: { amount: true },
			}),
			db.commission.aggregate({
				where: { affiliateId, status: "approved" },
				_sum: { amount: true },
			}),
			db.commission.aggregate({
				where: { affiliateId, status: "paid" },
				_sum: { amount: true },
			}),
			db.referral.count({ where: { affiliateId } }),
		]);

	const convertedCount = await db.referral.count({
		where: { affiliateId, status: "converted" },
	});

	return {
		totalEarnings: Number(affiliate?.totalEarnings || 0),
		totalPaid: Number(affiliate?.totalPaid || 0),
		pendingAmount: Number(pendingSum._sum.amount || 0),
		approvedAmount: Number(approvedSum._sum.amount || 0),
		paidAmount: Number(paidSum._sum.amount || 0),
		commissionRate: Number(affiliate?.commissionRate || 0),
		totalReferrals: referralCount,
		convertedReferrals: convertedCount,
		conversionRate:
			referralCount > 0 ? (convertedCount / referralCount) * 100 : 0,
	};
}
