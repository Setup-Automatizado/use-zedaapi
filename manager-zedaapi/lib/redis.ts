import IORedis from "ioredis";

const globalForRedis = globalThis as unknown as {
	redis: IORedis | undefined;
};

function createClient(): IORedis {
	const url = process.env.REDIS_URL || "redis://localhost:6379";
	const conn = new IORedis(url, {
		maxRetriesPerRequest: null,
		enableReadyCheck: false,
		retryStrategy(times) {
			if (times > 10) return null;
			return Math.min(times * 500, 5000);
		},
	});

	conn.on("error", (err) => {
		console.error("[redis] connection error:", err.message);
	});

	return conn;
}

export const redis = globalForRedis.redis ?? createClient();

if (process.env.NODE_ENV !== "production") {
	globalForRedis.redis = redis;
}
