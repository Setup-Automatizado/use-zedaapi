"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
	CardDescription,
	CardFooter,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
	cancelSubscription,
	getBillingPortal,
} from "@/server/actions/subscriptions";
import { CreditCard, Calendar, Layers, AlertTriangle } from "lucide-react";

interface SubscriptionDetailsProps {
	subscription: {
		id: string;
		status: string;
		paymentMethod: string;
		cancelAtPeriodEnd: boolean;
		currentPeriodStart: Date;
		currentPeriodEnd: Date;
		stripeSubscriptionId: string | null;
		plan: {
			id: string;
			name: string;
			slug: string;
			description: string | null;
			price: unknown;
			currency: string;
			interval: string;
			maxInstances: number;
		};
		_count: {
			instances: number;
		};
	} | null;
}

const statusLabels: Record<
	string,
	{
		label: string;
		variant: "default" | "secondary" | "destructive" | "outline";
	}
> = {
	active: { label: "Ativa", variant: "default" },
	trialing: { label: "Periodo de teste", variant: "secondary" },
	past_due: { label: "Pagamento pendente", variant: "destructive" },
	canceled: { label: "Cancelada", variant: "outline" },
	incomplete: { label: "Incompleta", variant: "secondary" },
	paused: { label: "Pausada", variant: "outline" },
};

function formatDate(date: Date): string {
	return new Intl.DateTimeFormat("pt-BR", {
		day: "2-digit",
		month: "long",
		year: "numeric",
	}).format(new Date(date));
}

function formatCurrency(amount: unknown, currency = "BRL"): string {
	const num =
		typeof amount === "string" ? parseFloat(amount) : Number(amount);
	return new Intl.NumberFormat("pt-BR", {
		style: "currency",
		currency,
	}).format(num);
}

export function SubscriptionDetails({
	subscription,
}: SubscriptionDetailsProps) {
	const router = useRouter();
	const [cancelLoading, setCancelLoading] = useState(false);
	const [portalLoading, setPortalLoading] = useState(false);

	if (!subscription) {
		return (
			<Card>
				<CardContent className="py-12 text-center">
					<Layers className="mx-auto h-10 w-10 text-muted-foreground/50 mb-4" />
					<h3 className="font-medium mb-1">Sem assinatura ativa</h3>
					<p className="text-sm text-muted-foreground mb-6">
						Escolha um plano para comecar a usar o Zé da API
					</p>
					<Button onClick={() => router.push("/assinaturas/planos")}>
						Ver planos
					</Button>
				</CardContent>
			</Card>
		);
	}

	const status = statusLabels[subscription.status] ?? {
		label: subscription.status,
		variant: "secondary" as const,
	};

	async function handleCancel() {
		if (
			!confirm(
				"Tem certeza que deseja cancelar sua assinatura? Ela permanecera ativa ate o fim do periodo atual.",
			)
		) {
			return;
		}

		setCancelLoading(true);
		try {
			await cancelSubscription(subscription!.id);
			router.refresh();
		} catch (err) {
			console.error("Cancel error:", err);
		} finally {
			setCancelLoading(false);
		}
	}

	async function handleManageBilling() {
		setPortalLoading(true);
		try {
			const result = await getBillingPortal();
			if (result.success && result.data?.url) {
				window.location.href = result.data.url;
			}
		} catch (err) {
			console.error("Portal error:", err);
		} finally {
			setPortalLoading(false);
		}
	}

	return (
		<div className="space-y-6">
			<Card>
				<CardHeader>
					<div className="flex items-center justify-between">
						<div>
							<CardTitle>{subscription.plan.name}</CardTitle>
							{subscription.plan.description && (
								<CardDescription className="mt-1">
									{subscription.plan.description}
								</CardDescription>
							)}
						</div>
						<Badge variant={status.variant}>{status.label}</Badge>
					</div>
				</CardHeader>

				<CardContent className="space-y-4">
					<div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
						<div className="flex items-center gap-3">
							<CreditCard className="h-4 w-4 text-muted-foreground" />
							<div>
								<p className="text-xs text-muted-foreground">
									Valor
								</p>
								<p className="text-sm font-medium">
									{formatCurrency(
										subscription.plan.price,
										subscription.plan.currency,
									)}
									/
									{subscription.plan.interval === "year"
										? "ano"
										: "mes"}
								</p>
							</div>
						</div>

						<div className="flex items-center gap-3">
							<Calendar className="h-4 w-4 text-muted-foreground" />
							<div>
								<p className="text-xs text-muted-foreground">
									Proximo pagamento
								</p>
								<p className="text-sm font-medium">
									{formatDate(subscription.currentPeriodEnd)}
								</p>
							</div>
						</div>

						<div className="flex items-center gap-3">
							<Layers className="h-4 w-4 text-muted-foreground" />
							<div>
								<p className="text-xs text-muted-foreground">
									Instancias
								</p>
								<p className="text-sm font-medium">
									{subscription._count.instances} /{" "}
									{subscription.plan.maxInstances === -1
										? "ilimitadas"
										: subscription.plan.maxInstances}
								</p>
							</div>
						</div>
					</div>

					{subscription.cancelAtPeriodEnd && (
						<div className="flex items-center gap-2 rounded-lg bg-chart-2/10 border border-chart-2/30 px-4 py-3 text-sm">
							<AlertTriangle className="size-4 text-chart-2 shrink-0" />
							<p className="text-foreground">
								Sua assinatura sera cancelada em{" "}
								<strong>
									{formatDate(subscription.currentPeriodEnd)}
								</strong>
								.
							</p>
						</div>
					)}
				</CardContent>

				<Separator />

				<CardFooter className="pt-4 gap-3">
					<Button
						variant="outline"
						onClick={() => router.push("/assinaturas/planos")}
					>
						Trocar plano
					</Button>

					{subscription.stripeSubscriptionId && (
						<Button
							variant="outline"
							onClick={handleManageBilling}
							disabled={portalLoading}
						>
							{portalLoading
								? "Abrindo..."
								: "Gerenciar pagamento"}
						</Button>
					)}

					{subscription.status === "active" &&
						!subscription.cancelAtPeriodEnd && (
							<Button
								variant="ghost"
								className="text-destructive hover:text-destructive ml-auto"
								onClick={handleCancel}
								disabled={cancelLoading}
							>
								{cancelLoading
									? "Cancelando..."
									: "Cancelar assinatura"}
							</Button>
						)}
				</CardFooter>
			</Card>
		</div>
	);
}
