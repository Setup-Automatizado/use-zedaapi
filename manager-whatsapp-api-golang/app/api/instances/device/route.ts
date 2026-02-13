/**
 * Batch Device Info API Route
 *
 * Fetches device information (including avatar URLs) for multiple instances
 * in parallel. Only fetches for connected instances.
 *
 * POST /api/instances/device
 * Body: { instances: [{ instanceId, instanceToken, connected }] }
 * Response: { devices: { [instanceId]: DeviceInfo | null } }
 */

import { type NextRequest, NextResponse } from "next/server";

const API_BASE_URL = process.env.WHATSAPP_API_URL || "http://localhost:8080";
const CLIENT_TOKEN =
	process.env.WHATSAPP_CLIENT_TOKEN || process.env.CLIENT_AUTH_TOKEN || "";

interface InstanceRequest {
	instanceId: string;
	instanceToken: string;
	connected: boolean;
}

interface DeviceInfo {
	phone: string;
	imgUrl?: string;
	name: string;
	device: {
		sessionName: string;
		device_model: string;
		wa_version: string;
		platform: string;
		os_version: string;
		device_manufacturer: string;
		mcc?: string;
		mnc?: string;
	};
	isBusiness: boolean;
}

interface BatchRequest {
	instances: InstanceRequest[];
}

interface BatchResponse {
	devices: Record<string, DeviceInfo | null>;
}

async function fetchDeviceInfo(
	instanceId: string,
	instanceToken: string,
): Promise<DeviceInfo | null> {
	try {
		const url = `${API_BASE_URL}/instances/${instanceId}/token/${instanceToken}/device`;

		const response = await fetch(url, {
			method: "GET",
			headers: {
				"Content-Type": "application/json",
				"Client-Token": CLIENT_TOKEN,
			},
			cache: "no-store",
		});

		if (!response.ok) {
			return null;
		}

		return await response.json();
	} catch {
		return null;
	}
}

export async function POST(request: NextRequest) {
	try {
		const body: BatchRequest = await request.json();

		if (!body.instances || !Array.isArray(body.instances)) {
			return NextResponse.json(
				{
					error: "Invalid request body. Expected { instances: [...] }",
				},
				{ status: 400 },
			);
		}

		// Filter only connected instances
		const connectedInstances = body.instances.filter((i) => i.connected);

		// Limit concurrent requests to avoid overloading the backend
		const MAX_CONCURRENT = 10;
		const devices: Record<string, DeviceInfo | null> = {};

		// Initialize all as null
		for (const instance of body.instances) {
			devices[instance.instanceId] = null;
		}

		// Process in batches
		for (let i = 0; i < connectedInstances.length; i += MAX_CONCURRENT) {
			const batch = connectedInstances.slice(i, i + MAX_CONCURRENT);

			const results = await Promise.allSettled(
				batch.map((instance) =>
					fetchDeviceInfo(instance.instanceId, instance.instanceToken),
				),
			);

			results.forEach((result, index) => {
				const instance = batch[index];
				if (result.status === "fulfilled" && result.value) {
					devices[instance.instanceId] = result.value;
				}
			});
		}

		const response: BatchResponse = { devices };

		return NextResponse.json(response);
	} catch (error) {
		console.error("Error fetching device info batch:", error);
		return NextResponse.json(
			{ error: "Internal server error" },
			{ status: 500 },
		);
	}
}
