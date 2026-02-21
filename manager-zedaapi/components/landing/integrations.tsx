"use client";

import { useRef, type ComponentType, type SVGProps } from "react";
import { motion, useInView } from "framer-motion";
import type { Variants } from "motion-dom";
import { ArrowRight, Globe } from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { fadeUp } from "@/lib/design-tokens";
import N8nIcon from "@/components/icons/n8n";
import MakeIcon from "@/components/icons/make";
import ZapierIcon from "@/components/icons/zapier";
import NodejsIcon from "@/components/icons/nodejs";
import PythonIcon from "@/components/icons/python";
import PhpIcon from "@/components/icons/php";
import WebhookIcon from "@/components/icons/webhook";

// ── Types ──────────────────────────

interface Integration {
	icon: LucideIcon | ComponentType<SVGProps<SVGSVGElement>>;
	name: string;
	description: string;
	/** Brand hex color for glow effects */
	brandColor: string;
	tag?: string;
	/** External link for "Saiba mais" */
	href: string;
}

// ── Data ──────────────────────────

const noCodeTools: Integration[] = [
	{
		icon: N8nIcon,
		name: "n8n",
		description:
			"Open-source com nodes nativos prontos — arraste, conecte e automatize.",
		brandColor: "#EA4B71",
		tag: "Popular",
		href: "https://www.npmjs.com/package/@setup-automatizado/n8n-nodes-zedaapi",
	},
	{
		icon: MakeIcon,
		name: "Make (Integromat)",
		description:
			"Builder visual com centenas de cenários prontos, sem código.",
		brandColor: "#6D00CC",
		href: "https://api.zedaapi.com/docs",
	},
	{
		icon: ZapierIcon,
		name: "Zapier",
		description:
			"Conecte WhatsApp a 6.000+ apps com workflows automatizados.",
		brandColor: "#FF4A00",
		href: "https://api.zedaapi.com/docs",
	},
];

const devTools: Integration[] = [
	{
		icon: NodejsIcon,
		name: "Node.js / TypeScript",
		description: "Fetch ou Axios com respostas JSON tipadas.",
		brandColor: "#539E43",
		href: "https://api.zedaapi.com/docs",
	},
	{
		icon: PythonIcon,
		name: "Python",
		description: "Requests e pronto. Ideal para bots e IA.",
		brandColor: "#387EB8",
		href: "https://api.zedaapi.com/docs",
	},
	{
		icon: PhpIcon,
		name: "PHP / Laravel",
		description: "Guzzle ou Http facade do Laravel.",
		brandColor: "#777BB4",
		href: "https://api.zedaapi.com/docs",
	},
	{
		icon: Globe,
		name: "Qualquer Linguagem",
		description: "REST padrão com HTTP/JSON.",
		brandColor: "#71717a",
		href: "https://api.zedaapi.com/docs",
	},
	{
		icon: WebhookIcon,
		name: "Webhooks",
		description: "Eventos em tempo real com retry.",
		brandColor: "#e91e63",
		href: "https://api.zedaapi.com/docs",
	},
];

// ── Code snippet lines ──────────────────────────

const codeLines = [
	{
		tokens: [
			{ text: "const ", color: "text-violet-400" },
			{ text: "res", color: "text-zinc-300" },
			{ text: " = ", color: "text-zinc-500" },
			{ text: "await ", color: "text-violet-400" },
			{ text: "fetch", color: "text-sky-400" },
			{ text: "(url, {", color: "text-zinc-400" },
		],
	},
	{
		tokens: [
			{ text: "  method", color: "text-zinc-300" },
			{ text: ": ", color: "text-zinc-500" },
			{ text: '"POST"', color: "text-amber-300" },
			{ text: ",", color: "text-zinc-500" },
		],
	},
	{
		tokens: [
			{ text: "  headers", color: "text-zinc-300" },
			{ text: ": { ", color: "text-zinc-500" },
			{ text: '"Client-Token"', color: "text-amber-300" },
			{ text: ": token },", color: "text-zinc-500" },
		],
	},
	{
		tokens: [
			{ text: "  body", color: "text-zinc-300" },
			{ text: ": ", color: "text-zinc-500" },
			{ text: "JSON", color: "text-sky-400" },
			{ text: ".", color: "text-zinc-500" },
			{ text: "stringify", color: "text-sky-400" },
			{ text: "({", color: "text-zinc-400" },
		],
	},
	{
		tokens: [
			{ text: "    phone", color: "text-zinc-300" },
			{ text: ": ", color: "text-zinc-500" },
			{ text: '"5521999887766"', color: "text-amber-300" },
			{ text: ",", color: "text-zinc-500" },
		],
	},
	{
		tokens: [
			{ text: "    message", color: "text-zinc-300" },
			{ text: ": ", color: "text-zinc-500" },
			{ text: '"Hello from ZedaAPI!"', color: "text-emerald-300" },
		],
	},
	{ tokens: [{ text: "  })", color: "text-zinc-400" }] },
	{ tokens: [{ text: "});", color: "text-zinc-400" }] },
];

// ── Animations ──────────────────────────

const stagger: Variants = {
	hidden: {},
	visible: { transition: { staggerChildren: 0.06 } },
};

const cardIn: Variants = {
	hidden: { opacity: 0, y: 20 },
	visible: {
		opacity: 1,
		y: 0,
		transition: { duration: 0.5, ease: [0.22, 1, 0.36, 1] },
	},
};

const codeLineIn: Variants = {
	hidden: { opacity: 0, x: -8 },
	visible: {
		opacity: 1,
		x: 0,
		transition: { duration: 0.4, ease: [0.22, 1, 0.36, 1] },
	},
};

// ── No-Code Card ──────────────────────────

function NoCodeCard({ integration }: { integration: Integration }) {
	const Icon = integration.icon;

	return (
		<motion.a
			href={integration.href}
			target="_blank"
			rel="noopener noreferrer"
			variants={cardIn}
			className="group relative flex flex-col overflow-hidden rounded-2xl border border-border/60 bg-card transition-all duration-300 hover:-translate-y-1 hover:shadow-xl hover:shadow-black/[0.08] dark:hover:shadow-black/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
		>
			{/* Top gradient accent line */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute inset-x-0 top-0 h-px"
				style={{
					background: `linear-gradient(90deg, transparent, ${integration.brandColor}40, transparent)`,
				}}
			/>

			<div className="flex flex-1 flex-col p-6">
				{/* Icon + Tag row */}
				<div className="flex items-start justify-between">
					<div
						className="flex size-12 items-center justify-center rounded-xl border border-border/50 transition-all duration-300 group-hover:scale-105"
						style={{
							background: `linear-gradient(135deg, ${integration.brandColor}08, ${integration.brandColor}15)`,
							boxShadow: `0 0 0 1px ${integration.brandColor}10`,
						}}
					>
						<Icon className="size-6" />
					</div>

					{integration.tag && (
						<span
							className="rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider"
							style={{
								background: `${integration.brandColor}15`,
								color: integration.brandColor,
							}}
						>
							{integration.tag}
						</span>
					)}
				</div>

				{/* Content */}
				<h3 className="mt-5 text-base font-semibold text-foreground">
					{integration.name}
				</h3>
				<p className="mt-2 flex-1 text-sm leading-relaxed text-muted-foreground">
					{integration.description}
				</p>

				{/* Hover arrow */}
				<div className="mt-4 flex items-center gap-1.5 text-xs font-medium text-muted-foreground/60 transition-all duration-300 group-hover:text-foreground group-hover:translate-x-0.5">
					<span>Saiba mais</span>
					<ArrowRight className="size-3 transition-transform duration-300 group-hover:translate-x-0.5" />
				</div>
			</div>

			{/* Bottom glow on hover */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute inset-x-0 bottom-0 h-24 opacity-0 transition-opacity duration-500 group-hover:opacity-100"
				style={{
					background: `radial-gradient(ellipse at bottom, ${integration.brandColor}06, transparent 70%)`,
				}}
			/>
		</motion.a>
	);
}

// ── Dev Tool Compact Card ──────────────────────────

function DevToolCard({ integration }: { integration: Integration }) {
	const Icon = integration.icon;

	return (
		<motion.div
			variants={cardIn}
			className="group relative flex items-center gap-4 rounded-xl border border-border/50 bg-card/80 p-4 transition-all duration-300 hover:-translate-y-0.5 hover:border-border hover:bg-card hover:shadow-lg hover:shadow-black/[0.04] dark:hover:shadow-black/15"
		>
			<div
				className="flex size-10 shrink-0 items-center justify-center rounded-lg transition-all duration-300 group-hover:scale-105"
				style={{
					background: `${integration.brandColor}12`,
				}}
			>
				<Icon className="size-5" />
			</div>
			<div className="min-w-0">
				<h3 className="text-sm font-semibold text-foreground">
					{integration.name}
				</h3>
				<p className="mt-0.5 text-xs leading-relaxed text-muted-foreground">
					{integration.description}
				</p>
			</div>
		</motion.div>
	);
}

// ── Code Snippet Panel ──────────────────────────

function CodeSnippet({ isInView }: { isInView: boolean }) {
	return (
		<motion.div
			initial="hidden"
			animate={isInView ? "visible" : "hidden"}
			variants={{
				hidden: { opacity: 0, y: 20 },
				visible: {
					opacity: 1,
					y: 0,
					transition: {
						duration: 0.6,
						ease: [0.22, 1, 0.36, 1],
						delay: 0.3,
					},
				},
			}}
			className="relative overflow-hidden rounded-2xl border border-border/40 bg-zinc-950 shadow-2xl shadow-black/20"
		>
			{/* Chrome bar */}
			<div className="flex items-center gap-2 border-b border-zinc-800/60 px-4 py-3">
				<div className="flex gap-1.5">
					<div className="size-2.5 rounded-full bg-zinc-800" />
					<div className="size-2.5 rounded-full bg-zinc-800" />
					<div className="size-2.5 rounded-full bg-zinc-800" />
				</div>
				<span className="ml-2 text-[10px] font-medium text-zinc-600">
					send-text.ts
				</span>
				<div className="ml-auto flex items-center gap-1.5">
					<span className="size-1.5 rounded-full bg-emerald-500/80 shadow-[0_0_4px_rgba(16,185,129,0.5)]" />
					<span className="text-[9px] font-medium text-emerald-500/70">
						200 OK
					</span>
				</div>
			</div>

			{/* Code lines */}
			<motion.div
				variants={{
					hidden: {},
					visible: {
						transition: {
							staggerChildren: 0.05,
							delayChildren: 0.5,
						},
					},
				}}
				initial="hidden"
				animate={isInView ? "visible" : "hidden"}
				className="p-4"
			>
				<pre className="font-mono text-[12px] leading-[1.8] sm:text-[13px]">
					{codeLines.map((line, li) => (
						<motion.div key={li} variants={codeLineIn}>
							<span className="mr-4 select-none text-zinc-700">
								{String(li + 1).padStart(2, " ")}
							</span>
							{line.tokens.map((token, ti) => (
								<span key={ti} className={token.color}>
									{token.text}
								</span>
							))}
						</motion.div>
					))}
				</pre>
			</motion.div>

			{/* Gradient glow */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute -right-20 -top-20 size-60 rounded-full bg-violet-500/[0.04] blur-3xl"
			/>
			<div
				aria-hidden="true"
				className="pointer-events-none absolute -bottom-10 -left-10 size-40 rounded-full bg-emerald-500/[0.03] blur-3xl"
			/>
		</motion.div>
	);
}

// ── Main ──────────────────────────

export function Integrations() {
	const ref = useRef<HTMLDivElement>(null);
	const isInView = useInView(ref, { once: true, margin: "-80px" });

	return (
		<section
			id="integracoes"
			className="relative border-t border-border/50 bg-muted/20 py-20 sm:py-28"
		>
			<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8" ref={ref}>
				{/* Section header */}
				<motion.div
					initial="hidden"
					animate={isInView ? "visible" : "hidden"}
					className="mx-auto max-w-2xl text-center"
				>
					<motion.p
						custom={0}
						variants={fadeUp}
						className="text-sm font-semibold uppercase tracking-widest text-primary"
					>
						Integrações
					</motion.p>
					<motion.h2
						custom={0.5}
						variants={fadeUp}
						className="mt-3 text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl"
					>
						Conecte com as ferramentas{" "}
						<span className="text-primary">que você já usa</span>
					</motion.h2>
					<motion.p
						custom={1}
						variants={fadeUp}
						className="mt-4 text-base leading-relaxed text-muted-foreground sm:text-lg"
					>
						API REST padrão com HTTP/JSON — funciona com qualquer
						linguagem ou ferramenta no-code.
					</motion.p>
				</motion.div>

				{/* No-Code section label */}
				<motion.div
					initial={{ opacity: 0 }}
					animate={isInView ? { opacity: 1 } : { opacity: 0 }}
					transition={{ delay: 0.2, duration: 0.5 }}
					className="mb-5 mt-14 flex items-center gap-3 sm:mt-20"
				>
					<span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
						No-Code
					</span>
					<div className="h-px flex-1 bg-border/50" />
				</motion.div>

				{/* No-Code tools — 3 featured cards */}
				<motion.div
					variants={stagger}
					initial="hidden"
					animate={isInView ? "visible" : "hidden"}
					className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3"
				>
					{noCodeTools.map((integration) => (
						<NoCodeCard
							key={integration.name}
							integration={integration}
						/>
					))}
				</motion.div>

				{/* Code snippet + Dev tools — two-column layout */}
				<div className="mt-10 grid grid-cols-1 gap-8 lg:grid-cols-2">
					{/* Left: Code snippet */}
					<div className="flex flex-col">
						<motion.div
							initial={{ opacity: 0 }}
							animate={isInView ? { opacity: 1 } : { opacity: 0 }}
							transition={{ delay: 0.35, duration: 0.5 }}
							className="mb-5 flex items-center gap-3"
						>
							<span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
								Para Devs
							</span>
							<div className="h-px flex-1 bg-border/50" />
						</motion.div>

						<CodeSnippet isInView={isInView} />

						{/* Caption under code */}
						<motion.p
							initial={{ opacity: 0 }}
							animate={isInView ? { opacity: 1 } : { opacity: 0 }}
							transition={{ delay: 0.8, duration: 0.5 }}
							className="mt-3 text-center text-xs text-muted-foreground/60"
						>
							8 linhas para enviar a primeira mensagem
						</motion.p>
					</div>

					{/* Right: Dev tool compact cards */}
					<div className="flex flex-col">
						{/* Spacer to align with "Para Devs" label on mobile (hidden on lg since left side has it) */}
						<div className="mb-5 hidden items-center gap-3 lg:flex">
							<span className="text-xs font-semibold uppercase tracking-wider text-transparent">
								&nbsp;
							</span>
							<div className="h-px flex-1 bg-transparent" />
						</div>

						<motion.div
							variants={stagger}
							initial="hidden"
							animate={isInView ? "visible" : "hidden"}
							className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-1"
						>
							{devTools.map((integration) => (
								<DevToolCard
									key={integration.name}
									integration={integration}
								/>
							))}
						</motion.div>
					</div>
				</div>
			</div>
		</section>
	);
}
