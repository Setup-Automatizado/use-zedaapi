import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import { requireContentApiKey } from "@/lib/api-auth";
import { db } from "@/lib/db";

// DELETE /api/admin/content/blog/tags/[id]
export async function DELETE(
	req: NextRequest,
	{ params }: { params: Promise<{ id: string }> },
) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const { id } = await params;

		const existing = await db.blogTag.findUnique({ where: { id } });
		if (!existing) {
			return NextResponse.json(
				{ error: "Tag not found" },
				{ status: 404 },
			);
		}

		await db.blogTag.delete({ where: { id } });

		revalidatePath("/admin/blog");

		return NextResponse.json({ success: true });
	} catch (error) {
		console.error("Failed to delete blog tag:", error);
		return NextResponse.json(
			{ error: "Failed to delete blog tag" },
			{ status: 500 },
		);
	}
}
