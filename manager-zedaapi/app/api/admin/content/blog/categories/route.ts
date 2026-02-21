import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import { requireContentApiKey } from "@/lib/api-auth";
import { db } from "@/lib/db";
import { createLogger } from "@/lib/logger";
import { slugify } from "@/lib/slugify";

const log = createLogger("api:blog-categories");

// GET /api/admin/content/blog/categories
export async function GET(req: NextRequest) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const categories = await db.blogCategory.findMany({
			orderBy: { sortOrder: "asc" },
			include: {
				_count: { select: { posts: true } },
			},
		});

		return NextResponse.json(categories);
	} catch (error) {
		log.error("Failed to list blog categories", { error });
		return NextResponse.json(
			{ error: "Failed to list blog categories" },
			{ status: 500 },
		);
	}
}

// POST /api/admin/content/blog/categories
export async function POST(req: NextRequest) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const body = (await req.json()) as Record<string, unknown>;

		if (typeof body.name !== "string" || !body.name.trim()) {
			return NextResponse.json(
				{ error: "name is required" },
				{ status: 400 },
			);
		}

		const slug = slugify(body.name);
		const existing = await db.blogCategory.findUnique({
			where: { slug },
		});
		if (existing) {
			return NextResponse.json(
				{ error: "A category with this slug already exists" },
				{ status: 409 },
			);
		}

		const category = await db.blogCategory.create({
			data: {
				name: body.name,
				slug,
				description:
					typeof body.description === "string"
						? body.description
						: null,
				sortOrder:
					typeof body.sortOrder === "number" ? body.sortOrder : 0,
			},
			include: {
				_count: { select: { posts: true } },
			},
		});

		revalidatePath("/blog");
		revalidatePath("/admin/blog");

		return NextResponse.json(category, { status: 201 });
	} catch (error) {
		log.error("Failed to create blog category", { error });
		return NextResponse.json(
			{ error: "Failed to create blog category" },
			{ status: 500 },
		);
	}
}
