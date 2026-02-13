/**
 * Config API Route
 *
 * Returns safe configuration values for the frontend.
 * Used to expose WHATSAPP_CLIENT_TOKEN safely.
 */

import { NextResponse } from "next/server";

export async function GET() {
	return NextResponse.json({
		clientToken: process.env.WHATSAPP_CLIENT_TOKEN || "",
	});
}
