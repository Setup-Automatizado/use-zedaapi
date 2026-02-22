import "dotenv/config";
import { PrismaPg } from "@prisma/adapter-pg";
import { hashPassword } from "better-auth/crypto";
import Stripe from "stripe";
import { PrismaClient } from "../generated/prisma/client";

const adapter = new PrismaPg({
	connectionString: process.env.DATABASE_URL!,
});
const prisma = new PrismaClient({ adapter, log: ["error"] });

// =========================================================================
// Stripe Helpers
// =========================================================================

function getStripe(stripeKey: string): Stripe | null {
	if (!stripeKey || stripeKey.includes("placeholder")) return null;
	return new Stripe(stripeKey, { apiVersion: "2025-08-27.basil" });
}

async function syncStripePrice(
	plan: {
		slug: string;
		name: string;
		description: string;
		price: number;
		currency: string;
		interval: string;
	},
	stripe: Stripe,
): Promise<string | null> {
	try {
		// 1. Find or create Product
		const products = await stripe.products.search({
			query: `metadata['slug']:'${plan.slug}'`,
		});

		let product = products.data[0];
		if (!product) {
			console.log(`  Creating Stripe Product: ${plan.name}...`);
			product = await stripe.products.create({
				name: `Zé da API - ${plan.name}`,
				description: plan.description,
				metadata: { slug: plan.slug },
			});
		}

		// 2. Find or create Price
		const prices = await stripe.prices.list({
			product: product.id,
			active: true,
			limit: 10,
		});

		const unitAmount = Math.round(plan.price * 100);
		let price = prices.data.find(
			(p) =>
				p.unit_amount === unitAmount &&
				p.currency.toLowerCase() === plan.currency.toLowerCase() &&
				p.recurring?.interval === plan.interval,
		);

		if (!price) {
			console.log(
				`  Creating Stripe Price: ${plan.currency} ${plan.price}/${plan.interval}...`,
			);
			price = await stripe.prices.create({
				product: product.id,
				unit_amount: unitAmount,
				currency: plan.currency,
				recurring: {
					interval: plan.interval as Stripe.Price.Recurring.Interval,
				},
				metadata: { slug: plan.slug },
			});
		}

		return price.id;
	} catch (error: unknown) {
		const message = error instanceof Error ? error.message : String(error);
		console.log(`  ❌ Stripe Sync Error (${plan.slug}):`, message);
		return null;
	}
}

// =========================================================================
// Stripe Webhook Auto-Setup
// =========================================================================

const WEBHOOK_URL = "https://zedaapi.com/api/webhooks/stripe";
const WEBHOOK_EVENTS: Stripe.WebhookEndpointCreateParams.EnabledEvent[] = [
	// Checkout
	"checkout.session.completed",
	"checkout.session.expired",
	"checkout.session.async_payment_succeeded",
	// Subscriptions
	"customer.subscription.created",
	"customer.subscription.updated",
	"customer.subscription.deleted",
	// Invoices
	"invoice.paid",
	"invoice.payment_failed",
	// Refunds
	"charge.refunded",
	// Connect
	"account.updated",
	"transfer.created",
	"transfer.updated",
	"transfer.reversed",
];

async function ensureStripeWebhook(stripe: Stripe): Promise<string | null> {
	try {
		// List existing webhooks
		const endpoints = await stripe.webhookEndpoints.list({ limit: 100 });
		const existing = endpoints.data.find((ep) => ep.url === WEBHOOK_URL);

		if (existing) {
			// Update events if needed
			const currentEvents = new Set(existing.enabled_events);
			const needsUpdate = WEBHOOK_EVENTS.some(
				(e) => !currentEvents.has(e),
			);

			if (needsUpdate) {
				console.log("  Updating Stripe webhook events...");
				await stripe.webhookEndpoints.update(existing.id, {
					enabled_events: WEBHOOK_EVENTS,
				});
				console.log(`  ✓ Webhook atualizado: ${WEBHOOK_URL}`);
			} else {
				console.log(`  ○ Webhook ja existe: ${WEBHOOK_URL}`);
			}
			return null; // Can't retrieve secret of existing webhook
		}

		// Create new webhook
		console.log("  Creating Stripe webhook...");
		const endpoint = await stripe.webhookEndpoints.create({
			url: WEBHOOK_URL,
			enabled_events: WEBHOOK_EVENTS,
			description: "ZedaAPI Manager",
		});

		console.log(`  ✓ Webhook criado: ${WEBHOOK_URL}`);
		console.log(`  ✓ Webhook ID: ${endpoint.id}`);
		if (endpoint.secret) {
			console.log(`  ✓ Webhook Secret: ${endpoint.secret}`);
			console.log(
				`  ⚠ ATUALIZE STRIPE_WEBHOOK_SECRET no docker-compose.prod.yml!`,
			);
		}

		return endpoint.secret ?? null;
	} catch (error: unknown) {
		const message = error instanceof Error ? error.message : String(error);
		console.log(`  ❌ Webhook Error:`, message);
		return null;
	}
}

// =========================================================================
// Features — EXATAMENTE iguais a landing page (lib/billing/plan-config.ts)
// =========================================================================

const BASE_FEATURES = [
	"Envio e recebimento ilimitado",
	"Webhooks em tempo real",
	"API REST completa",
	"Botões interativos",
	"Gerenciamento de grupos",
	"Multi-dispositivo",
	"Métricas e analytics",
	"Nodes community n8n",
	"Documentação completa",
	"Suporte por e-mail e WhatsApp",
];

const PREMIUM_FEATURES = [...BASE_FEATURES, "SLA garantido 99.9%"];

// =========================================================================
// Plan Definitions — Monthly + Annual (20% discount)
// =========================================================================

interface PlanDef {
	name: string;
	slug: string;
	description: string;
	price: number;
	currency: string;
	interval: string;
	maxInstances: number;
	features: string[];
	active: boolean;
	sortOrder: number;
}

function buildPlans(): PlanDef[] {
	const tiers = [
		{
			name: "Starter",
			slug: "starter",
			description:
				"Ideal para quem está começando. 1 instância WhatsApp por apenas R$9/mês.",
			monthlyPrice: 9,
			maxInstances: 1,
			features: BASE_FEATURES,
			sortOrder: 1,
		},
		{
			name: "Pro",
			slug: "pro",
			description:
				"Para equipes em crescimento. Até 10 instâncias por R$2,90/cada.",
			monthlyPrice: 29,
			maxInstances: 10,
			features: BASE_FEATURES,
			sortOrder: 2,
		},
		{
			name: "Business",
			slug: "business",
			description:
				"Para operações em escala. Até 30 instâncias por R$1,97/cada.",
			monthlyPrice: 59,
			maxInstances: 30,
			features: PREMIUM_FEATURES,
			sortOrder: 3,
		},
		{
			name: "Scale",
			slug: "scale",
			description:
				"Para quem precisa escalar. Até 100 instâncias por R$0,99/cada.",
			monthlyPrice: 99,
			maxInstances: 100,
			features: PREMIUM_FEATURES,
			sortOrder: 4,
		},
		{
			name: "Enterprise",
			slug: "enterprise",
			description:
				"Para grandes operações. Até 300 instâncias por R$0,50/cada.",
			monthlyPrice: 149,
			maxInstances: 300,
			features: PREMIUM_FEATURES,
			sortOrder: 5,
		},
		{
			name: "Ultimate",
			slug: "ultimate",
			description:
				"Capacidade máxima. Até 500 instâncias por apenas R$0,40/cada.",
			monthlyPrice: 199,
			maxInstances: 500,
			features: PREMIUM_FEATURES,
			sortOrder: 6,
		},
	];

	const plans: PlanDef[] = [];

	for (const tier of tiers) {
		// Monthly plan
		plans.push({
			name: tier.name,
			slug: tier.slug,
			description: tier.description,
			price: tier.monthlyPrice,
			currency: "BRL",
			interval: "month",
			maxInstances: tier.maxInstances,
			features: tier.features,
			active: true,
			sortOrder: tier.sortOrder,
		});

		// Annual plan (20% discount — matches landing page toggle)
		const annualMonthly = Math.round(tier.monthlyPrice * 0.8);
		const annualTotal = annualMonthly * 12;
		plans.push({
			name: `${tier.name} Anual`,
			slug: `${tier.slug}-annual`,
			description: `${tier.description} Economia de 20% no plano anual.`,
			price: annualTotal,
			currency: "BRL",
			interval: "year",
			maxInstances: tier.maxInstances,
			features: tier.features,
			active: true,
			sortOrder: tier.sortOrder + 100, // Annual plans after monthly
		});
	}

	return plans;
}

// =========================================================================
// Main
// =========================================================================

async function main() {
	console.log("Seeding database...\n");

	const stripeKey = process.env.STRIPE_SECRET_KEY || "";
	const stripe = getStripe(stripeKey);

	// =========================================================================
	// 1. Stripe Webhook (auto-setup)
	// =========================================================================
	console.log("[Stripe Webhook]");
	if (stripe) {
		await ensureStripeWebhook(stripe);
	} else {
		console.log("  ⚠ Skipping Stripe webhook (Missing/Invalid Key)");
	}

	// =========================================================================
	// 2. Plans — Monthly + Annual with Stripe sync
	// =========================================================================
	console.log("\n[Plans]");
	const plans = buildPlans();

	for (const plan of plans) {
		let stripePriceId: string | null = null;

		if (stripe) {
			stripePriceId = await syncStripePrice(plan, stripe);
		}

		await prisma.plan.upsert({
			where: { slug: plan.slug },
			update: {
				name: plan.name,
				description: plan.description,
				price: plan.price,
				currency: plan.currency,
				interval: plan.interval,
				maxInstances: plan.maxInstances,
				features: plan.features,
				active: plan.active,
				sortOrder: plan.sortOrder,
				...(stripePriceId ? { stripePriceId } : {}),
			},
			create: {
				...plan,
				stripePriceId,
			},
		});

		const intervalLabel = plan.interval === "year" ? "ano" : "mês";
		console.log(
			`  ✓ "${plan.name}" — R$${plan.price}/${intervalLabel} [${stripePriceId || "No Stripe ID"}]`,
		);
	}

	// =========================================================================
	// 3. Admin User
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
		if (existingAdmin.role !== "admin") {
			await prisma.user.update({
				where: { id: existingAdmin.id },
				data: { role: "admin" },
			});
			console.log(`  ✓ Atualizado para admin: ${adminEmail}`);
		} else {
			console.log(`  ○ Ja existe: ${adminEmail}`);
		}
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
	// 4. NFS-e Config (placeholder - inativo por padrao)
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
	// 5. Feature Flags
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
	// 6. System Settings
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
	// 7. Email Templates
	// =========================================================================
	console.log("\n[Email Templates]");
	const emailTemplates = [
		// =====================================================================
		// Auth & Conta (5 templates)
		// =====================================================================
		{
			slug: "welcome",
			name: "Boas-vindas",
			subject: "Bem-vindo ao Zé da API Manager!",
			body: "Olá {{name}}, seja bem-vindo ao Zé da API Manager! Sua conta foi criada com sucesso. Acesse o painel para começar: {{dashboardUrl}}",
			variables: JSON.stringify(["name", "email", "dashboardUrl"]),
			active: true,
		},
		{
			slug: "verify-email",
			name: "Verificação de e-mail",
			subject: "Verifique seu e-mail — Zé da API Manager",
			body: "Olá {{name}}, clique no link abaixo para verificar seu e-mail: {{verifyUrl}}. Este link expira em 24 horas.",
			variables: JSON.stringify(["name", "email", "verifyUrl"]),
			active: true,
		},
		{
			slug: "reset-password",
			name: "Redefinição de senha",
			subject: "Redefinir senha — Zé da API Manager",
			body: "Olá {{name}}, clique no link para redefinir sua senha: {{resetUrl}}. Este link expira em 1 hora. Se você não solicitou a redefinição, ignore este e-mail.",
			variables: JSON.stringify(["name", "email", "resetUrl"]),
			active: true,
		},
		{
			slug: "magic-link",
			name: "Login via link mágico",
			subject: "Seu link de acesso — Zé da API Manager",
			body: "Olá {{name}}, clique no link abaixo para acessar sua conta: {{magicLinkUrl}}. Este link expira em 15 minutos e só pode ser usado uma vez.",
			variables: JSON.stringify(["name", "email", "magicLinkUrl"]),
			active: true,
		},
		{
			slug: "account-deactivated",
			name: "Conta desativada",
			subject: "Sua conta foi desativada — Zé da API Manager",
			body: "Olá {{name}}, sua conta no Zé da API Manager foi desativada. Suas instâncias foram desconectadas. Se deseja reativar, entre em contato: {{supportEmail}}",
			variables: JSON.stringify([
				"name",
				"email",
				"supportEmail",
				"deactivatedAt",
			]),
			active: true,
		},

		// =====================================================================
		// Segurança (5 templates)
		// =====================================================================
		{
			slug: "two-factor-enabled",
			name: "2FA ativado",
			subject: "Autenticação de dois fatores ativada — Zé da API Manager",
			body: "Olá {{name}}, a autenticação de dois fatores (2FA) foi ativada com sucesso na sua conta. Guarde seus códigos de recuperação em local seguro. Se você não ativou o 2FA, entre em contato imediatamente: {{supportEmail}}",
			variables: JSON.stringify([
				"name",
				"email",
				"enabledAt",
				"supportEmail",
			]),
			active: true,
		},
		{
			slug: "two-factor-code",
			name: "Código de verificação 2FA",
			subject: "Código de verificação: {{code}} — Zé da API Manager",
			body: "Olá {{name}}, seu código de verificação é: {{code}}. Este código expira em {{expiresIn}} minutos. Não compartilhe este código com ninguém.",
			variables: JSON.stringify(["name", "code", "expiresIn"]),
			active: true,
		},
		{
			slug: "password-changed",
			name: "Senha alterada",
			subject: "Sua senha foi alterada — Zé da API Manager",
			body: "Olá {{name}}, sua senha foi alterada com sucesso em {{changedAt}}. Se você não realizou esta alteração, redefina sua senha imediatamente: {{resetUrl}} e entre em contato com o suporte.",
			variables: JSON.stringify([
				"name",
				"changedAt",
				"resetUrl",
				"ipAddress",
				"userAgent",
			]),
			active: true,
		},
		{
			slug: "email-changed",
			name: "E-mail alterado",
			subject: "Seu e-mail foi alterado — Zé da API Manager",
			body: "Olá {{name}}, o e-mail da sua conta foi alterado de {{oldEmail}} para {{newEmail}}. Se você não realizou esta alteração, entre em contato imediatamente: {{supportEmail}}",
			variables: JSON.stringify([
				"name",
				"oldEmail",
				"newEmail",
				"changedAt",
				"supportEmail",
			]),
			active: true,
		},
		{
			slug: "login-new-device",
			name: "Login em novo dispositivo",
			subject: "Novo acesso detectado na sua conta — Zé da API Manager",
			body: "Olá {{name}}, detectamos um novo acesso na sua conta. Dispositivo: {{deviceName}}. Localização: {{location}}. Data: {{loginAt}}. Se não foi você, altere sua senha imediatamente: {{changePasswordUrl}}",
			variables: JSON.stringify([
				"name",
				"deviceName",
				"location",
				"ipAddress",
				"loginAt",
				"changePasswordUrl",
			]),
			active: true,
		},

		// =====================================================================
		// Ciclo de Vida da Assinatura (8 templates)
		// =====================================================================
		{
			slug: "subscription-created",
			name: "Assinatura criada",
			subject: "Assinatura ativada — Plano {{planName}}",
			body: "Olá {{name}}, sua assinatura do plano {{planName}} foi ativada com sucesso! Você pode criar até {{maxInstances}} instâncias WhatsApp. Valor: R${{price}}/{{interval}}. Acesse o painel: {{dashboardUrl}}",
			variables: JSON.stringify([
				"name",
				"planName",
				"maxInstances",
				"price",
				"interval",
				"dashboardUrl",
			]),
			active: true,
		},
		{
			slug: "subscription-canceled",
			name: "Assinatura cancelada",
			subject: "Assinatura cancelada — Plano {{planName}}",
			body: "Olá {{name}}, sua assinatura do plano {{planName}} foi cancelada. Suas instâncias continuarão funcionando até {{endDate}}. Se mudar de ideia, reative sua assinatura: {{reactivateUrl}}",
			variables: JSON.stringify([
				"name",
				"planName",
				"endDate",
				"reactivateUrl",
			]),
			active: true,
		},
		{
			slug: "subscription-renewed",
			name: "Assinatura renovada",
			subject: "Assinatura renovada — Plano {{planName}}",
			body: "Olá {{name}}, sua assinatura do plano {{planName}} foi renovada automaticamente. Valor: R${{amount}}. Próximo vencimento: {{nextBillingDate}}. Fatura disponível em: {{invoiceUrl}}",
			variables: JSON.stringify([
				"name",
				"planName",
				"amount",
				"nextBillingDate",
				"invoiceUrl",
			]),
			active: true,
		},
		{
			slug: "subscription-upgraded",
			name: "Upgrade de plano",
			subject: "Upgrade realizado — {{oldPlan}} para {{newPlan}}",
			body: "Olá {{name}}, seu plano foi atualizado de {{oldPlan}} para {{newPlan}}! Agora você tem acesso a até {{maxInstances}} instâncias. Novo valor: R${{newPrice}}/{{interval}}. O valor proporcional será ajustado na próxima fatura.",
			variables: JSON.stringify([
				"name",
				"oldPlan",
				"newPlan",
				"maxInstances",
				"newPrice",
				"interval",
				"effectiveDate",
			]),
			active: true,
		},
		{
			slug: "subscription-downgraded",
			name: "Downgrade de plano",
			subject: "Plano alterado — {{oldPlan}} para {{newPlan}}",
			body: "Olá {{name}}, seu plano será alterado de {{oldPlan}} para {{newPlan}} a partir de {{effectiveDate}}. Novo limite: {{maxInstances}} instâncias. Novo valor: R${{newPrice}}/{{interval}}. Verifique se suas instâncias ativas estão dentro do novo limite.",
			variables: JSON.stringify([
				"name",
				"oldPlan",
				"newPlan",
				"maxInstances",
				"newPrice",
				"interval",
				"effectiveDate",
				"currentInstances",
			]),
			active: true,
		},
		{
			slug: "subscription-resumed",
			name: "Assinatura reativada",
			subject: "Assinatura reativada — Plano {{planName}}",
			body: "Olá {{name}}, sua assinatura do plano {{planName}} foi reativada com sucesso! Suas instâncias continuam funcionando normalmente. Próximo vencimento: {{nextBillingDate}}.",
			variables: JSON.stringify([
				"name",
				"planName",
				"nextBillingDate",
				"dashboardUrl",
			]),
			active: true,
		},
		{
			slug: "trial-started",
			name: "Trial iniciado",
			subject:
				"Seu período de avaliação começou — {{trialDays}} dias grátis!",
			body: "Olá {{name}}, seu período de avaliação do plano {{planName}} começou! Você tem {{trialDays}} dias grátis para testar todas as funcionalidades. O trial expira em {{trialEndDate}}. Nenhuma cobrança será feita até lá.",
			variables: JSON.stringify([
				"name",
				"planName",
				"trialDays",
				"trialEndDate",
				"dashboardUrl",
			]),
			active: true,
		},
		{
			slug: "trial-ending",
			name: "Trial prestes a expirar",
			subject:
				"Seu trial expira em {{daysLeft}} dias — Escolha seu plano",
			body: "Olá {{name}}, seu período de avaliação expira em {{daysLeft}} dias ({{trialEndDate}}). Para continuar usando o Zé da API Manager, escolha um plano: {{pricingUrl}}. Após o término do trial, suas instâncias serão desconectadas.",
			variables: JSON.stringify([
				"name",
				"planName",
				"daysLeft",
				"trialEndDate",
				"pricingUrl",
			]),
			active: true,
		},

		// =====================================================================
		// Pagamento e Faturamento (10 templates)
		// =====================================================================
		{
			slug: "invoice-paid",
			name: "Fatura paga",
			subject: "Pagamento confirmado — Fatura #{{invoiceNumber}}",
			body: "Olá {{name}}, seu pagamento de R${{amount}} foi confirmado com sucesso. Fatura #{{invoiceNumber}}. Método: {{paymentMethod}}. Acesse o comprovante: {{pdfUrl}}",
			variables: JSON.stringify([
				"name",
				"amount",
				"invoiceNumber",
				"paymentMethod",
				"pdfUrl",
				"paidAt",
			]),
			active: true,
		},
		{
			slug: "invoice-created",
			name: "Nova fatura gerada",
			subject: "Nova fatura #{{invoiceNumber}} — R${{amount}}",
			body: "Olá {{name}}, uma nova fatura foi gerada para sua assinatura. Valor: R${{amount}}. Vencimento: {{dueDate}}. Acesse para realizar o pagamento: {{paymentUrl}}",
			variables: JSON.stringify([
				"name",
				"amount",
				"invoiceNumber",
				"dueDate",
				"paymentUrl",
				"planName",
			]),
			active: true,
		},
		{
			slug: "payment-failed",
			name: "Falha no pagamento",
			subject: "Falha no pagamento — Ação necessária",
			body: "Olá {{name}}, houve uma falha no pagamento da sua assinatura (Plano {{planName}}). Valor: R${{amount}}. Por favor, atualize seu método de pagamento até {{dueDate}} para evitar a suspensão do serviço: {{retryUrl}}",
			variables: JSON.stringify([
				"name",
				"planName",
				"amount",
				"retryUrl",
				"dueDate",
				"attemptCount",
			]),
			active: true,
		},
		{
			slug: "charge-refunded",
			name: "Reembolso processado",
			subject: "Reembolso de R${{amount}} processado",
			body: "Olá {{name}}, seu reembolso de R${{amount}} referente à fatura #{{invoiceNumber}} foi processado com sucesso. O valor será creditado no seu método de pagamento original em até 10 dias úteis.",
			variables: JSON.stringify([
				"name",
				"amount",
				"invoiceNumber",
				"refundedAt",
				"reason",
			]),
			active: true,
		},
		{
			slug: "pix-charge-created",
			name: "Cobrança PIX gerada",
			subject: "PIX gerado — R${{amount}} — Zé da API Manager",
			body: "Olá {{name}}, sua cobrança PIX de R${{amount}} foi gerada. Copie o código PIX abaixo para pagar: {{pixCopiaECola}}. Este código expira em {{expiresAt}}. Após o pagamento, sua assinatura será ativada automaticamente.",
			variables: JSON.stringify([
				"name",
				"amount",
				"pixCopiaECola",
				"expiresAt",
				"invoiceNumber",
				"planName",
			]),
			active: true,
		},
		{
			slug: "boleto-charge-created",
			name: "Boleto gerado",
			subject: "Boleto gerado — R${{amount}} — Zé da API Manager",
			body: "Olá {{name}}, seu boleto de R${{amount}} foi gerado. Linha digitável: {{linhaDigitavel}}. Vencimento: {{dueDate}}. O boleto também possui QR Code PIX para pagamento imediato. Após a compensação, sua assinatura será ativada automaticamente.",
			variables: JSON.stringify([
				"name",
				"amount",
				"linhaDigitavel",
				"codigoBarras",
				"pixCopiaECola",
				"dueDate",
				"invoiceNumber",
				"planName",
			]),
			active: true,
		},
		{
			slug: "pix-expired",
			name: "PIX expirado",
			subject: "Cobrança PIX expirada — Gere um novo pagamento",
			body: "Olá {{name}}, sua cobrança PIX de R${{amount}} expirou sem pagamento. Para continuar com a assinatura do plano {{planName}}, gere um novo pagamento: {{retryUrl}}",
			variables: JSON.stringify([
				"name",
				"amount",
				"planName",
				"retryUrl",
				"expiredAt",
			]),
			active: true,
		},
		{
			slug: "upcoming-renewal",
			name: "Renovação próxima",
			subject: "Sua assinatura será renovada em {{daysLeft}} dias",
			body: "Olá {{name}}, sua assinatura do plano {{planName}} será renovada automaticamente em {{renewalDate}}. Valor: R${{amount}}. Se deseja alterar o plano ou cancelar, acesse: {{billingUrl}}",
			variables: JSON.stringify([
				"name",
				"planName",
				"amount",
				"renewalDate",
				"daysLeft",
				"billingUrl",
			]),
			active: true,
		},
		{
			slug: "payment-method-updated",
			name: "Método de pagamento atualizado",
			subject: "Método de pagamento atualizado — Zé da API Manager",
			body: "Olá {{name}}, seu método de pagamento foi atualizado com sucesso. Novo método: {{newMethod}}. Se você não realizou esta alteração, entre em contato: {{supportEmail}}",
			variables: JSON.stringify([
				"name",
				"newMethod",
				"updatedAt",
				"supportEmail",
			]),
			active: true,
		},
		{
			slug: "subscription-past-due",
			name: "Assinatura em atraso",
			subject: "Pagamento em atraso — Ação necessária",
			body: "Olá {{name}}, sua assinatura do plano {{planName}} está com pagamento em atraso. Valor pendente: R${{amount}}. Suas instâncias podem ser suspensas em {{suspensionDate}}. Atualize o pagamento: {{retryUrl}}",
			variables: JSON.stringify([
				"name",
				"planName",
				"amount",
				"retryUrl",
				"suspensionDate",
			]),
			active: true,
		},

		// =====================================================================
		// NFS-e (2 templates)
		// =====================================================================
		{
			slug: "nfse-issued",
			name: "NFS-e emitida",
			subject: "NFS-e #{{nfseNumber}} emitida com sucesso",
			body: "Olá {{name}}, sua NFS-e #{{nfseNumber}} no valor de R${{amount}} foi emitida com sucesso. Acesse o DANFSE: {{pdfUrl}}. XML disponível em: {{xmlUrl}}",
			variables: JSON.stringify([
				"name",
				"nfseNumber",
				"amount",
				"pdfUrl",
				"xmlUrl",
				"chaveAcesso",
			]),
			active: true,
		},
		{
			slug: "nfse-cancelled",
			name: "NFS-e cancelada",
			subject: "NFS-e #{{nfseNumber}} cancelada",
			body: "Olá {{name}}, a NFS-e #{{nfseNumber}} no valor de R${{amount}} foi cancelada. Motivo: {{reason}}. Uma nova NFS-e será emitida se aplicável.",
			variables: JSON.stringify([
				"name",
				"nfseNumber",
				"amount",
				"reason",
				"cancelledAt",
			]),
			active: true,
		},

		// =====================================================================
		// Gerenciamento de Instâncias (4 templates)
		// =====================================================================
		{
			slug: "instance-created",
			name: "Instância criada",
			subject: "Instância {{instanceName}} criada com sucesso",
			body: "Olá {{name}}, sua instância WhatsApp '{{instanceName}}' foi criada com sucesso! Próximo passo: conecte escaneando o QR Code no painel: {{connectUrl}}",
			variables: JSON.stringify([
				"name",
				"instanceName",
				"connectUrl",
				"planName",
				"instancesUsed",
				"instancesMax",
			]),
			active: true,
		},
		{
			slug: "instance-connected",
			name: "Instância conectada",
			subject: "Instância {{instanceName}} conectada ao WhatsApp",
			body: "Olá {{name}}, a instância '{{instanceName}}' foi conectada com sucesso ao WhatsApp ({{phone}}). Sua API está pronta para enviar e receber mensagens. Configure seus webhooks: {{webhookUrl}}",
			variables: JSON.stringify([
				"name",
				"instanceName",
				"phone",
				"webhookUrl",
				"apiDocsUrl",
			]),
			active: true,
		},
		{
			slug: "instance-disconnected",
			name: "Instância desconectada",
			subject: "Alerta: Instância {{instanceName}} desconectada",
			body: "Olá {{name}}, a instância '{{instanceName}}' ({{phone}}) foi desconectada do WhatsApp. Motivo possível: {{reason}}. Reconecte pelo painel: {{reconnectUrl}}",
			variables: JSON.stringify([
				"name",
				"instanceName",
				"phone",
				"reason",
				"reconnectUrl",
				"disconnectedAt",
			]),
			active: true,
		},
		{
			slug: "instance-deleted",
			name: "Instância removida",
			subject: "Instância {{instanceName}} removida",
			body: "Olá {{name}}, a instância '{{instanceName}}' ({{phone}}) foi removida da sua conta. Esta ação não pode ser desfeita. Você pode criar uma nova instância a qualquer momento: {{dashboardUrl}}",
			variables: JSON.stringify([
				"name",
				"instanceName",
				"phone",
				"deletedAt",
				"dashboardUrl",
			]),
			active: true,
		},

		// =====================================================================
		// Lista de Espera (2 templates)
		// =====================================================================
		{
			slug: "waitlist-approved",
			name: "Aprovado na lista de espera",
			subject: "Você foi aprovado! Crie sua conta no Zé da API Manager",
			body: "Olá {{name}}, você foi aprovado na lista de espera! Crie sua conta agora: {{registerUrl}}. Esta aprovação é válida por 7 dias.",
			variables: JSON.stringify([
				"name",
				"email",
				"registerUrl",
				"approvedAt",
			]),
			active: true,
		},
		{
			slug: "waitlist-registered",
			name: "Inscrito na lista de espera",
			subject: "Você está na lista de espera — Zé da API Manager",
			body: "Olá {{name}}, recebemos sua inscrição na lista de espera do Zé da API Manager! Sua posição: #{{position}}. Avisaremos assim que houver uma vaga disponível. Obrigado pela paciência!",
			variables: JSON.stringify([
				"name",
				"email",
				"position",
				"registeredAt",
			]),
			active: true,
		},

		// =====================================================================
		// Sistema de Afiliados (6 templates)
		// =====================================================================
		{
			slug: "affiliate-registered",
			name: "Registro de afiliado",
			subject: "Solicitação de afiliado recebida — Zé da API Manager",
			body: "Olá {{name}}, recebemos sua solicitação para o programa de afiliados! Seu código de referência é: {{affiliateCode}}. Comece a indicar e ganhe {{commissionRate}}% de comissão: {{dashboardUrl}}",
			variables: JSON.stringify([
				"name",
				"affiliateCode",
				"commissionRate",
				"dashboardUrl",
			]),
			active: true,
		},
		{
			slug: "affiliate-approved",
			name: "Afiliado aprovado",
			subject: "Parabéns! Você é um afiliado ZedaAPI",
			body: "Olá {{name}}, sua conta de afiliado foi ativada! Seu código: {{affiliateCode}}. Comissão: {{commissionRate}}%. Compartilhe seu link de indicação e acompanhe seus ganhos: {{dashboardUrl}}",
			variables: JSON.stringify([
				"name",
				"affiliateCode",
				"commissionRate",
				"dashboardUrl",
				"referralLink",
			]),
			active: true,
		},
		{
			slug: "commission-earned",
			name: "Comissão recebida",
			subject: "Nova comissão de R${{amount}} — Zé da API Manager",
			body: "Olá {{name}}, você ganhou uma nova comissão de R${{amount}} pela indicação de {{referredName}}! Plano contratado: {{planName}}. Total acumulado: R${{totalEarnings}}. Veja seus ganhos: {{dashboardUrl}}",
			variables: JSON.stringify([
				"name",
				"amount",
				"referredName",
				"planName",
				"totalEarnings",
				"dashboardUrl",
			]),
			active: true,
		},
		{
			slug: "referral-converted",
			name: "Indicação convertida",
			subject:
				"Indicação convertida! {{referredName}} assinou o plano {{planName}}",
			body: "Olá {{name}}, sua indicação {{referredName}} acabou de assinar o plano {{planName}}! Uma comissão de R${{commissionAmount}} foi registrada. Acompanhe suas indicações: {{dashboardUrl}}",
			variables: JSON.stringify([
				"name",
				"referredName",
				"planName",
				"commissionAmount",
				"dashboardUrl",
			]),
			active: true,
		},
		{
			slug: "payout-processed",
			name: "Pagamento de comissão",
			subject: "Pagamento de R${{amount}} processado",
			body: "Olá {{name}}, seu pagamento de comissão de R${{amount}} foi processado com sucesso via {{method}}. Data: {{processedAt}}. O valor será creditado em até 3 dias úteis.",
			variables: JSON.stringify([
				"name",
				"amount",
				"method",
				"processedAt",
				"payoutId",
			]),
			active: true,
		},
		{
			slug: "payout-failed",
			name: "Falha no pagamento de comissão",
			subject: "Falha no pagamento de comissão — Ação necessária",
			body: "Olá {{name}}, houve uma falha ao processar seu pagamento de comissão de R${{amount}}. Motivo: {{reason}}. Verifique seus dados bancários/Stripe Connect e tente novamente: {{dashboardUrl}}",
			variables: JSON.stringify([
				"name",
				"amount",
				"reason",
				"dashboardUrl",
			]),
			active: true,
		},

		// =====================================================================
		// Organização e Equipe (3 templates)
		// =====================================================================
		{
			slug: "member-invited",
			name: "Convite para organização",
			subject:
				"Você foi convidado para {{organizationName}} — Zé da API Manager",
			body: "Olá {{name}}, {{inviterName}} convidou você para a organização '{{organizationName}}' no Zé da API Manager. Função: {{role}}. Aceite o convite: {{acceptUrl}}. Este convite expira em 7 dias.",
			variables: JSON.stringify([
				"name",
				"inviterName",
				"organizationName",
				"role",
				"acceptUrl",
			]),
			active: true,
		},
		{
			slug: "member-joined",
			name: "Membro entrou na organização",
			subject:
				"{{memberName}} entrou na organização {{organizationName}}",
			body: "Olá {{name}}, {{memberName}} ({{memberEmail}}) aceitou o convite e entrou na organização '{{organizationName}}' com a função de {{role}}. Gerencie membros: {{manageUrl}}",
			variables: JSON.stringify([
				"name",
				"memberName",
				"memberEmail",
				"organizationName",
				"role",
				"manageUrl",
			]),
			active: true,
		},
		{
			slug: "member-removed",
			name: "Membro removido da organização",
			subject: "Você foi removido da organização {{organizationName}}",
			body: "Olá {{name}}, você foi removido da organização '{{organizationName}}' por {{removedBy}}. Você não tem mais acesso aos recursos desta organização. Se acredita que isso foi um erro, entre em contato com o administrador.",
			variables: JSON.stringify([
				"name",
				"organizationName",
				"removedBy",
				"removedAt",
			]),
			active: true,
		},

		// =====================================================================
		// Chaves API (2 templates)
		// =====================================================================
		{
			slug: "api-key-created",
			name: "Chave API criada",
			subject: "Nova chave API criada — Zé da API Manager",
			body: "Olá {{name}}, uma nova chave API foi criada. Nome: {{keyName}}. Prefixo: {{keyPrefix}}... Esta é a única vez que você pode ver a chave completa. Guarde-a em local seguro.",
			variables: JSON.stringify([
				"name",
				"keyName",
				"keyPrefix",
				"createdAt",
				"docsUrl",
			]),
			active: true,
		},
		{
			slug: "api-key-revoked",
			name: "Chave API revogada",
			subject: "Chave API revogada — Zé da API Manager",
			body: "Olá {{name}}, a chave API '{{keyName}}' ({{keyPrefix}}...) foi revogada e não pode mais ser usada. Se você não revogou esta chave, verifique a segurança da sua conta: {{securityUrl}}",
			variables: JSON.stringify([
				"name",
				"keyName",
				"keyPrefix",
				"revokedAt",
				"securityUrl",
			]),
			active: true,
		},

		// =====================================================================
		// Notificações Admin (5 templates)
		// =====================================================================
		{
			slug: "admin-new-user",
			name: "Admin: novo usuário",
			subject: "[Admin] Novo usuário: {{userName}} ({{userEmail}})",
			body: "Novo usuário cadastrado no Zé da API Manager. Nome: {{userName}}. E-mail: {{userEmail}}. País: {{country}}. Data: {{registeredAt}}. Origem: {{source}}. Total de usuários: {{totalUsers}}.",
			variables: JSON.stringify([
				"userName",
				"userEmail",
				"country",
				"registeredAt",
				"source",
				"totalUsers",
			]),
			active: true,
		},
		{
			slug: "admin-new-subscription",
			name: "Admin: nova assinatura",
			subject: "[Admin] Nova assinatura: {{userName}} — {{planName}}",
			body: "Nova assinatura criada. Usuário: {{userName}} ({{userEmail}}). Plano: {{planName}}. Valor: R${{amount}}/{{interval}}. Método: {{paymentMethod}}. Receita mensal total: R${{totalMRR}}.",
			variables: JSON.stringify([
				"userName",
				"userEmail",
				"planName",
				"amount",
				"interval",
				"paymentMethod",
				"totalMRR",
			]),
			active: true,
		},
		{
			slug: "admin-nfse-error",
			name: "Admin: erro NFS-e",
			subject:
				"[Admin] Erro na emissão NFS-e — Fatura #{{invoiceNumber}}",
			body: "Falha na emissão de NFS-e. Fatura: #{{invoiceNumber}}. Usuário: {{userName}} ({{userEmail}}). Valor: R${{amount}}. Erro: {{errorMessage}}. Tentativas: {{retryCount}}. Verifique no painel admin.",
			variables: JSON.stringify([
				"invoiceNumber",
				"userName",
				"userEmail",
				"amount",
				"errorMessage",
				"retryCount",
			]),
			active: true,
		},
		{
			slug: "admin-subscription-canceled",
			name: "Admin: assinatura cancelada",
			subject: "[Admin] Cancelamento: {{userName}} — {{planName}}",
			body: "Assinatura cancelada. Usuário: {{userName}} ({{userEmail}}). Plano: {{planName}}. Motivo: {{reason}}. Ativa desde: {{activeSince}}. Instâncias: {{instanceCount}}. Impacto no MRR: -R${{amount}}.",
			variables: JSON.stringify([
				"userName",
				"userEmail",
				"planName",
				"reason",
				"activeSince",
				"instanceCount",
				"amount",
			]),
			active: true,
		},
		{
			slug: "admin-support-request",
			name: "Admin: ticket de suporte",
			subject: "[Admin] Novo ticket de suporte: {{subject}}",
			body: "Novo ticket de suporte recebido. De: {{userName}} ({{userEmail}}). Assunto: {{subject}}. Plano: {{planName}}. Prioridade: {{priority}}. Mensagem: {{message}}",
			variables: JSON.stringify([
				"userName",
				"userEmail",
				"subject",
				"planName",
				"priority",
				"message",
			]),
			active: true,
		},

		// =====================================================================
		// Sistema e Conformidade (4 templates)
		// =====================================================================
		{
			slug: "data-deletion-requested",
			name: "Solicitação de exclusão de dados",
			subject:
				"Solicitação de exclusão de dados recebida — Zé da API Manager",
			body: "Olá {{name}}, recebemos sua solicitação de exclusão de dados (LGPD). Protocolo: {{requestId}}. Seus dados serão excluídos em até 15 dias úteis. Você receberá uma confirmação quando o processo for concluído.",
			variables: JSON.stringify([
				"name",
				"email",
				"requestId",
				"requestedAt",
			]),
			active: true,
		},
		{
			slug: "data-deletion-completed",
			name: "Exclusão de dados concluída",
			subject: "Exclusão de dados concluída — Zé da API Manager",
			body: "Olá {{name}}, a exclusão dos seus dados foi concluída com sucesso. Protocolo: {{requestId}}. Todos os seus dados pessoais, instâncias e histórico foram permanentemente removidos conforme a LGPD.",
			variables: JSON.stringify([
				"name",
				"email",
				"requestId",
				"completedAt",
			]),
			active: true,
		},
		{
			slug: "contact-form-received",
			name: "Formulário de contato recebido",
			subject: "Recebemos sua mensagem — Zé da API Manager",
			body: "Olá {{name}}, recebemos sua mensagem e responderemos em até 24 horas úteis. Protocolo: {{ticketId}}. Assunto: {{subject}}. Enquanto isso, confira nossa documentação: {{docsUrl}}",
			variables: JSON.stringify([
				"name",
				"email",
				"ticketId",
				"subject",
				"docsUrl",
			]),
			active: true,
		},
		{
			slug: "maintenance-scheduled",
			name: "Manutenção programada",
			subject: "Manutenção programada — {{maintenanceDate}}",
			body: "Olá {{name}}, informamos que uma manutenção programada será realizada em {{maintenanceDate}} das {{startTime}} às {{endTime}} (horário de Brasília). Durante este período, o serviço pode ficar indisponível. Suas instâncias serão reconectadas automaticamente após a manutenção.",
			variables: JSON.stringify([
				"name",
				"maintenanceDate",
				"startTime",
				"endTime",
				"reason",
				"statusPageUrl",
			]),
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
	// 8. NFS-e Sequence (initialize if not exists)
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
	console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
	console.log("\n  Planos criados: 12 (6 mensais + 6 anuais)");
	console.log(
		`  Email templates: ${emailTemplates.length} (todos os fluxos)`,
	);
	console.log("  Webhook Stripe: https://zedaapi.com/api/webhooks/stripe");
	console.log("  Eventos inscritos: 13 tipos");
	console.log("");
}

main()
	.catch((e) => {
		console.error("\nSeed falhou:", e);
		process.exit(1);
	})
	.finally(async () => {
		await prisma.$disconnect();
	});
