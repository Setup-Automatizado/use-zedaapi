"use client";

import { useRef } from "react";
import Link from "next/link";
import { motion, useInView } from "framer-motion";
import type { Variants } from "motion-dom";
import {
	ArrowRightIcon,
	ZapIcon,
	ShieldCheckIcon,
	ClockIcon,
	CreditCardIcon,
	HeadphonesIcon,
	ServerIcon,
} from "lucide-react";
import { Button } from "@/components/ui/button";

// ── Animations ──────────────────────────

const fadeUp: Variants = {
	hidden: { opacity: 0, y: 24 },
	visible: {
		opacity: 1,
		y: 0,
		transition: { duration: 0.6, ease: [0.22, 1, 0.36, 1] as const },
	},
};

const staggerContainer: Variants = {
	hidden: {},
	visible: { transition: { staggerChildren: 0.06 } },
};

const pillIn: Variants = {
	hidden: { opacity: 0, scale: 0.9, y: 8 },
	visible: {
		opacity: 1,
		scale: 1,
		y: 0,
		transition: { duration: 0.4, ease: [0.22, 1, 0.36, 1] as const },
	},
};

// ── Trust items ──────────────────────────

const trustItems = [
	{
		icon: ShieldCheckIcon,
		label: "Conforme a LGPD",
		color: "text-emerald-400",
		bg: "bg-emerald-500/10",
	},
	{
		icon: ServerIcon,
		label: "Infra brasileira",
		color: "text-sky-400",
		bg: "bg-sky-500/10",
	},
	{
		icon: CreditCardIcon,
		label: "Sem cartão",
		color: "text-amber-400",
		bg: "bg-amber-500/10",
	},
	{
		icon: ClockIcon,
		label: "Setup em 5 min",
		color: "text-violet-400",
		bg: "bg-violet-500/10",
	},
	{
		icon: HeadphonesIcon,
		label: "Suporte PT-BR",
		color: "text-rose-400",
		bg: "bg-rose-500/10",
	},
	{
		icon: ZapIcon,
		label: "47ms latência",
		color: "text-primary",
		bg: "bg-primary/10",
	},
] as const;

// ── Component ──────────────────────────

export function CTA() {
	const ref = useRef<HTMLDivElement>(null);
	const isInView = useInView(ref, { once: true, margin: "-80px" });

	return (
		<section id="contato" className="relative py-20 sm:py-28" ref={ref}>
			{/* Top divider gradient */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute left-1/2 top-0 h-px w-2/3 -translate-x-1/2 bg-gradient-to-r from-transparent via-border to-transparent"
			/>

			<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				<motion.div
					initial="hidden"
					animate={isInView ? "visible" : "hidden"}
					className="relative overflow-hidden rounded-[2rem] border border-border/40"
				>
					{/* === Multi-layer background === */}

					{/* Base gradient */}
					<div
						aria-hidden="true"
						className="pointer-events-none absolute inset-0 bg-gradient-to-b from-card via-card/95 to-muted/30"
					/>

					{/* Grid pattern */}
					<div
						aria-hidden="true"
						className="pointer-events-none absolute inset-0 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] bg-[size:3rem_3rem] opacity-[0.08] [mask-image:radial-gradient(ellipse_80%_60%_at_50%_40%,black_20%,transparent_100%)]"
					/>

					{/* Primary orb - top center */}
					<div
						aria-hidden="true"
						className="pointer-events-none absolute left-1/2 -top-32 h-[500px] w-[700px] -translate-x-1/2 rounded-full bg-primary/10 blur-[160px]"
					/>

					{/* Secondary orb - bottom right */}
					<div
						aria-hidden="true"
						className="pointer-events-none absolute -bottom-20 -right-20 h-[350px] w-[450px] rounded-full bg-primary/6 blur-[120px]"
					/>

					{/* Tertiary orb - bottom left */}
					<div
						aria-hidden="true"
						className="pointer-events-none absolute -bottom-16 -left-16 h-[300px] w-[400px] rounded-full bg-violet-500/4 blur-[100px]"
					/>

					{/* Radial highlight from top */}
					<div
						aria-hidden="true"
						className="pointer-events-none absolute inset-x-0 top-0 h-64 bg-gradient-to-b from-primary/[0.04] to-transparent"
					/>

					{/* Inner border glow */}
					<div
						aria-hidden="true"
						className="pointer-events-none absolute inset-0 rounded-[2rem] ring-1 ring-inset ring-white/[0.04]"
					/>

					{/* === Content === */}
					<div className="relative px-6 py-20 sm:px-12 sm:py-28 md:px-20 lg:py-32">
						<div className="mx-auto max-w-3xl text-center">
							{/* Animated badge */}
							<motion.div
								variants={fadeUp}
								className="mb-8 flex justify-center"
							>
								<div className="flex items-center gap-2.5 rounded-full border border-primary/20 bg-primary/5 px-4 py-2">
									<span className="relative flex size-2 shrink-0">
										<span className="absolute inline-flex size-full animate-ping rounded-full bg-primary/60" />
										<span className="relative inline-flex size-2 rounded-full bg-primary" />
									</span>
									<span className="text-sm font-medium text-primary whitespace-nowrap">
										+500 empresas confiam
									</span>
								</div>
							</motion.div>

							{/* Headline */}
							<motion.h2
								variants={fadeUp}
								className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl lg:text-6xl"
							>
								Comece a enviar
								<br className="hidden sm:block" />
								em 5 minutos.{" "}
								<span className="relative">
									<span className="text-primary">
										Teste grátis.
									</span>
									{/* Underline accent */}
									<svg
										aria-hidden="true"
										viewBox="0 0 286 8"
										fill="none"
										className="absolute -bottom-1 left-0 w-full"
									>
										<path
											d="M2 5.5C50 2.5 100 1.5 143 3.5C186 5.5 236 4.5 284 2"
											stroke="currentColor"
											strokeWidth="3"
											strokeLinecap="round"
											className="text-primary/30"
										/>
									</svg>
								</span>
							</motion.h2>

							{/* Subtitle */}
							<motion.p
								variants={fadeUp}
								className="mx-auto mt-6 max-w-xl text-base text-muted-foreground sm:text-lg leading-relaxed"
							>
								Crie sua conta, conecte pelo QR Code e envie a
								primeira mensagem via API. Sem aprovação manual,
								sem cartão de crédito.
							</motion.p>

							{/* CTAs */}
							<motion.div
								variants={fadeUp}
								className="mt-10 flex flex-col items-center justify-center gap-3 sm:flex-row sm:gap-4"
							>
								<Button
									size="lg"
									asChild
									className="group/btn relative h-13 px-8 text-base font-semibold shadow-xl shadow-primary/25 hover:shadow-2xl hover:shadow-primary/30 transition-all duration-300"
								>
									<Link href="/cadastro">
										<span className="relative z-10 flex items-center gap-2">
											Criar Conta Grátis
											<ArrowRightIcon className="size-4 transition-transform duration-200 group-hover/btn:translate-x-0.5" />
										</span>
									</Link>
								</Button>
								<Button
									variant="outline"
									size="lg"
									asChild
									className="h-13 px-8 text-base font-medium backdrop-blur-sm"
								>
									<Link href="/contato">
										Falar com Especialista
									</Link>
								</Button>
							</motion.div>

							{/* Free trial callout */}
							<motion.p
								variants={fadeUp}
								className="mt-4 text-xs text-muted-foreground/60"
							>
								7 dias grátis em todos os planos. Cancele quando
								quiser, sem multa.
							</motion.p>
						</div>

						{/* Trust pills */}
						<motion.div
							variants={staggerContainer}
							className="mx-auto mt-14 flex max-w-2xl flex-wrap items-center justify-center gap-2.5"
						>
							{trustItems.map((item) => {
								const Icon = item.icon;
								return (
									<motion.div
										key={item.label}
										variants={pillIn}
										className="flex items-center gap-2 rounded-full border border-border/50 bg-background/60 px-3.5 py-2 backdrop-blur-sm transition-all duration-300 hover:border-border hover:bg-background/80 hover:shadow-sm"
									>
										<div
											className={`flex size-5 items-center justify-center rounded-full ${item.bg}`}
										>
											<Icon
												className={`size-2.5 ${item.color}`}
											/>
										</div>
										<span className="text-xs font-medium text-muted-foreground">
											{item.label}
										</span>
									</motion.div>
								);
							})}
						</motion.div>
					</div>
				</motion.div>
			</div>
		</section>
	);
}
