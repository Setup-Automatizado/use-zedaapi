import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getWaitlist } from "@/server/actions/admin";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { WaitlistTableClient } from "./waitlist-table-client";

export const metadata: Metadata = {
	title: "Waitlist | Admin Zé da API Manager",
};

export default async function AdminWaitlistPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">Waitlist</h1>
				<p className="text-sm text-muted-foreground">
					Gerencie a lista de espera para acesso a plataforma.
				</p>
			</div>

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
