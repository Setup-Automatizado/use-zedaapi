import type { Metadata } from "next";
import { notFound } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, ChevronRight } from "lucide-react";
import { PostContent } from "@/components/blog/post-content";
import { db } from "@/lib/db";

interface SupportArticlePageProps {
	params: Promise<{ slug: string }>;
}

export async function generateMetadata({
	params,
}: SupportArticlePageProps): Promise<Metadata> {
	const { slug } = await params;

	const article = await db.supportArticle.findFirst({
		where: { slug, status: "published" },
		select: {
			title: true,
			excerpt: true,
			seoTitle: true,
			seoDescription: true,
			slug: true,
		},
	});

	if (!article) return { title: "Artigo nao encontrado" };

	const title = article.seoTitle || article.title;
	const description = article.seoDescription || article.excerpt || "";

	return {
		title: `${title} - Suporte Zé da API`,
		description,
		alternates: {
			canonical: `https://zedaapi.com/suporte/${article.slug}`,
		},
	};
}

export default async function SupportArticlePage({
	params,
}: SupportArticlePageProps) {
	const { slug } = await params;

	const article = await db.supportArticle.findFirst({
		where: { slug, status: "published" },
		include: {
			category: { select: { name: true, slug: true } },
		},
	});

	if (!article) notFound();

	// Increment view count (fire and forget)
	db.supportArticle
		.update({
			where: { id: article.id },
			data: { viewCount: { increment: 1 } },
		})
		.catch(() => {});

	// Related articles in same category
	const relatedArticles = await db.supportArticle.findMany({
		where: {
			status: "published",
			categoryId: article.categoryId,
			id: { not: article.id },
		},
		orderBy: { sortOrder: "asc" },
		take: 5,
		select: { title: true, slug: true },
	});

	const jsonLd = {
		"@context": "https://schema.org",
		"@type": "Article",
		headline: article.title,
		description: article.excerpt ?? "",
		url: `https://zedaapi.com/suporte/${article.slug}`,
		dateModified: article.updatedAt.toISOString(),
		publisher: {
			"@type": "Organization",
			name: "Zé da API",
			url: "https://zedaapi.com",
		},
	};

	return (
		<div>
			<script
				type="application/ld+json"
				dangerouslySetInnerHTML={{
					__html: JSON.stringify(jsonLd),
				}}
			/>

			<section className="py-10 sm:py-14">
				<div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
					{/* Breadcrumb */}
					<nav className="mb-6 flex items-center gap-1.5 text-sm text-muted-foreground">
						<Link
							href="/suporte"
							className="transition-colors hover:text-foreground"
						>
							Suporte
						</Link>
						<ChevronRight className="size-3.5" />
						{article.category && (
							<>
								<span>{article.category.name}</span>
								<ChevronRight className="size-3.5" />
							</>
						)}
						<span className="text-foreground">{article.title}</span>
					</nav>

					<article>
						<h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
							{article.title}
						</h1>

						{article.excerpt && (
							<p className="mt-3 text-base text-muted-foreground">
								{article.excerpt}
							</p>
						)}

						<div className="mt-8 border-t border-border pt-8">
							<PostContent content={article.content} />
						</div>
					</article>

					{/* Related articles */}
					{relatedArticles.length > 0 && (
						<div className="mt-10 rounded-2xl border border-border bg-muted/30 p-6">
							<h3 className="text-sm font-semibold text-foreground">
								Artigos relacionados
							</h3>
							<ul className="mt-3 space-y-2">
								{relatedArticles.map((a) => (
									<li key={a.slug}>
										<Link
											href={`/suporte/${a.slug}`}
											className="text-sm text-muted-foreground underline-offset-4 transition-colors hover:text-primary hover:underline"
										>
											{a.title}
										</Link>
									</li>
								))}
							</ul>
						</div>
					)}

					{/* Back */}
					<div className="mt-8">
						<Link
							href="/suporte"
							className="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
						>
							<ArrowLeft className="size-3.5" />
							Voltar ao suporte
						</Link>
					</div>
				</div>
			</section>
		</div>
	);
}
