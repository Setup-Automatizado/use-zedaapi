/**
 * Profile API Route
 *
 * Handles user profile updates including name.
 *
 * @module app/api/profile
 */

import { headers } from "next/headers";
import { type NextRequest, NextResponse } from "next/server";
import { auth } from "@/lib/auth";
import prisma from "@/lib/prisma";

/**
 * GET /api/profile
 * Get current user profile with linked accounts
 */
export async function GET() {
	try {
		const session = await auth.api.getSession({
			headers: await headers(),
		});

		if (!session?.user) {
			return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
		}

		// Get user with linked accounts
		const user = await prisma.user.findUnique({
			where: { id: session.user.id },
			select: {
				id: true,
				name: true,
				email: true,
				image: true,
				role: true,
				createdAt: true,
				accounts: {
					select: {
						id: true,
						providerId: true,
						createdAt: true,
					},
				},
			},
		});

		if (!user) {
			return NextResponse.json({ error: "User not found" }, { status: 404 });
		}

		// Transform accounts to a simpler format
		const linkedAccounts = user.accounts.map((account) => ({
			id: account.id,
			provider: account.providerId,
			linkedAt: account.createdAt,
		}));

		return NextResponse.json({
			id: user.id,
			name: user.name,
			email: user.email,
			image: user.image,
			role: user.role,
			createdAt: user.createdAt,
			linkedAccounts,
		});
	} catch (error) {
		console.error("Profile fetch error:", error);
		return NextResponse.json(
			{ error: "Failed to fetch profile" },
			{ status: 500 },
		);
	}
}

/**
 * PATCH /api/profile
 * Update user profile
 */
export async function PATCH(request: NextRequest) {
	try {
		const session = await auth.api.getSession({
			headers: await headers(),
		});

		if (!session?.user) {
			return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
		}

		const body = await request.json();
		const { name } = body;

		// Validate name
		if (name !== undefined) {
			if (typeof name !== "string") {
				return NextResponse.json(
					{ error: "Name must be a string" },
					{ status: 400 },
				);
			}

			if (name.length < 2 || name.length > 100) {
				return NextResponse.json(
					{ error: "Name must be between 2 and 100 characters" },
					{ status: 400 },
				);
			}
		}

		// Update user
		const updatedUser = await prisma.user.update({
			where: { id: session.user.id },
			data: {
				...(name !== undefined && { name: name.trim() }),
			},
			select: {
				id: true,
				name: true,
				email: true,
				image: true,
			},
		});

		return NextResponse.json(updatedUser);
	} catch (error) {
		console.error("Profile update error:", error);
		return NextResponse.json(
			{ error: "Failed to update profile" },
			{ status: 500 },
		);
	}
}
