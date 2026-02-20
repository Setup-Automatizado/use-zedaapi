import "dotenv/config";
import { PrismaPg } from "@prisma/adapter-pg";
import { PrismaClient } from "../generated/prisma/client";
import { hashPassword } from "better-auth/crypto";

const adapter = new PrismaPg({
	connectionString: process.env.DATABASE_URL!,
});
const prisma = new PrismaClient({ adapter, log: ["error"] });

async function main() {
	console.log("Seeding database...\n");

	// =========================================================================
	// 1. Plans — Pricing agressivo (undercut em TODOS os tiers)
	//
	// Concorrente:
	//   10 devices  = R$190  (R$19,00/device)
	//   100 devices = R$138  (R$1,38/device)
	//   300 devices = R$195  (R$0,65/device)
	//
	// Nosso (sempre mais barato):
	//   1   = R$9    (R$9,00)   — sem equivalente no concorrente
	//   10  = R$29   (R$2,90)   — vs R$190 (85% mais barato)
	//   30  = R$59   (R$1,97)   — sem equivalente
	//   100 = R$99   (R$0,99)   — vs R$138 (28% mais barato)
	//   300 = R$149  (R$0,50)   — vs R$195 (24% mais barato)
	//   500 = R$199  (R$0,40)   — sem equivalente
	// =========================================================================
	const plans = [
		{
			name: "Starter",
			slug: "starter",
			description:
				"Ideal para quem esta comecando. 1 instancia WhatsApp por apenas R$9/mes.",
			price: 9,
			currency: "BRL",
			interval: "month",
			maxInstances: 1,
			features: [
				"1 instancia WhatsApp",
				"Envio e recebimento ilimitado",
				"Webhooks em tempo real",
				"API REST completa",
				"Botoes interativos*",
				"Nodes community n8n",
				"Suporte por email",
				"Documentacao completa",
			],
			active: true,
			sortOrder: 1,
		},
		{
			name: "Pro",
			slug: "pro",
			description:
				"Para equipes em crescimento. 10 instancias por R$2,90/cada.",
			price: 29,
			currency: "BRL",
			interval: "month",
			maxInstances: 10,
			features: [
				"10 instancias WhatsApp",
				"Envio e recebimento ilimitado",
				"Webhooks em tempo real",
				"API REST completa",
				"Botoes interativos*",
				"Nodes community n8n",
				"Suporte prioritario",
				"Gerenciamento de grupos",
				"Multi-dispositivo",
				"Metricas e analytics",
			],
			active: true,
			sortOrder: 2,
		},
		{
			name: "Business",
			slug: "business",
			description:
				"Para operacoes em escala. 30 instancias por R$1,97/cada.",
			price: 59,
			currency: "BRL",
			interval: "month",
			maxInstances: 30,
			features: [
				"30 instancias WhatsApp",
				"Envio e recebimento ilimitado",
				"Webhooks em tempo real",
				"API REST completa",
				"Botoes interativos*",
				"Nodes community n8n",
				"Suporte prioritario 24/7",
				"Gerenciamento de grupos",
				"Multi-dispositivo",
				"Metricas e analytics",
				"API estavel e sempre atualizada",
				"SLA garantido 99.9%",
			],
			active: true,
			sortOrder: 3,
		},
		{
			name: "Scale",
			slug: "scale",
			description:
				"Para quem precisa escalar. 100 instancias por R$0,99/cada.",
			price: 99,
			currency: "BRL",
			interval: "month",
			maxInstances: 100,
			features: [
				"100 instancias WhatsApp",
				"Envio e recebimento ilimitado",
				"Webhooks em tempo real",
				"API REST completa",
				"Botoes interativos*",
				"Nodes community n8n",
				"Suporte dedicado 24/7",
				"Gerenciamento de grupos",
				"Multi-dispositivo",
				"Metricas e analytics avancadas",
				"API estavel e sempre atualizada",
				"SLA garantido 99.9%",
				"Gerente de conta dedicado",
			],
			active: true,
			sortOrder: 4,
		},
		{
			name: "Enterprise",
			slug: "enterprise",
			description:
				"Para grandes operacoes. 300 instancias por R$0,50/cada.",
			price: 149,
			currency: "BRL",
			interval: "month",
			maxInstances: 300,
			features: [
				"300 instancias WhatsApp",
				"Envio e recebimento ilimitado",
				"Webhooks em tempo real",
				"API REST completa",
				"Botoes interativos*",
				"Nodes community n8n",
				"Suporte dedicado 24/7 com SLA",
				"Gerenciamento de grupos",
				"Multi-dispositivo",
				"Metricas e analytics avancadas",
				"API estavel e sempre atualizada",
				"SLA garantido 99.95%",
				"Gerente de conta dedicado",
				"Onboarding personalizado",
				"Infraestrutura dedicada",
			],
			active: true,
			sortOrder: 5,
		},
		{
			name: "Ultimate",
			slug: "ultimate",
			description:
				"Capacidade maxima. 500 instancias por apenas R$0,40/cada.",
			price: 199,
			currency: "BRL",
			interval: "month",
			maxInstances: 500,
			features: [
				"500 instancias WhatsApp",
				"Envio e recebimento ilimitado",
				"Webhooks em tempo real",
				"API REST completa",
				"Botoes interativos*",
				"Nodes community n8n",
				"Suporte dedicado 24/7 com SLA premium",
				"Gerenciamento de grupos",
				"Multi-dispositivo",
				"Metricas e analytics avancadas",
				"API estavel e sempre atualizada",
				"SLA garantido 99.99%",
				"Gerente de conta dedicado",
				"Onboarding personalizado",
				"Infraestrutura dedicada",
				"Prioridade em novos recursos",
				"Consultoria tecnica inclusa",
			],
			active: true,
			sortOrder: 6,
		},
	];

	console.log("[Plans]");
	for (const plan of plans) {
		await prisma.plan.upsert({
			where: { slug: plan.slug },
			update: plan,
			create: plan,
		});
		console.log(
			`  ✓ "${plan.name}" — R$${plan.price}/mes, ${plan.maxInstances} instancias`,
		);
	}

	// =========================================================================
	// 2. Admin User
	// =========================================================================
	console.log("\n[Admin User]");
	const adminEmail = process.env.ADMIN_EMAIL || "admin@zedaapi.com";
	const adminPassword = process.env.ADMIN_PASSWORD || "Admin@ZedaAPI2026!";
	const existingAdmin = await prisma.user.findUnique({
		where: { email: adminEmail },
	});

	if (!existingAdmin) {
		const hashedPassword = await hashPassword(adminPassword);
		const adminUser = await prisma.user.create({
			data: {
				name: "Admin Zé da API",
				email: adminEmail,
				emailVerified: true,
				role: "admin",
				country: "BR",
			},
		});
		// Create credential account for Better Auth login
		await prisma.account.create({
			data: {
				accountId: adminUser.id,
				providerId: "credential",
				userId: adminUser.id,
				password: hashedPassword,
			},
		});
		console.log(`  ✓ Criado: ${adminEmail}`);
		console.log(
			`  ✓ Senha: ${adminPassword.slice(0, 3)}${"*".repeat(adminPassword.length - 3)}`,
		);
	} else {
		// Ensure admin role is set
		if (existingAdmin.role !== "admin") {
			await prisma.user.update({
				where: { id: existingAdmin.id },
				data: { role: "admin" },
			});
			console.log(`  ✓ Atualizado para admin: ${adminEmail}`);
		} else {
			console.log(`  ○ Ja existe: ${adminEmail}`);
		}
		// Ensure credential account exists
		const existingAccount = await prisma.account.findFirst({
			where: { userId: existingAdmin.id, providerId: "credential" },
		});
		if (!existingAccount) {
			const hashedPassword = await hashPassword(adminPassword);
			await prisma.account.create({
				data: {
					accountId: existingAdmin.id,
					providerId: "credential",
					userId: existingAdmin.id,
					password: hashedPassword,
				},
			});
			console.log(`  ✓ Account credential criado para admin`);
		}
	}

	// =========================================================================
	// 3. NFS-e Config (placeholder - inativo por padrao)
	// =========================================================================
	console.log("\n[NFS-e Config]");
	const existingNfse = await prisma.nfseConfig.findFirst();

	if (!existingNfse) {
		await prisma.nfseConfig.create({
			data: {
				active: false,
				cnpj: "00000000000000",
				inscricaoMunicipal: "",
				codigoMunicipio: "4106902",
				uf: "PR",
				certificatePfxUrl: "",
				certificatePassword: "",
				certificateExpiresAt: new Date("2027-01-01"),
				codigoServico: "1.05",
				codigoServicoMunicipal: "",
				codigoNbs: "",
				cnae: "6311900",
				aliquotaIss: 5.0,
				descricaoServico:
					"Licenciamento de direito de uso de software - SaaS WhatsApp API",
				codigoServicoPf: "1.05",
				codigoServicoMunicipalPf: "",
				codigoNbsPf: "",
				cnaePf: "6311900",
				aliquotaIssPf: 5.0,
				descricaoServicoPf:
					"Licenciamento de direito de uso de software - SaaS WhatsApp API",
				opSimpNac: 3,
				regApTribSN: 0,
				regEspTrib: 0,
				ambiente: "HOMOLOGACAO",
			},
		});
		console.log("  ✓ Config NFS-e criada (inativa — placeholder)");
	} else {
		console.log("  ○ Config NFS-e ja existe");
	}

	// =========================================================================
	// 4. Feature Flags
	// =========================================================================
	console.log("\n[Feature Flags]");
	const featureFlags = [
		{
			key: "waitlist_enabled",
			enabled: true,
			description:
				"Ativar sistema de lista de espera para novos usuarios",
		},
		{
			key: "affiliates_enabled",
			enabled: true,
			description: "Ativar sistema de afiliados",
		},
		{
			key: "sicredi_pix_enabled",
			enabled: true,
			description: "Ativar pagamento via PIX Sicredi",
		},
		{
			key: "sicredi_boleto_enabled",
			enabled: true,
			description: "Ativar pagamento via Boleto Hibrido Sicredi",
		},
		{
			key: "nfse_auto_issue",
			enabled: true,
			description: "Emitir NFS-e automaticamente apos pagamento",
		},
		{
			key: "maintenance_mode",
			enabled: true,
			description:
				"Ativar modo de manutencao (bloqueia acesso ao dashboard)",
		},
		{
			key: "pair_phone_enabled",
			enabled: true,
			description:
				"Ativar pareamento por codigo de telefone (alem do QR Code)",
		},
		{
			key: "proxy_management_enabled",
			enabled: true,
			description: "Ativar gerenciamento de proxy por instancia",
		},
		{
			key: "organization_enabled",
			enabled: true,
			description: "Ativar sistema de organizacoes (multi-tenant)",
		},
		{
			key: "trial_enabled",
			enabled: true,
			description: "Ativar periodo de trial gratuito para novos usuarios",
		},
	];

	for (const flag of featureFlags) {
		await prisma.featureFlag.upsert({
			where: { key: flag.key },
			update: { description: flag.description },
			create: flag,
		});
		console.log(
			`  ${flag.enabled ? "✓" : "○"} "${flag.key}" — ${flag.enabled ? "ON" : "OFF"}`,
		);
	}

	// =========================================================================
	// 5. System Settings
	// =========================================================================
	console.log("\n[System Settings]");
	const settings = [
		{
			key: "platform_name",
			value: "Zé da API Manager",
			description: "Nome da plataforma exibido no UI",
		},
		{
			key: "support_email",
			value: "suporte@zedaapi.com",
			description: "Email de suporte exibido para usuarios",
		},
		{
			key: "default_currency",
			value: "BRL",
			description: "Moeda padrao para cobranças",
		},
		{
			key: "affiliate_commission_rate",
			value: "0.20",
			description: "Taxa de comissao de afiliados (20%)",
		},
		{
			key: "trial_days",
			value: "7",
			description: "Dias de trial gratuito para novos usuarios",
		},
		{
			key: "max_api_keys_per_user",
			value: "5",
			description: "Numero maximo de API keys por usuario",
		},
		{
			key: "instance_sync_interval_seconds",
			value: "300",
			description:
				"Intervalo em segundos para sync automatico de instancias",
		},
		{
			key: "webhook_timeout_seconds",
			value: "10",
			description: "Timeout para entrega de webhooks",
		},
		{
			key: "nfse_ambiente",
			value: "HOMOLOGACAO",
			description: "Ambiente NFS-e: HOMOLOGACAO ou PRODUCAO",
		},
		{
			key: "pix_expiration_seconds",
			value: "3600",
			description: "Tempo de expiracao da cobranca PIX (1 hora)",
		},
		{
			key: "boleto_days_to_expire",
			value: "3",
			description: "Dias ate vencimento do boleto",
		},
	];

	for (const setting of settings) {
		await prisma.systemSetting.upsert({
			where: { key: setting.key },
			update: { value: setting.value, description: setting.description },
			create: setting,
		});
		console.log(`  ✓ "${setting.key}" = "${setting.value}"`);
	}

	// =========================================================================
	// 6. Email Templates
	// =========================================================================
	console.log("\n[Email Templates]");
	const emailTemplates = [
		{
			slug: "welcome",
			name: "Boas-vindas",
			subject: "Bem-vindo ao Zé da API Manager!",
			body: "Ola {{name}}, seja bem-vindo ao Zé da API Manager! Sua conta foi criada com sucesso.",
			variables: JSON.stringify(["name", "email", "loginUrl"]),
			active: true,
		},
		{
			slug: "verify-email",
			name: "Verificacao de email",
			subject: "Verifique seu email - Zé da API Manager",
			body: "Ola {{name}}, clique no link abaixo para verificar seu email: {{verifyUrl}}",
			variables: JSON.stringify(["name", "email", "verifyUrl"]),
			active: true,
		},
		{
			slug: "reset-password",
			name: "Redefinicao de senha",
			subject: "Redefinir senha - Zé da API Manager",
			body: "Ola {{name}}, clique no link para redefinir sua senha: {{resetUrl}}. Este link expira em 1 hora.",
			variables: JSON.stringify(["name", "email", "resetUrl"]),
			active: true,
		},
		{
			slug: "invoice-paid",
			name: "Fatura paga",
			subject: "Pagamento confirmado — Fatura #{{invoiceNumber}}",
			body: "Ola {{name}}, seu pagamento de R${{amount}} foi confirmado. Fatura #{{invoiceNumber}}.",
			variables: JSON.stringify([
				"name",
				"amount",
				"invoiceNumber",
				"pdfUrl",
			]),
			active: true,
		},
		{
			slug: "payment-failed",
			name: "Falha no pagamento",
			subject: "Falha no pagamento — Acao necessaria",
			body: "Ola {{name}}, houve uma falha no pagamento da sua assinatura. Por favor, atualize seu metodo de pagamento.",
			variables: JSON.stringify([
				"name",
				"planName",
				"retryUrl",
				"dueDate",
			]),
			active: true,
		},
		{
			slug: "subscription-created",
			name: "Assinatura criada",
			subject: "Assinatura ativada — Plano {{planName}}",
			body: "Ola {{name}}, sua assinatura do plano {{planName}} foi ativada com sucesso!",
			variables: JSON.stringify([
				"name",
				"planName",
				"maxInstances",
				"price",
			]),
			active: true,
		},
		{
			slug: "subscription-canceled",
			name: "Assinatura cancelada",
			subject: "Assinatura cancelada — Plano {{planName}}",
			body: "Ola {{name}}, sua assinatura do plano {{planName}} foi cancelada. O acesso continua ate {{endDate}}.",
			variables: JSON.stringify(["name", "planName", "endDate"]),
			active: true,
		},
		{
			slug: "nfse-issued",
			name: "NFS-e emitida",
			subject: "NFS-e #{{nfseNumber}} emitida com sucesso",
			body: "Ola {{name}}, sua NFS-e #{{nfseNumber}} no valor de R${{amount}} foi emitida. Acesse o DANFSE: {{pdfUrl}}",
			variables: JSON.stringify([
				"name",
				"nfseNumber",
				"amount",
				"pdfUrl",
				"xmlUrl",
			]),
			active: true,
		},
		{
			slug: "waitlist-approved",
			name: "Aprovado na lista de espera",
			subject: "Voce foi aprovado! Crie sua conta no Zé da API Manager",
			body: "Ola {{name}}, voce foi aprovado na lista de espera! Use o codigo {{inviteCode}} para criar sua conta: {{registerUrl}}",
			variables: JSON.stringify([
				"name",
				"email",
				"inviteCode",
				"registerUrl",
			]),
			active: true,
		},
		{
			slug: "instance-disconnected",
			name: "Instancia desconectada",
			subject: "Alerta: Instancia {{instanceName}} desconectada",
			body: "Ola {{name}}, a instancia {{instanceName}} ({{phone}}) foi desconectada. Reconecte pelo painel.",
			variables: JSON.stringify([
				"name",
				"instanceName",
				"phone",
				"reconnectUrl",
			]),
			active: true,
		},
		{
			slug: "affiliate-approved",
			name: "Afiliado aprovado",
			subject: "Parabens! Voce e um afiliado ZedaAPI",
			body: "Ola {{name}}, sua solicitacao de afiliado foi aprovada! Seu codigo de referencia e {{affiliateCode}}. Comece a indicar: {{dashboardUrl}}",
			variables: JSON.stringify([
				"name",
				"affiliateCode",
				"commissionRate",
				"dashboardUrl",
			]),
			active: true,
		},
		{
			slug: "payout-processed",
			name: "Pagamento de comissao",
			subject: "Pagamento de R${{amount}} processado",
			body: "Ola {{name}}, seu pagamento de comissao de R${{amount}} foi processado via {{method}}.",
			variables: JSON.stringify(["name", "amount", "method", "date"]),
			active: true,
		},
	];

	for (const template of emailTemplates) {
		await prisma.emailTemplate.upsert({
			where: { slug: template.slug },
			update: {
				name: template.name,
				subject: template.subject,
				body: template.body,
				variables: template.variables,
				active: template.active,
			},
			create: template,
		});
		console.log(`  ✓ "${template.slug}" — ${template.name}`);
	}

	// =========================================================================
	// 7. NFS-e Sequence (initialize if not exists)
	// =========================================================================
	console.log("\n[NFS-e Sequence]");
	const nfseConfig = await prisma.nfseConfig.findFirst({
		where: { active: true },
	});

	if (nfseConfig) {
		const existing = await prisma.nfseSequence.findUnique({
			where: { cnpj: nfseConfig.cnpj },
		});
		if (!existing) {
			await prisma.nfseSequence.create({
				data: {
					cnpj: nfseConfig.cnpj,
					lastNumber: 0,
					year: new Date().getFullYear(),
				},
			});
			console.log(
				`  ✓ Sequencia criada para CNPJ ${nfseConfig.cnpj} (ano ${new Date().getFullYear()})`,
			);
		} else {
			console.log(`  ○ Sequencia ja existe para CNPJ ${nfseConfig.cnpj}`);
		}
	} else {
		console.log(
			"  ○ Nenhuma config NFS-e ativa — sequencia sera criada quando ativada",
		);
	}

	// =========================================================================
	// Done
	// =========================================================================
	console.log("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
	console.log("  Seed concluido com sucesso!");
	console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n");
}

main()
	.catch((e) => {
		console.error("\nSeed falhou:", e);
		process.exit(1);
	})
	.finally(async () => {
		await prisma.$disconnect();
	});
