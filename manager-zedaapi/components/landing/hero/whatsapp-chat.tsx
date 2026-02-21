"use client";

import { useRef, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { LockIcon, PlayIcon, ImageIcon } from "lucide-react";

import type { ChatMessage } from "./use-hero-animation";

interface WhatsAppChatProps {
	messages: ChatMessage[];
	showTypingIndicator: boolean;
	reducedMotion: boolean;
}

// WhatsApp Web dark theme (fixed colors, independent of site theme)
const WA = {
	bg: "#0b141a",
	header: "#202c33",
	footer: "#202c33",
	bubble: "#005c4b",
	text: "#e9edef",
	timestamp: "#8696a0",
	tickRead: "#53bdeb",
	tickGray: "#8696a0",
	inputBg: "#2a3942",
} as const;

function TickIcon({
	status,
}: {
	status: "sending" | "sent" | "delivered" | "read";
}) {
	if (status === "sending") {
		return (
			<svg
				width="16"
				height="11"
				viewBox="0 0 16 11"
				className="inline-block ml-1"
			>
				<path
					d="M11.071.653a.457.457 0 0 0-.304-.102.493.493 0 0 0-.381.178l-6.19 7.636-2.011-2.095a.463.463 0 0 0-.349-.145.458.458 0 0 0-.345.161.41.41 0 0 0 .014.573l2.344 2.442a.469.469 0 0 0 .345.156h.027a.462.462 0 0 0 .345-.178l6.533-8.054a.41.41 0 0 0-.028-.572z"
					fill={WA.tickGray}
					opacity={0.5}
				/>
			</svg>
		);
	}

	if (status === "sent") {
		return (
			<svg
				width="16"
				height="11"
				viewBox="0 0 16 11"
				className="inline-block ml-1"
			>
				<path
					d="M11.071.653a.457.457 0 0 0-.304-.102.493.493 0 0 0-.381.178l-6.19 7.636-2.011-2.095a.463.463 0 0 0-.349-.145.458.458 0 0 0-.345.161.41.41 0 0 0 .014.573l2.344 2.442a.469.469 0 0 0 .345.156h.027a.462.462 0 0 0 .345-.178l6.533-8.054a.41.41 0 0 0-.028-.572z"
					fill={WA.tickGray}
				/>
			</svg>
		);
	}

	const color = status === "read" ? WA.tickRead : WA.tickGray;

	return (
		<svg
			width="16"
			height="11"
			viewBox="0 0 16 11"
			className="inline-block ml-1"
		>
			<path
				d="M11.071.653a.457.457 0 0 0-.304-.102.493.493 0 0 0-.381.178l-6.19 7.636-2.011-2.095a.463.463 0 0 0-.349-.145.458.458 0 0 0-.345.161.41.41 0 0 0 .014.573l2.344 2.442a.469.469 0 0 0 .345.156h.027a.462.462 0 0 0 .345-.178l6.533-8.054a.41.41 0 0 0-.028-.572z"
				fill={color}
			/>
			<path
				d="M15.071.653a.457.457 0 0 0-.304-.102.493.493 0 0 0-.381.178l-6.19 7.636-1.2-1.25-.353.435 1.218 1.27a.469.469 0 0 0 .345.155h.027a.462.462 0 0 0 .345-.178l6.533-8.054a.41.41 0 0 0-.04-.09z"
				fill={color}
			/>
		</svg>
	);
}

function TypingIndicator() {
	return (
		<div
			className="flex items-center gap-0.5 px-2.5 py-1.5 rounded-lg w-fit"
			style={{ backgroundColor: WA.bubble }}
		>
			{[0, 1, 2].map((i) => (
				<span
					key={i}
					className="size-[5px] rounded-full animate-bounce"
					style={{
						backgroundColor: WA.timestamp,
						animationDelay: `${i * 150}ms`,
						animationDuration: "0.8s",
					}}
				/>
			))}
		</div>
	);
}

function AudioWaveform() {
	const bars = [3, 5, 8, 4, 7, 10, 6, 9, 4, 7, 5, 8, 3, 6, 9, 5, 7, 4, 8, 6];
	return (
		<div className="flex items-center gap-[2px] h-4">
			{bars.map((h, i) => (
				<div
					key={i}
					className="w-[2px] rounded-full"
					style={{
						height: `${h * 1.6}px`,
						backgroundColor: WA.text,
						opacity: 0.7,
					}}
				/>
			))}
		</div>
	);
}

function TextBubble({ message }: { message: ChatMessage }) {
	return (
		<div
			className="max-w-[85%] rounded-lg px-2.5 py-1 ml-auto"
			style={{ backgroundColor: WA.bubble }}
		>
			<p className="text-xs leading-relaxed" style={{ color: WA.text }}>
				{message.text}
			</p>
			<div className="flex items-center justify-end gap-0.5 mt-0.5">
				<span className="text-[10px]" style={{ color: WA.timestamp }}>
					{message.timestamp}
				</span>
				<TickIcon status={message.status} />
			</div>
		</div>
	);
}

function ImageBubble({ message }: { message: ChatMessage }) {
	return (
		<div
			className="max-w-[85%] rounded-lg overflow-hidden ml-auto"
			style={{ backgroundColor: WA.bubble }}
		>
			{/* Image placeholder */}
			<div
				className="relative aspect-[4/3] w-40 flex items-center justify-center"
				style={{ backgroundColor: "#1a2e2a" }}
			>
				<ImageIcon
					className="size-6 opacity-30"
					style={{ color: WA.text }}
				/>
			</div>
			{message.caption && (
				<div className="px-2.5 py-1">
					<p
						className="text-xs leading-relaxed"
						style={{ color: WA.text }}
					>
						{message.caption}
					</p>
					<div className="flex items-center justify-end gap-0.5 mt-0.5">
						<span
							className="text-[10px]"
							style={{ color: WA.timestamp }}
						>
							{message.timestamp}
						</span>
						<TickIcon status={message.status} />
					</div>
				</div>
			)}
		</div>
	);
}

function AudioBubble({ message }: { message: ChatMessage }) {
	return (
		<div
			className="max-w-[85%] rounded-lg px-2.5 py-1.5 ml-auto"
			style={{ backgroundColor: WA.bubble }}
		>
			<div className="flex items-center gap-2">
				{/* Play button */}
				<div
					className="flex-shrink-0 size-6 rounded-full flex items-center justify-center"
					style={{ backgroundColor: "rgba(255,255,255,0.15)" }}
				>
					<PlayIcon
						className="size-3 ml-0.5"
						style={{ color: WA.text }}
					/>
				</div>
				<div className="flex-1 flex flex-col gap-0.5">
					<AudioWaveform />
					<div className="flex items-center justify-between">
						<span
							className="text-[10px]"
							style={{ color: WA.timestamp }}
						>
							{message.audioDuration}
						</span>
						<div className="flex items-center gap-0.5">
							<span
								className="text-[10px]"
								style={{ color: WA.timestamp }}
							>
								{message.timestamp}
							</span>
							<TickIcon status={message.status} />
						</div>
					</div>
				</div>
			</div>
		</div>
	);
}

function MessageBubble({
	message,
	reducedMotion,
}: {
	message: ChatMessage;
	reducedMotion: boolean;
}) {
	const content = (() => {
		switch (message.type) {
			case "image":
				return <ImageBubble message={message} />;
			case "audio":
				return <AudioBubble message={message} />;
			default:
				return <TextBubble message={message} />;
		}
	})();

	if (reducedMotion) {
		return <div className="flex justify-end">{content}</div>;
	}

	return (
		<motion.div
			initial={{ opacity: 0, y: 8, scale: 0.95 }}
			animate={{ opacity: 1, y: 0, scale: 1 }}
			transition={{ duration: 0.3, ease: [0.22, 1, 0.36, 1] }}
			className="flex justify-end"
		>
			{content}
		</motion.div>
	);
}

export function WhatsAppChat({
	messages,
	showTypingIndicator,
	reducedMotion,
}: WhatsAppChatProps) {
	const chatAreaRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		if (chatAreaRef.current) {
			chatAreaRef.current.scrollTop = chatAreaRef.current.scrollHeight;
		}
	}, [messages, showTypingIndicator]);

	return (
		<div
			role="img"
			aria-label="Preview do WhatsApp mostrando mensagens recebidas via API"
			className="w-full overflow-hidden rounded-xl border border-white/[0.08] shadow-2xl shadow-primary/5"
			style={{ backgroundColor: WA.bg }}
		>
			{/* Header */}
			<div
				className="flex items-center gap-2.5 px-3 py-2 border-b"
				style={{
					backgroundColor: WA.header,
					borderColor: "rgba(255,255,255,0.06)",
				}}
			>
				{/* Avatar */}
				<div className="size-7 rounded-full bg-gradient-to-br from-emerald-400 to-teal-600 flex items-center justify-center flex-shrink-0">
					<span
						className="text-[10px] font-bold"
						style={{ color: "#fff" }}
					>
						LO
					</span>
				</div>
				<div className="flex flex-col">
					<span
						className="text-xs font-medium"
						style={{ color: WA.text }}
					>
						Loja Online
					</span>
					<span className="text-[10px]" style={{ color: "#00a884" }}>
						online
					</span>
				</div>
			</div>

			{/* Chat area */}
			<div
				ref={chatAreaRef}
				className="flex flex-col gap-1.5 p-2.5 overflow-y-auto"
				style={{
					height: "200px",
					backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='0.02'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
				}}
			>
				<AnimatePresence mode="popLayout">
					{messages.map((msg) => (
						<MessageBubble
							key={msg.id}
							message={msg}
							reducedMotion={reducedMotion}
						/>
					))}
				</AnimatePresence>

				{showTypingIndicator && (
					<div className="flex justify-end">
						<TypingIndicator />
					</div>
				)}

				{messages.length === 0 && !showTypingIndicator && (
					<div className="flex-1 flex items-center justify-center">
						<p
							className="text-xs text-center opacity-40"
							style={{ color: WA.timestamp }}
						>
							As mensagens aparecerão aqui
						</p>
					</div>
				)}
			</div>

			{/* Footer */}
			<div
				className="flex items-center gap-1.5 px-2.5 py-1.5 border-t"
				style={{
					backgroundColor: WA.footer,
					borderColor: "rgba(255,255,255,0.06)",
				}}
			>
				<LockIcon
					className="size-2.5 flex-shrink-0"
					style={{ color: WA.timestamp }}
				/>
				<span className="text-[10px]" style={{ color: WA.timestamp }}>
					Criptografia de ponta a ponta
				</span>
			</div>
		</div>
	);
}
