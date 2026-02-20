"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { AnimatePresence, motion } from "framer-motion";
import {
	CookieIcon,
	ShieldCheckIcon,
	BarChart3Icon,
	MegaphoneIcon,
	SlidersHorizontalIcon,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Switch } from "@/components/ui/switch";
import {
	type CookiePreferences,
	COOKIE_CONSENT_VERSION,
	acceptAllCookies,
	getCookieConsent,
	rejectOptionalCookies,
	setCookieConsent,
} from "@/lib/cookies";

const CATEGORIES = [
	{
		id: "essential" as const,
		label: "Cookies Essenciais",
		description:
			"Necessários para o funcionamento do site. Incluem autenticação, segurança e preferências básicas.",
		icon: ShieldCheckIcon,
		locked: true,
	},
	{
		id: "analytics" as const,
		label: "Análise e Desempenho",
		description:
			"Nos ajudam a entender como você usa o site para melhorar a experiência.",
		icon: BarChart3Icon,
		locked: false,
	},
	{
		id: "marketing" as const,
		label: "Marketing",
		description:
			"Utilizados para exibir anúncios relevantes e medir a eficácia de campanhas.",
		icon: MegaphoneIcon,
		locked: false,
	},
	{
		id: "functionality" as const,
		label: "Funcionalidade",
		description:
			"Permitem recursos avançados e personalização da sua experiência.",
		icon: SlidersHorizontalIcon,
		locked: false,
	},
] as const;

export function CookieConsent() {
	const [visible, setVisible] = useState(false);
	const [modalOpen, setModalOpen] = useState(false);
	const [preferences, setPreferences] = useState({
		analytics: false,
		marketing: false,
		functionality: false,
	});

	useEffect(() => {
		const existing = getCookieConsent();
		if (existing) return;

		const timer = setTimeout(() => setVisible(true), 1500);
		return () => clearTimeout(timer);
	}, []);

	const dismiss = useCallback(() => {
		setVisible(false);
		setModalOpen(false);
	}, []);

	const handleAcceptAll = useCallback(() => {
		acceptAllCookies();
		dismiss();
	}, [dismiss]);

	const handleRejectOptional = useCallback(() => {
		rejectOptionalCookies();
		dismiss();
	}, [dismiss]);

	const handleSavePreferences = useCallback(() => {
		const prefs: CookiePreferences = {
			essential: true,
			analytics: preferences.analytics,
			marketing: preferences.marketing,
			functionality: preferences.functionality,
			timestamp: new Date().toISOString(),
			version: COOKIE_CONSENT_VERSION,
		};
		setCookieConsent(prefs);
		dismiss();
	}, [preferences, dismiss]);

	const handleOpenModal = useCallback(() => {
		const existing = getCookieConsent();
		if (existing) {
			setPreferences({
				analytics: existing.analytics,
				marketing: existing.marketing,
				functionality: existing.functionality,
			});
		}
		setModalOpen(true);
	}, []);

	// Listen for footer cookie button event
	useEffect(() => {
		function handleOpenCookiePreferences() {
			handleOpenModal();
		}
		window.addEventListener(
			"open-cookie-preferences",
			handleOpenCookiePreferences,
		);
		return () =>
			window.removeEventListener(
				"open-cookie-preferences",
				handleOpenCookiePreferences,
			);
	}, [handleOpenModal]);

	return (
		<>
			<AnimatePresence>
				{visible && (
					<motion.div
						initial={{ y: 100, opacity: 0 }}
						animate={{ y: 0, opacity: 1 }}
						exit={{ y: 100, opacity: 0 }}
						transition={{
							type: "spring",
							damping: 25,
							stiffness: 300,
						}}
						className="fixed inset-x-0 bottom-0 z-50 p-4"
					>
						<div className="mx-auto max-w-5xl rounded-2xl border border-border bg-background/95 p-5 shadow-2xl backdrop-blur-xl sm:p-6">
							<div className="flex flex-col items-start gap-4 sm:flex-row sm:items-center sm:justify-between">
								<div className="flex items-start gap-3 sm:items-center">
									<div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10">
										<CookieIcon className="size-4.5 text-primary" />
									</div>
									<p className="text-sm leading-relaxed text-muted-foreground">
										Utilizamos cookies para melhorar sua
										experiência. Ao continuar navegando,
										você concorda com nossa{" "}
										<Link
											href="/politica-de-cookies"
											className="font-medium text-foreground underline underline-offset-4 transition-colors hover:text-primary"
										>
											Política de Cookies
										</Link>
										.
									</p>
								</div>
								<div className="flex w-full shrink-0 gap-2 sm:w-auto">
									<Button
										variant="outline"
										size="sm"
										onClick={handleOpenModal}
										className="flex-1 rounded-4xl sm:flex-none"
									>
										Gerenciar
									</Button>
									<Button
										size="sm"
										onClick={handleAcceptAll}
										className="flex-1 rounded-4xl sm:flex-none"
									>
										Aceitar todos
									</Button>
								</div>
							</div>
						</div>
					</motion.div>
				)}
			</AnimatePresence>

			<Dialog open={modalOpen} onOpenChange={setModalOpen}>
				<DialogContent className="sm:max-w-lg">
					<DialogHeader>
						<DialogTitle>Preferências de Cookies</DialogTitle>
						<DialogDescription>
							Gerencie como utilizamos cookies no seu navegador.
							Você pode alterar suas preferências a qualquer
							momento.
						</DialogDescription>
					</DialogHeader>

					<div className="flex flex-col gap-3">
						{CATEGORIES.map((cat) => {
							const Icon = cat.icon;
							return (
								<div
									key={cat.id}
									className="flex items-start justify-between gap-4 rounded-xl border border-border p-4 transition-colors hover:bg-muted/30"
								>
									<div className="flex items-start gap-3">
										<div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted">
											<Icon className="size-4 text-muted-foreground" />
										</div>
										<div className="flex-1">
											<p className="text-sm font-medium text-foreground">
												{cat.label}
												{cat.locked && (
													<span className="ml-2 inline-block rounded-md bg-muted px-1.5 py-0.5 text-[10px] font-normal text-muted-foreground">
														Obrigatório
													</span>
												)}
											</p>
											<p className="mt-1 text-xs leading-relaxed text-muted-foreground">
												{cat.description}
											</p>
										</div>
									</div>
									<Switch
										checked={
											cat.locked
												? true
												: preferences[
														cat.id as keyof typeof preferences
													]
										}
										disabled={cat.locked}
										onCheckedChange={
											cat.locked
												? undefined
												: (checked: boolean) =>
														setPreferences(
															(prev) => ({
																...prev,
																[cat.id]:
																	checked,
															}),
														)
										}
										size="sm"
									/>
								</div>
							);
						})}
					</div>

					<DialogFooter>
						<Button
							variant="outline"
							onClick={handleRejectOptional}
							className="rounded-4xl"
						>
							Rejeitar opcionais
						</Button>
						<Button
							onClick={handleSavePreferences}
							className="rounded-4xl"
						>
							Salvar preferências
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</>
	);
}
