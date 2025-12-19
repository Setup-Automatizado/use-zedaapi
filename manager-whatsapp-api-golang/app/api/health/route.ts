/**
 * Health Check API Route
 *
 * Provides basic service health status for container orchestration.
 * Public endpoint - no authentication required.
 *
 * This endpoint checks ONLY the manager service itself, NOT external dependencies.
 * It's designed for:
 * - Docker HEALTHCHECK
 * - AWS ECS/ALB health checks
 * - Kubernetes liveness probes
 *
 * The manager's own health is determined by its ability to:
 * 1. Respond to HTTP requests (Next.js is running)
 * 2. Connect to its database (Prisma connection)
 *
 * @module app/api/health
 */

import { NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";

/**
 * GET /api/health
 * Get manager service health status
 *
 * Checks:
 * 1. Service is responding (Next.js process is alive)
 * 2. Database connectivity (Prisma can query database)
 *
 * Returns:
 * - 200 OK: Manager is healthy
 * - 503 Service Unavailable: Manager has critical failure
 *
 * Note: This does NOT check WhatsApp API backend - that's checked separately
 * in /api/health/ready endpoint for external dependency monitoring.
 */
export async function GET() {
	const timestamp = new Date().toISOString();

	// Check database connectivity (critical for manager operations)
	try {
		await prisma.$queryRaw`SELECT 1`;

		// Manager is healthy
		return NextResponse.json({
			status: "healthy",
			timestamp,
			service: "manager",
			database: "connected",
			version: process.env.npm_package_version || "0.1.0",
			uptime: process.uptime(),
		});
	} catch (error) {
		// Database failure is critical - manager cannot operate
		return NextResponse.json(
			{
				status: "unhealthy",
				timestamp,
				service: "manager",
				database: "disconnected",
				error:
					error instanceof Error
						? error.message
						: "Database connection failed",
				version: process.env.npm_package_version || "0.1.0",
			},
			{ status: 503 },
		);
	}
}
