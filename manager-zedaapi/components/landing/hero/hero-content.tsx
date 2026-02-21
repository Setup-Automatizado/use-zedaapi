"use client";

import Link from "next/link";
import { motion } from "framer-motion";
import { ArrowRightIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { fadeUp } from "@/lib/design-tokens";

interface HeroContentProps {
	reducedMotion: boolean;
}

const TRUST_SIGNALS = [
	"Integração em 5 minutos",
	"7 dias grátis, sem cartão",
	"99.9% de uptime garantido",
	"+500 empresas em produção",
] as const;

export function HeroContent({ reducedMotion }: HeroContentProps) {
	const Wrapper = reducedMotion ? "div" : motion.div;

	function motionProps(delay: number) {
		if (reducedMotion) return {};
		return {
			custom: delay,
			variants: fadeUp,
			initial: "hidden" as const,
			animate: "visible" as const,
		};
	}

	return (
		<div className="flex flex-col items-center text-center">
			{/* Badge */}
			<Wrapper {...motionProps(0)}>
				<Badge
					variant="outline"
					className="mb-7 gap-2 px-4 py-1.5 text-sm font-medium border-primary/20 bg-primary/5 text-primary"
				>
					<svg
						viewBox="0 0 24 24"
						fill="currentColor"
						className="size-4"
						aria-hidden="true"
					>
						<path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z" />
					</svg>
					Preço fixo — sem cobrança por mensagem
				</Badge>
			</Wrapper>

			{/* Headline */}
			<Wrapper {...motionProps(1)}>
				<h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl lg:text-6xl">
					Uma chamada API.
					<br />
					<span className="text-primary">Mensagem entregue.</span>
				</h1>
			</Wrapper>

			{/* Subtitle */}
			<Wrapper {...motionProps(2)}>
				<p className="mt-6 max-w-xl text-lg text-muted-foreground leading-relaxed">
					Envie texto, imagem, áudio e documentos pelo WhatsApp com
					uma API REST. 47ms de latência, webhooks em tempo real e
					preço fixo mensal — sem surpresa na fatura.
				</p>
			</Wrapper>

			{/* CTAs */}
			<Wrapper {...motionProps(3)}>
				<div className="mt-10 flex flex-col items-center gap-3 sm:flex-row sm:gap-4">
					<Button
						size="lg"
						asChild
						className="h-12 px-7 text-base shadow-lg shadow-primary/20 hover:shadow-xl hover:shadow-primary/25 transition-all duration-300"
					>
						<Link href="/sign-up">
							Criar Conta Grátis
							<ArrowRightIcon
								className="size-4"
								data-icon="inline-end"
							/>
						</Link>
					</Button>
					<Button
						variant="outline"
						size="lg"
						asChild
						className="h-12 px-7 text-base"
					>
						<a
							href="https://api.zedaapi.com/docs"
							target="_blank"
							rel="noopener noreferrer"
						>
							Ver Documentação
						</a>
					</Button>
				</div>
			</Wrapper>

			{/* Trust signals */}
			<Wrapper {...motionProps(4)}>
				<div className="mt-8 flex flex-wrap items-center justify-center gap-x-6 gap-y-2 text-sm text-muted-foreground">
					{TRUST_SIGNALS.map((signal) => (
						<span
							key={signal}
							className="flex items-center gap-1.5"
						>
							<span className="size-1.5 rounded-full bg-emerald-500" />
							{signal}
						</span>
					))}
				</div>
			</Wrapper>
		</div>
	);
}
