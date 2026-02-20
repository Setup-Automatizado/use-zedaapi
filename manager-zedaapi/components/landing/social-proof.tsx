"use client";

import { useEffect, useRef, useState } from "react";
import { motion, useInView } from "framer-motion";

interface StatItemProps {
	value: number;
	suffix: string;
	label: string;
	description: string;
	delay: number;
}

function useCountUp(target: number, inView: boolean, duration = 2000) {
	const [count, setCount] = useState(0);

	useEffect(() => {
		if (!inView) return;

		let start = 0;
		const startTime = performance.now();

		function step(now: number) {
			const elapsed = now - startTime;
			const progress = Math.min(elapsed / duration, 1);
			const eased = 1 - Math.pow(1 - progress, 3);
			const current = Math.floor(eased * target);

			if (current !== start) {
				start = current;
				setCount(current);
			}

			if (progress < 1) {
				requestAnimationFrame(step);
			} else {
				setCount(target);
			}
		}

		requestAnimationFrame(step);
	}, [target, inView, duration]);

	return count;
}

function StatItem({ value, suffix, label, description, delay }: StatItemProps) {
	const ref = useRef<HTMLDivElement>(null);
	const isInView = useInView(ref, { once: true, margin: "-40px" });
	const count = useCountUp(value, isInView);

	return (
		<motion.div
			ref={ref}
			initial={{ opacity: 0, y: 20 }}
			animate={isInView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
			transition={{
				delay,
				duration: 0.5,
				ease: [0.22, 1, 0.36, 1],
			}}
			className="flex flex-col items-center gap-1.5 px-6 py-6"
		>
			<span className="tabular-nums text-4xl font-bold tracking-tight text-foreground sm:text-5xl">
				{count.toLocaleString("pt-BR")}
				<span className="text-primary">{suffix}</span>
			</span>
			<span className="text-sm font-semibold text-foreground">
				{label}
			</span>
			<span className="text-xs text-muted-foreground">{description}</span>
		</motion.div>
	);
}

const stats = [
	{
		value: 500,
		suffix: "+",
		label: "Empresas em Produção",
		description: "De startups a grandes operações",
	},
	{
		value: 10,
		suffix: "M+",
		label: "Mensagens Entregues por Mês",
		description: "Com latência média de 47ms",
	},
	{
		value: 99,
		suffix: ".9%",
		label: "Uptime Comprovado",
		description: "SLA por contrato, p99 < 200ms",
	},
	{
		value: 5000,
		suffix: "+",
		label: "Instâncias em Produção",
		description: "Operando agora, neste momento",
	},
] as const;

export function SocialProof() {
	return (
		<section className="relative border-y border-border/50 bg-muted/20">
			<h2 className="sr-only">Números da plataforma</h2>

			{/* Subtle gradient overlay */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute inset-0 bg-gradient-to-r from-transparent via-primary/[0.02] to-transparent"
			/>

			<div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				<div className="grid grid-cols-2 divide-x divide-border/50 lg:grid-cols-4">
					{stats.map((stat, i) => (
						<StatItem
							key={stat.label}
							value={stat.value}
							suffix={stat.suffix}
							label={stat.label}
							description={stat.description}
							delay={i * 0.1}
						/>
					))}
				</div>
			</div>
		</section>
	);
}
