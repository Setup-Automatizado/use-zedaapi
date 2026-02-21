"use client";

import { useRef } from "react";
import { motion, useInView } from "framer-motion";
import type { Variants } from "motion-dom";
import {
	Code2,
	Zap,
	BadgeDollarSign,
	Smartphone,
	ImageIcon,
	MousePointerClick,
	Users,
	BarChart3,
	ShieldCheck,
	BookOpen,
	Headset,
	Lock,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

// ── Primary features (3 large cards) ──────────────────────────

interface PrimaryFeature {
	icon: LucideIcon;
	title: string;
	description: string;
	detail: string;
}

const primaryFeatures: PrimaryFeature[] = [
	{
		icon: Code2,
		title: "API REST completa",
		description:
			"Um curl e você envia a primeira mensagem. Documentação interativa com Swagger, exemplos prontos em 4 linguagens e Postman collection inclusa.",
		detail: "Texto, imagem, áudio, vídeo, documentos, localização, contatos, botões e listas",
	},
	{
		icon: Zap,
		title: "Webhooks instantâneos",
		description:
			"Cada mensagem recebida, leitura e status de entrega chega ao seu servidor em tempo real. Latência mediana de 47ms — você reage antes do usuário piscar.",
		detail: "Mensagens, status de entrega, leitura, presença e eventos de grupo",
	},
	{
		icon: BadgeDollarSign,
		title: "Preço fixo, sem surpresa",
		description:
			"Zero cobrança por mensagem enviada ou recebida. Pague pela faixa de instâncias e envie sem limite — até 10 milhões de mensagens por mês comprovados.",
		detail: "Sem taxa por mensagem, sem teto de envio, sem cobrança excedente",
	},
];

// ── Secondary features (compact grid) ──────────────────────────

interface SecondaryFeature {
	icon: LucideIcon;
	title: string;
	description: string;
}

const secondaryFeatures: SecondaryFeature[] = [
	{
		icon: Smartphone,
		title: "Multi-instância",
		description:
			"Conecte múltiplos números WhatsApp e controle todos pelo mesmo painel. Cada instância opera de forma independente com QR Code próprio.",
	},
	{
		icon: ImageIcon,
		title: "Mídia sem bloqueio",
		description:
			"Imagens, vídeos, áudios e PDFs processados em background. A entrega do evento nunca espera o upload terminar.",
	},
	{
		icon: MousePointerClick,
		title: "Botões e listas",
		description:
			"Botões de resposta rápida, listas de seleção e menus interativos direto pela API. Aumente conversão sem templates da Meta.",
	},
	{
		icon: Users,
		title: "Grupos via API",
		description:
			"Crie grupos, adicione participantes, defina admins e envie em massa — tudo automatizado com endpoints dedicados.",
	},
	{
		icon: BarChart3,
		title: "Dashboard em tempo real",
		description:
			"Acompanhe envios, entregas, leituras e saúde de cada instância. Métricas exportáveis para seu sistema de monitoramento.",
	},
	{
		icon: ShieldCheck,
		title: "99.9% de uptime",
		description:
			"Infraestrutura redundante com failover automático e monitoramento 24/7. Downtime máximo de 43 min/mês por contrato.",
	},
	{
		icon: Lock,
		title: "Segurança completa",
		description:
			"Criptografia em trânsito e em repouso, tokens com hash criptográfico, rate limiting por instância. Conformidade com a LGPD.",
	},
	{
		icon: BookOpen,
		title: "Docs em 4 linguagens",
		description:
			"Guias passo a passo com exemplos em Node.js, Python, PHP e Go. Postman collection pronta e referência Swagger interativa.",
	},
	{
		icon: Headset,
		title: "Suporte em português",
		description:
			"Equipe técnica brasileira por e-mail e WhatsApp. Planos Business+ com suporte prioritário e SLA de resposta de 4h.",
	},
];

// ── Animations ──────────────────────────

const containerVariants: Variants = {
	hidden: {},
	visible: { transition: { staggerChildren: 0.08 } },
};

const cardVariants: Variants = {
	hidden: { opacity: 0, y: 20 },
	visible: {
		opacity: 1,
		y: 0,
		transition: { duration: 0.5, ease: [0.22, 1, 0.36, 1] },
	},
};

const fadeIn: Variants = {
	hidden: { opacity: 0, y: 16 },
	visible: {
		opacity: 1,
		y: 0,
		transition: { duration: 0.5, ease: [0.22, 1, 0.36, 1] },
	},
};

// ── Components ──────────────────────────

function PrimaryCard({ feature }: { feature: PrimaryFeature }) {
	return (
		<motion.div
			variants={cardVariants}
			className="group relative flex flex-col rounded-xl border border-primary/10 bg-primary/[0.02] p-6 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-primary/[0.04]"
		>
			{/* Gradient accent top */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-primary/30 to-transparent"
			/>

			<div className="flex size-11 items-center justify-center rounded-xl bg-primary/10 transition-colors duration-300 group-hover:bg-primary/15">
				<feature.icon className="size-5 text-primary" />
			</div>

			<h3 className="mt-4 text-base font-semibold text-foreground">
				{feature.title}
			</h3>

			<p className="mt-2 flex-1 text-sm leading-relaxed text-muted-foreground">
				{feature.description}
			</p>

			{/* Detail chip */}
			<div className="mt-4 rounded-lg border border-border/50 bg-muted/30 px-3 py-2">
				<p className="text-xs leading-relaxed text-muted-foreground">
					{feature.detail}
				</p>
			</div>
		</motion.div>
	);
}

function SecondaryCard({ feature }: { feature: SecondaryFeature }) {
	return (
		<motion.div
			variants={cardVariants}
			className="group flex gap-4 rounded-xl border border-border bg-card p-4 transition-all duration-300 hover:-translate-y-0.5 hover:border-primary/15 hover:shadow-md hover:shadow-primary/[0.03]"
		>
			<div className="flex size-9 flex-shrink-0 items-center justify-center rounded-lg bg-muted transition-colors duration-300 group-hover:bg-primary/10">
				<feature.icon className="size-4 text-muted-foreground transition-colors duration-300 group-hover:text-primary" />
			</div>
			<div className="min-w-0">
				<h3 className="text-sm font-semibold text-foreground">
					{feature.title}
				</h3>
				<p className="mt-1 text-[13px] leading-relaxed text-muted-foreground">
					{feature.description}
				</p>
			</div>
		</motion.div>
	);
}

// ── Main ──────────────────────────

export function Features() {
	const ref = useRef<HTMLDivElement>(null);
	const isInView = useInView(ref, { once: true, margin: "-80px" });

	return (
		<section id="recursos" className="relative py-20 sm:py-28" ref={ref}>
			{/* Top divider */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute left-1/2 top-0 h-px w-2/3 -translate-x-1/2 bg-gradient-to-r from-transparent via-border to-transparent"
			/>

			<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				{/* Section header */}
				<div className="mx-auto max-w-2xl text-center">
					<motion.p
						variants={fadeIn}
						initial="hidden"
						animate={isInView ? "visible" : "hidden"}
						className="text-sm font-semibold uppercase tracking-widest text-primary"
					>
						Recursos
					</motion.p>
					<motion.h2
						variants={fadeIn}
						initial="hidden"
						animate={isInView ? "visible" : "hidden"}
						transition={{ delay: 0.05 }}
						className="mt-3 text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl"
					>
						Tudo que sua integração{" "}
						<span className="text-primary">precisa</span>
					</motion.h2>
					<motion.p
						variants={fadeIn}
						initial="hidden"
						animate={isInView ? "visible" : "hidden"}
						transition={{ delay: 0.1 }}
						className="mt-4 text-base text-muted-foreground sm:text-lg leading-relaxed"
					>
						API documentada, webhooks em tempo real, envio de mídia
						e preço fixo — sem cobrança por mensagem.
					</motion.p>
				</div>

				{/* Primary features — 3 large cards */}
				<motion.div
					variants={containerVariants}
					initial="hidden"
					animate={isInView ? "visible" : "hidden"}
					className="mt-14 grid grid-cols-1 gap-4 sm:mt-16 md:grid-cols-3"
				>
					{primaryFeatures.map((feature) => (
						<PrimaryCard key={feature.title} feature={feature} />
					))}
				</motion.div>

				{/* Secondary features — compact grid */}
				<motion.div
					variants={containerVariants}
					initial="hidden"
					animate={isInView ? "visible" : "hidden"}
					className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3"
				>
					{secondaryFeatures.map((feature) => (
						<SecondaryCard key={feature.title} feature={feature} />
					))}
				</motion.div>
			</div>
		</section>
	);
}
