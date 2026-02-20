export const QUEUE_NAMES = {
	STRIPE_WEBHOOKS: "stripe-webhooks",
	NFSE_ISSUANCE: "nfse-issuance",
	EMAIL_SENDING: "email-sending",
	SICREDI_BILLING: "sicredi-billing",
	INSTANCE_SYNC: "instance-sync",
	AFFILIATE_PAYOUTS: "affiliate-payouts",
} as const;

export type QueueName = (typeof QUEUE_NAMES)[keyof typeof QUEUE_NAMES];

export const PRIORITIES = {
	CRITICAL: 1,
	HIGH: 2,
	NORMAL: 5,
	LOW: 10,
} as const;

export const DEFAULT_JOB_OPTIONS = {
	attempts: 3,
	backoff: {
		type: "exponential" as const,
		delay: 30_000,
	},
	removeOnComplete: {
		count: 1000,
		age: 86_400,
	},
	removeOnFail: {
		count: 5000,
		age: 604_800,
	},
};

export const QUEUE_CONFIG = {
	STRIPE_WEBHOOKS: {
		concurrency: 5,
		attempts: 5,
		backoff: { type: "exponential" as const, delay: 5_000 },
		rateLimit: { max: 50, duration: 60_000 },
	},
	NFSE_ISSUANCE: {
		concurrency: 2,
		attempts: 3,
		backoff: { type: "exponential" as const, delay: 60_000 },
		rateLimit: { max: 10, duration: 60_000 },
	},
	EMAIL_SENDING: {
		concurrency: 10,
		attempts: 3,
		backoff: { type: "exponential" as const, delay: 10_000 },
		rateLimit: { max: 80, duration: 60_000 },
	},
	SICREDI_BILLING: {
		concurrency: 2,
		attempts: 3,
		backoff: { type: "exponential" as const, delay: 30_000 },
		rateLimit: { max: 20, duration: 60_000 },
	},
	INSTANCE_SYNC: {
		concurrency: 1,
		attempts: 2,
		backoff: { type: "fixed" as const, delay: 15_000 },
		rateLimit: { max: 30, duration: 60_000 },
	},
	AFFILIATE_PAYOUTS: {
		concurrency: 1,
		attempts: 3,
		backoff: { type: "exponential" as const, delay: 30_000 },
		rateLimit: { max: 10, duration: 60_000 },
	},
} as const;
