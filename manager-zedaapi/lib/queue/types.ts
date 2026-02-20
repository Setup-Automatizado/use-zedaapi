export interface StripeWebhookJobData {
	eventId: string;
	type: string;
	payload: Record<string, unknown>;
}

export interface NfseIssuanceJobData {
	invoiceId: string;
	action: "emit" | "cancel" | "query_status";
	motivo?: string;
	retryCount?: number;
}

export interface EmailSendingJobData {
	to: string;
	template: string;
	data: Record<string, unknown>;
	attachments?: Array<{
		filename: string;
		content: string;
		encoding?: string;
	}>;
	priority?: number;
}

export interface SicrediBillingJobData {
	invoiceId: string;
	type: "pix" | "boleto_hibrido";
	action: "create" | "check_status";
}

export interface InstanceSyncJobData {
	instanceId?: string;
	userId?: string;
	syncAll?: boolean;
}

export interface AffiliatePayoutJobData {
	affiliateId: string;
	amount: number;
	method: "stripe_transfer" | "manual";
}
