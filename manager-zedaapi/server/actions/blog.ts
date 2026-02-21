"use server";

import { revalidatePath } from "next/cache";
import { requireAdmin } from "@/lib/auth-server";
import { db } from "@/lib/db";
import { slugify } from "@/lib/slugify";
import { deleteFile } from "@/lib/services/storage/s3-client";
import type { ActionResult } from "@/types";

// =============================================================================
// Blog Posts
// =============================================================================

export async function getAdminBlogPosts(
	page: number = 1,
	search?: string,
	categoryId?: string,
	status?: string,
): Promise<
	ActionResult<{ items: Array<Record<string, unknown>>; total: number }>
> {
	try {
		await requireAdmin();

		const pageSize = 20;
		const where: Record<string, unknown> = {};

		if (search) {
			where.title = { contains: search, mode: "insensitive" as const };
		}
		if (categoryId) {
			where.categoryId = categoryId;
		}
		if (status) {
			where.status = status;
		}

		const [posts, total] = await Promise.all([
			db.blogPost.findMany({
				where,
				skip: (page - 1) * pageSize,
				take: pageSize,
				orderBy: { createdAt: "desc" },
				include: {
					category: true,
					author: { select: { name: true } },
					tags: { include: { tag: true } },
					_count: { select: { media: true } },
				},
			}),
			db.blogPost.count({ where }),
		]);

		return {
			success: true,
			data: {
				items: posts.map((p) => ({
					...p,
					mediaCount: p._count.media,
				})),
				total,
			},
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch blog posts" };
	}
}

export async function getAdminBlogPost(
	id: string,
): Promise<ActionResult<Record<string, unknown>>> {
	try {
		await requireAdmin();

		const post = await db.blogPost.findUnique({
			where: { id },
			include: {
				tags: { include: { tag: true } },
				category: true,
				media: true,
				author: { select: { id: true, name: true, email: true } },
			},
		});

		if (!post) {
			return { success: false, error: "Blog post not found" };
		}

		return { success: true, data: post };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch blog post" };
	}
}

export async function createBlogPost(data: {
	title: string;
	content: string;
	excerpt?: string;
	coverImageUrl?: string;
	categoryId?: string;
	seoTitle?: string;
	seoDescription?: string;
	seoKeywords?: string;
	status?: string;
	tagIds?: string[];
}): Promise<ActionResult<{ id: string }>> {
	try {
		const session = await requireAdmin();

		const slug = slugify(data.title);

		const existing = await db.blogPost.findUnique({ where: { slug } });
		if (existing) {
			return {
				success: false,
				error: "A post with this slug already exists",
			};
		}

		const wordCount = data.content
			.replace(/<[^>]*>/g, "")
			.split(/\s+/)
			.filter(Boolean).length;
		const readingTimeMin = Math.ceil(wordCount / 200);

		const post = await db.blogPost.create({
			data: {
				title: data.title,
				slug,
				content: data.content,
				excerpt: data.excerpt || null,
				coverImageUrl: data.coverImageUrl || null,
				categoryId: data.categoryId || null,
				seoTitle: data.seoTitle || null,
				seoDescription: data.seoDescription || null,
				seoKeywords: data.seoKeywords || null,
				status: data.status || "draft",
				readingTimeMin,
				publishedAt: data.status === "published" ? new Date() : null,
				authorId: session.user.id,
				tags: data.tagIds?.length
					? {
							create: data.tagIds.map((tagId) => ({
								tagId,
							})),
						}
					: undefined,
			},
		});

		revalidatePath("/admin/blog");
		return { success: true, data: { id: post.id } };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to create blog post" };
	}
}

export async function updateBlogPost(
	id: string,
	data: {
		title?: string;
		content?: string;
		excerpt?: string;
		coverImageUrl?: string;
		categoryId?: string;
		seoTitle?: string;
		seoDescription?: string;
		seoKeywords?: string;
		status?: string;
		tagIds?: string[];
	},
): Promise<ActionResult> {
	try {
		await requireAdmin();

		const existing = await db.blogPost.findUnique({ where: { id } });
		if (!existing) {
			return { success: false, error: "Blog post not found" };
		}

		const updateData: Record<string, unknown> = {};

		if (data.title !== undefined) {
			updateData.title = data.title;
			updateData.slug = slugify(data.title);
		}
		if (data.content !== undefined) {
			updateData.content = data.content;
			const wordCount = data.content
				.replace(/<[^>]*>/g, "")
				.split(/\s+/)
				.filter(Boolean).length;
			updateData.readingTimeMin = Math.ceil(wordCount / 200);
		}
		if (data.excerpt !== undefined)
			updateData.excerpt = data.excerpt || null;
		if (data.coverImageUrl !== undefined)
			updateData.coverImageUrl = data.coverImageUrl || null;
		if (data.categoryId !== undefined)
			updateData.categoryId = data.categoryId || null;
		if (data.seoTitle !== undefined)
			updateData.seoTitle = data.seoTitle || null;
		if (data.seoDescription !== undefined)
			updateData.seoDescription = data.seoDescription || null;
		if (data.seoKeywords !== undefined)
			updateData.seoKeywords = data.seoKeywords || null;

		if (data.status !== undefined) {
			updateData.status = data.status;
			if (data.status === "published" && !existing.publishedAt) {
				updateData.publishedAt = new Date();
			}
		}

		if (data.tagIds !== undefined) {
			await db.blogPostTag.deleteMany({ where: { postId: id } });
			if (data.tagIds.length > 0) {
				await db.blogPostTag.createMany({
					data: data.tagIds.map((tagId) => ({
						postId: id,
						tagId,
					})),
				});
			}
		}

		await db.blogPost.update({
			where: { id },
			data: updateData,
		});

		revalidatePath("/admin/blog");
		revalidatePath(`/blog/${existing.slug}`);
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to update blog post" };
	}
}

export async function deleteBlogPost(id: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		await db.blogPost.delete({ where: { id } });

		revalidatePath("/admin/blog");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to delete blog post" };
	}
}

export async function publishBlogPost(id: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		await db.blogPost.update({
			where: { id },
			data: { status: "published", publishedAt: new Date() },
		});

		revalidatePath("/admin/blog");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to publish blog post" };
	}
}

export async function archiveBlogPost(id: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		await db.blogPost.update({
			where: { id },
			data: { status: "archived" },
		});

		revalidatePath("/admin/blog");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to archive blog post" };
	}
}

// =============================================================================
// Blog Categories
// =============================================================================

export async function getAdminBlogCategories(): Promise<
	ActionResult<Array<Record<string, unknown>>>
> {
	try {
		await requireAdmin();

		const categories = await db.blogCategory.findMany({
			orderBy: { sortOrder: "asc" },
			include: { _count: { select: { posts: true } } },
		});

		return { success: true, data: categories };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch blog categories" };
	}
}

export async function createBlogCategory(data: {
	name: string;
	slug: string;
	description?: string;
	sortOrder?: number;
}): Promise<ActionResult> {
	try {
		await requireAdmin();

		const existing = await db.blogCategory.findUnique({
			where: { slug: data.slug },
		});
		if (existing) {
			return {
				success: false,
				error: "A category with this slug already exists",
			};
		}

		await db.blogCategory.create({
			data: {
				name: data.name,
				slug: data.slug,
				description: data.description || null,
				sortOrder: data.sortOrder ?? 0,
			},
		});

		revalidatePath("/admin/blog/categories");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to create blog category" };
	}
}

export async function updateBlogCategory(
	id: string,
	data: {
		name?: string;
		slug?: string;
		description?: string;
		sortOrder?: number;
	},
): Promise<ActionResult> {
	try {
		await requireAdmin();

		await db.blogCategory.update({
			where: { id },
			data: {
				...(data.name !== undefined && { name: data.name }),
				...(data.slug !== undefined && { slug: data.slug }),
				...(data.description !== undefined && {
					description: data.description || null,
				}),
				...(data.sortOrder !== undefined && {
					sortOrder: data.sortOrder,
				}),
			},
		});

		revalidatePath("/admin/blog/categories");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to update blog category" };
	}
}

export async function deleteBlogCategory(id: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		const postsCount = await db.blogPost.count({
			where: { categoryId: id },
		});
		if (postsCount > 0) {
			return {
				success: false,
				error: `Cannot delete category with ${postsCount} post(s). Move or delete them first.`,
			};
		}

		await db.blogCategory.delete({ where: { id } });

		revalidatePath("/admin/blog/categories");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to delete blog category" };
	}
}

// =============================================================================
// Blog Tags
// =============================================================================

export async function getAdminBlogTags(): Promise<
	ActionResult<Array<Record<string, unknown>>>
> {
	try {
		await requireAdmin();

		const tags = await db.blogTag.findMany({
			orderBy: { name: "asc" },
			include: { _count: { select: { posts: true } } },
		});

		return { success: true, data: tags };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch blog tags" };
	}
}

export async function createBlogTag(data: {
	name: string;
}): Promise<ActionResult> {
	try {
		await requireAdmin();

		const slug = slugify(data.name);

		const existing = await db.blogTag.findUnique({ where: { slug } });
		if (existing) {
			return {
				success: false,
				error: "A tag with this name already exists",
			};
		}

		await db.blogTag.create({
			data: { name: data.name, slug },
		});

		revalidatePath("/admin/blog/tags");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to create blog tag" };
	}
}

export async function deleteBlogTag(id: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		await db.blogTag.delete({ where: { id } });

		revalidatePath("/admin/blog/tags");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to delete blog tag" };
	}
}

// =============================================================================
// Blog Media
// =============================================================================

export async function deleteBlogMedia(id: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		const media = await db.blogMedia.findUnique({ where: { id } });
		if (!media) {
			return { success: false, error: "Media not found" };
		}

		if (media.s3Key) {
			try {
				await deleteFile(media.s3Key);
			} catch {
				// Log but don't fail if S3 delete fails
			}
		}

		await db.blogMedia.delete({ where: { id } });

		revalidatePath("/admin/blog");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to delete blog media" };
	}
}
