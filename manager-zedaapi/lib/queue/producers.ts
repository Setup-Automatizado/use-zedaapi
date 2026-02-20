import { DEFAULT_JOB_OPTIONS, PRIORITIES, QUEUE_CONFIG } from "./config";
import { createLogger } from "./logger";
import type {
	AffiliatePayoutJobData,
	EmailSendingJobData,
	InstanceSyncJobData,
	NfseIssuanceJobData,
	SicrediBillingJobData,
	StripeWebhookJobData,
} from "./types";

const log = createLogger("producer");

export async function enqueueStripeWebhook(
	data: StripeWebhookJobData,
): Promise<void> {
	const { getStripeWebhooksQueue } = await import("./queues");
	const queue = getStripeWebhooksQueue();
	const result = await queue.add(`stripe-${data.type}`, data, {
		...DEFAULT_JOB_OPTIONS,
		attempts: QUEUE_CONFIG.STRIPE_WEBHOOKS.attempts,
		backoff: QUEUE_CONFIG.STRIPE_WEBHOOKS.backoff,
		priority: PRIORITIES.HIGH,
		jobId: `stripe-${data.eventId}`,
	});

	log.info("Stripe webhook job enqueued", {
		jobId: result.id,
		eventId: data.eventId,
		type: data.type,
	});
}

export async function enqueueNfseIssuance(
	data: NfseIssuanceJobData,
): Promise<void> {
	const { getNfseIssuanceQueue } = await import("./queues");
	const queue = getNfseIssuanceQueue();
	const result = await queue.add(`nfse-${data.action}`, data, {
		...DEFAULT_JOB_OPTIONS,
		attempts: QUEUE_CONFIG.NFSE_ISSUANCE.attempts,
		backoff: QUEUE_CONFIG.NFSE_ISSUANCE.backoff,
		priority: PRIORITIES.NORMAL,
	});

	log.info("NFS-e issuance job enqueued", {
		jobId: result.id,
		invoiceId: data.invoiceId,
		action: data.action,
	});
}

export async function enqueueEmailSending(
	data: EmailSendingJobData,
): Promise<void> {
	const { getEmailSendingQueue } = await import("./queues");
	const queue = getEmailSendingQueue();
	const result = await queue.add("send-email", data, {
		...DEFAULT_JOB_OPTIONS,
		attempts: QUEUE_CONFIG.EMAIL_SENDING.attempts,
		backoff: QUEUE_CONFIG.EMAIL_SENDING.backoff,
		priority: data.priority ?? PRIORITIES.NORMAL,
	});

	log.info("Email sending job enqueued", {
		jobId: result.id,
		template: data.template,
		to: data.to,
	});
}

export async function enqueueSicrediBilling(
	data: SicrediBillingJobData,
	opts?: { delay?: number },
): Promise<void> {
	const { getSicrediBillingQueue } = await import("./queues");
	const queue = getSicrediBillingQueue();
	const result = await queue.add(`sicredi-${data.action}`, data, {
		...DEFAULT_JOB_OPTIONS,
		attempts: QUEUE_CONFIG.SICREDI_BILLING.attempts,
		backoff: QUEUE_CONFIG.SICREDI_BILLING.backoff,
		priority: PRIORITIES.HIGH,
		...(opts?.delay ? { delay: opts.delay } : {}),
	});

	log.info("Sicredi billing job enqueued", {
		jobId: result.id,
		invoiceId: data.invoiceId,
		type: data.type,
		action: data.action,
		delay: opts?.delay,
	});
}

export async function enqueueInstanceSync(
	data: InstanceSyncJobData,
): Promise<void> {
	const { getInstanceSyncQueue } = await import("./queues");
	const queue = getInstanceSyncQueue();

	const jobId = data.syncAll
		? "sync-all"
		: data.instanceId
			? `sync-${data.instanceId}`
			: data.userId
				? `sync-user-${data.userId}`
				: undefined;

	const result = await queue.add("instance-sync", data, {
		...DEFAULT_JOB_OPTIONS,
		attempts: QUEUE_CONFIG.INSTANCE_SYNC.attempts,
		backoff: QUEUE_CONFIG.INSTANCE_SYNC.backoff,
		priority: PRIORITIES.LOW,
		jobId,
	});

	log.info("Instance sync job enqueued", {
		jobId: result.id,
		instanceId: data.instanceId,
		userId: data.userId,
		syncAll: data.syncAll,
	});
}

export async function enqueueAffiliatePayout(
	data: AffiliatePayoutJobData,
): Promise<void> {
	const { getAffiliatePayoutsQueue } = await import("./queues");
	const queue = getAffiliatePayoutsQueue();
	const result = await queue.add("affiliate-payout", data, {
		...DEFAULT_JOB_OPTIONS,
		attempts: QUEUE_CONFIG.AFFILIATE_PAYOUTS.attempts,
		backoff: QUEUE_CONFIG.AFFILIATE_PAYOUTS.backoff,
		priority: PRIORITIES.NORMAL,
		jobId: `payout-${data.affiliateId}-${Date.now()}`,
	});

	log.info("Affiliate payout job enqueued", {
		jobId: result.id,
		affiliateId: data.affiliateId,
		amount: data.amount,
		method: data.method,
	});
}
