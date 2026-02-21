import { Search } from "lucide-react";
import type { Metadata } from "next";
import Link from "next/link";
import { Suspense } from "react";
import { PostCard } from "@/components/blog/post-card";
import { Badge } from "@/components/ui/badge";
import { db } from "@/lib/db";

export const metadata: Metadata = {
	title: "Blog - Zé da API",
	description:
		"Artigos, tutoriais e novidades sobre WhatsApp API, automação de mensagens, integrações e boas praticas para desenvolvedores.",
	openGraph: {
		title: "Blog - Zé da API",
		description:
			"Artigos e tutoriais sobre WhatsApp API e automação de mensagens.",
		url: "https://zedaapi.com/blog",
	},
	alternates: {
		canonical: "https://zedaapi.com/blog",
	},
};

interface BlogPageProps {
	searchParams: Promise<{ page?: string; categoria?: string }>;
}

export default async function BlogPage({ searchParams }: BlogPageProps) {
	const params = await searchParams;
	const page = Number(params.page) || 1;
	const categorySlug = params.categoria;

	return (
		<div>
			{/* Hero */}
			<section className="border-b border-border bg-muted/30 py-16 sm:py-20">
				<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
					<div className="mx-auto max-w-2xl text-center">
						<h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl">
							Blog da Zé da API
						</h1>
						<p className="mt-4 text-base leading-relaxed text-muted-foreground sm:text-lg">
							Artigos, tutoriais e novidades sobre WhatsApp API,
							automação e integrações.
						</p>
					</div>
				</div>
			</section>

			{/* Content */}
			<section className="py-12 sm:py-16">
				<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
					<Suspense
						fallback={
							<div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
								{Array.from({ length: 6 }).map((_, i) => (
									<div
										key={i}
										className="animate-pulse rounded-2xl border border-border"
									>
										<div className="aspect-video bg-muted" />
										<div className="space-y-3 p-5">
											<div className="h-4 w-20 rounded bg-muted" />
											<div className="h-5 w-3/4 rounded bg-muted" />
											<div className="h-4 w-full rounded bg-muted" />
											<div className="h-3 w-1/2 rounded bg-muted" />
										</div>
									</div>
								))}
							</div>
						}
					>
						<BlogContent page={page} categorySlug={categorySlug} />
					</Suspense>
				</div>
			</section>

			{/* JSON-LD */}
			<script
				type="application/ld+json"
				dangerouslySetInnerHTML={{
					__html: JSON.stringify({
						"@context": "https://schema.org",
						"@type": "Blog",
						name: "Blog da Zé da API",
						description:
							"Artigos sobre WhatsApp API e automação de mensagens.",
						url: "https://zedaapi.com/blog",
						publisher: {
							"@type": "Organization",
							name: "Zé da API",
							url: "https://zedaapi.com",
						},
					}),
				}}
			/>
		</div>
	);
}

async function BlogContent({
	page,
	categorySlug,
}: {
	page: number;
	categorySlug?: string;
}) {
	const pageSize = 12;

	const categoryFilter = categorySlug
		? { category: { slug: categorySlug } }
		: {};

	const [posts, total, categories] = await Promise.all([
		db.blogPost.findMany({
			where: { status: "published", ...categoryFilter },
			include: {
				category: { select: { name: true, slug: true } },
				author: { select: { name: true } },
				tags: { include: { tag: true } },
			},
			orderBy: { publishedAt: "desc" },
			skip: (page - 1) * pageSize,
			take: pageSize,
		}),
		db.blogPost.count({
			where: { status: "published", ...categoryFilter },
		}),
		db.blogCategory.findMany({
			where: {
				posts: { some: { status: "published" } },
			},
			orderBy: { sortOrder: "asc" },
			select: { name: true, slug: true },
		}),
	]);

	const totalPages = Math.ceil(total / pageSize);

	return (
		<div>
			{/* Category pills */}
			{categories.length > 0 && (
				<div className="mb-8 flex flex-wrap gap-2">
					<Link href="/blog">
						<Badge
							variant={!categorySlug ? "default" : "outline"}
							className="cursor-pointer px-3 py-1"
						>
							Todos
						</Badge>
					</Link>
					{categories.map((cat) => (
						<Link
							key={cat.slug}
							href={`/blog?categoria=${cat.slug}`}
						>
							<Badge
								variant={
									categorySlug === cat.slug
										? "default"
										: "outline"
								}
								className="cursor-pointer px-3 py-1"
							>
								{cat.name}
							</Badge>
						</Link>
					))}
				</div>
			)}

			{/* Posts grid */}
			{posts.length === 0 ? (
				<div className="flex min-h-[300px] items-center justify-center rounded-2xl border border-dashed border-border">
					<div className="text-center">
						<Search className="mx-auto size-8 text-muted-foreground/50" />
						<p className="mt-3 text-sm font-medium">
							Nenhum artigo encontrado
						</p>
						<p className="mt-1 text-xs text-muted-foreground">
							Novos artigos em breve.
						</p>
					</div>
				</div>
			) : (
				<div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
					{posts.map((post) => (
						<PostCard
							key={post.id}
							slug={post.slug}
							title={post.title}
							excerpt={post.excerpt}
							coverImageUrl={post.coverImageUrl}
							categoryName={post.category?.name ?? null}
							categorySlug={post.category?.slug ?? null}
							authorName={post.author.name}
							publishedAt={post.publishedAt}
							readingTimeMin={post.readingTimeMin}
							viewCount={post.viewCount}
						/>
					))}
				</div>
			)}

			{/* Pagination */}
			{totalPages > 1 && (
				<div className="mt-10 flex items-center justify-center gap-2">
					{page > 1 && (
						<Link
							href={`/blog?page=${page - 1}${categorySlug ? `&categoria=${categorySlug}` : ""}`}
							className="rounded-lg border border-border px-4 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
						>
							Anterior
						</Link>
					)}
					<span className="px-3 text-sm text-muted-foreground">
						{page} de {totalPages}
					</span>
					{page < totalPages && (
						<Link
							href={`/blog?page=${page + 1}${categorySlug ? `&categoria=${categorySlug}` : ""}`}
							className="rounded-lg border border-border px-4 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
						>
							Proximo
						</Link>
					)}
				</div>
			)}
		</div>
	);
}
