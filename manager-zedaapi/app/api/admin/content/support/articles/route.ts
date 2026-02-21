import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import { requireContentApiKey, parsePaginationParams } from "@/lib/api-auth";
import { db } from "@/lib/db";
import { slugify } from "@/lib/slugify";

const PAGE_SIZE = 20;

export async function GET(req: NextRequest) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const url = new URL(req.url);
		const { page, search } = parsePaginationParams(url);
		const categoryId = url.searchParams.get("categoryId") || undefined;

		const where: {
			title?: { contains: string; mode: "insensitive" };
			categoryId?: string;
		} = {};

		if (search) {
			where.title = { contains: search, mode: "insensitive" };
		}
		if (categoryId) {
			where.categoryId = categoryId;
		}

		const [articles, total] = await Promise.all([
			db.supportArticle.findMany({
				where,
				include: { category: true },
				orderBy: [{ sortOrder: "asc" }, { createdAt: "desc" }],
				skip: (page - 1) * PAGE_SIZE,
				take: PAGE_SIZE,
			}),
			db.supportArticle.count({ where }),
		]);

		return NextResponse.json({
			data: articles,
			pagination: {
				page,
				pageSize: PAGE_SIZE,
				total,
				totalPages: Math.ceil(total / PAGE_SIZE),
			},
		});
	} catch (error) {
		console.error("Failed to list support articles:", error);
		return NextResponse.json(
			{ error: "Failed to list articles" },
			{ status: 500 },
		);
	}
}

export async function POST(req: NextRequest) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const body = (await req.json()) as {
			title?: string;
			content?: string;
			excerpt?: string;
			categoryId?: string;
			seoTitle?: string;
			seoDescription?: string;
			status?: string;
			sortOrder?: number;
		};

		if (!body.title || !body.content || !body.categoryId) {
			return NextResponse.json(
				{
					error: "Missing required fields: title, content, categoryId",
				},
				{ status: 400 },
			);
		}

		const slug = slugify(body.title);

		const existing = await db.supportArticle.findUnique({
			where: { slug },
		});
		if (existing) {
			return NextResponse.json(
				{ error: "An article with this slug already exists" },
				{ status: 409 },
			);
		}

		const article = await db.supportArticle.create({
			data: {
				title: body.title,
				slug,
				content: body.content,
				excerpt: body.excerpt ?? undefined,
				categoryId: body.categoryId,
				seoTitle: body.seoTitle ?? undefined,
				seoDescription: body.seoDescription ?? undefined,
				status: body.status ?? "draft",
				sortOrder: body.sortOrder ?? 0,
			},
			include: { category: true },
		});

		revalidatePath("/suporte");
		return NextResponse.json(article, { status: 201 });
	} catch (error) {
		console.error("Failed to create support article:", error);
		return NextResponse.json(
			{ error: "Failed to create article" },
			{ status: 500 },
		);
	}
}
