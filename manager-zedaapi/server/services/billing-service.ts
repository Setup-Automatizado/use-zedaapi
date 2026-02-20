"use server";

import { db } from "@/lib/db";
import { createPixChargeWithTxid } from "@/lib/services/sicredi/pix";
import { createBoletoHibrido } from "@/lib/services/sicredi/boleto-hibrido";
import { generateTxid, calculateDueDate } from "@/lib/services/sicredi/utils";

/**
 * Create a Sicredi charge (PIX or Boleto Hibrido) for an invoice.
 */
export async function createSicrediCharge(
	invoiceId: string,
	type: "pix" | "boleto_hibrido",
) {
	const invoice = await db.invoice.findUnique({
		where: { id: invoiceId },
		include: {
			user: {
				select: {
					id: true,
					name: true,
					email: true,
					cpfCnpj: true,
					address: true,
					city: true,
					state: true,
					zipCode: true,
				},
			},
			subscription: {
				select: { id: true },
			},
		},
	});

	if (!invoice) {
		throw new Error("Fatura não encontrada");
	}

	if (invoice.status === "paid") {
		throw new Error("Fatura já foi paga");
	}

	const amountCents = Math.round(Number(invoice.amount) * 100);

	if (type === "pix") {
		const txid = generateTxid(invoice.id);
		const pixResponse = await createPixChargeWithTxid(txid, {
			amountCents,
			description: `Assinatura Zé da API - Fatura ${invoice.id.slice(0, 8)}`,
			expirationSeconds: 3600, // 1 hour
			devedor: invoice.user.cpfCnpj
				? {
						nome: invoice.user.name,
						...(invoice.user.cpfCnpj.length <= 11
							? { cpf: invoice.user.cpfCnpj }
							: { cnpj: invoice.user.cpfCnpj }),
					}
				: undefined,
		});

		const charge = await db.sicrediCharge.create({
			data: {
				userId: invoice.user.id,
				invoiceId: invoice.id,
				subscriptionId: invoice.subscription?.id,
				type: "COB",
				txid: pixResponse.txid,
				pixCopiaECola: pixResponse.pixCopiaECola || null,
				amount: invoice.amount,
				status: "ATIVA",
				expiresAt: new Date(Date.now() + 3600 * 1000),
			},
		});

		return {
			chargeId: charge.id,
			txid: pixResponse.txid,
			pixCopiaECola: pixResponse.pixCopiaECola,
			qrCodeUrl: pixResponse.location,
			expiresAt: charge.expiresAt,
		};
	}

	// Boleto Hibrido
	if (
		!invoice.user.cpfCnpj ||
		!invoice.user.zipCode ||
		!invoice.user.city ||
		!invoice.user.state
	) {
		throw new Error(
			"Dados cadastrais incompletos. CPF/CNPJ, CEP, cidade e estado são obrigatórios para boleto.",
		);
	}

	const dueDate = calculateDueDate(3);

	const boletoResponse = await createBoletoHibrido({
		amountCents,
		dueDate,
		description: `Assinatura Zé da API - Fatura ${invoice.id.slice(0, 8)}`,
		pagador: {
			nome: invoice.user.name,
			documento: invoice.user.cpfCnpj.replace(/\D/g, ""),
			cep: invoice.user.zipCode.replace(/\D/g, ""),
			cidade: invoice.user.city,
			endereco: invoice.user.address || "Nao informado",
			uf: invoice.user.state,
			email: invoice.user.email,
		},
		seuNumero: invoice.id.slice(0, 15),
	});

	const charge = await db.sicrediCharge.create({
		data: {
			userId: invoice.user.id,
			invoiceId: invoice.id,
			subscriptionId: invoice.subscription?.id,
			type: "COBV",
			txid: boletoResponse.nossoNumero,
			codigoBarras: boletoResponse.codigoBarras,
			linhaDigitavel: boletoResponse.linhaDigitavel,
			pixCopiaECola: boletoResponse.qrCode || null,
			amount: invoice.amount,
			status: "ATIVA",
			expiresAt: new Date(dueDate),
		},
	});

	return {
		chargeId: charge.id,
		nossoNumero: boletoResponse.nossoNumero,
		codigoBarras: boletoResponse.codigoBarras,
		linhaDigitavel: boletoResponse.linhaDigitavel,
		pixCopiaECola: boletoResponse.qrCode,
		expiresAt: charge.expiresAt,
	};
}

/**
 * Handle successful Stripe payment (called from webhook).
 */
export async function handleStripePayment(
	stripeSubscriptionId: string,
	stripeInvoiceId: string,
	amountPaid: number,
	currency: string,
	invoiceUrl: string | null,
) {
	const subscription = await db.subscription.findUnique({
		where: { stripeSubscriptionId },
		select: { id: true, userId: true },
	});

	if (!subscription) {
		console.warn(`[billing] No subscription for ${stripeSubscriptionId}`);
		return;
	}

	// Create or update the invoice
	await db.invoice.upsert({
		where: { stripeInvoiceId },
		create: {
			userId: subscription.userId,
			subscriptionId: subscription.id,
			stripeInvoiceId,
			amount: amountPaid / 100,
			currency: currency.toUpperCase(),
			status: "paid",
			paidAt: new Date(),
			pdfUrl: invoiceUrl,
			paymentMethod: "stripe",
		},
		update: {
			status: "paid",
			paidAt: new Date(),
			pdfUrl: invoiceUrl,
		},
	});
}

/**
 * Handle successful Sicredi PIX payment (called from webhook or status check).
 * Activates the subscription when the charge is paid.
 *
 * The caller (webhook route or billing worker) is responsible for updating
 * the SicrediCharge status to "CONCLUIDA". This function only handles
 * invoice update and subscription activation.
 *
 * @param txid - The transaction ID
 * @param paidAt - The actual payment timestamp from BACEN (horario field)
 */
export async function handleSicrediPayment(txid: string, paidAt?: Date) {
	const charge = await db.sicrediCharge.findUnique({
		where: { txid },
		include: {
			invoice: true,
			subscription: true,
		},
	});

	if (!charge) {
		console.warn(`[billing] No Sicredi charge for txid: ${txid}`);
		return;
	}

	const paymentDate = paidAt || charge.paidAt || new Date();

	// Update invoice
	if (charge.invoice && charge.invoice.status !== "paid") {
		await db.invoice.update({
			where: { id: charge.invoice.id },
			data: {
				status: "paid",
				paidAt: paymentDate,
			},
		});
	}

	// Activate subscription if it was pending
	if (charge.subscription && charge.subscription.status !== "active") {
		const now = new Date();
		const periodEnd = new Date(now);
		if (charge.subscription.currentPeriodEnd > now) {
			// Keep existing period end
		} else {
			periodEnd.setMonth(periodEnd.getMonth() + 1);
		}

		await db.subscription.update({
			where: { id: charge.subscription.id },
			data: {
				status: "active",
				currentPeriodStart: now,
				currentPeriodEnd:
					charge.subscription.currentPeriodEnd > now
						? charge.subscription.currentPeriodEnd
						: periodEnd,
			},
		});
	}
}

/**
 * Get invoices for a user.
 */
export async function getInvoices(userId: string, page = 1, pageSize = 20) {
	const [invoices, total] = await Promise.all([
		db.invoice.findMany({
			where: { userId },
			include: {
				subscription: {
					select: {
						plan: { select: { name: true, slug: true } },
					},
				},
				sicrediCharge: {
					select: {
						type: true,
						status: true,
						pixCopiaECola: true,
						linhaDigitavel: true,
					},
				},
			},
			orderBy: { createdAt: "desc" },
			skip: (page - 1) * pageSize,
			take: pageSize,
		}),
		db.invoice.count({ where: { userId } }),
	]);

	return { invoices, total, page, pageSize };
}

/**
 * Get a single invoice.
 */
export async function getInvoice(invoiceId: string) {
	return db.invoice.findUnique({
		where: { id: invoiceId },
		include: {
			subscription: {
				select: {
					plan: true,
				},
			},
			sicrediCharge: true,
		},
	});
}
