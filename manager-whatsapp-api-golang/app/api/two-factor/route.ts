/**
 * Two-Factor Method API
 *
 * GET: Get current 2FA method for authenticated user
 * PATCH: Update 2FA method for authenticated user
 */

import { headers } from "next/headers";
import { type NextRequest, NextResponse } from "next/server";
import { auth } from "@/lib/auth";
import prisma from "@/lib/prisma";

export async function GET() {
	try {
		const session = await auth.api.getSession({
			headers: await headers(),
		});

		if (!session?.user) {
			return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
		}

		const twoFactor = await prisma.twoFactor.findUnique({
			where: { userId: session.user.id },
			select: { method: true },
		});

		return NextResponse.json({
			method: twoFactor?.method || null,
			enabled: !!twoFactor,
		});
	} catch (error) {
		console.error("Error fetching 2FA method:", error);
		return NextResponse.json(
			{ error: "Internal server error" },
			{ status: 500 },
		);
	}
}

export async function PATCH(request: NextRequest) {
	try {
		const session = await auth.api.getSession({
			headers: await headers(),
		});

		if (!session?.user) {
			return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
		}

		const body = await request.json();
		const { method } = body;

		if (!method || !["totp", "email"].includes(method)) {
			return NextResponse.json(
				{ error: "Invalid method. Must be 'totp' or 'email'" },
				{ status: 400 },
			);
		}

		const twoFactor = await prisma.twoFactor.findUnique({
			where: { userId: session.user.id },
		});

		if (!twoFactor) {
			return NextResponse.json({ error: "2FA not enabled" }, { status: 404 });
		}

		await prisma.twoFactor.update({
			where: { userId: session.user.id },
			data: { method },
		});

		return NextResponse.json({ success: true, method });
	} catch (error) {
		console.error("Error updating 2FA method:", error);
		return NextResponse.json(
			{ error: "Internal server error" },
			{ status: 500 },
		);
	}
}
