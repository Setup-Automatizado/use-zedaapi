"use client";

import { motion, type Variants } from "framer-motion";
import {
	Code2,
	Webhook,
	Infinity,
	Smartphone,
	ImageIcon,
	MousePointerClick,
	Users,
	BarChart3,
	ShieldCheck,
	Lock,
	BookOpen,
	Headset,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

interface Feature {
	icon: LucideIcon;
	title: string;
	description: string;
	highlight?: boolean;
}

const features: Feature[] = [
	{
		icon: Code2,
		title: "API REST com Swagger UI",
		description:
			"Endpoints documentados com OpenAPI 3.0 e Swagger interativo. Um curl e você envia a primeira mensagem — sem SDK obrigatório.",
		highlight: true,
	},
	{
		icon: Webhook,
		title: "Webhooks com 47ms (p50)",
		description:
			"Cada mensagem, leitura e status de entrega dispara um webhook ao seu servidor. Latência p50 de 47ms, p99 abaixo de 200ms.",
		highlight: true,
	},
	{
		icon: Infinity,
		title: "Preço Fixo por Faixa",
		description:
			"Zero cobrança por mensagem, sem teto por instância. Pague pela faixa de instâncias e envie sem limite — até 10M mensagens/mês comprovados.",
	},
	{
		icon: Smartphone,
		title: "Multi-instância Independente",
		description:
			"Conecte múltiplas instâncias WhatsApp e controle todas pelo mesmo painel. Cada número opera com fila FIFO isolada e QR Code próprio.",
		highlight: true,
	},
	{
		icon: ImageIcon,
		title: "Mídia Assíncrona com S3",
		description:
			"Imagens, vídeos, áudios, PDFs e stickers processados de forma assíncrona. Fast path de 5s com fallback para fila — nunca bloqueia eventos.",
	},
	{
		icon: MousePointerClick,
		title: "Botões e Listas Interativas",
		description:
			"Botões de resposta rápida, listas de seleção e menus interativos via API. Aumente conversão sem depender de templates da Meta.",
	},
	{
		icon: Users,
		title: "Grupos via API Completa",
		description:
			"Crie grupos, adicione participantes, defina admins e envie em massa — tudo automatizado com endpoints REST dedicados.",
	},
	{
		icon: BarChart3,
		title: "Métricas com Prometheus",
		description:
			"Counters e histogramas exportados para Prometheus. Dashboard de envio, entrega, leitura e saúde de cada instância em tempo real.",
	},
	{
		icon: ShieldCheck,
		title: "SLA de 99.9% por Contrato",
		description:
			"Infraestrutura redundante com monitoramento 24/7, failover automático e SLA formal. Downtime máximo de 43min/mês nos planos Business+.",
	},
	{
		icon: Lock,
		title: "TLS 1.3 + AES-256-GCM",
		description:
			"Criptografia em trânsito e em repouso, tokens com hash criptográfico, rate limiting por instância. Conformidade total com a LGPD.",
	},
	{
		icon: BookOpen,
		title: "Docs com Exemplos em 4 Linguagens",
		description:
			"Guias passo a passo, exemplos em Node.js, Python, PHP e Go, Postman collection pronta e referência Swagger interativa.",
	},
	{
		icon: Headset,
		title: "Suporte Técnico em Português",
		description:
			"Equipe técnica brasileira por e-mail e WhatsApp. Planos Business+ incluem suporte prioritário com SLA de resposta de 4h.",
	},
];

const containerVariants: Variants = {
	hidden: {},
	visible: {
		transition: {
			staggerChildren: 0.05,
		},
	},
};

const cardVariants: Variants = {
	hidden: { opacity: 0, y: 24 },
	visible: {
		opacity: 1,
		y: 0,
		transition: { duration: 0.5, ease: [0.22, 1, 0.36, 1] },
	},
};

export function Features() {
	return (
		<section id="recursos" className="relative py-20 sm:py-28">
			{/* Subtle background accent */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute left-1/2 top-0 h-px w-2/3 -translate-x-1/2 bg-gradient-to-r from-transparent via-border to-transparent"
			/>

			<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				{/* Section header */}
				<div className="mx-auto max-w-2xl text-center">
					<motion.p
						initial={{ opacity: 0, y: 16 }}
						whileInView={{ opacity: 1, y: 0 }}
						viewport={{ once: true, margin: "-80px" }}
						transition={{ duration: 0.5 }}
						className="text-sm font-semibold uppercase tracking-widest text-primary"
					>
						Recursos
					</motion.p>
					<motion.h2
						initial={{ opacity: 0, y: 16 }}
						whileInView={{ opacity: 1, y: 0 }}
						viewport={{ once: true, margin: "-80px" }}
						transition={{ duration: 0.5, delay: 0.05 }}
						className="mt-3 text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl"
					>
						Tudo que sua integração WhatsApp{" "}
						<span className="text-primary">precisa</span>
					</motion.h2>
					<motion.p
						initial={{ opacity: 0, y: 16 }}
						whileInView={{ opacity: 1, y: 0 }}
						viewport={{ once: true, margin: "-80px" }}
						transition={{ duration: 0.5, delay: 0.1 }}
						className="mt-4 text-base text-muted-foreground sm:text-lg leading-relaxed"
					>
						API REST documentada, webhooks instantâneos, mídia
						assíncrona e preço fixo sem cobrança por mensagem.
					</motion.p>
				</div>

				{/* Features grid */}
				<motion.div
					variants={containerVariants}
					initial="hidden"
					whileInView="visible"
					viewport={{ once: true, margin: "-60px" }}
					className="mt-14 grid grid-cols-1 gap-4 sm:mt-20 sm:grid-cols-2 lg:grid-cols-3"
				>
					{features.map((feature) => (
						<motion.div
							key={feature.title}
							variants={cardVariants}
							className={`group relative rounded-2xl border p-6 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-primary/[0.03] ${
								feature.highlight
									? "border-primary/15 bg-primary/[0.02]"
									: "border-border bg-card"
							}`}
						>
							<div
								className={`flex size-11 items-center justify-center rounded-xl transition-colors duration-300 ${
									feature.highlight
										? "bg-primary/10 group-hover:bg-primary/15"
										: "bg-muted group-hover:bg-primary/10"
								}`}
							>
								<feature.icon
									className={`size-5 transition-colors duration-300 ${
										feature.highlight
											? "text-primary"
											: "text-muted-foreground group-hover:text-primary"
									}`}
								/>
							</div>
							<h3 className="mt-4 text-sm font-semibold text-foreground">
								{feature.title}
							</h3>
							<p className="mt-2 text-sm leading-relaxed text-muted-foreground">
								{feature.description}
							</p>
						</motion.div>
					))}
				</motion.div>
			</div>
		</section>
	);
}
