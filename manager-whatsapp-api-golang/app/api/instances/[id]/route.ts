/**
 * Single Instance API Route
 *
 * Handles operations on individual WhatsApp instances.
 * Fetches instance from the list endpoint using Partner token.
 *
 * @module app/api/instances/[id]
 */

import { NextRequest, NextResponse } from "next/server";
import { getInstance } from "@/lib/api/instances";

/**
 * GET /api/instances/[id]
 * Get instance details by ID
 *
 * Uses Partner token to fetch from list endpoint
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
				{ status: 404 },
			);
		}

		return NextResponse.json(instance);
	} catch (error) {
		console.error("Error fetching instance:", error);

		if (error && typeof error === "object" && "status" in error) {
			const apiError = error as { message?: string; status?: number };
			return NextResponse.json(
				{ error: apiError.message || "Failed to fetch instance" },
				{ status: apiError.status || 500 },
			);
		}

		return NextResponse.json(
			{ error: "Internal server error" },
			{ status: 500 },
		);
	}
}
