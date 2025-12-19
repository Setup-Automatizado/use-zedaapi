/**
 * Invitations API Route
 *
 * Handles CRUD operations for user invitations.
 * Only ADMIN users can access these endpoints.
 *
 * @module app/api/invitations
 */

import { NextRequest, NextResponse } from "next/server";
import { auth } from "@/lib/auth";
import prisma from "@/lib/prisma";
import { sendUserInvite } from "@/lib/email";

/**
 * GET /api/invitations
 * List all invitations (admin only)
 */
export async function GET(request: NextRequest) {
	try {
		const session = await auth.api.getSession({ headers: request.headers });

		if (!session?.user) {
			return NextResponse.json(
				{ error: "Unauthorized" },
				{ status: 401 },
			);
		}

		// Check if user is admin
		const user = await prisma.user.findUnique({
			where: { id: session.user.id },
			select: { role: true },
		});

		if (user?.role !== "ADMIN") {
			return NextResponse.json({ error: "Forbidden" }, { status: 403 });
		}

		const invitations = await prisma.allowedUser.findMany({
			include: {
				invitedBy: { select: { name: true, email: true } },
				user: { select: { name: true, email: true, createdAt: true } },
			},
			orderBy: { invitedAt: "desc" },
		});

		return NextResponse.json({ invitations });
	} catch (error) {
		console.error("Error listing invitations:", error);
		return NextResponse.json(
			{ error: "Internal server error" },
			{ status: 500 },
		);
	}
}

/**
 * POST /api/invitations
 * Create a new invitation (admin only)
 */
export async function POST(request: NextRequest) {
	try {
		const session = await auth.api.getSession({ headers: request.headers });

		if (!session?.user) {
			return NextResponse.json(
				{ error: "Unauthorized" },
				{ status: 401 },
			);
		}

		// Check if user is admin
		const adminUser = await prisma.user.findUnique({
			where: { id: session.user.id },
			select: { role: true, name: true, email: true },
		});

		if (adminUser?.role !== "ADMIN") {
			return NextResponse.json({ error: "Forbidden" }, { status: 403 });
		}

		const body = await request.json();
		const { email, role } = body;

		if (!email || !role) {
			return NextResponse.json(
				{ error: "Email and role are required" },
				{ status: 400 },
			);
		}

		// Validate email format
		const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
		if (!emailRegex.test(email)) {
			return NextResponse.json(
				{ error: "Invalid email format" },
				{ status: 400 },
			);
		}

		// Validate role
		if (!["ADMIN", "USER"].includes(role)) {
			return NextResponse.json(
				{ error: "Invalid role" },
				{ status: 400 },
			);
		}

		// Check if already exists
		const existing = await prisma.allowedUser.findUnique({
			where: { email: email.toLowerCase() },
		});

		if (existing) {
			return NextResponse.json(
				{ error: "Este email ja foi convidado" },
				{ status: 409 },
			);
		}

		// Create invitation
		const invitation = await prisma.allowedUser.create({
			data: {
				email: email.toLowerCase(),
				role,
				invitedById: session.user.id,
			},
		});

		// Send invitation email
		const inviteUrl = `${process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000"}/register?invited=true&email=${encodeURIComponent(email)}`;

		try {
			await sendUserInvite(email, {
				inviteeEmail: email,
				inviterName: adminUser.name || "Admin",
				inviterEmail: adminUser.email,
				role,
				inviteUrl,
				expiresIn: 48, // 48 hours
			});
		} catch (emailError) {
			console.error("Failed to send invitation email:", emailError);
			// Don't fail the request if email fails - invitation is still created
		}

		return NextResponse.json({ invitation }, { status: 201 });
	} catch (error) {
		console.error("Error creating invitation:", error);
		return NextResponse.json(
			{ error: "Internal server error" },
			{ status: 500 },
		);
	}
}

/**
 * DELETE /api/invitations?email=xxx
 * Remove an invitation (admin only)
 */
export async function DELETE(request: NextRequest) {
	try {
		const session = await auth.api.getSession({ headers: request.headers });

		if (!session?.user) {
			return NextResponse.json(
				{ error: "Unauthorized" },
				{ status: 401 },
			);
		}

		// Check if user is admin
		const user = await prisma.user.findUnique({
			where: { id: session.user.id },
			select: { role: true, email: true },
		});

		if (user?.role !== "ADMIN") {
			return NextResponse.json({ error: "Forbidden" }, { status: 403 });
		}

		const { searchParams } = new URL(request.url);
		const email = searchParams.get("email");

		if (!email) {
			return NextResponse.json(
				{ error: "Email is required" },
				{ status: 400 },
			);
		}

		// Don't allow removing yourself
		if (email.toLowerCase() === user.email.toLowerCase()) {
			return NextResponse.json(
				{ error: "Voce nao pode remover a si mesmo" },
				{ status: 400 },
			);
		}

		// Check if invitation exists
		const existing = await prisma.allowedUser.findUnique({
			where: { email: email.toLowerCase() },
		});

		if (!existing) {
			return NextResponse.json(
				{ error: "Convite nao encontrado" },
				{ status: 404 },
			);
		}

		// Delete the invitation
		await prisma.allowedUser.delete({
			where: { email: email.toLowerCase() },
		});

		// If user exists, also delete them (revoke access)
		if (existing.userId) {
			try {
				// Delete sessions first
				await prisma.session.deleteMany({
					where: { userId: existing.userId },
				});
				// Delete accounts
				await prisma.account.deleteMany({
					where: { userId: existing.userId },
				});
				// Delete two factors
				await prisma.twoFactor.deleteMany({
					where: { userId: existing.userId },
				});
				// Delete user
				await prisma.user.delete({
					where: { id: existing.userId },
				});
			} catch (userDeleteError) {
				console.error("Error deleting user:", userDeleteError);
				// Continue even if user deletion fails
			}
		}

		return NextResponse.json({ success: true });
	} catch (error) {
		console.error("Error deleting invitation:", error);
		return NextResponse.json(
			{ error: "Internal server error" },
			{ status: 500 },
		);
	}
}
