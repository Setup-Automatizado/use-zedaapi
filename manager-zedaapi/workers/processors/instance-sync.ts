import type { Job } from "bullmq";
import type { InstanceSyncJobData } from "@/lib/queue/types";
import type { db as dbClient } from "@/lib/db";
import { createLogger } from "@/lib/queue/logger";

type DbClient = typeof dbClient;

const log = createLogger("processor:instance-sync");

async function sendEmailNotification(
	to: string,
	template: string,
	data: Record<string, unknown>,
): Promise<void> {
	try {
		const { enqueueEmailSending } = await import("@/lib/queue/producers");
		await enqueueEmailSending({ to, template, data });
	} catch (error) {
		log.error("Failed to enqueue email notification", {
			template,
			to,
			error: error instanceof Error ? error.message : "Unknown error",
		});
	}
}

export async function processInstanceSyncJob(
	job: Job<InstanceSyncJobData>,
): Promise<void> {
	const { instanceId, userId, syncAll } = job.data;

	log.info("Processing instance sync job", {
		jobId: job.id,
		instanceId,
		userId,
		syncAll,
	});

	const done = log.timer("Instance sync", { instanceId, userId, syncAll });
	const { db } = await import("@/lib/db");

	if (syncAll) {
		await syncAllInstances(db);
	} else if (instanceId) {
		await syncSingleInstance(instanceId, db);
	} else if (userId) {
		await syncUserInstances(userId, db);
	} else {
		log.warn("No sync target specified");
		return;
	}

	done();
}

async function syncSingleInstance(
	instanceId: string,
	db: DbClient,
): Promise<void> {
	const instance = await db.instance.findUnique({
		where: { id: instanceId },
	});

	if (!instance) {
		log.warn("Instance not found", { instanceId });
		return;
	}

	await fetchAndUpdateStatus(instance, db);
}

async function syncUserInstances(userId: string, db: DbClient): Promise<void> {
	const instances = await db.instance.findMany({
		where: { userId },
	});

	log.info("Syncing user instances", { userId, count: instances.length });

	for (const instance of instances) {
		await fetchAndUpdateStatus(instance, db);
	}
}

async function syncAllInstances(db: DbClient): Promise<void> {
	const instances = await db.instance.findMany();

	log.info("Syncing all instances", { count: instances.length });

	for (const instance of instances) {
		await fetchAndUpdateStatus(instance, db);
	}
}

async function fetchAndUpdateStatus(
	instance: { id: string; zedaapiInstanceId: string; status: string },
	db: DbClient,
): Promise<void> {
	const baseUrl = process.env.ZEDAAPI_BASE_URL;
	const token = process.env.ZEDAAPI_CLIENT_TOKEN;

	if (!baseUrl || !token) {
		log.error("ZedaAPI config missing", { instanceId: instance.id });
		return;
	}

	try {
		const response = await fetch(
			`${baseUrl}/instances/${instance.zedaapiInstanceId}/status`,
			{
				headers: {
					Authorization: `Bearer ${token}`,
					"Content-Type": "application/json",
				},
				signal: AbortSignal.timeout(10_000),
			},
		);

		if (!response.ok) {
			log.warn("ZedaAPI status request failed", {
				instanceId: instance.id,
				status: response.status,
			});

			if (response.status === 404) {
				await db.instance.update({
					where: { id: instance.id },
					data: {
						status: "not_found",
						whatsappConnected: false,
						lastSyncAt: new Date(),
					},
				});
			}
			return;
		}

		const data = (await response.json()) as {
			connected?: boolean;
			status?: string;
			phone?: string;
			profileName?: string;
			profilePicUrl?: string;
		};

		const newStatus = data.connected ? "connected" : "disconnected";
		const statusChanged = instance.status !== newStatus;

		await db.instance.update({
			where: { id: instance.id },
			data: {
				status: newStatus,
				whatsappConnected: data.connected ?? false,
				phone: data.phone ?? undefined,
				profileName: data.profileName ?? undefined,
				profilePicUrl: data.profilePicUrl ?? undefined,
				lastSyncAt: new Date(),
			},
		});

		if (statusChanged) {
			log.info("Instance status changed", {
				instanceId: instance.id,
				from: instance.status,
				to: newStatus,
			});

			// Send notification when instance disconnects
			if (
				newStatus === "disconnected" &&
				instance.status === "connected"
			) {
				const fullInstance = await db.instance.findUnique({
					where: { id: instance.id },
					include: {
						user: { select: { email: true, name: true } },
					},
				});

				if (fullInstance?.user?.email) {
					await sendEmailNotification(
						fullInstance.user.email,
						"instance-connected",
						{
							userName: fullInstance.user.name || "Usuário",
							userId: fullInstance.userId,
							instanceName: fullInstance.name,
							phone: fullInstance.phone || "",
							status: "disconnected",
							reconnectUrl: `${process.env.NEXT_PUBLIC_APP_URL}/instancias`,
						},
					);
				}
			}
		} else {
			log.debug("Instance status unchanged", {
				instanceId: instance.id,
				status: newStatus,
			});
		}
	} catch (error) {
		log.error("Failed to sync instance", {
			instanceId: instance.id,
			error: error instanceof Error ? error.message : "Unknown error",
		});
	}
}
