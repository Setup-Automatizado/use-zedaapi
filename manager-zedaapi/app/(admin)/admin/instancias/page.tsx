import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminInstances } from "@/server/actions/admin";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { PageHeader } from "@/components/shared/page-header";
import { InstancesTableClient } from "./instances-table-client";

export const metadata: Metadata = {
	title: "Instâncias | Admin Zé da API Manager",
};

export default async function AdminInstancesPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Instâncias"
				description="Todas as instâncias de todos os usuários."
			/>

			<Suspense fallback={<TableSkeleton />}>
				<InstancesList />
			</Suspense>
		</div>
	);
}

async function InstancesList() {
	const res = await getAdminInstances(1);
	const items = res.data?.items ?? [];
	const total = res.data?.total ?? 0;

	return <InstancesTableClient initialData={items} initialTotal={total} />;
}
