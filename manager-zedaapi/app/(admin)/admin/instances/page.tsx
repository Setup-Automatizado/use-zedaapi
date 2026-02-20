import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminInstances } from "@/server/actions/admin";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { InstancesTableClient } from "./instances-table-client";

export const metadata: Metadata = {
	title: "Instancias | Admin Zé da API Manager",
};

export default async function AdminInstancesPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">
					Instancias
				</h1>
				<p className="text-sm text-muted-foreground">
					Todas as instancias de todos os usuarios.
				</p>
			</div>

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
