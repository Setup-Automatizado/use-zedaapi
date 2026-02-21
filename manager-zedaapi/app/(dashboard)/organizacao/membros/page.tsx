import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { PageHeader } from "@/components/shared/page-header";
import { MembersClient } from "./members-client";

export const metadata: Metadata = {
	title: "Membros | Zé da API Manager",
};

export default async function MembersPage() {
	const session = await requireAuth();

	// TODO: Fetch real members from organization when org system is implemented
	// For now, show the current user as owner
	const members = [
		{
			id: session.user.id,
			name: session.user.name ?? "Usuário",
			email: session.user.email ?? "",
			role: "owner" as const,
			joinedAt: new Date().toISOString(),
		},
	];

	return (
		<div className="space-y-6">
			<PageHeader
				title="Membros"
				description="Gerencie os membros da sua organização."
			/>

			<MembersClient members={members} />
		</div>
	);
}
