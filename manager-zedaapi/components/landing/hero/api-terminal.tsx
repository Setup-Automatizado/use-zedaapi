"use client";

import { useState, useMemo } from "react";
import { motion } from "framer-motion";
import { CopyIcon, CheckIcon } from "lucide-react";

import type { DemoScenario } from "./animation-data";
interface ApiTerminalProps {
	scenario: DemoScenario;
	typedChars: number;
	showResponse: boolean;
	reducedMotion: boolean;
}

interface Token {
	text: string;
	className: string;
}

function tokenize(command: string): Token[] {
	const tokens: Token[] = [];
	const lines = command.split("\n");

	for (let i = 0; i < lines.length; i++) {
		const line = lines[i] ?? "";
		if (i > 0) tokens.push({ text: "\n", className: "" });

		// First line: $ curl -X POST \
		const curlMatch = line.match(/^(\$ )(curl)( -X POST)(.*)/);
		if (curlMatch) {
			tokens.push({
				text: curlMatch[1] ?? "",
				className: "text-zinc-600",
			});
			tokens.push({
				text: curlMatch[2] ?? "",
				className: "text-emerald-400 font-medium",
			});
			tokens.push({
				text: curlMatch[3] ?? "",
				className: "text-zinc-300",
			});
			if (curlMatch[4])
				tokens.push({ text: curlMatch[4], className: "text-zinc-600" });
			continue;
		}

		// URL line: https://...
		const urlMatch = line.match(/^(\s+)(https?:\/\/\S+)(.*)/);
		if (urlMatch) {
			tokens.push({ text: urlMatch[1] ?? "", className: "" });
			tokens.push({ text: urlMatch[2] ?? "", className: "text-sky-400" });
			if (urlMatch[3])
				tokens.push({ text: urlMatch[3], className: "text-zinc-600" });
			continue;
		}

		// Header lines: -H "..."
		const headerMatch = line.match(/^(\s+-H )(".*?")(.*)/);
		if (headerMatch) {
			tokens.push({
				text: headerMatch[1] ?? "",
				className: "text-zinc-300",
			});
			tokens.push({
				text: headerMatch[2] ?? "",
				className: "text-amber-300",
			});
			if (headerMatch[3])
				tokens.push({
					text: headerMatch[3],
					className: "text-zinc-600",
				});
			continue;
		}

		// Data flag: -d '{
		const dataMatch = line.match(/^(\s+-d )('{)/);
		if (dataMatch) {
			tokens.push({
				text: dataMatch[1] ?? "",
				className: "text-zinc-300",
			});
			tokens.push({
				text: dataMatch[2] ?? "",
				className: "text-amber-300",
			});
			continue;
		}

		// JSON key-value: "key": "value"
		const kvMatch = line.match(/^(\s+)("[\w]+?")(: )(".*?")(,?)$/);
		if (kvMatch) {
			tokens.push({ text: kvMatch[1] ?? "", className: "" });
			tokens.push({ text: kvMatch[2] ?? "", className: "text-sky-300" });
			tokens.push({ text: kvMatch[3] ?? "", className: "text-zinc-300" });
			tokens.push({
				text: kvMatch[4] ?? "",
				className: "text-emerald-300",
			});
			if (kvMatch[5])
				tokens.push({ text: kvMatch[5], className: "text-zinc-300" });
			continue;
		}

		// Closing brace: }'
		const closeMatch = line.match(/^(\s+)(}')/);
		if (closeMatch) {
			tokens.push({ text: closeMatch[1] ?? "", className: "" });
			tokens.push({
				text: closeMatch[2] ?? "",
				className: "text-amber-300",
			});
			continue;
		}

		// Fallback
		tokens.push({ text: line, className: "text-zinc-300" });
	}

	return tokens;
}

function tokenizeResponse(response: string): Token[] {
	const tokens: Token[] = [];
	const lines = response.split("\n");

	for (let i = 0; i < lines.length; i++) {
		const line = lines[i] ?? "";
		if (i > 0) tokens.push({ text: "\n", className: "" });

		// Braces
		if (line.trim() === "{" || line.trim() === "}") {
			tokens.push({ text: line, className: "text-zinc-600" });
			continue;
		}

		// Key-value pairs
		const kvMatch = line.match(/^(\s+)("[\w]+?")(: )(".*?")(,?)$/);
		if (kvMatch) {
			tokens.push({ text: kvMatch[1] ?? "", className: "" });
			tokens.push({ text: kvMatch[2] ?? "", className: "text-sky-300" });
			tokens.push({ text: kvMatch[3] ?? "", className: "text-zinc-300" });
			tokens.push({
				text: kvMatch[4] ?? "",
				className: "text-emerald-300",
			});
			if (kvMatch[5])
				tokens.push({ text: kvMatch[5], className: "text-zinc-300" });
			continue;
		}

		tokens.push({ text: line, className: "text-zinc-300" });
	}

	return tokens;
}

export function ApiTerminal({
	scenario,
	typedChars,
	showResponse,
	reducedMotion,
}: ApiTerminalProps) {
	const [copied, setCopied] = useState(false);

	const commandTokens = useMemo(
		() => tokenize(scenario.command),
		[scenario.command],
	);
	const responseTokens = useMemo(
		() => tokenizeResponse(scenario.response),
		[scenario.response],
	);

	function handleCopy() {
		navigator.clipboard.writeText(scenario.copyCommand);
		setCopied(true);
		setTimeout(() => setCopied(false), 2000);
	}

	// Map visible text onto tokens
	function renderTypedTokens() {
		if (reducedMotion) {
			return commandTokens.map((token, i) => (
				<span key={i} className={token.className}>
					{token.text}
				</span>
			));
		}

		let charCount = 0;
		const elements: React.ReactNode[] = [];

		for (let i = 0; i < commandTokens.length; i++) {
			const token = commandTokens[i];
			if (!token) continue;
			const tokenStart = charCount;
			const tokenEnd = charCount + token.text.length;

			if (typedChars <= tokenStart) break;

			const visibleLength = Math.min(
				typedChars - tokenStart,
				token.text.length,
			);
			elements.push(
				<span key={i} className={token.className}>
					{token.text.slice(0, visibleLength)}
				</span>,
			);

			charCount = tokenEnd;
		}

		// Cursor
		if (typedChars < scenario.command.length) {
			elements.push(
				<span
					key="cursor"
					className="inline-block w-[2px] h-[1.1em] bg-emerald-400 align-text-bottom animate-pulse"
				/>,
			);
		}

		return elements;
	}

	const shouldShowResponse = reducedMotion || showResponse;

	return (
		<div
			role="img"
			aria-label="Terminal demonstrando chamada API para envio de mensagem WhatsApp"
			className="w-full overflow-hidden rounded-xl border border-white/[0.08] bg-zinc-950 shadow-2xl shadow-primary/5"
		>
			{/* Window chrome */}
			<div className="flex items-center justify-between border-b border-white/[0.06] px-3 py-2">
				<div className="flex items-center gap-1.5">
					<div className="size-2 rounded-full bg-zinc-700" />
					<div className="size-2 rounded-full bg-zinc-700" />
					<div className="size-2 rounded-full bg-zinc-700" />
				</div>
				<span className="text-[10px] font-medium tracking-wide text-zinc-500 uppercase">
					Terminal
				</span>
				<button
					type="button"
					onClick={handleCopy}
					className="flex items-center gap-1 rounded-md px-2 py-0.5 text-[11px] text-zinc-500 transition-all duration-150 hover:bg-white/[0.06] hover:text-zinc-300 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950"
					aria-label="Copiar comando"
				>
					{copied ? (
						<CheckIcon className="size-3 text-emerald-400" />
					) : (
						<CopyIcon className="size-3" />
					)}
					{copied ? "Copiado!" : "Copiar"}
				</button>
			</div>

			{/* Code content */}
			<div className="px-3 py-3 sm:px-4">
				<pre className="font-mono text-[11px] leading-relaxed whitespace-pre-wrap break-all sm:break-normal sm:whitespace-pre">
					<code>{renderTypedTokens()}</code>
				</pre>

				{/* Response */}
				{shouldShowResponse && (
					<motion.div
						initial={reducedMotion ? false : { opacity: 0 }}
						animate={{ opacity: 1 }}
						transition={{ duration: 0.3 }}
						className="mt-3 border-t border-white/[0.04] pt-3"
					>
						<div className="mb-2 flex items-center gap-2">
							<div className="size-1.5 rounded-full bg-emerald-400 animate-pulse" />
							<p className="text-[10px] font-medium uppercase tracking-wider text-zinc-500">
								Resposta &mdash; 47ms
							</p>
						</div>
						<pre className="font-mono text-[11px] leading-relaxed">
							<code>
								{responseTokens.map((token, i) => (
									<span key={i} className={token.className}>
										{token.text}
									</span>
								))}
							</code>
						</pre>
					</motion.div>
				)}
			</div>
		</div>
	);
}
