// =============================================================================
// NFS-e Nacional — Type Definitions
// =============================================================================

/** Ambiente de emissao */
export type NfseAmbiente = "HOMOLOGACAO" | "PRODUCAO";

/** Status da NFS-e no invoice */
export type NfseInvoiceStatus =
	| "PENDING"
	| "PROCESSING"
	| "ISSUED"
	| "ERROR"
	| "CANCELLED";

/** Dados do prestador de servico (empresa emissora) */
export interface Prestador {
	cnpj: string;
	inscricaoMunicipal: string;
	codigoMunicipio: string;
	uf: string;
}

/** Dados do tomador de servico (contribuidor) */
export interface Tomador {
	cpfCnpj: string;
	nome: string;
	email: string;
	phone?: string;
	endereco?: {
		logradouro: string;
		numero: string;
		complemento?: string;
		bairro: string;
		codigoMunicipio: string;
		cidade: string;
		uf: string;
		cep: string;
	};
}

/** Dados do servico prestado */
export interface ServicoNfse {
	codigoServico: string; // Item LC 116 (e.g., "1.05")
	cnae: string;
	descricao: string;
	valorServicos: number; // Em reais (decimal)
	valorIss: number; // ISS calculado
	aliquotaIss: number; // Aliquota ISS (decimal, ex: 0.02 = 2%)
}

/** DPS — Declaracao de Prestacao de Servicos (payload XML v1.01) */
export interface DPS {
	infDPS: {
		tpAmb: 1 | 2; // 1=Producao, 2=Homologacao
		dhEmi: string; // TSDateTimeUTC: YYYY-MM-DDTHH:MM:SS-03:00
		verAplic: string;
		serie: string; // Raw number (sem zeros a esquerda)
		nDPS: string; // Raw number (sem zeros a esquerda)
		dCompet: string; // YYYY-MM-DD
		tpEmit: 1; // 1 = Prestador
		cLocEmi: string; // Codigo IBGE municipio
		prest: {
			CNPJ: string;
			IM?: string; // Omitir se municipio nao tem CNC NFS-e
			regTrib: {
				opSimpNac: 1 | 2 | 3; // 1=Nao optante, 2=MEI, 3=ME/EPP
				regApTribSN?: 1; // Obrigatorio se opSimpNac=3
				regEspTrib: number;
			};
		};
		toma: {
			CPF?: string;
			CNPJ?: string;
			xNome: string;
			email?: string;
		};
		serv: {
			locPrest: { cLocPrestacao: string };
			cServ: {
				cTribNac: string; // 6 digitos (ex: "010701")
				cTribMun?: string; // 3 digitos — codigo complementar municipal (obrigatorio em RJ)
				xDescServ: string;
				cNBS?: string; // Nomenclatura Brasileira de Servicos (ex: "111032900")
			};
		};
		valores: {
			vServPrest: { vServ: string };
			trib: {
				tribMun: { tribISSQN: 1; tpRetISSQN: 1 };
				totTrib: {
					pTotTribSN?: string; // % total SN (para opSimpNac=3)
					indTotTrib?: 0; // Para nao-optantes SN
				};
			};
		};
	};
}

/** Resposta da API SEFIN apos envio do DPS */
export interface NfseSefinResponse {
	chaveAcesso?: string;
	numero?: string;
	codigoStatus: number;
	mensagemStatus: string;
	dataEmissao?: string;
	/** GZip+Base64 encoded official NFS-e XML (returned by SEFIN on success) */
	nfseXmlGZipB64?: string;
	/** DPS identifier assigned by SEFIN */
	idDps?: string;
}

/** Configuracao NFS-e carregada do banco */
export interface NfseConfigData {
	id: string;
	active: boolean;
	cnpj: string;
	inscricaoMunicipal: string;
	codigoMunicipio: string;
	uf: string;
	certificatePfxUrl: string;
	certificatePassword: string;
	certificateExpiresAt: Date;
	// PJ — Pessoa Juridica (Anexo V — 15,50%)
	codigoServico: string;
	codigoServicoMunicipal: string;
	codigoNbs: string;
	cnae: string;
	aliquotaIss: number;
	descricaoServico: string;
	// PF — Pessoa Fisica (Anexo III — 7,69%)
	codigoServicoPf: string;
	codigoServicoMunicipalPf: string;
	codigoNbsPf: string;
	cnaePf: string;
	aliquotaIssPf: number;
	descricaoServicoPf: string;
	ambiente: string;
	opSimpNac?: number; // 1=Nao optante, 2=MEI, 3=ME/EPP (default: 3)
	regApTribSN?: number; // Regime apuracao tributos SN (default: 1)
	regEspTrib?: number; // Regime especial tributacao (default: 0)
}

/** Resultado de emissao */
export interface NfseEmitResult {
	success: boolean;
	chaveAcesso?: string;
	numero?: string;
	danfseUrl?: string;
	pdfUrl?: string;
	xmlUrl?: string;
	dpsXmlUrl?: string;
	error?: string;
}

// =============================================================================
// Tomador Validation & Friendly Error Messages
// =============================================================================

/**
 * Validates that the tomador (contributor) has all required fields for NFS-e emission.
 * Returns null if valid, or a user-friendly error message in pt-BR.
 */
export function validateTomadorForNfse(tomador: Tomador): string | null {
	if (!tomador.cpfCnpj || tomador.cpfCnpj.replace(/\D/g, "").length === 0) {
		return "CPF/CNPJ do contribuidor nao informado. Solicite o preenchimento dos dados fiscais no perfil.";
	}

	const digits = tomador.cpfCnpj.replace(/\D/g, "");
	if (digits.length !== 11 && digits.length !== 14) {
		return "CPF/CNPJ do contribuidor e invalido. Solicite a correcao no perfil.";
	}

	if (!tomador.endereco) {
		return "Endereco do contribuidor nao informado. Solicite o preenchimento do endereco completo no perfil.";
	}

	const addr = tomador.endereco;
	const missing: string[] = [];
	if (!addr.logradouro) missing.push("rua");
	if (!addr.codigoMunicipio) missing.push("codigo do municipio");
	if (!addr.uf) missing.push("estado");
	if (!addr.cep) missing.push("CEP");

	if (missing.length > 0) {
		return `Endereco do contribuidor incompleto (falta: ${missing.join(", ")}). Solicite o preenchimento no perfil.`;
	}

	return null;
}

/**
 * Converts a raw SEFIN error response to a user-friendly pt-BR message.
 * Hides technical XML/schema details from the developer UI.
 */
export function parseSefinErrorToFriendly(rawError: string): string {
	// Pattern: SEFIN JSON error with "erros" array
	if (
		rawError.includes("enderToma") ||
		rawError.includes("'end, fone, email'") ||
		rawError.includes("'end'")
	) {
		return "Endereco do contribuidor incompleto ou com formato invalido. Solicite o preenchimento do endereco completo no perfil.";
	}

	if (rawError.includes("Falha Schema Xml") || rawError.includes("RNG6110")) {
		// Try to extract specific field info
		if (rawError.includes("'toma'")) {
			return "Dados do contribuidor incompletos ou com formato invalido. Verifique CPF/CNPJ, endereco e e-mail.";
		}
		if (rawError.includes("'prest'")) {
			return "Dados do prestador (desenvolvedor) incompletos. Verifique sua configuracao de NFS-e.";
		}
		return "Dados da nota fiscal com formato invalido. Verifique as configuracoes de NFS-e e os dados do contribuidor.";
	}

	if (rawError.includes("CPF") && rawError.includes("invalido")) {
		return "CPF do contribuidor e invalido. Solicite a correcao no perfil.";
	}

	if (rawError.includes("CNPJ") && rawError.includes("invalido")) {
		return "CNPJ do contribuidor e invalido. Solicite a correcao no perfil.";
	}

	if (
		rawError.includes("certificado") ||
		rawError.includes("Certificate") ||
		rawError.includes("ssl") ||
		rawError.includes("tls")
	) {
		return "Problema com o certificado digital. Verifique se o certificado A1 esta valido e nao expirado.";
	}

	if (rawError.includes("SEFIN API error 5")) {
		return "Erro temporario no servidor da SEFIN. A emissao sera retentada automaticamente.";
	}

	if (rawError.includes("SEFIN API error")) {
		return "Erro na comunicacao com a SEFIN. Tente novamente em alguns minutos.";
	}

	// Already friendly message (from our validation)
	if (
		rawError.includes("Solicite") ||
		rawError.includes("Verifique") ||
		rawError.includes("Configure")
	) {
		return rawError;
	}

	// Fallback: truncate raw error to avoid exposing technical details
	if (rawError.length > 200) {
		return "Erro na emissao da NFS-e. Verifique os dados do contribuidor e tente novamente.";
	}

	return rawError;
}
