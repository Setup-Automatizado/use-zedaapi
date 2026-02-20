import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import {
	getDashboardStats,
	getRevenueHistory,
	getSubscriptionsByPlan,
	getRecentActivity,
} from "@/server/actions/admin";
import { KpiCards } from "@/components/admin/kpi-cards";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
	CardsSkeleton,
	ChartSkeleton,
} from "@/components/shared/loading-skeleton";
import { Activity } from "lucide-react";
import { AdminChartsClient } from "./admin-charts-client";

export const metadata: Metadata = {
	title: "Painel Admin | Zé da API Manager",
};

export default async function AdminDashboardPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">
					Painel Admin
				</h1>
				<p className="text-sm text-muted-foreground">
					Visao geral do sistema.
				</p>
			</div>

			<Suspense fallback={<CardsSkeleton count={6} />}>
				<AdminStats />
			</Suspense>

			<Suspense
				fallback={
					<div className="grid gap-4 lg:grid-cols-2">
						<ChartSkeleton />
						<ChartSkeleton />
					</div>
				}
			>
				<AdminCharts />
			</Suspense>

			<Suspense
				fallback={
					<Card>
						<CardContent className="p-6">
							<div className="space-y-3">
								{Array.from({ length: 5 }).map((_, i) => (
									<div
										key={i}
										className="h-10 animate-pulse rounded bg-muted"
									/>
								))}
							</div>
						</CardContent>
					</Card>
				}
			>
				<RecentActivitySection />
			</Suspense>
		</div>
	);
}

async function AdminStats() {
	const stats = await getDashboardStats();
	return (
		<KpiCards
			data={{
				mrr: stats.mrr,
				churnRate: stats.churnRate,
				activeSubscriptions: stats.activeSubscriptions,
				totalInstances: stats.totalInstances,
				newUsersLast30d: stats.newUsers30d,
				totalUsers: 0,
			}}
		/>
	);
}

async function AdminCharts() {
	const [revenueRes, planRes] = await Promise.all([
		getRevenueHistory(),
		getSubscriptionsByPlan(),
	]);

	const revenue = revenueRes.data ?? [];
	const plans = planRes.data ?? [];

	return <AdminChartsClient revenue={revenue} plans={plans} />;
}

async function RecentActivitySection() {
	const activityRes = await getRecentActivity();
	const activity = activityRes.data ?? [];

	return (
		<Card>
			<CardHeader>
				<CardTitle className="flex items-center gap-2 text-sm font-medium">
					<Activity className="size-4" />
					Atividade Recente
				</CardTitle>
			</CardHeader>
			<CardContent>
				{activity.length === 0 ? (
					<p className="py-8 text-center text-sm text-muted-foreground">
						Nenhuma atividade recente
					</p>
				) : (
					<div className="space-y-2">
						{activity.map((item) => (
							<div
								key={item.id}
								className="flex items-center justify-between gap-4 rounded-lg border p-3"
							>
								<div className="flex items-center gap-3 min-w-0">
									<Badge
										variant="outline"
										className="shrink-0"
									>
										{item.action}
									</Badge>
									<span className="truncate text-sm">
										{item.resource}
									</span>
								</div>
								<div className="flex items-center gap-3 shrink-0 text-xs text-muted-foreground">
									{item.userName && (
										<span>{item.userName}</span>
									)}
									<span>
										{new Date(
											item.timestamp,
										).toLocaleDateString("pt-BR", {
											day: "2-digit",
											month: "2-digit",
											hour: "2-digit",
											minute: "2-digit",
										})}
									</span>
								</div>
							</div>
						))}
					</div>
				)}
			</CardContent>
		</Card>
	);
}
