import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import { requireContentApiKey } from "@/lib/api-auth";
import { db } from "@/lib/db";
import { createLogger } from "@/lib/logger";

const log = createLogger("api:blog-posts");

// POST /api/admin/content/blog/posts/[id]/archive
export async function POST(
	req: NextRequest,
	{ params }: { params: Promise<{ id: string }> },
) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const { id } = await params;

		const existing = await db.blogPost.findUnique({ where: { id } });
		if (!existing) {
			return NextResponse.json(
				{ error: "Post not found" },
				{ status: 404 },
			);
		}

		if (existing.status === "archived") {
			return NextResponse.json(
				{ error: "Post is already archived" },
				{ status: 400 },
			);
		}

		const post = await db.blogPost.update({
			where: { id },
			data: { status: "archived" },
			include: {
				category: true,
				tags: { include: { tag: true } },
			},
		});

		revalidatePath("/blog");
		revalidatePath(`/blog/${post.slug}`);
		revalidatePath("/admin/blog");

		return NextResponse.json(post);
	} catch (error) {
		log.error("Failed to archive blog post", { error });
		return NextResponse.json(
			{ error: "Failed to archive blog post" },
			{ status: 500 },
		);
	}
}
