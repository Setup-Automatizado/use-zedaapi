import { Worker } from "bullmq";
import { createWorkerConnection } from "@/lib/queue/connection";
import { createLogger } from "@/lib/queue/logger";
import { QUEUE_NAMES, QUEUE_CONFIG } from "@/lib/queue/config";
import { processStripeWebhookJob } from "./processors/stripe-webhook";
import { processNfseJob } from "./processors/nfse.processor";
import { processEmailSendingJob } from "./processors/email-sending";
import { processSicrediBillingJob } from "./processors/sicredi-billing";
import { processInstanceSyncJob } from "./processors/instance-sync";
import { processAffiliatePayoutJob } from "./processors/affiliate-payout";
import type { StripeWebhookJobData } from "@/lib/queue/types";
import type { NfseIssuanceJobData } from "@/lib/queue/types";
import type { EmailSendingJobData } from "@/lib/queue/types";
import type { SicrediBillingJobData } from "@/lib/queue/types";
import type { InstanceSyncJobData } from "@/lib/queue/types";
import type { AffiliatePayoutJobData } from "@/lib/queue/types";

const log = createLogger("workers");

interface WorkerEntry {
	name: string;
	worker: Worker;
}

function createWorkers(): WorkerEntry[] {
	const workers: WorkerEntry[] = [];

	// Stripe Webhooks Worker
	const stripeWorker = new Worker<StripeWebhookJobData>(
		QUEUE_NAMES.STRIPE_WEBHOOKS,
		processStripeWebhookJob,
		{
			connection: createWorkerConnection(),
			concurrency: QUEUE_CONFIG.STRIPE_WEBHOOKS.concurrency,
			limiter: QUEUE_CONFIG.STRIPE_WEBHOOKS.rateLimit,
		},
	);
	workers.push({ name: "stripe-webhooks", worker: stripeWorker });

	// NFS-e Issuance Worker
	const nfseWorker = new Worker<NfseIssuanceJobData>(
		QUEUE_NAMES.NFSE_ISSUANCE,
		processNfseJob,
		{
			connection: createWorkerConnection(),
			concurrency: QUEUE_CONFIG.NFSE_ISSUANCE.concurrency,
			limiter: QUEUE_CONFIG.NFSE_ISSUANCE.rateLimit,
		},
	);
	workers.push({ name: "nfse-issuance", worker: nfseWorker });

	// Email Sending Worker
	const emailWorker = new Worker<EmailSendingJobData>(
		QUEUE_NAMES.EMAIL_SENDING,
		processEmailSendingJob,
		{
			connection: createWorkerConnection(),
			concurrency: QUEUE_CONFIG.EMAIL_SENDING.concurrency,
			limiter: QUEUE_CONFIG.EMAIL_SENDING.rateLimit,
		},
	);
	workers.push({ name: "email-sending", worker: emailWorker });

	// Sicredi Billing Worker
	const sicrediWorker = new Worker<SicrediBillingJobData>(
		QUEUE_NAMES.SICREDI_BILLING,
		processSicrediBillingJob,
		{
			connection: createWorkerConnection(),
			concurrency: QUEUE_CONFIG.SICREDI_BILLING.concurrency,
			limiter: QUEUE_CONFIG.SICREDI_BILLING.rateLimit,
		},
	);
	workers.push({ name: "sicredi-billing", worker: sicrediWorker });

	// Instance Sync Worker
	const instanceSyncWorker = new Worker<InstanceSyncJobData>(
		QUEUE_NAMES.INSTANCE_SYNC,
		processInstanceSyncJob,
		{
			connection: createWorkerConnection(),
			concurrency: QUEUE_CONFIG.INSTANCE_SYNC.concurrency,
			limiter: QUEUE_CONFIG.INSTANCE_SYNC.rateLimit,
		},
	);
	workers.push({ name: "instance-sync", worker: instanceSyncWorker });

	// Affiliate Payouts Worker
	const affiliateWorker = new Worker<AffiliatePayoutJobData>(
		QUEUE_NAMES.AFFILIATE_PAYOUTS,
		processAffiliatePayoutJob,
		{
			connection: createWorkerConnection(),
			concurrency: QUEUE_CONFIG.AFFILIATE_PAYOUTS.concurrency,
			limiter: QUEUE_CONFIG.AFFILIATE_PAYOUTS.rateLimit,
		},
	);
	workers.push({ name: "affiliate-payouts", worker: affiliateWorker });

	// Attach event listeners to all workers
	for (const { name, worker } of workers) {
		worker.on("completed", (job) => {
			log.info(`[${name}] Job completed`, {
				jobId: job.id,
				name: job.name,
			});
		});

		worker.on("failed", (job, error) => {
			log.error(`[${name}] Job failed`, {
				jobId: job?.id,
				name: job?.name,
				error: error.message,
				attemptsMade: job?.attemptsMade,
			});
		});

		worker.on("error", (error) => {
			log.error(`[${name}] Worker error`, { error: error.message });
		});

		worker.on("stalled", (jobId) => {
			log.warn(`[${name}] Job stalled`, { jobId });
		});
	}

	return workers;
}

// Start all workers
log.info("Starting all workers...");
const workers = createWorkers();
log.info(`Workers started: ${workers.map((w) => w.name).join(", ")}`);

// Graceful shutdown
let isShuttingDown = false;

async function shutdown(signal: string) {
	if (isShuttingDown) return;
	isShuttingDown = true;

	log.info(`Received ${signal}, shutting down workers...`);

	const closePromises = workers.map(async ({ name, worker }) => {
		log.info(`Closing ${name} worker...`);
		try {
			await worker.close();
			log.info(`${name} worker closed`);
		} catch (error) {
			log.error(`Error closing ${name} worker`, {
				error: error instanceof Error ? error.message : "Unknown error",
			});
		}
	});

	await Promise.all(closePromises);

	log.info("All workers shut down");
	process.exit(0);
}

process.on("SIGINT", () => shutdown("SIGINT"));
process.on("SIGTERM", () => shutdown("SIGTERM"));
process.on("uncaughtException", (error) => {
	log.error("Uncaught exception", {
		error: error.message,
		stack: error.stack,
	});
	shutdown("uncaughtException");
});
process.on("unhandledRejection", (reason) => {
	log.error("Unhandled rejection", {
		reason: reason instanceof Error ? reason.message : String(reason),
	});
});
