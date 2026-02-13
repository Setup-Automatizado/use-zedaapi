import type { QueueJob, QueueStats } from "@/types/queue";

export async function getQueueJobs(
	instanceId: string,
	instanceToken: string,
): Promise<QueueJob[]> {
	const response = await fetch(
		`/api/instances/${instanceId}/queue?instanceToken=${encodeURIComponent(instanceToken)}`,
	);

	if (!response.ok) {
		throw new Error(`Failed to fetch queue jobs: ${response.statusText}`);
	}

	const data = await response.json();
	return data.jobs || [];
}

export async function getQueueCount(
	instanceId: string,
	instanceToken: string,
): Promise<number> {
	const response = await fetch(
		`/api/instances/${instanceId}/queue/count?instanceToken=${encodeURIComponent(instanceToken)}`,
	);

	if (!response.ok) {
		throw new Error(`Failed to fetch queue count: ${response.statusText}`);
	}

	const data = await response.json();
	return data.count || 0;
}

export async function getQueueStats(
	instanceId: string,
	instanceToken: string,
): Promise<QueueStats> {
	const jobs = await getQueueJobs(instanceId, instanceToken);

	const stats: QueueStats = {
		total: jobs.length,
		pending: jobs.filter((j) => j.status === "pending").length,
		processing: jobs.filter((j) => j.status === "processing").length,
		sent: jobs.filter((j) => j.status === "sent").length,
		failed: jobs.filter((j) => j.status === "failed").length,
		canceled: jobs.filter((j) => j.status === "canceled").length,
	};

	return stats;
}

export async function clearQueue(
	instanceId: string,
	instanceToken: string,
): Promise<void> {
	const response = await fetch(
		`/api/instances/${instanceId}/queue?instanceToken=${encodeURIComponent(instanceToken)}`,
		{
			method: "DELETE",
		},
	);

	if (!response.ok) {
		throw new Error(`Failed to clear queue: ${response.statusText}`);
	}
}
