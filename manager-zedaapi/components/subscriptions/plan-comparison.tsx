"use client";

import { useCallback, useRef } from "react";
import { PricingCard } from "@/components/shared/pricing-card";

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
	const currentTierSlugRef = useRef<string>("");

	const handleTierChange = useCallback((tierSlug: string) => {
		currentTierSlugRef.current = tierSlug;
	}, []);

	const handleCtaClick = useCallback(() => {
		const matchingPlan = plans.find(
			(p) => p.slug === currentTierSlugRef.current,
		);
		if (matchingPlan) {
			onSelectPlan(matchingPlan.id);
		}
	}, [plans, onSelectPlan]);

	return (
		<div className="relative mx-auto max-w-5xl">
			<PricingCard
				showAnnualToggle
				ctaContent="Assinar agora"
				onCtaClick={handleCtaClick}
				ctaLoading={loading}
				currentPlanSlug={currentPlanSlug}
				onTierChange={handleTierChange}
			/>
		</div>
	);
}
