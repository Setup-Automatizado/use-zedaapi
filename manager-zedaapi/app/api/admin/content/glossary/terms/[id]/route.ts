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

		const term = await db.glossaryTerm.findUnique({ where: { id } });

		if (!term) {
			return NextResponse.json(
				{ error: "Term not found" },
				{ status: 404 },
			);
		}

		return NextResponse.json(term);
	} catch (error) {
		console.error("Failed to get glossary term:", error);
		return NextResponse.json(
			{ error: "Failed to get term" },
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

		const existing = await db.glossaryTerm.findUnique({ where: { id } });
		if (!existing) {
			return NextResponse.json(
				{ error: "Term not found" },
				{ status: 404 },
			);
		}

		const body = (await req.json()) as {
			term?: string;
			definition?: string;
			content?: string;
			seoTitle?: string;
			seoDescription?: string;
			status?: string;
			relatedSlugs?: string[];
		};

		const data: Record<string, unknown> = {};

		if (body.term !== undefined) {
			data.term = body.term;
			data.slug = slugify(body.term);

			const slugConflict = await db.glossaryTerm.findUnique({
				where: { slug: data.slug as string },
			});
			if (slugConflict && slugConflict.id !== id) {
				return NextResponse.json(
					{ error: "A term with this slug already exists" },
					{ status: 409 },
				);
			}
		}

		if (body.definition !== undefined) data.definition = body.definition;
		if (body.content !== undefined) data.content = body.content;
		if (body.seoTitle !== undefined) data.seoTitle = body.seoTitle;
		if (body.seoDescription !== undefined)
			data.seoDescription = body.seoDescription;
		if (body.status !== undefined) data.status = body.status;
		if (body.relatedSlugs !== undefined)
			data.relatedSlugs = body.relatedSlugs ?? undefined;

		const term = await db.glossaryTerm.update({
			where: { id },
			data,
		});

		revalidatePath("/glossario");
		return NextResponse.json(term);
	} catch (error) {
		console.error("Failed to update glossary term:", error);
		return NextResponse.json(
			{ error: "Failed to update term" },
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

		const existing = await db.glossaryTerm.findUnique({ where: { id } });
		if (!existing) {
			return NextResponse.json(
				{ error: "Term not found" },
				{ status: 404 },
			);
		}

		await db.glossaryTerm.delete({ where: { id } });

		revalidatePath("/glossario");
		return NextResponse.json({ success: true });
	} catch (error) {
		console.error("Failed to delete glossary term:", error);
		return NextResponse.json(
			{ error: "Failed to delete term" },
			{ status: 500 },
		);
	}
}
