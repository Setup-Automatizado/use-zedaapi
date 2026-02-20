"use client";

import { useMemo, useState, type CSSProperties } from "react";
import Link from "next/link";
import { motion, AnimatePresence } from "framer-motion";
import NumberFlow from "@number-flow/react";
import {
	Check,
	Minus,
	Plus,
	ArrowRightIcon,
	Building2,
	Clock,
	Headphones,
	TrendingDown,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
	detectPlanTierByInstanceCount,
	getPriceBreakdown,
	getDisplayFeatures,
	getPlanCheckpoints,
	MIN_INSTANCES,
	MAX_INSTANCES,
	type PlanSlug,
} from "@/lib/billing/plan-config";

const SLIDER_RANGE = MAX_INSTANCES - MIN_INSTANCES;

function getSavingsMessage(instanceCount: number): string | null {
	if (instanceCount >= 100) return "Até 28% mais barato por instância";
	if (instanceCount >= 10) return "Até 85% mais barato por instância";
	if (instanceCount >= 2) return "Preço imbatível para começar";
	return null;
}

export function Pricing() {
	const [instanceCount, setInstanceCount] = useState(10);
	const [annual, setAnnual] = useState(false);

	const tier = useMemo(
		() => detectPlanTierByInstanceCount(instanceCount),
		[instanceCount],
	);
	const breakdown = useMemo(
		() => getPriceBreakdown(instanceCount),
		[instanceCount],
	);
	const features = useMemo(
		() => getDisplayFeatures(tier.slug as PlanSlug),
		[tier.slug],
	);
	const checkpoints = useMemo(() => getPlanCheckpoints(), []);

	const displayPrice = annual
		? Math.round(breakdown.price * 0.8)
		: breakdown.price;
	const annualSavings = annual ? breakdown.price * 12 - displayPrice * 12 : 0;
	const pricePerInstance =
		Math.round((displayPrice / instanceCount) * 100) / 100;

	const sliderProgress =
		((instanceCount - MIN_INSTANCES) / SLIDER_RANGE) * 100;

	const sliderStyle = useMemo(
		() =>
			({
				"--slider-progress": `${sliderProgress}%`,
			}) as CSSProperties,
		[sliderProgress],
	);

	const savingsMessage = getSavingsMessage(instanceCount);

	return (
		<section id="precos" className="relative py-20 sm:py-28">
			<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				{/* Header */}
				<motion.div
					initial={{ opacity: 0, y: 20 }}
					whileInView={{ opacity: 1, y: 0 }}
					viewport={{ once: true, margin: "-80px" }}
					transition={{ duration: 0.5 }}
					className="mx-auto max-w-2xl text-center"
				>
					<p className="text-sm font-semibold uppercase tracking-widest text-primary">
						Preços
					</p>
					<h2 className="mt-3 text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl">
						Preço fixo por faixa.{" "}
						<span className="text-primary">
							Sem taxa por mensagem.
						</span>
					</h2>
					<p className="mt-4 text-base leading-relaxed text-muted-foreground sm:text-lg">
						Ajuste o número de instâncias e pague pela faixa. Todas
						as funcionalidades incluídas.
					</p>

					{/* Monthly/Annual Toggle */}
					<div
						className="mt-8 inline-flex items-center gap-3 rounded-4xl border border-border bg-muted/50 p-1"
						role="radiogroup"
						aria-label="Alternar entre mensal e anual"
					>
						<button
							type="button"
							role="radio"
							aria-checked={!annual}
							onClick={() => setAnnual(false)}
							className={cn(
								"rounded-4xl px-4 py-2 text-sm font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background",
								!annual
									? "bg-background text-foreground shadow-sm"
									: "text-muted-foreground hover:text-foreground",
							)}
						>
							Mensal
						</button>
						<button
							type="button"
							role="radio"
							aria-checked={annual}
							onClick={() => setAnnual(true)}
							className={cn(
								"flex items-center gap-2 rounded-4xl px-4 py-2 text-sm font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background",
								annual
									? "bg-background text-foreground shadow-sm"
									: "text-muted-foreground hover:text-foreground",
							)}
						>
							Anual
							<Badge
								variant="default"
								className="px-1.5 text-[10px]"
							>
								-20%
							</Badge>
						</button>
					</div>
				</motion.div>

				{/* Main Card */}
				<motion.div
					initial={{ opacity: 0, y: 30 }}
					whileInView={{ opacity: 1, y: 0 }}
					viewport={{ once: true, margin: "-60px" }}
					transition={{ duration: 0.6, delay: 0.1 }}
					className="mx-auto mt-14 max-w-5xl"
				>
					<div className="rounded-2xl border border-border bg-card p-6 shadow-lg shadow-primary/[0.03] sm:p-10">
						<div className="grid grid-cols-1 gap-10 lg:grid-cols-2">
							{/* LEFT COLUMN: Slider + Price */}
							<div className="space-y-6">
								{/* Label + Tier Badge */}
								<div className="flex items-center justify-between">
									<div className="flex items-center gap-2">
										<span className="text-sm font-medium text-foreground">
											Número de Instâncias WhatsApp
										</span>
										<AnimatePresence mode="wait">
											<motion.span
												key={tier.slug}
												initial={{
													opacity: 0,
													scale: 0.9,
												}}
												animate={{
													opacity: 1,
													scale: 1,
												}}
												exit={{
													opacity: 0,
													scale: 0.9,
												}}
												transition={{ duration: 0.2 }}
												className="inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-semibold text-primary"
											>
												{tier.name}
											</motion.span>
										</AnimatePresence>
									</div>
								</div>

								{/* Increment/Decrement + NumberFlow */}
								<div className="flex items-center justify-between">
									<div className="flex items-center gap-3">
										<Button
											variant="outline"
											size="icon"
											className="size-8 rounded-lg"
											disabled={
												instanceCount <= MIN_INSTANCES
											}
											onClick={() =>
												setInstanceCount((c) =>
													Math.max(
														MIN_INSTANCES,
														c - 1,
													),
												)
											}
											aria-label="Diminuir instâncias"
										>
											<Minus className="size-3.5" />
										</Button>
										<div className="w-16 text-center">
											<NumberFlow
												value={instanceCount}
												className="text-xl font-semibold text-foreground"
												format={{
													minimumIntegerDigits: 1,
												}}
											/>
										</div>
										<Button
											variant="outline"
											size="icon"
											className="size-8 rounded-lg"
											disabled={
												instanceCount >= MAX_INSTANCES
											}
											onClick={() =>
												setInstanceCount((c) =>
													Math.min(
														MAX_INSTANCES,
														c + 1,
													),
												)
											}
											aria-label="Aumentar instâncias"
										>
											<Plus className="size-3.5" />
										</Button>
									</div>
								</div>

								{/* Tier Checkpoints - Desktop */}
								<div className="relative pt-8 sm:pt-10">
									<div className="absolute -top-0 left-0 right-0 hidden sm:block">
										<div className="flex items-end justify-between px-0.5">
											{checkpoints.map((cp) => {
												const isActive =
													tier.slug === cp.planSlug;
												return (
													<div
														key={cp.planSlug}
														className="flex flex-col items-center"
													>
														<span
															className={cn(
																"whitespace-nowrap rounded-full px-2 py-0.5 text-[10px] font-semibold transition-all",
																isActive
																	? "bg-primary/15 text-primary"
																	: "bg-muted text-muted-foreground",
															)}
														>
															{cp.range}
														</span>
														<div
															className={cn(
																"mt-1 h-2 w-px transition-colors",
																isActive
																	? "bg-primary/50"
																	: "bg-border",
															)}
														/>
													</div>
												);
											})}
										</div>
									</div>

									{/* Range Slider */}
									<input
										type="range"
										min={MIN_INSTANCES}
										max={MAX_INSTANCES}
										value={instanceCount}
										onChange={(e) =>
											setInstanceCount(
												Number.parseInt(
													e.target.value,
													10,
												),
											)
										}
										style={sliderStyle}
										className="range-slider h-1.5 w-full cursor-pointer appearance-none rounded-lg bg-muted transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background [&::-moz-range-thumb]:size-4 [&::-moz-range-thumb]:appearance-none [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:border-2 [&::-moz-range-thumb]:border-primary [&::-moz-range-thumb]:bg-background [&::-moz-range-thumb]:shadow-sm [&::-moz-range-thumb]:transition-transform [&::-moz-range-thumb]:hover:scale-110 [&::-moz-range-track]:rounded-lg [&::-webkit-slider-runnable-track]:rounded-lg [&::-webkit-slider-thumb]:size-4 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:border-2 [&::-webkit-slider-thumb]:border-primary [&::-webkit-slider-thumb]:bg-background [&::-webkit-slider-thumb]:shadow-sm [&::-webkit-slider-thumb]:transition-transform [&::-webkit-slider-thumb]:hover:scale-110"
										aria-label="Número de instâncias"
									/>
									<div className="mt-1.5 flex justify-between">
										<span className="text-xs text-muted-foreground">
											{MIN_INSTANCES} instância
										</span>
										<span className="text-xs text-muted-foreground">
											{MAX_INSTANCES} instâncias
										</span>
									</div>

									{/* Tier Checkpoints - Mobile */}
									<div className="mt-3 flex flex-wrap items-center justify-center gap-1.5 sm:hidden">
										{checkpoints.map((cp) => {
											const isActive =
												tier.slug === cp.planSlug;
											return (
												<span
													key={cp.planSlug}
													className={cn(
														"rounded-full px-2 py-0.5 text-[9px] font-semibold transition-all",
														isActive
															? "bg-primary/15 text-primary"
															: "bg-muted text-muted-foreground",
													)}
												>
													{cp.range} {cp.planName}
												</span>
											);
										})}
									</div>
								</div>

								{/* Divider */}
								<div className="border-t border-border" />

								{/* Price Display */}
								<div className="space-y-2">
									<div className="flex items-end gap-1.5">
										<NumberFlow
											value={displayPrice}
											format={{
												style: "currency",
												currency: "BRL",
												minimumFractionDigits: 0,
												maximumFractionDigits: 0,
											}}
											className="text-5xl font-bold tracking-tight text-foreground sm:text-6xl"
										/>
										<span className="mb-2 text-base text-muted-foreground">
											/mês
										</span>
									</div>

									<div className="flex items-center gap-3">
										<span className="text-sm text-muted-foreground">
											<NumberFlow
												value={pricePerInstance}
												format={{
													style: "currency",
													currency: "BRL",
													minimumFractionDigits: 2,
													maximumFractionDigits: 2,
												}}
												className="font-medium text-foreground"
											/>{" "}
											por instância
										</span>
									</div>

									{/* Savings Badge */}
									<AnimatePresence mode="wait">
										{savingsMessage && (
											<motion.div
												key={savingsMessage}
												initial={{ opacity: 0, y: 4 }}
												animate={{ opacity: 1, y: 0 }}
												exit={{ opacity: 0, y: -4 }}
												transition={{ duration: 0.2 }}
												className="flex items-center gap-1.5"
											>
												<TrendingDown className="size-3.5 text-primary" />
												<span className="text-xs font-medium text-primary">
													{savingsMessage}
												</span>
											</motion.div>
										)}
									</AnimatePresence>

									{/* Annual Savings */}
									<AnimatePresence>
										{annual && annualSavings > 0 && (
											<motion.div
												initial={{
													opacity: 0,
													height: 0,
												}}
												animate={{
													opacity: 1,
													height: "auto",
												}}
												exit={{ opacity: 0, height: 0 }}
												transition={{ duration: 0.2 }}
											>
												<span className="text-xs font-medium text-primary">
													Economize R${annualSavings}
													/ano com o plano anual
												</span>
											</motion.div>
										)}
									</AnimatePresence>

									<p className="text-xs text-muted-foreground">
										Cobrança {annual ? "anual" : "mensal"},
										cancele quando quiser, sem multa.
									</p>
								</div>
							</div>

							{/* RIGHT COLUMN: Features + CTA */}
							<div className="flex flex-col space-y-5">
								<div>
									<h4 className="mb-4 text-sm font-medium text-foreground">
										Tudo incluído:
									</h4>
									<div className="grid grid-cols-1 gap-2.5">
										<AnimatePresence mode="wait">
											<motion.div
												key={tier.slug}
												initial={{ opacity: 0 }}
												animate={{ opacity: 1 }}
												exit={{ opacity: 0 }}
												transition={{ duration: 0.2 }}
												className="grid grid-cols-1 gap-2.5"
											>
												{features.map((feature, i) => (
													<motion.div
														key={feature}
														initial={{
															opacity: 0,
															x: -5,
														}}
														animate={{
															opacity: 1,
															x: 0,
														}}
														transition={{
															delay:
																0.05 +
																i * 0.025,
															duration: 0.25,
														}}
														className="flex items-start gap-2.5"
													>
														<div className="mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-full bg-primary/10">
															<Check className="size-2.5 text-primary" />
														</div>
														<span className="text-sm leading-tight text-muted-foreground">
															{feature}
														</span>
													</motion.div>
												))}
											</motion.div>
										</AnimatePresence>
									</div>
								</div>

								{/* Divider */}
								<div className="border-t border-border" />

								{/* Stats Row */}
								<div className="grid grid-cols-3 gap-4 text-center">
									<div>
										<div className="flex items-center justify-center gap-1 text-foreground">
											<Building2 className="size-3.5 text-primary" />
											<span className="text-lg font-bold tracking-tight sm:text-xl">
												500+
											</span>
										</div>
										<p className="text-[11px] text-muted-foreground">
											Empresas ativas
										</p>
									</div>
									<div>
										<div className="flex items-center justify-center gap-1 text-foreground">
											<Clock className="size-3.5 text-primary" />
											<span className="text-lg font-bold tracking-tight sm:text-xl">
												99.9%
											</span>
										</div>
										<p className="text-[11px] text-muted-foreground">
											Uptime garantido
										</p>
									</div>
									<div>
										<div className="flex items-center justify-center gap-1 text-foreground">
											<Headphones className="size-3.5 text-primary" />
											<span className="text-lg font-bold tracking-tight sm:text-xl">
												24/7
											</span>
										</div>
										<p className="text-[11px] text-muted-foreground">
											Suporte técnico
										</p>
									</div>
								</div>

								{/* CTA Button */}
								<div className="mt-auto pt-2">
									<Button
										size="lg"
										asChild
										className="w-full shadow-lg shadow-primary/20"
									>
										<Link href="/sign-up">
											Testar 7 Dias Grátis
											<ArrowRightIcon
												className="size-4"
												data-icon="inline-end"
											/>
										</Link>
									</Button>
								</div>
							</div>
						</div>
					</div>
				</motion.div>

				{/* Bottom Notes */}
				<motion.div
					initial={{ opacity: 0 }}
					whileInView={{ opacity: 1 }}
					viewport={{ once: true }}
					transition={{ duration: 0.5, delay: 0.3 }}
					className="mt-12 flex flex-col items-center gap-2 text-center"
				>
					<p className="text-sm text-muted-foreground">
						Todos os planos incluem{" "}
						<span className="font-medium text-foreground">
							7 dias de teste gratuito
						</span>
						. Sem cartão de crédito. Cancele quando quiser.
					</p>
					<p className="text-xs text-muted-foreground">
						Aceitamos cartão de crédito internacional (Stripe) e
						PIX/Boleto para clientes brasileiros.
					</p>
				</motion.div>
			</div>
		</section>
	);
}
