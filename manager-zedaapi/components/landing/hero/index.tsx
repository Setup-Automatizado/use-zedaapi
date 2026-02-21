"use client";

import { useReducedMotion } from "@/components/shared/motion";
import { DEMO_SCENARIOS } from "./animation-data";
import { useHeroAnimation } from "./use-hero-animation";
import { HeroContent } from "./hero-content";
import { ApiTerminal } from "./api-terminal";
import { WhatsAppChat } from "./whatsapp-chat";

export function Hero() {
	const prefersReducedMotion = useReducedMotion();
	const reducedMotion = prefersReducedMotion ?? false;

	const {
		typedChars,
		showResponse,
		showTypingIndicator,
		chatMessages,
		currentScenario,
	} = useHeroAnimation(!reducedMotion);

	const scenario = currentScenario ?? DEMO_SCENARIOS[0]!;

	return (
		<section className="relative overflow-hidden">
			{/* Background grid */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute inset-0 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] bg-[size:4rem_4rem] opacity-[0.25] [mask-image:radial-gradient(ellipse_70%_50%_at_50%_0%,black_30%,transparent_100%)]"
			/>

			{/* Primary gradient glow */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute -top-40 left-1/2 h-[600px] w-[900px] -translate-x-1/2 rounded-full bg-primary/8 blur-[140px]"
			/>

			{/* Secondary glow */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute top-20 -right-40 h-[300px] w-[400px] rounded-full bg-primary/5 blur-[100px]"
			/>

			<div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				<div className="flex flex-col items-center gap-14 pb-20 pt-24 sm:pb-28 sm:pt-32 lg:pb-32 lg:pt-36">
					{/* Top — text content (centered) */}
					<div className="max-w-3xl">
						<HeroContent reducedMotion={reducedMotion} />
					</div>

					{/* Bottom — terminal + chat side by side */}
					<div className="grid w-full grid-cols-1 gap-4 lg:grid-cols-2">
						<ApiTerminal
							scenario={scenario}
							typedChars={typedChars}
							showResponse={showResponse}
							reducedMotion={reducedMotion}
						/>
						<WhatsAppChat
							messages={chatMessages}
							showTypingIndicator={showTypingIndicator}
							reducedMotion={reducedMotion}
						/>
					</div>
				</div>
			</div>
		</section>
	);
}
