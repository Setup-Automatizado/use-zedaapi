"use client";

import NumberFlow from "@number-flow/react";
import { Check, Minus, Plus } from "lucide-react";
import { motion } from "framer-motion";
import { useMemo, useState, type CSSProperties } from "react";
import { Button } from "@/components/ui/button";
import {
	getCurrentPlanInfo,
	getPriceBreakdown,
	getDisplayFeatures,
	PLAN_TIERS,
	MIN_INSTANCES,
	MAX_INSTANCES,
	type PlanSlug,
} from "@/lib/billing/plan-config";

const SLIDER_RANGE = MAX_INSTANCES - MIN_INSTANCES;
const PERCENTAGE_MULTIPLIER = 100;
const ANIMATION_DELAY_BASE = 0.08;
const ANIMATION_DELAY_INCREMENT = 0.025;
const ANIMATION_DURATION = 0.25;

const DEFAULT_INSTANCES = 10;

/**
 * Badge positions — visually balanced across the slider
 * (hardcoded for readability, not linear %)
 */
const BADGE_POSITIONS = {
	starter: "3%",
	pro: "12%",
	business: "25%",
	scale: "42%",
	enterprise: "65%",
	ultimate: "88%",
} as const;

interface Plan {
	id: string;
	name: string;
	slug: string;
	description: string | null;
	price: number | string;
	currency: string;
	interval: string;
	maxInstances: number;
	features: unknown;
	active: boolean;
}

interface PlanComparisonProps {
	plans: Plan[];
	currentPlanSlug?: string | null;
	onSelectPlan: (planId: string) => void;
	loading?: boolean;
	loadingPlanId?: string | null;
}

export function PlanComparison({
	plans,
	currentPlanSlug,
	onSelectPlan,
	loading = false,
}: PlanComparisonProps) {
	const [instanceCount, setInstanceCount] = useState(DEFAULT_INSTANCES);

	const currentPlanInfo = useMemo(
		() => getCurrentPlanInfo(instanceCount),
		[instanceCount],
	);

	const priceBreakdown = useMemo(
		() => getPriceBreakdown(instanceCount),
		[instanceCount],
	);

	const features = useMemo(() => {
		return getDisplayFeatures(currentPlanInfo.slug as PlanSlug);
	}, [currentPlanInfo.slug]);

	const { price: totalPrice, pricePerInstance } = priceBreakdown;

	const handleIncrement = () => {
		if (instanceCount < MAX_INSTANCES) {
			setInstanceCount(instanceCount + 1);
		}
	};

	const handleDecrement = () => {
		if (instanceCount > MIN_INSTANCES) {
			setInstanceCount(instanceCount - 1);
		}
	};

	const handleSelectCurrentPlan = () => {
		const matchingPlan = plans.find((p) => p.slug === currentPlanInfo.slug);
		if (matchingPlan) {
			onSelectPlan(matchingPlan.id);
		}
	};

	const sliderProgress =
		((instanceCount - MIN_INSTANCES) / SLIDER_RANGE) *
		PERCENTAGE_MULTIPLIER;

	const sliderStyle = useMemo(
		() =>
			({
				["--slider-progress" as string]: `${sliderProgress}%`,
			}) as CSSProperties,
		[sliderProgress],
	);

	const isCurrentPlan = currentPlanSlug === currentPlanInfo.slug;

	return (
		<div className="relative mx-auto max-w-6xl">
			<div className="rounded-2xl border border-border bg-muted/30 p-6 sm:p-10">
				<div className="grid grid-cols-1 gap-10 lg:grid-cols-2">
					{/* Left — Slider + Price */}
					<div className="space-y-6">
						<div>
							<div className="mb-1.5 flex items-center gap-2">
								<span className="text-xs font-medium tracking-tight text-primary">
									Pricing
								</span>
							</div>
							<h3 className="mb-1.5 text-2xl font-medium tracking-tight text-foreground sm:text-3xl">
								Escale conforme sua necessidade.
							</h3>
							<p className="text-sm tracking-tight text-muted-foreground">
								Escolha o numero de instancias WhatsApp e ajuste
								quando quiser. Sem compromissos.
							</p>
						</div>

						{/* Slider controls */}
						<div className="space-y-3">
							<div className="flex items-center justify-between">
								<div className="flex items-center gap-2">
									<span className="text-sm font-medium tracking-tight text-foreground">
										Instancias WhatsApp
									</span>
									<span className="inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-semibold text-primary">
										Plano {currentPlanInfo.name}
									</span>
								</div>
								<div className="flex items-center gap-2.5">
									<Button
										className="h-7 w-7"
										disabled={
											instanceCount <= MIN_INSTANCES
										}
										onClick={handleDecrement}
										size="icon"
										variant="outline"
									>
										<Minus className="h-3.5 w-3.5" />
									</Button>
									<div className="w-14 text-center">
										<NumberFlow
											className="text-xl font-semibold tracking-tight text-foreground"
											format={{
												minimumIntegerDigits: 1,
											}}
											value={instanceCount}
										/>
									</div>
									<Button
										className="h-7 w-7"
										disabled={
											instanceCount >= MAX_INSTANCES
										}
										onClick={handleIncrement}
										size="icon"
										variant="outline"
									>
										<Plus className="h-3.5 w-3.5" />
									</Button>
								</div>
							</div>

							{/* Slider + badges */}
							<div className="relative pt-1 sm:pt-10">
								{/* Desktop badges */}
								<div className="absolute -top-1 left-0 right-0 hidden sm:block">
									{PLAN_TIERS.map((tier) => {
										const isActive =
											instanceCount >=
												tier.minInstances &&
											instanceCount <= tier.maxInstances;
										const position =
											BADGE_POSITIONS[
												tier.slug as keyof typeof BADGE_POSITIONS
											];

										return (
											<div
												key={tier.slug}
												className="absolute flex flex-col items-center"
												style={{
													left: position,
													transform:
														"translateX(-50%)",
												}}
											>
												<span
													className={`whitespace-nowrap rounded-full px-2 py-0.5 text-[10px] font-semibold transition-all md:px-3 md:py-1 md:text-xs ${
														isActive
															? "bg-primary/15 text-primary"
															: "bg-muted text-muted-foreground/40"
													}`}
												>
													{tier.minInstances ===
													tier.maxInstances
														? tier.minInstances
														: `${tier.minInstances}-${tier.maxInstances}`}{" "}
													{tier.name}
												</span>
												<div
													className={`mt-1 h-2 w-px transition-colors md:mt-1.5 md:h-3 ${
														isActive
															? "bg-primary/50"
															: "bg-border"
													}`}
												/>
											</div>
										);
									})}
								</div>

								<input
									aria-label="Quantidade de instancias"
									className="range-slider h-1.5 w-full cursor-pointer rounded-lg accent-primary [&::-moz-range-thumb]:h-3.5 [&::-moz-range-thumb]:w-3.5 [&::-moz-range-thumb]:appearance-none [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:border-none [&::-moz-range-thumb]:bg-primary [&::-webkit-slider-thumb]:h-3.5 [&::-webkit-slider-thumb]:w-3.5 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-primary"
									max={MAX_INSTANCES}
									min={MIN_INSTANCES}
									onChange={(e) =>
										setInstanceCount(
											Number.parseInt(e.target.value, 10),
										)
									}
									style={sliderStyle}
									type="range"
									value={instanceCount}
								/>
								<div className="mt-1.5 flex justify-between">
									<span className="text-xs tracking-tight text-muted-foreground/60">
										{MIN_INSTANCES} instancia
									</span>
									<span className="text-xs tracking-tight text-muted-foreground/60">
										{MAX_INSTANCES} instancias
									</span>
								</div>

								{/* Mobile badges */}
								<div className="mt-3 flex flex-wrap items-center justify-center gap-1.5 sm:hidden">
									{PLAN_TIERS.map((tier) => {
										const isActive =
											instanceCount >=
												tier.minInstances &&
											instanceCount <= tier.maxInstances;
										return (
											<span
												key={tier.slug}
												className={`rounded-full px-2 py-0.5 text-[9px] font-semibold transition-all ${
													isActive
														? "bg-primary/15 text-primary"
														: "bg-muted text-muted-foreground/40"
												}`}
											>
												{tier.minInstances ===
												tier.maxInstances
													? tier.minInstances
													: `${tier.minInstances}-${tier.maxInstances}`}{" "}
												{tier.name}
											</span>
										);
									})}
								</div>
							</div>
						</div>

						{/* Price display */}
						<div className="border-t border-border pt-4">
							<div className="mb-3 flex items-end justify-between gap-4">
								<div>
									<div className="flex items-end gap-1.5">
										<NumberFlow
											className="text-5xl font-semibold tracking-tighter text-foreground sm:text-6xl"
											format={{
												minimumIntegerDigits: 1,
											}}
											prefix="R$"
											value={totalPrice}
										/>
										<span className="mb-1.5 text-base tracking-tight text-muted-foreground">
											/mes
										</span>
									</div>
									<div className="mt-1 flex items-center gap-1.5 text-sm text-muted-foreground">
										<span>
											R$
											<NumberFlow
												className="inline font-medium text-foreground"
												value={pricePerInstance}
												format={{
													minimumFractionDigits: 2,
													maximumFractionDigits: 2,
												}}
											/>
											/instancia
										</span>
										<span className="text-xs">
											({instanceCount}{" "}
											{instanceCount === 1
												? "instancia"
												: "instancias"}
											)
										</span>
									</div>
								</div>

								<Button
									className="flex-shrink-0"
									size="lg"
									disabled={isCurrentPlan || loading}
									onClick={handleSelectCurrentPlan}
								>
									{loading
										? "Processando..."
										: isCurrentPlan
											? "Plano atual"
											: "Assinar agora"}
								</Button>
							</div>
							<p className="text-xs tracking-tight text-muted-foreground/60">
								Cobranca mensal, cancele quando quiser.
							</p>
						</div>
					</div>

					{/* Right — Features */}
					<div className="space-y-4">
						<div>
							<h4 className="mb-3 text-sm font-medium tracking-tight text-foreground">
								Tudo incluido no{" "}
								<span className="text-primary">
									{currentPlanInfo.name}
								</span>
								:
							</h4>
						</div>
						<div className="grid grid-cols-1 gap-2.5">
							{features.map((feature, index) => (
								<motion.div
									animate={{ opacity: 1, x: 0 }}
									className="flex items-start gap-2.5"
									initial={{ opacity: 0, x: -5 }}
									key={feature}
									transition={{
										delay:
											ANIMATION_DELAY_BASE +
											index * ANIMATION_DELAY_INCREMENT,
										duration: ANIMATION_DURATION,
									}}
								>
									<div className="mt-0.5 flex-shrink-0">
										<div className="flex h-3.5 w-3.5 items-center justify-center rounded-full bg-primary/10">
											<Check className="h-2.5 w-2.5 text-primary" />
										</div>
									</div>
									<span className="text-sm leading-tight tracking-tight text-foreground/70">
										{feature}
									</span>
								</motion.div>
							))}
						</div>

						{/* Footnote */}
						<p className="text-[10px] text-muted-foreground/50">
							* Botoes interativos disponiveis ate quando houver
							suporte oficial do WhatsApp.
						</p>

						{/* Stats */}
						<div className="mt-6 border-t border-border pt-6">
							<div className="grid grid-cols-3 gap-6 text-center">
								<div>
									<div className="mb-1 text-3xl font-semibold tracking-tighter text-foreground sm:text-4xl">
										10K+
									</div>
									<div className="text-xs tracking-tight text-muted-foreground">
										Mensagens/dia
									</div>
								</div>
								<div>
									<div className="mb-1 text-3xl font-semibold tracking-tighter text-foreground sm:text-4xl">
										99.9%
									</div>
									<div className="text-xs tracking-tight text-muted-foreground">
										Uptime garantido
									</div>
								</div>
								<div>
									<div className="mb-1 text-3xl font-semibold tracking-tighter text-foreground sm:text-4xl">
										500+
									</div>
									<div className="text-xs tracking-tight text-muted-foreground">
										Instancias ativas
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>

			{/* Bottom info strip */}
			<div className="relative mt-10">
				<div className="mx-auto grid w-full max-w-6xl grid-cols-1 gap-6 text-center md:grid-cols-3">
					<div className="space-y-1.5">
						<p className="font-medium tracking-tight text-foreground">
							Suporte dedicado
						</p>
						<p className="text-sm tracking-tight text-muted-foreground">
							Equipe disponivel para ajudar na integracao e
							operacao da sua API.
						</p>
					</div>
					<div className="space-y-1.5">
						<p className="font-medium tracking-tight text-foreground">
							Setup gratuito
						</p>
						<p className="text-sm tracking-tight text-muted-foreground">
							Configure suas instancias em minutos com nossa
							documentacao completa.
						</p>
					</div>
					<div className="space-y-1.5">
						<p className="font-medium tracking-tight text-foreground">
							Sempre atualizado
						</p>
						<p className="text-sm tracking-tight text-muted-foreground">
							Melhorias continuas e novos recursos sem custo
							adicional.
						</p>
					</div>
				</div>
			</div>
		</div>
	);
}
