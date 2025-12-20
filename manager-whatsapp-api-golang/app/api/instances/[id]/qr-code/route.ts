/**
 * QR Code API Route
 *
 * Provides QR code image for WhatsApp pairing.
 * Returns base64-encoded image data.
 *
 * @module app/api/instances/[id]/qr-code
 */

import { type NextRequest, NextResponse } from "next/server";
import { getInstance, getQRCodeImage } from "@/lib/api/instances";

/**
 * GET /api/instances/[id]/qr-code
 * Get QR code as base64-encoded image
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

		const qrCode = await getQRCodeImage(id, instance.token);

		return NextResponse.json(qrCode);
	} catch (error) {
		console.error("Error fetching QR code:", error);

		if (error && typeof error === "object" && "status" in error) {
			const apiError = error as { message?: string; status?: number };
			return NextResponse.json(
				{ error: apiError.message || "Failed to fetch QR code" },
				{ status: apiError.status || 500 },
			);
		}

		return NextResponse.json(
			{ error: "Internal server error" },
			{ status: 500 },
		);
	}
}
