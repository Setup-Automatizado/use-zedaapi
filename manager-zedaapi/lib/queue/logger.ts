/**
 * Structured logger for BullMQ workers.
 *
 * Features:
 * - ISO timestamps on every line
 * - Log levels: DEBUG, INFO, WARN, ERROR
 * - Duration tracking via timer()
 * - Structured context (key=value pairs)
 * - Scoped loggers per worker/processor
 */

type LogLevel = "DEBUG" | "INFO" | "WARN" | "ERROR";

const LOG_LEVEL_PRIORITY: Record<LogLevel, number> = {
	DEBUG: 0,
	INFO: 1,
	WARN: 2,
	ERROR: 3,
};

function getMinLevel(): LogLevel {
	const env = process.env.WORKER_LOG_LEVEL?.toUpperCase();
	if (env && env in LOG_LEVEL_PRIORITY) return env as LogLevel;
	return "DEBUG";
}

function shouldLog(level: LogLevel): boolean {
	return LOG_LEVEL_PRIORITY[level] >= LOG_LEVEL_PRIORITY[getMinLevel()];
}

function formatContext(ctx?: Record<string, unknown>): string {
	if (!ctx || Object.keys(ctx).length === 0) return "";
	const parts = Object.entries(ctx)
		.filter(([, v]) => v !== undefined && v !== null)
		.map(
			([k, v]) => `${k}=${typeof v === "object" ? JSON.stringify(v) : v}`,
		);
	return parts.length > 0 ? ` | ${parts.join(", ")}` : "";
}

function formatDuration(ms: number): string {
	if (ms < 1000) return `${ms}ms`;
	if (ms < 60_000) return `${(ms / 1000).toFixed(1)}s`;
	return `${(ms / 60_000).toFixed(1)}min`;
}

function log(
	level: LogLevel,
	scope: string,
	message: string,
	ctx?: Record<string, unknown>,
): void {
	if (!shouldLog(level)) return;

	const timestamp = new Date().toISOString();
	const contextStr = formatContext(ctx);
	const line = `${timestamp} [${level}] [${scope}] ${message}${contextStr}`;

	switch (level) {
		case "ERROR":
			console.error(line);
			break;
		case "WARN":
			console.warn(line);
			break;
		default:
			console.log(line);
	}
}

export interface Logger {
	debug(message: string, ctx?: Record<string, unknown>): void;
	info(message: string, ctx?: Record<string, unknown>): void;
	warn(message: string, ctx?: Record<string, unknown>): void;
	error(message: string, ctx?: Record<string, unknown>): void;
	timer(label: string, ctx?: Record<string, unknown>): () => void;
	child(subscope: string): Logger;
}

export function createLogger(scope: string): Logger {
	return {
		debug: (msg, ctx) => log("DEBUG", scope, msg, ctx),
		info: (msg, ctx) => log("INFO", scope, msg, ctx),
		warn: (msg, ctx) => log("WARN", scope, msg, ctx),
		error: (msg, ctx) => log("ERROR", scope, msg, ctx),

		timer(label: string, ctx?: Record<string, unknown>) {
			const start = Date.now();
			log("DEBUG", scope, `${label} started`, ctx);
			return () => {
				const elapsed = Date.now() - start;
				log("INFO", scope, `${label} completed`, {
					...ctx,
					duration: formatDuration(elapsed),
				});
			};
		},

		child(subscope: string): Logger {
			return createLogger(`${scope}:${subscope}`);
		},
	};
}
