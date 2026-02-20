export interface WebhookPayload {
	event: "contact_form" | "whatsapp_widget" | "data_deletion_request";
	timestamp: string;
	data: Record<string, unknown>;
	utm: {
		source: string | null;
		medium: string | null;
		campaign: string | null;
		term: string | null;
		content: string | null;
	};
	metadata: {
		page_url: string | null;
		referrer: string | null;
		user_agent: string | null;
	};
}

export async function sendWebhook(payload: WebhookPayload): Promise<void> {
	const url = process.env.CONTACT_WEBHOOK_URL;
	if (!url) return;

	await fetch(url, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
		signal: AbortSignal.timeout(10_000),
	}).catch(() => {});
}
