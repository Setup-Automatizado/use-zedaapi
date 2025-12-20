/**
 * Health check and monitoring API functions
 *
 * Provides endpoints for service health, readiness, and metrics.
 * These are public endpoints that don't require authentication.
 *
 * @module lib/api/health
 */

import "server-only";
import type { HealthResponse, ReadinessResponse } from "@/types";
import { api } from "./client";

/**
 * Gets overall service health status
 *
 * Returns basic health information. Always returns 200 if service is running.
 * Use for basic uptime monitoring.
 *
 * @returns Health status object
 */
export async function getHealth(): Promise<HealthResponse> {
	return api.get<HealthResponse>("/health");
}

/**
 * Gets service readiness status
 *
 * Performs deep health checks on critical dependencies:
 * - Database connectivity
 * - Redis connectivity
 * - S3/MinIO availability
 *
 * Returns 503 if any critical dependency is unavailable.
 * Use for load balancer readiness probes.
 *
 * @returns Readiness status with dependency details
 */
export async function getReadiness(): Promise<ReadinessResponse> {
	return api.get<ReadinessResponse>("/ready");
}

/**
 * Gets Prometheus metrics in text format
 *
 * Returns metrics in Prometheus exposition format for scraping.
 * Includes:
 * - HTTP request metrics (duration, status codes)
 * - Event processing metrics
 * - Message queue metrics
 * - Connection metrics
 *
 * @returns Plain text Prometheus metrics
 */
export async function getMetrics(): Promise<string> {
	return api.get<string>("/metrics");
}
