"use client";

import type { ComponentType, SVGProps } from "react";
import { motion, type Variants } from "framer-motion";
import { Globe, Webhook } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import N8nIcon from "@/components/icons/n8n";
import MakeIcon from "@/components/icons/make";
import ZapierIcon from "@/components/icons/zapier";
import NodejsIcon from "@/components/icons/nodejs";
import PythonIcon from "@/components/icons/python";
import PhpIcon from "@/components/icons/php";
import WebhookIcon from "@/components/icons/webhook";

interface Integration {
	icon: LucideIcon | ComponentType<SVGProps<SVGSVGElement>>;
	name: string;
	description: string;
	tag?: string;
}

const integrations: Integration[] = [
	{
		icon: N8nIcon,
		name: "n8n",
		description:
			"Open-source, sem custos adicionais de plataforma. Nodes nativos prontos para usar — arraste, conecte e automatize em minutos.",
		tag: "Popular",
	},
	{
		icon: MakeIcon,
		name: "Make (Integromat)",
		description:
			"Centenas de cenários prontos e um builder visual poderoso. Crie automações complexas sem escrever uma linha de código.",
	},
	{
		icon: ZapierIcon,
		name: "Zapier",
		description:
			"Conecte WhatsApp a 6.000+ apps em minutos. Workflows automatizados que funcionam enquanto você foca no que importa.",
	},
	{
		icon: NodejsIcon,
		name: "Node.js / TypeScript",
		description:
			"Fetch, Axios ou sua lib favorita. API REST com respostas JSON tipáveis — integre em minutos com qualquer projeto Node.",
	},
	{
		icon: PythonIcon,
		name: "Python",
		description:
			"Requests ou httpx e pronto. Ideal para bots inteligentes, automações e integrações com IA via API REST.",
	},
	{
		icon: PhpIcon,
		name: "PHP / Laravel",
		description:
			"Guzzle ou Http facade do Laravel. Integre com o sistema de filas e deploy com o stack que seu time já domina.",
	},
	{
		icon: Globe,
		name: "Qualquer Linguagem",
		description:
			"REST padrão com HTTP/JSON — sem vendor lock-in. Go, Java, Ruby, C# ou qualquer linguagem que fale HTTP.",
	},
	{
		icon: WebhookIcon,
		name: "Webhooks Universais",
		description:
			"Eventos em tempo real para qualquer endpoint. Payloads JSON estruturados com retry automático e entrega garantida.",
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
	hidden: { opacity: 0, y: 20 },
	visible: {
		opacity: 1,
		y: 0,
		transition: { duration: 0.5, ease: [0.22, 1, 0.36, 1] },
	},
};

export function Integrations() {
	return (
		<section
			id="integracoes"
			className="relative border-t border-border/50 bg-muted/20 py-20 sm:py-28"
		>
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
						Integrações
					</motion.p>
					<motion.h2
						initial={{ opacity: 0, y: 16 }}
						whileInView={{ opacity: 1, y: 0 }}
						viewport={{ once: true, margin: "-80px" }}
						transition={{ duration: 0.5, delay: 0.05 }}
						className="mt-3 text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl"
					>
						Conecte com as ferramentas{" "}
						<span className="text-primary">que você já usa</span>
					</motion.h2>
					<motion.p
						initial={{ opacity: 0, y: 16 }}
						whileInView={{ opacity: 1, y: 0 }}
						viewport={{ once: true, margin: "-80px" }}
						transition={{ duration: 0.5, delay: 0.1 }}
						className="mt-4 text-base text-muted-foreground sm:text-lg leading-relaxed"
					>
						SDKs oficiais, nodes para no-code e API REST completa.
						Sua equipe integra em minutos — não em semanas.
					</motion.p>
				</div>

				{/* Integrations grid */}
				<motion.div
					variants={containerVariants}
					initial="hidden"
					whileInView="visible"
					viewport={{ once: true, margin: "-60px" }}
					className="mt-14 grid grid-cols-1 gap-4 sm:mt-20 sm:grid-cols-2 lg:grid-cols-4"
				>
					{integrations.map((integration) => {
						const Icon = integration.icon;
						return (
							<motion.div
								key={integration.name}
								variants={cardVariants}
								className="group relative flex flex-col rounded-2xl border border-border bg-card p-6 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-primary/[0.03] hover:border-primary/30"
							>
								{integration.tag && (
									<span className="absolute right-4 top-4 rounded-4xl bg-primary/10 px-2 py-0.5 text-[10px] font-semibold text-primary uppercase tracking-wider">
										{integration.tag}
									</span>
								)}
								<div className="flex size-12 items-center justify-center rounded-xl bg-muted transition-colors duration-300 group-hover:bg-primary/10">
									<Icon className="size-6 text-muted-foreground transition-colors duration-300 group-hover:text-primary" />
								</div>
								<h3 className="mt-4 text-sm font-semibold text-foreground">
									{integration.name}
								</h3>
								<p className="mt-2 text-sm leading-relaxed text-muted-foreground">
									{integration.description}
								</p>
							</motion.div>
						);
					})}
				</motion.div>
			</div>
		</section>
	);
}
