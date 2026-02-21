import { ArrowLeft } from "lucide-react";
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { PostCard } from "@/components/blog/post-card";
import { db } from "@/lib/db";

interface CategoryPageProps {
	params: Promise<{ slug: string }>;
	searchParams: Promise<{ page?: string }>;
}

export async function generateMetadata({
	params,
}: CategoryPageProps): Promise<Metadata> {
	const { slug } = await params;
	const category = await db.blogCategory.findFirst({
		where: { slug },
		select: { name: true, description: true, slug: true },
	});

	if (!category) return { title: "Categoria nao encontrada" };

	return {
		title: `${category.name} - Blog Zé da API`,
		description:
			category.description ??
			`Artigos sobre ${category.name} no blog do Zé da API.`,
		alternates: {
			canonical: `https://zedaapi.com/blog/categoria/${category.slug}`,
		},
	};
}

export default async function BlogCategoryPage({
	params,
	searchParams,
}: CategoryPageProps) {
	const { slug } = await params;
	const sp = await searchParams;
	const page = Number(sp.page) || 1;
	const pageSize = 12;

	const category = await db.blogCategory.findFirst({
		where: { slug },
	});

	if (!category) notFound();

	const [posts, total] = await Promise.all([
		db.blogPost.findMany({
			where: { status: "published", categoryId: category.id },
			include: {
				category: { select: { name: true, slug: true } },
				author: { select: { name: true } },
			},
			orderBy: { publishedAt: "desc" },
			skip: (page - 1) * pageSize,
			take: pageSize,
		}),
		db.blogPost.count({
			where: { status: "published", categoryId: category.id },
		}),
	]);

	const totalPages = Math.ceil(total / pageSize);

	return (
		<div>
			{/* Hero */}
			<section className="border-b border-border bg-muted/30 py-16 sm:py-20">
				<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
					<Link
						href="/blog"
						className="mb-4 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
					>
						<ArrowLeft className="size-3.5" />
						Voltar ao blog
					</Link>
					<h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
						{category.name}
					</h1>
					{category.description && (
						<p className="mt-3 text-base text-muted-foreground sm:text-lg">
							{category.description}
						</p>
					)}
					<p className="mt-2 text-sm text-muted-foreground">
						{total} {total === 1 ? "artigo" : "artigos"}
					</p>
				</div>
			</section>

			{/* Posts */}
			<section className="py-12 sm:py-16">
				<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
					{posts.length === 0 ? (
						<div className="flex min-h-[200px] items-center justify-center rounded-2xl border border-dashed border-border">
							<p className="text-sm text-muted-foreground">
								Nenhum artigo nesta categoria.
							</p>
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

					{totalPages > 1 && (
						<div className="mt-10 flex items-center justify-center gap-2">
							{page > 1 && (
								<Link
									href={`/blog/categoria/${slug}?page=${page - 1}`}
									className="rounded-lg border border-border px-4 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent"
								>
									Anterior
								</Link>
							)}
							<span className="px-3 text-sm text-muted-foreground">
								{page} de {totalPages}
							</span>
							{page < totalPages && (
								<Link
									href={`/blog/categoria/${slug}?page=${page + 1}`}
									className="rounded-lg border border-border px-4 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent"
								>
									Proximo
								</Link>
							)}
						</div>
					)}
				</div>
			</section>

			{/* JSON-LD */}
			<script
				type="application/ld+json"
				dangerouslySetInnerHTML={{
					__html: JSON.stringify({
						"@context": "https://schema.org",
						"@type": "CollectionPage",
						name: category.name,
						description: category.description ?? "",
						url: `https://zedaapi.com/blog/categoria/${category.slug}`,
					}),
				}}
			/>
		</div>
	);
}
