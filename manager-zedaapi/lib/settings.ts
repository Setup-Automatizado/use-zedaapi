"use server";

import { db } from "@/lib/db";

/**
 * Get a system setting value by key.
 */
export async function getSystemSetting(key: string): Promise<string | null> {
	try {
		const setting = await db.systemSetting.findUnique({
			where: { key },
			select: { value: true },
		});
		return setting?.value ?? null;
	} catch {
		return null;
	}
}

/**
 * Get a feature flag status by key.
 */
export async function getFeatureFlag(key: string): Promise<boolean> {
	try {
		const flag = await db.featureFlag.findUnique({
			where: { key },
			select: { enabled: true },
		});
		return flag?.enabled ?? false;
	} catch {
		return false;
	}
}
