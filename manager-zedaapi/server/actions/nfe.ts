"use server";

import { revalidatePath } from "next/cache";
import { isRedirectError } from "next/dist/client/components/redirect-error";
import { requireAdmin } from "@/lib/auth-server";
import { db } from "@/lib/db";
import {
	emitNfse,
	cancelNfse,
	queryNfseStatus,
	getActiveConfig,
} from "@/lib/services/nfse/service";

/**
 * Manually trigger NFS-e issuance for an invoice.
 */
export async function issueNfse(invoiceId: string) {
	try {
		await requireAdmin();

		const result = await emitNfse(invoiceId);

		revalidatePath("/admin/nfe");
		revalidatePath(`/admin/faturas/${invoiceId}`);

		return { success: true as const, data: result };
	} catch (error) {
		if (isRedirectError(error)) throw error;
		return {
			success: false as const,
			error:
				error instanceof Error
					? error.message
					: "Failed to issue NFS-e",
		};
	}
}

/**
 * Cancel a previously issued NFS-e.
 */
export async function cancelNfseAction(
	invoiceId: string,
	motivo: string = "Cancelamento solicitado pelo administrador",
) {
	try {
		await requireAdmin();

		const result = await cancelNfse(invoiceId, motivo);

		revalidatePath("/admin/nfe");
		revalidatePath(`/admin/faturas/${invoiceId}`);

		return { success: true as const, data: result };
	} catch (error) {
		if (isRedirectError(error)) throw error;
		return {
			success: false as const,
			error:
				error instanceof Error
					? error.message
					: "Failed to cancel NFS-e",
		};
	}
}

/**
 * Check NFS-e status for an invoice.
 */
export async function getNfseStatus(invoiceId: string) {
	try {
		await requireAdmin();

		await queryNfseStatus(invoiceId);

		const invoice = await db.invoice.findUnique({
			where: { id: invoiceId },
			select: {
				id: true,
				nfseStatus: true,
				nfseNumber: true,
				nfseProtocol: true,
				nfseXmlUrl: true,
				nfsePdfUrl: true,
				nfseIssuedAt: true,
				nfseError: true,
				nfseCanceledAt: true,
			},
		});

		revalidatePath("/admin/nfe");

		return invoice;
	} catch (error) {
		if (isRedirectError(error)) throw error;
		return null;
	}
}

/**
 * Get active NFS-e configuration (admin only).
 */
export async function getNfseConfig() {
	try {
		await requireAdmin();

		const config = await getActiveConfig();
		if (!config) return null;

		return {
			id: config.id,
			active: config.active,
			cnpj: config.cnpj,
			inscricaoMunicipal: config.inscricaoMunicipal,
			codigoMunicipio: config.codigoMunicipio,
			uf: config.uf,
			certificateExpiresAt: config.certificateExpiresAt,
			codigoServico: config.codigoServico,
			cnae: config.cnae,
			aliquotaIss: config.aliquotaIss,
			descricaoServico: config.descricaoServico,
			codigoServicoPf: config.codigoServicoPf,
			cnaePf: config.cnaePf,
			aliquotaIssPf: config.aliquotaIssPf,
			descricaoServicoPf: config.descricaoServicoPf,
			ambiente: config.ambiente,
		};
	} catch (error) {
		if (isRedirectError(error)) throw error;
		return null;
	}
}

/**
 * Retry failed NFS-e issuance.
 */
export async function retryNfse(invoiceId: string) {
	try {
		await requireAdmin();

		const invoice = await db.invoice.findUnique({
			where: { id: invoiceId },
			select: { nfseStatus: true },
		});

		if (!invoice) {
			return { success: false as const, error: "Invoice not found" };
		}

		if (invoice.nfseStatus !== "ERROR") {
			return {
				success: false as const,
				error: "Only NFS-e with error status can be retried",
			};
		}

		await db.invoice.update({
			where: { id: invoiceId },
			data: {
				nfseStatus: "PENDING",
				nfseError: null,
			},
		});

		const result = await emitNfse(invoiceId);

		revalidatePath("/admin/nfe");
		revalidatePath(`/admin/faturas/${invoiceId}`);

		return { success: true as const, data: result };
	} catch (error) {
		if (isRedirectError(error)) throw error;
		return {
			success: false as const,
			error:
				error instanceof Error
					? error.message
					: "Failed to retry NFS-e",
		};
	}
}
