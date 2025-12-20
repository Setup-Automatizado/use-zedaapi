/**
 * Social Account Management API Route
 *
 * Handles unlinking social accounts from user profile.
 *
 * @module app/api/profile/accounts/[accountId]
 */

import { headers } from "next/headers";
import { type NextRequest, NextResponse } from "next/server";
import { auth } from "@/lib/auth";
import prisma from "@/lib/prisma";

interface RouteParams {
	params: Promise<{ accountId: string }>;
}

/**
 * DELETE /api/profile/accounts/[accountId]
 * Unlink a social account
 */
export async function DELETE(request: NextRequest, { params }: RouteParams) {
	try {
		const { accountId } = await params;

		const session = await auth.api.getSession({
			headers: await headers(),
		});

		if (!session?.user) {
			return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
		}

		// Verify the account belongs to the user
		const account = await prisma.account.findFirst({
			where: {
				id: accountId,
				userId: session.user.id,
			},
		});

		if (!account) {
			return NextResponse.json({ error: "Account not found" }, { status: 404 });
		}

		// Check if user has a password or other accounts before unlinking
		const user = await prisma.user.findUnique({
			where: { id: session.user.id },
			include: {
				accounts: true,
			},
		});

		// Check if user has email/password login
		const hasPassword = await prisma.account.findFirst({
			where: {
				userId: session.user.id,
				providerId: "credential",
			},
		});

		const otherAccounts =
			user?.accounts.filter((a) => a.id !== accountId) || [];

		// User must have at least one way to login
		if (!hasPassword && otherAccounts.length === 0) {
			return NextResponse.json(
				{
					error:
						"Cannot unlink the only login method. Add a password or link another account first.",
				},
				{ status: 400 },
			);
		}

		// Delete the account
		await prisma.account.delete({
			where: { id: accountId },
		});

		return NextResponse.json({ success: true });
	} catch (error) {
		console.error("Account unlink error:", error);
		return NextResponse.json(
			{ error: "Failed to unlink account" },
			{ status: 500 },
		);
	}
}
