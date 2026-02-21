import { sicrediPix } from "./client";
import { createLogger } from "@/lib/logger";

const log = createLogger("service:sicredi-webhook");

// =============================================================================
// SICREDI WEBHOOK MANAGEMENT (via API PIX mTLS)
// =============================================================================

interface WebhookResponse {
	webhookUrl: string;
	chave: string;
	criacao: string;
}

/**
 * Register webhook URL for a PIX key.
 * PUT /api/v2/webhook/{chave}
 */
export async function registerWebhook(
	pixKey: string,
	webhookUrl: string,
): Promise<void> {
	await sicrediPix.request("PUT", `/api/v2/webhook/${pixKey}`, {
		webhookUrl,
	});
}

/**
 * Get webhook configuration for a PIX key.
 * GET /api/v2/webhook/{chave}
 */
export async function getWebhook(pixKey: string): Promise<WebhookResponse> {
	return sicrediPix.request<WebhookResponse>(
		"GET",
		`/api/v2/webhook/${pixKey}`,
	);
}

/**
 * Delete webhook for a PIX key.
 * DELETE /api/v2/webhook/{chave}
 */
export async function deleteWebhook(pixKey: string): Promise<void> {
	await sicrediPix.request("DELETE", `/api/v2/webhook/${pixKey}`);
}

/**
 * Register webhook using environment variables.
 * Convenience function for setup.
 */
export async function registerSicrediWebhook(): Promise<void> {
	const pixKey = process.env.SICREDI_PIX_KEY;
	const baseUrl = process.env.BETTER_AUTH_URL;

	if (!pixKey) throw new Error("SICREDI_PIX_KEY not configured");
	if (!baseUrl) throw new Error("BETTER_AUTH_URL not configured");

	const webhookUrl = `${baseUrl}/api/webhooks/sicredi`;
	await registerWebhook(pixKey, webhookUrl);

	log.info("Sicredi webhook registered", { webhookUrl });
}
