import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import { requireContentApiKey } from "@/lib/api-auth";
import { db } from "@/lib/db";
import { createLogger } from "@/lib/logger";
import { slugify } from "@/lib/slugify";

const log = createLogger("api:blog-tags");

// GET /api/admin/content/blog/tags
export async function GET(req: NextRequest) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const tags = await db.blogTag.findMany({
			orderBy: { name: "asc" },
			include: {
				_count: { select: { posts: true } },
			},
		});

		return NextResponse.json(tags);
	} catch (error) {
		log.error("Failed to list blog tags", { error });
		return NextResponse.json(
			{ error: "Failed to list blog tags" },
			{ status: 500 },
		);
	}
}

// POST /api/admin/content/blog/tags
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
		const existing = await db.blogTag.findUnique({ where: { slug } });
		if (existing) {
			return NextResponse.json(
				{ error: "A tag with this slug already exists" },
				{ status: 409 },
			);
		}

		const tag = await db.blogTag.create({
			data: {
				name: body.name,
				slug,
			},
			include: {
				_count: { select: { posts: true } },
			},
		});

		revalidatePath("/admin/blog");

		return NextResponse.json(tag, { status: 201 });
	} catch (error) {
		log.error("Failed to create blog tag", { error });
		return NextResponse.json(
			{ error: "Failed to create blog tag" },
			{ status: 500 },
		);
	}
}
