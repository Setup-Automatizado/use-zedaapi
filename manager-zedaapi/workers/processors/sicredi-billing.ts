import type { Job } from "bullmq";
import type { SicrediBillingJobData } from "@/lib/queue/types";
import { createLogger } from "@/lib/queue/logger";

const log = createLogger("processor:sicredi-billing");

export async function processSicrediBillingJob(
	job: Job<SicrediBillingJobData>,
): Promise<void> {
	const { invoiceId, type, action } = job.data;

	log.info("Processing Sicredi billing job", {
		jobId: job.id,
		invoiceId,
		type,
		action,
	});

	const done = log.timer(`Sicredi ${action}`, { invoiceId, type });
	const { db } = await import("@/lib/db");

	switch (action) {
		case "create": {
			const invoice = await db.invoice.findUnique({
				where: { id: invoiceId },
				include: {
					user: true,
					subscription: { include: { plan: true } },
				},
			});

			if (!invoice) {
				log.error("Invoice not found", { invoiceId });
				throw new Error(`Invoice ${invoiceId} not found`);
			}

			// Check if charge already exists
			const existingCharge = await db.sicrediCharge.findUnique({
				where: { invoiceId },
			});

			if (existingCharge) {
				log.info("Sicredi charge already exists", {
					invoiceId,
					chargeId: existingCharge.id,
				});
				return;
			}

			if (type === "pix") {
				await createPixCharge(invoice, db);
			} else {
				await createBoletoCharge(invoice, db);
			}

			done();
			break;
		}

		case "check_status": {
			const charge = await db.sicrediCharge.findFirst({
				where: { invoiceId },
			});

			if (!charge) {
				log.warn("Sicredi charge not found for status check", {
					invoiceId,
				});
				return;
			}

			if (
				charge.status === "CONCLUIDA" ||
				charge.status === "LIQUIDADO"
			) {
				log.info("Charge already settled", {
					invoiceId,
					status: charge.status,
				});
				return;
			}

			await checkChargeStatus(charge, db);
			done();
			break;
		}

		default:
			log.error("Unknown Sicredi billing action", { action });
	}
}

async function createPixCharge(
	invoice: Record<string, unknown>,
	db: Awaited<ReturnType<typeof getDb>>,
): Promise<void> {
	const userId = invoice.userId as string;
	const amount = Number(invoice.amount);
	const invoiceId = invoice.id as string;
	const user = invoice.user as Record<string, unknown> | undefined;

	// Generate txid for idempotency
	const txid = `MZA${invoiceId.replace(/-/g, "").slice(0, 25)}`;
	const amountCents = Math.round(amount * 100);

	const { createPixChargeWithTxid } =
		await import("@/lib/services/sicredi/pix");

	const cpfCnpj = user?.cpfCnpj as string | undefined;
	const pixResponse = await createPixChargeWithTxid(txid, {
		amountCents,
		description: `Assinatura Zé da API - Fatura ${invoiceId.slice(0, 8)}`,
		expirationSeconds: 86400, // 24h
		devedor: cpfCnpj
			? {
					nome: (user?.name as string) || "Cliente",
					...(cpfCnpj.length <= 11
						? { cpf: cpfCnpj }
						: { cnpj: cpfCnpj }),
				}
			: undefined,
	});

	const charge = await db.sicrediCharge.create({
		data: {
			userId,
			invoiceId,
			type: "pix",
			txid: pixResponse.txid,
			pixCopiaECola: pixResponse.pixCopiaECola || null,
			amount,
			status: "ATIVA",
			expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000), // 24h
		},
	});

	// Update invoice with charge reference
	await db.invoice.update({
		where: { id: invoiceId },
		data: {
			paymentMethod: "pix",
			sicrediChargeId: charge.id,
		},
	});

	log.info("PIX charge created", {
		chargeId: charge.id,
		txid: pixResponse.txid,
		invoiceId,
		amount,
		pixCopiaECola: !!pixResponse.pixCopiaECola,
	});

	// Schedule status check after 5 minutes
	const { enqueueSicrediBilling } = await import("@/lib/queue/producers");
	await enqueueSicrediBilling(
		{ invoiceId, type: "pix", action: "check_status" },
		{ delay: 5 * 60 * 1000 },
	);
}

async function createBoletoCharge(
	invoice: Record<string, unknown>,
	db: Awaited<ReturnType<typeof getDb>>,
): Promise<void> {
	const userId = invoice.userId as string;
	const amount = Number(invoice.amount);
	const invoiceId = invoice.id as string;
	const user = invoice.user as Record<string, unknown> | undefined;
	const amountCents = Math.round(amount * 100);

	const cpfCnpj = (user?.cpfCnpj as string | undefined)?.replace(/\D/g, "");
	if (!cpfCnpj || !user?.zipCode || !user?.city || !user?.state) {
		throw new Error(
			`Dados cadastrais incompletos para boleto. invoiceId=${invoiceId}`,
		);
	}

	const { createBoletoHibrido } =
		await import("@/lib/services/sicredi/boleto-hibrido");
	const { calculateDueDate } = await import("@/lib/services/sicredi/utils");

	const dueDate = calculateDueDate(3);

	const boletoResponse = await createBoletoHibrido({
		amountCents,
		dueDate,
		description: `Assinatura Zé da API - Fatura ${invoiceId.slice(0, 8)}`,
		pagador: {
			nome: (user.name as string) || "Cliente",
			documento: cpfCnpj,
			cep: (user.zipCode as string).replace(/\D/g, ""),
			cidade: user.city as string,
			endereco: (user.address as string) || "Nao informado",
			uf: user.state as string,
			email: (user.email as string) || "",
		},
		seuNumero: invoiceId.slice(0, 15),
	});

	const charge = await db.sicrediCharge.create({
		data: {
			userId,
			invoiceId,
			type: "boleto_hibrido",
			txid: boletoResponse.nossoNumero,
			codigoBarras: boletoResponse.codigoBarras,
			linhaDigitavel: boletoResponse.linhaDigitavel,
			pixCopiaECola: boletoResponse.qrCode || null,
			amount,
			status: "ATIVA",
			expiresAt: new Date(dueDate),
		},
	});

	await db.invoice.update({
		where: { id: invoiceId },
		data: {
			paymentMethod: "boleto_hibrido",
			sicrediChargeId: charge.id,
		},
	});

	log.info("Boleto hibrido charge created", {
		chargeId: charge.id,
		nossoNumero: boletoResponse.nossoNumero,
		invoiceId,
		amount,
	});

	// Schedule status check after 5 minutes
	const { enqueueSicrediBilling } = await import("@/lib/queue/producers");
	await enqueueSicrediBilling(
		{ invoiceId, type: "boleto_hibrido", action: "check_status" },
		{ delay: 5 * 60 * 1000 },
	);
}

async function checkChargeStatus(
	charge: Record<string, unknown>,
	db: Awaited<ReturnType<typeof getDb>>,
): Promise<void> {
	const chargeId = charge.id as string;
	const txid = charge.txid as string | null;
	const chargeType = charge.type as string;
	const invoiceId = charge.invoiceId as string | null;

	log.info("Checking charge status", { chargeId, txid, type: chargeType });

	if (chargeType === "pix" && txid) {
		const { getPixCharge } = await import("@/lib/services/sicredi/pix");
		const pixResponse = await getPixCharge(txid);

		if (pixResponse.status === "CONCLUIDA" && pixResponse.pix?.[0]) {
			const paidAt = new Date(pixResponse.pix[0].horario);
			await db.sicrediCharge.update({
				where: { id: chargeId },
				data: { status: "CONCLUIDA", paidAt },
			});

			const { handleSicrediPayment } =
				await import("@/server/services/billing-service");
			await handleSicrediPayment(txid, paidAt);

			log.info("PIX charge confirmed via status check", {
				chargeId,
				txid,
			});
			return;
		}

		log.info("PIX charge still pending", {
			chargeId,
			txid,
			apiStatus: pixResponse.status,
		});
	} else if (chargeType === "boleto_hibrido" && txid) {
		const { getBoletoHibrido, mapBoletoSituacao } =
			await import("@/lib/services/sicredi/boleto-hibrido");
		const boletoResponse = await getBoletoHibrido(txid);
		const { isPaid, isCancelled, rawSituacao } =
			mapBoletoSituacao(boletoResponse);

		if (isPaid) {
			await db.sicrediCharge.update({
				where: { id: chargeId },
				data: { status: "CONCLUIDA", paidAt: new Date() },
			});

			const { handleSicrediPayment } =
				await import("@/server/services/billing-service");
			await handleSicrediPayment(txid, new Date());

			log.info("Boleto hibrido confirmed via status check", {
				chargeId,
				txid,
				situacao: rawSituacao,
			});
			return;
		}

		if (isCancelled) {
			await db.sicrediCharge.update({
				where: { id: chargeId },
				data: { status: "BAIXADO" },
			});
			log.info("Boleto hibrido cancelled", {
				chargeId,
				txid,
				situacao: rawSituacao,
			});
			return;
		}

		log.info("Boleto hibrido still pending", {
			chargeId,
			txid,
			situacao: rawSituacao,
		});
	}

	// Re-enqueue for another check if still pending
	if (invoiceId) {
		const { enqueueSicrediBilling } = await import("@/lib/queue/producers");
		await enqueueSicrediBilling(
			{
				invoiceId,
				type: chargeType as "pix" | "boleto_hibrido",
				action: "check_status",
			},
			{ delay: 5 * 60 * 1000 },
		);
	}
}

async function getDb() {
	const { db } = await import("@/lib/db");
	return db;
}
