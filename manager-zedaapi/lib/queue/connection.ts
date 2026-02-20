import IORedis from "ioredis";
import type { ConnectionOptions } from "bullmq";
import { createLogger } from "./logger";

const log = createLogger("redis");

const globalForRedis = global as unknown as {
	redisConnection: IORedis | undefined;
};

function attachConnectionLogs(conn: IORedis, label: string): void {
	conn.on("connect", () => {
		log.info(`${label} connected`);
	});
	conn.on("ready", () => {
		log.info(`${label} ready`);
	});
	conn.on("error", (err) => {
		log.error(`${label} error`, { error: err.message });
	});
	conn.on("close", () => {
		log.warn(`${label} closed`);
	});
	conn.on("reconnecting", () => {
		log.info(`${label} reconnecting...`);
	});
	conn.on("end", () => {
		log.warn(`${label} ended`);
	});
}

function createConnection(): IORedis {
	const url = process.env.REDIS_URL || "redis://localhost:6379";
	log.info("Creating shared Redis connection", {
		url: url.replace(/\/\/.*@/, "//<credentials>@"),
	});

	const conn = new IORedis(url, {
		maxRetriesPerRequest: null,
		enableReadyCheck: false,
		retryStrategy(times) {
			if (times > 10) {
				log.error("Max reconnect attempts reached (10), giving up");
				return null;
			}
			const delay = Math.min(times * 500, 5000);
			log.warn(`Reconnect attempt ${times}/10`, { delay: `${delay}ms` });
			return delay;
		},
	});

	attachConnectionLogs(conn, "shared");
	return conn;
}

export function getConnection(): ConnectionOptions {
	if (!globalForRedis.redisConnection) {
		globalForRedis.redisConnection = createConnection();
	}
	return globalForRedis.redisConnection as unknown as ConnectionOptions;
}

let workerConnCount = 0;

export function createWorkerConnection(): ConnectionOptions {
	workerConnCount++;
	const label = `worker-${workerConnCount}`;
	const url = process.env.REDIS_URL || "redis://localhost:6379";

	log.debug(`Creating ${label} connection`);

	const conn = new IORedis(url, {
		maxRetriesPerRequest: null,
		enableReadyCheck: false,
	});

	attachConnectionLogs(conn, label);
	return conn as unknown as ConnectionOptions;
}
