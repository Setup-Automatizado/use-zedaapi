import { Queue } from "bullmq";
import { getConnection } from "./connection";
import { QUEUE_NAMES } from "./config";
import type {
	AffiliatePayoutJobData,
	EmailSendingJobData,
	InstanceSyncJobData,
	NfseIssuanceJobData,
	SicrediBillingJobData,
	StripeWebhookJobData,
} from "./types";

const globalForQueues = global as unknown as {
	stripeWebhooksQueue: Queue | undefined;
	nfseIssuanceQueue: Queue | undefined;
	emailSendingQueue: Queue | undefined;
	sicrediBillingQueue: Queue | undefined;
	instanceSyncQueue: Queue | undefined;
	affiliatePayoutsQueue: Queue | undefined;
};

export function getStripeWebhooksQueue(): Queue<StripeWebhookJobData> {
	if (!globalForQueues.stripeWebhooksQueue) {
		globalForQueues.stripeWebhooksQueue = new Queue(
			QUEUE_NAMES.STRIPE_WEBHOOKS,
			{ connection: getConnection() },
		);
	}
	return globalForQueues.stripeWebhooksQueue as Queue<StripeWebhookJobData>;
}

export function getNfseIssuanceQueue(): Queue<NfseIssuanceJobData> {
	if (!globalForQueues.nfseIssuanceQueue) {
		globalForQueues.nfseIssuanceQueue = new Queue(
			QUEUE_NAMES.NFSE_ISSUANCE,
			{ connection: getConnection() },
		);
	}
	return globalForQueues.nfseIssuanceQueue as Queue<NfseIssuanceJobData>;
}

export function getEmailSendingQueue(): Queue<EmailSendingJobData> {
	if (!globalForQueues.emailSendingQueue) {
		globalForQueues.emailSendingQueue = new Queue(
			QUEUE_NAMES.EMAIL_SENDING,
			{ connection: getConnection() },
		);
	}
	return globalForQueues.emailSendingQueue as Queue<EmailSendingJobData>;
}

export function getSicrediBillingQueue(): Queue<SicrediBillingJobData> {
	if (!globalForQueues.sicrediBillingQueue) {
		globalForQueues.sicrediBillingQueue = new Queue(
			QUEUE_NAMES.SICREDI_BILLING,
			{ connection: getConnection() },
		);
	}
	return globalForQueues.sicrediBillingQueue as Queue<SicrediBillingJobData>;
}

export function getInstanceSyncQueue(): Queue<InstanceSyncJobData> {
	if (!globalForQueues.instanceSyncQueue) {
		globalForQueues.instanceSyncQueue = new Queue(
			QUEUE_NAMES.INSTANCE_SYNC,
			{ connection: getConnection() },
		);
	}
	return globalForQueues.instanceSyncQueue as Queue<InstanceSyncJobData>;
}

export function getAffiliatePayoutsQueue(): Queue<AffiliatePayoutJobData> {
	if (!globalForQueues.affiliatePayoutsQueue) {
		globalForQueues.affiliatePayoutsQueue = new Queue(
			QUEUE_NAMES.AFFILIATE_PAYOUTS,
			{ connection: getConnection() },
		);
	}
	return globalForQueues.affiliatePayoutsQueue as Queue<AffiliatePayoutJobData>;
}
