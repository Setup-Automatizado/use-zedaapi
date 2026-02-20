// =============================================================================
// NFS-e Nacional — HTTP Client (mTLS) for SEFIN API
// =============================================================================

import { gzipSync } from "node:zlib";
import type { NfseConfigData, NfseSefinResponse } from "./types";

const SEFIN_URLS = {
	HOMOLOGACAO: "https://sefin.producaorestrita.nfse.gov.br/SefinNacional",
	PRODUCAO: "https://sefin.nfse.gov.br/SefinNacional",
} as const;

const DANFSE_URLS = {
	HOMOLOGACAO: "https://adn.producaorestrita.nfse.gov.br/danfse",
	PRODUCAO: "https://adn.nfse.gov.br/danfse",
} as const;

/**
 * Submit a signed DPS XML to SEFIN Nacional API.
 * Compresses with GZip, encodes as Base64, sends via POST with mTLS.
 *
 * NOTE: The `tls` option on fetch() is a Bun-specific extension and is
 * not available in Node.js or other runtimes.
 */
export async function submitDps(
	signedXml: string,
	certPem: string,
	keyPem: string,
	config: NfseConfigData,
): Promise<NfseSefinResponse> {
	const baseUrl =
		SEFIN_URLS[config.ambiente as keyof typeof SEFIN_URLS] ||
		SEFIN_URLS.HOMOLOGACAO;

	// GZip compress the signed XML
	const compressed = gzipSync(Buffer.from(signedXml, "utf-8"));
	const base64Payload = compressed.toString("base64");

	const response = await fetch(`${baseUrl}/nfse`, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
			Accept: "application/json",
		},
		body: JSON.stringify({
			dpsXmlGZipB64: base64Payload,
		}),
		// Bun supports TLS options in fetch
		tls: {
			cert: certPem,
			key: keyPem,
		},
	} as RequestInit);

	if (!response.ok) {
		const errorBody = await response.text().catch(() => "");
		if (response.status >= 500) {
			// 5xx: SEFIN server error — throw for retry
			throw new Error(
				`SEFIN API error ${response.status}: ${errorBody || response.statusText}`,
			);
		}
		// 4xx: Client error (invalid data) — do NOT retry
		return {
			codigoStatus: response.status,
			mensagemStatus: `Erro SEFIN: ${errorBody || response.statusText}`,
		};
	}

	const data = (await response.json()) as NfseSefinResponse;
	return data;
}

/**
 * Query NFS-e status by chave de acesso.
 */
export async function queryNfse(
	chaveAcesso: string,
	certPem: string,
	keyPem: string,
	config: NfseConfigData,
): Promise<NfseSefinResponse> {
	const baseUrl =
		SEFIN_URLS[config.ambiente as keyof typeof SEFIN_URLS] ||
		SEFIN_URLS.HOMOLOGACAO;

	const response = await fetch(`${baseUrl}/nfse/${chaveAcesso}`, {
		method: "GET",
		headers: {
			Accept: "application/json",
		},
		tls: {
			cert: certPem,
			key: keyPem,
		},
	} as RequestInit);

	if (!response.ok) {
		const errorBody = await response.text().catch(() => "");
		throw new Error(
			`SEFIN query error ${response.status}: ${errorBody || response.statusText}`,
		);
	}

	return (await response.json()) as NfseSefinResponse;
}

/**
 * Cancel an issued NFS-e via Eventos API (pedRegEvento e101101).
 * Builds the event XML, signs it, gzips+base64, and POSTs to /nfse/{chave}/eventos.
 */
export async function cancelNfseRequest(
	chaveAcesso: string,
	motivo: string,
	certPem: string,
	keyPem: string,
	config: NfseConfigData,
): Promise<NfseSefinResponse> {
	const baseUrl =
		SEFIN_URLS[config.ambiente as keyof typeof SEFIN_URLS] ||
		SEFIN_URLS.HOMOLOGACAO;

	// Build cancel event XML
	const { buildCancelEventXml } = await import("./xml-builder");
	const eventXml = buildCancelEventXml({
		chaveAcesso,
		motivo,
		cnpjAutor: config.cnpj,
		ambiente: config.ambiente,
	});

	// Sign the event XML
	const { signEventXml } = await import("./signer");
	const signedXml = signEventXml(eventXml, certPem, keyPem);

	// GZip + Base64
	const compressed = gzipSync(Buffer.from(signedXml, "utf-8"));
	const base64Payload = compressed.toString("base64");

	console.log(`[nfse:cancel] Sending cancel event for chave=[REDACTED]`);

	const response = await fetch(`${baseUrl}/nfse/${chaveAcesso}/eventos`, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
			Accept: "application/json",
		},
		body: JSON.stringify({
			pedidoRegistroEventoXmlGZipB64: base64Payload,
		}),
		tls: {
			cert: certPem,
			key: keyPem,
		},
	} as RequestInit);

	if (!response.ok) {
		const errorBody = await response.text().catch(() => "");
		if (response.status >= 500) {
			throw new Error(
				`SEFIN cancel error ${response.status}: ${errorBody || response.statusText}`,
			);
		}
		return {
			codigoStatus: response.status,
			mensagemStatus: `Erro SEFIN cancelamento: ${errorBody || response.statusText}`,
		};
	}

	const data = await response.json();
	console.log("[nfse:cancel] SEFIN response:", JSON.stringify(data));

	// Success: SEFIN returns eventoXmlGZipB64 when event is registered
	if (data.eventoXmlGZipB64) {
		return {
			codigoStatus: 0,
			mensagemStatus: "Cancelamento registrado com sucesso",
		};
	}

	// Alternative success indicators
	if (data.chaveAcesso || data.nSeqEvento || data.cStat === 135) {
		return { codigoStatus: 0, mensagemStatus: "Cancelamento registrado" };
	}

	return {
		codigoStatus: data.cStat || response.status,
		mensagemStatus:
			data.xMotivo || data.mensagemStatus || JSON.stringify(data),
	};
}

/**
 * Fetch DANFSE PDF from SEFIN via mTLS.
 * Returns the raw PDF buffer, or null if the PDF is not yet available.
 */
export async function fetchDanfsePdf(
	chaveAcesso: string,
	certPem: string,
	keyPem: string,
	config: NfseConfigData,
): Promise<Buffer | null> {
	const danfseBaseUrl =
		DANFSE_URLS[config.ambiente as keyof typeof DANFSE_URLS] ||
		DANFSE_URLS.HOMOLOGACAO;

	const url = `${danfseBaseUrl}/${chaveAcesso}`;
	console.log(`[nfse:danfse] Fetching DANFSE PDF from ${url}`);

	try {
		const response = await fetch(url, {
			method: "GET",
			headers: {
				Accept: "application/pdf",
			},
			tls: {
				cert: certPem,
				key: keyPem,
			},
		} as RequestInit);

		if (!response.ok) {
			console.warn(
				`[nfse:danfse] DANFSE fetch failed: ${response.status} ${response.statusText}`,
			);
			return null;
		}

		const contentType = response.headers.get("content-type") || "";
		const arrayBuffer = await response.arrayBuffer();
		const buffer = Buffer.from(arrayBuffer);

		// Validate we got something reasonable (at least 1KB for a PDF)
		if (buffer.length < 1024) {
			console.warn(
				`[nfse:danfse] DANFSE response too small (${buffer.length} bytes), may not be valid PDF`,
			);
		}

		console.log(
			`[nfse:danfse] DANFSE PDF fetched successfully | size=${buffer.length} bytes, contentType=${contentType}`,
		);
		return buffer;
	} catch (error) {
		console.warn(
			`[nfse:danfse] Failed to fetch DANFSE PDF: ${error instanceof Error ? error.message : "unknown error"}`,
		);
		return null;
	}
}

/**
 * Get the public DANFSE portal URL (requires mTLS — for reference only).
 */
export function getDanfsePortalUrl(
	chaveAcesso: string,
	ambiente: string = "HOMOLOGACAO",
): string {
	const danfseBaseUrl =
		DANFSE_URLS[ambiente as keyof typeof DANFSE_URLS] ||
		DANFSE_URLS.HOMOLOGACAO;
	return `${danfseBaseUrl}/${chaveAcesso}`;
}
