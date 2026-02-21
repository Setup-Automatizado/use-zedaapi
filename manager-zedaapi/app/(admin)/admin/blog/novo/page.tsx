import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import {
	getAdminBlogCategories,
	getAdminBlogTags,
} from "@/server/actions/blog";
import { PageHeader } from "@/components/shared/page-header";
import { BlogFormClient } from "./blog-form-client";

export const metadata: Metadata = {
	title: "Novo Post | Admin Zé da API Manager",
};

export default async function NovoBlogPostPage() {
	await requireAdmin();

	const [catRes, tagRes] = await Promise.all([
		getAdminBlogCategories(),
		getAdminBlogTags(),
	]);

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
				title="Novo Post"
				description="Crie um novo post para o blog."
				backHref="/admin/blog"
			/>
			<BlogFormClient categories={categories} tags={tags} />
		</div>
	);
}
