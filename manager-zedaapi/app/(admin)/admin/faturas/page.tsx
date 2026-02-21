import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminInvoices } from "@/server/actions/admin";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { PageHeader } from "@/components/shared/page-header";
import { InvoicesTableClient } from "./invoices-table-client";

export const metadata: Metadata = {
	title: "Faturas | Admin Zé da API Manager",
};

export default async function AdminInvoicesPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Faturas"
				description="Todas as faturas geradas na plataforma."
			/>

			<Suspense fallback={<TableSkeleton />}>
				<InvoicesList />
			</Suspense>
		</div>
	);
}

async function InvoicesList() {
	const res = await getAdminInvoices(1);
	const items = res.data?.items ?? [];
	const total = res.data?.total ?? 0;

	return <InvoicesTableClient initialData={items} initialTotal={total} />;
}
