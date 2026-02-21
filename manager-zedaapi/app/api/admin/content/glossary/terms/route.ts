import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import { requireContentApiKey, parsePaginationParams } from "@/lib/api-auth";
import { db } from "@/lib/db";
import { createLogger } from "@/lib/logger";
import { slugify } from "@/lib/slugify";

const log = createLogger("api:glossary");

const PAGE_SIZE = 20;

export async function GET(req: NextRequest) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const url = new URL(req.url);
		const { page, search } = parsePaginationParams(url);
		const letter = url.searchParams.get("letter") || undefined;

		const where: {
			term?:
				| { contains: string; mode: "insensitive" }
				| { startsWith: string; mode: "insensitive" };
			AND?: Array<{ term: { startsWith: string; mode: "insensitive" } }>;
		} = {};

		if (search) {
			where.term = { contains: search, mode: "insensitive" };
		}

		if (letter) {
			if (search) {
				where.AND = [
					{ term: { startsWith: letter, mode: "insensitive" } },
				];
			} else {
				where.term = { startsWith: letter, mode: "insensitive" };
			}
		}

		const [terms, total] = await Promise.all([
			db.glossaryTerm.findMany({
				where,
				orderBy: { term: "asc" },
				skip: (page - 1) * PAGE_SIZE,
				take: PAGE_SIZE,
			}),
			db.glossaryTerm.count({ where }),
		]);

		return NextResponse.json({
			data: terms,
			pagination: {
				page,
				pageSize: PAGE_SIZE,
				total,
				totalPages: Math.ceil(total / PAGE_SIZE),
			},
		});
	} catch (error) {
		log.error("Failed to list glossary terms", { error });
		return NextResponse.json(
			{ error: "Failed to list terms" },
			{ status: 500 },
		);
	}
}

export async function POST(req: NextRequest) {
	const authError = requireContentApiKey(req);
	if (authError) return authError;

	try {
		const body = (await req.json()) as {
			term?: string;
			definition?: string;
			content?: string;
			seoTitle?: string;
			seoDescription?: string;
			status?: string;
			relatedSlugs?: string[];
		};

		if (!body.term || !body.definition) {
			return NextResponse.json(
				{ error: "Missing required fields: term, definition" },
				{ status: 400 },
			);
		}

		const slug = slugify(body.term);

		const existing = await db.glossaryTerm.findUnique({ where: { slug } });
		if (existing) {
			return NextResponse.json(
				{ error: "A term with this slug already exists" },
				{ status: 409 },
			);
		}

		const term = await db.glossaryTerm.create({
			data: {
				term: body.term,
				slug,
				definition: body.definition,
				content: body.content ?? undefined,
				seoTitle: body.seoTitle ?? undefined,
				seoDescription: body.seoDescription ?? undefined,
				status: body.status ?? "draft",
				relatedSlugs: body.relatedSlugs ?? undefined,
			},
		});

		revalidatePath("/glossario");
		return NextResponse.json(term, { status: 201 });
	} catch (error) {
		log.error("Failed to create glossary term", { error });
		return NextResponse.json(
			{ error: "Failed to create term" },
			{ status: 500 },
		);
	}
}
