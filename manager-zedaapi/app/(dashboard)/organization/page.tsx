import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { PageHeader } from "@/components/shared/page-header";
import { OrganizationForm } from "./organization-form";

export const metadata: Metadata = {
	title: "Organização | Zé da API Manager",
};

export default async function OrganizationPage() {
	const session = await requireAuth();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Organização"
				description="Gerencie as configurações da sua organização."
			/>

			<OrganizationForm
				user={{
					name: session.user.name ?? "",
					email: session.user.email ?? "",
				}}
			/>
		</div>
	);
}
