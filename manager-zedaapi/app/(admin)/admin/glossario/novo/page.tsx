import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { PageHeader } from "@/components/shared/page-header";
import { GlossarioFormClient } from "./glossario-form-client";

export const metadata: Metadata = {
	title: "Novo Termo | Admin Zé da API Manager",
};

export default async function NovoGlossarioPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Novo Termo"
				description="Adicione um novo termo ao glossario."
				backHref="/admin/glossario"
			/>

			<GlossarioFormClient />
		</div>
	);
}
