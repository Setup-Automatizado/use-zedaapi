import { ArrowLeft, Calendar, Clock, Eye } from "lucide-react";
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { MediaRenderer } from "@/components/blog/media-renderer";
import { PostCard } from "@/components/blog/post-card";
import { PostContent } from "@/components/blog/post-content";
import { ShareButtons } from "@/components/blog/share-buttons";
import { TableOfContents } from "@/components/blog/toc";
import { Badge } from "@/components/ui/badge";
import { db } from "@/lib/db";

interface BlogPostPageProps {
	params: Promise<{ slug: string }>;
}

export async function generateMetadata({
	params,
}: BlogPostPageProps): Promise<Metadata> {
	const { slug } = await params;

	const post = await db.blogPost.findFirst({
		where: { slug, status: "published" },
		select: {
			title: true,
			excerpt: true,
			seoTitle: true,
			seoDescription: true,
			seoKeywords: true,
			coverImageUrl: true,
			slug: true,
		},
	});

	if (!post) return { title: "Post não encontrado" };

	const title = post.seoTitle || post.title;
	const description = post.seoDescription || post.excerpt || "";

	return {
		title: `${title} - Blog Zé da API`,
		description,
		keywords: post.seoKeywords?.split(",").map((k) => k.trim()),
		openGraph: {
			title,
			description,
			url: `https://zedaapi.com/blog/${post.slug}`,
			type: "article",
			images: post.coverImageUrl ? [post.coverImageUrl] : [],
		},
		twitter: {
			card: "summary_large_image",
			title,
			description,
			images: post.coverImageUrl ? [post.coverImageUrl] : [],
		},
		alternates: {
			canonical: `https://zedaapi.com/blog/${post.slug}`,
		},
	};
}

export default async function BlogPostPage({ params }: BlogPostPageProps) {
	const { slug } = await params;

	const post = await db.blogPost.findFirst({
		where: { slug, status: "published" },
		include: {
			category: { select: { name: true, slug: true } },
			author: { select: { name: true, image: true } },
			tags: { include: { tag: true } },
			media: true,
		},
	});

	if (!post) notFound();

	// Increment view count (fire and forget)
	db.blogPost
		.update({
			where: { id: post.id },
			data: { viewCount: { increment: 1 } },
		})
		.catch(() => {});

	// Related posts (same category)
	const relatedPosts = post.categoryId
		? await db.blogPost.findMany({
				where: {
					status: "published",
					categoryId: post.categoryId,
					id: { not: post.id },
				},
				include: {
					category: { select: { name: true, slug: true } },
					author: { select: { name: true } },
				},
				orderBy: { publishedAt: "desc" },
				take: 3,
			})
		: [];

	const postUrl = `https://zedaapi.com/blog/${post.slug}`;

	const jsonLd = {
		"@context": "https://schema.org",
		"@type": "BlogPosting",
		headline: post.title,
		description: post.excerpt ?? "",
		url: postUrl,
		datePublished: post.publishedAt?.toISOString(),
		dateModified: post.updatedAt.toISOString(),
		author: {
			"@type": "Person",
			name: post.author.name,
		},
		publisher: {
			"@type": "Organization",
			name: "Zé da API",
			url: "https://zedaapi.com",
		},
		image: post.coverImageUrl ?? undefined,
		wordCount: post.content.split(/\s+/).length,
		mainEntityOfPage: {
			"@type": "WebPage",
			"@id": postUrl,
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

			{/* Cover Hero */}
			{post.coverImageUrl && (
				<section className="relative h-64 overflow-hidden sm:h-80 lg:h-96">
					<img
						src={post.coverImageUrl}
						alt={post.title}
						className="size-full object-cover"
					/>
					<div className="absolute inset-0 bg-gradient-to-t from-background via-background/60 to-transparent" />
				</section>
			)}

			<section className="py-10 sm:py-14">
				<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
					<div className="mx-auto grid max-w-7xl gap-10 lg:grid-cols-[1fr_260px]">
						{/* Main content */}
						<article className="min-w-0">
							{/* Back link */}
							<Link
								href="/blog"
								className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
							>
								<ArrowLeft className="size-3.5" />
								Voltar ao blog
							</Link>

							{/* Meta */}
							<div className="mb-6 flex flex-wrap items-center gap-3">
								{post.category && (
									<Link
										href={`/blog?categoria=${post.category.slug}`}
									>
										<Badge variant="secondary">
											{post.category.name}
										</Badge>
									</Link>
								)}
								{post.tags.map(({ tag }) => (
									<Badge
										key={tag.id}
										variant="outline"
										className="text-xs"
									>
										{tag.name}
									</Badge>
								))}
							</div>

							<h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
								{post.title}
							</h1>

							<div className="mt-4 flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
								<div className="flex items-center gap-2">
									{post.author.image ? (
										<img
											src={post.author.image}
											alt={post.author.name}
											className="size-7 rounded-full"
										/>
									) : (
										<div className="flex size-7 items-center justify-center rounded-full bg-primary/10 text-xs font-medium text-primary">
											{post.author.name
												.charAt(0)
												.toUpperCase()}
										</div>
									)}
									<span className="font-medium text-foreground/80">
										{post.author.name}
									</span>
								</div>
								{post.publishedAt && (
									<span className="flex items-center gap-1">
										<Calendar className="size-3.5" />
										{new Date(
											post.publishedAt,
										).toLocaleDateString("pt-BR", {
											day: "2-digit",
											month: "long",
											year: "numeric",
										})}
									</span>
								)}
								{post.readingTimeMin > 0 && (
									<span className="flex items-center gap-1">
										<Clock className="size-3.5" />
										{post.readingTimeMin} min de leitura
									</span>
								)}
								<span className="flex items-center gap-1">
									<Eye className="size-3.5" />
									{post.viewCount} visualizacoes
								</span>
							</div>

							{/* Content */}
							<div className="mt-8 border-t border-border pt-8">
								<PostContent content={post.content} />
							</div>

							{/* Media gallery */}
							{post.media.length > 0 && (
								<div className="mt-8">
									{post.media.map((m) => (
										<MediaRenderer
											key={m.id}
											type={m.type}
											url={m.url}
											alt={m.alt}
											caption={m.caption}
										/>
									))}
								</div>
							)}

							{/* Share */}
							<div className="mt-8 border-t border-border pt-6">
								<ShareButtons
									url={postUrl}
									title={post.title}
								/>
							</div>
						</article>

						{/* Sidebar */}
						<aside className="hidden lg:block">
							<div className="sticky top-24 space-y-8">
								<TableOfContents content={post.content} />
							</div>
						</aside>
					</div>

					{/* Related posts */}
					{relatedPosts.length > 0 && (
						<div className="mt-16 border-t border-border pt-12">
							<h2 className="mb-6 text-xl font-semibold text-foreground">
								Artigos relacionados
							</h2>
							<div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
								{relatedPosts.map((related) => (
									<PostCard
										key={related.id}
										slug={related.slug}
										title={related.title}
										excerpt={related.excerpt}
										coverImageUrl={related.coverImageUrl}
										categoryName={
											related.category?.name ?? null
										}
										categorySlug={
											related.category?.slug ?? null
										}
										authorName={related.author.name}
										publishedAt={related.publishedAt}
										readingTimeMin={related.readingTimeMin}
										viewCount={related.viewCount}
									/>
								))}
							</div>
						</div>
					)}
				</div>
			</section>
		</div>
	);
}
