import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { OrganizationForm } from "./organization-form";

export const metadata: Metadata = {
	title: "Organizacao | Zé da API Manager",
};

export default async function OrganizationPage() {
	const session = await requireAuth();

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">
					Organizacao
				</h1>
				<p className="text-sm text-muted-foreground">
					Gerencie as configuracoes da sua organizacao.
				</p>
			</div>

			<OrganizationForm
				user={{
					name: session.user.name ?? "",
					email: session.user.email ?? "",
				}}
			/>
		</div>
	);
}
