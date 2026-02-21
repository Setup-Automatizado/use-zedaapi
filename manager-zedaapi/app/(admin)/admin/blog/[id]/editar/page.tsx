import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { requireAdmin } from "@/lib/auth-server";
import {
	getAdminBlogPost,
	getAdminBlogCategories,
	getAdminBlogTags,
} from "@/server/actions/blog";
import { PageHeader } from "@/components/shared/page-header";
import { BlogFormClient } from "./blog-form-client";

export const metadata: Metadata = {
	title: "Editar Post | Admin Zé da API Manager",
};

export default async function EditarBlogPostPage({
	params,
}: {
	params: Promise<{ id: string }>;
}) {
	await requireAdmin();

	const { id } = await params;

	const [postRes, catRes, tagRes] = await Promise.all([
		getAdminBlogPost(id),
		getAdminBlogCategories(),
		getAdminBlogTags(),
	]);

	if (!postRes.success || !postRes.data) {
		notFound();
	}

	const post = postRes.data as {
		id: string;
		title: string;
		slug: string;
		content: string;
		excerpt: string | null;
		coverImageUrl: string | null;
		categoryId: string | null;
		seoTitle: string | null;
		seoDescription: string | null;
		seoKeywords: string | null;
		status: string;
		tags?: Array<{ tag: { id: string; name: string } }>;
		media?: Array<{
			id: string;
			url: string;
			type: string;
			filename: string;
			alt: string | null;
			caption: string | null;
		}>;
	};

	const categories = (catRes.data ?? []) as Array<{
		id: string;
		name: string;
	}>;
	const tags = (tagRes.data ?? []) as Array<{
		id: string;
		name: string;
		slug: string;
	}>;

	return (
		<div className="space-y-6">
			<PageHeader
				title="Editar Post"
				description={`Editando: ${post.title}`}
				backHref="/admin/blog"
			/>
			<BlogFormClient
				categories={categories}
				tags={tags}
				initialData={post}
			/>
		</div>
	);
}
