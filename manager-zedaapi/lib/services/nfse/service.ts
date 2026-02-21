// =============================================================================
// NFS-e Nacional — Main Orchestrator Service
// =============================================================================

import { gunzipSync } from "node:zlib";
import { db } from "@/lib/db";
import { createLogger } from "@/lib/logger";
import { loadCertificate } from "./certificate";
import { buildDpsXml } from "./xml-builder";
import { signDpsXml } from "./signer";
import {
	submitDps,
	queryNfse,
	cancelNfseRequest,
	fetchDanfsePdf,
	getDanfsePortalUrl,
} from "./client";
import type { NfseConfigData, NfseEmitResult, Tomador } from "./types";
import { validateTomadorForNfse, parseSefinErrorToFriendly } from "./types";
import { fetchCnpjData } from "@/lib/services/brasil-api/cnpj";

const log = createLogger("service:nfse");

/**
 * Get active NFS-e configuration from database.
 * Returns null if no active config exists.
 */
async function getActiveConfig(): Promise<NfseConfigData | null> {
	const config = await db.nfseConfig.findFirst({
		where: { active: true },
	});
	return config;
}

/**
 * Reserve a sequential DPS number for an invoice.
 * Uses atomic increment on NfseSequence to prevent race conditions.
 */
async function reserveDpsNumber(
	invoiceId: string,
	cnpj: string,
): Promise<number> {
	const currentYear = new Date().getFullYear();

	// Atomically get next number
	const seq = await db.nfseSequence.upsert({
		where: { cnpj },
		create: { cnpj, lastNumber: 1, year: currentYear },
		update: { lastNumber: { increment: 1 } },
	});

	// If year rolled over, reset
	if (seq.year < currentYear) {
		const updated = await db.nfseSequence.update({
			where: { cnpj },
			data: { lastNumber: 1, year: currentYear },
		});
		return updated.lastNumber;
	}

	return seq.lastNumber;
}

/**
 * Upload signed DPS XML to S3 and return the URL.
 */
async function uploadDpsXmlToS3(
	signedXml: string,
	invoiceId: string,
): Promise<string> {
	const { uploadFile } = await import("@/lib/services/storage/s3-client");
	const buffer = Buffer.from(signedXml, "utf-8");
	const key = `nfse/xml/${invoiceId}-dps.xml`;
	return await uploadFile(key, buffer, "application/xml");
}

/**
 * Decompress and save the official NFS-e XML from SEFIN response to S3.
 */
async function saveResponseXmlToS3(
	nfseXmlGZipB64: string,
	invoiceId: string,
): Promise<string> {
	const { uploadFile } = await import("@/lib/services/storage/s3-client");
	const compressed = Buffer.from(nfseXmlGZipB64, "base64");
	const xmlBuffer = gunzipSync(compressed);
	const key = `nfse/xml/${invoiceId}-nfse.xml`;
	return await uploadFile(key, xmlBuffer, "application/xml");
}

/**
 * Upload DANFSE PDF to S3 and return the URL.
 */
async function uploadPdfToS3(
	pdfBuffer: Buffer,
	invoiceId: string,
): Promise<string> {
	const { uploadFile } = await import("@/lib/services/storage/s3-client");
	const key = `nfse/pdf/${invoiceId}.pdf`;
	return await uploadFile(key, pdfBuffer, "application/pdf");
}

/**
 * Emit NFS-e for a given invoice.
 * This is the main entry point called by the worker/action.
 */
export async function emitNfse(invoiceId: string): Promise<NfseEmitResult> {
	// 1. Load config
	const config = await getActiveConfig();
	if (!config) {
		return {
			success: false,
			error: "NFS-e nao configurada. Configure no painel admin.",
		};
	}

	// 2. Load invoice with user data
	const invoice = await db.invoice.findUnique({
		where: { id: invoiceId },
		include: {
			user: {
				select: {
					id: true,
					name: true,
					email: true,
					cpfCnpj: true,
					phone: true,
					address: true,
					city: true,
					state: true,
					zipCode: true,
				},
			},
		},
	});

	if (!invoice) {
		return { success: false, error: `Invoice ${invoiceId} nao encontrada` };
	}

	// Skip if already issued (only if we have a valid protocol)
	if (invoice.nfseStatus === "ISSUED" && invoice.nfseProtocol) {
		return {
			success: true,
			chaveAcesso: invoice.nfseProtocol,
			numero: invoice.nfseNumber || undefined,
			pdfUrl: invoice.nfsePdfUrl || undefined,
		};
	}

	// Reset corrupted ISSUED state (no protocol) back to PENDING for re-emission
	if (invoice.nfseStatus === "ISSUED" && !invoice.nfseProtocol) {
		await db.invoice.update({
			where: { id: invoiceId },
			data: { nfseStatus: "PENDING", nfseError: null },
		});
	}

	// Skip if zero amount (nothing to invoice)
	const amountCents = Number(invoice.amount) * 100;
	if (!amountCents || amountCents <= 0) {
		await db.invoice.update({
			where: { id: invoiceId },
			data: {
				nfseStatus: "CANCELLED",
				nfseError: "Valor zero — NFS-e nao aplicavel",
			},
		});
		return { success: false, error: "Valor zero" };
	}

	// 3. Build tomador data from user
	const cpfCnpj = invoice.user.cpfCnpj || "";
	const cpfCnpjDigits = cpfCnpj.replace(/\D/g, "");
	const isPJ = cpfCnpjDigits.length === 14;

	const tomador: Tomador = {
		cpfCnpj,
		nome: invoice.user.name,
		email: invoice.user.email,
		phone: invoice.user.phone || undefined,
	};

	// Build address if available
	if (
		invoice.user.address &&
		invoice.user.city &&
		invoice.user.state &&
		invoice.user.zipCode
	) {
		tomador.endereco = {
			logradouro: invoice.user.address,
			numero: "S/N",
			bairro: "Nao informado",
			codigoMunicipio: "",
			cidade: invoice.user.city,
			uf: invoice.user.state,
			cep: invoice.user.zipCode,
		};
	}

	// If codigoMunicipio is empty but we have a CNPJ, try to resolve via BrasilAPI
	if (tomador.endereco && !tomador.endereco.codigoMunicipio && isPJ) {
		try {
			const cnpjInfo = await fetchCnpjData(cpfCnpjDigits);
			if (cnpjInfo?.codigoMunicipio) {
				tomador.endereco.codigoMunicipio = cnpjInfo.codigoMunicipio;
			}
		} catch {
			// CNPJ lookup failed, validation will catch missing code
		}
	}

	// Validate ALL required tomador fields before proceeding
	const validationError = validateTomadorForNfse(tomador);
	if (validationError) {
		await db.invoice.update({
			where: { id: invoiceId },
			data: {
				nfseStatus: "ERROR",
				nfseError: validationError,
			},
		});
		return { success: false, error: validationError };
	}

	// 4. Mark as processing
	await db.invoice.update({
		where: { id: invoiceId },
		data: { nfseStatus: "PROCESSING" },
	});

	try {
		// 5. Load certificate
		const { certPem, keyPem } = await loadCertificate(config);

		// 6. Reserve atomic DPS number
		const dpsNumber = await reserveDpsNumber(invoiceId, config.cnpj);
		const tipoPessoa = isPJ ? "PJ" : "PF";
		const codigoUsado = isPJ
			? config.codigoServico
			: config.codigoServicoPf;
		log.info("Preparing DPS submission", {
			invoiceId,
			dpsNumber,
			amountCents,
			tomadorCpfCnpj: tomador.cpfCnpj,
			tipoPessoa,
			codigoServico: codigoUsado,
			tomadorNome: tomador.nome,
		});

		// 7. Build DPS XML
		const dpsXml = buildDpsXml({
			invoiceId,
			amountCents,
			tomador,
			config,
			dpsNumber: String(dpsNumber),
		});

		// 8. Sign XML (XMLDSIG)
		const signedXml = signDpsXml(dpsXml, certPem, keyPem);

		// 9. Upload signed DPS XML to S3
		const dpsXmlUrl = await uploadDpsXmlToS3(signedXml, invoiceId);

		// 10. Submit to SEFIN
		log.info("Submitting to SEFIN", { invoiceId });
		const response = await submitDps(signedXml, certPem, keyPem, config);
		log.info("SEFIN response received", {
			invoiceId,
			response,
		});

		// 11. Process response
		if (response.chaveAcesso) {
			// --- Save official NFS-e XML from SEFIN response ---
			let nfseXmlUrl = dpsXmlUrl; // fallback to DPS XML
			if (response.nfseXmlGZipB64) {
				try {
					nfseXmlUrl = await saveResponseXmlToS3(
						response.nfseXmlGZipB64,
						invoiceId,
					);
					log.info("Official NFS-e XML saved to S3", { invoiceId });
				} catch (xmlError) {
					log.error("Failed to save NFS-e XML", {
						invoiceId,
						error:
							xmlError instanceof Error
								? xmlError.message
								: "unknown",
					});
				}
			}

			// --- Fetch and save DANFSE PDF ---
			let pdfUrl: string | null = null;
			try {
				const pdfBuffer = await fetchDanfsePdf(
					response.chaveAcesso,
					certPem,
					keyPem,
					config,
				);
				if (pdfBuffer) {
					pdfUrl = await uploadPdfToS3(pdfBuffer, invoiceId);
					log.info("DANFSE PDF saved to S3", {
						invoiceId,
						sizeBytes: pdfBuffer.length,
					});
				} else {
					log.warn("DANFSE PDF not available yet", { invoiceId });
				}
			} catch (pdfError) {
				log.warn("Failed to fetch DANFSE PDF", {
					invoiceId,
					error:
						pdfError instanceof Error
							? pdfError.message
							: "unknown",
				});
			}

			const danfseUrl =
				pdfUrl ||
				getDanfsePortalUrl(response.chaveAcesso, config.ambiente);

			await db.invoice.update({
				where: { id: invoiceId },
				data: {
					nfseProtocol: response.chaveAcesso,
					nfseNumber: response.numero || null,
					nfseStatus: "ISSUED",
					nfseIssuedAt: response.dataEmissao
						? new Date(response.dataEmissao)
						: new Date(),
					nfsePdfUrl: danfseUrl,
					nfseXmlUrl: nfseXmlUrl,
					nfseError: null,
				},
			});

			return {
				success: true,
				chaveAcesso: response.chaveAcesso,
				numero: response.numero,
				danfseUrl,
				pdfUrl: pdfUrl || undefined,
				xmlUrl: nfseXmlUrl,
				dpsXmlUrl,
			};
		}

		// SEFIN returned error (4xx — non-retryable)
		const friendlyError = parseSefinErrorToFriendly(
			response.mensagemStatus,
		);
		await db.invoice.update({
			where: { id: invoiceId },
			data: {
				nfseStatus: "ERROR",
				nfseError: friendlyError,
			},
		});

		log.error("SEFIN rejected submission", {
			invoiceId,
			mensagemStatus: response.mensagemStatus,
		});
		return { success: false, error: friendlyError };
	} catch (error) {
		const rawMessage =
			error instanceof Error ? error.message : "Erro desconhecido";
		const friendlyMessage = parseSefinErrorToFriendly(rawMessage);

		await db.invoice.update({
			where: { id: invoiceId },
			data: {
				nfseStatus: "ERROR",
				nfseError: friendlyMessage,
			},
		});

		// Re-throw 5xx errors for retry
		if (rawMessage.includes("SEFIN API error")) {
			throw error;
		}

		return { success: false, error: friendlyMessage };
	}
}

/**
 * Cancel an issued NFS-e.
 */
export async function cancelNfse(
	invoiceId: string,
	motivo: string,
): Promise<NfseEmitResult> {
	const config = await getActiveConfig();
	if (!config) {
		return { success: false, error: "NFS-e nao configurada" };
	}

	const invoice = await db.invoice.findUnique({
		where: { id: invoiceId },
	});

	if (!invoice || !invoice.nfseProtocol) {
		return {
			success: false,
			error: "NFS-e nao encontrada para este invoice",
		};
	}

	if (invoice.nfseStatus !== "ISSUED") {
		return {
			success: false,
			error: "So e possivel cancelar NFS-e emitidas",
		};
	}

	try {
		const { certPem, keyPem } = await loadCertificate(config);

		// Sanitize motivo to prevent XML injection
		const sanitizedMotivo = motivo.replace(/[<>&"']/g, (c) => {
			const map: Record<string, string> = {
				"<": "&lt;",
				">": "&gt;",
				"&": "&amp;",
				'"': "&quot;",
				"'": "&apos;",
			};
			return map[c] || c;
		});

		const response = await cancelNfseRequest(
			invoice.nfseProtocol,
			sanitizedMotivo,
			certPem,
			keyPem,
			config,
		);

		if (response.codigoStatus === 200 || response.codigoStatus === 0) {
			await db.invoice.update({
				where: { id: invoiceId },
				data: {
					nfseStatus: "CANCELLED",
					nfseCanceledAt: new Date(),
					nfseError: `Cancelada: ${motivo}`,
				},
			});

			return { success: true };
		}

		return { success: false, error: response.mensagemStatus };
	} catch (error) {
		const message =
			error instanceof Error ? error.message : "Erro ao cancelar NFS-e";
		return { success: false, error: message };
	}
}

/**
 * Query and update NFS-e status for a given invoice.
 * Also fetches and saves DANFSE PDF if not yet available.
 */
export async function queryNfseStatus(invoiceId: string): Promise<void> {
	const config = await getActiveConfig();
	if (!config) return;

	const invoice = await db.invoice.findUnique({
		where: { id: invoiceId },
	});

	if (!invoice || !invoice.nfseProtocol) return;
	if (invoice.nfseStatus === "ISSUED" || invoice.nfseStatus === "CANCELLED") {
		// If already issued but missing PDF, try to fetch it
		if (
			invoice.nfseStatus === "ISSUED" &&
			invoice.nfsePdfUrl &&
			invoice.nfsePdfUrl.includes("nfse.gov.br")
		) {
			await tryFetchMissingPdf(invoice.id, invoice.nfseProtocol, config);
		}
		return;
	}

	try {
		const { certPem, keyPem } = await loadCertificate(config);
		const response = await queryNfse(
			invoice.nfseProtocol,
			certPem,
			keyPem,
			config,
		);

		if (response.chaveAcesso && response.numero) {
			// Save official XML if available
			let nfseXmlUrl = invoice.nfseXmlUrl;
			if (response.nfseXmlGZipB64) {
				try {
					nfseXmlUrl = await saveResponseXmlToS3(
						response.nfseXmlGZipB64,
						invoiceId,
					);
				} catch {
					// Keep existing XML URL
				}
			}

			// Fetch and save PDF
			let pdfUrl = getDanfsePortalUrl(
				response.chaveAcesso,
				config.ambiente,
			);
			try {
				const pdfBuffer = await fetchDanfsePdf(
					response.chaveAcesso,
					certPem,
					keyPem,
					config,
				);
				if (pdfBuffer) {
					pdfUrl = await uploadPdfToS3(pdfBuffer, invoiceId);
				}
			} catch {
				// Use portal URL as fallback
			}

			await db.invoice.update({
				where: { id: invoiceId },
				data: {
					nfseNumber: response.numero,
					nfseStatus: "ISSUED",
					nfseIssuedAt: response.dataEmissao
						? new Date(response.dataEmissao)
						: new Date(),
					nfsePdfUrl: pdfUrl,
					nfseXmlUrl: nfseXmlUrl,
				},
			});
		}
	} catch (error) {
		log.error("Query status error", {
			invoiceId,
			error: error instanceof Error ? error.message : String(error),
		});
	}
}

/**
 * Try to fetch and save the DANFSE PDF for an issued NFS-e that only has a portal URL.
 */
async function tryFetchMissingPdf(
	invoiceId: string,
	chaveAcesso: string,
	config: NfseConfigData,
): Promise<void> {
	try {
		const { certPem, keyPem } = await loadCertificate(config);
		const pdfBuffer = await fetchDanfsePdf(
			chaveAcesso,
			certPem,
			keyPem,
			config,
		);
		if (pdfBuffer) {
			const pdfUrl = await uploadPdfToS3(pdfBuffer, invoiceId);
			await db.invoice.update({
				where: { id: invoiceId },
				data: { nfsePdfUrl: pdfUrl },
			});
			log.info("Backfilled DANFSE PDF", { invoiceId });
		}
	} catch (error) {
		log.warn("Failed to backfill PDF", {
			invoiceId,
			error: error instanceof Error ? error.message : "unknown",
		});
	}
}

/**
 * Get active NFS-e configuration (public, for admin UI).
 */
export { getActiveConfig };
