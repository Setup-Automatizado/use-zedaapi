import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { requireAdmin } from "@/lib/auth-server";
import {
	getAdminSupportArticle,
	getAdminSupportCategories,
} from "@/server/actions/content";
import { PageHeader } from "@/components/shared/page-header";
import { SuporteFormClient } from "./suporte-form-client";

export const metadata: Metadata = {
	title: "Editar Artigo de Suporte | Admin Zé da API Manager",
};

export default async function EditarSuportePage({
	params,
}: {
	params: Promise<{ id: string }>;
}) {
	await requireAdmin();

	const { id } = await params;

	const [articleRes, categoriesRes] = await Promise.all([
		getAdminSupportArticle(id),
		getAdminSupportCategories(),
	]);

	if (!articleRes.success || !articleRes.data) {
		notFound();
	}

	const article = articleRes.data as {
		id: string;
		title: string;
		slug: string;
		content: string;
		excerpt: string | null;
		categoryId: string | null;
		seoTitle: string | null;
		seoDescription: string | null;
		status: string;
		sortOrder: number;
	};

	const categories = (categoriesRes.data ?? []) as Array<{
		id: string;
		name: string;
	}>;

	return (
		<div className="space-y-6">
			<PageHeader
				title="Editar Artigo"
				description="Atualize as informacoes do artigo de suporte."
				backHref="/admin/suporte"
			/>

			<SuporteFormClient categories={categories} initialData={article} />
		</div>
	);
}
