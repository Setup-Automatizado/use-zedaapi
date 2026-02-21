import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminUsers } from "@/server/actions/admin";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { PageHeader } from "@/components/shared/page-header";
import { UsersTableClient } from "./users-table-client";

export const metadata: Metadata = {
	title: "Usuários | Admin Zé da API Manager",
};

export default async function AdminUsersPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Usuários"
				description="Gerencie todos os usuários da plataforma."
			/>

			<Suspense fallback={<TableSkeleton />}>
				<UsersList />
			</Suspense>
		</div>
	);
}

async function UsersList() {
	const res = await getAdminUsers(1);
	const items = res.data?.items ?? [];
	const total = res.data?.total ?? 0;

	return <UsersTableClient initialData={items} initialTotal={total} />;
}
