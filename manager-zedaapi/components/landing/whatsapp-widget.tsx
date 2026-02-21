"use client";

import { useRef, useState, useTransition, useEffect } from "react";
import { AnimatePresence, motion } from "framer-motion";
import {
	CheckCircle2,
	MessageCircle,
	X,
	ArrowLeft,
	ArrowRight,
	Send,
	Mail,
	Loader2,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import WhatsAppIcon from "@/components/icons/whatsapp";
import { submitWidgetForm } from "@/server/actions/contact";

const WHATSAPP_NUMBER =
	process.env.NEXT_PUBLIC_CONTACT_WHATSAPP ?? "5521971532700";
const WHATSAPP_URL = `https://wa.me/${WHATSAPP_NUMBER}?text=${encodeURIComponent("Olá! Vim pelo site do Zé da API e gostaria de saber mais sobre a plataforma.")}`;

type WidgetView = "closed" | "menu" | "form" | "success";

export function WhatsAppWidget() {
	const [view, setView] = useState<WidgetView>("closed");
	const [isPending, startTransition] = useTransition();
	const [formError, setFormError] = useState<string | null>(null);
	const [preference, setPreference] = useState<"email" | "whatsapp">(
		"whatsapp",
	);
	const [hasInteracted, setHasInteracted] = useState(false);
	const popupRef = useRef<HTMLDivElement>(null);

	// Auto-stop pulse animation after 8s
	useEffect(() => {
		const timer = setTimeout(() => setHasInteracted(true), 8000);
		return () => clearTimeout(timer);
	}, []);

	// Close on outside click
	useEffect(() => {
		if (view === "closed") return;

		function handleClick(e: MouseEvent) {
			if (
				popupRef.current &&
				!popupRef.current.contains(e.target as Node)
			) {
				setView("closed");
			}
		}

		document.addEventListener("mousedown", handleClick);
		return () => document.removeEventListener("mousedown", handleClick);
	}, [view]);

	// Close on Escape
	useEffect(() => {
		if (view === "closed") return;

		function handleEscape(e: KeyboardEvent) {
			if (e.key === "Escape") setView("closed");
		}

		document.addEventListener("keydown", handleEscape);
		return () => document.removeEventListener("keydown", handleEscape);
	}, [view]);

	function handleFormSubmit(e: React.FormEvent<HTMLFormElement>) {
		e.preventDefault();
		setFormError(null);

		const form = e.currentTarget;
		const fd = new FormData(form);
		fd.set("preferencia", preference);
		fd.set("page_url", window.location.href);
		fd.set("referrer", document.referrer);

		startTransition(async () => {
			const result = await submitWidgetForm(fd);
			if (result.success) {
				setView("success");
			} else {
				setFormError(
					result.error ?? "Erro ao enviar. Tente novamente.",
				);
			}
		});
	}

	function handleToggle() {
		if (view === "closed") {
			setView("menu");
			setHasInteracted(true);
		} else {
			setView("closed");
			setTimeout(() => {
				setFormError(null);
			}, 300);
		}
	}

	return (
		<div
			ref={popupRef}
			className="fixed bottom-6 right-6 z-40 flex flex-col items-end gap-3"
		>
			<AnimatePresence mode="wait">
				{view !== "closed" && (
					<motion.div
						key="popup"
						initial={{ opacity: 0, scale: 0.92, y: 16 }}
						animate={{ opacity: 1, scale: 1, y: 0 }}
						exit={{ opacity: 0, scale: 0.92, y: 16 }}
						transition={{
							type: "spring",
							damping: 25,
							stiffness: 300,
						}}
						className="w-[340px] max-w-[calc(100vw-48px)] overflow-hidden rounded-2xl border border-black/10 bg-white/95 shadow-2xl shadow-black/10 backdrop-blur-xl dark:border-white/10 dark:bg-black/95 dark:shadow-white/5"
					>
						{/* Header */}
						<div className="flex items-center justify-between border-b border-black/5 bg-gradient-to-r from-emerald-500/10 via-emerald-400/5 to-transparent px-4 py-3 dark:border-white/5 dark:from-emerald-500/15">
							<div className="flex items-center gap-2.5">
								{view === "form" && (
									<button
										type="button"
										onClick={() => setView("menu")}
										className="rounded-full p-0.5 text-muted-foreground transition-colors hover:text-foreground"
										aria-label="Voltar"
									>
										<ArrowLeft className="size-4" />
									</button>
								)}
								<div className="flex size-8 items-center justify-center rounded-full bg-emerald-500/15 ring-1 ring-emerald-500/30">
									<MessageCircle className="size-4 text-emerald-600 dark:text-emerald-400" />
								</div>
								<div>
									<span className="text-sm font-semibold text-foreground">
										{view === "success"
											? "Recebemos!"
											: "Fale conosco"}
									</span>
									{view !== "success" && (
										<p className="text-[11px] text-muted-foreground">
											Resposta rápida e humana
										</p>
									)}
								</div>
							</div>
							<button
								type="button"
								onClick={() => setView("closed")}
								className="rounded-full p-1 text-muted-foreground transition-colors hover:text-foreground"
								aria-label="Fechar"
							>
								<X className="size-4" />
							</button>
						</div>

						{/* Body */}
						<div className="p-4">
							<AnimatePresence mode="wait">
								{view === "menu" && (
									<motion.div
										key="menu"
										initial={{ opacity: 0, x: -8 }}
										animate={{ opacity: 1, x: 0 }}
										exit={{ opacity: 0, x: 8 }}
										transition={{ duration: 0.15 }}
										className="flex flex-col gap-3"
									>
										<p className="text-sm text-muted-foreground">
											Como prefere falar conosco?
										</p>
										<a
											href={WHATSAPP_URL}
											target="_blank"
											rel="noopener noreferrer"
											className="group flex items-center gap-3 rounded-xl border border-emerald-500/30 bg-emerald-500/5 px-4 py-3.5 text-sm transition-all duration-150 hover:-translate-y-0.5 hover:border-emerald-500/50 hover:bg-emerald-500/10 hover:shadow-md"
										>
											<WhatsAppIcon className="size-5 shrink-0 text-emerald-600 dark:text-emerald-400" />
											<div className="flex-1">
												<p className="font-medium text-emerald-700 dark:text-emerald-400">
													Falar agora pelo WhatsApp
												</p>
												<p className="mt-0.5 text-xs text-emerald-600/70 dark:text-emerald-500/70">
													Resposta imediata
												</p>
											</div>
											<ArrowRight className="size-4 text-emerald-500/50 transition-transform duration-150 group-hover:translate-x-0.5 group-hover:text-emerald-500" />
										</a>
										<button
											type="button"
											onClick={() => setView("form")}
											className="group flex items-center gap-3 rounded-xl border border-black/10 bg-white px-4 py-3.5 text-left text-sm transition-all duration-150 hover:-translate-y-0.5 hover:shadow-md dark:border-white/10 dark:bg-white/5"
										>
											<MessageCircle className="size-5 shrink-0 text-muted-foreground" />
											<div className="flex-1">
												<p className="font-medium text-foreground">
													Solicitar que entremos em
													contato
												</p>
												<p className="mt-0.5 text-xs text-muted-foreground">
													Retorno em até 24h
												</p>
											</div>
											<ArrowRight className="size-4 text-muted-foreground/50 transition-transform duration-150 group-hover:translate-x-0.5 group-hover:text-muted-foreground" />
										</button>
									</motion.div>
								)}

								{view === "form" && (
									<motion.form
										key="form"
										initial={{ opacity: 0, x: 8 }}
										animate={{ opacity: 1, x: 0 }}
										exit={{ opacity: 0, x: -8 }}
										transition={{ duration: 0.15 }}
										onSubmit={handleFormSubmit}
										className="flex flex-col gap-3"
									>
										{formError && (
											<motion.p
												initial={{
													opacity: 0,
													y: -4,
												}}
												animate={{
													opacity: 1,
													y: 0,
												}}
												className="rounded-lg bg-destructive/10 px-3 py-2 text-xs text-destructive"
											>
												{formError}
											</motion.p>
										)}

										<Input
											name="nome"
											placeholder="Seu nome *"
											required
											minLength={2}
											className="h-10"
										/>

										{/* Contact preference toggle */}
										<div className="flex gap-2">
											<button
												type="button"
												onClick={() =>
													setPreference("whatsapp")
												}
												className={cn(
													"flex flex-1 items-center justify-center gap-1.5 rounded-lg border px-3 py-2 text-xs font-medium transition-all duration-150",
													preference === "whatsapp"
														? "border-emerald-500/50 bg-emerald-50 text-emerald-700 dark:bg-emerald-950/30 dark:text-emerald-400"
														: "border-border bg-background text-muted-foreground hover:bg-muted/50",
												)}
											>
												<WhatsAppIcon className="size-3.5" />
												WhatsApp
											</button>
											<button
												type="button"
												onClick={() =>
													setPreference("email")
												}
												className={cn(
													"flex flex-1 items-center justify-center gap-1.5 rounded-lg border px-3 py-2 text-xs font-medium transition-all duration-150",
													preference === "email"
														? "border-primary/50 bg-primary/5 text-foreground"
														: "border-border bg-background text-muted-foreground hover:bg-muted/50",
												)}
											>
												<Mail className="size-3.5" />
												E-mail
											</button>
										</div>

										<Input
											name="contato"
											placeholder={
												preference === "email"
													? "seu@email.com *"
													: "(21) 97153-2700 *"
											}
											type={
												preference === "email"
													? "email"
													: "tel"
											}
											required
											minLength={3}
											className="h-10"
										/>

										<Textarea
											name="mensagem"
											placeholder="Mensagem breve (opcional)"
											rows={2}
											className="resize-none"
										/>

										<Button
											type="submit"
											size="sm"
											disabled={isPending}
											className="gap-2 bg-emerald-600 text-white hover:bg-emerald-700"
										>
											{isPending ? (
												<>
													<Loader2 className="size-3.5 animate-spin" />
													Enviando...
												</>
											) : (
												<>
													<Send className="size-3.5" />
													Enviar
												</>
											)}
										</Button>

										<p className="text-center text-[10px] text-muted-foreground/60">
											Seus dados estão seguros e não serão
											compartilhados.
										</p>
									</motion.form>
								)}

								{view === "success" && (
									<motion.div
										key="success"
										initial={{ opacity: 0, scale: 0.95 }}
										animate={{ opacity: 1, scale: 1 }}
										transition={{
											duration: 0.25,
											ease: "easeOut",
										}}
										className="flex flex-col items-center gap-3 py-6 text-center"
									>
										<motion.div
											initial={{ scale: 0 }}
											animate={{ scale: 1 }}
											transition={{
												type: "spring",
												stiffness: 300,
												damping: 20,
												delay: 0.1,
											}}
											className="flex size-12 items-center justify-center rounded-full bg-emerald-500/10"
										>
											<CheckCircle2 className="size-6 text-emerald-500" />
										</motion.div>
										<div>
											<p className="text-sm font-medium text-foreground">
												Recebemos!
											</p>
											<p className="mt-1 text-xs leading-relaxed text-muted-foreground">
												Entraremos em contato em breve.
											</p>
										</div>
										<Button
											variant="ghost"
											size="sm"
											onClick={() => setView("closed")}
											className="mt-1 text-xs text-muted-foreground hover:text-foreground"
										>
											Fechar
										</Button>
									</motion.div>
								)}
							</AnimatePresence>
						</div>
					</motion.div>
				)}
			</AnimatePresence>

			{/* Floating button */}
			<motion.button
				type="button"
				onClick={handleToggle}
				className={cn(
					"relative flex size-14 items-center justify-center rounded-full shadow-lg transition-all duration-200",
					view === "closed"
						? "bg-emerald-500 text-white shadow-emerald-500/30 hover:bg-emerald-600 hover:shadow-xl hover:shadow-emerald-500/40"
						: "bg-black text-white shadow-black/20 hover:bg-black/80 dark:bg-white dark:text-black dark:shadow-white/20 dark:hover:bg-white/80",
				)}
				aria-label={
					view === "closed" ? "Abrir chat de contato" : "Fechar chat"
				}
				whileHover={{ scale: 1.05 }}
				whileTap={{ scale: 0.95 }}
			>
				<AnimatePresence mode="wait">
					{view === "closed" ? (
						<motion.div
							key="whatsapp"
							initial={{ rotate: -90, opacity: 0 }}
							animate={{ rotate: 0, opacity: 1 }}
							exit={{ rotate: 90, opacity: 0 }}
							transition={{ duration: 0.2 }}
						>
							<WhatsAppIcon className="size-6" />
						</motion.div>
					) : (
						<motion.div
							key="close"
							initial={{ rotate: 90, opacity: 0 }}
							animate={{ rotate: 0, opacity: 1 }}
							exit={{ rotate: -90, opacity: 0 }}
							transition={{ duration: 0.2 }}
						>
							<X className="size-6" />
						</motion.div>
					)}
				</AnimatePresence>
				{view === "closed" && !hasInteracted && (
					<>
						<span className="absolute inset-0 animate-ping rounded-full bg-emerald-400 opacity-20" />
						<span className="absolute inset-0 animate-pulse rounded-full bg-emerald-400 opacity-10" />
					</>
				)}
				{view === "closed" && (
					<span className="absolute -right-1 -top-1 flex size-5 items-center justify-center rounded-full bg-red-500 text-[10px] font-bold text-white shadow-sm">
						1
					</span>
				)}
			</motion.button>
		</div>
	);
}
