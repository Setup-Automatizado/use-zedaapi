import { db } from "@/lib/db";
import { ZedaAPIError, NotFoundError } from "@/lib/errors";
import { zedaapi } from "@/lib/zedaapi-client";
import { INSTANCE_STATUS } from "@/lib/constants";
import type {
	CreateInstanceRequest,
	InstanceStatusResponse,
	ZedaAPIInstance,
} from "@/types/zedaapi";

// =============================================================================
// ZedaAPI Service - Combines ZedaAPI calls with database operations
// =============================================================================

/**
 * Provision a new WhatsApp instance via ZedaAPI and record it in the DB.
 */
export async function provisionInstance(
	userId: string,
	subscriptionId: string,
	name: string,
): Promise<{ instanceId: string; zedaapiInstanceId: string }> {
	const request: CreateInstanceRequest = {
		name,
	};

	const response = await zedaapi.createInstance(request);

	// Activate subscription on ZedaAPI side
	await zedaapi.activateSubscription(response.id, response.token);

	// Store in our database
	const instance = await db.instance.create({
		data: {
			userId,
			subscriptionId,
			zedaapiInstanceId: response.id,
			zedaapiToken: response.token,
			name,
			status: INSTANCE_STATUS.CONNECTING,
		},
	});

	return {
		instanceId: instance.id,
		zedaapiInstanceId: response.id,
	};
}

/**
 * Deprovision an instance: cancel on ZedaAPI side and mark as deleted in DB.
 */
export async function deprovisionInstance(instanceId: string): Promise<void> {
	const instance = await db.instance.findUnique({
		where: { id: instanceId },
	});

	if (!instance) {
		throw new NotFoundError("Instance");
	}

	// Cancel subscription on ZedaAPI
	try {
		await zedaapi.cancelSubscription(
			instance.zedaapiInstanceId,
			instance.zedaapiToken,
		);
	} catch (error) {
		// If instance already gone on ZedaAPI side, continue with local cleanup
		if (!(error instanceof ZedaAPIError && error.statusCode === 404)) {
			throw error;
		}
	}

	// Mark as deleted
	await db.instance.update({
		where: { id: instanceId },
		data: {
			status: INSTANCE_STATUS.DELETED,
			whatsappConnected: false,
		},
	});
}

/**
 * Fetch live status from ZedaAPI and update the DB record.
 */
export async function syncInstanceStatus(
	instanceId: string,
): Promise<InstanceStatusResponse> {
	const instance = await db.instance.findUnique({
		where: { id: instanceId },
	});

	if (!instance) {
		throw new NotFoundError("Instance");
	}

	const status = await zedaapi.getStatus(
		instance.zedaapiInstanceId,
		instance.zedaapiToken,
	);

	// Determine local status based on ZedaAPI response
	let localStatus: string = INSTANCE_STATUS.DISCONNECTED;
	if (status.connected) {
		localStatus = INSTANCE_STATUS.CONNECTED;
	} else if (status.error) {
		localStatus = INSTANCE_STATUS.ERROR;
	}

	await db.instance.update({
		where: { id: instanceId },
		data: {
			status: localStatus,
			whatsappConnected: status.connected,
			lastSyncAt: new Date(),
		},
	});

	return status;
}

/**
 * Bulk sync all instances for a user.
 */
export async function syncAllInstances(userId: string): Promise<number> {
	const instances = await db.instance.findMany({
		where: {
			userId,
			status: { not: INSTANCE_STATUS.DELETED },
		},
		select: { id: true },
	});

	let synced = 0;
	for (const instance of instances) {
		try {
			await syncInstanceStatus(instance.id);
			synced++;
		} catch {
			// Individual sync failures should not block others
		}
	}

	return synced;
}

/**
 * Get instance from DB with live status from ZedaAPI.
 */
export async function getInstanceWithStatus(instanceId: string) {
	const instance = await db.instance.findUnique({
		where: { id: instanceId },
		include: {
			subscription: {
				include: { plan: true },
			},
		},
	});

	if (!instance) {
		throw new NotFoundError("Instance");
	}

	let liveStatus: InstanceStatusResponse | null = null;
	try {
		liveStatus = await zedaapi.getStatus(
			instance.zedaapiInstanceId,
			instance.zedaapiToken,
		);
	} catch {
		// If ZedaAPI unreachable, return cached data
	}

	return {
		...instance,
		liveStatus,
	};
}

/**
 * List remote ZedaAPI instances (partner-level).
 */
export async function listRemoteInstances(
	page = 1,
	pageSize = 20,
): Promise<{ instances: ZedaAPIInstance[]; total: number }> {
	const response = await zedaapi.listInstances({ page, pageSize });
	return {
		instances: response.content,
		total: response.total,
	};
}
