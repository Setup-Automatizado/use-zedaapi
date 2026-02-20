"use client";

import Link from "next/link";
import { motion } from "framer-motion";
import type { Variants } from "motion-dom";
import {
	ArrowRightIcon,
	CopyIcon,
	CheckIcon,
	SparklesIcon,
} from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

const fadeUp: Variants = {
	hidden: { opacity: 0, y: 24 },
	visible: (i: number) => ({
		opacity: 1,
		y: 0,
		transition: {
			delay: i * 0.1,
			duration: 0.6,
			ease: [0.22, 1, 0.36, 1],
		},
	}),
};

const CODE_EXAMPLE = `curl -X POST https://api.zedaapi.com/v1/send-text \\
  -H "Client-Token: seu_token_aqui" \\
  -H "Content-Type: application/json" \\
  -d '{
    "phone": "5521999999999",
    "message": "Olá! Seu pedido #1234 foi confirmado."
  }'`;

function CodeBlock() {
	const [copied, setCopied] = useState(false);

	function handleCopy() {
		navigator.clipboard.writeText(CODE_EXAMPLE);
		setCopied(true);
		setTimeout(() => setCopied(false), 2000);
	}

	return (
		<motion.div
			custom={5}
			variants={fadeUp}
			initial="hidden"
			animate="visible"
			className="relative w-full max-w-2xl overflow-hidden rounded-2xl border border-white/[0.08] bg-zinc-950 shadow-2xl shadow-primary/5"
		>
			{/* Window chrome */}
			<div className="flex items-center justify-between border-b border-white/[0.06] px-4 py-3">
				<div className="flex items-center gap-1.5">
					<div className="size-2.5 rounded-full bg-zinc-700" />
					<div className="size-2.5 rounded-full bg-zinc-700" />
					<div className="size-2.5 rounded-full bg-zinc-700" />
				</div>
				<span className="text-[11px] font-medium tracking-wide text-zinc-500 uppercase">
					Terminal
				</span>
				<button
					type="button"
					onClick={handleCopy}
					className="flex items-center gap-1.5 rounded-lg px-2.5 py-1 text-xs text-zinc-500 transition-all duration-150 hover:bg-white/[0.06] hover:text-zinc-300"
					aria-label="Copiar comando"
				>
					{copied ? (
						<CheckIcon className="size-3.5 text-emerald-400" />
					) : (
						<CopyIcon className="size-3.5" />
					)}
					{copied ? "Copiado!" : "Copiar"}
				</button>
			</div>

			{/* Code content */}
			<div className="p-5 sm:p-6">
				<pre className="overflow-x-auto text-[13px] leading-relaxed">
					<code>
						<span className="text-zinc-600">$ </span>
						<span className="text-emerald-400 font-medium">
							curl
						</span>
						<span className="text-zinc-300"> -X POST </span>
						<span className="text-sky-400">
							https://api.zedaapi.com/v1/send-text
						</span>
						<span className="text-zinc-600"> \</span>
						{"\n"}
						<span className="text-zinc-300">{"  "}-H </span>
						<span className="text-amber-300">
							&quot;Client-Token: seu_token_aqui&quot;
						</span>
						<span className="text-zinc-600"> \</span>
						{"\n"}
						<span className="text-zinc-300">{"  "}-H </span>
						<span className="text-amber-300">
							&quot;Content-Type: application/json&quot;
						</span>
						<span className="text-zinc-600"> \</span>
						{"\n"}
						<span className="text-zinc-300">{"  "}-d </span>
						<span className="text-amber-300">&apos;{"{"}</span>
						{"\n"}
						<span className="text-zinc-300">{"    "}</span>
						<span className="text-sky-300">&quot;phone&quot;</span>
						<span className="text-zinc-300">: </span>
						<span className="text-emerald-300">
							&quot;5521999999999&quot;
						</span>
						<span className="text-zinc-300">,</span>
						{"\n"}
						<span className="text-zinc-300">{"    "}</span>
						<span className="text-sky-300">
							&quot;message&quot;
						</span>
						<span className="text-zinc-300">: </span>
						<span className="text-emerald-300">
							&quot;Olá! Seu pedido #1234 foi confirmado.&quot;
						</span>
						{"\n"}
						<span className="text-amber-300">{"  }"}&apos;</span>
					</code>
				</pre>

				{/* Response */}
				<div className="mt-5 border-t border-white/[0.04] pt-5">
					<div className="mb-3 flex items-center gap-2">
						<div className="size-1.5 rounded-full bg-emerald-400 animate-pulse" />
						<p className="text-[11px] font-medium uppercase tracking-wider text-zinc-500">
							Resposta &mdash; 47ms
						</p>
					</div>
					<pre className="text-[13px] leading-relaxed">
						<code>
							<span className="text-zinc-600">{"{"}</span>
							{"\n"}
							<span className="text-zinc-300">{"  "}</span>
							<span className="text-sky-300">
								&quot;success&quot;
							</span>
							<span className="text-zinc-300">: </span>
							<span className="text-emerald-400 font-medium">
								true
							</span>
							<span className="text-zinc-300">,</span>
							{"\n"}
							<span className="text-zinc-300">{"  "}</span>
							<span className="text-sky-300">
								&quot;messageId&quot;
							</span>
							<span className="text-zinc-300">: </span>
							<span className="text-emerald-300">
								&quot;BAE5F4C2A1B3D6E8&quot;
							</span>
							<span className="text-zinc-300">,</span>
							{"\n"}
							<span className="text-zinc-300">{"  "}</span>
							<span className="text-sky-300">
								&quot;status&quot;
							</span>
							<span className="text-zinc-300">: </span>
							<span className="text-emerald-300">
								&quot;sent&quot;
							</span>
							{"\n"}
							<span className="text-zinc-600">{"}"}</span>
						</code>
					</pre>
				</div>
			</div>
		</motion.div>
	);
}

export function Hero() {
	return (
		<section className="relative overflow-hidden">
			{/* Background grid */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute inset-0 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] bg-[size:4rem_4rem] opacity-[0.25] [mask-image:radial-gradient(ellipse_70%_50%_at_50%_0%,black_30%,transparent_100%)]"
			/>

			{/* Primary gradient glow */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute -top-40 left-1/2 h-[600px] w-[900px] -translate-x-1/2 rounded-full bg-primary/8 blur-[140px]"
			/>

			{/* Secondary glow */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute top-20 -right-40 h-[300px] w-[400px] rounded-full bg-primary/5 blur-[100px]"
			/>

			<div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				<div className="flex flex-col items-center gap-14 pb-20 pt-24 sm:pb-28 sm:pt-32 lg:pb-32 lg:pt-36">
					{/* Text content */}
					<div className="flex max-w-3xl flex-col items-center text-center">
						{/* Badge */}
						<motion.div
							custom={0}
							variants={fadeUp}
							initial="hidden"
							animate="visible"
						>
							<Badge
								variant="outline"
								className="mb-7 gap-2 px-4 py-1.5 text-sm font-medium border-primary/20 bg-primary/5 text-primary"
							>
								<SparklesIcon className="size-3.5" />
								Nova versão 3.6 disponível
							</Badge>
						</motion.div>

						{/* Headline */}
						<motion.h1
							custom={1}
							variants={fadeUp}
							initial="hidden"
							animate="visible"
							className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl lg:text-6xl xl:text-7xl"
						>
							+500 empresas já enviam{" "}
							<br className="hidden sm:block" />
							pelo WhatsApp com{" "}
							<span className="text-primary">
								47ms de latência
							</span>
						</motion.h1>

						{/* Subtitle */}
						<motion.p
							custom={2}
							variants={fadeUp}
							initial="hidden"
							animate="visible"
							className="mt-6 max-w-2xl text-lg text-muted-foreground sm:text-xl leading-relaxed"
						>
							Sua equipe ainda depende de ferramentas lentas e
							instáveis para integrar o WhatsApp? Com a ZedaAPI,
							você conecta em minutos, escala sem limites e nunca
							mais perde uma mensagem.{" "}
							<span className="font-medium text-foreground">
								10 milhões de mensagens por mês
							</span>{" "}
							não mentem.
						</motion.p>

						{/* CTAs */}
						<motion.div
							custom={3}
							variants={fadeUp}
							initial="hidden"
							animate="visible"
							className="mt-10 flex flex-col items-center gap-3 sm:flex-row sm:gap-4"
						>
							<Button
								size="lg"
								asChild
								className="h-12 px-7 text-base shadow-lg shadow-primary/20 hover:shadow-xl hover:shadow-primary/25 transition-all duration-300"
							>
								<Link href="/sign-up">
									Começar Grátis — Vagas Limitadas
									<ArrowRightIcon
										className="size-4"
										data-icon="inline-end"
									/>
								</Link>
							</Button>
							<Button
								variant="outline"
								size="lg"
								asChild
								className="h-12 px-7 text-base"
							>
								<Link href="#precos">Comparar Planos</Link>
							</Button>
						</motion.div>

						{/* Trust signals */}
						<motion.div
							custom={4}
							variants={fadeUp}
							initial="hidden"
							animate="visible"
							className="mt-6 flex flex-wrap items-center justify-center gap-x-6 gap-y-2 text-sm text-muted-foreground"
						>
							<span className="flex items-center gap-1.5">
								<span className="size-1.5 rounded-full bg-emerald-500" />
								Pronto em 30 segundos
							</span>
							<span className="flex items-center gap-1.5">
								<span className="size-1.5 rounded-full bg-emerald-500" />
								Sem cartão de crédito
							</span>
							<span className="flex items-center gap-1.5">
								<span className="size-1.5 rounded-full bg-emerald-500" />
								99.9% de uptime garantido
							</span>
							<span className="flex items-center gap-1.5">
								<span className="size-1.5 rounded-full bg-emerald-500" />
								Suporte 100% brasileiro
							</span>
						</motion.div>
					</div>

					{/* Code block */}
					<CodeBlock />
				</div>
			</div>
		</section>
	);
}
