import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import crypto from "node:crypto";

/**
 * Validates the Content API key from the Authorization header.
 * Expects: Authorization: Bearer <CONTENT_API_KEY>
 *
 * Returns null if valid, or a NextResponse with 401 status if invalid.
 */
export function requireContentApiKey(req: NextRequest): NextResponse | null {
	const apiKey = process.env.CONTENT_API_KEY;

	if (!apiKey) {
		return NextResponse.json(
			{ error: "Content API is not configured" },
			{ status: 503 },
		);
	}

	const authHeader = req.headers.get("authorization");
	if (!authHeader) {
		return NextResponse.json(
			{ error: "Missing Authorization header" },
			{ status: 401 },
		);
	}

	const token = authHeader.startsWith("Bearer ")
		? authHeader.slice(7)
		: authHeader;

	// Constant-time comparison to prevent timing attacks
	const isValid =
		token.length === apiKey.length &&
		crypto.timingSafeEqual(Buffer.from(token), Buffer.from(apiKey));

	if (!isValid) {
		return NextResponse.json({ error: "Invalid API key" }, { status: 401 });
	}

	return null;
}

/**
 * Parses page/search query params from the URL.
 */
export function parsePaginationParams(url: URL): {
	page: number;
	search: string | undefined;
} {
	const page = Math.max(1, Number(url.searchParams.get("page")) || 1);
	const search = url.searchParams.get("search") || undefined;
	return { page, search };
}
