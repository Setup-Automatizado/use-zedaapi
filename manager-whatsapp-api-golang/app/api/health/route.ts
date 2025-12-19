/**
 * Health Check API Route
 *
 * Provides basic service health status.
 * Public endpoint - no authentication required.
 *
 * @module app/api/health
 */

import { NextResponse } from 'next/server';
import { getHealth } from '@/lib/api/health';

/**
 * GET /api/health
 * Get overall service health status
 *
 * Returns basic health information. Always returns 200 if service is running.
 * Use for basic uptime monitoring.
 */
export async function GET() {
  try {
    const health = await getHealth();
    return NextResponse.json(health);
  } catch (error) {
    console.error('Error fetching health:', error);

    // Even if backend is down, return a degraded status
    return NextResponse.json(
      {
        status: 'degraded',
        timestamp: new Date().toISOString(),
        error: 'Failed to reach backend service',
      },
      { status: 503 }
    );
  }
}
