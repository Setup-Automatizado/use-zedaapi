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
		title: "Integre em 5 Minutos",
		description:
			"Um único curl e você já está enviando mensagens. API REST documentada, compatível com qualquer linguagem — sem SDKs obrigatórios.",
		highlight: true,
	},
	{
		icon: Webhook,
		title: "Saiba Tudo em Tempo Real",
		description:
			"Receba webhooks instantâneos a cada mensagem, leitura e status de entrega. Latência média de 47ms do evento ao seu servidor.",
		highlight: true,
	},
	{
		icon: Infinity,
		title: "Escale Sem Limites",
		description:
			"Zero cobrança por mensagem, sem teto por instância. Empresas enviam mais de 10 milhões de mensagens por mês conosco — sem custo extra.",
	},
	{
		icon: Smartphone,
		title: "Gerencie Tudo de Um Só Lugar",
		description:
			"Conecte múltiplas instâncias WhatsApp e controle todas pelo mesmo painel. Cada número opera de forma 100% independente.",
		highlight: true,
	},
	{
		icon: ImageIcon,
		title: "Envie Qualquer Formato",
		description:
			"Imagens, vídeos, áudios, PDFs, stickers e localização — tudo via API. Processamento assíncrono para nunca travar sua fila.",
	},
	{
		icon: MousePointerClick,
		title: "Converta com Interatividade",
		description:
			"Botões de resposta rápida, listas de seleção e menus interativos que aumentam suas taxas de conversão em até 3x.",
	},
	{
		icon: Users,
		title: "Domine Grupos e Comunidades",
		description:
			"Crie grupos, adicione participantes, defina admins e envie mensagens em massa — tudo automatizado pela API.",
	},
	{
		icon: BarChart3,
		title: "Decisões com Dados Reais",
		description:
			"Dashboard completo com métricas de envio, entrega, leitura e saúde das instâncias. Monitore tudo em tempo real.",
	},
	{
		icon: ShieldCheck,
		title: "99.9% Online, Sempre",
		description:
			"Infraestrutura redundante com monitoramento 24/7 e SLA garantido por contrato. Sua operação nunca para.",
	},
	{
		icon: Lock,
		title: "Segurança Nível Enterprise",
		description:
			"TLS 1.3 em trânsito, AES-256 em repouso, tokens rotativos e rate limiting. 100% em conformidade com a LGPD.",
	},
	{
		icon: BookOpen,
		title: "Documente-se em Minutos",
		description:
			"Guias passo a passo, exemplos em 8 linguagens, Postman collection pronta e referência interativa da API.",
	},
	{
		icon: Headset,
		title: "Suporte que Fala Sua Língua",
		description:
			"Time técnico 100% brasileiro, disponível por e-mail e WhatsApp. Planos avançados têm atendimento dedicado e prioritário.",
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
						Resultados que sua concorrência{" "}
						<span className="text-primary">não consegue</span>{" "}
						entregar
					</motion.h2>
					<motion.p
						initial={{ opacity: 0, y: 16 }}
						whileInView={{ opacity: 1, y: 0 }}
						viewport={{ once: true, margin: "-80px" }}
						transition={{ duration: 0.5, delay: 0.1 }}
						className="mt-4 text-base text-muted-foreground sm:text-lg leading-relaxed"
					>
						Enquanto outras APIs travam na escalabilidade, a ZedaAPI
						processa mais de 10 milhões de mensagens por mês com
						latência média de 47ms.
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
