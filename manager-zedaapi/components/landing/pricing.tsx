"use client";

import { motion } from "framer-motion";
import { ArrowRightIcon } from "lucide-react";

import { PricingCard } from "@/components/shared/pricing-card";

export function Pricing() {
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
				</motion.div>

				{/* Main Card */}
				<motion.div
					initial={{ opacity: 0, y: 30 }}
					whileInView={{ opacity: 1, y: 0 }}
					viewport={{ once: true, margin: "-60px" }}
					transition={{ duration: 0.6, delay: 0.1 }}
					className="mx-auto mt-14 max-w-5xl"
				>
					<PricingCard
						showAnnualToggle
						ctaHref="/cadastro"
						ctaContent={
							<>
								Testar 7 Dias Grátis
								<ArrowRightIcon
									className="size-4"
									data-icon="inline-end"
								/>
							</>
						}
					/>
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
