import type { Variants } from "motion-dom";

// Instance status config - consolidates 5 duplicated statusConfig objects
export const INSTANCE_STATUS_CONFIG: Record<
	string,
	{ label: string; className: string; dot: string }
> = {
	connected: {
		label: "Conectado",
		className: "bg-primary/10 text-primary",
		dot: "bg-primary",
	},
	connecting: {
		label: "Conectando",
		className: "bg-chart-2/10 text-chart-2",
		dot: "bg-chart-2",
	},
	disconnected: {
		label: "Desconectado",
		className: "bg-muted text-muted-foreground",
		dot: "bg-muted-foreground",
	},
	error: {
		label: "Erro",
		className: "bg-destructive/10 text-destructive",
		dot: "bg-destructive",
	},
	banned: {
		label: "Banido",
		className: "bg-destructive/10 text-destructive",
		dot: "bg-destructive",
	},
	creating: {
		label: "Criando",
		className: "bg-chart-2/10 text-chart-2",
		dot: "bg-chart-2",
	},
};

// Subscription status config
export const SUBSCRIPTION_STATUS_CONFIG: Record<
	string,
	{ label: string; className: string }
> = {
	active: { label: "Ativo", className: "text-primary" },
	trialing: { label: "Teste", className: "text-chart-2" },
	past_due: { label: "Atrasado", className: "text-destructive" },
	canceled: { label: "Cancelado", className: "text-muted-foreground" },
};

// Invoice status config
export const INVOICE_STATUS_CONFIG: Record<
	string,
	{ label: string; className: string }
> = {
	paid: { label: "Pago", className: "bg-primary/10 text-primary" },
	pending: { label: "Pendente", className: "bg-chart-2/10 text-chart-2" },
	overdue: {
		label: "Atrasado",
		className: "bg-destructive/10 text-destructive",
	},
};

// Framer Motion animation variants (reusable)
export const fadeUp: Variants = {
	hidden: { opacity: 0, y: 24 },
	visible: (i: number = 0) => ({
		opacity: 1,
		y: 0,
		transition: { delay: i * 0.1, duration: 0.6, ease: [0.22, 1, 0.36, 1] },
	}),
};

export const staggerContainer: Variants = {
	hidden: { opacity: 0 },
	visible: {
		opacity: 1,
		transition: { staggerChildren: 0.08, delayChildren: 0.1 },
	},
};

export const scaleIn: Variants = {
	hidden: { opacity: 0, scale: 0.95 },
	visible: {
		opacity: 1,
		scale: 1,
		transition: { duration: 0.4, ease: [0.22, 1, 0.36, 1] },
	},
};

// Transition presets
export const TRANSITION = {
	micro: { duration: 0.1, ease: "easeOut" },
	standard: { duration: 0.15, ease: "easeOut" },
	emphasis: { duration: 0.2, ease: [0.22, 1, 0.36, 1] },
} as const;
