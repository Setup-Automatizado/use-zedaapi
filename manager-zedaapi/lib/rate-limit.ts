import Redis from "ioredis";

let _redis: Redis | null = null;

function getRedis(): Redis {
	if (!_redis) {
		_redis = new Redis(process.env.REDIS_URL || "redis://localhost:6379", {
			maxRetriesPerRequest: 3,
			lazyConnect: true,
		});
	}
	return _redis;
}

interface RateLimitConfig {
	max: number;
	windowMs: number;
	prefix: string;
}

interface RateLimitResult {
	success: boolean;
	remaining: number;
	resetMs: number;
}

export const RATE_LIMIT_CONFIGS = {
	API: { max: 100, windowMs: 60_000, prefix: "rl:api" },
	AUTH: { max: 10, windowMs: 900_000, prefix: "rl:auth" },
	WEBHOOK: { max: 1000, windowMs: 60_000, prefix: "rl:webhook" },
	INSTANCE_CREATE: { max: 5, windowMs: 3_600_000, prefix: "rl:instance" },
} as const;

export async function checkRateLimit(
	identifier: string,
	config: RateLimitConfig,
): Promise<RateLimitResult> {
	const now = Date.now();
	const bucket = Math.floor(now / config.windowMs);
	const key = `${config.prefix}:${identifier}:${bucket}`;
	const ttlSeconds = Math.ceil(config.windowMs / 1000);
	const resetMs = (bucket + 1) * config.windowMs;

	try {
		const redis = getRedis();
		const count = await redis.incr(key);

		if (count === 1) {
			await redis.expire(key, ttlSeconds);
		}

		return {
			success: count <= config.max,
			remaining: Math.max(0, config.max - count),
			resetMs,
		};
	} catch {
		return { success: true, remaining: config.max, resetMs };
	}
}

export function validateOrigin(request: Request): boolean {
	const appUrl =
		process.env.NEXT_PUBLIC_APP_URL || process.env.BETTER_AUTH_URL;

	if (!appUrl) return true;

	const allowed = new URL(appUrl);
	const isDev =
		allowed.hostname === "localhost" || allowed.hostname === "127.0.0.1";

	const origin = request.headers.get("origin");
	const referer = request.headers.get("referer");

	if (!origin && !referer) return false;

	if (origin) {
		try {
			const parsed = new URL(origin);
			if (
				isDev &&
				(parsed.hostname === "localhost" ||
					parsed.hostname === "127.0.0.1")
			)
				return true;
			return parsed.origin === allowed.origin;
		} catch {
			return false;
		}
	}

	if (referer) {
		try {
			const parsed = new URL(referer);
			if (
				isDev &&
				(parsed.hostname === "localhost" ||
					parsed.hostname === "127.0.0.1")
			)
				return true;
			return parsed.origin === allowed.origin;
		} catch {
			return false;
		}
	}

	return false;
}

export function getClientIp(request: Request): string {
	const forwarded = request.headers.get("x-forwarded-for");
	if (forwarded) return forwarded.split(",")[0]!.trim();

	const realIp = request.headers.get("x-real-ip");
	if (realIp) return realIp;

	return "unknown";
}

export function rateLimitHeaders(
	result: RateLimitResult,
	config: RateLimitConfig,
) {
	return {
		"X-RateLimit-Limit": String(config.max),
		"X-RateLimit-Remaining": String(result.remaining),
		"X-RateLimit-Reset": String(Math.ceil(result.resetMs / 1000)),
	};
}
