/**
 * Avatar Upload API Route
 *
 * Handles avatar image uploads to S3 storage.
 * Requires authentication.
 *
 * @module app/api/upload/avatar
 */

import { NextRequest, NextResponse } from "next/server";
import { auth } from "@/lib/auth";
import { headers } from "next/headers";
import prisma from "@/lib/prisma";
import {
	uploadFile,
	deleteFile,
	generateUniqueFilename,
	validateImageFile,
	isS3Configured,
	s3PublicUrl,
} from "@/lib/s3";

/**
 * POST /api/upload/avatar
 * Upload a new avatar image
 */
export async function POST(request: NextRequest) {
	try {
		// Check S3 configuration
		if (!isS3Configured()) {
			return NextResponse.json(
				{ error: "Storage not configured" },
				{ status: 503 },
			);
		}

		// Get session
		const session = await auth.api.getSession({
			headers: await headers(),
		});

		if (!session?.user) {
			return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
		}

		// Parse form data
		const formData = await request.formData();
		const file = formData.get("file") as File | null;

		if (!file) {
			return NextResponse.json(
				{ error: "No file provided" },
				{ status: 400 },
			);
		}

		// Validate file
		const validation = validateImageFile(file);
		if (!validation.valid) {
			return NextResponse.json(
				{ error: validation.error },
				{ status: 400 },
			);
		}

		// Generate unique filename
		const key = generateUniqueFilename(file.name, "avatars");

		// Read file as buffer
		const buffer = Buffer.from(await file.arrayBuffer());

		// Upload to S3
		const imageUrl = await uploadFile(key, buffer, file.type);

		// Delete old avatar if exists
		const user = await prisma.user.findUnique({
			where: { id: session.user.id },
			select: { image: true },
		});

		if (user?.image && user.image.startsWith(s3PublicUrl)) {
			const oldKey = user.image.replace(`${s3PublicUrl}/`, "");
			try {
				await deleteFile(oldKey);
			} catch (e) {
				// Ignore deletion errors for old avatar
				console.warn("Failed to delete old avatar:", e);
			}
		}

		// Update user avatar in database
		await prisma.user.update({
			where: { id: session.user.id },
			data: { image: imageUrl },
		});

		return NextResponse.json({
			success: true,
			imageUrl,
		});
	} catch (error) {
		console.error("Avatar upload error:", error);
		return NextResponse.json(
			{ error: "Failed to upload avatar" },
			{ status: 500 },
		);
	}
}

/**
 * DELETE /api/upload/avatar
 * Remove current avatar
 */
export async function DELETE() {
	try {
		// Get session
		const session = await auth.api.getSession({
			headers: await headers(),
		});

		if (!session?.user) {
			return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
		}

		// Get current user
		const user = await prisma.user.findUnique({
			where: { id: session.user.id },
			select: { image: true },
		});

		// Delete from S3 if it's our uploaded avatar
		if (user?.image && user.image.startsWith(s3PublicUrl)) {
			const key = user.image.replace(`${s3PublicUrl}/`, "");
			try {
				await deleteFile(key);
			} catch (e) {
				console.warn("Failed to delete avatar from S3:", e);
			}
		}

		// Remove avatar from user
		await prisma.user.update({
			where: { id: session.user.id },
			data: { image: null },
		});

		return NextResponse.json({ success: true });
	} catch (error) {
		console.error("Avatar delete error:", error);
		return NextResponse.json(
			{ error: "Failed to delete avatar" },
			{ status: 500 },
		);
	}
}
