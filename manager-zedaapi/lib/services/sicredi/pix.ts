import { sicrediPix } from "./client";
import { formatSicrediAmount } from "./utils";

// =============================================================================
// SICREDI COB — Cobranças PIX Imediatas (via API PIX mTLS)
// =============================================================================

interface SicrediCobResponse {
	txid: string;
	revisao: number;
	calendario: {
		criacao: string;
		expiracao: number;
	};
	devedor?: {
		nome?: string;
		cpf?: string;
		cnpj?: string;
	};
	valor: {
		original: string;
	};
	chave: string;
	status: string; // ATIVA, CONCLUIDA, REMOVIDA_PELO_USUARIO_RECEBEDOR, REMOVIDA_PELO_PSP
	pixCopiaECola?: string;
	location?: string;
	loc?: {
		id: number;
		location: string;
	};
	pix?: Array<{
		endToEndId: string;
		txid: string;
		valor: string;
		horario: string;
	}>;
}

interface CreatePixChargeOptions {
	amountCents: number;
	description?: string;
	expirationSeconds?: number;
	devedor?: {
		nome: string;
		cpf?: string;
		cnpj?: string;
	};
}

/**
 * Create PIX charge without pre-defined txid (Sicredi generates it).
 * POST /api/v2/cob
 */
export async function createPixCharge(
	opts: CreatePixChargeOptions,
): Promise<SicrediCobResponse> {
	const pixKey = process.env.SICREDI_PIX_KEY;
	if (!pixKey) throw new Error("SICREDI_PIX_KEY not configured");

	const body: Record<string, unknown> = {
		calendario: {
			expiracao: opts.expirationSeconds || 3600, // 1h default
		},
		valor: {
			original: formatSicrediAmount(opts.amountCents),
		},
		chave: pixKey,
	};

	if (opts.description) {
		body.solicitacaoPagador = opts.description;
	}

	if (opts.devedor) {
		body.devedor = opts.devedor;
	}

	return sicrediPix.request<SicrediCobResponse>("POST", "/api/v2/cob", body);
}

/**
 * Create PIX charge with pre-defined txid (for idempotency).
 * PUT /api/v2/cob/{txid}
 */
export async function createPixChargeWithTxid(
	txid: string,
	opts: CreatePixChargeOptions,
): Promise<SicrediCobResponse> {
	const pixKey = process.env.SICREDI_PIX_KEY;
	if (!pixKey) throw new Error("SICREDI_PIX_KEY not configured");

	const body: Record<string, unknown> = {
		calendario: {
			expiracao: opts.expirationSeconds || 3600,
		},
		valor: {
			original: formatSicrediAmount(opts.amountCents),
		},
		chave: pixKey,
	};

	if (opts.description) {
		body.solicitacaoPagador = opts.description;
	}

	if (opts.devedor) {
		body.devedor = opts.devedor;
	}

	return sicrediPix.request<SicrediCobResponse>(
		"PUT",
		`/api/v2/cob/${txid}`,
		body,
	);
}

/**
 * Get PIX charge details.
 * GET /api/v2/cob/{txid}
 */
export async function getPixCharge(txid: string): Promise<SicrediCobResponse> {
	return sicrediPix.request<SicrediCobResponse>("GET", `/api/v2/cob/${txid}`);
}

/**
 * Cancel a PIX charge.
 * Verifies the charge is ATIVA before attempting cancellation.
 * PATCH /api/v2/cob/{txid}
 */
export async function cancelPixCharge(
	txid: string,
): Promise<SicrediCobResponse> {
	const charge = await getPixCharge(txid);
	if (charge.status !== "ATIVA") {
		throw new Error(
			`Cannot cancel PIX charge ${txid}: status is "${charge.status}", expected "ATIVA"`,
		);
	}

	return sicrediPix.request<SicrediCobResponse>(
		"PATCH",
		`/api/v2/cob/${txid}`,
		{
			status: "REMOVIDA_PELO_USUARIO_RECEBEDOR",
		},
	);
}

export type { SicrediCobResponse, CreatePixChargeOptions };
