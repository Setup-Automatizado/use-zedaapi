import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminUsers } from "@/server/actions/admin";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { UsersTableClient } from "./users-table-client";

export const metadata: Metadata = {
	title: "Usuarios | Admin Zé da API Manager",
};

export default async function AdminUsersPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">Usuarios</h1>
				<p className="text-sm text-muted-foreground">
					Gerencie todos os usuarios da plataforma.
				</p>
			</div>

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
