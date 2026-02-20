// =============================================================================
// NFS-e Nacional — DPS XML Builder (v1.01 schema)
// =============================================================================

import { XMLBuilder } from "fast-xml-parser";
import type { NfseConfigData, Tomador } from "./types";

interface BuildDpsOptions {
	/** Invoice ID (used as DPS number) */
	invoiceId: string;
	/** Amount in cents (base for NFS-e — subscription amount) */
	amountCents: number;
	/** Tomador (contributor) data */
	tomador: Tomador;
	/** NFS-e config from database */
	config: NfseConfigData;
	/** Sequential DPS number (raw, e.g. "3") */
	dpsNumber: string;
}

/**
 * Build DPS XML according to NFS-e Nacional v1.01 schema.
 * The NFS-e is emitted on the subscription amount.
 *
 * Element order inside infDPS (strict XSD sequence):
 *   tpAmb -> dhEmi -> verAplic -> serie -> nDPS -> dCompet -> tpEmit -> cLocEmi
 *   -> [subst] -> prest -> toma -> [interm] -> serv -> valores
 */
export function buildDpsXml(options: BuildDpsOptions): string {
	const { amountCents, tomador, config, dpsNumber } = options;

	const valorServicos = amountCents / 100;

	const now = new Date();
	const competencia = now.toISOString().slice(0, 10); // YYYY-MM-DD

	// SEFIN requires TSDateTimeUTC format: YYYY-MM-DDTHH:MM:SS-03:00
	// Must use BRT offset (-03:00), no milliseconds, no "Z"
	const pad = (n: number) => String(n).padStart(2, "0");
	const brtOffset = -3;
	const brt = new Date(now.getTime() + brtOffset * 60 * 60 * 1000);
	const dataEmissao = `${brt.getUTCFullYear()}-${pad(brt.getUTCMonth() + 1)}-${pad(brt.getUTCDate())}T${pad(brt.getUTCHours())}:${pad(brt.getUTCMinutes())}:${pad(brt.getUTCSeconds())}-03:00`;

	const isProducao = config.ambiente === "PRODUCAO";

	// -- DPS Id composition --
	// Format: "DPS" + cMunEmi(7) + tpInscFed(1) + CNPJ(14) + serie(5) + nDPS(15) = 45 chars total
	// Zero-padding is ONLY for the Id attribute. XML elements use raw values.
	const cnpjDigits = config.cnpj.replace(/\D/g, "").padStart(14, "0");
	const tpInscFed = "2"; // 2 = CNPJ
	const serieRaw = "1";
	const seriePadded = serieRaw.padStart(5, "0");
	const cMunEmi = config.codigoMunicipio.padStart(7, "0");
	const nDPSPadded = dpsNumber.padStart(15, "0");
	const dpsId = `DPS${cMunEmi}${tpInscFed}${cnpjDigits}${seriePadded}${nDPSPadded}`;

	// -- Tomador (service taker) --
	// XSD strict order: CPF/CNPJ -> xNome -> end -> fone -> email
	const tomaNode: Record<string, unknown> = {};
	const tomaDigits = tomador.cpfCnpj.replace(/\D/g, "");
	const isTomadorPf = tomaDigits.length === 11;
	if (isTomadorPf) {
		tomaNode.CPF = tomaDigits;
	} else if (tomaDigits.length === 14) {
		tomaNode.CNPJ = tomaDigits;
	}
	tomaNode.xNome = tomador.nome;

	// end — Endereco nacional do tomador (required for complete NFS-e)
	if (tomador.endereco) {
		const addr = tomador.endereco;
		const endNacNode: Record<string, string> = {
			xLgr: addr.logradouro,
			nro: addr.numero || "S/N",
		};
		if (addr.complemento) {
			endNacNode.xCpl = addr.complemento;
		}
		endNacNode.xBairro = addr.bairro || "Nao informado";
		endNacNode.cMun = addr.codigoMunicipio;
		endNacNode.UF = addr.uf;
		endNacNode.CEP = addr.cep.replace(/\D/g, "");
		tomaNode.end = { endNac: endNacNode };
	}

	// fone — Telefone do tomador (optional)
	if (tomador.phone) {
		tomaNode.fone = tomador.phone.replace(/\D/g, "");
	}

	if (tomador.email) {
		tomaNode.email = tomador.email;
	}

	// -- Selecao de configuracao fiscal PF/PJ --
	// PF (CPF): Anexo III — 7,69%
	// PJ (CNPJ): Anexo V — 15,50%
	const codigoServico = isTomadorPf
		? config.codigoServicoPf
		: config.codigoServico;
	const codigoServicoMunicipal = isTomadorPf
		? config.codigoServicoMunicipalPf
		: config.codigoServicoMunicipal;
	const codigoNbs = isTomadorPf ? config.codigoNbsPf : config.codigoNbs;
	const aliquotaIss = isTomadorPf ? config.aliquotaIssPf : config.aliquotaIss;
	const descricaoServico = isTomadorPf
		? config.descricaoServicoPf
		: config.descricaoServico;

	// -- Build DPS object (strict XSD element order) --
	const opSimpNac = config.opSimpNac ?? 3;
	const dpsObj = {
		DPS: {
			"@_xmlns": "http://www.sped.fazenda.gov.br/nfse",
			"@_versao": "1.01",
			infDPS: {
				"@_Id": dpsId,
				tpAmb: isProducao ? 1 : 2,
				dhEmi: dataEmissao,
				verAplic: "ManagerZedaAPI1.0",
				serie: serieRaw,
				nDPS: dpsNumber,
				dCompet: competencia,
				tpEmit: 1, // 1 = Prestador (proprio)
				cLocEmi: config.codigoMunicipio,
				prest: {
					CNPJ: cnpjDigits,
					// IM is only included when the municipality has CNC NFS-e complementary registration.
					// Error E0120 means the municipality doesn't support IM — omit it.
					...(config.inscricaoMunicipal &&
					config.inscricaoMunicipal !== "SKIP"
						? { IM: config.inscricaoMunicipal }
						: {}),
					regTrib: {
						opSimpNac,
						...(opSimpNac === 3
							? { regApTribSN: config.regApTribSN ?? 1 }
							: {}),
						regEspTrib: config.regEspTrib ?? 0,
					},
				},
				toma: tomaNode,
				serv: {
					locPrest: {
						cLocPrestacao: config.codigoMunicipio,
					},
					cServ: {
						cTribNac: codigoServico,
						// cTribMun — Codigo complementar municipal (obrigatorio em RJ e outros municipios)
						...(codigoServicoMunicipal
							? { cTribMun: codigoServicoMunicipal }
							: {}),
						xDescServ: descricaoServico,
						// cNBS — Nomenclatura Brasileira de Servicos
						...(codigoNbs ? { cNBS: codigoNbs } : {}),
					},
				},
				valores: {
					vServPrest: {
						vServ: valorServicos.toFixed(2),
					},
					trib: {
						// XSD xs:sequence — tribMun -> tribFed -> totTrib (order matters)
						tribMun: {
							tribISSQN: 1, // 1 = Operacao tributavel
							tpRetISSQN: 1, // 1 = Nao retido (prestador recolhe)
						},
						totTrib:
							opSimpNac === 3
								? { pTotTribSN: (aliquotaIss * 100).toFixed(2) }
								: { indTotTrib: 0 },
					},
				},
			},
		},
	};

	const builder = new XMLBuilder({
		ignoreAttributes: false,
		attributeNamePrefix: "@_",
		processEntities: true,
		format: false, // Compact for signing (C14N requires no extra whitespace)
		suppressEmptyNode: true,
	});

	const xmlOutput = builder.build(dpsObj);
	return '<?xml version="1.0" encoding="UTF-8"?>\n' + xmlOutput;
}

// =============================================================================
// Cancel Event XML Builder (pedRegEvento e101101)
// =============================================================================

interface BuildCancelEventOptions {
	/** Chave de acesso da NFS-e a cancelar */
	chaveAcesso: string;
	/** Motivo do cancelamento (xMotivo) */
	motivo: string;
	/** CNPJ do autor (prestador/emitente) */
	cnpjAutor: string;
	/** Ambiente: PRODUCAO ou HOMOLOGACAO */
	ambiente: string;
}

/**
 * Build pedRegEvento XML for NFS-e cancellation (event type e101101).
 */
export function buildCancelEventXml(options: BuildCancelEventOptions): string {
	const { chaveAcesso, motivo, cnpjAutor, ambiente } = options;

	const isProducao = ambiente === "PRODUCAO";
	const cnpjDigits = cnpjAutor.replace(/\D/g, "").padStart(14, "0");

	// Id format: PRE(3) + chaveAcesso(50) + tpEvento(6) = 59 chars
	const infPedRegId = `PRE${chaveAcesso}101101`;

	// BRT datetime
	const now = new Date();
	const pad = (n: number) => String(n).padStart(2, "0");
	const brtOffset = -3;
	const brt = new Date(now.getTime() + brtOffset * 60 * 60 * 1000);
	const dhEvento = `${brt.getUTCFullYear()}-${pad(brt.getUTCMonth() + 1)}-${pad(brt.getUTCDate())}T${pad(brt.getUTCHours())}:${pad(brt.getUTCMinutes())}:${pad(brt.getUTCSeconds())}-03:00`;

	const eventObj = {
		pedRegEvento: {
			"@_xmlns": "http://www.sped.fazenda.gov.br/nfse",
			"@_versao": "1.00",
			infPedReg: {
				"@_Id": infPedRegId,
				tpAmb: isProducao ? 1 : 2,
				verAplic: "ManagerZedaAPI1.0",
				dhEvento,
				CNPJAutor: cnpjDigits,
				chNFSe: chaveAcesso,
				e101101: {
					xDesc: "Cancelamento de NFS-e",
					cMotivo: 2, // 2 = Servico nao prestado / Erro na emissao
					xMotivo: motivo,
				},
			},
		},
	};

	const builder = new XMLBuilder({
		ignoreAttributes: false,
		attributeNamePrefix: "@_",
		processEntities: true,
		format: false,
		suppressEmptyNode: true,
	});

	const xmlOutput = builder.build(eventObj);
	return '<?xml version="1.0" encoding="UTF-8"?>\n' + xmlOutput;
}
