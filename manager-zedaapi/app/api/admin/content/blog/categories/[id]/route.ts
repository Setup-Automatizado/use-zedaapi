import { revalidatePath } from "next/cache";
import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { requireContentApiKey } from "@/lib/api-auth";
import { db } from "@/lib/db";
import { slugify } from "@/lib/slugify";

// PATCH /api/admin/content/blog/categories/[id]
export async function PATCH(
	req: NextRequest,
	{ params }: { params: Promise<{ id: string }> },
) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const { id } = await params;
		const body = (await req.json()) as Record<string, unknown>;

		const existing = await db.blogCategory.findUnique({ where: { id } });
		if (!existing) {
			return NextResponse.json(
				{ error: "Category not found" },
				{ status: 404 },
			);
		}

		const data: Record<string, unknown> = {};

		if (typeof body.name === "string" && body.name.trim()) {
			data.name = body.name;
			const newSlug = slugify(body.name);
			if (newSlug !== existing.slug) {
				const slugTaken = await db.blogCategory.findUnique({
					where: { slug: newSlug },
				});
				if (slugTaken) {
					return NextResponse.json(
						{
							error: "A category with this slug already exists",
						},
						{ status: 409 },
					);
				}
				data.slug = newSlug;
			}
		}

		if (typeof body.description === "string")
			data.description = body.description;
		if (typeof body.sortOrder === "number") data.sortOrder = body.sortOrder;

		const category = await db.blogCategory.update({
			where: { id },
			data,
			include: {
				_count: { select: { posts: true } },
			},
		});

		revalidatePath("/blog");
		revalidatePath("/admin/blog");

		return NextResponse.json(category);
	} catch (error) {
		console.error("Failed to update blog category:", error);
		return NextResponse.json(
			{ error: "Failed to update blog category" },
			{ status: 500 },
		);
	}
}

// DELETE /api/admin/content/blog/categories/[id]
export async function DELETE(
	req: NextRequest,
	{ params }: { params: Promise<{ id: string }> },
) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const { id } = await params;

		const existing = await db.blogCategory.findUnique({
			where: { id },
			include: { _count: { select: { posts: true } } },
		});
		if (!existing) {
			return NextResponse.json(
				{ error: "Category not found" },
				{ status: 404 },
			);
		}

		if (existing._count.posts > 0) {
			return NextResponse.json(
				{
					error: `Cannot delete category with ${existing._count.posts} post(s). Reassign or delete posts first.`,
				},
				{ status: 409 },
			);
		}

		await db.blogCategory.delete({ where: { id } });

		revalidatePath("/blog");
		revalidatePath("/admin/blog");

		return NextResponse.json({ success: true });
	} catch (error) {
		console.error("Failed to delete blog category:", error);
		return NextResponse.json(
			{ error: "Failed to delete blog category" },
			{ status: 500 },
		);
	}
}
