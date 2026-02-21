import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import { requireContentApiKey, parsePaginationParams } from "@/lib/api-auth";
import { db } from "@/lib/db";
import { createLogger } from "@/lib/logger";
import { slugify } from "@/lib/slugify";

const log = createLogger("api:blog-posts");

// GET /api/admin/content/blog/posts?page=1&search=&categoryId=&status=
export async function GET(req: NextRequest) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const url = new URL(req.url);
		const { page, search } = parsePaginationParams(url);
		const categoryId = url.searchParams.get("categoryId") || undefined;
		const status = url.searchParams.get("status") || undefined;
		const pageSize = 20;

		const where: Record<string, unknown> = {};
		if (search) where.title = { contains: search, mode: "insensitive" };
		if (categoryId) where.categoryId = categoryId;
		if (status) where.status = status;

		const [posts, total] = await Promise.all([
			db.blogPost.findMany({
				where,
				skip: (page - 1) * pageSize,
				take: pageSize,
				orderBy: { createdAt: "desc" },
				include: {
					category: true,
					author: { select: { id: true, name: true } },
					tags: { include: { tag: true } },
				},
			}),
			db.blogPost.count({ where }),
		]);

		return NextResponse.json({
			items: posts,
			total,
			page,
			pageSize,
			totalPages: Math.ceil(total / pageSize),
		});
	} catch (error) {
		log.error("Failed to list blog posts", { error });
		return NextResponse.json(
			{ error: "Failed to list blog posts" },
			{ status: 500 },
		);
	}
}

// POST /api/admin/content/blog/posts
export async function POST(req: NextRequest) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const body = (await req.json()) as Record<string, unknown>;

		if (
			typeof body.title !== "string" ||
			!body.title.trim() ||
			typeof body.content !== "string" ||
			!body.content.trim()
		) {
			return NextResponse.json(
				{ error: "title and content are required" },
				{ status: 400 },
			);
		}

		const slug = slugify(body.title);
		const existing = await db.blogPost.findUnique({ where: { slug } });
		if (existing) {
			return NextResponse.json(
				{ error: "A post with this slug already exists" },
				{ status: 409 },
			);
		}

		const wordCount = body.content
			.replace(/<[^>]*>/g, "")
			.split(/\s+/)
			.filter(Boolean).length;
		const readingTimeMin = Math.ceil(wordCount / 200);

		const admin = await db.user.findFirst({ where: { role: "admin" } });
		if (!admin) {
			return NextResponse.json(
				{ error: "No admin user found to assign as author" },
				{ status: 500 },
			);
		}

		const tagIds = Array.isArray(body.tagIds) ? body.tagIds : [];
		const postStatus =
			typeof body.status === "string" ? body.status : "draft";

		const post = await db.blogPost.create({
			data: {
				title: body.title,
				slug,
				content: body.content,
				excerpt: typeof body.excerpt === "string" ? body.excerpt : null,
				coverImageUrl:
					typeof body.coverImageUrl === "string"
						? body.coverImageUrl
						: null,
				categoryId:
					typeof body.categoryId === "string"
						? body.categoryId
						: null,
				seoTitle:
					typeof body.seoTitle === "string" ? body.seoTitle : null,
				seoDescription:
					typeof body.seoDescription === "string"
						? body.seoDescription
						: null,
				seoKeywords:
					typeof body.seoKeywords === "string"
						? body.seoKeywords
						: null,
				status: postStatus,
				readingTimeMin,
				publishedAt: postStatus === "published" ? new Date() : null,
				authorId: admin.id,
				tags: tagIds.length
					? {
							create: tagIds.map((tagId: unknown) => ({
								tagId: String(tagId),
							})),
						}
					: undefined,
			},
			include: { category: true, tags: { include: { tag: true } } },
		});

		revalidatePath("/blog");
		revalidatePath("/admin/blog");

		return NextResponse.json(post, { status: 201 });
	} catch (error) {
		log.error("Failed to create blog post", { error });
		return NextResponse.json(
			{ error: "Failed to create blog post" },
			{ status: 500 },
		);
	}
}
