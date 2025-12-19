/**
 * Instance Status API Route
 *
 * Provides current connection status and metadata for a WhatsApp instance.
 *
 * @module app/api/instances/[id]/status
 */

import { NextRequest, NextResponse } from "next/server";
import { getInstance, getInstanceStatus } from "@/lib/api/instances";

/**
 * GET /api/instances/[id]/status
 * Get instance connection status
 *
 * Automatically fetches instance token from the instances list
 */
export async function GET(
	request: NextRequest,
	context: { params: Promise<{ id: string }> },
) {
	try {
		const { id } = await context.params;

		// Fetch instance to get the token
		const instance = await getInstance(id);

		if (!instance) {
			return NextResponse.json(
				{ error: "Instance not found" },
				{ status: 404 },
			);
		}

		const status = await getInstanceStatus(id, instance.token);

		return NextResponse.json(status);
	} catch (error) {
		console.error("Error fetching instance status:", error);

		if (error && typeof error === "object" && "status" in error) {
			const apiError = error as { message?: string; status?: number };
			return NextResponse.json(
				{ error: apiError.message || "Failed to fetch status" },
				{ status: apiError.status || 500 },
			);
		}

		return NextResponse.json(
			{ error: "Internal server error" },
			{ status: 500 },
		);
	}
}
