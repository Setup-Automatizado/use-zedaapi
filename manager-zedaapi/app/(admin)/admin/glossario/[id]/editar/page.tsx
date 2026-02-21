import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminGlossaryTerm } from "@/server/actions/content";
import { PageHeader } from "@/components/shared/page-header";
import { GlossarioFormClient } from "./glossario-form-client";

export const metadata: Metadata = {
	title: "Editar Termo | Admin Zé da API Manager",
};

export default async function EditarGlossarioPage({
	params,
}: {
	params: Promise<{ id: string }>;
}) {
	await requireAdmin();

	const { id } = await params;
	const res = await getAdminGlossaryTerm(id);

	if (!res.success || !res.data) {
		notFound();
	}

	const term = res.data as {
		id: string;
		term: string;
		slug: string;
		definition: string;
		content: string | null;
		seoTitle: string | null;
		seoDescription: string | null;
		status: string;
		relatedSlugs: string[] | null;
	};

	return (
		<div className="space-y-6">
			<PageHeader
				title="Editar Termo"
				description={`Editando "${term.term}"`}
				backHref="/admin/glossario"
			/>

			<GlossarioFormClient initialData={term} />
		</div>
	);
}
