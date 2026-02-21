"use server";

import { revalidatePath } from "next/cache";
import { requireAdmin } from "@/lib/auth-server";
import { db } from "@/lib/db";
import { slugify } from "@/lib/slugify";
import type { ActionResult } from "@/types";

// =============================================================================
// Support Categories
// =============================================================================

export async function getAdminSupportCategories(): Promise<
	ActionResult<Array<Record<string, unknown>>>
> {
	try {
		await requireAdmin();

		const categories = await db.supportCategory.findMany({
			orderBy: { sortOrder: "asc" },
			include: { _count: { select: { articles: true } } },
		});

		return { success: true, data: categories };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch support categories" };
	}
}

export async function createSupportCategory(data: {
	name: string;
	slug: string;
	description?: string;
	icon?: string;
	sortOrder?: number;
}): Promise<ActionResult> {
	try {
		await requireAdmin();

		const existing = await db.supportCategory.findUnique({
			where: { slug: data.slug },
		});
		if (existing) {
			return {
				success: false,
				error: "A category with this slug already exists",
			};
		}

		await db.supportCategory.create({
			data: {
				name: data.name,
				slug: data.slug,
				description: data.description || null,
				icon: data.icon || null,
				sortOrder: data.sortOrder ?? 0,
			},
		});

		revalidatePath("/admin/support");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to create support category" };
	}
}

export async function updateSupportCategory(
	id: string,
	data: {
		name?: string;
		slug?: string;
		description?: string;
		icon?: string;
		sortOrder?: number;
	},
): Promise<ActionResult> {
	try {
		await requireAdmin();

		await db.supportCategory.update({
			where: { id },
			data: {
				...(data.name !== undefined && { name: data.name }),
				...(data.slug !== undefined && { slug: data.slug }),
				...(data.description !== undefined && {
					description: data.description || null,
				}),
				...(data.icon !== undefined && { icon: data.icon || null }),
				...(data.sortOrder !== undefined && {
					sortOrder: data.sortOrder,
				}),
			},
		});

		revalidatePath("/admin/support");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to update support category" };
	}
}

export async function deleteSupportCategory(id: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		const articlesCount = await db.supportArticle.count({
			where: { categoryId: id },
		});
		if (articlesCount > 0) {
			return {
				success: false,
				error: `Cannot delete category with ${articlesCount} article(s). Move or delete them first.`,
			};
		}

		await db.supportCategory.delete({ where: { id } });

		revalidatePath("/admin/support");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to delete support category" };
	}
}

// =============================================================================
// Support Articles
// =============================================================================

export async function getAdminSupportArticles(
	page: number = 1,
	search?: string,
	categoryId?: string,
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

		const [articles, total] = await Promise.all([
			db.supportArticle.findMany({
				where,
				skip: (page - 1) * pageSize,
				take: pageSize,
				orderBy: { sortOrder: "asc" },
				include: { category: true },
			}),
			db.supportArticle.count({ where }),
		]);

		return {
			success: true,
			data: { items: articles, total },
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch support articles" };
	}
}

export async function getAdminSupportArticle(
	id: string,
): Promise<ActionResult<Record<string, unknown>>> {
	try {
		await requireAdmin();

		const article = await db.supportArticle.findUnique({
			where: { id },
			include: { category: true },
		});

		if (!article) {
			return { success: false, error: "Support article not found" };
		}

		return { success: true, data: article };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch support article" };
	}
}

export async function createSupportArticle(data: {
	title: string;
	content: string;
	excerpt?: string;
	categoryId: string;
	seoTitle?: string;
	seoDescription?: string;
	status?: string;
	sortOrder?: number;
}): Promise<ActionResult<{ id: string }>> {
	try {
		await requireAdmin();

		const slug = slugify(data.title);

		const existing = await db.supportArticle.findUnique({
			where: { slug },
		});
		if (existing) {
			return {
				success: false,
				error: "An article with this slug already exists",
			};
		}

		const article = await db.supportArticle.create({
			data: {
				title: data.title,
				slug,
				content: data.content,
				excerpt: data.excerpt || null,
				categoryId: data.categoryId,
				seoTitle: data.seoTitle || null,
				seoDescription: data.seoDescription || null,
				status: data.status || "draft",
				sortOrder: data.sortOrder ?? 0,
			},
		});

		revalidatePath("/admin/support");
		return { success: true, data: { id: article.id } };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to create support article" };
	}
}

export async function updateSupportArticle(
	id: string,
	data: {
		title?: string;
		content?: string;
		excerpt?: string;
		categoryId?: string;
		seoTitle?: string;
		seoDescription?: string;
		status?: string;
		sortOrder?: number;
	},
): Promise<ActionResult> {
	try {
		await requireAdmin();

		const updateData: Record<string, unknown> = {};

		if (data.title !== undefined) {
			updateData.title = data.title;
			updateData.slug = slugify(data.title);
		}
		if (data.content !== undefined) updateData.content = data.content;
		if (data.excerpt !== undefined)
			updateData.excerpt = data.excerpt || null;
		if (data.categoryId !== undefined)
			updateData.categoryId = data.categoryId;
		if (data.seoTitle !== undefined)
			updateData.seoTitle = data.seoTitle || null;
		if (data.seoDescription !== undefined)
			updateData.seoDescription = data.seoDescription || null;
		if (data.status !== undefined) updateData.status = data.status;
		if (data.sortOrder !== undefined) updateData.sortOrder = data.sortOrder;

		await db.supportArticle.update({
			where: { id },
			data: updateData,
		});

		revalidatePath("/admin/support");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to update support article" };
	}
}

export async function deleteSupportArticle(id: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		await db.supportArticle.delete({ where: { id } });

		revalidatePath("/admin/support");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to delete support article" };
	}
}

// =============================================================================
// Glossary Terms
// =============================================================================

export async function getAdminGlossaryTerms(
	page: number = 1,
	search?: string,
	letter?: string,
): Promise<
	ActionResult<{ items: Array<Record<string, unknown>>; total: number }>
> {
	try {
		await requireAdmin();

		const pageSize = 20;
		const where: Record<string, unknown> = {};

		if (search) {
			where.term = { contains: search, mode: "insensitive" as const };
		}
		if (letter) {
			where.term = {
				...(where.term as object),
				startsWith: letter,
				mode: "insensitive" as const,
			};
		}

		const [terms, total] = await Promise.all([
			db.glossaryTerm.findMany({
				where,
				skip: (page - 1) * pageSize,
				take: pageSize,
				orderBy: { term: "asc" },
			}),
			db.glossaryTerm.count({ where }),
		]);

		return {
			success: true,
			data: { items: terms, total },
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch glossary terms" };
	}
}

export async function getAdminGlossaryTerm(
	id: string,
): Promise<ActionResult<Record<string, unknown>>> {
	try {
		await requireAdmin();

		const term = await db.glossaryTerm.findUnique({ where: { id } });

		if (!term) {
			return { success: false, error: "Glossary term not found" };
		}

		return { success: true, data: term };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch glossary term" };
	}
}

export async function createGlossaryTerm(data: {
	term: string;
	definition: string;
	content?: string;
	seoTitle?: string;
	seoDescription?: string;
	status?: string;
	relatedSlugs?: string[];
}): Promise<ActionResult<{ id: string }>> {
	try {
		await requireAdmin();

		const slug = slugify(data.term);

		const existing = await db.glossaryTerm.findUnique({ where: { slug } });
		if (existing) {
			return {
				success: false,
				error: "A term with this slug already exists",
			};
		}

		const term = await db.glossaryTerm.create({
			data: {
				term: data.term,
				slug,
				definition: data.definition,
				content: data.content || null,
				seoTitle: data.seoTitle || null,
				seoDescription: data.seoDescription || null,
				status: data.status || "draft",
				relatedSlugs: data.relatedSlugs ?? undefined,
			},
		});

		revalidatePath("/admin/glossary");
		return { success: true, data: { id: term.id } };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to create glossary term" };
	}
}

export async function updateGlossaryTerm(
	id: string,
	data: {
		term?: string;
		definition?: string;
		content?: string;
		seoTitle?: string;
		seoDescription?: string;
		status?: string;
		relatedSlugs?: string[];
	},
): Promise<ActionResult> {
	try {
		await requireAdmin();

		const updateData: Record<string, unknown> = {};

		if (data.term !== undefined) {
			updateData.term = data.term;
			updateData.slug = slugify(data.term);
		}
		if (data.definition !== undefined)
			updateData.definition = data.definition;
		if (data.content !== undefined)
			updateData.content = data.content || null;
		if (data.seoTitle !== undefined)
			updateData.seoTitle = data.seoTitle || null;
		if (data.seoDescription !== undefined)
			updateData.seoDescription = data.seoDescription || null;
		if (data.status !== undefined) updateData.status = data.status;
		if (data.relatedSlugs !== undefined)
			updateData.relatedSlugs = data.relatedSlugs;

		await db.glossaryTerm.update({
			where: { id },
			data: updateData,
		});

		revalidatePath("/admin/glossary");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to update glossary term" };
	}
}

export async function deleteGlossaryTerm(id: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		await db.glossaryTerm.delete({ where: { id } });

		revalidatePath("/admin/glossary");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to delete glossary term" };
	}
}
