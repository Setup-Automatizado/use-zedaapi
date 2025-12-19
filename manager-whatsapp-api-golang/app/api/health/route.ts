/**
 * Health Check API Route
 *
 * Provides basic service health status for container orchestration.
 * Public endpoint - no authentication required.
 *
 * This endpoint is used by:
 * - Docker HEALTHCHECK
 * - AWS ECS/ALB health checks
 * - Kubernetes liveness probes
 *
 * @module app/api/health
 */

import { NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";

/**
 * GET /api/health
 * Get service health status
 *
 * Checks:
 * 1. Service is responding (always true if this code runs)
 * 2. Database connectivity (optional - returns degraded if DB fails)
 * 3. Backend API connectivity (optional - returns degraded if API fails)
 *
 * Returns:
 * - 200 OK: Service is healthy (may have degraded dependencies)
 * - 503 Service Unavailable: Critical failure (database unreachable)
 */
export async function GET() {
	const timestamp = new Date().toISOString();
	const checks: Record<
		string,
		{ status: "healthy" | "degraded" | "unhealthy"; message?: string }
	> = {
		service: { status: "healthy" },
	};

	// Check database connectivity (critical)
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

		// Database is critical - return 503
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

	// Check backend API connectivity (non-critical)
	try {
		const backendUrl = process.env.WHATSAPP_API_URL;
		if (backendUrl) {
			const controller = new AbortController();
			const timeoutId = setTimeout(() => controller.abort(), 3000);

			const response = await fetch(`${backendUrl}/health`, {
				signal: controller.signal,
				headers: { Accept: "application/json" },
			});

			clearTimeout(timeoutId);

			if (response.ok) {
				checks.backend = { status: "healthy" };
			} else {
				checks.backend = {
					status: "degraded",
					message: `Backend returned ${response.status}`,
				};
			}
		} else {
			checks.backend = {
				status: "degraded",
				message: "Backend URL not configured",
			};
		}
	} catch (error) {
		checks.backend = {
			status: "degraded",
			message:
				error instanceof Error ? error.message : "Backend unreachable",
		};
	}

	// Determine overall status
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

	// Return 200 even with degraded status (service is running)
	// Only critical failures (database) return 503
	return NextResponse.json({
		status: overallStatus,
		timestamp,
		checks,
		version: process.env.npm_package_version || "0.1.0",
		uptime: process.uptime(),
	});
}
