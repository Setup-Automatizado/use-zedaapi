"use client";

import { useRef, useState, useEffect, useCallback } from "react";
import { motion, useInView } from "framer-motion";
import type { Variants } from "motion-dom";
import { CheckIcon } from "lucide-react";

import { fadeUp } from "@/lib/design-tokens";
import { useReducedMotion } from "@/components/shared/motion";

// ── Animations ──────────────────────────

const containerVariants: Variants = {
	hidden: {},
	visible: { transition: { staggerChildren: 0.2 } },
};

const stepVariants: Variants = {
	hidden: { opacity: 0, y: 24 },
	visible: {
		opacity: 1,
		y: 0,
		transition: { duration: 0.6, ease: [0.22, 1, 0.36, 1] },
	},
};

// ── Hook: looping timer-based animation ──────────────────────────

function useLoopingAnimation(active: boolean, reducedMotion: boolean | null) {
	const [cycle, setCycle] = useState(0);

	useEffect(() => {
		if (!active || reducedMotion) return;
		// increment cycle to restart animations
		const interval = setInterval(() => {
			setCycle((c) => c + 1);
		}, 7000); // full cycle every 7s
		return () => clearInterval(interval);
	}, [active, reducedMotion]);

	return cycle;
}

// ── Demo 1: Sign Up Form ──────────────────────────

function SignUpDemo({ active }: { active: boolean }) {
	const [emailChars, setEmailChars] = useState(0);
	const [passwordDots, setPasswordDots] = useState(0);
	const [buttonActive, setButtonActive] = useState(false);
	const [success, setSuccess] = useState(false);
	const reducedMotion = useReducedMotion();

	const email = "maria@empresa.com";
	const totalPassword = 8;

	const reset = useCallback(() => {
		setEmailChars(0);
		setPasswordDots(0);
		setButtonActive(false);
		setSuccess(false);
	}, []);

	const cycle = useLoopingAnimation(active, reducedMotion);

	useEffect(() => {
		if (!active) return;
		if (reducedMotion) {
			setSuccess(true);
			return;
		}

		reset();
		const timers: ReturnType<typeof setTimeout>[] = [];

		// Type email
		for (let i = 0; i <= email.length; i++) {
			timers.push(setTimeout(() => setEmailChars(i), 300 + i * 55));
		}

		const emailDone = 300 + email.length * 55 + 150;

		// Type password dots
		for (let i = 0; i <= totalPassword; i++) {
			timers.push(
				setTimeout(() => setPasswordDots(i), emailDone + i * 70),
			);
		}

		const passwordDone = emailDone + totalPassword * 70 + 200;

		// Button press
		timers.push(setTimeout(() => setButtonActive(true), passwordDone));

		// Success
		timers.push(setTimeout(() => setSuccess(true), passwordDone + 500));

		return () => timers.forEach(clearTimeout);
	}, [active, reducedMotion, reset, cycle]);

	return (
		<div className="relative flex h-48 flex-col items-center justify-center gap-2.5 overflow-hidden rounded-t-xl bg-zinc-950 p-5">
			{/* Subtle gradient glow */}
			<div className="pointer-events-none absolute inset-0 bg-gradient-to-b from-emerald-500/[0.03] to-transparent" />

			{success ? (
				<motion.div
					initial={{ opacity: 0, scale: 0.8 }}
					animate={{ opacity: 1, scale: 1 }}
					transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
					className="flex flex-col items-center gap-3"
				>
					<div className="relative">
						<div className="flex size-12 items-center justify-center rounded-full bg-emerald-500/20 ring-1 ring-emerald-500/30">
							<CheckIcon className="size-6 text-emerald-400" />
						</div>
						{/* Pulse ring */}
						<div className="absolute inset-0 animate-ping rounded-full bg-emerald-500/10" />
					</div>
					<span className="text-sm font-medium text-emerald-400">
						Conta criada!
					</span>
				</motion.div>
			) : (
				<div className="w-full max-w-[220px] space-y-2.5">
					{/* Email field */}
					<div>
						<div className="mb-1 text-[10px] font-medium text-zinc-500">
							E-mail
						</div>
						<div
							className={`flex h-8 items-center rounded-lg border px-2.5 transition-colors duration-200 ${
								emailChars > 0 && emailChars < email.length
									? "border-emerald-500/40 bg-zinc-900 shadow-[0_0_8px_rgba(16,185,129,0.08)]"
									: "border-zinc-800 bg-zinc-900/80"
							}`}
						>
							<span className="font-mono text-[12px] text-zinc-300">
								{email.slice(0, emailChars)}
							</span>
							{emailChars > 0 && emailChars < email.length && (
								<span className="ml-px inline-block h-3.5 w-[2px] animate-pulse rounded-full bg-emerald-400" />
							)}
						</div>
					</div>

					{/* Password field */}
					<div>
						<div className="mb-1 text-[10px] font-medium text-zinc-500">
							Senha
						</div>
						<div
							className={`flex h-8 items-center gap-[3px] rounded-lg border px-2.5 transition-colors duration-200 ${
								passwordDots > 0 && passwordDots < totalPassword
									? "border-emerald-500/40 bg-zinc-900 shadow-[0_0_8px_rgba(16,185,129,0.08)]"
									: "border-zinc-800 bg-zinc-900/80"
							}`}
						>
							{Array.from({ length: passwordDots }).map(
								(_, i) => (
									<motion.span
										key={i}
										initial={{ scale: 0 }}
										animate={{ scale: 1 }}
										transition={{ duration: 0.1 }}
										className="size-[5px] rounded-full bg-zinc-400"
									/>
								),
							)}
							{passwordDots > 0 &&
								passwordDots < totalPassword && (
									<span className="ml-px inline-block h-3.5 w-[2px] animate-pulse rounded-full bg-emerald-400" />
								)}
						</div>
					</div>

					{/* Button */}
					<motion.div
						animate={
							buttonActive
								? { scale: [1, 0.96, 1] }
								: { scale: 1 }
						}
						transition={{ duration: 0.3 }}
						className={`flex h-8 cursor-default items-center justify-center rounded-lg text-[12px] font-semibold transition-all duration-300 ${
							buttonActive
								? "bg-emerald-500 text-white shadow-lg shadow-emerald-500/25"
								: "bg-zinc-800 text-zinc-500"
						}`}
					>
						Criar conta
					</motion.div>
				</div>
			)}
		</div>
	);
}

// ── Demo 2: QR Code Scan ──────────────────────────

const QR_SIZE = 11;
const QR_PATTERN = [
	[1, 1, 1, 1, 1, 0, 1, 0, 1, 1, 1],
	[1, 0, 0, 0, 1, 0, 0, 1, 1, 0, 1],
	[1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1],
	[1, 0, 0, 0, 1, 0, 0, 1, 1, 0, 1],
	[1, 1, 1, 1, 1, 0, 1, 0, 1, 1, 1],
	[0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0],
	[1, 0, 1, 0, 1, 1, 0, 1, 1, 0, 1],
	[0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0],
	[1, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1],
	[1, 0, 0, 0, 1, 0, 1, 0, 1, 0, 1],
	[1, 1, 1, 1, 1, 0, 1, 0, 1, 1, 1],
];

function QRCodeDemo({ active }: { active: boolean }) {
	const [scanProgress, setScanProgress] = useState(0);
	const [connected, setConnected] = useState(false);
	const reducedMotion = useReducedMotion();

	const reset = useCallback(() => {
		setScanProgress(0);
		setConnected(false);
	}, []);

	const cycle = useLoopingAnimation(active, reducedMotion);

	useEffect(() => {
		if (!active) return;
		if (reducedMotion) {
			setConnected(true);
			return;
		}

		reset();
		const timers: ReturnType<typeof setTimeout>[] = [];
		const totalSteps = 50;

		for (let i = 0; i <= totalSteps; i++) {
			timers.push(
				setTimeout(() => setScanProgress(i / totalSteps), 400 + i * 60),
			);
		}

		const scanDone = 400 + totalSteps * 60 + 300;
		timers.push(setTimeout(() => setConnected(true), scanDone));

		return () => timers.forEach(clearTimeout);
	}, [active, reducedMotion, reset, cycle]);

	return (
		<div className="relative flex h-48 items-center justify-center overflow-hidden rounded-t-xl bg-zinc-950 p-5">
			{/* Subtle gradient glow */}
			<div className="pointer-events-none absolute inset-0 bg-gradient-to-b from-sky-500/[0.03] to-transparent" />

			{connected ? (
				<motion.div
					initial={{ opacity: 0, scale: 0.8 }}
					animate={{ opacity: 1, scale: 1 }}
					transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
					className="flex flex-col items-center gap-3"
				>
					<div className="relative">
						<div className="flex size-12 items-center justify-center rounded-full bg-sky-500/20 ring-1 ring-sky-500/30">
							<CheckIcon className="size-6 text-sky-400" />
						</div>
						<div className="absolute inset-0 animate-ping rounded-full bg-sky-500/10" />
					</div>
					<div className="flex items-center gap-2">
						<span className="size-2 rounded-full bg-sky-400 shadow-[0_0_6px_rgba(56,189,248,0.6)]" />
						<span className="text-sm font-medium text-sky-400">
							WhatsApp conectado
						</span>
					</div>
				</motion.div>
			) : (
				<div className="relative">
					{/* QR Code grid */}
					<div
						className="grid gap-[3px]"
						style={{
							gridTemplateColumns: `repeat(${QR_SIZE}, 1fr)`,
						}}
					>
						{QR_PATTERN.flat().map((filled, i) => {
							const row = Math.floor(i / QR_SIZE);
							const isScanned =
								scanProgress > 0 &&
								row / QR_SIZE < scanProgress;
							return (
								<div
									key={i}
									className={`size-[7px] rounded-[1px] transition-all duration-200 ${
										filled
											? isScanned
												? "bg-sky-400/80 shadow-[0_0_4px_rgba(56,189,248,0.3)]"
												: "bg-zinc-400"
											: "bg-transparent"
									}`}
								/>
							);
						})}
					</div>

					{/* Scan line */}
					{scanProgress > 0 && scanProgress < 1 && (
						<div
							className="pointer-events-none absolute -left-2 -right-2 h-[2px]"
							style={{
								top: `${scanProgress * 100}%`,
								transition: "top 60ms linear",
								background:
									"linear-gradient(90deg, transparent 0%, rgba(56,189,248,0.8) 30%, rgba(56,189,248,1) 50%, rgba(56,189,248,0.8) 70%, transparent 100%)",
								boxShadow:
									"0 0 12px rgba(56,189,248,0.5), 0 0 24px rgba(56,189,248,0.2)",
							}}
						/>
					)}
				</div>
			)}
		</div>
	);
}

// ── Demo 3: Terminal API ──────────────────────────

function TerminalDemo({ active }: { active: boolean }) {
	const [typedChars, setTypedChars] = useState(0);
	const [showResponse, setShowResponse] = useState(false);
	const reducedMotion = useReducedMotion();

	const command =
		'$ curl -X POST .../send-text \\\n  -d \'{"phone":"5521..."}\'\n  -H "Client-Token: zApi..."';
	const responseLine1 = '{ "status": "sent",';
	const responseLine2 = '  "messageId": "BAE5..." }';

	const reset = useCallback(() => {
		setTypedChars(0);
		setShowResponse(false);
	}, []);

	const cycle = useLoopingAnimation(active, reducedMotion);

	useEffect(() => {
		if (!active) return;
		if (reducedMotion) {
			setTypedChars(command.length);
			setShowResponse(true);
			return;
		}

		reset();
		const timers: ReturnType<typeof setTimeout>[] = [];

		for (let i = 0; i <= command.length; i++) {
			timers.push(setTimeout(() => setTypedChars(i), 300 + i * 28));
		}

		const typeDone = 300 + command.length * 28 + 400;
		timers.push(setTimeout(() => setShowResponse(true), typeDone));

		return () => timers.forEach(clearTimeout);
	}, [active, reducedMotion, command.length, reset, cycle]);

	function renderCommand() {
		const visible = command.slice(0, typedChars);
		const lines = visible.split("\n");

		return lines.map((line, li) => (
			<span key={li}>
				{li > 0 && "\n"}
				{line.startsWith("$ ") ? (
					<>
						<span className="text-zinc-600">$ </span>
						<span className="font-medium text-violet-400">
							curl
						</span>
						<span className="text-zinc-400"> -X POST</span>
						<span className="text-sky-400">{line.slice(17)}</span>
					</>
				) : line.trimStart().startsWith("-d") ? (
					<>
						<span className="text-zinc-500">
							{line.slice(0, line.indexOf("-d") + 3)}
						</span>
						<span className="text-amber-300">
							{line.slice(line.indexOf("-d") + 3)}
						</span>
					</>
				) : line.trimStart().startsWith("-H") ? (
					<>
						<span className="text-zinc-500">
							{line.slice(0, line.indexOf("-H") + 3)}
						</span>
						<span className="text-amber-300">
							{line.slice(line.indexOf("-H") + 3)}
						</span>
					</>
				) : (
					<span className="text-zinc-300">{line}</span>
				)}
			</span>
		));
	}

	return (
		<div className="relative flex h-48 flex-col overflow-hidden rounded-t-xl bg-zinc-950 p-4">
			{/* Subtle gradient glow */}
			<div className="pointer-events-none absolute inset-0 bg-gradient-to-b from-violet-500/[0.03] to-transparent" />

			{/* Chrome */}
			<div className="mb-2.5 flex items-center gap-1.5">
				<div className="size-2 rounded-full bg-zinc-800" />
				<div className="size-2 rounded-full bg-zinc-800" />
				<div className="size-2 rounded-full bg-zinc-800" />
				<span className="ml-2 text-[9px] font-medium uppercase tracking-wider text-zinc-600">
					Terminal
				</span>
			</div>

			{/* Command */}
			<pre className="flex-1 font-mono text-[11px] leading-[1.6] whitespace-pre-wrap">
				<code>
					{renderCommand()}
					{typedChars > 0 && typedChars < command.length && (
						<span className="ml-px inline-block h-[1.1em] w-[2px] animate-pulse rounded-full bg-violet-400 align-text-bottom" />
					)}
				</code>
			</pre>

			{/* Response */}
			{showResponse && (
				<motion.div
					initial={{ opacity: 0, y: 4 }}
					animate={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.3 }}
					className="mt-2 border-t border-zinc-800/60 pt-2"
				>
					<div className="mb-1 flex items-center gap-2">
						<div className="size-1.5 rounded-full bg-emerald-400 shadow-[0_0_6px_rgba(52,211,153,0.6)]" />
						<span className="font-mono text-[10px] font-medium text-emerald-400/80">
							200 OK
						</span>
						<span className="font-mono text-[10px] text-zinc-600">
							— 47ms
						</span>
					</div>
					<pre className="font-mono text-[10px] leading-relaxed">
						<span className="text-zinc-500">{responseLine1}</span>
						{"\n"}
						<span className="text-emerald-300/70">
							{responseLine2}
						</span>
					</pre>
				</motion.div>
			)}
		</div>
	);
}

// ── Step Data ──────────────────────────

interface Step {
	number: string;
	title: string;
	description: string;
	accentColor: string;
	accentBg: string;
	glowColor: string;
	Demo: React.ComponentType<{ active: boolean }>;
}

const steps: Step[] = [
	{
		number: "1",
		title: "Crie sua conta",
		description:
			"Nome, e-mail e senha. Acesso imediato ao painel com 7 dias grátis.",
		accentColor: "text-emerald-500",
		accentBg: "bg-emerald-500/10",
		glowColor: "group-hover:shadow-emerald-500/[0.06]",
		Demo: SignUpDemo,
	},
	{
		number: "2",
		title: "Conecte o WhatsApp",
		description:
			"Escaneie o QR Code no painel — como abrir o WhatsApp Web.",
		accentColor: "text-sky-500",
		accentBg: "bg-sky-500/10",
		glowColor: "group-hover:shadow-sky-500/[0.06]",
		Demo: QRCodeDemo,
	},
	{
		number: "3",
		title: "Envie via API",
		description:
			"Um curl no terminal e a mensagem sai. Documentação com exemplos.",
		accentColor: "text-violet-500",
		accentBg: "bg-violet-500/10",
		glowColor: "group-hover:shadow-violet-500/[0.06]",
		Demo: TerminalDemo,
	},
];

// ── Step Card ──────────────────────────

function StepCard({
	step,
	isLast,
	active,
	index,
}: {
	step: Step;
	isLast: boolean;
	active: boolean;
	index: number;
}) {
	return (
		<motion.div
			variants={stepVariants}
			className={`group relative flex flex-col overflow-hidden rounded-2xl border border-border/60 bg-card transition-all duration-300 hover:-translate-y-1 hover:shadow-xl ${step.glowColor}`}
		>
			{/* Gradient accent top line */}
			<div
				aria-hidden="true"
				className={`pointer-events-none absolute inset-x-0 top-0 h-px ${
					index === 0
						? "bg-gradient-to-r from-transparent via-emerald-500/40 to-transparent"
						: index === 1
							? "bg-gradient-to-r from-transparent via-sky-500/40 to-transparent"
							: "bg-gradient-to-r from-transparent via-violet-500/40 to-transparent"
				}`}
			/>

			{/* Demo area */}
			<step.Demo active={active} />

			{/* Text content */}
			<div className="flex flex-col items-center border-t border-border/40 px-5 pb-6 pt-5 text-center">
				{/* Step number badge */}
				<div
					className={`mb-3 flex size-7 items-center justify-center rounded-full text-xs font-bold ${step.accentBg} ${step.accentColor}`}
				>
					{step.number}
				</div>

				<h3 className="text-base font-semibold text-foreground">
					{step.title}
				</h3>
				<p className="mt-2 max-w-[240px] text-sm leading-relaxed text-muted-foreground">
					{step.description}
				</p>
			</div>

			{/* Connector — desktop only */}
			{!isLast && (
				<div
					aria-hidden="true"
					className="pointer-events-none absolute right-0 top-24 hidden w-[calc(100%-3rem)] translate-x-[calc(50%+1.5rem)] md:block"
				>
					<div className="h-px w-full border-t border-dashed border-border/60" />
				</div>
			)}
		</motion.div>
	);
}

// ── Main ──────────────────────────

export function HowItWorks() {
	const ref = useRef<HTMLDivElement>(null);
	const isInView = useInView(ref, { once: true, margin: "-80px" });

	return (
		<section
			id="como-funciona"
			className="relative border-t border-border/50 bg-muted/20 py-20 sm:py-28"
		>
			<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8" ref={ref}>
				{/* Section header */}
				<motion.div
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
						Da conta à primeira mensagem{" "}
						<br className="hidden sm:block" />
						em <span className="text-primary">5 minutos</span>
					</motion.h2>
					<motion.p
						custom={1}
						variants={fadeUp}
						className="mt-4 max-w-md text-base text-muted-foreground sm:text-lg"
					>
						Três passos, nenhum obstáculo. Sem aprovação manual.
					</motion.p>
				</motion.div>

				{/* Steps */}
				<motion.div
					variants={containerVariants}
					initial="hidden"
					animate={isInView ? "visible" : "hidden"}
					className="relative mt-14 grid grid-cols-1 gap-6 sm:mt-20 md:grid-cols-3 md:gap-5"
				>
					{steps.map((step, i) => (
						<StepCard
							key={step.number}
							step={step}
							isLast={i === steps.length - 1}
							active={isInView}
							index={i}
						/>
					))}
				</motion.div>
			</div>
		</section>
	);
}
