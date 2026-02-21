import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminGlossaryTerms } from "@/server/actions/content";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { PageHeader } from "@/components/shared/page-header";
import { GlossarioTableClient } from "./glossario-table-client";

export const metadata: Metadata = {
	title: "Glossario | Admin Zé da API Manager",
};

export default async function GlossarioAdminPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Glossario"
				description="Gerencie os termos do glossario."
			/>

			<Suspense fallback={<TableSkeleton />}>
				<GlossarioContent />
			</Suspense>
		</div>
	);
}

async function GlossarioContent() {
	const res = await getAdminGlossaryTerms(1);
	const items = (res.data?.items ?? []) as Array<{
		id: string;
		term: string;
		slug: string;
		definition: string;
		status: string;
		createdAt: Date;
	}>;
	const total = res.data?.total ?? 0;

	return <GlossarioTableClient initialData={items} initialTotal={total} />;
}
