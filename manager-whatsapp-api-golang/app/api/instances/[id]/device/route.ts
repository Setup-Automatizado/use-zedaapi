/**
 * Device Info API Route
 *
 * Provides device information for connected WhatsApp instances.
 *
 * @module app/api/instances/[id]/device
 */

import { NextRequest, NextResponse } from "next/server";
import { getInstance, getDeviceInfo } from "@/lib/api/instances";

/**
 * GET /api/instances/[id]/device
 * Get device information for connected instance
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

		const deviceInfo = await getDeviceInfo(id, instance.token);

		return NextResponse.json(deviceInfo);
	} catch (error) {
		console.error("Error fetching device info:", error);

		if (error && typeof error === "object" && "status" in error) {
			const apiError = error as { message?: string; status?: number };
			return NextResponse.json(
				{
					error: apiError.message || "Failed to fetch device info",
				},
				{ status: apiError.status || 500 },
			);
		}

		return NextResponse.json(
			{ error: "Internal server error" },
			{ status: 500 },
		);
	}
}
