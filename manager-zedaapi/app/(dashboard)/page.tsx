import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { db } from "@/lib/db";
import {
	CardsSkeleton,
	GridSkeleton,
} from "@/components/shared/loading-skeleton";
import { StatsCards } from "@/components/dashboard/stats-cards";
import { InstanceOverview } from "@/components/dashboard/instance-overview";

export const metadata: Metadata = {
	title: "Dashboard | Zé da API Manager",
};

async function DashboardStats({ userId }: { userId: string }) {
	const [instances, subscription] = await Promise.all([
		db.instance.findMany({
			where: { userId },
			select: { id: true, status: true, name: true, phone: true },
		}),
		db.subscription.findFirst({
			where: { userId, status: "active" },
			include: { plan: true },
		}),
	]);

	const activeInstances = instances.filter(
		(i) => i.status === "connected",
	).length;

	return (
		<StatsCards
			data={{
				activeInstances,
				totalInstances: instances.length,
				subscriptionPlan: subscription?.plan.name ?? "Nenhum",
				subscriptionStatus: subscription?.status ?? "inactive",
				messagesSent: 0,
				nextBillingDate:
					subscription?.currentPeriodEnd?.toISOString() ?? null,
			}}
		/>
	);
}

async function DashboardInstances({ userId }: { userId: string }) {
	const instances = await db.instance.findMany({
		where: { userId },
		select: { id: true, name: true, status: true, phone: true },
		orderBy: { createdAt: "desc" },
		take: 6,
	});

	return (
		<InstanceOverview
			instances={instances.map((i) => ({
				id: i.id,
				name: i.name,
				status: i.status,
				phoneNumber: i.phone ?? undefined,
			}))}
		/>
	);
}

export default async function DashboardPage() {
	const session = await requireAuth();

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
				<p className="text-sm text-muted-foreground">
					Bem-vindo, {session.user.name?.split(" ")[0] ?? "Usuario"}
				</p>
			</div>

			<Suspense fallback={<CardsSkeleton count={4} />}>
				<DashboardStats userId={session.user.id} />
			</Suspense>

			<Suspense fallback={<GridSkeleton count={3} />}>
				<DashboardInstances userId={session.user.id} />
			</Suspense>
		</div>
	);
}
