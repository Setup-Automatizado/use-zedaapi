import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import { requireContentApiKey } from "@/lib/api-auth";
import { db } from "@/lib/db";
import { slugify } from "@/lib/slugify";

export async function GET(
	req: NextRequest,
	{ params }: { params: Promise<{ id: string }> },
) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const { id } = await params;

		const article = await db.supportArticle.findUnique({
			where: { id },
			include: { category: true },
		});

		if (!article) {
			return NextResponse.json(
				{ error: "Article not found" },
				{ status: 404 },
			);
		}

		return NextResponse.json(article);
	} catch (error) {
		console.error("Failed to get support article:", error);
		return NextResponse.json(
			{ error: "Failed to get article" },
			{ status: 500 },
		);
	}
}

export async function PATCH(
	req: NextRequest,
	{ params }: { params: Promise<{ id: string }> },
) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const { id } = await params;

		const existing = await db.supportArticle.findUnique({ where: { id } });
		if (!existing) {
			return NextResponse.json(
				{ error: "Article not found" },
				{ status: 404 },
			);
		}

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

		const data: Record<string, unknown> = {};

		if (body.title !== undefined) {
			data.title = body.title;
			data.slug = slugify(body.title);

			const slugConflict = await db.supportArticle.findUnique({
				where: { slug: data.slug as string },
			});
			if (slugConflict && slugConflict.id !== id) {
				return NextResponse.json(
					{ error: "An article with this slug already exists" },
					{ status: 409 },
				);
			}
		}

		if (body.content !== undefined) data.content = body.content;
		if (body.excerpt !== undefined) data.excerpt = body.excerpt;
		if (body.categoryId !== undefined) data.categoryId = body.categoryId;
		if (body.seoTitle !== undefined) data.seoTitle = body.seoTitle;
		if (body.seoDescription !== undefined)
			data.seoDescription = body.seoDescription;
		if (body.status !== undefined) data.status = body.status;
		if (body.sortOrder !== undefined) data.sortOrder = body.sortOrder;

		const article = await db.supportArticle.update({
			where: { id },
			data,
			include: { category: true },
		});

		revalidatePath("/suporte");
		return NextResponse.json(article);
	} catch (error) {
		console.error("Failed to update support article:", error);
		return NextResponse.json(
			{ error: "Failed to update article" },
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

		const existing = await db.supportArticle.findUnique({ where: { id } });
		if (!existing) {
			return NextResponse.json(
				{ error: "Article not found" },
				{ status: 404 },
			);
		}

		await db.supportArticle.delete({ where: { id } });

		revalidatePath("/suporte");
		return NextResponse.json({ success: true });
	} catch (error) {
		console.error("Failed to delete support article:", error);
		return NextResponse.json(
			{ error: "Failed to delete article" },
			{ status: 500 },
		);
	}
}
