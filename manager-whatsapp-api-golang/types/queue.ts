export interface QueueJob {
	id: string;
	type: string;
	status: "pending" | "processing" | "sent" | "failed" | "canceled";
	payload: Record<string, unknown>;
	attempt: number;
	max_attempts: number;
	errors?: string[];
	created_at: string;
	updated_at: string;
}

export interface QueueStats {
	total: number;
	pending: number;
	processing: number;
	sent: number;
	failed: number;
	canceled: number;
}

export interface QueueResponse {
	jobs: QueueJob[];
	count: number;
	stats: QueueStats;
}
