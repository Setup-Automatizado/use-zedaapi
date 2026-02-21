import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminSupportArticles } from "@/server/actions/content";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { PageHeader } from "@/components/shared/page-header";
import { SuporteTableClient } from "./suporte-table-client";

export const metadata: Metadata = {
	title: "Artigos de Suporte | Admin Zé da API Manager",
};

export default async function AdminSuportePage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Artigos de Suporte"
				description="Gerencie os artigos da central de ajuda."
			/>

			<Suspense fallback={<TableSkeleton />}>
				<SuporteContent />
			</Suspense>
		</div>
	);
}

async function SuporteContent() {
	const res = await getAdminSupportArticles(1);
	const items = (res.data?.items ?? []) as unknown as Parameters<
		typeof SuporteTableClient
	>[0]["initialData"];
	const total = res.data?.total ?? 0;

	return <SuporteTableClient initialData={items} initialTotal={total} />;
}
