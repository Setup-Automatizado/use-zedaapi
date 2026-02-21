import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import { requireContentApiKey } from "@/lib/api-auth";
import { db } from "@/lib/db";
import { createLogger } from "@/lib/logger";
import { slugify } from "@/lib/slugify";

const log = createLogger("api:blog-posts");

// GET /api/admin/content/blog/posts/[id]
export async function GET(
	req: NextRequest,
	{ params }: { params: Promise<{ id: string }> },
) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const { id } = await params;

		const post = await db.blogPost.findUnique({
			where: { id },
			include: {
				category: true,
				author: { select: { id: true, name: true } },
				tags: { include: { tag: true } },
				media: true,
			},
		});

		if (!post) {
			return NextResponse.json(
				{ error: "Post not found" },
				{ status: 404 },
			);
		}

		return NextResponse.json(post);
	} catch (error) {
		log.error("Failed to get blog post", { error });
		return NextResponse.json(
			{ error: "Failed to get blog post" },
			{ status: 500 },
		);
	}
}

// PATCH /api/admin/content/blog/posts/[id]
export async function PATCH(
	req: NextRequest,
	{ params }: { params: Promise<{ id: string }> },
) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const { id } = await params;
		const body = (await req.json()) as Record<string, unknown>;

		const existing = await db.blogPost.findUnique({ where: { id } });
		if (!existing) {
			return NextResponse.json(
				{ error: "Post not found" },
				{ status: 404 },
			);
		}

		const data: Record<string, unknown> = {};

		if (typeof body.title === "string" && body.title.trim()) {
			data.title = body.title;
			const newSlug = slugify(body.title);
			if (newSlug !== existing.slug) {
				const slugTaken = await db.blogPost.findUnique({
					where: { slug: newSlug },
				});
				if (slugTaken) {
					return NextResponse.json(
						{ error: "A post with this slug already exists" },
						{ status: 409 },
					);
				}
				data.slug = newSlug;
			}
		}

		if (typeof body.content === "string") {
			data.content = body.content;
			const wordCount = body.content
				.replace(/<[^>]*>/g, "")
				.split(/\s+/)
				.filter(Boolean).length;
			data.readingTimeMin = Math.ceil(wordCount / 200);
		}

		if (typeof body.excerpt === "string") data.excerpt = body.excerpt;
		if (typeof body.coverImageUrl === "string")
			data.coverImageUrl = body.coverImageUrl;
		if (typeof body.categoryId === "string")
			data.categoryId = body.categoryId || null;
		if (typeof body.seoTitle === "string") data.seoTitle = body.seoTitle;
		if (typeof body.seoDescription === "string")
			data.seoDescription = body.seoDescription;
		if (typeof body.seoKeywords === "string")
			data.seoKeywords = body.seoKeywords;

		if (typeof body.status === "string") {
			data.status = body.status;
			if (body.status === "published" && !existing.publishedAt) {
				data.publishedAt = new Date();
			}
		}

		// Handle tag updates: delete existing, create new
		if (Array.isArray(body.tagIds)) {
			await db.blogPostTag.deleteMany({ where: { postId: id } });
			if (body.tagIds.length > 0) {
				await db.blogPostTag.createMany({
					data: body.tagIds.map((tagId: unknown) => ({
						postId: id,
						tagId: String(tagId),
					})),
				});
			}
		}

		const post = await db.blogPost.update({
			where: { id },
			data,
			include: {
				category: true,
				author: { select: { id: true, name: true } },
				tags: { include: { tag: true } },
			},
		});

		revalidatePath("/blog");
		revalidatePath(`/blog/${post.slug}`);
		revalidatePath("/admin/blog");

		return NextResponse.json(post);
	} catch (error) {
		log.error("Failed to update blog post", { error });
		return NextResponse.json(
			{ error: "Failed to update blog post" },
			{ status: 500 },
		);
	}
}

// DELETE /api/admin/content/blog/posts/[id]
export async function DELETE(
	req: NextRequest,
	{ params }: { params: Promise<{ id: string }> },
) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const { id } = await params;

		const existing = await db.blogPost.findUnique({ where: { id } });
		if (!existing) {
			return NextResponse.json(
				{ error: "Post not found" },
				{ status: 404 },
			);
		}

		await db.blogPost.delete({ where: { id } });

		revalidatePath("/blog");
		revalidatePath(`/blog/${existing.slug}`);
		revalidatePath("/admin/blog");

		return NextResponse.json({ success: true });
	} catch (error) {
		log.error("Failed to delete blog post", { error });
		return NextResponse.json(
			{ error: "Failed to delete blog post" },
			{ status: 500 },
		);
	}
}
