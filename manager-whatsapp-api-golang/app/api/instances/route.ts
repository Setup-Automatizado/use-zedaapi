/**
 * Instances API Route
 *
 * Handles listing WhatsApp instances via Next.js Route Handler.
 * This proxies to the Go backend API which requires Partner-Token.
 *
 * @module app/api/instances
 */

import { type NextRequest, NextResponse } from "next/server";
import { listInstances } from "@/lib/api/instances";

/**
 * GET /api/instances
 * List all instances for the authenticated partner
 *
 * Query Parameters:
 * - page: Page number (default: 1)
 * - pageSize: Items per page (default: 20)
 * - query: Optional search query
 * - status: Optional status filter
 */
export async function GET(request: NextRequest) {
	try {
		const { searchParams } = new URL(request.url);

		const page = parseInt(searchParams.get("page") || "1", 10);
		const pageSize = parseInt(searchParams.get("pageSize") || "20", 10);
		const query = searchParams.get("query") || undefined;

		// Validate pagination params
		if (page < 1 || pageSize < 1 || pageSize > 100) {
			return NextResponse.json(
				{ error: "Invalid pagination parameters" },
				{ status: 400 },
			);
		}

		const response = await listInstances(page, pageSize, query);

		return NextResponse.json(response);
	} catch (error) {
		console.error("Error listing instances:", error);

		// Handle API errors
		if (error && typeof error === "object" && "status" in error) {
			const apiError = error as { message?: string; status?: number };
			return NextResponse.json(
				{ error: apiError.message || "Failed to list instances" },
				{ status: apiError.status || 500 },
			);
		}

		return NextResponse.json(
			{ error: "Internal server error" },
			{ status: 500 },
		);
	}
}
