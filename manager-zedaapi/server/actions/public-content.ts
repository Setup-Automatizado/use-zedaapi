"use server";

import { db } from "@/lib/db";

// =============================================================================
// Blog (Public)
// =============================================================================

export async function getPublishedBlogPosts(
	page: number = 1,
	categorySlug?: string,
): Promise<{
	items: Array<Record<string, unknown>>;
	total: number;
	totalPages: number;
}> {
	try {
		const pageSize = 12;
		const where: Record<string, unknown> = { status: "published" };

		if (categorySlug) {
			where.category = { slug: categorySlug };
		}

		const [posts, total] = await Promise.all([
			db.blogPost.findMany({
				where,
				skip: (page - 1) * pageSize,
				take: pageSize,
				orderBy: { publishedAt: "desc" },
				include: {
					category: true,
					author: { select: { name: true } },
					tags: { include: { tag: true } },
				},
			}),
			db.blogPost.count({ where }),
		]);

		return {
			items: posts,
			total,
			totalPages: Math.ceil(total / pageSize),
		};
	} catch {
		return { items: [], total: 0, totalPages: 0 };
	}
}

export async function getBlogPostBySlug(
	slug: string,
): Promise<Record<string, unknown> | null> {
	try {
		const post = await db.blogPost.findFirst({
			where: { slug, status: "published" },
			include: {
				category: true,
				author: { select: { name: true } },
				tags: { include: { tag: true } },
				media: true,
			},
		});

		if (!post) return null;

		// Increment view count (fire and forget)
		db.blogPost
			.update({
				where: { id: post.id },
				data: { viewCount: { increment: 1 } },
			})
			.catch(() => {});

		return post;
	} catch {
		return null;
	}
}

export async function getBlogCategories(): Promise<
	Array<Record<string, unknown>>
> {
	try {
		const categories = await db.blogCategory.findMany({
			where: {
				posts: { some: { status: "published" } },
			},
			orderBy: { sortOrder: "asc" },
			include: {
				_count: {
					select: { posts: { where: { status: "published" } } },
				},
			},
		});

		return categories;
	} catch {
		return [];
	}
}

export async function getPopularBlogPosts(
	limit: number = 5,
): Promise<Array<Record<string, unknown>>> {
	try {
		const posts = await db.blogPost.findMany({
			where: { status: "published" },
			orderBy: { viewCount: "desc" },
			take: limit,
			include: {
				category: true,
				author: { select: { name: true } },
				tags: { include: { tag: true } },
			},
		});

		return posts;
	} catch {
		return [];
	}
}

export async function getRelatedBlogPosts(
	postId: string,
	limit: number = 3,
): Promise<Array<Record<string, unknown>>> {
	try {
		const post = await db.blogPost.findUnique({
			where: { id: postId },
			select: { categoryId: true },
		});

		if (!post?.categoryId) return [];

		const posts = await db.blogPost.findMany({
			where: {
				status: "published",
				categoryId: post.categoryId,
				id: { not: postId },
			},
			orderBy: { publishedAt: "desc" },
			take: limit,
			include: {
				category: true,
				author: { select: { name: true } },
				tags: { include: { tag: true } },
			},
		});

		return posts;
	} catch {
		return [];
	}
}

// =============================================================================
// Support (Public)
// =============================================================================

export async function getSupportCategories(): Promise<
	Array<Record<string, unknown>>
> {
	try {
		const categories = await db.supportCategory.findMany({
			orderBy: { sortOrder: "asc" },
			include: {
				_count: {
					select: {
						articles: { where: { status: "published" } },
					},
				},
			},
		});

		return categories.filter(
			(c) => (c._count as { articles: number }).articles > 0,
		);
	} catch {
		return [];
	}
}

export async function getSupportArticlesByCategory(
	categorySlug: string,
): Promise<{
	category: Record<string, unknown> | null;
	articles: Array<Record<string, unknown>>;
}> {
	try {
		const category = await db.supportCategory.findUnique({
			where: { slug: categorySlug },
		});

		if (!category) {
			return { category: null, articles: [] };
		}

		const articles = await db.supportArticle.findMany({
			where: {
				categoryId: category.id,
				status: "published",
			},
			orderBy: { sortOrder: "asc" },
		});

		return { category, articles };
	} catch {
		return { category: null, articles: [] };
	}
}

export async function getSupportArticleBySlug(
	slug: string,
): Promise<Record<string, unknown> | null> {
	try {
		const article = await db.supportArticle.findFirst({
			where: { slug, status: "published" },
			include: { category: true },
		});

		if (!article) return null;

		// Increment view count (fire and forget)
		db.supportArticle
			.update({
				where: { id: article.id },
				data: { viewCount: { increment: 1 } },
			})
			.catch(() => {});

		return article;
	} catch {
		return null;
	}
}

// =============================================================================
// Glossary (Public)
// =============================================================================

export async function getGlossaryTerms(): Promise<
	Array<Record<string, unknown>>
> {
	try {
		const terms = await db.glossaryTerm.findMany({
			where: { status: "published" },
			orderBy: { term: "asc" },
		});

		return terms;
	} catch {
		return [];
	}
}

export async function getGlossaryTermBySlug(
	slug: string,
): Promise<Record<string, unknown> | null> {
	try {
		const term = await db.glossaryTerm.findFirst({
			where: { slug, status: "published" },
		});

		return term;
	} catch {
		return null;
	}
}
