import type { StatusCacheStats } from "@/types/status-cache";

export async function getStatusCacheStats(
	instanceId: string,
	instanceToken: string,
): Promise<StatusCacheStats> {
	const response = await fetch(
		`/api/instances/${instanceId}/messages-status/stats?instanceToken=${encodeURIComponent(instanceToken)}`,
	);

	if (!response.ok) {
		throw new Error(
			`Failed to fetch status cache stats: ${response.statusText}`,
		);
	}

	return response.json();
}
