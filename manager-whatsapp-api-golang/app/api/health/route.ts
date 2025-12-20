/**
 * Health Check API Route
 *
 * Provides health status compatible with both container orchestration
 * and frontend dashboard monitoring.
 * Public endpoint - no authentication required.
 *
 * This endpoint proxies the backend health check and adds manager status.
 *
 * @module app/api/health
 */

import { NextResponse } from "next/server";
import { getHealth } from "@/lib/api/health";
import prisma from "@/lib/prisma";

/**
 * GET /api/health
 * Get service health status
 *
 * Returns backend health status if available, otherwise returns manager-only status.
 * Always returns 200 unless manager database is down (503).
 *
 * This ensures:
 * - Frontend gets backend health data
 * - Docker/ECS health checks work correctly
 * - Manager database failures cause container restart
 * - Backend failures don't cause container restart
 */
export async function GET() {
	const timestamp = new Date().toISOString();

	// Check manager database first (critical)
	try {
		await prisma.$queryRaw`SELECT 1`;
	} catch (error) {
		// Manager database down = container should restart
		return NextResponse.json(
			{
				status: "unhealthy",
				service: "manager",
				timestamp,
				error:
					error instanceof Error ? error.message : "Database connection failed",
			},
			{ status: 503 },
		);
	}

	// Try to get backend health (non-critical)
	try {
		const backendHealth = await getHealth();
		// Return backend health response
		return NextResponse.json(backendHealth);
	} catch {
		// Backend is down, but manager is still healthy
		// Return ok status so frontend doesn't show error
		// Frontend will show "API Offline" based on backend status
		return NextResponse.json({
			status: "ok",
			service: "manager",
			timestamp,
		});
	}
}
