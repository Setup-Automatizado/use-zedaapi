/**
 * Health Check API Route
 *
 * Provides comprehensive health status for both container orchestration
 * and frontend dashboard monitoring.
 * Public endpoint - no authentication required.
 *
 * This endpoint is used by:
 * - Docker HEALTHCHECK (checks manager only)
 * - AWS ECS/ALB health checks (checks manager only)
 * - Kubernetes liveness probes (checks manager only)
 * - Frontend dashboard (shows full status including backend)
 *
 * @module app/api/health
 */

import { NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getHealth } from "@/lib/api/health";

/**
 * GET /api/health
 * Get comprehensive health status
 *
 * Checks:
 * 1. Manager service (Next.js process)
 * 2. Manager database (Prisma connection)
 * 3. Backend API (WhatsApp API health endpoint)
 *
 * Returns:
 * - 200 OK: Manager is healthy (backend status is informational only)
 * - 503 Service Unavailable: Manager has critical failure (database down)
 *
 * Note: Backend failures do NOT cause 503 response to prevent container restarts.
 * Container orchestration cares only about manager health, not backend health.
 */
export async function GET() {
	const timestamp = new Date().toISOString();
	const checks: Record<
		string,
		{ status: "healthy" | "degraded" | "unhealthy"; message?: string }
	> = {
		service: { status: "healthy" },
	};

	// Check database connectivity (critical for manager)
	try {
		await prisma.$queryRaw`SELECT 1`;
		checks.database = { status: "healthy" };
	} catch (error) {
		checks.database = {
			status: "unhealthy",
			message:
				error instanceof Error
					? error.message
					: "Database connection failed",
		};

		// Database is critical - return 503 (container should restart)
		return NextResponse.json(
			{
				status: "unhealthy",
				timestamp,
				checks,
				version: process.env.npm_package_version || "0.1.0",
			},
			{ status: 503 },
		);
	}

	// Check backend API connectivity (non-critical, informational for frontend)
	try {
		const backendHealth = await getHealth();
		checks.backend = {
			status: "healthy",
			message: `Backend status: ${backendHealth.status}`,
		};
	} catch (error) {
		// Backend failure is NOT critical for manager container health
		// Frontend will show degraded status, but container stays up
		checks.backend = {
			status: "degraded",
			message:
				error instanceof Error
					? error.message
					: "Backend API unreachable",
		};
	}

	// Determine overall status
	// Only database failures make status "unhealthy"
	// Backend failures only make it "degraded"
	const hasUnhealthy = Object.values(checks).some(
		(check) => check.status === "unhealthy",
	);
	const hasDegraded = Object.values(checks).some(
		(check) => check.status === "degraded",
	);

	const overallStatus = hasUnhealthy
		? "unhealthy"
		: hasDegraded
			? "degraded"
			: "healthy";

	// Always return 200 unless database is down
	// This prevents container restarts when only backend is down
	return NextResponse.json({
		status: overallStatus,
		timestamp,
		checks,
		version: process.env.npm_package_version || "0.1.0",
		uptime: process.uptime(),
	});
}
