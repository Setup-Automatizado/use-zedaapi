import type { InstanceMetrics } from "@/types/metrics";

const API_BASE_URL =
	process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:3333";

export async function getPrometheusMetrics(): Promise<string> {
	const response = await fetch(`${API_BASE_URL}/metrics`);

	if (!response.ok) {
		throw new Error(`Failed to fetch metrics: ${response.statusText}`);
	}

	return response.text();
}

export async function parseInstanceMetrics(
	prometheusText: string,
	instanceId: string,
): Promise<InstanceMetrics> {
	// Parse Prometheus format (simplified)
	const lines = prometheusText.split("\n");

	const getMetricValue = (metricName: string): number => {
		const regex = new RegExp(
			`${metricName}\\{.*instance_id="${instanceId}".*\\}\\s+(\\d+\\.?\\d*)`,
		);
		for (const line of lines) {
			const match = line.match(regex);
			if (match) return parseFloat(match[1]);
		}
		return 0;
	};

	return {
		messages_sent: getMetricValue("messages_sent_total"),
		messages_received: getMetricValue("messages_received_total"),
		messages_failed: getMetricValue("messages_failed_total"),
		avg_latency_ms: getMetricValue("message_send_duration_milliseconds"),
		transport_errors: getMetricValue("transport_errors_total"),
	};
}
