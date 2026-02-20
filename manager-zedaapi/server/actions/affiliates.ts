"use server";

import { requireAuth } from "@/lib/auth-server";
import {
	registerAffiliate,
	getAffiliate,
	getCommissions as getCommissionsList,
	getPayouts as getPayoutsList,
	getReferrals as getReferralsList,
	getAffiliateStats,
	createPayout,
} from "@/server/services/affiliate-service";
import type { ActionResult } from "@/types";

// =============================================================================
// Affiliate Server Actions
// =============================================================================

export async function registerAsAffiliate(): Promise<ActionResult> {
	try {
		const session = await requireAuth();
		await registerAffiliate(session.user.id);
		return { success: true };
	} catch (error) {
		return {
			success: false,
			error:
				error instanceof Error
					? error.message
					: "Erro ao registrar como afiliado",
		};
	}
}

export async function getAffiliateProfile(): Promise<
	ActionResult<{
		affiliate: Awaited<ReturnType<typeof getAffiliate>>;
		stats: Awaited<ReturnType<typeof getAffiliateStats>> | null;
	}>
> {
	try {
		const session = await requireAuth();
		const affiliate = await getAffiliate(session.user.id);

		if (!affiliate) {
			return { success: true, data: { affiliate: null, stats: null } };
		}

		const stats = await getAffiliateStats(affiliate.id);
		return { success: true, data: { affiliate, stats } };
	} catch (error) {
		return {
			success: false,
			error:
				error instanceof Error
					? error.message
					: "Erro ao buscar perfil de afiliado",
		};
	}
}

export async function getCommissions(
	page = 1,
): Promise<ActionResult<Awaited<ReturnType<typeof getCommissionsList>>>> {
	try {
		const session = await requireAuth();
		const affiliate = await getAffiliate(session.user.id);

		if (!affiliate) {
			return {
				success: false,
				error: "Você não é afiliado",
				code: "NOT_AFFILIATE",
			};
		}

		const result = await getCommissionsList(affiliate.id, page);
		return { success: true, data: result };
	} catch (error) {
		return {
			success: false,
			error:
				error instanceof Error
					? error.message
					: "Erro ao buscar comissões",
		};
	}
}

export async function getPayouts(
	page = 1,
): Promise<ActionResult<Awaited<ReturnType<typeof getPayoutsList>>>> {
	try {
		const session = await requireAuth();
		const affiliate = await getAffiliate(session.user.id);

		if (!affiliate) {
			return {
				success: false,
				error: "Você não é afiliado",
				code: "NOT_AFFILIATE",
			};
		}

		const result = await getPayoutsList(affiliate.id, page);
		return { success: true, data: result };
	} catch (error) {
		return {
			success: false,
			error:
				error instanceof Error
					? error.message
					: "Erro ao buscar pagamentos",
		};
	}
}

export async function getReferrals(
	page = 1,
): Promise<ActionResult<Awaited<ReturnType<typeof getReferralsList>>>> {
	try {
		const session = await requireAuth();
		const affiliate = await getAffiliate(session.user.id);

		if (!affiliate) {
			return {
				success: false,
				error: "Você não é afiliado",
				code: "NOT_AFFILIATE",
			};
		}

		const result = await getReferralsList(affiliate.id, page);
		return { success: true, data: result };
	} catch (error) {
		return {
			success: false,
			error:
				error instanceof Error
					? error.message
					: "Erro ao buscar indicações",
		};
	}
}

export async function requestPayout(
	amount: number,
	method: string,
): Promise<ActionResult> {
	try {
		const session = await requireAuth();
		const affiliate = await getAffiliate(session.user.id);

		if (!affiliate) {
			return {
				success: false,
				error: "Você não é afiliado",
				code: "NOT_AFFILIATE",
			};
		}

		if (amount <= 0) {
			return {
				success: false,
				error: "Valor deve ser maior que zero",
			};
		}

		if (!["pix", "bank_transfer"].includes(method)) {
			return {
				success: false,
				error: "Método de pagamento inválido",
			};
		}

		await createPayout(affiliate.id, amount, method);
		return { success: true };
	} catch (error) {
		return {
			success: false,
			error:
				error instanceof Error
					? error.message
					: "Erro ao solicitar pagamento",
		};
	}
}
