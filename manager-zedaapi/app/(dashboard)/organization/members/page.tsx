import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
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
			name: session.user.name ?? "Usuario",
			email: session.user.email ?? "",
			role: "owner" as const,
			joinedAt: new Date().toISOString(),
		},
	];

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">Membros</h1>
				<p className="text-sm text-muted-foreground">
					Gerencie os membros da sua organizacao.
				</p>
			</div>

			<MembersClient members={members} />
		</div>
	);
}
