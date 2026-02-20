"use server";

import { requireAuth } from "@/lib/auth-server";
import { db } from "@/lib/db";
import {
	getInvoices as getInvoicesList,
	getInvoice as getInvoiceDetail,
	createSicrediCharge,
} from "@/server/services/billing-service";
import { getBillingPortalUrl } from "@/server/services/subscription-service";
import type { ActionResult } from "@/types";

/**
 * Get paginated invoice list for current user.
 */
export async function getInvoices(page = 1) {
	try {
		const session = await requireAuth();
		return getInvoicesList(session.user.id, page);
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { invoices: [], total: 0, page: 1, pageSize: 20 };
	}
}

/**
 * Get a single invoice detail.
 */
export async function getInvoice(
	invoiceId: string,
): Promise<ActionResult<Awaited<ReturnType<typeof getInvoiceDetail>>>> {
	try {
		const session = await requireAuth();

		const invoice = await getInvoiceDetail(invoiceId);
		if (!invoice || invoice.userId !== session.user.id) {
			return { success: false, error: "Invoice not found" };
		}

		return { success: true, data: invoice };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch invoice" };
	}
}

/**
 * Request a Sicredi payment (PIX or Boleto Hibrido) for an invoice.
 */
export async function requestSicrediPayment(
	invoiceId: string,
	type: "pix" | "boleto_hibrido",
): Promise<ActionResult> {
	try {
		const session = await requireAuth();

		const invoice = await db.invoice.findFirst({
			where: {
				id: invoiceId,
				userId: session.user.id,
			},
		});

		if (!invoice) {
			return { success: false, error: "Invoice not found" };
		}

		if (invoice.status === "paid") {
			return { success: false, error: "Invoice already paid" };
		}

		await createSicrediCharge(invoiceId, type);
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return {
			success: false,
			error:
				error instanceof Error
					? error.message
					: "Failed to create payment",
		};
	}
}

/**
 * Create Stripe billing portal session URL.
 */
export async function createBillingPortal(): Promise<
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
