import { timingSafeEqual } from "crypto";
import { NextRequest, NextResponse } from "next/server";
import { db } from "@/lib/db";
import { handleSicrediPayment } from "@/server/services/billing-service";
import { parseSicrediAmount } from "@/lib/services/sicredi/utils";

export const dynamic = "force-dynamic";

const WEBHOOK_SECRET = process.env.SICREDI_WEBHOOK_SECRET;

interface SicrediPixCallback {
	endToEndId: string;
	txid: string;
	valor: string;
	horario: string;
	infoPagador?: string;
	componentesValor?: {
		original?: { valor: string };
	};
}

interface SicrediWebhookBody {
	pix: SicrediPixCallback[];
}

/**
 * POST /api/webhooks/sicredi
 *
 * Receives PIX payment confirmations from Sicredi (BACEN).
 * Handles both COB (PIX direto) and COBV (Boleto hibrido pago via PIX).
 */
export async function POST(request: NextRequest) {
	// Webhook secret validation
	if (!WEBHOOK_SECRET) {
		console.error(
			"[sicredi-webhook] SICREDI_WEBHOOK_SECRET not configured",
		);
		return NextResponse.json(
			{ error: "Webhook not configured" },
			{ status: 503 },
		);
	}

	const providedSecret = request.headers.get("x-webhook-secret");
	const isValid =
		providedSecret &&
		WEBHOOK_SECRET &&
		providedSecret.length === WEBHOOK_SECRET.length &&
		timingSafeEqual(
			Buffer.from(providedSecret),
			Buffer.from(WEBHOOK_SECRET),
		);
	if (!isValid) {
		console.warn("[sicredi-webhook] Invalid webhook secret");
		return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
	}

	let body: SicrediWebhookBody;

	try {
		body = (await request.json()) as SicrediWebhookBody;
	} catch {
		return NextResponse.json(
			{ error: "Invalid JSON body" },
			{ status: 400 },
		);
	}

	if (!body.pix || !Array.isArray(body.pix)) {
		return NextResponse.json(
			{ error: "Missing pix array" },
			{ status: 400 },
		);
	}

	const results: Array<{ txid: string; status: string }> = [];

	for (const pixPayment of body.pix) {
		const { txid } = pixPayment;

		if (!txid) {
			results.push({ txid: "unknown", status: "skipped_no_txid" });
			continue;
		}

		try {
			// Atomic guard: only process if status is still not CONCLUIDA
			const result = await db.$transaction(async (tx) => {
				const { count } = await tx.sicrediCharge.updateMany({
					where: { txid, status: { not: "CONCLUIDA" } },
					data: {
						status: "CONCLUIDA",
						paidAt: new Date(pixPayment.horario),
					},
				});

				if (count === 0) {
					// Either txid doesn't exist or already processed
					const exists = await tx.sicrediCharge.findUnique({
						where: { txid },
						select: { status: true },
					});

					if (!exists) {
						// Boleto hibrido fallback: txid from BACEN may differ from nossoNumero
						const paidCents = parseSicrediAmount(pixPayment.valor);
						if (paidCents > 0) {
							const boletoMatch =
								await tx.sicrediCharge.findFirst({
									where: {
										type: "boleto_hibrido",
										status: { not: "CONCLUIDA" },
										amount: paidCents / 100,
									},
									orderBy: { createdAt: "desc" },
								});

							if (boletoMatch) {
								console.log(
									`[sicredi-webhook] Boleto hibrido fallback match: txid=${txid} -> chargeId=${boletoMatch.id}`,
								);

								await tx.sicrediCharge.update({
									where: { id: boletoMatch.id },
									data: {
										status: "CONCLUIDA",
										paidAt: new Date(pixPayment.horario),
										txid,
									},
								});

								return "processed_boleto_fallback";
							}
						}

						console.warn(`[sicredi-webhook] Unknown txid: ${txid}`);
						return "unknown_txid";
					}

					return "already_processed";
				}

				// Validate payment amount
				const charge = await tx.sicrediCharge.findUnique({
					where: { txid },
					select: { id: true, amount: true },
				});

				if (charge) {
					const paidCents = parseSicrediAmount(pixPayment.valor);
					const expectedCents = Math.round(
						Number(charge.amount) * 100,
					);

					if (paidCents > 0 && paidCents !== expectedCents) {
						console.error(
							`[sicredi-webhook] Amount mismatch for txid=${txid}: expected=${expectedCents}, received=${paidCents}`,
						);

						// Revert status
						await tx.sicrediCharge.update({
							where: { id: charge.id },
							data: { status: "ATIVA" },
						});
						return "amount_mismatch";
					}
				}

				return "processed";
			});

			// Record webhook event only for non-duplicate results
			if (result !== "already_processed") {
				await db.sicrediWebhookEvent.create({
					data: {
						txid,
						type: "pix_payment",
						status:
							result === "processed" ||
							result === "processed_boleto_fallback"
								? "processing"
								: result,
						payload: pixPayment as object,
					},
				});
			}

			// Handle payment activation outside the transaction
			if (
				result === "processed" ||
				result === "processed_boleto_fallback"
			) {
				await handleSicrediPayment(txid, new Date(pixPayment.horario));
			}

			results.push({ txid, status: result });
		} catch (error) {
			console.error(
				`[sicredi-webhook] Error processing txid=${txid}:`,
				error,
			);
			results.push({ txid, status: "error" });
		}
	}

	return NextResponse.json({ received: true, results });
}

/**
 * GET /api/webhooks/sicredi
 *
 * Sicredi may send a GET request to verify the webhook URL.
 * Must respond with 200.
 */
export async function GET() {
	return NextResponse.json({ status: "ok" });
}
