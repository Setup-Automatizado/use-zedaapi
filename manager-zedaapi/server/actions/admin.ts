"use server";

import { revalidatePath } from "next/cache";
import { requireAdmin } from "@/lib/auth-server";
import { db } from "@/lib/db";
import type { ActionResult } from "@/types";

// =============================================================================
// Dashboard Stats
// =============================================================================

export async function getDashboardStats() {
	try {
		await requireAdmin();

		const now = new Date();
		const thirtyDaysAgo = new Date(now);
		thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);

		const [
			activeSubscriptions,
			totalInstances,
			newUsers30d,
			currentRevenue,
			previousRevenue,
			canceledLast30d,
		] = await Promise.all([
			db.subscription.count({ where: { status: "active" } }),
			db.instance.count({ where: { status: { not: "deleted" } } }),
			db.user.count({ where: { createdAt: { gte: thirtyDaysAgo } } }),
			db.subscription.findMany({
				where: { status: "active" },
				include: { plan: { select: { price: true } } },
			}),
			db.subscription.findMany({
				where: {
					status: "active",
					createdAt: { lte: thirtyDaysAgo },
				},
				include: { plan: { select: { price: true } } },
			}),
			db.subscription.count({
				where: {
					status: "canceled",
					canceledAt: { gte: thirtyDaysAgo },
				},
			}),
		]);

		const mrr = currentRevenue.reduce(
			(sum, sub) => sum + Number(sub.plan.price),
			0,
		);
		const previousMrr = previousRevenue.reduce(
			(sum, sub) => sum + Number(sub.plan.price),
			0,
		);
		const mrrChange =
			previousMrr > 0 ? ((mrr - previousMrr) / previousMrr) * 100 : 0;

		const totalAtStart = activeSubscriptions + canceledLast30d;
		const churnRate =
			totalAtStart > 0 ? (canceledLast30d / totalAtStart) * 100 : 0;

		return {
			mrr,
			mrrChange: Math.round(mrrChange * 10) / 10,
			churnRate: Math.round(churnRate * 10) / 10,
			activeSubscriptions,
			totalInstances,
			newUsers30d,
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return {
			mrr: 0,
			mrrChange: 0,
			churnRate: 0,
			activeSubscriptions: 0,
			totalInstances: 0,
			newUsers30d: 0,
		};
	}
}

// =============================================================================
// Revenue History
// =============================================================================

export async function getRevenueHistory(): Promise<
	ActionResult<Array<{ month: string; value: number }>>
> {
	try {
		await requireAdmin();

		const months = [
			"Jan",
			"Fev",
			"Mar",
			"Abr",
			"Mai",
			"Jun",
			"Jul",
			"Ago",
			"Set",
			"Out",
			"Nov",
			"Dez",
		];
		const now = new Date();
		const results: Array<{ month: string; value: number }> = [];

		for (let i = 5; i >= 0; i--) {
			const date = new Date(now);
			date.setMonth(date.getMonth() - i);
			const startOfMonth = new Date(
				date.getFullYear(),
				date.getMonth(),
				1,
			);
			const endOfMonth = new Date(
				date.getFullYear(),
				date.getMonth() + 1,
				0,
			);

			const invoices = await db.invoice.aggregate({
				where: {
					status: "paid",
					paidAt: { gte: startOfMonth, lte: endOfMonth },
				},
				_sum: { amount: true },
			});

			results.push({
				month: months[date.getMonth()] ?? "?",
				value: Number(invoices._sum.amount || 0),
			});
		}

		return { success: true, data: results };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch revenue history" };
	}
}

// =============================================================================
// Subscriptions by Plan
// =============================================================================

export async function getSubscriptionsByPlan(): Promise<
	ActionResult<Array<{ name: string; value: number; fill: string }>>
> {
	try {
		await requireAdmin();

		const plans = await db.plan.findMany({
			where: { active: true },
			include: {
				_count: {
					select: { subscriptions: { where: { status: "active" } } },
				},
			},
			orderBy: { sortOrder: "asc" },
		});

		const chartColors = [
			"var(--chart-1)",
			"var(--chart-2)",
			"var(--chart-3)",
			"var(--chart-4)",
			"var(--chart-5)",
		];

		const data = plans.map((plan, i) => ({
			name: plan.name,
			value: plan._count.subscriptions,
			fill: chartColors[i % chartColors.length] ?? "var(--chart-1)",
		}));

		return { success: true, data };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return {
			success: false,
			error: "Failed to fetch subscriptions by plan",
		};
	}
}

// =============================================================================
// Recent Activity
// =============================================================================

export async function getRecentActivity(): Promise<
	ActionResult<
		Array<{
			id: string;
			action: string;
			resource: string;
			userName: string;
			timestamp: string;
		}>
	>
> {
	try {
		await requireAdmin();

		const logs = await db.activityLog.findMany({
			take: 10,
			orderBy: { createdAt: "desc" },
			include: {
				user: { select: { name: true } },
			},
		});

		const data = logs.map((log) => ({
			id: log.id,
			action: log.action,
			resource: log.resource,
			userName: log.user?.name || "System",
			timestamp: log.createdAt.toISOString(),
		}));

		return { success: true, data };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch recent activity" };
	}
}

// =============================================================================
// Users Management
// =============================================================================

export async function getUsers(params: {
	page: number;
	pageSize: number;
	search?: string;
}) {
	try {
		await requireAdmin();

		const where = params.search
			? {
					OR: [
						{
							name: {
								contains: params.search,
								mode: "insensitive" as const,
							},
						},
						{
							email: {
								contains: params.search,
								mode: "insensitive" as const,
							},
						},
					],
				}
			: {};

		const [data, totalCount] = await Promise.all([
			db.user.findMany({
				where,
				skip: (params.page - 1) * params.pageSize,
				take: params.pageSize,
				orderBy: { createdAt: "desc" },
				select: {
					id: true,
					name: true,
					email: true,
					role: true,
					banned: true,
					createdAt: true,
					_count: { select: { instances: true } },
				},
			}),
			db.user.count({ where }),
		]);

		return {
			data: data.map((u) => ({
				...u,
				instanceCount: u._count.instances,
			})),
			totalCount,
			page: params.page,
			pageSize: params.pageSize,
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return {
			data: [],
			totalCount: 0,
			page: params.page,
			pageSize: params.pageSize,
		};
	}
}

export async function getAdminUsers(
	page: number = 1,
	search?: string,
): Promise<
	ActionResult<{
		items: Array<{
			id: string;
			name: string;
			email: string;
			role: string;
			banned: boolean;
			createdAt: Date;
			instanceCount: number;
		}>;
		total: number;
	}>
> {
	try {
		await requireAdmin();

		const pageSize = 20;
		const where = search
			? {
					OR: [
						{
							name: {
								contains: search,
								mode: "insensitive" as const,
							},
						},
						{
							email: {
								contains: search,
								mode: "insensitive" as const,
							},
						},
					],
				}
			: {};

		const [users, total] = await Promise.all([
			db.user.findMany({
				where,
				skip: (page - 1) * pageSize,
				take: pageSize,
				orderBy: { createdAt: "desc" },
				select: {
					id: true,
					name: true,
					email: true,
					role: true,
					banned: true,
					createdAt: true,
					_count: { select: { instances: true } },
				},
			}),
			db.user.count({ where }),
		]);

		return {
			success: true,
			data: {
				items: users.map((u) => ({
					id: u.id,
					name: u.name,
					email: u.email,
					role: u.role,
					banned: u.banned,
					createdAt: u.createdAt,
					instanceCount: u._count.instances,
				})),
				total,
			},
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch users" };
	}
}

export async function banUser(userId: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		await db.user.update({
			where: { id: userId },
			data: { banned: true },
		});

		revalidatePath("/admin/usuarios");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to ban user" };
	}
}

export async function unbanUser(userId: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		await db.user.update({
			where: { id: userId },
			data: { banned: false, banReason: null, banExpires: null },
		});

		revalidatePath("/admin/usuarios");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to unban user" };
	}
}

export async function setUserRole(
	userId: string,
	role: string,
): Promise<ActionResult> {
	try {
		await requireAdmin();

		if (!["user", "admin"].includes(role)) {
			return { success: false, error: "Invalid role" };
		}

		await db.user.update({
			where: { id: userId },
			data: { role },
		});

		revalidatePath("/admin/usuarios");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to update user role" };
	}
}

// =============================================================================
// Waitlist
// =============================================================================

export async function getWaitlist(
	page: number = 1,
	search?: string,
): Promise<
	ActionResult<{
		items: Array<{
			id: string;
			email: string;
			name: string | null;
			status: string;
			createdAt: Date;
		}>;
		total: number;
	}>
> {
	try {
		await requireAdmin();

		const pageSize = 20;
		const where = search
			? {
					OR: [
						{
							email: {
								contains: search,
								mode: "insensitive" as const,
							},
						},
						{
							name: {
								contains: search,
								mode: "insensitive" as const,
							},
						},
					],
				}
			: {};

		const [entries, total] = await Promise.all([
			db.waitlist.findMany({
				where,
				skip: (page - 1) * pageSize,
				take: pageSize,
				orderBy: { createdAt: "desc" },
			}),
			db.waitlist.count({ where }),
		]);

		return {
			success: true,
			data: {
				items: entries.map((e) => ({
					id: e.id,
					email: e.email,
					name: e.name,
					status: e.status,
					createdAt: e.createdAt,
				})),
				total,
			},
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch waitlist" };
	}
}

export async function approveWaitlist(entryId: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		// Look up the entry to get the email
		const entry = await db.waitlist.findUnique({
			where: { id: entryId },
			select: { email: true, name: true },
		});

		if (!entry) {
			return { success: false, error: "Entrada não encontrada" };
		}

		// Use the Better Auth waitlist plugin API to approve
		// This generates an invite code and triggers sendInviteEmail callback
		const { auth } = await import("@/lib/auth");
		await auth.api.approveEntry({
			body: { email: entry.email },
		});

		revalidatePath("/admin/lista-de-espera");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return {
			success: false,
			error: "Falha ao aprovar entrada da waitlist",
		};
	}
}

export async function rejectWaitlist(entryId: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		// Look up the entry to get the email
		const entry = await db.waitlist.findUnique({
			where: { id: entryId },
			select: { email: true },
		});

		if (!entry) {
			return { success: false, error: "Entrada não encontrada" };
		}

		// Use the Better Auth waitlist plugin API to reject
		const { auth } = await import("@/lib/auth");
		await auth.api.rejectEntry({
			body: { email: entry.email },
		});

		revalidatePath("/admin/lista-de-espera");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return {
			success: false,
			error: "Falha ao rejeitar entrada da waitlist",
		};
	}
}

// =============================================================================
// Activity Log
// =============================================================================

export async function getActivityLog(params: {
	page: number;
	pageSize: number;
	search?: string;
}) {
	try {
		await requireAdmin();

		const where = params.search
			? {
					OR: [
						{
							action: {
								contains: params.search,
								mode: "insensitive" as const,
							},
						},
						{
							resource: {
								contains: params.search,
								mode: "insensitive" as const,
							},
						},
					],
				}
			: {};

		const [data, totalCount] = await Promise.all([
			db.activityLog.findMany({
				where,
				skip: (params.page - 1) * params.pageSize,
				take: params.pageSize,
				orderBy: { createdAt: "desc" },
				include: {
					user: { select: { name: true, email: true } },
				},
			}),
			db.activityLog.count({ where }),
		]);

		return {
			data: data.map((log) => ({
				id: log.id,
				action: log.action,
				resource: log.resource,
				resourceId: log.resourceId,
				userName: log.user?.name || "System",
				userEmail: log.user?.email || "",
				timestamp: log.createdAt.toISOString(),
				details: log.metadata ? JSON.stringify(log.metadata) : null,
				ipAddress: log.ipAddress,
			})),
			totalCount,
			page: params.page,
			pageSize: params.pageSize,
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return {
			data: [],
			totalCount: 0,
			page: params.page,
			pageSize: params.pageSize,
		};
	}
}

export async function getAdminActivityLog(
	page: number = 1,
	search?: string,
): Promise<
	ActionResult<{
		items: Array<{
			id: string;
			action: string;
			resource: string;
			resourceId: string;
			userName: string;
			userEmail: string;
			timestamp: string;
			details: string | null;
		}>;
		total: number;
	}>
> {
	try {
		await requireAdmin();

		const pageSize = 20;
		const where = search
			? {
					OR: [
						{
							action: {
								contains: search,
								mode: "insensitive" as const,
							},
						},
						{
							resource: {
								contains: search,
								mode: "insensitive" as const,
							},
						},
					],
				}
			: {};

		const [logs, total] = await Promise.all([
			db.activityLog.findMany({
				where,
				skip: (page - 1) * pageSize,
				take: pageSize,
				orderBy: { createdAt: "desc" },
				include: {
					user: { select: { name: true, email: true } },
				},
			}),
			db.activityLog.count({ where }),
		]);

		return {
			success: true,
			data: {
				items: logs.map((log) => ({
					id: log.id,
					action: log.action,
					resource: log.resource,
					resourceId: log.resourceId || "",
					userName: log.user?.name || "System",
					userEmail: log.user?.email || "",
					timestamp: log.createdAt.toISOString(),
					details: log.metadata ? JSON.stringify(log.metadata) : null,
				})),
				total,
			},
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch activity log" };
	}
}

// =============================================================================
// System Settings
// =============================================================================

export async function getSystemSettings(): Promise<
	ActionResult<
		Array<{
			id: string;
			key: string;
			value: string;
			description: string | null;
		}>
	>
> {
	try {
		await requireAdmin();

		const settings = await db.systemSetting.findMany({
			orderBy: { key: "asc" },
		});

		return {
			success: true,
			data: settings.map((s) => ({
				id: s.id,
				key: s.key,
				value: s.value,
				description: s.description,
			})),
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch system settings" };
	}
}

export async function updateSystemSetting(
	key: string,
	value: string,
): Promise<ActionResult> {
	try {
		const session = await requireAdmin();

		await db.systemSetting.upsert({
			where: { key },
			create: { key, value, updatedBy: session.user.id },
			update: { value, updatedBy: session.user.id },
		});

		revalidatePath("/admin/configuracoes");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to update system setting" };
	}
}

// =============================================================================
// Feature Flags
// =============================================================================

export async function getFeatureFlags(): Promise<
	ActionResult<
		Array<{
			id: string;
			key: string;
			enabled: boolean;
			description: string | null;
		}>
	>
> {
	try {
		await requireAdmin();

		const flags = await db.featureFlag.findMany({
			orderBy: { key: "asc" },
		});

		return {
			success: true,
			data: flags.map((f) => ({
				id: f.id,
				key: f.key,
				enabled: f.enabled,
				description: f.description,
			})),
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch feature flags" };
	}
}

export async function toggleFeatureFlag(key: string): Promise<ActionResult> {
	try {
		await requireAdmin();

		const flag = await db.featureFlag.findUnique({ where: { key } });
		if (!flag) {
			return { success: false, error: "Feature flag not found" };
		}

		await db.featureFlag.update({
			where: { key },
			data: { enabled: !flag.enabled },
		});

		revalidatePath("/admin/funcionalidades");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to toggle feature flag" };
	}
}

// =============================================================================
// Plans Management
// =============================================================================

export async function getAdminPlans(): Promise<
	ActionResult<
		Array<{
			id: string;
			name: string;
			slug: string;
			description: string | null;
			price: number;
			currency: string;
			interval: string;
			maxInstances: number;
			features: unknown;
			active: boolean;
			sortOrder: number;
		}>
	>
> {
	try {
		await requireAdmin();

		const plans = await db.plan.findMany({
			orderBy: { sortOrder: "asc" },
		});

		return {
			success: true,
			data: plans.map((p) => ({
				id: p.id,
				name: p.name,
				slug: p.slug,
				description: p.description,
				price: Number(p.price),
				currency: p.currency,
				interval: p.interval,
				maxInstances: p.maxInstances,
				features: p.features,
				active: p.active,
				sortOrder: p.sortOrder,
			})),
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch plans" };
	}
}

export async function createPlan(data: {
	name: string;
	slug: string;
	price: number;
	maxInstances: number;
	features: unknown;
}): Promise<ActionResult> {
	try {
		await requireAdmin();

		const existing = await db.plan.findUnique({
			where: { slug: data.slug },
		});
		if (existing) {
			return {
				success: false,
				error: "A plan with this slug already exists",
			};
		}

		await db.plan.create({
			data: {
				name: data.name,
				slug: data.slug,
				price: data.price,
				maxInstances: data.maxInstances,
				features: data.features as object,
			},
		});

		revalidatePath("/admin/planos");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to create plan" };
	}
}

export async function updatePlan(
	id: string,
	data: {
		name?: string;
		price?: number;
		maxInstances?: number;
		active?: boolean;
	},
): Promise<ActionResult> {
	try {
		await requireAdmin();

		await db.plan.update({
			where: { id },
			data,
		});

		revalidatePath("/admin/planos");
		return { success: true };
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to update plan" };
	}
}

// =============================================================================
// Admin Invoices
// =============================================================================

export async function getAdminInvoices(page: number = 1): Promise<
	ActionResult<{
		items: Array<{
			id: string;
			userName: string;
			amount: number;
			status: string;
			paymentMethod: string | null;
			nfseStatus: string | null;
			createdAt: Date;
		}>;
		total: number;
	}>
> {
	try {
		await requireAdmin();

		const pageSize = 20;
		const [invoices, total] = await Promise.all([
			db.invoice.findMany({
				skip: (page - 1) * pageSize,
				take: pageSize,
				orderBy: { createdAt: "desc" },
				include: {
					user: { select: { name: true } },
				},
			}),
			db.invoice.count(),
		]);

		return {
			success: true,
			data: {
				items: invoices.map((inv) => ({
					id: inv.id,
					userName: inv.user.name,
					amount: Number(inv.amount),
					status: inv.status,
					paymentMethod: inv.paymentMethod,
					nfseStatus: inv.nfseStatus,
					createdAt: inv.createdAt,
				})),
				total,
			},
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch invoices" };
	}
}

// =============================================================================
// Admin Subscriptions
// =============================================================================

export async function getAdminSubscriptions(page: number = 1): Promise<
	ActionResult<{
		items: Array<{
			id: string;
			userName: string;
			userEmail: string;
			planName: string;
			status: string;
			amount: number;
			paymentMethod: string;
			currentPeriodEnd: Date;
		}>;
		total: number;
	}>
> {
	try {
		await requireAdmin();

		const pageSize = 20;
		const [subscriptions, total] = await Promise.all([
			db.subscription.findMany({
				skip: (page - 1) * pageSize,
				take: pageSize,
				orderBy: { createdAt: "desc" },
				include: {
					user: { select: { name: true, email: true } },
					plan: { select: { name: true, price: true } },
				},
			}),
			db.subscription.count(),
		]);

		return {
			success: true,
			data: {
				items: subscriptions.map((sub) => ({
					id: sub.id,
					userName: sub.user.name,
					userEmail: sub.user.email,
					planName: sub.plan.name,
					status: sub.status,
					amount: Number(sub.plan.price),
					paymentMethod: sub.paymentMethod,
					currentPeriodEnd: sub.currentPeriodEnd,
				})),
				total,
			},
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch subscriptions" };
	}
}

// =============================================================================
// Admin Instances
// =============================================================================

export async function getAdminInstances(page: number = 1): Promise<
	ActionResult<{
		items: Array<{
			id: string;
			name: string;
			userName: string;
			status: string;
			phone: string | null;
			whatsappConnected: boolean;
			planName: string | null;
			createdAt: Date;
		}>;
		total: number;
	}>
> {
	try {
		await requireAdmin();

		const pageSize = 20;
		const [instances, total] = await Promise.all([
			db.instance.findMany({
				where: { status: { not: "deleted" } },
				skip: (page - 1) * pageSize,
				take: pageSize,
				orderBy: { createdAt: "desc" },
				include: {
					user: { select: { name: true } },
					subscription: {
						include: { plan: { select: { name: true } } },
					},
				},
			}),
			db.instance.count({ where: { status: { not: "deleted" } } }),
		]);

		return {
			success: true,
			data: {
				items: instances.map((inst) => ({
					id: inst.id,
					name: inst.name,
					userName: inst.user.name,
					status: inst.status,
					phone: inst.phone,
					whatsappConnected: inst.whatsappConnected,
					planName: inst.subscription?.plan?.name || null,
					createdAt: inst.createdAt,
				})),
				total,
			},
		};
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT")
			throw error;
		return { success: false, error: "Failed to fetch instances" };
	}
}
