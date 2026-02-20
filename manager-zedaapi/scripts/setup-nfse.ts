import "dotenv/config";
import * as fs from "fs";
import * as forge from "node-forge";
import { Pool } from "pg";
import { PrismaPg } from "@prisma/adapter-pg";
import { PrismaClient } from "../generated/prisma/client";

// =============================================================================
// Setup NFS-e — Configura certificado A1 e cria NfseConfig no banco
//
// USO:
//   bun run scripts/setup-nfse.ts setup <caminho-pfx> <senha>
//   bun run scripts/setup-nfse.ts status
//   bun run scripts/setup-nfse.ts toggle <HOMOLOGACAO|PRODUCAO>
//
// EXEMPLO:
//   bun run scripts/setup-nfse.ts setup certs/cert.p12 12345678
//   bun run scripts/setup-nfse.ts status
//   bun run scripts/setup-nfse.ts toggle PRODUCAO
//
// REQUISITOS:
//   - DATABASE_URL configurada no .env
//   - ENCRYPTION_KEY configurada no .env (64-char hex)
//   - S3_* configuradas no .env (endpoint, bucket, credentials)
// =============================================================================

const command = process.argv[2] || "status";

function getDb() {
	const pool = new Pool({ connectionString: process.env.DATABASE_URL });
	const adapter = new PrismaPg(pool);
	return new PrismaClient({ adapter, log: ["error"] });
}

// Import encrypt dynamically (needs ENCRYPTION_KEY)
async function encryptPassword(password: string): Promise<string> {
	const { encrypt } = await import("../lib/crypto/encryption");
	return encrypt(password);
}

// Upload file to S3
async function uploadToS3(
	fileBuffer: Buffer,
	key: string,
	contentType: string,
): Promise<string> {
	const { S3Client } = await import("bun");
	const s3 = new S3Client({
		accessKeyId: process.env.S3_ACCESS_KEY_ID!,
		secretAccessKey: process.env.S3_SECRET_ACCESS_KEY!,
		bucket: process.env.S3_BUCKET!,
		endpoint: process.env.S3_ENDPOINT!,
		region: process.env.S3_REGION || "us-east-1",
	});

	const s3File = s3.file(key);
	await s3File.write(fileBuffer, { type: contentType });

	const publicUrl = process.env.S3_PUBLIC_URL || process.env.S3_ENDPOINT;
	const bucket = process.env.S3_BUCKET;
	const pathStyle = process.env.S3_PATH_STYLE === "true";

	if (pathStyle) {
		return `${publicUrl}/${bucket}/${key}`;
	}
	return `${publicUrl}/${key}`;
}

// Parse PFX and extract info
function parsePfx(
	pfxPath: string,
	password: string,
): {
	cnpj: string;
	commonName: string;
	expiresAt: Date;
	isValid: boolean;
} {
	const pfxBuffer = fs.readFileSync(pfxPath);
	const pfxDer = forge.util.decode64(pfxBuffer.toString("base64"));
	const p12Asn1 = forge.asn1.fromDer(pfxDer);
	const p12 = forge.pkcs12.pkcs12FromAsn1(p12Asn1, password);

	// Extract certificate
	const certBagType = forge.pki.oids.certBag;
	const certBags = p12.getBags({ bagType: certBagType });
	const certBag = certBagType ? certBags[certBagType] : undefined;
	if (!certBag || certBag.length === 0 || !certBag[0]?.cert) {
		throw new Error("Certificado nao encontrado no arquivo PFX");
	}

	const cert = certBag[0].cert;
	const expiresAt = cert.validity.notAfter;
	const isValid = new Date() < expiresAt;

	// Extract subject info
	const cn = cert.subject.getField("CN")?.value || "Desconhecido";

	// Try to extract CNPJ from subject (common patterns in Brazilian A1 certs)
	let cnpj = "";

	// Pattern 1: OID 2.16.76.1.3.3 (ICP-Brasil CNPJ)
	const cnpjField = cert.subject.attributes.find(
		(attr: { type?: string; value?: unknown }) =>
			attr.type === "2.16.76.1.3.3",
	);
	if (cnpjField) {
		cnpj = String(cnpjField.value).replace(/\D/g, "");
	}

	// Pattern 2: Extract from CN (e.g., "EMPRESA LTDA:54246473000100")
	if (!cnpj) {
		const cnMatch = cn.match(/(\d{14})/);
		if (cnMatch) {
			cnpj = cnMatch[1];
		}
	}

	// Pattern 3: OU field
	if (!cnpj) {
		const ou = cert.subject.getField("OU")?.value || "";
		const ouMatch = ou.match(/(\d{14})/);
		if (ouMatch) {
			cnpj = ouMatch[1];
		}
	}

	return { cnpj, commonName: cn, expiresAt, isValid };
}

// ==================== COMMANDS ====================

async function setupCommand() {
	const pfxPath = process.argv[3];
	const password = process.argv[4];

	if (!pfxPath || !password) {
		console.error(
			"Uso: bun run scripts/setup-nfse.ts setup <caminho-pfx> <senha>",
		);
		console.error(
			"Exemplo: bun run scripts/setup-nfse.ts setup certs/cert.p12 12345678",
		);
		process.exit(1);
	}

	if (!fs.existsSync(pfxPath)) {
		console.error(`Arquivo nao encontrado: ${pfxPath}`);
		process.exit(1);
	}

	// Validate env vars
	if (
		!process.env.ENCRYPTION_KEY ||
		process.env.ENCRYPTION_KEY.length !== 64
	) {
		console.error(
			"ENCRYPTION_KEY nao configurada ou invalida (precisa ter 64 chars hex)",
		);
		process.exit(1);
	}
	if (!process.env.S3_ENDPOINT || !process.env.S3_BUCKET) {
		console.error("S3_ENDPOINT e S3_BUCKET sao obrigatorios");
		process.exit(1);
	}

	console.log("\n=== Setup NFS-e Nacional (Zé da API Manager) ===\n");

	// 1. Parse certificate
	console.log("1. Analisando certificado PFX...");
	let certInfo;
	try {
		certInfo = parsePfx(pfxPath, password);
	} catch (error) {
		console.error(
			"   Erro ao abrir PFX. Verifique a senha.",
			error instanceof Error ? error.message : error,
		);
		process.exit(1);
	}

	console.log(`   Nome: ${certInfo.commonName}`);
	console.log(
		`   CNPJ: ${certInfo.cnpj || "(nao encontrado no certificado)"}`,
	);
	console.log(`   Validade: ${certInfo.expiresAt.toISOString()}`);
	console.log(`   Status: ${certInfo.isValid ? "VALIDO" : "EXPIRADO"}`);

	if (!certInfo.isValid) {
		console.error(
			"\n   ATENCAO: Certificado expirado! Renove antes de usar em producao.",
		);
	}

	// Prompt for CNPJ if not found in cert
	let cnpj = certInfo.cnpj;
	if (!cnpj) {
		console.log("\n   CNPJ nao encontrado automaticamente no certificado.");
		console.log(
			"   Informe o CNPJ da empresa emissora (apenas numeros, 14 digitos):",
		);
		const response = prompt("   CNPJ: ");
		cnpj = (response || "").replace(/\D/g, "");
		if (cnpj.length !== 14) {
			console.error("   CNPJ invalido. Deve ter 14 digitos.");
			process.exit(1);
		}
	}

	// 2. Upload to S3
	console.log("\n2. Enviando certificado para S3...");
	const pfxBuffer = fs.readFileSync(pfxPath);
	const s3Key = `nfse/certs/${cnpj}.p12`;
	const pfxUrl = await uploadToS3(pfxBuffer, s3Key, "application/x-pkcs12");
	console.log(`   URL: ${pfxUrl}`);

	// 3. Encrypt password
	console.log("\n3. Criptografando senha do certificado...");
	const encryptedPassword = await encryptPassword(password);
	console.log(`   Senha criptografada (${encryptedPassword.length} chars)`);

	// 4. Create/Update NfseConfig
	console.log("\n4. Salvando configuracao no banco...");
	const prisma = getDb();

	// Check if config already exists for this CNPJ
	const existing = await prisma.nfseConfig.findUnique({
		where: { cnpj },
	});

	const configData = {
		active: true,
		cnpj,
		inscricaoMunicipal: "", // User needs to fill this
		codigoMunicipio: "", // User needs to fill this
		uf: "", // User needs to fill this
		certificatePfxUrl: pfxUrl,
		certificatePassword: encryptedPassword,
		certificateExpiresAt: certInfo.expiresAt,
		// PJ — Pessoa Juridica (Anexo V — 15,50%)
		codigoServico: "010701", // cTribNac — Licenciamento de software / SaaS
		codigoServicoMunicipal: "", // cTribMun — Complementar municipal
		codigoNbs: "111032900", // cNBS — Licenciamento de software e bancos de dados
		cnae: "6311900", // Tratamento de dados, provedores de aplicacao e hospedagem
		aliquotaIss: 0.155, // 15,50% — Simples Nacional Anexo V, 1a faixa
		descricaoServico:
			"Licenciamento e hospedagem de plataforma SaaS para gestao de comunicacao empresarial via WhatsApp - assinatura mensal de API",
		// PF — Pessoa Fisica (Anexo III — 7,69%)
		codigoServicoPf: "010701", // cTribNac — Same service code for PF
		codigoServicoMunicipalPf: "", // cTribMun — Configurar quando disponivel
		codigoNbsPf: "", // cNBS — Configurar quando disponivel
		cnaePf: "6311900", // Same CNAE for PF
		aliquotaIssPf: 0.0769, // 7,69% — Simples Nacional Anexo III efetiva
		descricaoServicoPf:
			"Licenciamento e hospedagem de plataforma SaaS para gestao de comunicacao empresarial via WhatsApp - assinatura mensal de API",
		ambiente: "HOMOLOGACAO",
	};

	if (existing) {
		await prisma.nfseConfig.update({
			where: { cnpj },
			data: configData,
		});
		console.log(`   NfseConfig atualizado (ID: ${existing.id})`);
	} else {
		const created = await prisma.nfseConfig.create({
			data: configData,
		});
		console.log(`   NfseConfig criado (ID: ${created.id})`);
	}

	// 5. Summary
	console.log("\n=== Setup Concluido ===\n");
	console.log("IMPORTANTE: Complete os dados manualmente no banco:");
	console.log("  - inscricaoMunicipal  (Inscricao Municipal da empresa)");
	console.log(
		"  - codigoMunicipio     (Codigo IBGE do municipio, ex: 4314902 = Porto Alegre)",
	);
	console.log("  - uf                  (UF, ex: RS)");
	console.log("");
	console.log("Exemplo SQL:");
	console.log(`  UPDATE nfse_configs SET`);
	console.log(`    inscricao_municipal = 'SEU_IM',`);
	console.log(`    codigo_municipio = '4314902',`);
	console.log(`    uf = 'RS'`);
	console.log(`  WHERE cnpj = '${cnpj}';`);
	console.log("");
	console.log("Para ativar em producao:");
	console.log("  bun run scripts/setup-nfse.ts toggle PRODUCAO");
	console.log("");

	await prisma.$disconnect();
}

async function statusCommand() {
	console.log("\n=== Status NFS-e (Zé da API Manager) ===\n");
	const prisma = getDb();

	const configs = await prisma.nfseConfig.findMany();

	if (configs.length === 0) {
		console.log("Nenhuma configuracao NFS-e encontrada.");
		console.log(
			"Execute: bun run scripts/setup-nfse.ts setup <pfx> <senha>",
		);
		await prisma.$disconnect();
		return;
	}

	for (const config of configs) {
		const isExpired = new Date() >= config.certificateExpiresAt;
		console.log(`ID: ${config.id}`);
		console.log(`  CNPJ: ${config.cnpj}`);
		console.log(`  IM: ${config.inscricaoMunicipal || "(nao preenchido)"}`);
		console.log(
			`  Municipio: ${config.codigoMunicipio || "(nao preenchido)"} / ${config.uf || "(nao preenchido)"}`,
		);
		console.log(`  Ambiente: ${config.ambiente}`);
		console.log(`  Ativo: ${config.active ? "SIM" : "NAO"}`);
		console.log(
			`  Certificado: ${isExpired ? "EXPIRADO" : "VALIDO"} (ate ${config.certificateExpiresAt.toISOString()})`,
		);
		console.log(`  PFX URL: ${config.certificatePfxUrl}`);
		console.log(
			`  PJ: CNAE ${config.cnae} / Servico ${config.codigoServico} / ${(config.aliquotaIss * 100).toFixed(2)}%`,
		);
		console.log(
			`  PF: CNAE ${config.cnaePf} / Servico ${config.codigoServicoPf} / ${(config.aliquotaIssPf * 100).toFixed(2)}%`,
		);
		console.log("");
	}

	// NFS-e stats
	const stats = await prisma.invoice.groupBy({
		by: ["nfseStatus"],
		_count: true,
		where: { nfseStatus: { not: null } },
	});

	if (stats.length > 0) {
		console.log("Estatisticas NFS-e:");
		for (const s of stats) {
			console.log(`  ${s.nfseStatus}: ${s._count}`);
		}
	}

	await prisma.$disconnect();
}

async function toggleCommand() {
	const ambiente = process.argv[3]?.toUpperCase();

	if (!ambiente || !["HOMOLOGACAO", "PRODUCAO"].includes(ambiente)) {
		console.error(
			"Uso: bun run scripts/setup-nfse.ts toggle <HOMOLOGACAO|PRODUCAO>",
		);
		process.exit(1);
	}

	const prisma = getDb();

	const config = await prisma.nfseConfig.findFirst({
		where: { active: true },
	});
	if (!config) {
		console.error("Nenhuma configuracao NFS-e ativa encontrada.");
		await prisma.$disconnect();
		process.exit(1);
	}

	await prisma.nfseConfig.update({
		where: { id: config.id },
		data: { ambiente },
	});

	console.log(`Ambiente alterado para: ${ambiente}`);
	console.log(`CNPJ: ${config.cnpj}`);

	await prisma.$disconnect();
}

// ==================== MAIN ====================

async function main() {
	switch (command) {
		case "setup":
			await setupCommand();
			break;
		case "status":
			await statusCommand();
			break;
		case "toggle":
			await toggleCommand();
			break;
		default:
			console.log("Comandos disponiveis:");
			console.log(
				"  setup <pfx> <senha>      — Configura certificado e cria NfseConfig",
			);
			console.log(
				"  status                   — Mostra configuracao atual",
			);
			console.log("  toggle <HOMOLOGACAO|PRODUCAO> — Altera ambiente");
	}
}

main().catch((error) => {
	console.error("Erro:", error);
	process.exit(1);
});
