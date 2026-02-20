/**
 * Plan Config - Configuracao estatica de planos (client-safe)
 *
 * Preco FIXO por faixa de instancias WhatsApp.
 * Pricing agressivo para undercut concorrencia em TODOS os tiers.
 *
 * Concorrente:
 *   10 devices  = R$190  (R$19,00/device)
 *   100 devices = R$138  (R$1,38/device)
 *   300 devices = R$195  (R$0,65/device)
 *
 * Nosso (sempre mais barato):
 *   1   = R$9    (R$9,00)   — sem equivalente
 *   10  = R$29   (R$2,90)   — vs R$190 (85% mais barato)
 *   30  = R$59   (R$1,97)   — sem equivalente
 *   100 = R$99   (R$0,99)   — vs R$138 (28% mais barato)
 *   300 = R$149  (R$0,50)   — vs R$195 (24% mais barato)
 *   500 = R$199  (R$0,40)   — sem equivalente
 */

export const PLAN_TIERS = [
	{
		slug: "starter",
		name: "Starter",
		minInstances: 1,
		maxInstances: 1,
		price: 9,
	},
	{
		slug: "pro",
		name: "Pro",
		minInstances: 2,
		maxInstances: 10,
		price: 29,
	},
	{
		slug: "business",
		name: "Business",
		minInstances: 11,
		maxInstances: 30,
		price: 59,
	},
	{
		slug: "scale",
		name: "Scale",
		minInstances: 31,
		maxInstances: 100,
		price: 99,
	},
	{
		slug: "enterprise",
		name: "Enterprise",
		minInstances: 101,
		maxInstances: 300,
		price: 149,
	},
	{
		slug: "ultimate",
		name: "Ultimate",
		minInstances: 301,
		maxInstances: 500,
		price: 199,
	},
] as const;

export type PlanTier = (typeof PLAN_TIERS)[number];
export type PlanSlug = PlanTier["slug"];

/**
 * Features base disponiveis em TODOS os planos
 */
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

/**
 * Features especificas por tier (extras alem da base)
 */
const TIER_FEATURES: Record<PlanSlug, string[]> = {
	starter: [],
	pro: [],
	business: ["SLA garantido 99.9%"],
	scale: ["SLA garantido 99.9%"],
	enterprise: ["SLA garantido 99.9%"],
	ultimate: ["SLA garantido 99.9%"],
};

/**
 * Detecta o plano apropriado baseado na quantidade de instancias
 */
export function detectPlanTierByInstanceCount(instanceCount: number): PlanTier {
	for (const tier of PLAN_TIERS) {
		if (
			instanceCount >= tier.minInstances &&
			instanceCount <= tier.maxInstances
		) {
			return tier;
		}
	}
	// Fallback para Ultimate se acima de 500
	return PLAN_TIERS[5];
}

/**
 * Retorna informacao do plano atual baseado no instanceCount
 */
export function getCurrentPlanInfo(instanceCount: number) {
	const tier = detectPlanTierByInstanceCount(instanceCount);
	return {
		name: tier.name,
		slug: tier.slug,
		price: tier.price,
		maxInstances: tier.maxInstances,
	};
}

/**
 * Retorna o breakdown de preco para exibicao
 */
export function getPriceBreakdown(instanceCount: number) {
	const tier = detectPlanTierByInstanceCount(instanceCount);
	const pricePerInstance = tier.price / instanceCount;

	return {
		planName: tier.name,
		price: tier.price,
		pricePerInstance: Math.round(pricePerInstance * 100) / 100,
		maxInstances: tier.maxInstances,
		instanceCount,
	};
}

/**
 * Retorna todas as features do plano (base + tier)
 */
export function getDisplayFeatures(slug: PlanSlug): string[] {
	const tierFeatures = TIER_FEATURES[slug] ?? [];
	return [...BASE_FEATURES, ...tierFeatures];
}

/**
 * Retorna os checkpoints para badges do slider
 */
export function getPlanCheckpoints() {
	return PLAN_TIERS.map((tier) => ({
		instanceCount: tier.minInstances,
		planName: tier.name,
		planSlug: tier.slug,
		range: `${tier.minInstances}${tier.maxInstances === 500 ? "+" : `-${tier.maxInstances}`}`,
	}));
}

/**
 * Constantes do slider
 */
export const MIN_INSTANCES = 1;
export const MAX_INSTANCES = 500;
