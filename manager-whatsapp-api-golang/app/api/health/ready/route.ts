/**
 * Readiness Check API Route
 *
 * Performs deep health checks on critical dependencies.
 * Public endpoint - no authentication required.
 *
 * @module app/api/health/ready
 */

import { NextResponse } from "next/server";
import { getReadiness } from "@/lib/api/health";

/**
 * GET /api/health/ready
 * Get service readiness status
 *
 * Performs deep health checks on critical dependencies:
 * - Database connectivity
 * - Redis connectivity
 * - S3/MinIO availability
 *
 * Returns 503 if any critical dependency is unavailable.
 * Use for load balancer readiness probes.
 */
export async function GET() {
	try {
		const readiness = await getReadiness();

		// Check if service is ready based on the ready flag
		const isReady = readiness.ready;

		return NextResponse.json(readiness, { status: isReady ? 200 : 503 });
	} catch (error) {
		console.error("Error fetching readiness:", error);

		return NextResponse.json(
			{
				status: "unavailable",
				timestamp: new Date().toISOString(),
				error: "Failed to perform readiness check",
			},
			{ status: 503 },
		);
	}
}
