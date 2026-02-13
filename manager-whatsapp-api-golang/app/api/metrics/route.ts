/**
 * Metrics API Route
 *
 * Fetches raw Prometheus metrics from backend, parses and transforms
 * into dashboard-friendly format.
 *
 * GET /api/metrics
 * Query params:
 *   - instance_id: Filter by specific instance
 *   - format: 'raw' | 'dashboard' (default: dashboard)
 *
 * @module app/api/metrics/route
 */

import { NextResponse } from "next/server";
import { parsePrometheusMetrics, transformToDashboard } from "@/lib/metrics";
import type { MetricsResponse, ParsedMetrics } from "@/types/metrics";

/**
 * Base API URL from environment
 */
const API_BASE_URL = process.env.WHATSAPP_API_URL || "http://localhost:8080";

/**
 * Fetch raw Prometheus metrics from backend
 */
async function fetchRawMetrics(): Promise<string> {
	const response = await fetch(`${API_BASE_URL}/metrics`, {
		method: "GET",
		headers: {
			Accept: "text/plain",
		},
		cache: "no-store",
	});

	if (!response.ok) {
		throw new Error(`Failed to fetch metrics: ${response.status} ${response.statusText}`);
	}

	return response.text();
}

/**
 * GET /api/metrics
 *
 * Fetches and transforms Prometheus metrics
 */
export async function GET(request: Request): Promise<NextResponse<MetricsResponse | ParsedMetrics>> {
	const { searchParams } = new URL(request.url);
	const instanceId = searchParams.get("instance_id");
	const format = searchParams.get("format") || "dashboard";

	try {
		// Fetch raw Prometheus metrics
		const rawMetrics = await fetchRawMetrics();

		// Parse Prometheus format
		const parsed = parsePrometheusMetrics(rawMetrics);

		// Return raw parsed format if requested
		if (format === "raw") {
			return NextResponse.json(parsed);
		}

		// Transform to dashboard format
		const dashboard = transformToDashboard(parsed, {
			instanceId: instanceId || undefined,
		});

		const response: MetricsResponse = {
			success: true,
			data: dashboard,
			timestamp: new Date().toISOString(),
		};

		return NextResponse.json(response);
	} catch (error) {
		console.error("Failed to fetch metrics:", error);

		const response: MetricsResponse = {
			success: false,
			error: error instanceof Error ? error.message : "Failed to fetch metrics",
			timestamp: new Date().toISOString(),
		};

		return NextResponse.json(response, { status: 500 });
	}
}

/**
 * Revalidation settings
 */
export const dynamic = "force-dynamic";
export const revalidate = 0;
