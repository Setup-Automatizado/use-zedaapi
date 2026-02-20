"use client";

import Link from "next/link";
import { motion } from "framer-motion";
import { ArrowRightIcon, ShieldCheckIcon } from "lucide-react";
import { Button } from "@/components/ui/button";

export function CTA() {
	return (
		<section id="contato" className="relative py-20 sm:py-28">
			<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				<motion.div
					initial={{ opacity: 0, y: 24 }}
					whileInView={{ opacity: 1, y: 0 }}
					viewport={{ once: true, margin: "-80px" }}
					transition={{ duration: 0.6, ease: [0.22, 1, 0.36, 1] }}
					className="relative overflow-hidden rounded-3xl border border-border/50 bg-gradient-to-b from-muted/40 to-muted/10 px-6 py-20 text-center sm:px-16 sm:py-24"
				>
					{/* Background decorative elements */}
					<div
						aria-hidden="true"
						className="pointer-events-none absolute inset-0 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] bg-[size:4rem_4rem] opacity-[0.15] [mask-image:radial-gradient(ellipse_60%_60%_at_50%_50%,black_20%,transparent_100%)]"
					/>

					{/* Primary glow */}
					<div
						aria-hidden="true"
						className="pointer-events-none absolute left-1/2 top-0 h-[400px] w-[600px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-primary/8 blur-[120px]"
					/>

					{/* Secondary glow */}
					<div
						aria-hidden="true"
						className="pointer-events-none absolute bottom-0 left-1/2 h-[300px] w-[500px] -translate-x-1/2 translate-y-1/2 rounded-full bg-primary/5 blur-[100px]"
					/>

					<div className="relative mx-auto max-w-2xl">
						<div className="mx-auto mb-6 flex size-14 items-center justify-center rounded-2xl bg-primary/10">
							<ShieldCheckIcon className="size-7 text-primary" />
						</div>

						<h2 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl">
							Comece a enviar em 5 minutos.{" "}
							<span className="text-primary">Teste grátis.</span>
						</h2>

						<p className="mt-5 text-base text-muted-foreground sm:text-lg leading-relaxed">
							+500 empresas já usam a ZedaAPI em produção. Crie
							sua conta, conecte pelo QR Code e comece a enviar.
						</p>

						<div className="mt-10 flex flex-col items-center justify-center gap-3 sm:flex-row sm:gap-4">
							<Button
								size="lg"
								asChild
								className="h-12 px-8 text-base shadow-lg shadow-primary/20 hover:shadow-xl hover:shadow-primary/25 transition-all duration-300"
							>
								<Link href="/sign-up">
									Criar Conta Grátis
									<ArrowRightIcon
										data-icon="inline-end"
										className="size-4"
									/>
								</Link>
							</Button>
							<Button
								variant="outline"
								size="lg"
								asChild
								className="h-12 px-8 text-base"
							>
								<Link href="/contato">
									Falar com Especialista
								</Link>
							</Button>
						</div>

						<div className="mt-8 flex flex-wrap items-center justify-center gap-x-6 gap-y-2 text-sm text-muted-foreground">
							<span className="flex items-center gap-1.5">
								<span className="size-1.5 rounded-full bg-emerald-500" />
								LGPD compliant
							</span>
							<span className="flex items-center gap-1.5">
								<span className="size-1.5 rounded-full bg-emerald-500" />
								Infraestrutura brasileira
							</span>
							<span className="flex items-center gap-1.5">
								<span className="size-1.5 rounded-full bg-emerald-500" />
								Sem cartão, sem contrato
							</span>
							<span className="flex items-center gap-1.5">
								<span className="size-1.5 rounded-full bg-emerald-500" />
								Suporte 100% em português
							</span>
						</div>
					</div>
				</motion.div>
			</div>
		</section>
	);
}
