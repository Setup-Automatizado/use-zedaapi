"use client";

import { useRef } from "react";
import { motion, useInView } from "framer-motion";
import type { Variants } from "motion-dom";
import {
	UserPlusIcon,
	SmartphoneIcon,
	Code2Icon,
	ArrowRightIcon,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

interface Step {
	number: string;
	icon: LucideIcon;
	title: string;
	description: string;
	accent: string;
}

const steps: Step[] = [
	{
		number: "01",
		icon: UserPlusIcon,
		title: "Crie sua conta grátis",
		description:
			"Sem burocracia, sem formulários longos e sem cartão de crédito. Preencha nome, e-mail e senha — pronto. Acesso imediato ao painel completo com 7 dias grátis em qualquer plano.",
		accent: "from-emerald-500/20 to-emerald-500/0",
	},
	{
		number: "02",
		icon: SmartphoneIcon,
		title: "Conecte seu WhatsApp — sem instalar nada",
		description:
			"Escaneie um QR Code pelo painel e pronto, instância online. Funciona com qualquer celular, não exige app extra nem conhecimento técnico. É tão simples quanto abrir o WhatsApp Web.",
		accent: "from-sky-500/20 to-sky-500/0",
	},
	{
		number: "03",
		icon: Code2Icon,
		title: "Integre e envie a primeira mensagem",
		description:
			"Documentação em 4 linguagens (Node.js, Python, PHP e Go), coleção Postman pronta para importar. Copie, cole, envie. Sua primeira mensagem automática em menos de 2 minutos.",
		accent: "from-violet-500/20 to-violet-500/0",
	},
];

const fadeUp: Variants = {
	hidden: { opacity: 0, y: 24 },
	visible: (i: number) => ({
		opacity: 1,
		y: 0,
		transition: {
			delay: i * 0.15,
			duration: 0.6,
			ease: [0.22, 1, 0.36, 1],
		},
	}),
};

function StepCard({ step, index }: { step: Step; index: number }) {
	const Icon = step.icon;

	return (
		<motion.div
			custom={index + 1}
			variants={fadeUp}
			className="group relative flex flex-col items-center text-center"
		>
			{/* Card */}
			<div className="relative mb-6 w-full rounded-2xl border border-border bg-card p-8 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-primary/[0.03]">
				{/* Gradient accent */}
				<div
					aria-hidden="true"
					className={`pointer-events-none absolute inset-x-0 top-0 h-32 rounded-t-2xl bg-gradient-to-b ${step.accent} opacity-50`}
				/>

				{/* Number + Icon */}
				<div className="relative flex items-center justify-between">
					<span className="text-5xl font-bold text-border tracking-tighter">
						{step.number}
					</span>
					<div className="flex size-12 items-center justify-center rounded-xl bg-primary/10">
						<Icon className="size-6 text-primary" />
					</div>
				</div>

				{/* Content */}
				<div className="relative mt-6 text-left">
					<h3 className="text-lg font-semibold text-foreground">
						{step.title}
					</h3>
					<p className="mt-2 text-sm leading-relaxed text-muted-foreground">
						{step.description}
					</p>
				</div>
			</div>

			{/* Connector arrow (not on last item, desktop only) */}
			{index < steps.length - 1 && (
				<div
					aria-hidden="true"
					className="pointer-events-none absolute -right-6 top-1/2 hidden -translate-y-1/2 text-border lg:block"
				>
					<ArrowRightIcon className="size-5" />
				</div>
			)}
		</motion.div>
	);
}

export function HowItWorks() {
	const ref = useRef<HTMLDivElement>(null);
	const isInView = useInView(ref, { once: true, margin: "-80px" });

	return (
		<section
			id="como-funciona"
			className="relative border-t border-border/50 bg-muted/20 py-20 sm:py-28"
		>
			<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				{/* Section header */}
				<motion.div
					ref={ref}
					initial="hidden"
					animate={isInView ? "visible" : "hidden"}
					className="flex flex-col items-center text-center"
				>
					<motion.p
						custom={0}
						variants={fadeUp}
						className="text-sm font-semibold uppercase tracking-widest text-primary"
					>
						Como funciona
					</motion.p>
					<motion.h2
						custom={0.5}
						variants={fadeUp}
						className="mt-3 text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl"
					>
						Do cadastro à primeira mensagem automática{" "}
						<br className="hidden sm:block" />
						em menos de{" "}
						<span className="text-primary">5 minutos</span>
					</motion.h2>
					<motion.p
						custom={0.7}
						variants={fadeUp}
						className="mt-4 max-w-lg text-base text-muted-foreground sm:text-lg"
					>
						Três passos — nenhum obstáculo. Sem aprovação manual,
						sem espera por suporte.
					</motion.p>
				</motion.div>

				{/* Steps grid */}
				<motion.div
					initial="hidden"
					animate={isInView ? "visible" : "hidden"}
					className="relative mt-14 grid grid-cols-1 gap-8 sm:mt-20 md:grid-cols-3 md:gap-12"
				>
					{steps.map((step, i) => (
						<StepCard key={step.number} step={step} index={i} />
					))}
				</motion.div>
			</div>
		</section>
	);
}
