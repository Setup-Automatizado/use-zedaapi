import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminSupportCategories } from "@/server/actions/content";
import { PageHeader } from "@/components/shared/page-header";
import { SuporteFormClient } from "./suporte-form-client";

export const metadata: Metadata = {
	title: "Novo Artigo de Suporte | Admin Zé da API Manager",
};

export default async function NovoSuportePage() {
	await requireAdmin();

	const res = await getAdminSupportCategories();
	const categories = (res.data ?? []) as Array<{
		id: string;
		name: string;
	}>;

	return (
		<div className="space-y-6">
			<PageHeader
				title="Novo Artigo"
				description="Crie um novo artigo para a central de ajuda."
				backHref="/admin/suporte"
			/>

			<SuporteFormClient categories={categories} />
		</div>
	);
}
