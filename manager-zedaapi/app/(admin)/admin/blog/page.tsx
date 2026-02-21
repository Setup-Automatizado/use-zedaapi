import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminBlogPosts } from "@/server/actions/blog";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { BlogTableClient } from "./blog-table-client";

export const metadata: Metadata = {
	title: "Blog | Admin Zé da API Manager",
};

export default async function AdminBlogPage() {
	await requireAdmin();

	return (
		<Suspense fallback={<BlogPageSkeleton />}>
			<BlogContent />
		</Suspense>
	);
}

async function BlogContent() {
	const res = await getAdminBlogPosts(1);
	const items =
		res.success && res.data
			? (res.data.items as unknown as Parameters<
					typeof BlogTableClient
				>[0]["initialItems"])
			: [];
	const total = res.success && res.data ? res.data.total : 0;

	return <BlogTableClient initialItems={items} initialTotal={total} />;
}

function BlogPageSkeleton() {
	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<div className="space-y-1">
					<div className="h-7 w-32 animate-pulse rounded-lg bg-muted" />
					<div className="mt-1 h-4 w-56 animate-pulse rounded bg-muted" />
				</div>
				<div className="h-9 w-28 animate-pulse rounded bg-muted" />
			</div>
			<TableSkeleton />
		</div>
	);
}
