import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import { requireContentApiKey } from "@/lib/api-auth";
import { db } from "@/lib/db";
import { createLogger } from "@/lib/logger";
import { slugify } from "@/lib/slugify";

const log = createLogger("api:support");

export async function PATCH(
	req: NextRequest,
	{ params }: { params: Promise<{ id: string }> },
) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const { id } = await params;

		const existing = await db.supportCategory.findUnique({ where: { id } });
		if (!existing) {
			return NextResponse.json(
				{ error: "Category not found" },
				{ status: 404 },
			);
		}

		const body = (await req.json()) as {
			name?: string;
			slug?: string;
			description?: string;
			icon?: string;
			sortOrder?: number;
		};

		const data: Record<string, unknown> = {};

		if (body.name !== undefined) data.name = body.name;
		if (body.description !== undefined) data.description = body.description;
		if (body.icon !== undefined) data.icon = body.icon;
		if (body.sortOrder !== undefined) data.sortOrder = body.sortOrder;

		if (body.slug !== undefined) {
			data.slug = body.slug;
		} else if (body.name !== undefined) {
			data.slug = slugify(body.name);
		}

		if (data.slug) {
			const slugConflict = await db.supportCategory.findUnique({
				where: { slug: data.slug as string },
			});
			if (slugConflict && slugConflict.id !== id) {
				return NextResponse.json(
					{ error: "A category with this slug already exists" },
					{ status: 409 },
				);
			}
		}

		const category = await db.supportCategory.update({
			where: { id },
			data,
			include: {
				_count: { select: { articles: true } },
			},
		});

		revalidatePath("/suporte");
		return NextResponse.json(category);
	} catch (error) {
		log.error("Failed to update support category", { error });
		return NextResponse.json(
			{ error: "Failed to update category" },
			{ status: 500 },
		);
	}
}

export async function DELETE(
	req: NextRequest,
	{ params }: { params: Promise<{ id: string }> },
) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const { id } = await params;

		const existing = await db.supportCategory.findUnique({
			where: { id },
			include: { _count: { select: { articles: true } } },
		});

		if (!existing) {
			return NextResponse.json(
				{ error: "Category not found" },
				{ status: 404 },
			);
		}

		if (existing._count.articles > 0) {
			return NextResponse.json(
				{
					error: `Cannot delete category with ${existing._count.articles} article(s). Move or delete them first.`,
				},
				{ status: 400 },
			);
		}

		await db.supportCategory.delete({ where: { id } });

		revalidatePath("/suporte");
		return NextResponse.json({ success: true });
	} catch (error) {
		log.error("Failed to delete support category", { error });
		return NextResponse.json(
			{ error: "Failed to delete category" },
			{ status: 500 },
		);
	}
}
