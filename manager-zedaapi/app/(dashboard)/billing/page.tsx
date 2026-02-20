import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { getInvoices } from "@/server/actions/billing";
import { getSubscription } from "@/server/actions/subscriptions";
import {
	TableSkeleton,
	CardSkeleton,
} from "@/components/shared/loading-skeleton";
import { BillingClient } from "./billing-client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { CreditCard, CheckCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import Link from "next/link";

export const metadata: Metadata = {
	title: "Assinatura e Cobranca | Zé da API Manager",
};

const statusConfig: Record<string, { label: string; className: string }> = {
	active: { label: "Ativa", className: "bg-primary/10 text-primary" },
	trialing: { label: "Trial", className: "bg-chart-2/10 text-chart-2" },
	past_due: {
		label: "Vencida",
		className: "bg-destructive/10 text-destructive",
	},
	canceled: {
		label: "Cancelada",
		className: "bg-muted text-muted-foreground",
	},
};

async function SubscriptionSummary() {
	const subscription = await getSubscription();

	if (!subscription) {
		return (
			<Card className="lg:col-span-2">
				<CardContent className="py-8 text-center">
					<p className="text-sm text-muted-foreground">
						Nenhuma assinatura ativa.
					</p>
					<Link
						href="/subscriptions/plans"
						className="mt-2 inline-block text-sm font-medium text-primary hover:underline"
					>
						Ver planos disponiveis
					</Link>
				</CardContent>
			</Card>
		);
	}

	const config = statusConfig[subscription.status] ?? {
		label: subscription.status,
		className: "bg-muted text-muted-foreground",
	};

	const price =
		typeof subscription.plan.price === "string"
			? parseFloat(subscription.plan.price)
			: Number(subscription.plan.price);

	const features = (subscription.plan.features as string[] | null) ?? [];

	return (
		<Card className="lg:col-span-2">
			<CardHeader>
				<div className="flex items-center justify-between">
					<CardTitle className="flex items-center gap-2">
						<CreditCard className="size-4" />
						Plano Atual - {subscription.plan.name}
					</CardTitle>
					<Badge variant="secondary" className={cn(config.className)}>
						{config.label}
					</Badge>
				</div>
			</CardHeader>
			<CardContent className="space-y-4">
				<div className="flex items-baseline gap-2">
					<span className="text-3xl font-bold">
						R${" "}
						{price.toLocaleString("pt-BR", {
							minimumFractionDigits: 2,
						})}
					</span>
					<span className="text-sm text-muted-foreground">
						/{subscription.plan.interval === "year" ? "ano" : "mes"}
					</span>
				</div>

				{subscription.currentPeriodEnd && (
					<div className="text-sm text-muted-foreground">
						Proximo vencimento:{" "}
						{new Date(
							subscription.currentPeriodEnd,
						).toLocaleDateString("pt-BR")}
					</div>
				)}

				{features.length > 0 && (
					<>
						<Separator />
						<div className="space-y-2">
							{features.map((feature) => (
								<div
									key={feature}
									className="flex items-center gap-2 text-sm"
								>
									<CheckCircle className="size-4 text-primary" />
									{feature}
								</div>
							))}
						</div>
					</>
				)}

				<div className="flex gap-2 pt-2">
					<Link
						href="/subscriptions/plans"
						className="inline-flex items-center justify-center rounded-md border border-input bg-background px-3 py-1.5 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground"
					>
						Alterar plano
					</Link>
					<Link
						href="/subscriptions"
						className="inline-flex items-center justify-center rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
					>
						Gerenciar assinatura
					</Link>
				</div>
			</CardContent>
		</Card>
	);
}

async function InvoiceList() {
	const result = await getInvoices(1);

	const invoices = result.invoices.map((inv) => ({
		id: inv.id,
		amount:
			typeof inv.amount === "string"
				? parseFloat(inv.amount)
				: Number(inv.amount),
		currency: inv.currency,
		status: inv.status,
		paidAt: inv.paidAt ? new Date(inv.paidAt).toISOString() : null,
		dueDate: inv.dueDate ? new Date(inv.dueDate).toISOString() : null,
		paymentMethod: inv.paymentMethod,
		pdfUrl: inv.pdfUrl ?? null,
		createdAt: new Date(inv.createdAt).toISOString(),
		subscription: inv.subscription
			? {
					plan: {
						name: inv.subscription.plan.name,
						slug: inv.subscription.plan.slug,
					},
				}
			: null,
		sicrediCharge: inv.sicrediCharge
			? {
					type: inv.sicrediCharge.type,
					status: inv.sicrediCharge.status,
					pixCopiaECola: inv.sicrediCharge.pixCopiaECola,
					linhaDigitavel: inv.sicrediCharge.linhaDigitavel,
				}
			: null,
	}));

	return (
		<BillingClient
			initialInvoices={invoices}
			total={result.total}
			pageSize={result.pageSize}
		/>
	);
}

export default async function BillingPage() {
	await requireAuth();

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">
					Assinatura e Cobranca
				</h1>
				<p className="text-sm text-muted-foreground">
					Gerencie sua assinatura e metodos de pagamento.
				</p>
			</div>

			<div className="grid gap-4 lg:grid-cols-3">
				<Suspense fallback={<CardSkeleton />}>
					<SubscriptionSummary />
				</Suspense>
			</div>

			<Suspense fallback={<TableSkeleton />}>
				<InvoiceList />
			</Suspense>
		</div>
	);
}
