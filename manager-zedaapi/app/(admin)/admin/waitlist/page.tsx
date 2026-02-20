import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getWaitlist } from "@/server/actions/admin";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { PageHeader } from "@/components/shared/page-header";
import { WaitlistTableClient } from "./waitlist-table-client";

export const metadata: Metadata = {
	title: "Waitlist | Admin Zé da API Manager",
};

export default async function AdminWaitlistPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Waitlist"
				description="Gerencie a lista de espera para acesso à plataforma."
			/>

			<Suspense fallback={<TableSkeleton />}>
				<WaitlistList />
			</Suspense>
		</div>
	);
}

async function WaitlistList() {
	const res = await getWaitlist(1);
	const items = res.data?.items ?? [];
	const total = res.data?.total ?? 0;

	return <WaitlistTableClient initialData={items} initialTotal={total} />;
}
