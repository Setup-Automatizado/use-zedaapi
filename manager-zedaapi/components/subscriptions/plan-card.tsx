"use client";

import { cn } from "@/lib/utils";
import {
	Card,
	CardContent,
	CardFooter,
	CardHeader,
	CardTitle,
	CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Check, Loader2 } from "lucide-react";

interface PlanCardProps {
	plan: {
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
	};
	isCurrentPlan?: boolean;
	isPopular?: boolean;
	onSelect?: (planId: string) => void;
	loading?: boolean;
}

export function PlanCard({
	plan,
	isCurrentPlan = false,
	isPopular = false,
	onSelect,
	loading = false,
}: PlanCardProps) {
	const price =
		typeof plan.price === "string" ? parseFloat(plan.price) : plan.price;
	const features = Array.isArray(plan.features) ? plan.features : [];

	return (
		<Card
			className={cn(
				"relative flex flex-col transition-all duration-200",
				isPopular && "ring-2 ring-primary shadow-lg scale-[1.02]",
				isCurrentPlan && !isPopular && "ring-2 ring-primary/50",
			)}
		>
			{isPopular && (
				<div className="absolute -top-3 left-1/2 -translate-x-1/2">
					<Badge className="bg-primary text-primary-foreground px-3 py-0.5 text-xs font-medium">
						Mais popular
					</Badge>
				</div>
			)}

			{isCurrentPlan && !isPopular && (
				<div className="absolute -top-3 left-1/2 -translate-x-1/2">
					<Badge
						variant="outline"
						className="border-primary/50 text-primary bg-primary/5 px-3 py-0.5 text-xs font-medium"
					>
						Plano atual
					</Badge>
				</div>
			)}

			<CardHeader className="pt-8">
				<CardTitle className="text-lg">{plan.name}</CardTitle>
				{plan.description && (
					<CardDescription>{plan.description}</CardDescription>
				)}
			</CardHeader>

			<CardContent className="flex-1">
				<div className="mb-6">
					<span className="text-4xl font-bold tracking-tight tabular-nums">
						{new Intl.NumberFormat("pt-BR", {
							style: "currency",
							currency: plan.currency || "BRL",
						}).format(price)}
					</span>
					<span className="ml-1 text-sm text-muted-foreground">
						/{plan.interval === "year" ? "ano" : "mes"}
					</span>
				</div>

				<div className="space-y-3">
					<div className="flex items-center gap-2 text-sm">
						<Check className="size-4 shrink-0 text-primary" />
						<span>
							Ate{" "}
							<strong>
								{plan.maxInstances === -1
									? "ilimitadas"
									: plan.maxInstances}
							</strong>{" "}
							instancias
						</span>
					</div>

					{features.map((feature, i) => (
						<div
							key={i}
							className="flex items-center gap-2 text-sm"
						>
							<Check className="size-4 shrink-0 text-primary" />
							<span>{String(feature)}</span>
						</div>
					))}
				</div>
			</CardContent>

			<CardFooter>
				<Button
					className="w-full"
					variant={
						isCurrentPlan
							? "outline"
							: isPopular
								? "default"
								: "outline"
					}
					disabled={isCurrentPlan || loading}
					onClick={() => onSelect?.(plan.id)}
				>
					{loading && <Loader2 className="size-4 animate-spin" />}
					{loading
						? "Processando..."
						: isCurrentPlan
							? "Plano atual"
							: "Selecionar plano"}
				</Button>
			</CardFooter>
		</Card>
	);
}
