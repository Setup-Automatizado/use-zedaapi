/**
 * Message Status Stats API Route
 *
 * Proxies message status statistics requests to the WhatsApp API backend.
 *
 * @route GET /api/instances/[id]/messages-status/stats
 */

import { type NextRequest, NextResponse } from "next/server";

interface RouteParams {
	params: Promise<{
		id: string;
	}>;
}

export async function GET(request: NextRequest, props: RouteParams) {
	const params = await props.params;
	try {
		const { searchParams } = new URL(request.url);
		const instanceToken = searchParams.get("instanceToken");

		if (!instanceToken) {
			return NextResponse.json(
				{
					error: "Missing instanceToken",
					message: "instanceToken query parameter is required",
				},
				{ status: 400 },
			);
		}

		const clientToken = process.env.WHATSAPP_CLIENT_TOKEN;
		if (!clientToken) {
			return NextResponse.json(
				{
					error: "Configuration error",
					message: "WHATSAPP_CLIENT_TOKEN not configured",
				},
				{ status: 500 },
			);
		}

		const apiUrl = process.env.WHATSAPP_API_URL;
		if (!apiUrl) {
			return NextResponse.json(
				{
					error: "Configuration error",
					message: "WHATSAPP_API_URL not configured",
				},
				{ status: 500 },
			);
		}

		const url = `${apiUrl}/instances/${params.id}/token/${instanceToken}/messages-status/stats`;

		const response = await fetch(url, {
			headers: {
				"Client-Token": clientToken,
			},
		});

		const data = await response.json();

		return NextResponse.json(data, { status: response.status });
	} catch (error) {
		console.error("Error fetching message status stats:", error);

		return NextResponse.json(
			{
				error: "Internal server error",
				message:
					error instanceof Error ? error.message : "Unknown error occurred",
			},
			{ status: 500 },
		);
	}
}
