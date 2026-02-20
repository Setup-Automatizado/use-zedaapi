"use client";

import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	Smartphone,
	CreditCard,
	MessageSquare,
	CalendarClock,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface StatsData {
	activeInstances: number;
	totalInstances: number;
	subscriptionPlan: string;
	subscriptionStatus: string;
	messagesSent: number;
	nextBillingDate: string | null;
}

interface StatsCardsProps {
	data?: StatsData;
}

const statusColors: Record<string, string> = {
	active: "text-primary",
	trialing: "text-chart-2",
	past_due: "text-destructive",
	canceled: "text-muted-foreground",
};

export function StatsCards({ data }: StatsCardsProps) {
	const stats = [
		{
			title: "Instancias Ativas",
			value: data
				? `${data.activeInstances}/${data.totalInstances}`
				: "--",
			icon: Smartphone,
			description: "conectadas agora",
		},
		{
			title: "Assinatura",
			value: data?.subscriptionPlan ?? "--",
			icon: CreditCard,
			description: data?.subscriptionStatus ?? "",
			valueClassName:
				statusColors[data?.subscriptionStatus ?? ""] ?? "",
		},
		{
			title: "Mensagens Enviadas",
			value: data?.messagesSent?.toLocaleString("pt-BR") ?? "--",
			icon: MessageSquare,
			description: "ultimos 30 dias",
		},
		{
			title: "Proximo Vencimento",
			value: data?.nextBillingDate
				? new Date(data.nextBillingDate).toLocaleDateString("pt-BR", {
						day: "2-digit",
						month: "short",
					})
				: "--",
			icon: CalendarClock,
			description: data?.nextBillingDate ? "fatura pendente" : "",
		},
	];

	return (
		<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			{stats.map((stat) => (
				<Card key={stat.title}>
					<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
						<CardTitle className="text-sm font-medium text-muted-foreground">
							{stat.title}
						</CardTitle>
						<stat.icon className="size-4 text-muted-foreground" />
					</CardHeader>
					<CardContent>
						<div
							className={cn(
								"text-2xl font-bold tabular-nums",
								stat.valueClassName,
							)}
						>
							{stat.value}
						</div>
						{stat.description && (
							<p className="mt-1 text-xs text-muted-foreground">
								{stat.description}
							</p>
						)}
					</CardContent>
				</Card>
			))}
		</div>
	);
}
