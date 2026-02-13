/**
 * Send Text Message API Route
 *
 * Proxies send-text requests to the WhatsApp API backend.
 * Handles CORS issues by making server-side requests.
 *
 * @route POST /api/send-text
 */

import { type NextRequest, NextResponse } from "next/server";

interface SendTextBody {
	instanceId: string;
	instanceToken: string;
	phone: string;
	message: string;
	delayMessage?: number;
	delayTyping?: number;
	editMessageId?: string;
}

export async function POST(request: NextRequest) {
	try {
		const body: SendTextBody = await request.json();

		const {
			instanceId,
			instanceToken,
			phone,
			message,
			delayMessage,
			delayTyping,
			editMessageId,
		} = body;

		// Validate required fields
		if (!instanceId || !instanceToken || !phone || !message) {
			return NextResponse.json(
				{
					error: "Missing required fields",
					message: "instanceId, instanceToken, phone, and message are required",
				},
				{ status: 400 },
			);
		}

		// Get client token from environment
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

		// Get API URL from environment
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

		// Build request body for WhatsApp API
		const requestBody: Record<string, string | number> = {
			phone,
			message,
		};

		if (delayMessage && delayMessage > 0) {
			requestBody.delayMessage = delayMessage;
		}

		if (delayTyping && delayTyping > 0) {
			requestBody.delayTyping = delayTyping;
		}

		if (editMessageId && editMessageId.trim() !== "") {
			requestBody.editMessageId = editMessageId.trim();
		}

		// Make request to WhatsApp API
		const url = `${apiUrl}/instances/${instanceId}/token/${instanceToken}/send-text`;

		const response = await fetch(url, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
				"Client-Token": clientToken,
			},
			body: JSON.stringify(requestBody),
		});

		const data = await response.json();

		// Return the response from WhatsApp API
		return NextResponse.json(data, { status: response.status });
	} catch (error) {
		console.error("Error sending text message:", error);

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
