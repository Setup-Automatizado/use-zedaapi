/**
 * Phone Pairing Code API Route
 *
 * Provides phone-based pairing code for WhatsApp linking.
 *
 * @module app/api/instances/[id]/phone-code/[phone]
 */

import { type NextRequest, NextResponse } from "next/server";
import { getInstance, getPhonePairingCode } from "@/lib/api/instances";

/**
 * GET /api/instances/[id]/phone-code/[phone]
 * Get phone pairing code for device linking
 *
 * Automatically fetches instance token from the instances list
 */
export async function GET(
	request: NextRequest,
	context: { params: Promise<{ id: string; phone: string }> },
) {
	try {
		const { id, phone } = await context.params;

		if (!phone) {
			return NextResponse.json(
				{ error: "Phone number is required" },
				{ status: 400 },
			);
		}

		// Fetch instance to get the token
		const instance = await getInstance(id);

		if (!instance) {
			return NextResponse.json(
				{ error: "Instance not found" },
				{ status: 404 },
			);
		}

		const pairingCode = await getPhonePairingCode(id, instance.token, phone);

		return NextResponse.json(pairingCode);
	} catch (error) {
		console.error("Error fetching phone pairing code:", error);

		if (error && typeof error === "object" && "status" in error) {
			const apiError = error as { message?: string; status?: number };
			return NextResponse.json(
				{
					error: apiError.message || "Failed to fetch pairing code",
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
