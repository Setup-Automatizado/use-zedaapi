import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminActivityLog } from "@/server/actions/admin";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { PageHeader } from "@/components/shared/page-header";
import { ActivityLogClient } from "./activity-log-client";

export const metadata: Metadata = {
	title: "Log de Atividade | Admin Zé da API Manager",
};

export default async function AdminActivityLogPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Log de Atividade"
				description="Histórico de ações realizadas na plataforma."
			/>

			<Suspense fallback={<TableSkeleton />}>
				<ActivityLogContent />
			</Suspense>
		</div>
	);
}

async function ActivityLogContent() {
	const res = await getAdminActivityLog(1);
	const items = res.data?.items ?? [];
	const total = res.data?.total ?? 0;

	return <ActivityLogClient initialData={items} initialTotal={total} />;
}
