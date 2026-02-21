import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminBlogCategories } from "@/server/actions/blog";
import { CategoriasClient } from "./categorias-client";

export const metadata: Metadata = {
	title: "Categorias do Blog | Admin Zé da API Manager",
};

export default async function AdminBlogCategoriasPage() {
	await requireAdmin();

	const res = await getAdminBlogCategories();
	const categories = (res.data ?? []) as Array<{
		id: string;
		name: string;
		slug: string;
		description: string | null;
		sortOrder: number;
		_count: { posts: number };
	}>;

	return <CategoriasClient initialCategories={categories} />;
}
