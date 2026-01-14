/**
 * Single Instance API Route
 *
 * Handles operations on individual WhatsApp instances.
 * Fetches instance from the list endpoint using Partner token.
 *
 * @module app/api/instances/[id]
 */

import { type NextRequest, NextResponse } from "next/server";
import { getInstance } from "@/lib/api/instances";

/**
 * Cache control headers to prevent stale data after webhook updates
 */
const NO_CACHE_HEADERS = {
	"Cache-Control": "no-cache, no-store, must-revalidate",
	Pragma: "no-cache",
	Expires: "0",
} as const;

/**
 * GET /api/instances/[id]
 * Get instance details by ID
 *
 * Uses Partner token to fetch from list endpoint
 * Returns no-cache headers to ensure fresh data after updates
 */
export async function GET(
	request: NextRequest,
	context: { params: Promise<{ id: string }> },
) {
	try {
		const { id } = await context.params;

		// Fetch instance from list using Partner token
		const instance = await getInstance(id);

		if (!instance) {
			return NextResponse.json(
				{ error: "Instance not found" },
				{ status: 404, headers: NO_CACHE_HEADERS },
			);
		}

		return NextResponse.json(instance, { headers: NO_CACHE_HEADERS });
	} catch (error) {
		console.error("Error fetching instance:", error);

		if (error && typeof error === "object" && "status" in error) {
			const apiError = error as { message?: string; status?: number };
			return NextResponse.json(
				{ error: apiError.message || "Failed to fetch instance" },
				{ status: apiError.status || 500, headers: NO_CACHE_HEADERS },
			);
		}

		return NextResponse.json(
			{ error: "Internal server error" },
			{ status: 500, headers: NO_CACHE_HEADERS },
		);
	}
}
