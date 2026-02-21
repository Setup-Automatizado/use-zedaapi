import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminSupportCategories } from "@/server/actions/content";
import { PageHeader } from "@/components/shared/page-header";
import { CategoriasClient } from "./categorias-client";

export const metadata: Metadata = {
	title: "Categorias de Suporte | Admin Zé da API Manager",
};

export default async function CategoriasSuportePage() {
	await requireAdmin();

	const res = await getAdminSupportCategories();
	const categories = (res.data ?? []) as Array<{
		id: string;
		name: string;
		slug: string;
		description: string | null;
		icon: string | null;
		sortOrder: number;
		_count: { articles: number };
	}>;

	return (
		<div className="space-y-6">
			<PageHeader
				title="Categorias de Suporte"
				description="Gerencie as categorias da central de ajuda."
				backHref="/admin/suporte"
			/>

			<CategoriasClient initialCategories={categories} />
		</div>
	);
}
