"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { PlanComparison } from "@/components/subscriptions/plan-comparison";
import { PaymentMethodForm } from "@/components/billing/payment-method-form";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { checkout } from "@/server/actions/subscriptions";
import { ArrowLeft } from "lucide-react";

interface Plan {
	id: string;
	name: string;
	slug: string;
	description: string | null;
	price: number;
	currency: string;
	interval: string;
	maxInstances: number;
	features: string[];
	active: boolean;
}

interface PlansClientProps {
	plans: Plan[];
	currentPlanSlug: string | null;
}

export function PlansClient({ plans, currentPlanSlug }: PlansClientProps) {
	const router = useRouter();
	const [selectedPlanId, setSelectedPlanId] = useState<string | null>(null);
	const [loading, setLoading] = useState(false);
	const [step, setStep] = useState<"plan" | "payment">("plan");

	const selectedPlan = plans.find((p) => p.id === selectedPlanId);

	function handleSelectPlan(planId: string) {
		setSelectedPlanId(planId);
		setStep("payment");
	}

	async function handlePayment(method: "stripe" | "pix" | "boleto") {
		if (!selectedPlanId) return;

		setLoading(true);
		try {
			const result = await checkout(selectedPlanId, method);

			if (result.success && result.data?.url) {
				window.location.href = result.data.url;
			} else if (result.success) {
				router.push("/billing");
			} else {
				console.error("Checkout failed:", result.error);
			}
		} catch (err) {
			console.error("Checkout error:", err);
		} finally {
			setLoading(false);
		}
	}

	if (step === "payment" && selectedPlan) {
		return (
			<div className="mx-auto max-w-lg space-y-6">
				<Button
					variant="ghost"
					size="sm"
					onClick={() => setStep("plan")}
					className="gap-1"
				>
					<ArrowLeft className="size-3.5" />
					Voltar aos planos
				</Button>

				<Card>
					<CardHeader>
						<CardTitle className="text-lg">
							Forma de pagamento — {selectedPlan.name}
						</CardTitle>
						<p className="text-sm text-muted-foreground">
							{new Intl.NumberFormat("pt-BR", {
								style: "currency",
								currency: "BRL",
							}).format(selectedPlan.price)}
							/mes — ate {selectedPlan.maxInstances} instancias
						</p>
					</CardHeader>
					<CardContent>
						<PaymentMethodForm
							onSelect={handlePayment}
							loading={loading}
						/>
					</CardContent>
				</Card>
			</div>
		);
	}

	return (
		<PlanComparison
			plans={plans}
			currentPlanSlug={currentPlanSlug}
			onSelectPlan={handleSelectPlan}
			loading={loading}
			loadingPlanId={selectedPlanId}
		/>
	);
}
