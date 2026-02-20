"use server";

import { db } from "@/lib/db";
import { zedaapi } from "@/lib/zedaapi-client";
import { AppError, NotFoundError } from "@/lib/errors";
import type { PaginatedResult } from "@/types";

// =============================================================================
// Instance Management Service
// =============================================================================

/**
 * Provision a new WhatsApp instance via ZedaAPI and persist locally.
 */
export async function provisionInstance(
	userId: string,
	subscriptionId: string,
	name: string,
) {
	await checkInstanceLimit(userId);

	const result = await zedaapi.createInstance({ name });

	const instance = await db.instance.create({
		data: {
			userId,
			subscriptionId,
			zedaapiInstanceId: result.id,
			zedaapiToken: result.token,
			name,
			status: "disconnected",
		},
	});

	await zedaapi.activateSubscription(result.id, result.token);

	return instance;
}

/**
 * Deprovision an instance: cancel on ZedaAPI and soft-delete locally.
 */
export async function deprovisionInstance(instanceId: string) {
	const instance = await db.instance.findUnique({
		where: { id: instanceId },
		select: { zedaapiInstanceId: true, zedaapiToken: true },
	});

	if (!instance) {
		throw new NotFoundError("Instance");
	}

	try {
		await zedaapi.cancelSubscription(
			instance.zedaapiInstanceId,
			instance.zedaapiToken,
		);
	} catch {
		// Continue even if ZedaAPI fails — instance may already be gone
	}

	await db.instance.update({
		where: { id: instanceId },
		data: { status: "deleted" },
	});
}

/**
 * Sync instance status from ZedaAPI to local database.
 */
export async function syncInstanceStatus(instanceId: string) {
	const instance = await db.instance.findUnique({
		where: { id: instanceId },
		select: { zedaapiInstanceId: true, zedaapiToken: true },
	});

	if (!instance) {
		throw new NotFoundError("Instance");
	}

	const status = await zedaapi.getStatus(
		instance.zedaapiInstanceId,
		instance.zedaapiToken,
	);

	const updatedInstance = await db.instance.update({
		where: { id: instanceId },
		data: {
			whatsappConnected: status.connected,
			status: status.connected ? "connected" : "disconnected",
			lastSyncAt: new Date(),
		},
	});

	return updatedInstance;
}

/**
 * List instances for a user with pagination.
 */
export async function getInstancesForUser(
	userId: string,
	page = 1,
	pageSize = 10,
): Promise<
	PaginatedResult<Awaited<ReturnType<typeof db.instance.findMany>>[0]>
> {
	const skip = (page - 1) * pageSize;

	const [items, total] = await Promise.all([
		db.instance.findMany({
			where: { userId, status: { not: "deleted" } },
			include: {
				subscription: { select: { plan: { select: { name: true } } } },
			},
			orderBy: { createdAt: "desc" },
			skip,
			take: pageSize,
		}),
		db.instance.count({
			where: { userId, status: { not: "deleted" } },
		}),
	]);

	const totalPages = Math.ceil(total / pageSize);

	return {
		items,
		total,
		page,
		pageSize,
		totalPages,
		hasMore: page < totalPages,
	};
}

/**
 * Get a single instance by ID.
 */
export async function getInstanceById(instanceId: string) {
	const instance = await db.instance.findUnique({
		where: { id: instanceId },
		include: {
			subscription: {
				select: {
					plan: { select: { name: true, maxInstances: true } },
				},
			},
		},
	});

	if (!instance) {
		throw new NotFoundError("Instance");
	}

	return instance;
}

/**
 * Check if the user has reached their instance limit based on their active subscription.
 */
export async function checkInstanceLimit(userId: string): Promise<void> {
	const subscription = await db.subscription.findFirst({
		where: {
			userId,
			status: { in: ["active", "trialing"] },
		},
		include: {
			plan: { select: { maxInstances: true } },
		},
		orderBy: { createdAt: "desc" },
	});

	if (!subscription) {
		throw new AppError(
			"Nenhuma assinatura ativa encontrada",
			402,
			"NO_SUBSCRIPTION",
		);
	}

	const activeCount = await db.instance.count({
		where: {
			userId,
			status: { not: "deleted" },
		},
	});

	if (activeCount >= subscription.plan.maxInstances) {
		throw new AppError(
			`Limite de ${subscription.plan.maxInstances} instancias atingido. Faca upgrade do plano para criar mais.`,
			403,
			"INSTANCE_LIMIT_REACHED",
		);
	}
}
