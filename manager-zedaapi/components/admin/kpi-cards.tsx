"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
	DollarSign,
	TrendingDown,
	CreditCard,
	Smartphone,
	Users,
	UserPlus,
} from "lucide-react";

interface KpiData {
	mrr: number;
	churnRate: number;
	activeSubscriptions: number;
	totalInstances: number;
	newUsersLast30d: number;
	totalUsers: number;
}

interface KpiCardsProps {
	data?: KpiData;
}

export function KpiCards({ data }: KpiCardsProps) {
	const kpis = [
		{
			title: "MRR",
			value: data
				? `R$ ${data.mrr.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}`
				: "--",
			icon: DollarSign,
			color: "text-primary",
			bg: "bg-primary/10",
		},
		{
			title: "Taxa de Churn",
			value: data ? `${data.churnRate.toFixed(1)}%` : "--",
			icon: TrendingDown,
			color: "text-destructive",
			bg: "bg-destructive/10",
			description: "últimos 30 dias",
		},
		{
			title: "Assinaturas Ativas",
			value: data?.activeSubscriptions?.toLocaleString("pt-BR") ?? "--",
			icon: CreditCard,
			color: "text-chart-3",
			bg: "bg-chart-3/10",
		},
		{
			title: "Total Instâncias",
			value: data?.totalInstances?.toLocaleString("pt-BR") ?? "--",
			icon: Smartphone,
			color: "text-chart-4",
			bg: "bg-chart-4/10",
		},
		{
			title: "Novos (30d)",
			value: data?.newUsersLast30d?.toLocaleString("pt-BR") ?? "--",
			icon: UserPlus,
			color: "text-chart-2",
			bg: "bg-chart-2/10",
		},
		{
			title: "Total Usuários",
			value: data?.totalUsers?.toLocaleString("pt-BR") ?? "--",
			icon: Users,
			color: "text-chart-1",
			bg: "bg-chart-1/10",
		},
	];

	return (
		<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
			{kpis.map((kpi) => (
				<Card key={kpi.title}>
					<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
						<CardTitle className="text-sm font-medium text-muted-foreground">
							{kpi.title}
						</CardTitle>
						<div
							className={`flex size-8 items-center justify-center rounded-lg ${kpi.bg}`}
						>
							<kpi.icon className={`size-4 ${kpi.color}`} />
						</div>
					</CardHeader>
					<CardContent>
						<div className="text-xl font-bold tracking-tight tabular-nums">
							{kpi.value}
						</div>
						{kpi.description && (
							<p className="mt-1 text-xs text-muted-foreground">
								{kpi.description}
							</p>
						)}
					</CardContent>
				</Card>
			))}
		</div>
	);
}
