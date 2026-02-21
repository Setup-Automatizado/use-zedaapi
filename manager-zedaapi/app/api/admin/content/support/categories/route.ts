import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import { requireContentApiKey } from "@/lib/api-auth";
import { db } from "@/lib/db";
import { slugify } from "@/lib/slugify";

export async function GET(req: NextRequest) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const categories = await db.supportCategory.findMany({
			orderBy: [{ sortOrder: "asc" }, { name: "asc" }],
			include: {
				_count: { select: { articles: true } },
			},
		});

		return NextResponse.json({ data: categories });
	} catch (error) {
		console.error("Failed to list support categories:", error);
		return NextResponse.json(
			{ error: "Failed to list categories" },
			{ status: 500 },
		);
	}
}

export async function POST(req: NextRequest) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const body = (await req.json()) as {
			name?: string;
			slug?: string;
			description?: string;
			icon?: string;
			sortOrder?: number;
		};

		if (!body.name) {
			return NextResponse.json(
				{ error: "Missing required field: name" },
				{ status: 400 },
			);
		}

		const slug = body.slug || slugify(body.name);

		const existing = await db.supportCategory.findUnique({
			where: { slug },
		});
		if (existing) {
			return NextResponse.json(
				{ error: "A category with this slug already exists" },
				{ status: 409 },
			);
		}

		const category = await db.supportCategory.create({
			data: {
				name: body.name,
				slug,
				description: body.description ?? undefined,
				icon: body.icon ?? undefined,
				sortOrder: body.sortOrder ?? 0,
			},
			include: {
				_count: { select: { articles: true } },
			},
		});

		revalidatePath("/suporte");
		return NextResponse.json(category, { status: 201 });
	} catch (error) {
		console.error("Failed to create support category:", error);
		return NextResponse.json(
			{ error: "Failed to create category" },
			{ status: 500 },
		);
	}
}
