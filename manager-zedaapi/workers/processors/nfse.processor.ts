import type { Job } from "bullmq";
import type { NfseIssuanceJobData } from "@/lib/queue/types";
import { createLogger } from "@/lib/queue/logger";

const log = createLogger("processor:nfse");

/**
 * Process NFS-e jobs -- emit, cancel, or query status.
 * Delegates to the NFS-e service layer.
 */
export async function processNfseJob(
	job: Job<NfseIssuanceJobData>,
): Promise<void> {
	const { invoiceId, action, motivo } = job.data;

	log.info("Processing NFS-e job", {
		jobId: job.id,
		invoiceId,
		action,
		hasMotivo: !!motivo,
	});

	const done = log.timer(`NFS-e ${action}`, { invoiceId });
	const { db } = await import("@/lib/db");

	switch (action) {
		case "emit": {
			log.info("Emitting NFS-e", { invoiceId });

			// Load invoice with user data for tomador
			const invoice = await db.invoice.findUnique({
				where: { id: invoiceId },
				include: { user: true },
			});

			if (!invoice) {
				log.error("Invoice not found", { invoiceId });
				throw new Error(`Invoice ${invoiceId} not found`);
			}

			if (invoice.nfseStatus === "emitted") {
				log.info("NFS-e already emitted, skipping", { invoiceId });
				return;
			}

			try {
				// Dynamic import to avoid loading NFS-e service when not needed
				const { emitNfse } =
					await import("@/lib/services/nfse/service");
				const result = await emitNfse(invoiceId);

				if (result.success) {
					done();
					log.info("NFS-e emitted successfully", {
						invoiceId,
						chaveAcesso: result.chaveAcesso ?? "?",
						numero: result.numero ?? "?",
					});
				} else {
					log.error("NFS-e emission failed", {
						invoiceId,
						error: result.error,
					});

					// Server errors are retryable
					if (result.error?.includes("SEFIN")) {
						log.warn("SEFIN error -- will retry via BullMQ", {
							invoiceId,
						});
						throw new Error(result.error);
					}
				}
			} catch (error) {
				const msg =
					error instanceof Error ? error.message : "Unknown error";
				await db.invoice.update({
					where: { id: invoiceId },
					data: { nfseStatus: "error", nfseError: msg },
				});
				throw error;
			}
			break;
		}

		case "cancel": {
			if (!motivo) {
				log.error("Missing motivo for cancellation", { invoiceId });
				throw new Error("Motivo is required for NFS-e cancellation");
			}

			log.info("Cancelling NFS-e", { invoiceId, motivo });

			try {
				const { cancelNfse } =
					await import("@/lib/services/nfse/service");
				const result = await cancelNfse(invoiceId, motivo);

				if (result.success) {
					done();
					log.info("NFS-e cancelled successfully", { invoiceId });
				} else {
					log.error("NFS-e cancellation failed", {
						invoiceId,
						error: result.error,
					});
				}
			} catch (error) {
				const msg =
					error instanceof Error ? error.message : "Unknown error";
				await db.invoice.update({
					where: { id: invoiceId },
					data: { nfseError: msg },
				});
				throw error;
			}
			break;
		}

		case "query_status": {
			log.info("Querying NFS-e status", { invoiceId });

			try {
				const { queryNfseStatus } =
					await import("@/lib/services/nfse/service");
				await queryNfseStatus(invoiceId);
				done();
				log.info("NFS-e status query completed", { invoiceId });
			} catch (error) {
				log.error("NFS-e status query failed", {
					invoiceId,
					error:
						error instanceof Error
							? error.message
							: "Unknown error",
				});
				throw error;
			}
			break;
		}

		default:
			log.error("Unknown NFS-e action", { action, invoiceId });
	}
}
