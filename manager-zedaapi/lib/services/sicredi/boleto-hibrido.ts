import { sicrediParceiro } from "./client";

// =============================================================================
// SICREDI BOLETO HIBRIDO — API Parceiro (api-parceiro.sicredi.com.br)
// POST /cobranca/boleto/v1/boletos com tipoCobranca: "HIBRIDO"
// Gera boleto com QR Code PIX embutido (pague via PIX ou boleto)
// =============================================================================

// ---- Response types ----

interface SicrediBoletoResponse {
	nossoNumero: string;
	codigoBarras: string;
	linhaDigitavel: string;
	cooperativa: string;
	posto: string;
	/** QR Code PIX copy-paste string (HIBRIDO only) */
	qrCode?: string;
	/** PIX URL (HIBRIDO only) */
	urlPix?: string;
	/** Whether this is a hybrid boleto */
	hibrido?: boolean;
	/** Sicredi may return additional fields */
	[key: string]: unknown;
}

/**
 * Boleto query response includes `situacao` field from API Parceiro.
 * Known values: "EM_ABERTO", "LIQUIDADO", "BAIXADO", "VENCIDO"
 */
interface SicrediBoletoQueryResponse extends SicrediBoletoResponse {
	/** Boleto status from API Parceiro */
	situacao?: string;
}

// ---- Request types ----

interface CreateBoletoHibridoOptions {
	amountCents: number;
	dueDate: string; // YYYY-MM-DD
	description?: string;
	pagador: {
		nome: string;
		documento: string; // CPF (11 digits) or CNPJ (14 digits)
		tipoPessoa?: "PESSOA_FISICA" | "PESSOA_JURIDICA";
		cep: string; // 8 numeric digits (required by Sicredi API)
		cidade: string;
		endereco: string; // logradouro
		uf: string; // 2-letter state code
		telefone?: string;
		email?: string;
	};
	/** Our internal reference number */
	seuNumero: string;
	/** Days valid after due date (default: 30) */
	validadeAposVencimento?: number;
}

/**
 * Create Boleto Hibrido via API Parceiro.
 * POST /cobranca/boleto/v1/boletos
 */
export async function createBoletoHibrido(
	opts: CreateBoletoHibridoOptions,
): Promise<SicrediBoletoResponse> {
	const codigoBeneficiario = sicrediParceiro.getCodigoBeneficiario();
	if (!codigoBeneficiario) {
		throw new Error("SICREDI_BENEFICIARIO_CNPJ not configured");
	}

	// Detect pessoa fisica vs juridica based on document length
	const tipoPessoa =
		opts.pagador.tipoPessoa ||
		(opts.pagador.documento.length <= 11
			? "PESSOA_FISICA"
			: "PESSOA_JURIDICA");

	const body = {
		codigoBeneficiario,
		tipoCobranca: "HIBRIDO",
		dataVencimento: opts.dueDate,
		especieDocumento: "DUPLICATA_MERCANTIL_INDICACAO",
		seuNumero: opts.seuNumero,
		valor: opts.amountCents / 100, // API expects BRL, not cents
		validadeAposVencimento: opts.validadeAposVencimento ?? 30,
		pagador: {
			documento: opts.pagador.documento,
			nome: opts.pagador.nome,
			tipoPessoa,
			cep: opts.pagador.cep.replace(/\D/g, "").slice(0, 8),
			cidade: opts.pagador.cidade,
			endereco: opts.pagador.endereco,
			uf: opts.pagador.uf.toUpperCase().slice(0, 2),
			telefone: opts.pagador.telefone || "",
			email: opts.pagador.email || "",
		},
		// No discount, interest, or fine — omit these fields entirely.
		// Sicredi API rejects discount fields with value 0
		// ("deve ser superior a zero, ou nao informado").
		informativos: [
			(opts.description || "Assinatura Zé da API").slice(0, 80),
		],
		mensagens: [(opts.description || "Assinatura Zé da API").slice(0, 80)],
	};

	return sicrediParceiro.request<SicrediBoletoResponse>(
		"POST",
		"/cobranca/boleto/v1/boletos",
		body,
	);
}

/**
 * Query boleto status via API Parceiro.
 * GET /cobranca/boleto/v1/boletos?nossoNumero={nossoNumero}
 * Response includes `situacao` field: "EM_ABERTO", "LIQUIDADO", "BAIXADO", "VENCIDO"
 */
export async function getBoletoHibrido(
	nossoNumero: string,
): Promise<SicrediBoletoQueryResponse> {
	return sicrediParceiro.request<SicrediBoletoQueryResponse>(
		"GET",
		`/cobranca/boleto/v1/boletos?nossoNumero=${encodeURIComponent(nossoNumero)}`,
	);
}

/**
 * Check if a boleto has been paid (LIQUIDADO).
 * Maps API Parceiro situacao to our internal status.
 */
export function mapBoletoSituacao(response: SicrediBoletoQueryResponse): {
	isPaid: boolean;
	isCancelled: boolean;
	rawSituacao: string;
} {
	// Check `situacao` field (primary) and fallback to `status` or `situacaoTitulo`
	const situacao = (response.situacao ||
		(response as Record<string, unknown>).situacaoTitulo ||
		(response as Record<string, unknown>).status ||
		"") as string;

	const normalized = situacao.toUpperCase();

	return {
		isPaid: normalized === "LIQUIDADO" || normalized === "CONCLUIDA",
		isCancelled: normalized === "BAIXADO",
		rawSituacao: situacao,
	};
}

/**
 * Cancel (baixar) a boleto via API Parceiro.
 * PATCH /cobranca/boleto/v1/boletos/{nossoNumero}
 * Sets situacao to "BAIXADO" — the boleto becomes invalid for payment.
 */
export async function cancelBoletoHibrido(
	nossoNumero: string,
): Promise<SicrediBoletoQueryResponse> {
	return sicrediParceiro.request<SicrediBoletoQueryResponse>(
		"PATCH",
		`/cobranca/boleto/v1/boletos/${encodeURIComponent(nossoNumero)}`,
		{ situacao: "BAIXADO" },
	);
}

/**
 * Download boleto PDF via authenticated API Parceiro request.
 * Server-side only — requires valid Parceiro credentials.
 * GET /cobranca/boleto/v1/boletos/pdf?linhaDigitavel={linhaDigitavel}
 * @returns PDF file contents as ArrayBuffer
 */
export async function downloadBoletoHibridoPdf(
	linhaDigitavel: string,
): Promise<ArrayBuffer> {
	const token = await sicrediParceiro.getToken();
	const baseUrl =
		process.env.SICREDI_PARCEIRO_BASE_URL ||
		"https://api-parceiro.sicredi.com.br";
	const url = `${baseUrl}/cobranca/boleto/v1/boletos/pdf?linhaDigitavel=${encodeURIComponent(linhaDigitavel)}`;

	const response = await fetch(url, {
		method: "GET",
		headers: {
			Authorization: `Bearer ${token}`,
			"x-api-key": process.env.SICREDI_PARCEIRO_API_KEY || "",
			Accept: "application/pdf",
		},
	});

	if (!response.ok) {
		const errorText = await response.text();
		throw new Error(
			`Sicredi Parceiro PDF download error (${response.status}): ${errorText}`,
		);
	}

	return response.arrayBuffer();
}

export type {
	SicrediBoletoResponse,
	SicrediBoletoQueryResponse,
	CreateBoletoHibridoOptions,
};
