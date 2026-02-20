import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminSubscriptions } from "@/server/actions/admin";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { PageHeader } from "@/components/shared/page-header";
import { SubscriptionsTableClient } from "./subscriptions-table-client";

export const metadata: Metadata = {
	title: "Assinaturas | Admin Zé da API Manager",
};

export default async function AdminSubscriptionsPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Assinaturas"
				description="Todas as assinaturas da plataforma."
			/>

			<Suspense fallback={<TableSkeleton />}>
				<SubscriptionsList />
			</Suspense>
		</div>
	);
}

async function SubscriptionsList() {
	const res = await getAdminSubscriptions(1);
	const items = res.data?.items ?? [];
	const total = res.data?.total ?? 0;

	return (
		<SubscriptionsTableClient initialData={items} initialTotal={total} />
	);
}
