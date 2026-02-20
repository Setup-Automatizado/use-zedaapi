"use server";

import { z } from "zod";
import { revalidatePath } from "next/cache";
import { requireAuth } from "@/lib/auth-server";
import { db } from "@/lib/db";
import { zedaapi } from "@/lib/zedaapi-client";
import { ZedaAPIError, NotFoundError } from "@/lib/errors";
import { INSTANCE_STATUS, ROUTES } from "@/lib/constants";
import {
	provisionInstance,
	deprovisionInstance,
	syncInstanceStatus,
	getInstanceWithStatus,
} from "@/server/services/zedaapi-service";
import type { ActionResult } from "@/types";

// =============================================================================
// Validation schemas
// =============================================================================

const createInstanceSchema = z.object({
	name: z.string().min(2, "Name must be at least 2 characters").max(64),
	subscriptionId: z.string().min(1, "Subscription is required"),
});

// =============================================================================
// Server Actions
// =============================================================================

export async function createInstance(
	formData: FormData,
): Promise<ActionResult<{ instanceId: string }>> {
	const session = await requireAuth();

	const raw = {
		name: formData.get("name"),
		subscriptionId: formData.get("subscriptionId"),
	};

	const validation = createInstanceSchema.safeParse(raw);
	if (!validation.success) {
		return {
			success: false,
			errors: validation.error.flatten().fieldErrors,
		};
	}

	const { name, subscriptionId } = validation.data;

	const subscription = await db.subscription.findFirst({
		where: {
			id: subscriptionId,
			userId: session.user.id,
			status: "active",
		},
		include: {
			plan: true,
			instances: { where: { status: { not: INSTANCE_STATUS.DELETED } } },
		},
	});

	if (!subscription) {
		return { success: false, error: "Active subscription not found" };
	}

	if (subscription.instances.length >= subscription.plan.maxInstances) {
		return {
			success: false,
			error: `Instance limit reached (${subscription.plan.maxInstances} max)`,
		};
	}

	try {
		const result = await provisionInstance(
			session.user.id,
			subscriptionId,
			name,
		);

		revalidatePath(ROUTES.INSTANCES);

		return { success: true, data: { instanceId: result.instanceId } };
	} catch (error) {
		if (error instanceof ZedaAPIError) {
			return { success: false, error: `ZedaAPI error: ${error.message}` };
		}
		return { success: false, error: "Failed to create instance" };
	}
}

export async function deleteInstance(
	instanceId: string,
): Promise<ActionResult> {
	const session = await requireAuth();

	const instance = await db.instance.findFirst({
		where: { id: instanceId, userId: session.user.id },
	});

	if (!instance) {
		return { success: false, error: "Instance not found" };
	}

	try {
		await deprovisionInstance(instanceId);
		revalidatePath(ROUTES.INSTANCES);
		return { success: true };
	} catch (error) {
		if (error instanceof ZedaAPIError) {
			return { success: false, error: `ZedaAPI error: ${error.message}` };
		}
		return { success: false, error: "Failed to delete instance" };
	}
}

export async function getInstances(): Promise<
	ActionResult<
		Array<{
			id: string;
			name: string;
			status: string;
			phone: string | null;
			whatsappConnected: boolean;
			createdAt: Date;
		}>
	>
> {
	const session = await requireAuth();

	const instances = await db.instance.findMany({
		where: {
			userId: session.user.id,
			status: { not: INSTANCE_STATUS.DELETED },
		},
		select: {
			id: true,
			name: true,
			status: true,
			phone: true,
			whatsappConnected: true,
			profileName: true,
			profilePicUrl: true,
			lastSyncAt: true,
			createdAt: true,
		},
		orderBy: { createdAt: "desc" },
	});

	return { success: true, data: instances };
}

export async function getInstance(
	instanceId: string,
): Promise<ActionResult<Awaited<ReturnType<typeof getInstanceWithStatus>>>> {
	const session = await requireAuth();

	const owned = await db.instance.findFirst({
		where: { id: instanceId, userId: session.user.id },
		select: { id: true },
	});

	if (!owned) {
		return { success: false, error: "Instance not found" };
	}

	try {
		const instance = await getInstanceWithStatus(instanceId);
		return { success: true, data: instance };
	} catch (error) {
		if (error instanceof NotFoundError) {
			return { success: false, error: "Instance not found" };
		}
		return { success: false, error: "Failed to fetch instance" };
	}
}

export async function getQRCode(
	instanceId: string,
): Promise<ActionResult<{ qrCode: string }>> {
	const session = await requireAuth();

	const instance = await db.instance.findFirst({
		where: { id: instanceId, userId: session.user.id },
	});

	if (!instance) {
		return { success: false, error: "Instance not found" };
	}

	try {
		const response = await zedaapi.getQRCode(
			instance.zedaapiInstanceId,
			instance.zedaapiToken,
		);
		return { success: true, data: { qrCode: response.value } };
	} catch (error) {
		if (error instanceof ZedaAPIError) {
			return { success: false, error: `ZedaAPI error: ${error.message}` };
		}
		return { success: false, error: "Failed to get QR code" };
	}
}

export async function restartInstance(
	instanceId: string,
): Promise<ActionResult> {
	const session = await requireAuth();

	const instance = await db.instance.findFirst({
		where: { id: instanceId, userId: session.user.id },
	});

	if (!instance) {
		return { success: false, error: "Instance not found" };
	}

	try {
		await zedaapi.restart(
			instance.zedaapiInstanceId,
			instance.zedaapiToken,
		);

		await db.instance.update({
			where: { id: instanceId },
			data: { status: INSTANCE_STATUS.CONNECTING },
		});

		revalidatePath(ROUTES.INSTANCE_DETAIL(instanceId));
		return { success: true };
	} catch (error) {
		if (error instanceof ZedaAPIError) {
			return { success: false, error: `ZedaAPI error: ${error.message}` };
		}
		return { success: false, error: "Failed to restart instance" };
	}
}

export async function disconnectInstance(
	instanceId: string,
): Promise<ActionResult> {
	const session = await requireAuth();

	const instance = await db.instance.findFirst({
		where: { id: instanceId, userId: session.user.id },
	});

	if (!instance) {
		return { success: false, error: "Instance not found" };
	}

	try {
		await zedaapi.disconnect(
			instance.zedaapiInstanceId,
			instance.zedaapiToken,
		);

		await db.instance.update({
			where: { id: instanceId },
			data: {
				status: INSTANCE_STATUS.DISCONNECTED,
				whatsappConnected: false,
			},
		});

		revalidatePath(ROUTES.INSTANCE_DETAIL(instanceId));
		return { success: true };
	} catch (error) {
		if (error instanceof ZedaAPIError) {
			return { success: false, error: `ZedaAPI error: ${error.message}` };
		}
		return { success: false, error: "Failed to disconnect instance" };
	}
}

export async function syncInstance(instanceId: string): Promise<ActionResult> {
	const session = await requireAuth();

	const instance = await db.instance.findFirst({
		where: { id: instanceId, userId: session.user.id },
		select: { id: true },
	});

	if (!instance) {
		return { success: false, error: "Instance not found" };
	}

	try {
		await syncInstanceStatus(instanceId);
		revalidatePath(ROUTES.INSTANCE_DETAIL(instanceId));
		return { success: true };
	} catch (error) {
		if (error instanceof ZedaAPIError) {
			return { success: false, error: `ZedaAPI error: ${error.message}` };
		}
		return { success: false, error: "Failed to sync instance" };
	}
}
